package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type PopulationRegistryRecord struct {
	Name       string `json:"name"`
	Gender     string `json:"gender"`
	Age        int    `json:"age"`
	Education  string `json:"education"`
	Occupation string `json:"occupation"`
}

type populationRegistryRow struct {
	Name       string
	Gender     string
	Age        sql.NullInt64
	Education  string
	Occupation string
}

func (h *PopulationHandler) GetPopulationRegistry(c *gin.Context) {
	dobColumn := h.populationMemberDateColumn()
	dobExpr := "NULL"
	if dobColumn != "" {
		dobExpr = fmt.Sprintf("fm.%s", dobColumn)
	}

	filter := strings.ToLower(strings.TrimSpace(c.Query("filter")))
	where := "1=1"
	switch filter {
	case "bpl":
		where = "UPPER(TRIM(COALESCE(f.FAMILY_BELONG_BPL_CATEGORY, ''))) = 'YES'"
	case "student":
		where = `(
			(
				UPPER(TRIM(COALESCE(fm.CURRENTLY_PURSUING_EDUCATION, ''))) = 'YES'
				OR NULLIF(TRIM(COALESCE(fm.STANDARD_WHICH_STUDYING, '')), '') IS NOT NULL
				OR NULLIF(TRIM(COALESCE(fm.TYPE_INSTITUTION, '')), '') IS NOT NULL
			)
			AND UPPER(TRIM(COALESCE(fm.CURRENTLY_PURSUING_EDUCATION, ''))) != 'NO'
		)`
	case "divyang":
		where = "UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES' OR UPPER(TRIM(COALESCE(fm.DISABILITY, ''))) = 'YES'"
	}

	query := fmt.Sprintf(`
		SELECT
			COALESCE(TRIM(CONCAT(
				COALESCE(fm.FIRST_NAME, ''), ' ',
				COALESCE(fm.LAST_NAME, '')
			)), '') AS name,
			COALESCE(fm.GENDER, '') AS gender,
			CASE
				WHEN STR_TO_DATE(%[1]s, '%%Y-%%m-%%d') IS NOT NULL THEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(%[1]s, '%%Y-%%m-%%d'), CURDATE())
				WHEN STR_TO_DATE(%[1]s, '%%d-%%m-%%Y') IS NOT NULL THEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(%[1]s, '%%d-%%m-%%Y'), CURDATE())
				ELSE NULL
			END AS age,
			COALESCE(NULLIF(TRIM(COALESCE(fm.QUALIFICATION, '')), ''), NULLIF(TRIM(COALESCE(fm.EDUCATION_STATUS, '')), ''), 'Not Available') AS education,
			COALESCE(NULLIF(TRIM(COALESCE(fm.OCCUPATION, '')), ''), 'Not Working') AS occupation
		FROM FAMILY_MEMBER fm
		LEFT JOIN FAMILY f ON fm.EXTERNAL_FAMILY_ID = f.FAMILY_ID
		WHERE %[2]s
		ORDER BY fm.FIRST_NAME, fm.LAST_NAME
	`, dobExpr, where)

	rows, err := h.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch population registry data", "detail": err.Error()})
		return
	}
	defer rows.Close()

	records := make([]PopulationRegistryRecord, 0)
	for rows.Next() {
		var row populationRegistryRow
		if err := rows.Scan(
			&row.Name,
			&row.Gender,
			&row.Age,
			&row.Education,
			&row.Occupation,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan population registry record", "detail": err.Error()})
			return
		}

		age := 0
		if row.Age.Valid {
			age = int(row.Age.Int64)
			if age < 0 {
				age = 0
			}
		}

		records = append(records, PopulationRegistryRecord{
			Name:       strings.TrimSpace(row.Name),
			Gender:     strings.TrimSpace(row.Gender),
			Age:        age,
			Education:  strings.TrimSpace(row.Education),
			Occupation: strings.TrimSpace(row.Occupation),
		})
	}

	if records == nil {
		records = []PopulationRegistryRecord{}
	}

	c.JSON(http.StatusOK, records)
}

func (h *PopulationHandler) populationMemberDateColumn() string {
	for _, column := range []string{"DOB", "D_O_B"} {
		var count int
		if err := h.DB.QueryRow(`
			SELECT COUNT(*)
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			  AND TABLE_NAME = 'FAMILY_MEMBER'
			  AND COLUMN_NAME = ?
		`, column).Scan(&count); err == nil && count > 0 {
			return column
		}
	}
	return ""
}
