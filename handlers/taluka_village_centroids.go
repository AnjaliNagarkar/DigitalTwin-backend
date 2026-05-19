package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// TalukaCentroidRow is one taluka aggregate for map drill-down (GET /map/taluka-centroids).
type TalukaCentroidRow struct {
	TalukaID   int     `json:"taluka_id"`
	TalukaName string  `json:"taluka_name"`
	Count      int64   `json:"count"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
}

// TalukaCentroidsHandler serves GET /map/taluka-centroids?district_id=…
type TalukaCentroidsHandler struct {
	DB *sql.DB
}

// GetTalukaCentroids returns mean lat/lng and survey count per taluka within a district.
func (h *TalukaCentroidsHandler) GetTalukaCentroids(c *gin.Context) {
	districtID := strings.TrimSpace(c.Query("district_id"))
	if districtID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "district_id is required"})
		return
	}
	did, err := strconv.Atoi(districtID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid district_id"})
		return
	}

	const query = `
SELECT
f.TALUKA_ID,
tm.vsTalukaName AS taluka_name,
COUNT(*) AS total_count,
AVG(CAST(f.LATITUDE AS DECIMAL(10,6))) AS lat,
AVG(CAST(f.LONGITUDE AS DECIMAL(10,6))) AS lng
FROM FAMILY f
LEFT JOIN taluka_master tm
ON tm.pklTalukaId = f.TALUKA_ID
WHERE f.DISTRICT_ID = ?
AND f.LATITUDE IS NOT NULL
AND f.LONGITUDE IS NOT NULL
GROUP BY f.TALUKA_ID, tm.vsTalukaName
`

	rows, err := h.DB.Query(query, did)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch taluka centroids", "detail": err.Error()})
		return
	}
	defer rows.Close()

	out := make([]TalukaCentroidRow, 0)
	for rows.Next() {
		var talukaID sql.NullInt64
		var name sql.NullString
		var count int64
		var lat, lng sql.NullFloat64

		if scanErr := rows.Scan(&talukaID, &name, &count, &lat, &lng); scanErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan taluka centroids", "detail": scanErr.Error()})
			return
		}
		if !talukaID.Valid {
			continue
		}
		row := TalukaCentroidRow{
			TalukaID: int(talukaID.Int64),
			Count:    count,
		}
		if name.Valid {
			row.TalukaName = name.String
		}
		if lat.Valid {
			row.Lat = lat.Float64
		}
		if lng.Valid {
			row.Lng = lng.Float64
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read taluka centroids", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}

// VillageCentroidRow is one village aggregate for map drill-down (GET /map/village-centroids).
type VillageCentroidRow struct {
	VillageID   int     `json:"village_id"`
	VillageName string  `json:"village_name"`
	Count       int64   `json:"count"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
}

// VillageCentroidsHandler serves GET /map/village-centroids?district_id=…&taluka_id=…
type VillageCentroidsHandler struct {
	DB *sql.DB
}

// GetVillageCentroids returns mean lat/lng and survey count per village within a taluka.
func (h *VillageCentroidsHandler) GetVillageCentroids(c *gin.Context) {
	districtID := strings.TrimSpace(c.Query("district_id"))
	talukaID := strings.TrimSpace(c.Query("taluka_id"))
	if districtID == "" || talukaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "district_id and taluka_id are required"})
		return
	}
	did, err := strconv.Atoi(districtID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid district_id"})
		return
	}
	tid, err := strconv.Atoi(talukaID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid taluka_id"})
		return
	}

	const query = `
SELECT
f.VILLAGE_ID,
vm.vsVillageName AS village_name,
COUNT(*) AS total_count,
AVG(CAST(f.LATITUDE AS DECIMAL(10,6))) AS lat,
AVG(CAST(f.LONGITUDE AS DECIMAL(10,6))) AS lng
FROM FAMILY f
LEFT JOIN village_master vm
ON vm.pklVillageId = f.VILLAGE_ID
WHERE f.DISTRICT_ID = ?
AND f.TALUKA_ID = ?
AND f.LATITUDE IS NOT NULL
AND f.LONGITUDE IS NOT NULL
GROUP BY f.VILLAGE_ID, vm.vsVillageName
`

	rows, err := h.DB.Query(query, did, tid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch village centroids", "detail": err.Error()})
		return
	}
	defer rows.Close()

	out := make([]VillageCentroidRow, 0)
	for rows.Next() {
		var villageID sql.NullInt64
		var name sql.NullString
		var count int64
		var lat, lng sql.NullFloat64

		if scanErr := rows.Scan(&villageID, &name, &count, &lat, &lng); scanErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan village centroids", "detail": scanErr.Error()})
			return
		}
		if !villageID.Valid {
			continue
		}
		row := VillageCentroidRow{
			VillageID: int(villageID.Int64),
			Count:     count,
		}
		if name.Valid {
			row.VillageName = name.String
		}
		if lat.Valid {
			row.Lat = lat.Float64
		}
		if lng.Valid {
			row.Lng = lng.Float64
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read village centroids", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}
