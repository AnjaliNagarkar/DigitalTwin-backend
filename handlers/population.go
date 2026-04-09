package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type PopulationHandler struct {
	DB *sql.DB
}

type PopulationDemographicsResponse struct {
	GenderDistribution map[string]int `json:"gender_distribution"`
	AgeDistribution    map[string]int `json:"age_distribution"`
}

type PopulationEducationResponse struct {
	LiteratePopulation        int            `json:"literate_population"`
	IlliteratePopulation      int            `json:"illiterate_population"`
	StudentsCount             int            `json:"students_count"`
	DropoutCount              int            `json:"dropout_count"`
	GraduatePopulation        int            `json:"graduate_population"`
	LiteracyRate              float64        `json:"literacy_rate"`
	QualificationDistribution map[string]int `json:"qualification_distribution"`
}

type PopulationEmploymentResponse struct {
	EmployedMembers        int            `json:"employed_members"`
	UnemployedMembers      int            `json:"unemployed_members"`
	DailyWageWorkers       int            `json:"daily_wage_workers"`
	SkilledWorkers         int            `json:"skilled_workers"`
	OccupationDistribution map[string]int `json:"occupation_distribution"`
}

type PopulationMapMarker struct {
	ExternalFamilyID        string  `json:"external_family_id"`
	HouseNo                 string  `json:"house_no"`
	HeadName                string  `json:"head_name"`
	Lat                     float64 `json:"lat"`
	Lng                     float64 `json:"lng"`
	TotalMembers            int     `json:"total_members"`
	MaleMembers             int     `json:"male_members"`
	FemaleMembers           int     `json:"female_members"`
	HasDisability           int     `json:"has_disability"`
	FamilyBelongBPLCategory string  `json:"FAMILY_BELONG_BPL_CATEGORY"`
	RationCardType          string  `json:"RATION_CARD_TYPE"`
	AnnualIncome            string  `json:"ANNUAL_INCOME"`
}

type PopulationMapInsightsResponse struct {
	BPLDistribution struct {
		BPL             int `json:"bpl"`
		NonBPL          int `json:"non_bpl"`
		TotalHouseholds int `json:"total_households"`
	} `json:"bpl_distribution"`
	EducationStatus struct {
		Literate   int `json:"literate"`
		Illiterate int `json:"illiterate"`
		Students   int `json:"students"`
		Dropouts   int `json:"dropouts"`
	} `json:"education_status"`
	WorkingVsDependent struct {
		Working         int `json:"working"`
		Dependent       int `json:"dependent"`
		TotalPopulation int `json:"total_population"`
	} `json:"working_vs_dependent"`
}

type PopulationMapSummaryResponse struct {
	TotalHouseholds int `json:"total_households"`
}

// GetPopulationDashboard handles GET /population/dashboard.
// It returns top-card population metrics using SELECT-only queries.
func (h *PopulationHandler) GetPopulationDashboard(c *gin.Context) {
	log.Println("[SELECT] GET /population/dashboard")

	var totalPopulation int
	var totalHouseholds int
	var workingPopulation int
	var dependentPopulation int

	h.DB.QueryRow("SELECT COUNT(*) FROM FAMILY_MEMBER").Scan(&totalPopulation)
	h.DB.QueryRow("SELECT COUNT(*) FROM FAMILY").Scan(&totalHouseholds)
	h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER
		WHERE OCCUPATION IS NOT NULL
		  AND TRIM(OCCUPATION) != ''
		  AND UPPER(TRIM(OCCUPATION)) NOT IN (
			'HOUSEWIFE',
			'STUDYING',
			'STUDENT',
			'UNEMPLOYED',
			'NOT APPLICABLE',
			'NA'
		  )
	`).Scan(&workingPopulation)
	h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER
		WHERE (
			STR_TO_DATE(DOB, '%Y-%m-%d') IS NOT NULL
			AND (
				TIMESTAMPDIFF(YEAR, STR_TO_DATE(DOB, '%Y-%m-%d'), CURDATE()) < 18
				OR TIMESTAMPDIFF(YEAR, STR_TO_DATE(DOB, '%Y-%m-%d'), CURDATE()) > 60
			)
		  )
		   OR (
			STR_TO_DATE(DOB, '%Y-%m-%d') IS NOT NULL
			AND TIMESTAMPDIFF(YEAR, STR_TO_DATE(DOB, '%Y-%m-%d'), CURDATE()) BETWEEN 18 AND 60
			AND UPPER(TRIM(COALESCE(OCCUPATION, ''))) IN (
				'HOUSEWIFE',
				'STUDYING',
				'STUDENT',
				'UNEMPLOYED',
				'NOT APPLICABLE'
			)
		  )
	`).Scan(&dependentPopulation)

	result := gin.H{
		"total_population":     totalPopulation,
		"total_households":     totalHouseholds,
		"working_population":   workingPopulation,
		"dependent_population": dependentPopulation,
	}

	c.JSON(http.StatusOK, result)
}

// GetPopulationDemographics handles GET /population/demographics.
// It returns gender and age bucket distributions from FAMILY_MEMBER.
func (h *PopulationHandler) GetPopulationDemographics(c *gin.Context) {
	log.Println("[SELECT] GET /population/demographics")

	var male int
	var female int
	var other int

	h.DB.QueryRow(`
		SELECT
			SUM(CASE WHEN LOWER(TRIM(GENDER)) = 'male' THEN 1 ELSE 0 END) AS male,
			SUM(CASE WHEN LOWER(TRIM(GENDER)) = 'female' THEN 1 ELSE 0 END) AS female,
			SUM(CASE WHEN LOWER(TRIM(COALESCE(GENDER, ''))) NOT IN ('male', 'female') THEN 1 ELSE 0 END) AS other
		FROM FAMILY_MEMBER
	`).Scan(&male, &female, &other)

	var age0To5 int
	var age6To18 int
	var age19To35 int
	var age36To60 int
	var age60Plus int

	h.DB.QueryRow(`
		SELECT
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(DOB, '%d-%m-%Y'), CURDATE()) BETWEEN 0 AND 5 THEN 1 ELSE 0 END) AS age_0_5,
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(DOB, '%d-%m-%Y'), CURDATE()) BETWEEN 6 AND 18 THEN 1 ELSE 0 END) AS age_6_18,
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(DOB, '%d-%m-%Y'), CURDATE()) BETWEEN 19 AND 35 THEN 1 ELSE 0 END) AS age_19_35,
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(DOB, '%d-%m-%Y'), CURDATE()) BETWEEN 36 AND 60 THEN 1 ELSE 0 END) AS age_36_60,
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(DOB, '%d-%m-%Y'), CURDATE()) > 60 THEN 1 ELSE 0 END) AS age_60_plus
		FROM FAMILY_MEMBER
	`).Scan(&age0To5, &age6To18, &age19To35, &age36To60, &age60Plus)

	response := PopulationDemographicsResponse{
		GenderDistribution: map[string]int{
			"male":   male,
			"female": female,
			"other":  other,
		},
		AgeDistribution: map[string]int{
			"age_0_5":     age0To5,
			"age_6_18":    age6To18,
			"age_19_35":   age19To35,
			"age_36_60":   age36To60,
			"age_60_plus": age60Plus,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetPopulationEducation handles GET /population/education.
// It returns education intelligence metrics and qualification breakdown.
func (h *PopulationHandler) GetPopulationEducation(c *gin.Context) {
	log.Println("[SELECT] GET /population/education")

	var totalPopulation int
	var literatePopulation int
	var illiteratePopulation int
	var studentsCount int
	var dropoutCount int
	var graduatePopulation int

	h.DB.QueryRow("SELECT COUNT(*) FROM FAMILY_MEMBER").Scan(&totalPopulation)
	h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER
		WHERE UPPER(TRIM(COALESCE(EVER_ATTENDED_SCHOOL, ''))) = 'YES'
	`).Scan(&literatePopulation)
	h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER
		WHERE UPPER(TRIM(COALESCE(EVER_ATTENDED_SCHOOL, ''))) = 'NO'
		   OR EVER_ATTENDED_SCHOOL IS NULL
	`).Scan(&illiteratePopulation)
	h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER
		WHERE UPPER(TRIM(COALESCE(CURRENTLY_PURSUING_EDUCATION, ''))) = 'YES'
	`).Scan(&studentsCount)
	h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER
		WHERE UPPER(TRIM(COALESCE(EVER_ATTENDED_SCHOOL, ''))) = 'YES'
		  AND UPPER(TRIM(COALESCE(CURRENTLY_PURSUING_EDUCATION, ''))) != 'YES'
		  AND DROP_OUT IS NOT NULL
		  AND TRIM(DROP_OUT) != ''
	`).Scan(&dropoutCount)
	h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER
		WHERE TRIM(COALESCE(QUALIFICATION, '')) = 'Graduation & Above'
	`).Scan(&graduatePopulation)

	var below10th int
	var tenth int
	var twelfth int
	var graduateAbove int
	h.DB.QueryRow(`
		SELECT
			SUM(CASE WHEN QUALIFICATION IS NULL OR TRIM(QUALIFICATION) = '' THEN 1 ELSE 0 END) AS below_10th,
			SUM(CASE WHEN TRIM(QUALIFICATION) = '10th' THEN 1 ELSE 0 END) AS tenth,
			SUM(CASE WHEN TRIM(QUALIFICATION) = '12th' THEN 1 ELSE 0 END) AS twelfth,
			SUM(CASE WHEN TRIM(QUALIFICATION) = 'Graduation & Above' THEN 1 ELSE 0 END) AS graduate_above
		FROM FAMILY_MEMBER
	`).Scan(&below10th, &tenth, &twelfth, &graduateAbove)

	literacyRate := 0.0
	if totalPopulation > 0 {
		literacyRate = (float64(literatePopulation) / float64(totalPopulation)) * 100
	}

	response := PopulationEducationResponse{
		LiteratePopulation:   literatePopulation,
		IlliteratePopulation: illiteratePopulation,
		StudentsCount:        studentsCount,
		DropoutCount:         dropoutCount,
		GraduatePopulation:   graduatePopulation,
		LiteracyRate:         literacyRate,
		QualificationDistribution: map[string]int{
			"below_10th":     below10th,
			"tenth":          tenth,
			"twelfth":        twelfth,
			"graduate_above": graduateAbove,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetPopulationEmployment handles GET /population/employment.
// It returns employment insight metrics and occupation distribution.
func (h *PopulationHandler) GetPopulationEmployment(c *gin.Context) {
	log.Println("[SELECT] GET /population/employment")

	var employedMembers int
	var unemployedMembers int
	var dailyWageWorkers int
	var skilledWorkers int

	h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER
		WHERE TRIM(COALESCE(OCCUPATION, '')) IN (
			'Salaried Job',
			'Self Employed - Farm based',
			'Self Employed- Non-farm based',
			'Self Employed-Agri allied',
			'Wage Work'
		)
	`).Scan(&employedMembers)

	h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER
		WHERE TRIM(COALESCE(OCCUPATION, '')) IN (
			'Unemployed',
			'Not Applicable'
		)
	`).Scan(&unemployedMembers)

	h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER
		WHERE TRIM(COALESCE(OCCUPATION, '')) = 'Wage Work'
		   OR NATURE_WAGE_WORK IS NOT NULL
	`).Scan(&dailyWageWorkers)

	h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER
		WHERE LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%driver%'
		   OR LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%electric%'
		   OR LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%mechanic%'
		   OR LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%tailor%'
		   OR LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%carpenter%'
		   OR LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%computer%'
		   OR LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%bank%'
		   OR LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%shop%'
		   OR LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%company worker%'
		   OR LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%security%'
		   OR LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%painter%'
		   OR LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%civil%'
		   OR LOWER(COALESCE(NATURE_WAGE_WORK, '')) LIKE '%technician%'
	`).Scan(&skilledWorkers)

	var farmBased int
	var agriAllied int
	var nonFarm int
	var salaried int
	var wageWorkers int
	var housewife int
	var students int
	var unemployed int
	var other int

	h.DB.QueryRow(`
		SELECT
			SUM(CASE WHEN TRIM(COALESCE(OCCUPATION, '')) = 'Self Employed - Farm based' THEN 1 ELSE 0 END) AS farm_based,
			SUM(CASE WHEN TRIM(COALESCE(OCCUPATION, '')) = 'Self Employed-Agri allied' THEN 1 ELSE 0 END) AS agri_allied,
			SUM(CASE WHEN TRIM(COALESCE(OCCUPATION, '')) = 'Self Employed- Non-farm based' THEN 1 ELSE 0 END) AS non_farm,
			SUM(CASE WHEN TRIM(COALESCE(OCCUPATION, '')) = 'Salaried Job' THEN 1 ELSE 0 END) AS salaried,
			SUM(CASE WHEN TRIM(COALESCE(OCCUPATION, '')) = 'Wage Work' THEN 1 ELSE 0 END) AS wage_workers,
			SUM(CASE WHEN TRIM(COALESCE(OCCUPATION, '')) = 'Housewife' THEN 1 ELSE 0 END) AS housewife,
			SUM(CASE WHEN TRIM(COALESCE(OCCUPATION, '')) = 'Studying' THEN 1 ELSE 0 END) AS students,
			SUM(CASE WHEN TRIM(COALESCE(OCCUPATION, '')) = 'Unemployed' THEN 1 ELSE 0 END) AS unemployed,
			SUM(CASE WHEN OCCUPATION IS NULL OR TRIM(COALESCE(OCCUPATION, '')) = '' THEN 1 ELSE 0 END) AS other
		FROM FAMILY_MEMBER
	`).Scan(&farmBased, &agriAllied, &nonFarm, &salaried, &wageWorkers, &housewife, &students, &unemployed, &other)

	response := PopulationEmploymentResponse{
		EmployedMembers:   employedMembers,
		UnemployedMembers: unemployedMembers,
		DailyWageWorkers:  dailyWageWorkers,
		SkilledWorkers:    skilledWorkers,
		OccupationDistribution: map[string]int{
			"farm_based":   farmBased,
			"agri_allied":  agriAllied,
			"non_farm":     nonFarm,
			"salaried":     salaried,
			"wage_workers": wageWorkers,
			"housewife":    housewife,
			"students":     students,
			"unemployed":   unemployed,
			"other":        other,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *PopulationHandler) familyColumnExists(column string) bool {
	var count int
	if err := h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'FAMILY'
		  AND COLUMN_NAME = ?
	`, column).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func (h *PopulationHandler) buildPopulationFamilyFilters(alias string, c *gin.Context) (string, []interface{}) {
	clauses := []string{"1=1"}
	args := []interface{}{}

	stateID := strings.TrimSpace(c.Query("state_id"))
	districtID := strings.TrimSpace(c.Query("district_id"))
	talukaID := strings.TrimSpace(c.Query("taluka_id"))
	villageID := strings.TrimSpace(c.Query("village_id"))

	if stateID != "" && h.familyColumnExists("STATE_ID") {
		clauses = append(clauses, fmt.Sprintf("CAST(%s.STATE_ID AS CHAR) = ?", alias))
		args = append(args, stateID)
	}
	if districtID != "" {
		clauses = append(clauses, fmt.Sprintf("CAST(%s.DISTRICT_ID AS CHAR) = ?", alias))
		args = append(args, districtID)
	}
	if talukaID != "" {
		clauses = append(clauses, fmt.Sprintf("CAST(%s.TALUKA_ID AS CHAR) = ?", alias))
		args = append(args, talukaID)
	}
	if villageID != "" {
		clauses = append(clauses, fmt.Sprintf("CAST(%s.VILLAGE_ID AS CHAR) = ?", alias))
		args = append(args, villageID)
	}

	return strings.Join(clauses, " AND "), args
}

// GetPopulationMapData handles GET /population/map-data.
// It returns household markers and total member counts for the population map.
func (h *PopulationHandler) GetPopulationMapData(c *gin.Context) {
	log.Println("[SELECT] GET /population/map-data")
	colorBy := strings.TrimSpace(c.Query("color_by"))
	if strings.EqualFold(colorBy, "bpl") {
		log.Println("[SELECT] population map color mode: bpl")
	}
	if strings.EqualFold(colorBy, "divyang") {
		log.Println("[SELECT] population map color mode: divyang")
	}

	where, args := h.buildPopulationFamilyFilters("f", c)
	where = fmt.Sprintf("WHERE f.LATITUDE IS NOT NULL AND f.LONGITUDE IS NOT NULL AND f.LATITUDE != 0 AND f.LONGITUDE != 0 AND %s", where)

	query := fmt.Sprintf(`
		SELECT
			COALESCE(CAST(f.EXTERNAL_FAMILY_ID AS CHAR), '') AS external_family_id,
			COALESCE(CAST(f.HOUSE_NO AS CHAR), '') AS house_no,
			COALESCE(TRIM(CONCAT(
				COALESCE(f.FIRST_NAME_HOUSEHOLD_HEAD, ''), ' ',
				COALESCE(f.MIDDLE_NAME_HOUSEHOLD_HEAD, ''), ' ',
				COALESCE(f.LAST_NAME_HOUSEHOLD_HEAD, '')
			)), '') AS head_name,
			f.LATITUDE AS lat,
			f.LONGITUDE AS lng,
			MAX(CASE
				WHEN UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES'
					OR NULLIF(TRIM(COALESCE(fm.DISABILITY, '')), '') IS NOT NULL
					OR NULLIF(TRIM(COALESCE(CAST(fm.DISABILITY_PERCENTAGE AS CHAR), '')), '') IS NOT NULL
				THEN 1
				ELSE 0
			END) AS has_disability,
			COALESCE(TRIM(COALESCE(f.FAMILY_BELONG_BPL_CATEGORY, '')), '') AS FAMILY_BELONG_BPL_CATEGORY,
			COALESCE(TRIM(COALESCE(f.RATION_CARD_TYPE, '')), '') AS RATION_CARD_TYPE,
			COALESCE(TRIM(CAST(f.ANNUAL_INCOME AS CHAR)), '') AS ANNUAL_INCOME,
			COUNT(fm.FAMILY_MEMBER_ID) AS total_members,
			SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER, ''))) = 'male' THEN 1 ELSE 0 END) AS male_members,
			SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER, ''))) = 'female' THEN 1 ELSE 0 END) AS female_members
		FROM FAMILY f
		LEFT JOIN FAMILY_MEMBER fm ON fm.EXTERNAL_FAMILY_ID = f.EXTERNAL_FAMILY_ID
		%s
		GROUP BY
			f.EXTERNAL_FAMILY_ID,
			f.HOUSE_NO,
			f.FIRST_NAME_HOUSEHOLD_HEAD,
			f.MIDDLE_NAME_HOUSEHOLD_HEAD,
			f.LAST_NAME_HOUSEHOLD_HEAD,
			f.LATITUDE,
			f.LONGITUDE,
			f.FAMILY_BELONG_BPL_CATEGORY,
			f.RATION_CARD_TYPE,
			f.ANNUAL_INCOME
		ORDER BY f.HOUSE_NO, f.EXTERNAL_FAMILY_ID
	`, where)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch population map data", "detail": err.Error()})
		return
	}
	defer rows.Close()

	markers := []PopulationMapMarker{}
	for rows.Next() {
		var marker PopulationMapMarker
		if err := rows.Scan(
			&marker.ExternalFamilyID,
			&marker.HouseNo,
			&marker.HeadName,
			&marker.Lat,
			&marker.Lng,
			&marker.HasDisability,
			&marker.FamilyBelongBPLCategory,
			&marker.RationCardType,
			&marker.AnnualIncome,
			&marker.TotalMembers,
			&marker.MaleMembers,
			&marker.FemaleMembers,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan population map marker", "detail": err.Error()})
			return
		}
		markers = append(markers, marker)
	}

	if markers == nil {
		markers = []PopulationMapMarker{}
	}

	c.JSON(http.StatusOK, markers)
}

// GetPopulationMapSummary handles GET /population/map-summary.
// It returns the filtered household count with valid map coordinates.
func (h *PopulationHandler) GetPopulationMapSummary(c *gin.Context) {
	log.Println("[SELECT] GET /population/map-summary")

	where, args := h.buildPopulationFamilyFilters("f", c)

	var totalHouseholds int
	err := h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(DISTINCT f.EXTERNAL_FAMILY_ID)
		FROM FAMILY f
		WHERE f.LATITUDE IS NOT NULL
		  AND f.LONGITUDE IS NOT NULL
		  AND f.LATITUDE != 0
		  AND f.LONGITUDE != 0
		  AND %s
	`, where), args...).Scan(&totalHouseholds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch population map summary", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, PopulationMapSummaryResponse{TotalHouseholds: totalHouseholds})
}

// GetPopulationMapInsights handles GET /population/map-insights.
// It returns BPL, education, and working/dependent summaries for the population map.
func (h *PopulationHandler) GetPopulationMapInsights(c *gin.Context) {
	log.Println("[SELECT] GET /population/map-insights")

	where, args := h.buildPopulationFamilyFilters("f", c)

	var totalHouseholds int
	var bplHouseholds int
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY f
		WHERE %s
	`, where), args...).Scan(&totalHouseholds)

	bplConditions := []string{}
	if h.familyColumnExists("FAMILY_BELONG_BPL_CATEGORY") {
		bplConditions = append(bplConditions, "UPPER(TRIM(COALESCE(f.FAMILY_BELONG_BPL_CATEGORY, ''))) = 'YES'")
	}
	if h.familyColumnExists("RATION_CARD_TYPE") {
		bplConditions = append(bplConditions, "UPPER(TRIM(COALESCE(f.RATION_CARD_TYPE, ''))) IN ('BPL', 'AAY')")
	}
	if len(bplConditions) > 0 {
		bplQuery := fmt.Sprintf(`
			SELECT COUNT(*)
			FROM FAMILY f
			WHERE %s
			  AND (%s)
		`, where, strings.Join(bplConditions, " OR "))
		h.DB.QueryRow(bplQuery, args...).Scan(&bplHouseholds)
	}

	var literate int
	var illiterate int
	var students int
	var dropouts int
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE %s
		  AND UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'YES'
	`, where), args...).Scan(&literate)
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE %s
		  AND (UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'NO' OR fm.EVER_ATTENDED_SCHOOL IS NULL)
	`, where), args...).Scan(&illiterate)
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE %s
		  AND UPPER(TRIM(COALESCE(fm.CURRENTLY_PURSUING_EDUCATION, ''))) = 'YES'
	`, where), args...).Scan(&students)
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE %s
		  AND (
			UPPER(TRIM(COALESCE(fm.DROP_OUT, ''))) = 'YES'
			OR TRIM(COALESCE(fm.DROP_OUT, '')) IN ('1','2','3','4','5','6','7','8','9','10')
		  )
	`, where), args...).Scan(&dropouts)

	var working int
	var dependent int
	var totalPopulation int
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE %s
	`, where), args...).Scan(&totalPopulation)
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE %s
		  AND UPPER(TRIM(COALESCE(fm.OCCUPATION, ''))) IN (
			'SELF EMPLOYED - FARM BASED',
			'SELF EMPLOYED- NON-FARM BASED',
			'SELF EMPLOYED-AGRI ALLIED',
			'WAGE WORK',
			'SALARIED JOB'
		  )
	`, where), args...).Scan(&working)
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE %s
		  AND (
			(
				STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y') IS NOT NULL
				AND (
					TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) < 18
					OR TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) > 60
				)
			)
			OR (
				STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y') IS NOT NULL
				AND TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) BETWEEN 18 AND 60
				AND UPPER(TRIM(COALESCE(fm.OCCUPATION, ''))) IN (
					'STUDENT',
					'STUDYING',
					'HOUSEWIFE',
					'UNEMPLOYED',
					'NOT APPLICABLE'
				)
			)
		  )
	`, where), args...).Scan(&dependent)

	response := PopulationMapInsightsResponse{}
	response.BPLDistribution.BPL = bplHouseholds
	response.BPLDistribution.NonBPL = totalHouseholds - bplHouseholds
	response.BPLDistribution.TotalHouseholds = totalHouseholds
	response.EducationStatus.Literate = literate
	response.EducationStatus.Illiterate = illiterate
	response.EducationStatus.Students = students
	response.EducationStatus.Dropouts = dropouts
	response.WorkingVsDependent.Working = working
	response.WorkingVsDependent.Dependent = dependent
	response.WorkingVsDependent.TotalPopulation = totalPopulation

	c.JSON(http.StatusOK, response)
}
