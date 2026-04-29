package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DistrictEmploymentCount struct {
	DistrictID    int64 `json:"district_id"`
	EmployedCount int64 `json:"employed_count"`
}

type DistrictEmploymentHandler struct {
	DB *sql.DB
}

// GetDistrictEmploymentCount returns district-wise employed person counts from FAMILY_MEMBER.
func (h *DistrictEmploymentHandler) GetDistrictEmploymentCount(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT
			DISTRICT_ID,
			COUNT(*) AS employed_count
		FROM FAMILY_MEMBER
		WHERE TRIM(LOWER(OCCUPATION)) IN (
			'self employed - farm based',
			'self employed - agri allied',
			'self employed - non-farm based',
			'wage work',
			'salaried job'
		)
		  AND DISTRICT_ID IS NOT NULL
		GROUP BY DISTRICT_ID
		ORDER BY DISTRICT_ID
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch district employment counts", "detail": err.Error()})
		return
	}
	defer rows.Close()

	result := []DistrictEmploymentCount{}
	for rows.Next() {
		var item DistrictEmploymentCount
		if scanErr := rows.Scan(&item.DistrictID, &item.EmployedCount); scanErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse district employment counts", "detail": scanErr.Error()})
			return
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read district employment counts", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
