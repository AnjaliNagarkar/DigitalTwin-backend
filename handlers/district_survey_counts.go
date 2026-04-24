package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type DistrictSurveyCount struct {
	DistrictID   int    `json:"district_id"`
	DistrictName string `json:"district_name"`
	SurveyCount  int    `json:"survey_count"`
}

type DistrictSurveyCountHandler struct {
	DB *sql.DB
}

func (h *DistrictSurveyCountHandler) GetDistrictSurveyCounts(c *gin.Context) {
	districtID := strings.TrimSpace(c.Query("district_id"))
	talukaID := strings.TrimSpace(c.Query("taluka_id"))
	villageID := strings.TrimSpace(c.Query("village_id"))

	where := "WHERE 1=1"
	args := make([]interface{}, 0, 3)

	if districtID != "" {
		where += " AND CAST(f.DISTRICT_ID AS CHAR) = ?"
		args = append(args, districtID)
	}
	if talukaID != "" {
		where += " AND CAST(f.TALUKA_ID AS CHAR) = ?"
		args = append(args, talukaID)
	}
	if villageID != "" {
		where += " AND CAST(f.VILLAGE_ID AS CHAR) = ?"
		args = append(args, villageID)
	}

	query := `
		SELECT
			CAST(f.DISTRICT_ID AS UNSIGNED) AS district_id,
			MAX(COALESCE(d.vsDisplayName, d.vsDistrictName, '')) AS district_name,
			COUNT(*) AS survey_count
		FROM FAMILY f
		JOIN district_master d ON d.pklDistrictId = f.DISTRICT_ID
		` + where + `
		GROUP BY f.DISTRICT_ID
		ORDER BY survey_count DESC
	`

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch district survey counts", "detail": err.Error()})
		return
	}
	defer rows.Close()

	result := make([]DistrictSurveyCount, 0)
	for rows.Next() {
		var item DistrictSurveyCount
		if err := rows.Scan(&item.DistrictID, &item.DistrictName, &item.SurveyCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse district survey counts", "detail": err.Error()})
			return
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read district survey counts", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
