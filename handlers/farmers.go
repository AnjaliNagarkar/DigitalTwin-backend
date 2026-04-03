package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FarmerRecord struct {
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	OwnAgricultureLand string `json:"ownAgricultureLand"`
}

type FarmerHandler struct {
	DB *sql.DB
}

func (h *FarmerHandler) GetFarmers(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT
			COALESCE(fm.FIRST_NAME, ''),
			COALESCE(fm.LAST_NAME, ''),
			COALESCE(f.OWN_AGRICULTURE_LAND, '')
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
		if err := rows.Scan(&farmer.FirstName, &farmer.LastName, &farmer.OwnAgricultureLand); err != nil {
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
