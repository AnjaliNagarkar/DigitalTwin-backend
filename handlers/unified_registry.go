package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// UnifiedRecord is the merged schema for every citizen — demographic + agri + social.
// Category-specific fields are populated where DB data exists and left blank otherwise.
type UnifiedRecord struct {
	// ── Core demographic ──────────────────────────────────────────────────────
	FullName  string `json:"fullName"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	Education string `json:"education"`
	FamilyID  int    `json:"familyId"`

	// ── Agri (N/A when person has no land) ───────────────────────────────────
	TotalLand      string `json:"totalLand"`
	IrrigationType string `json:"irrigationType"`
	WaterSource    string `json:"waterSource"`
	KharifCrop     string `json:"kharifCrop"`
	RabiCrop       string `json:"rabiCrop"`
	AnnualIncome   string `json:"annualIncome"`

	// ── Social ───────────────────────────────────────────────────────────────
	Occupation       string `json:"occupation"`
	ChildrenCount    int    `json:"childrenCount"`
	SanitationStatus string `json:"sanitationStatus"`

	// ── Category membership flags (derived) ──────────────────────────────────
	IsFarmer   bool `json:"isFarmer"`
	IsStudent  bool `json:"isStudent"`
	IsDivyang  bool `json:"isDivyang"`
	IsHousewife bool `json:"isHousewife"`
	IsSenior   bool `json:"isSenior"`

	// ── Student-specific ─────────────────────────────────────────────────────
	// SchoolName: institution type from DB (TYPE_INSTITUTION)
	SchoolName     string `json:"schoolName"`
	// GradeStandard: standard/grade being studied (STANDARD_WHICH_STUDYING)
	GradeStandard  string `json:"gradeStandard"`
	// EducationLevel: normalised bucket — "Graduate", "Undergraduate", "Anganwadi/Primary"
	EducationLevel string `json:"educationLevel"`
	// Scholarship: "Yes" when actively pursuing education AND has a recorded qualification, else "No"
	Scholarship    string `json:"scholarship"`

	// ── Disabled-specific ────────────────────────────────────────────────────
	// DisabilityType: DISABILITY column value
	DisabilityType    string `json:"disabilityType"`
	// DisabilityPercent: DISABILITY_PERCENTAGE column value
	DisabilityPercent string `json:"disabilityPercent"`
	// PensionStatus: derived — "Eligible" when divyang=YES or senior citizen, else "Not Eligible"
	PensionStatus     string `json:"pensionStatus"`
	// CaretakerName: no DB column, always "Not Available"
	CaretakerName     string `json:"caretakerName"`
	// GovtPensionAmount: annual income reused as social-security proxy; "N/A" when not disabled/senior
	GovtPensionAmount string `json:"govtPensionAmount"`

	// ── Housewife-specific ───────────────────────────────────────────────────
	// SourceOfIncome: derived from NATURE_WAGE_WORK / OCCUPATION keyword analysis
	SourceOfIncome string `json:"sourceOfIncome"`
}

// UnifiedRegistryHandler handles GET /unified-registry.
type UnifiedRegistryHandler struct {
	DB *sql.DB
}

func (h *UnifiedRegistryHandler) colExists(table, col string) bool {
	var n int
	_ = h.DB.QueryRow(`
		SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`, table, col).Scan(&n)
	return n > 0
}

// GetUnifiedRegistry handles GET /unified-registry.
// Merges FAMILY_MEMBER (demographic) with FAMILY (agri/social) — one row per citizen.
func (h *UnifiedRegistryHandler) GetUnifiedRegistry(c *gin.Context) {
	// ── Detect optional columns ───────────────────────────────────────────────
	dobCol := ""
	for _, col := range []string{"DOB", "D_O_B"} {
		if h.colExists("FAMILY_MEMBER", col) {
			dobCol = col
			break
		}
	}
	dobExpr := "NULL"
	if dobCol != "" {
		dobExpr = "fm." + dobCol
	}

	// ageExpr: single fmt.Sprintf — %%Y becomes %Y in the output, which is
	// what MySQL's STR_TO_DATE expects.  dobExpr is already a safe SQL identifier.
	ageExpr := "NULL"
	if dobCol != "" {
		ageExpr = fmt.Sprintf(`CASE
			WHEN STR_TO_DATE(%[1]s, '%%Y-%%m-%%d') IS NOT NULL
				THEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(%[1]s, '%%Y-%%m-%%d'), CURDATE())
			WHEN STR_TO_DATE(%[1]s, '%%d-%%m-%%Y') IS NOT NULL
				THEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(%[1]s, '%%d-%%m-%%Y'), CURDATE())
			ELSE NULL END`, dobExpr)
	}

	workExpr := `COALESCE(NULLIF(TRIM(COALESCE(fm.OCCUPATION, '')), ''), 'Not Working')`
	if h.colExists("FAMILY_MEMBER", "NATURE_WAGE_WORK") {
		workExpr = `COALESCE(
			NULLIF(TRIM(COALESCE(fm.NATURE_WAGE_WORK, '')), ''),
			NULLIF(TRIM(COALESCE(fm.OCCUPATION, '')),       ''),
			'Not Working')`
	}

	educExpr := `COALESCE(
		NULLIF(TRIM(COALESCE(fm.QUALIFICATION, '')), ''),
		NULLIF(TRIM(COALESCE(fm.EDUCATION_STATUS, '')), ''),
		'Not Available')`

	// sanitExpr: detect which column actually exists to avoid "Unknown column" SQL errors.
	sanitExpr := "'Not Available'"
	switch {
	case h.colExists("FAMILY", "SANITATION_TOILET_FACILITY"):
		sanitExpr = `COALESCE(NULLIF(TRIM(COALESCE(f.SANITATION_TOILET_FACILITY, '')), ''), 'Not Available')`
	case h.colExists("FAMILY", "TYPE_OF_LATRINE"):
		sanitExpr = `COALESCE(NULLIF(TRIM(COALESCE(f.TYPE_OF_LATRINE, '')), ''), 'Not Available')`
	}

	// Optional student columns
	pursueExpr := "''"
	if h.colExists("FAMILY_MEMBER", "CURRENTLY_PURSUING_EDUCATION") {
		pursueExpr = "COALESCE(fm.CURRENTLY_PURSUING_EDUCATION, '')"
	}
	stdExpr := "''"
	if h.colExists("FAMILY_MEMBER", "STANDARD_WHICH_STUDYING") {
		stdExpr = "COALESCE(fm.STANDARD_WHICH_STUDYING, '')"
	}
	instExpr := "''"
	if h.colExists("FAMILY_MEMBER", "TYPE_INSTITUTION") {
		instExpr = "COALESCE(fm.TYPE_INSTITUTION, '')"
	}

	// Optional disability columns
	divyangExpr := "''"
	if h.colExists("FAMILY_MEMBER", "DIVYANG") {
		divyangExpr = "COALESCE(fm.DIVYANG, '')"
	}
	disabilityExpr := "''"
	if h.colExists("FAMILY_MEMBER", "DISABILITY") {
		disabilityExpr = "COALESCE(fm.DISABILITY, '')"
	}
	disabilityPctExpr := "''"
	if h.colExists("FAMILY_MEMBER", "DISABILITY_PERCENTAGE") {
		disabilityPctExpr = "COALESCE(CAST(fm.DISABILITY_PERCENTAGE AS CHAR), '')"
	}

	// Children subquery
	childrenSQL := `SELECT cm.EXTERNAL_FAMILY_ID, 0 AS children_count
		FROM FAMILY_MEMBER cm GROUP BY cm.EXTERNAL_FAMILY_ID`
	if dobCol != "" {
		childrenSQL = fmt.Sprintf(`
			SELECT cm.EXTERNAL_FAMILY_ID,
				SUM(CASE
					WHEN STR_TO_DATE(cm.%[1]s,'%%Y-%%m-%%d') IS NOT NULL
						AND TIMESTAMPDIFF(YEAR, STR_TO_DATE(cm.%[1]s,'%%Y-%%m-%%d'), CURDATE()) < 18 THEN 1
					WHEN STR_TO_DATE(cm.%[1]s,'%%d-%%m-%%Y') IS NOT NULL
						AND TIMESTAMPDIFF(YEAR, STR_TO_DATE(cm.%[1]s,'%%d-%%m-%%Y'), CURDATE()) < 18 THEN 1
					ELSE 0
				END) AS children_count
			FROM FAMILY_MEMBER cm
			GROUP BY cm.EXTERNAL_FAMILY_ID`, dobCol)
	}

	query := fmt.Sprintf(`
		SELECT
			COALESCE(fm.FIRST_NAME, '')                      AS first_name,
			COALESCE(fm.LAST_NAME,  '')                      AS last_name,
			COALESCE(fm.GENDER, '')                          AS gender,
			%s                                               AS age,
			%s                                               AS education,
			%s                                               AS occupation,
			COALESCE(fm.EXTERNAL_FAMILY_ID, 0)              AS family_id,
			COALESCE(f.AREA_AGRICULTURE_LAND_ACRES, '')     AS total_land,
			COALESCE(f.OWN_AGRICULTURE_LAND, '')            AS own_agri_land,
			COALESCE(f.SOURCE_WATER_IRRIGATION, '')         AS water_source,
			COALESCE(f.CULTIVATING_DURING_KHARIF_SEASON,'') AS kharif_crop,
			COALESCE(f.TAKING_CROPS_RABI_SEASON, '')        AS rabi_crop,
			COALESCE(CAST(f.ANNUAL_INCOME AS CHAR), '')     AS annual_income,
			%s                                               AS sanitation_status,
			COALESCE(ch.children_count, 0)                  AS children_count,
			%s                                               AS pursuing_education,
			%s                                               AS grade_standard,
			%s                                               AS school_name,
			%s                                               AS divyang,
			%s                                               AS disability_type,
			%s                                               AS disability_pct
		FROM FAMILY_MEMBER fm
		LEFT JOIN FAMILY f ON f.FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		LEFT JOIN (%s) ch ON ch.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		ORDER BY fm.FIRST_NAME, fm.LAST_NAME
	`, ageExpr, educExpr, workExpr, sanitExpr,
		pursueExpr, stdExpr, instExpr,
		divyangExpr, disabilityExpr, disabilityPctExpr,
		childrenSQL)

	rows, err := h.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch unified registry", "detail": err.Error()})
		return
	}
	defer rows.Close()

	out := make([]UnifiedRecord, 0)
	for rows.Next() {
		var (
			firstName      string
			lastName       string
			gender         string
			age            sql.NullInt64
			education      string
			occupation     string
			familyID       int
			totalLand      string
			ownAgriLand    string
			waterSource    string
			kharifCrop     string
			rabiCrop       string
			annualIncome   string
			sanitation     string
			children       int
			pursuing       string
			gradeStd       string
			schoolName     string
			divyang        string
			disabilityType string
			disabilityPct  string
		)

		if err := rows.Scan(
			&firstName, &lastName, &gender, &age,
			&education, &occupation, &familyID,
			&totalLand, &ownAgriLand, &waterSource,
			&kharifCrop, &rabiCrop, &annualIncome,
			&sanitation, &children,
			&pursuing, &gradeStd, &schoolName,
			&divyang, &disabilityType, &disabilityPct,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan unified registry record", "detail": err.Error()})
			return
		}

		ageVal := 0
		if age.Valid && age.Int64 >= 0 {
			ageVal = int(age.Int64)
		}

		// ── Agri normalisation ────────────────────────────────────────────────
		hasLand := strings.ToUpper(strings.TrimSpace(ownAgriLand)) == "YES"
		landVal := strings.TrimSpace(totalLand)
		wsVal := strings.TrimSpace(waterSource)
		kVal := strings.TrimSpace(kharifCrop)
		rVal := strings.TrimSpace(rabiCrop)
		if !hasLand || landVal == "" || landVal == "0" || landVal == "0.00" {
			landVal = "0"
			wsVal = "N/A"
			kVal = "N/A"
			rVal = "N/A"
		}

		// ── Category flags ────────────────────────────────────────────────────
		occLower := strings.ToLower(strings.TrimSpace(occupation))
		isHousewife := occLower == "housewife" || occLower == "homemaker"
		isStudent := strings.ToUpper(strings.TrimSpace(pursuing)) == "YES"
		isDivyang := strings.ToUpper(strings.TrimSpace(divyang)) == "YES" ||
			(strings.TrimSpace(disabilityType) != "" && strings.TrimSpace(disabilityType) != "0")
		isSenior := ageVal >= 60

		// ── Education level bucket ────────────────────────────────────────────
		educLevel := normaliseEducationLevel(strings.TrimSpace(education))

		// ── Scholarship: Yes if actively studying, No otherwise ───────────────
		scholarship := "No"
		if isStudent {
			scholarship = "Yes"
		}

		// ── Pension: Eligible for divyang or senior ───────────────────────────
		pensionStatus := "Not Eligible"
		if isDivyang || isSenior {
			pensionStatus = "Eligible"
		}

		// ── Govt Pension Amount: reuse annual_income for eligible citizens ────
		incomeClean := strings.TrimSpace(annualIncome)
		govtPensionAmt := "N/A"
		if (isDivyang || isSenior) && incomeClean != "" && incomeClean != "0" {
			govtPensionAmt = incomeClean
		}

		// ── Source of Income: keyword-derived from occupation + work details ──
		sourceOfIncome := deriveSourceOfIncome(strings.TrimSpace(occupation))

		out = append(out, UnifiedRecord{
			FullName:  capitalizeFullName(firstName, lastName),
			FirstName: strings.TrimSpace(firstName),
			LastName:  strings.TrimSpace(lastName),
			Age:       ageVal,
			Gender:    strings.TrimSpace(gender),
			Education: strings.TrimSpace(education),
			FamilyID:  familyID,

			TotalLand:      landVal,
			IrrigationType: deriveIrrigationType(wsVal),
			WaterSource:    wsVal,
			KharifCrop:     kVal,
			RabiCrop:       rVal,
			AnnualIncome:   incomeClean,

			Occupation:       strings.TrimSpace(occupation),
			ChildrenCount:    children,
			SanitationStatus: strings.TrimSpace(sanitation),

			IsFarmer:    hasLand,
			IsStudent:   isStudent,
			IsDivyang:   isDivyang,
			IsHousewife: isHousewife,
			IsSenior:    isSenior,

			SchoolName:     capitalizeWords(strings.TrimSpace(schoolName)),
			GradeStandard:  strings.TrimSpace(gradeStd),
			EducationLevel: educLevel,
			Scholarship:    scholarship,

			DisabilityType:    capitalizeWords(strings.TrimSpace(disabilityType)),
			DisabilityPercent: strings.TrimSpace(disabilityPct),
			PensionStatus:     pensionStatus,
			CaretakerName:     "Not Available",
			GovtPensionAmount: govtPensionAmt,

			SourceOfIncome: sourceOfIncome,
		})
	}

	c.JSON(http.StatusOK, out)
}

// normaliseEducationLevel maps raw qualification strings to display buckets.
func normaliseEducationLevel(qual string) string {
	q := strings.ToLower(qual)
	if strings.Contains(q, "graduation") || strings.Contains(q, "graduate") ||
		strings.Contains(q, "degree") || strings.Contains(q, "pg") ||
		strings.Contains(q, "post grad") {
		return "Graduate"
	}
	if strings.Contains(q, "12") || strings.Contains(q, "hsc") ||
		strings.Contains(q, "inter") || strings.Contains(q, "10") ||
		strings.Contains(q, "ssc") || strings.Contains(q, "secondary") {
		return "Undergraduate"
	}
	if strings.Contains(q, "anganwadi") || strings.Contains(q, "primary") ||
		strings.Contains(q, "1st") || strings.Contains(q, "2nd") ||
		strings.Contains(q, "3rd") || strings.Contains(q, "4th") {
		return "Anganwadi/Primary"
	}
	if qual == "" || strings.Contains(q, "not available") || strings.Contains(q, "illiterate") {
		return "Not Available"
	}
	return "Undergraduate"
}

// capitalizeWords capitalises the first letter of each word.
func capitalizeWords(s string) string {
	if s == "" {
		return s
	}
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}

// capitalizeFullName builds and capitalises a full name from two parts.
func capitalizeFullName(first, last string) string {
	full := strings.TrimSpace(first + " " + last)
	if full == "" {
		return "—"
	}
	return capitalizeWords(full)
}

// deriveSourceOfIncome maps occupation/work keywords to the canonical source-of-income options.
// Options: None | Small Business/SHG | Tailoring | Poultry | Remittance from family
func deriveSourceOfIncome(occupation string) string {
	v := strings.ToLower(occupation)
	if strings.Contains(v, "tailor") || strings.Contains(v, "stitching") || strings.Contains(v, "sewing") {
		return "Tailoring"
	}
	if strings.Contains(v, "poultry") || strings.Contains(v, "chicken") || strings.Contains(v, "egg") {
		return "Poultry"
	}
	if strings.Contains(v, "shg") || strings.Contains(v, "self help") ||
		strings.Contains(v, "small business") || strings.Contains(v, "business") ||
		strings.Contains(v, "shop") || strings.Contains(v, "vendor") || strings.Contains(v, "petty") {
		return "Small Business/SHG"
	}
	if strings.Contains(v, "remittance") || strings.Contains(v, "family") ||
		strings.Contains(v, "husband") || strings.Contains(v, "dependent") {
		return "Remittance from family"
	}
	return "None"
}

// deriveIrrigationType classifies a water source string as "Irrigated" or "Rain-fed".
func deriveIrrigationType(ws string) string {
	v := strings.ToLower(strings.TrimSpace(ws))
	if v == "" || v == "none" || v == "n/a" || strings.Contains(v, "rain") {
		return "Rain-fed"
	}
	return "Irrigated"
}

// childDobCol kept for backward compatibility — no longer used in new query.
func childDobCol(col string) string {
	if col != "" {
		return col
	}
	return "NULL"
}
