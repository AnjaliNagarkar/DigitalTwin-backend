package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DistrictBpl struct {
	DistrictID int   `json:"district_id"`
	BplCount   int64 `json:"bpl_count"`
}

type DistrictBplHandler struct {
	DB *sql.DB
}

// GetDistrictBpl returns district-wise BPL counts from FAMILY.
func (h *DistrictBplHandler) GetDistrictBpl(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT
			f.DISTRICT_ID,
			COUNT(*) AS bpl_count
		FROM FAMILY f
		WHERE LOWER(TRIM(f.FAMILY_BELONG_BPL_CATEGORY)) = 'yes'
		GROUP BY f.DISTRICT_ID
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch district bpl counts", "detail": err.Error()})
		return
	}
	defer rows.Close()

	result := []DistrictBpl{}
	for rows.Next() {
		var item DistrictBpl
		if scanErr := rows.Scan(&item.DistrictID, &item.BplCount); scanErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse district bpl counts", "detail": scanErr.Error()})
			return
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read district bpl counts", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

