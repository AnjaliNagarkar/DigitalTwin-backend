package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	IsFarmer    bool `json:"isFarmer"`
	IsStudent   bool `json:"isStudent"`
	IsDivyang   bool `json:"isDivyang"`
	IsHousewife bool `json:"isHousewife"`
	IsSenior    bool `json:"isSenior"`

	// ── Student-specific ─────────────────────────────────────────────────────
	SchoolName     string `json:"schoolName"`
	GradeStandard  string `json:"gradeStandard"`
	EducationLevel string `json:"educationLevel"`
	Scholarship    string `json:"scholarship"`

	// ── Disabled-specific ────────────────────────────────────────────────────
	DisabilityType    string `json:"disabilityType"`
	DisabilityPercent string `json:"disabilityPercent"`
	PensionStatus     string `json:"pensionStatus"`
	CaretakerName     string `json:"caretakerName"`
	GovtPensionAmount string `json:"govtPensionAmount"`

	MaritalStatus string `json:"maritalStatus"` // New field
	// ── Housewife-specific ───────────────────────────────────────────────────
	SourceOfIncome string `json:"sourceOfIncome"`
}

// UnifiedRegistryHandler handles GET /unified-registry.
// Column detection happens once at construction (NewUnifiedRegistryHandler) so
// every request just runs the pre-built SQL — no per-request INFORMATION_SCHEMA queries.
type UnifiedRegistryHandler struct {
	DB    *sql.DB
	query string
}

// colsExist returns a set of columns that exist in table.
// Uses a single INFORMATION_SCHEMA query for all columns at once.
func colsExist(db *sql.DB, table string, cols []string) map[string]bool {
	result := make(map[string]bool, len(cols))
	if len(cols) == 0 {
		return result
	}
	placeholders := make([]string, len(cols))
	args := make([]any, len(cols)+1)
	args[0] = table
	for i, c := range cols {
		placeholders[i] = "?"
		args[i+1] = c
	}
	q := fmt.Sprintf(
		`SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS
		 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		   AND COLUMN_NAME IN (%s)`,
		strings.Join(placeholders, ","),
	)
	rows, err := db.Query(q, args...)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if rows.Scan(&name) == nil {
			result[name] = true
		}
	}
	return result
}

// NewUnifiedRegistryHandler detects optional DB columns (2 round-trips total)
// and pre-builds the query string once at startup.
func NewUnifiedRegistryHandler(db *sql.DB) *UnifiedRegistryHandler {
	h := &UnifiedRegistryHandler{DB: db}
	h.query = h.buildQuery()
	return h
}

func (h *UnifiedRegistryHandler) buildQuery() string {
	// Two INFORMATION_SCHEMA queries — one per table — instead of 10+ individual ones.
	fm := colsExist(h.DB, "FAMILY_MEMBER", []string{
		"DOB", "D_O_B",
		"NATURE_WAGE_WORK",
		"CURRENTLY_PURSUING_EDUCATION",
		"STANDARD_WHICH_STUDYING",
		"TYPE_INSTITUTION",
		"DIVYANG",
		"DISABILITY",
		"MARITAL_STATUS", // Add this
		"DISABILITY_PERCENTAGE",
	})
	fam := colsExist(h.DB, "FAMILY", []string{
		"SANITATION_TOILET_FACILITY",
		"TYPE_OF_LATRINE",
	})

	// ── DOB raw value (age is computed in Go to avoid expensive SQL STR_TO_DATE) ──
	dobCol := ""
	if fm["DOB"] {
		dobCol = "DOB"
	} else if fm["D_O_B"] {
		dobCol = "D_O_B"
	}
	dobExpr := "''"
	if dobCol != "" {
		dobExpr = fmt.Sprintf("COALESCE(fm.%s, '')", dobCol)
	}

	// ── Occupation / work ─────────────────────────────────────────────────────
	workExpr := `COALESCE(NULLIF(TRIM(COALESCE(fm.OCCUPATION,'')), ''), 'Not Working')`
	if fm["NATURE_WAGE_WORK"] {
		workExpr = `COALESCE(
			NULLIF(TRIM(COALESCE(fm.NATURE_WAGE_WORK,'')), ''),
			NULLIF(TRIM(COALESCE(fm.OCCUPATION,'')),       ''),
			'Not Working')`
	}

	// ── Education ─────────────────────────────────────────────────────────────
	educExpr := `COALESCE(
		NULLIF(TRIM(COALESCE(fm.QUALIFICATION,'')),      ''),
		NULLIF(TRIM(COALESCE(fm.EDUCATION_STATUS,'')),   ''),
		'Not Available')`

	// ── Sanitation ────────────────────────────────────────────────────────────
	sanitExpr := "'Not Available'"
	if fam["SANITATION_TOILET_FACILITY"] {
		sanitExpr = `COALESCE(NULLIF(TRIM(COALESCE(f.SANITATION_TOILET_FACILITY,'')), ''), 'Not Available')`
	} else if fam["TYPE_OF_LATRINE"] {
		sanitExpr = `COALESCE(NULLIF(TRIM(COALESCE(f.TYPE_OF_LATRINE,'')), ''), 'Not Available')`
	}

	// ── Optional student / disability columns ─────────────────────────────────
	opt := func(col, expr string) string {
		if fm[col] {
			return expr
		}
		return "''"
	}
	pursueExpr := opt("CURRENTLY_PURSUING_EDUCATION", "COALESCE(fm.CURRENTLY_PURSUING_EDUCATION,'')")
	stdExpr := opt("STANDARD_WHICH_STUDYING", "COALESCE(fm.STANDARD_WHICH_STUDYING,'')")
	instExpr := opt("TYPE_INSTITUTION", "COALESCE(fm.TYPE_INSTITUTION,'')")
	divyangExpr := opt("DIVYANG", "COALESCE(fm.DIVYANG,'')")
	disabilityExpr := opt("DISABILITY", "COALESCE(fm.DISABILITY,'')")
	disabilityPctExpr := opt("DISABILITY_PERCENTAGE", "COALESCE(CAST(fm.DISABILITY_PERCENTAGE AS CHAR),'')")
	maritalStatusExpr := opt("MARITAL_STATUS", "COALESCE(fm.MARITAL_STATUS,'')") // New expression

	// ORDER BY removed — no index on name columns; frontend handles sorting.
	return fmt.Sprintf(`
		SELECT
			COALESCE(fm.FIRST_NAME, '')                       AS first_name,
			COALESCE(fm.LAST_NAME,  '')                       AS last_name,
			COALESCE(fm.GENDER, '')                           AS gender,
			%s                                                AS dob_raw,
			%s                                                AS education,
			%s                                                AS occupation,
			COALESCE(fm.EXTERNAL_FAMILY_ID, 0)               AS family_id,
			COALESCE(f.AREA_AGRICULTURE_LAND_ACRES, '')      AS total_land,
			COALESCE(f.OWN_AGRICULTURE_LAND, '')             AS own_agri_land,
			COALESCE(f.SOURCE_WATER_IRRIGATION, '')          AS water_source,
			COALESCE(f.CULTIVATING_DURING_KHARIF_SEASON, '') AS kharif_crop,
			COALESCE(f.TAKING_CROPS_RABI_SEASON, '')         AS rabi_crop,
			COALESCE(CAST(f.ANNUAL_INCOME AS CHAR), '')      AS annual_income,
			%s                                                AS sanitation_status,
			0                                                 AS children_count,
			%s                                                AS pursuing_education,
			%s                                                AS grade_standard,
			%s                                                AS school_name,
			%s                                                AS divyang,
			%s                                                AS disability_type,
			%s                                                AS disability_pct,
			%s                                                AS marital_status_raw
		FROM FAMILY_MEMBER fm
		LEFT JOIN FAMILY f ON f.FAMILY_ID = fm.EXTERNAL_FAMILY_ID
	`, dobExpr, educExpr, workExpr, sanitExpr,
		pursueExpr, stdExpr, instExpr, divyangExpr, disabilityExpr, disabilityPctExpr,
		maritalStatusExpr) // Add this
}

// GetUnifiedRegistry handles GET /unified-registry.
func (h *UnifiedRegistryHandler) GetUnifiedRegistry(c *gin.Context) {
	query := h.query

	rows, err := h.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch unified registry", "detail": err.Error()})
		return
	}
	defer rows.Close()

	out := make([]UnifiedRecord, 0)
	childrenByFamily := map[int]int{}
	for rows.Next() {
		var (
			firstName        string
			lastName         string
			gender           string
			dobRaw           string
			education        string
			occupation       string
			familyID         int
			totalLand        string
			ownAgriLand      string
			waterSource      string
			kharifCrop       string
			rabiCrop         string
			annualIncome     string
			sanitation       string
			children         int
			pursuing         string
			gradeStd         string
			schoolName       string
			divyang          string
			disabilityType   string
			maritalStatusRaw string // New variable
			disabilityPct    string
		)

		if err := rows.Scan(
			&firstName, &lastName, &gender, &dobRaw,
			&education, &occupation, &familyID,
			&totalLand, &ownAgriLand, &waterSource,
			&kharifCrop, &rabiCrop, &annualIncome,
			&sanitation, &children,
			&pursuing, &gradeStd, &schoolName,
			&divyang, &disabilityType, &disabilityPct,
			&maritalStatusRaw,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan unified registry record", "detail": err.Error(), "query": h.query})
			return
		}

		ageVal := parseDOBAge(dobRaw)
		if ageVal >= 0 && ageVal < 18 {
			childrenByFamily[familyID]++
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

		// ── Scholarship: Yes if actively studying ─────────────────────────────
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

		// ── Source of Income: keyword-derived from occupation ─────────────────
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

			MaritalStatus:  strings.TrimSpace(maritalStatusRaw), // Populate new field
			SourceOfIncome: sourceOfIncome,
		})
	}

	for i := range out {
		out[i].ChildrenCount = childrenByFamily[out[i].FamilyID]
	}

	c.JSON(http.StatusOK, out)
}

func parseDOBAge(raw string) int {
	v := strings.TrimSpace(raw)
	if v == "" {
		return 0
	}
	layouts := []string{"2006-01-02", "02-01-2006", "02/01/2006", "2006/01/02"}
	var dob time.Time
	parsed := false
	for _, layout := range layouts {
		t, err := time.Parse(layout, v)
		if err == nil {
			dob = t
			parsed = true
			break
		}
	}
	if !parsed {
		return 0
	}
	now := time.Now()
	age := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		age--
	}
	if age < 0 {
		return 0
	}
	return age
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
