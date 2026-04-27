package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type PopulationHandler struct {
	DB *sql.DB
}

type PopulationDemographicsResponse struct {
	GenderDistribution          map[string]int               `json:"gender_distribution"`
	AgeDistribution             map[string]int               `json:"age_distribution"`
	AgeIncomeGenderDistribution []AgeIncomeGenderItem        `json:"age_income_gender_distribution"`
	TotalDivyang                int                          `json:"total_divyang"`
	DisabilityDistribution      []DisabilityDistributionItem `json:"disability_distribution"`
}

type AgeIncomeGenderItem struct {
	AgeRange      string  `json:"age_range"`
	AverageIncome float64 `json:"average_income"`
	TotalIncome   float64 `json:"total_income"`
	MaleCount     int     `json:"male_count"`
	FemaleCount   int     `json:"female_count"`
}

type DisabilityDistributionItem struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
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
	ExternalFamilyID        string   `json:"external_family_id"`
	HouseNo                 string   `json:"house_no"`
	HeadName                string   `json:"head_name"`
	VillageID               string   `json:"village_id"`
	VillageName             string   `json:"village_name"`
	Lat                     float64  `json:"lat"`
	Lng                     float64  `json:"lng"`
	TotalMembers            int      `json:"total_members"`
	LiterateMembers         int      `json:"literate_members"`
	IlliterateMembers       int      `json:"illiterate_members"`
	VillageWorkingMembers   int      `json:"village_working_members"`
	UnemployedMembers       int      `json:"unemployed_members"`
	DivyangMembers          int      `json:"divyang_members"`
	MaleMembers             int      `json:"male_members"`
	FemaleMembers           int      `json:"female_members"`
	WorkingMembers          int      `json:"working_members"`
	NonWorkingMembers       int      `json:"non_working_members"`
	WorkingOccupations      string   `json:"working_occupations"`
	OccupationListRaw       string   `json:"occupation_list"`
	OccupationList          []string `json:"occupation_list_array"`
	HasDisability           int      `json:"has_disability"`
	FamilyBelongBPLCategory string   `json:"FAMILY_BELONG_BPL_CATEGORY"`
	RationCardType          string   `json:"RATION_CARD_TYPE"`
	AnnualIncome            string   `json:"ANNUAL_INCOME"`
	LocationQuality         string   `json:"location_quality"`
	LocationReason          string   `json:"location_reason"`
	DuplicateCount          int      `json:"duplicate_count,omitempty"`
	DuplicateHouses         []string `json:"duplicate_houses,omitempty"`
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
	_, cancel, ok := ensureDBReady(c, h.DB, "/population/dashboard")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection unavailable"})
		return
	}
	defer cancel()
	districtID := strings.TrimSpace(c.Query("district_id"))
	talukaID := strings.TrimSpace(c.Query("taluka_id"))
	villageID := strings.TrimSpace(c.Query("village_id"))
	log.Printf("[FILTER] /population/dashboard district_id=%q taluka_id=%q village_id=%q", districtID, talukaID, villageID)
	where, args := h.buildPopulationFamilyFilters("f", c)

	var totalPopulation int
	var totalHouseholds int
	var workingPopulation int
	var dependentPopulation int

	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE %s
	`, where), args...).Scan(&totalPopulation)
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY f
		WHERE %s
	`, where), args...).Scan(&totalHouseholds)
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE fm.OCCUPATION IS NOT NULL
		  AND TRIM(fm.OCCUPATION) != ''
		  AND UPPER(TRIM(fm.OCCUPATION)) NOT IN (
			'HOUSEWIFE',
			'STUDYING',
			'STUDENT',
			'UNEMPLOYED',
			'NOT APPLICABLE',
			'NA'
		  )
		  AND %s
	`, where), args...).Scan(&workingPopulation)
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE (
			STR_TO_DATE(fm.DOB, '%%Y-%%m-%%d') IS NOT NULL
			AND (
				TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%Y-%%m-%%d'), CURDATE()) < 18
				OR TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%Y-%%m-%%d'), CURDATE()) > 60
			)
		  )
		   OR (
			STR_TO_DATE(fm.DOB, '%%Y-%%m-%%d') IS NOT NULL
			AND TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%Y-%%m-%%d'), CURDATE()) BETWEEN 18 AND 60
			AND UPPER(TRIM(COALESCE(fm.OCCUPATION, ''))) IN (
				'HOUSEWIFE',
				'STUDYING',
				'STUDENT',
				'UNEMPLOYED',
				'NOT APPLICABLE'
			)
		  ))
		  AND %s
	`, where), args...).Scan(&dependentPopulation)

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
	_, cancel, ok := ensureDBReady(c, h.DB, "/population/demographics")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection unavailable"})
		return
	}
	defer cancel()
	districtID := strings.TrimSpace(c.Query("district_id"))
	talukaID := strings.TrimSpace(c.Query("taluka_id"))
	villageID := strings.TrimSpace(c.Query("village_id"))
	log.Printf("[FILTER] /population/demographics district_id=%q taluka_id=%q village_id=%q", districtID, talukaID, villageID)
	where, args := h.buildPopulationFamilyFilters("f", c)

	var male int
	var female int
	var other int

	h.DB.QueryRow(fmt.Sprintf(`
		SELECT
			SUM(CASE WHEN LOWER(TRIM(fm.GENDER)) = 'male' THEN 1 ELSE 0 END) AS male,
			SUM(CASE WHEN LOWER(TRIM(fm.GENDER)) = 'female' THEN 1 ELSE 0 END) AS female,
			SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER, ''))) NOT IN ('male', 'female') THEN 1 ELSE 0 END) AS other
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE %s
	`, where), args...).Scan(&male, &female, &other)

	var age0To5 int
	var age6To18 int
	var age19To35 int
	var age36To60 int
	var age60Plus int

	h.DB.QueryRow(fmt.Sprintf(`
		SELECT
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) BETWEEN 0 AND 5 THEN 1 ELSE 0 END) AS age_0_5,
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) BETWEEN 6 AND 18 THEN 1 ELSE 0 END) AS age_6_18,
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) BETWEEN 19 AND 35 THEN 1 ELSE 0 END) AS age_19_35,
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) BETWEEN 36 AND 60 THEN 1 ELSE 0 END) AS age_36_60,
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) > 60 THEN 1 ELSE 0 END) AS age_60_plus
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE %s
	`, where), args...).Scan(&age0To5, &age6To18, &age19To35, &age36To60, &age60Plus)

	var totalDivyang int
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) AS total_divyang
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES'
		  AND %s
	`, where), args...).Scan(&totalDivyang)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	ageGroupOrder := []string{"0-18", "19-30", "31-45", "46-60", "60+"}
	ageIncomeGenderDistribution := make([]AgeIncomeGenderItem, 0, len(ageGroupOrder))
	ageIncomeGenderMap := make(map[string]AgeIncomeGenderItem, len(ageGroupOrder))

	ageGenderQuery := `
		SELECT
			CASE
				WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%d-%m-%Y'), CURDATE()) <= 18 THEN '0-18'
				WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%d-%m-%Y'), CURDATE()) BETWEEN 19 AND 30 THEN '19-30'
				WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%d-%m-%Y'), CURDATE()) BETWEEN 31 AND 45 THEN '31-45'
				WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%d-%m-%Y'), CURDATE()) BETWEEN 46 AND 60 THEN '46-60'
				ELSE '60+'
			END AS age_group,
			SUM(CASE WHEN LOWER(TRIM(fm.GENDER)) IN ('male','m') THEN 1 ELSE 0 END) AS male,
			SUM(CASE WHEN LOWER(TRIM(fm.GENDER)) IN ('female','f') THEN 1 ELSE 0 END) AS female,
			SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER, ''))) NOT IN ('male', 'female', 'm', 'f') THEN 1 ELSE 0 END) AS other,
			SUM(IFNULL(f.ANNUAL_INCOME, 0)) AS total_income
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON fm.EXTERNAL_FAMILY_ID = f.FAMILY_ID
		WHERE fm.DOB IS NOT NULL AND fm.DOB != ''
		  AND STR_TO_DATE(fm.DOB, '%d-%m-%Y') IS NOT NULL
		GROUP BY age_group
	`
	log.Println("AGE-GENDER QUERY:", ageGenderQuery)

	ageIncomeGenderRows, err := h.DB.QueryContext(ctx, ageGenderQuery)
	if err != nil {
		log.Println("AGE-GENDER QUERY FAILED:", err)
	} else {
		defer ageIncomeGenderRows.Close()

		for ageIncomeGenderRows.Next() {
			var item AgeIncomeGenderItem
			var other int
			var totalIncome sql.NullFloat64
			if scanErr := ageIncomeGenderRows.Scan(&item.AgeRange, &item.MaleCount, &item.FemaleCount, &other, &totalIncome); scanErr != nil {
				log.Println("SCAN ERROR:", scanErr)
				ageIncomeGenderMap = map[string]AgeIncomeGenderItem{}
				break
			}
			item.AverageIncome = 0
			if totalIncome.Valid {
				item.TotalIncome = totalIncome.Float64
			}
			ageIncomeGenderMap[item.AgeRange] = item
		}
		if rowsErr := ageIncomeGenderRows.Err(); rowsErr != nil {
			log.Println("ROWS ERROR:", rowsErr)
		}
	}

	for _, ageGroup := range ageGroupOrder {
		if item, ok := ageIncomeGenderMap[ageGroup]; ok {
			ageIncomeGenderDistribution = append(ageIncomeGenderDistribution, item)
			continue
		}
		ageIncomeGenderDistribution = append(ageIncomeGenderDistribution, AgeIncomeGenderItem{AgeRange: ageGroup})
	}

	disabilityRows, err := h.DB.Query(fmt.Sprintf(`
		SELECT
			CASE
				WHEN DISABILITY_CATEGORY LIKE '%%Blind%%'
					OR DISABILITY_CATEGORY LIKE '%%Low vision%%'
				THEN 'Visual Disability'

				WHEN DISABILITY_CATEGORY LIKE '%%Locomotor%%'
					OR DISABILITY_CATEGORY LIKE '%%Cerebral%%'
					OR DISABILITY_CATEGORY LIKE '%%Muscular%%'
					OR DISABILITY_CATEGORY LIKE '%%Dwarf%%'
				THEN 'Locomotor Disability'

				WHEN DISABILITY_CATEGORY LIKE '%%Mental%%'
					OR DISABILITY_CATEGORY LIKE '%%Autism%%'
					OR DISABILITY_CATEGORY LIKE '%%Intellectual%%'
					OR DISABILITY_CATEGORY LIKE '%%Learning%%'
				THEN 'Intellectual Disability'

				WHEN DISABILITY_CATEGORY LIKE '%%Hearing%%'
				THEN 'Hearing Disability'

				WHEN DISABILITY_CATEGORY LIKE '%%Speech%%'
				THEN 'Speech Disability'

				WHEN DISABILITY_CATEGORY LIKE '%%Multiple%%'
				THEN 'Multiple Disabilities'

				WHEN DISABILITY_CATEGORY LIKE '%%Parkinson%%'
					OR DISABILITY_CATEGORY LIKE '%%Sclerosis%%'
					OR DISABILITY_CATEGORY LIKE '%%Sickle%%'
					OR DISABILITY_CATEGORY LIKE '%%Thalassemia%%'
				THEN 'Chronic Conditions'

				ELSE 'Other'
			END AS disability_group,
			COUNT(*) AS total
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON fm.EXTERNAL_FAMILY_ID = f.EXTERNAL_FAMILY_ID
		WHERE UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES'
		  AND %s
		GROUP BY disability_group
		ORDER BY total DESC
	`, where), args...)
	if err != nil {
		log.Printf("demographics disability_distribution query failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch disability distribution"})
		return
	}
	defer disabilityRows.Close()

	disabilityDistribution := make([]DisabilityDistributionItem, 0)
	for disabilityRows.Next() {
		var item DisabilityDistributionItem
		if err := disabilityRows.Scan(&item.Name, &item.Value); err != nil {
			log.Printf("demographics disability_distribution scan failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse disability distribution"})
			return
		}
		disabilityDistribution = append(disabilityDistribution, item)
	}
	if err := disabilityRows.Err(); err != nil {
		log.Printf("demographics disability_distribution rows error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read disability distribution"})
		return
	}

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
		AgeIncomeGenderDistribution: ageIncomeGenderDistribution,
		TotalDivyang:                totalDivyang,
		DisabilityDistribution:      disabilityDistribution,
	}

	c.JSON(http.StatusOK, response)
}

// GetPopulationEducation handles GET /population/education.
// It returns education intelligence metrics and qualification breakdown.
func (h *PopulationHandler) GetPopulationEducation(c *gin.Context) {
	log.Println("[SELECT] GET /population/education")
	_, cancel, ok := ensureDBReady(c, h.DB, "/population/education")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection unavailable"})
		return
	}
	defer cancel()
	districtID := strings.TrimSpace(c.Query("district_id"))
	talukaID := strings.TrimSpace(c.Query("taluka_id"))
	villageID := strings.TrimSpace(c.Query("village_id"))
	log.Printf("[FILTER] /population/education district_id=%q taluka_id=%q village_id=%q", districtID, talukaID, villageID)
	where, args := h.buildPopulationFamilyFilters("f", c)

	var totalPopulation int
	var literatePopulation int
	var illiteratePopulation int
	var studentsCount int
	var dropoutCount int
	var graduatePopulation int

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
		WHERE UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'YES'
		  AND %s
	`, where), args...).Scan(&literatePopulation)
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE (UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'NO'
		   OR fm.EVER_ATTENDED_SCHOOL IS NULL)
		  AND %s
	`, where), args...).Scan(&illiteratePopulation)
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE UPPER(TRIM(COALESCE(fm.CURRENTLY_PURSUING_EDUCATION, ''))) = 'YES'
		  AND %s
	`, where), args...).Scan(&studentsCount)
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'YES'
		  AND UPPER(TRIM(COALESCE(fm.CURRENTLY_PURSUING_EDUCATION, ''))) != 'YES'
		  AND fm.DROP_OUT IS NOT NULL
		  AND TRIM(fm.DROP_OUT) != ''
		  AND %s
	`, where), args...).Scan(&dropoutCount)
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE TRIM(COALESCE(fm.QUALIFICATION, '')) = 'Graduation & Above'
		  AND %s
	`, where), args...).Scan(&graduatePopulation)

	var below10th int
	var tenth int
	var twelfth int
	var graduateAbove int
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT
			SUM(CASE WHEN fm.QUALIFICATION IS NULL OR TRIM(fm.QUALIFICATION) = '' THEN 1 ELSE 0 END) AS below_10th,
			SUM(CASE WHEN TRIM(fm.QUALIFICATION) = '10th' THEN 1 ELSE 0 END) AS tenth,
			SUM(CASE WHEN TRIM(fm.QUALIFICATION) = '12th' THEN 1 ELSE 0 END) AS twelfth,
			SUM(CASE WHEN TRIM(fm.QUALIFICATION) = 'Graduation & Above' THEN 1 ELSE 0 END) AS graduate_above
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE %s
	`, where), args...).Scan(&below10th, &tenth, &twelfth, &graduateAbove)

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
	_, cancel, ok := ensureDBReady(c, h.DB, "/population/employment")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection unavailable"})
		return
	}
	defer cancel()
	districtID := strings.TrimSpace(c.Query("district_id"))
	talukaID := strings.TrimSpace(c.Query("taluka_id"))
	villageID := strings.TrimSpace(c.Query("village_id"))
	log.Printf("[FILTER] /population/employment district_id=%q taluka_id=%q village_id=%q", districtID, talukaID, villageID)
	where, args := h.buildPopulationFamilyFilters("f", c)

	var employedMembers int
	var unemployedMembers int
	var dailyWageWorkers int
	var skilledWorkers int

	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE TRIM(COALESCE(fm.OCCUPATION, '')) IN (
			'Salaried Job',
			'Self Employed - Farm based',
			'Self Employed- Non-farm based',
			'Self Employed-Agri allied',
			'Wage Work'
		)
		  AND %s
	`, where), args...).Scan(&employedMembers)

	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE TRIM(COALESCE(fm.OCCUPATION, '')) IN (
			'Unemployed',
			'Not Applicable'
		)
		  AND %s
	`, where), args...).Scan(&unemployedMembers)

	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE (TRIM(COALESCE(fm.OCCUPATION, '')) = 'Wage Work'
		   OR fm.NATURE_WAGE_WORK IS NOT NULL)
		  AND %s
	`, where), args...).Scan(&dailyWageWorkers)

	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE (
		   LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%driver%%'
		   OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%electric%%'
		   OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%mechanic%%'
		   OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%tailor%%'
		   OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%carpenter%%'
		   OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%computer%%'
		   OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%bank%%'
		   OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%shop%%'
		   OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%company worker%%'
		   OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%security%%'
		   OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%painter%%'
		   OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%civil%%'
		   OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%technician%%'
		)
		  AND %s
	`, where), args...).Scan(&skilledWorkers)

	var farmBased int
	var agriAllied int
	var nonFarm int
	var salaried int
	var wageWorkers int
	var housewife int
	var students int
	var unemployed int
	var other int

	h.DB.QueryRow(fmt.Sprintf(`
		SELECT
			SUM(CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Self Employed - Farm based' THEN 1 ELSE 0 END) AS farm_based,
			SUM(CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Self Employed-Agri allied' THEN 1 ELSE 0 END) AS agri_allied,
			SUM(CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Self Employed- Non-farm based' THEN 1 ELSE 0 END) AS non_farm,
			SUM(CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Salaried Job' THEN 1 ELSE 0 END) AS salaried,
			SUM(CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Wage Work' THEN 1 ELSE 0 END) AS wage_workers,
			SUM(CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Housewife' THEN 1 ELSE 0 END) AS housewife,
			SUM(CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Studying' THEN 1 ELSE 0 END) AS students,
			SUM(CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Unemployed' THEN 1 ELSE 0 END) AS unemployed,
			SUM(CASE WHEN fm.OCCUPATION IS NULL OR TRIM(COALESCE(fm.OCCUPATION, '')) = '' THEN 1 ELSE 0 END) AS other
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE %s
	`, where), args...).Scan(&farmBased, &agriAllied, &nonFarm, &salaried, &wageWorkers, &housewife, &students, &unemployed, &other)

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

func (h *PopulationHandler) populationMemberColumnExists(column string) bool {
	var count int
	if err := h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'FAMILY_MEMBER'
		  AND COLUMN_NAME = ?
	`, column).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func (h *PopulationHandler) buildPopulationFamilyFilters(alias string, c *gin.Context) (string, []interface{}) {
	clauses := []string{"1=1"}
	args := []interface{}{}

	toNullable := func(value string) interface{} {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil
		}
		return trimmed
	}

	stateID := strings.TrimSpace(c.Query("state_id"))
	districtID := toNullable(c.Query("district_id"))
	talukaID := toNullable(c.Query("taluka_id"))
	villageID := toNullable(c.Query("village_id"))

	if stateID != "" && h.familyColumnExists("STATE_ID") {
		clauses = append(clauses, fmt.Sprintf("CAST(%s.STATE_ID AS CHAR) = ?", alias))
		args = append(args, stateID)
	}

	clauses = append(clauses, fmt.Sprintf("(? IS NULL OR ? = '' OR CAST(%s.DISTRICT_ID AS CHAR) = ?)", alias))
	args = append(args, districtID, districtID, districtID)

	clauses = append(clauses, fmt.Sprintf("(? IS NULL OR ? = '' OR CAST(%s.TALUKA_ID AS CHAR) = ?)", alias))
	args = append(args, talukaID, talukaID, talukaID)

	clauses = append(clauses, fmt.Sprintf("(? IS NULL OR ? = '' OR CAST(%s.VILLAGE_ID AS CHAR) = ?)", alias))
	args = append(args, villageID, villageID, villageID)

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
	if strings.EqualFold(colorBy, "employment") {
		log.Println("[SELECT] population map color mode: employment")
	}

	where, args := h.buildPopulationFamilyFilters("f", c)
	where = fmt.Sprintf("WHERE f.LATITUDE IS NOT NULL AND f.LONGITUDE IS NOT NULL AND f.LATITUDE != 0 AND f.LONGITUDE != 0 AND %s", where)
	villageNameExpr := "''"
	if h.familyColumnExists("VILLAGE_NAME") {
		villageNameExpr = "TRIM(COALESCE(f.VILLAGE_NAME, ''))"
	}

	literateMembersExpr := "0"
	if h.populationMemberColumnExists("EVER_ATTENDED_SCHOOL") {
		literateMembersExpr = "SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'YES' THEN 1 ELSE 0 END)"
	}

	illiterateMembersExpr := "0"
	if h.populationMemberColumnExists("EVER_ATTENDED_SCHOOL") {
		illiterateMembersExpr = "SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'NO' THEN 1 ELSE 0 END)"
	}

	unemployedMembersExpr := "0"
	if h.populationMemberColumnExists("OCCUPATION") {
		unemployedMembersExpr = "SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.OCCUPATION, ''))) IN ('UNEMPLOYED', 'NOT WORKING') THEN 1 ELSE 0 END)"
	}

	villageWorkingMembersExpr := `SUM(CASE
					WHEN (
						fm.OCCUPATION IS NOT NULL
						AND TRIM(fm.OCCUPATION) != ''
						AND LOWER(TRIM(fm.OCCUPATION)) NOT IN ('not working', 'unemployed')
					)
					THEN 1
					ELSE 0
				END)`
	if h.populationMemberColumnExists("NATURE_WAGE_WORK") {
		villageWorkingMembersExpr = `SUM(CASE
					WHEN (
						fm.OCCUPATION IS NOT NULL
						AND TRIM(fm.OCCUPATION) != ''
						AND LOWER(TRIM(fm.OCCUPATION)) NOT IN ('not working', 'unemployed')
					)
					OR NULLIF(TRIM(COALESCE(fm.NATURE_WAGE_WORK, '')), '') IS NOT NULL
					THEN 1
					ELSE 0
				END)`
	}

	divyangCase := "0"
	hasDivyang := h.populationMemberColumnExists("DIVYANG")
	hasDisability := h.populationMemberColumnExists("DISABILITY")
	if hasDivyang && hasDisability {
		divyangCase = "CASE WHEN UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES' OR UPPER(TRIM(COALESCE(fm.DISABILITY, ''))) = 'YES' THEN 1 ELSE 0 END"
	} else if hasDivyang {
		divyangCase = "CASE WHEN UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES' THEN 1 ELSE 0 END"
	} else if hasDisability {
		divyangCase = "CASE WHEN UPPER(TRIM(COALESCE(fm.DISABILITY, ''))) = 'YES' THEN 1 ELSE 0 END"
	}
	divyangMembersExpr := fmt.Sprintf("SUM(%s)", divyangCase)

	query := fmt.Sprintf(`
		SELECT
			COALESCE(CAST(f.EXTERNAL_FAMILY_ID AS CHAR), '') AS external_family_id,
			COALESCE(CAST(f.HOUSE_NO AS CHAR), '') AS house_no,
			COALESCE(TRIM(CONCAT(
				COALESCE(f.FIRST_NAME_HOUSEHOLD_HEAD, ''), ' ',
				COALESCE(f.MIDDLE_NAME_HOUSEHOLD_HEAD, ''), ' ',
				COALESCE(f.LAST_NAME_HOUSEHOLD_HEAD, '')
			)), '') AS head_name,
			COALESCE(CAST(f.VILLAGE_ID AS CHAR), '') AS village_id,
			COALESCE(%s, '') AS village_name,
			f.LATITUDE AS lat,
			f.LONGITUDE AS lng,
			COALESCE(fm_agg.has_disability, 0) AS has_disability,
			COALESCE(TRIM(COALESCE(f.FAMILY_BELONG_BPL_CATEGORY, '')), '') AS FAMILY_BELONG_BPL_CATEGORY,
			COALESCE(TRIM(COALESCE(f.RATION_CARD_TYPE, '')), '') AS RATION_CARD_TYPE,
			COALESCE(TRIM(CAST(f.ANNUAL_INCOME AS CHAR)), '') AS ANNUAL_INCOME,
			COALESCE(fm_agg.total_members, 0) AS total_members,
			COALESCE(fm_agg.literate_members, 0) AS literate_members,
			COALESCE(fm_agg.illiterate_members, 0) AS illiterate_members,
			COALESCE(fm_agg.village_working_members, 0) AS village_working_members,
			COALESCE(fm_agg.unemployed_members, 0) AS unemployed_members,
			COALESCE(fm_agg.divyang_members, 0) AS divyang_members,
			COALESCE(fm_agg.male_members, 0) AS male_members,
			COALESCE(fm_agg.female_members, 0) AS female_members,
			COALESCE(fm_agg.working_members, 0) AS working_members,
			(COALESCE(fm_agg.total_members, 0) - COALESCE(fm_agg.working_members, 0)) AS non_working_members,
			COALESCE(REPLACE(fm_agg.occupation_list, '|', ', '), '') AS working_occupations,
			COALESCE(fm_agg.occupation_list, '') AS occupation_list
		FROM FAMILY f
		LEFT JOIN (
			SELECT
				fm.EXTERNAL_FAMILY_ID,
				COUNT(fm.FAMILY_MEMBER_ID) AS total_members,
				%s AS literate_members,
				%s AS illiterate_members,
				%s AS village_working_members,
				%s AS unemployed_members,
				%s AS divyang_members,
				SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER, ''))) = 'male' THEN 1 ELSE 0 END) AS male_members,
				SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER, ''))) = 'female' THEN 1 ELSE 0 END) AS female_members,
				MAX(CASE
					WHEN UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES'
						OR NULLIF(TRIM(COALESCE(fm.DISABILITY, '')), '') IS NOT NULL
						OR NULLIF(TRIM(COALESCE(CAST(fm.DISABILITY_PERCENTAGE AS CHAR), '')), '') IS NOT NULL
					THEN 1
					ELSE 0
				END) AS has_disability,
				SUM(CASE
					WHEN fm.OCCUPATION IS NOT NULL
						AND TRIM(fm.OCCUPATION) != ''
						AND LOWER(TRIM(fm.OCCUPATION)) NOT IN (
							'housewife',
							'homemaker',
							'student',
							'studying',
							'not applicable',
							'na',
							'n/a',
							'none',
							'unemployed',
							'child',
							'nil',
							'no work'
						)
					THEN 1
					ELSE 0
				END) AS working_members,
				COALESCE(GROUP_CONCAT(
					DISTINCT
					CASE
						WHEN fm.OCCUPATION IS NOT NULL
							AND TRIM(fm.OCCUPATION) != ''
							AND LOWER(TRIM(fm.OCCUPATION)) NOT IN (
								'housewife',
								'homemaker',
								'student',
								'studying',
								'not applicable',
								'na',
								'n/a',
								'none',
								'unemployed',
								'child',
								'nil',
								'no work'
							)
						THEN TRIM(fm.OCCUPATION)
					END
					SEPARATOR '|'
				), '') AS occupation_list
			FROM FAMILY_MEMBER fm
			GROUP BY fm.EXTERNAL_FAMILY_ID
		) fm_agg ON fm_agg.EXTERNAL_FAMILY_ID = f.EXTERNAL_FAMILY_ID
		%s
		ORDER BY f.HOUSE_NO, f.EXTERNAL_FAMILY_ID
	`, villageNameExpr, literateMembersExpr, illiterateMembersExpr, villageWorkingMembersExpr, unemployedMembersExpr, divyangMembersExpr, where)

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
			&marker.VillageID,
			&marker.VillageName,
			&marker.Lat,
			&marker.Lng,
			&marker.HasDisability,
			&marker.FamilyBelongBPLCategory,
			&marker.RationCardType,
			&marker.AnnualIncome,
			&marker.TotalMembers,
			&marker.LiterateMembers,
			&marker.IlliterateMembers,
			&marker.VillageWorkingMembers,
			&marker.UnemployedMembers,
			&marker.DivyangMembers,
			&marker.MaleMembers,
			&marker.FemaleMembers,
			&marker.WorkingMembers,
			&marker.NonWorkingMembers,
			&marker.WorkingOccupations,
			&marker.OccupationListRaw,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan population map marker", "detail": err.Error()})
			return
		}

		if marker.OccupationListRaw != "" {
			marker.OccupationList = strings.Split(marker.OccupationListRaw, "|")
		}

		markers = append(markers, marker)
	}

	if markers == nil {
		markers = []PopulationMapMarker{}
	}

	h.annotatePopulationLocationQuality(markers)

	c.JSON(http.StatusOK, markers)
}

type populationVillageGeoStats struct {
	sumLat      float64
	sumLng      float64
	validCount  int
	coordCount  map[string]int
	coordHouses map[string][]string
}

// annotatePopulationLocationQuality adds location quality flags in O(n) by
// analysing each village independently with two linear passes.
func (h *PopulationHandler) annotatePopulationLocationQuality(markers []PopulationMapMarker) {
	if len(markers) == 0 {
		return
	}

	villageGroups := make(map[string][]int, len(markers))
	villageStats := make(map[string]*populationVillageGeoStats, len(markers))

	for i := range markers {
		villageID := strings.TrimSpace(markers[i].VillageID)
		if villageID == "" {
			villageID = "__unknown_village__"
		}

		villageGroups[villageID] = append(villageGroups[villageID], i)

		stats := villageStats[villageID]
		if stats == nil {
			stats = &populationVillageGeoStats{coordCount: map[string]int{}, coordHouses: map[string][]string{}}
			villageStats[villageID] = stats
		}

		lat, lng := markers[i].Lat, markers[i].Lng
		if isMissingPopulationCoordinate(lat, lng) || !isPopulationCoordinateRangeValid(lat, lng) {
			continue
		}

		coordKey := roundedPopulationCoordKey(lat, lng)
		stats.coordCount[coordKey]++
		houseNo := strings.TrimSpace(markers[i].HouseNo)
		if houseNo != "" {
			stats.coordHouses[coordKey] = append(stats.coordHouses[coordKey], houseNo)
		}
		stats.sumLat += lat
		stats.sumLng += lng
		stats.validCount++
	}

	for villageID, indices := range villageGroups {
		stats := villageStats[villageID]
		avgLat, avgLng := 0.0, 0.0
		if stats != nil && stats.validCount > 0 {
			avgLat = stats.sumLat / float64(stats.validCount)
			avgLng = stats.sumLng / float64(stats.validCount)
		}

		for _, idx := range indices {
			lat, lng := markers[idx].Lat, markers[idx].Lng

			if isMissingPopulationCoordinate(lat, lng) {
				markers[idx].LocationQuality = "missing"
				markers[idx].LocationReason = "GPS missing"
				continue
			}

			if !isPopulationCoordinateRangeValid(lat, lng) {
				markers[idx].LocationQuality = "suspicious"
				markers[idx].LocationReason = "Invalid coordinate range"
				continue
			}

			if looksLikePopulationSwappedCoordinate(lat, lng) {
				markers[idx].LocationQuality = "suspicious"
				markers[idx].LocationReason = "Latitude and longitude appear swapped"
				continue
			}

			coordKey := roundedPopulationCoordKey(lat, lng)
			duplicates := 0
			var duplicateHouses []string
			if stats != nil {
				duplicates = stats.coordCount[coordKey]
				duplicateHouses = stats.coordHouses[coordKey]
			}

			if duplicates >= 6 {
				markers[idx].DuplicateCount = duplicates
				markers[idx].DuplicateHouses = duplicateHouses
			}

			if duplicates >= 8 {
				markers[idx].LocationQuality = "suspicious"
				markers[idx].LocationReason = "Too many houses share same coordinates"
				continue
			}

			if stats != nil && stats.validCount > 0 {
				dLat := lat - avgLat
				dLng := lng - avgLng
				dist := math.Sqrt(dLat*dLat + dLng*dLng)

				if dist > 0.05 {
					markers[idx].LocationQuality = "suspicious"
					markers[idx].LocationReason = "Far outside village boundary"
					continue
				}

				if dist > 0.022 {
					markers[idx].LocationQuality = "approximate"
					markers[idx].LocationReason = "Outside main settlement area"
					continue
				}
			}

			if duplicates >= 6 {
				markers[idx].LocationQuality = "approximate"
				markers[idx].LocationReason = "Coordinates shared by many nearby households"
				continue
			}

			markers[idx].LocationQuality = "valid"
			markers[idx].LocationReason = "Within village cluster"
		}
	}
}

func isMissingPopulationCoordinate(lat, lng float64) bool {
	return lat == 0 || lng == 0 || math.IsNaN(lat) || math.IsNaN(lng)
}

func isPopulationCoordinateRangeValid(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

func roundedPopulationCoordKey(lat, lng float64) string {
	return fmt.Sprintf("%.5f,%.5f", lat, lng)
}

func looksLikePopulationSwappedCoordinate(lat, lng float64) bool {
	// India-focused heuristic: swapped values often look like lat ~ 70-98 and lng ~ 6-38.
	return lat >= 68 && lat <= 98 && lng >= 6 && lng <= 38
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
