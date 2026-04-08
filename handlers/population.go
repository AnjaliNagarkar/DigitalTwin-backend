package handlers

import (
	"database/sql"
	"log"
	"net/http"

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
