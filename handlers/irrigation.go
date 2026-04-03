package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IrrigationRecord struct {
	WaterSource    string `json:"waterSource"`
	PumpIrrigation string `json:"pumpIrrigation"`
}

type IrrigationHandler struct {
	DB *sql.DB
}

func (h *IrrigationHandler) GetIrrigation(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT
			COALESCE(SOURCE_WATER_IRRIGATION, ''),
			COALESCE(A_PUMP_IRRIGATION, '')
		FROM FAMILY
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch irrigation data"})
		return
	}
	defer rows.Close()

	var records []IrrigationRecord
	for rows.Next() {
		var record IrrigationRecord
		if err := rows.Scan(&record.WaterSource, &record.PumpIrrigation); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan irrigation record"})
			return
		}
		records = append(records, record)
	}

	if records == nil {
		records = []IrrigationRecord{}
	}
	c.JSON(http.StatusOK, records)
}
