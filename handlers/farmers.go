package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type FarmerRecord struct {
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	Occupation         string `json:"occupation"`
	WorkDetails        string `json:"workDetails"`
	OwnAgricultureLand string `json:"ownAgricultureLand"`
	TotalLand          string `json:"totalLand"`
	WaterSource        string `json:"waterSource"`
	RationCard         string `json:"rationCard"`
	AnnualIncome       string `json:"annualIncome"`
	ChildrenCount      int    `json:"childrenCount"`
}

type FarmerHandler struct {
	DB *sql.DB
}

func (h *FarmerHandler) familyMemberColumnExists(column string) bool {
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

func (h *FarmerHandler) GetFarmers(c *gin.Context) {
	workExpr := `COALESCE(NULLIF(TRIM(COALESCE(fm.OCCUPATION, '')), ''), 'Not Working')`
	if h.familyMemberColumnExists("NATURE_WAGE_WORK") {
		workExpr = `COALESCE(NULLIF(TRIM(COALESCE(fm.NATURE_WAGE_WORK, '')), ''), NULLIF(TRIM(COALESCE(fm.OCCUPATION, '')), ''), 'Not Working')`
	}

	query := strings.ReplaceAll(`
		SELECT
			COALESCE(fm.FIRST_NAME, ''),
			COALESCE(fm.LAST_NAME, ''),
			COALESCE(fm.OCCUPATION, ''),
			__WORK_DETAILS_EXPR__,
			COALESCE(f.OWN_AGRICULTURE_LAND, ''),
			COALESCE(f.AREA_AGRICULTURE_LAND_ACRES, ''),
			COALESCE(f.SOURCE_WATER_IRRIGATION, ''),
			COALESCE(f.RATION_CARD_TYPE, ''),
			COALESCE(f.ANNUAL_INCOME, ''),
			COALESCE(children.children_count, 0)
		FROM FAMILY f
		JOIN FAMILY_MEMBER fm ON fm.EXTERNAL_FAMILY_ID = f.FAMILY_ID
		LEFT JOIN (
			SELECT
				cm.EXTERNAL_FAMILY_ID,
				SUM(
					CASE
						WHEN STR_TO_DATE(cm.DOB, '%Y-%m-%d') IS NOT NULL
							AND TIMESTAMPDIFF(YEAR, STR_TO_DATE(cm.DOB, '%Y-%m-%d'), CURDATE()) < 18 THEN 1
						WHEN STR_TO_DATE(cm.DOB, '%d-%m-%Y') IS NOT NULL
							AND TIMESTAMPDIFF(YEAR, STR_TO_DATE(cm.DOB, '%d-%m-%Y'), CURDATE()) < 18 THEN 1
						ELSE 0
					END
				) AS children_count
			FROM FAMILY_MEMBER cm
			GROUP BY cm.EXTERNAL_FAMILY_ID
		) children ON children.EXTERNAL_FAMILY_ID = f.FAMILY_ID
	`, "__WORK_DETAILS_EXPR__", workExpr)

	rows, err := h.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch farmer data"})
		return
	}
	defer rows.Close()

	var farmers []FarmerRecord
	for rows.Next() {
		var farmer FarmerRecord
		if err := rows.Scan(&farmer.FirstName, &farmer.LastName, &farmer.Occupation, &farmer.WorkDetails, &farmer.OwnAgricultureLand, &farmer.TotalLand, &farmer.WaterSource, &farmer.RationCard, &farmer.AnnualIncome, &farmer.ChildrenCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan farmer record"})
			return
		}
		farmer.WorkDetails = strings.TrimSpace(farmer.WorkDetails)
		if farmer.WorkDetails == "" {
			farmer.WorkDetails = "Not Working"
		}
		farmers = append(farmers, farmer)
	}

	if farmers == nil {
		farmers = []FarmerRecord{}
	}
	c.JSON(http.StatusOK, farmers)
}
