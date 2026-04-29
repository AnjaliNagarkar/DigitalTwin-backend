package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DistrictPopulation struct {
	DistrictID int   `json:"district_id"`
	Population int64 `json:"population"`
}

type DistrictPopulationHandler struct {
	DB *sql.DB
}

// GetDistrictPopulation returns district-wise population totals from FAMILY_MEMBER.
func (h *DistrictPopulationHandler) GetDistrictPopulation(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT
			fm.DISTRICT_ID,
			COUNT(*) AS population
		FROM FAMILY_MEMBER fm
		WHERE fm.DISTRICT_ID IS NOT NULL
		GROUP BY fm.DISTRICT_ID
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch district population", "detail": err.Error()})
		return
	}
	defer rows.Close()

	result := []DistrictPopulation{}
	for rows.Next() {
		var item DistrictPopulation
		if scanErr := rows.Scan(&item.DistrictID, &item.Population); scanErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse district population", "detail": scanErr.Error()})
			return
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read district population", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
