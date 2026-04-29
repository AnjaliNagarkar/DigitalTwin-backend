package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DistrictLandHoldingCoverage struct {
	DistrictID         int64   `json:"district_id"`
	LandHolders        int64   `json:"land_holders"`
	TotalFamilies      int64   `json:"total_families"`
	CoveragePercentage float64 `json:"coverage_percentage"`
}

type DistrictLandHoldingHandler struct {
	DB *sql.DB
}

// GetDistrictLandHoldingCoverage returns district-wise land holding coverage percentages.
func (h *DistrictLandHoldingHandler) GetDistrictLandHoldingCoverage(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT
			DISTRICT_ID,
			COUNT(CASE WHEN LOWER(TRIM(OWN_AGRICULTURE_LAND)) = 'yes' THEN 1 END) AS land_holders,
			COUNT(*) AS total_families,
			ROUND(
				COUNT(CASE WHEN LOWER(TRIM(OWN_AGRICULTURE_LAND)) = 'yes' THEN 1 END) * 100.0 / COUNT(*),
				2
			) AS coverage_percentage
		FROM FAMILY
		WHERE DISTRICT_ID IS NOT NULL
		GROUP BY DISTRICT_ID
		ORDER BY DISTRICT_ID
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch district land holding coverage", "detail": err.Error()})
		return
	}
	defer rows.Close()

	result := []DistrictLandHoldingCoverage{}
	for rows.Next() {
		var item DistrictLandHoldingCoverage
		if scanErr := rows.Scan(&item.DistrictID, &item.LandHolders, &item.TotalFamilies, &item.CoveragePercentage); scanErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse district land holding coverage", "detail": scanErr.Error()})
			return
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read district land holding coverage", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
