package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type DistrictCentroid struct {
	DistrictID   int     `json:"district_id"`
	DistrictName string  `json:"district_name"`
	Count        int64   `json:"count"`
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
}

type DistrictCentroidsHandler struct {
	DB *sql.DB
}

// GetDistrictCentroids returns survey count and centroid for each district using DISTRICT_ID.
func (h *DistrictCentroidsHandler) GetDistrictCentroids(c *gin.Context) {
	idCol, latCol, lngCol, lookupErr := h.detectDistrictMasterColumns()
	if lookupErr != nil {
		log.Printf("[map/district-centroids] district_master column detection failed: %v", lookupErr)
	}

	joinCol := idCol
	if joinCol == "" {
		joinCol = "pklDistrictId"
	}

	query := fmt.Sprintf(`
		SELECT
			f.DISTRICT_ID,
			MAX(COALESCE(dm.vsDistrictName, '')) AS district_name,
			COUNT(*) AS total_count,
			AVG(f.LATITUDE) AS lat,
			AVG(f.LONGITUDE) AS lng
		FROM FAMILY f
		LEFT JOIN district_master dm
			ON f.DISTRICT_ID = dm.%s
		WHERE f.DISTRICT_ID IS NOT NULL
		GROUP BY f.DISTRICT_ID
	`, joinCol)

	// If district master has usable coordinate columns, use them as fallback.
	if idCol != "" && latCol != "" && lngCol != "" {
		query = fmt.Sprintf(`
			SELECT
				f.DISTRICT_ID,
				MAX(COALESCE(dm.vsDistrictName, '')) AS district_name,
				COUNT(*) AS total_count,
				COALESCE(AVG(f.LATITUDE), MAX(dm.%s)) AS lat,
				COALESCE(AVG(f.LONGITUDE), MAX(dm.%s)) AS lng
			FROM FAMILY f
			LEFT JOIN district_master dm
				ON f.DISTRICT_ID = dm.%s
			WHERE f.DISTRICT_ID IS NOT NULL
			GROUP BY f.DISTRICT_ID
		`, latCol, lngCol, idCol)
	}

	rows, err := h.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch district centroids", "detail": err.Error()})
		return
	}
	defer rows.Close()

	centroids := []DistrictCentroid{}
	for rows.Next() {
		var centroid DistrictCentroid
		var districtName sql.NullString
		var lat sql.NullFloat64
		var lng sql.NullFloat64

		if scanErr := rows.Scan(&centroid.DistrictID, &districtName, &centroid.Count, &lat, &lng); scanErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan district centroids", "detail": scanErr.Error()})
			return
		}

		if districtName.Valid {
			centroid.DistrictName = districtName.String
		}
		if lat.Valid {
			centroid.Lat = lat.Float64
		}
		if lng.Valid {
			centroid.Lng = lng.Float64
		}

		centroids = append(centroids, centroid)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read district centroids", "detail": err.Error()})
		return
	}

	log.Printf("[map/district-centroids] districts returned=%d", len(centroids))

	c.JSON(http.StatusOK, centroids)
}
func (h *DistrictCentroidsHandler) detectDistrictMasterColumns() (string, string, string, error) {
	cols := []string{}
	rows, err := h.DB.Query(`
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'district_master'
	`)
	if err != nil {
		return "", "", "", err
	}
	defer rows.Close()

	for rows.Next() {
		var col string
		if scanErr := rows.Scan(&col); scanErr != nil {
			return "", "", "", scanErr
		}
		cols = append(cols, col)
	}
	if err := rows.Err(); err != nil {
		return "", "", "", err
	}

	pick := func(candidates []string) string {
		for _, candidate := range candidates {
			for _, existing := range cols {
				if strings.EqualFold(existing, candidate) {
					return existing
				}
			}
		}
		return ""
	}

	idCol := pick([]string{"DISTRICT_ID", "pklDistrictId"})
	latCol := pick([]string{"LAT", "LATITUDE", "CENTER_LAT", "CENTROID_LAT"})
	lngCol := pick([]string{"LNG", "LONGITUDE", "CENTER_LNG", "CENTER_LON", "CENTROID_LNG", "CENTROID_LON"})

	return idCol, latCol, lngCol, nil
}

