package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type LocationOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LocationHandler struct {
	DB *sql.DB
}

type DistrictMasterOption struct {
	PklDistrictId  int    `json:"pklDistrictId"`
	VsDistrictName string `json:"vsDistrictName"`
}

// GetDistricts returns the district master list used for district-id mapping.
func (h *LocationHandler) GetDistricts(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT
			dm.pklDistrictId,
			COALESCE(dm.vsDistrictName, '')
		FROM district_master dm
		WHERE dm.bEnabled = 1
		ORDER BY dm.pklDistrictId
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch districts", "detail": err.Error()})
		return
	}
	defer rows.Close()

	districts := []DistrictMasterOption{}
	for rows.Next() {
		var item DistrictMasterOption
		if scanErr := rows.Scan(&item.PklDistrictId, &item.VsDistrictName); scanErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan districts", "detail": scanErr.Error()})
			return
		}
		districts = append(districts, item)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read districts", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, districts)
}

// GetLocationOptions returns district/taluka/village dropdown options.
// Optional filters:
// - district_id or district_ids: single or comma-separated district IDs (narrows taluka/village options)
// - taluka_id or taluka_ids: single or comma-separated taluka IDs (narrows village options)
// - village_id or village_ids: single or comma-separated village IDs (for filtering)
func (h *LocationHandler) GetLocationOptions(c *gin.Context) {
	// Get parameters - support both singular and plural forms
	districtID := c.Query("district_id")
	districtIDs := c.Query("district_ids")
	talukaID := c.Query("taluka_id")
	talukaIDs := c.Query("taluka_ids")

	// Normalize district IDs (support both single and multiple)
	var districtIDList []string
	if districtIDs != "" {
		districtIDList = strings.Split(districtIDs, ",")
		for i, id := range districtIDList {
			districtIDList[i] = strings.TrimSpace(id)
		}
	} else if districtID != "" {
		districtIDList = []string{strings.TrimSpace(districtID)}
	}

	// Normalize taluka IDs (support both single and multiple)
	var talukaIDList []string
	if talukaIDs != "" {
		talukaIDList = strings.Split(talukaIDs, ",")
		for i, id := range talukaIDList {
			talukaIDList[i] = strings.TrimSpace(id)
		}
	} else if talukaID != "" {
		talukaIDList = []string{strings.TrimSpace(talukaID)}
	}

	districts := []LocationOption{}
	talukas := []LocationOption{}
	villages := []LocationOption{}

	// Fetch all districts
	dRows, err := h.DB.Query(`
		SELECT CAST(dm.pklDistrictId AS CHAR), COALESCE(dm.vsDisplayName, dm.vsDistrictName, '')
		FROM district_master dm
		WHERE dm.bEnabled = 1
		ORDER BY COALESCE(dm.vsDisplayName, dm.vsDistrictName)
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch districts", "detail": err.Error()})
		return
	}
	defer dRows.Close()
	for dRows.Next() {
		var o LocationOption
		dRows.Scan(&o.ID, &o.Name)
		districts = append(districts, o)
	}

	// Build taluka query with support for multiple districts
	tQuery := `
		SELECT CAST(tm.pklTalukaId AS CHAR), COALESCE(tm.vsDisplayName, tm.vsTalukaName, '')
		FROM taluka_master tm
		WHERE tm.bEnabled = 1
	`
	tArgs := []interface{}{}

	// Apply district filter if provided
	if len(districtIDList) > 0 {
		placeholders := make([]string, len(districtIDList))
		for i, id := range districtIDList {
			placeholders[i] = "?"
			tArgs = append(tArgs, id)
		}
		tQuery += " AND CAST(tm.fklDistrictId AS CHAR) IN (" + strings.Join(placeholders, ",") + ")"
	}

	tQuery += " ORDER BY COALESCE(tm.vsDisplayName, tm.vsTalukaName)"

	tRows, err := h.DB.Query(tQuery, tArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch talukas", "detail": err.Error()})
		return
	}
	defer tRows.Close()
	for tRows.Next() {
		var o LocationOption
		tRows.Scan(&o.ID, &o.Name)
		talukas = append(talukas, o)
	}

	// Build village query with support for multiple talukas and districts
	vQuery := `
		SELECT DISTINCT CAST(vm.pklVillageId AS CHAR), COALESCE(vm.vsDisplayName, vm.vsVillageName, '')
		FROM village_master vm
		JOIN grampanchayat_master gm ON gm.pklGramPanchayatId = vm.fklGramPanchayatId
		JOIN taluka_master tm ON tm.pklTalukaId = gm.fklTalukaId
		WHERE vm.bEnabled = 1
	`
	vArgs := []interface{}{}

	// Apply district filter if provided
	if len(districtIDList) > 0 {
		placeholders := make([]string, len(districtIDList))
		for i, id := range districtIDList {
			placeholders[i] = "?"
			vArgs = append(vArgs, id)
		}
		vQuery += " AND CAST(tm.fklDistrictId AS CHAR) IN (" + strings.Join(placeholders, ",") + ")"
	}

	// Apply taluka filter if provided
	if len(talukaIDList) > 0 {
		placeholders := make([]string, len(talukaIDList))
		for i, id := range talukaIDList {
			placeholders[i] = "?"
			vArgs = append(vArgs, id)
		}
		vQuery += " AND CAST(tm.pklTalukaId AS CHAR) IN (" + strings.Join(placeholders, ",") + ")"
	}

	vQuery += " ORDER BY COALESCE(vm.vsDisplayName, vm.vsVillageName)"

	vRows, err := h.DB.Query(vQuery, vArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch villages", "detail": err.Error()})
		return
	}
	defer vRows.Close()
	for vRows.Next() {
		var o LocationOption
		vRows.Scan(&o.ID, &o.Name)
		villages = append(villages, o)
	}

	c.JSON(http.StatusOK, gin.H{
		"districts": districts,
		"talukas":   talukas,
		"villages":  villages,
	})
}
