package handlers

import (
	"database/sql"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DistrictIrrigationCoverage struct {
	DistrictID int64   `json:"district_id"`
	Coverage   float64 `json:"coverage"`
	Irrigated  int64   `json:"irrigated"`
	Total      int64   `json:"total"`
}

type DistrictIrrigationCoverageHandler struct {
	DB *sql.DB
}

// GetDistrictIrrigationCoverage returns district-wise irrigation coverage percentage.
func (h *DistrictIrrigationCoverageHandler) GetDistrictIrrigationCoverage(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT 
			DISTRICT_ID,
			COUNT(*) AS total_households,
			SUM(
				CASE 
					WHEN SOURCE_WATER_IRRIGATION IS NOT NULL
					 AND TRIM(SOURCE_WATER_IRRIGATION) <> ''
					 AND LOWER(TRIM(SOURCE_WATER_IRRIGATION)) != 'no source of water irrigation'
					THEN 1 ELSE 0
				END
			) AS irrigated_households
		FROM FAMILY
		WHERE DISTRICT_ID IS NOT NULL
		GROUP BY DISTRICT_ID
		ORDER BY DISTRICT_ID
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch district irrigation coverage", "detail": err.Error()})
		return
	}
	defer rows.Close()

	result := []DistrictIrrigationCoverage{}
	for rows.Next() {
		var districtID int64
		var total int64
		var irrigated int64
		if scanErr := rows.Scan(&districtID, &total, &irrigated); scanErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse district irrigation coverage", "detail": scanErr.Error()})
			return
		}

		coverage := 0.0
		if total > 0 {
			coverage = (float64(irrigated) / float64(total)) * 100
			coverage = math.Round(coverage*100) / 100
		}

		result = append(result, DistrictIrrigationCoverage{
			DistrictID: districtID,
			Coverage:   coverage,
			Irrigated:  irrigated,
			Total:      total,
		})
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read district irrigation coverage", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
