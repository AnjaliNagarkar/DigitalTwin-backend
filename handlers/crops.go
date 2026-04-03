package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CropRecord struct {
	Kharif          string `json:"kharif"`
	Rabi            string `json:"rabi"`
	RabiCultivation string `json:"rabiCultivation"`
}

type CropHandler struct {
	DB *sql.DB
}

func (h *CropHandler) GetCrops(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT
			COALESCE(CULTIVATING_DURING_KHARIF_SEASON, ''),
			COALESCE(TAKING_CROPS_RABI_SEASON, ''),
			COALESCE(CULTIVATING_DURING_RABI_SEASON, '')
		FROM FAMILY
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch crop data"})
		return
	}
	defer rows.Close()

	var crops []CropRecord
	for rows.Next() {
		var crop CropRecord
		if err := rows.Scan(&crop.Kharif, &crop.Rabi, &crop.RabiCultivation); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan crop record"})
			return
		}
		crops = append(crops, crop)
	}

	if crops == nil {
		crops = []CropRecord{}
	}
	c.JSON(http.StatusOK, crops)
}
