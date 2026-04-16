package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FarmerRecord struct {
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	Occupation         string `json:"occupation"`
	OwnAgricultureLand string `json:"ownAgricultureLand"`
	TotalLand          string `json:"totalLand"`
	WaterSource        string `json:"waterSource"`
	RationCard         string `json:"rationCard"`
}

type FarmerHandler struct {
	DB *sql.DB
}

func (h *FarmerHandler) GetFarmers(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT
			COALESCE(fm.FIRST_NAME, ''),
			COALESCE(fm.LAST_NAME, ''),
			COALESCE(fm.OCCUPATION, ''),
			COALESCE(f.OWN_AGRICULTURE_LAND, ''),
			COALESCE(f.AREA_AGRICULTURE_LAND_ACRES, ''),
			COALESCE(f.SOURCE_WATER_IRRIGATION, ''),
			COALESCE(f.RATION_CARD_TYPE, '')
		FROM FAMILY f
		JOIN FAMILY_MEMBER fm ON fm.EXTERNAL_FAMILY_ID = f.FAMILY_ID
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch farmer data"})
		return
	}
	defer rows.Close()

	var farmers []FarmerRecord
	for rows.Next() {
		var farmer FarmerRecord
		if err := rows.Scan(&farmer.FirstName, &farmer.LastName, &farmer.Occupation, &farmer.OwnAgricultureLand, &farmer.TotalLand, &farmer.WaterSource, &farmer.RationCard); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan farmer record"})
			return
		}
		farmers = append(farmers, farmer)
	}

	if farmers == nil {
		farmers = []FarmerRecord{}
	}
	c.JSON(http.StatusOK, farmers)
}
