package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LandRecord struct {
	TotalLand      string `json:"totalLand"`
	CultivatedLand string `json:"cultivatedLand"`
}

type LandHandler struct {
	DB *sql.DB
}

func (h *LandHandler) GetLand(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT
			COALESCE(AREA_AGRICULTURE_LAND_ACRES, ''),
			COALESCE(LAND_UNDER_CULTIVATION_ACRES, '')
		FROM FAMILY
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch land data"})
		return
	}
	defer rows.Close()

	var lands []LandRecord
	for rows.Next() {
		var land LandRecord
		if err := rows.Scan(&land.TotalLand, &land.CultivatedLand); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan land record"})
			return
		}
		lands = append(lands, land)
	}

	if lands == nil {
		lands = []LandRecord{}
	}
	c.JSON(http.StatusOK, lands)
}
