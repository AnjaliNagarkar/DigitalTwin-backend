package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DistrictDivyangCount struct {
	DistrictID   int64 `json:"district_id"`
	DivyangCount int64 `json:"divyang_count"`
}

type DistrictDivyangHandler struct {
	DB *sql.DB
}

// GetDistrictDivyangCount returns district-wise Divyang counts from FAMILY_MEMBER.
func (h *DistrictDivyangHandler) GetDistrictDivyangCount(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT
			DISTRICT_ID,
			COUNT(*) AS divyang_count
		FROM FAMILY_MEMBER
		WHERE LOWER(TRIM(DIVYANG)) = 'yes'
		  AND DISTRICT_ID IS NOT NULL
		GROUP BY DISTRICT_ID
		ORDER BY DISTRICT_ID
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch district divyang counts", "detail": err.Error()})
		return
	}
	defer rows.Close()

	result := []DistrictDivyangCount{}
	for rows.Next() {
		var item DistrictDivyangCount
		if scanErr := rows.Scan(&item.DistrictID, &item.DivyangCount); scanErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse district divyang counts", "detail": scanErr.Error()})
			return
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read district divyang counts", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
