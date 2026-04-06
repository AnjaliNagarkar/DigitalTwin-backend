package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LocationOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LocationHandler struct {
	DB *sql.DB
}

// GetLocationOptions returns district/taluka/village dropdown options.
// Optional filters:
// - district_id: narrows taluka/village options
// - taluka_id: narrows village options
func (h *LocationHandler) GetLocationOptions(c *gin.Context) {
	districtID := c.Query("district_id")
	talukaID := c.Query("taluka_id")

	districts := []LocationOption{}
	talukas := []LocationOption{}
	villages := []LocationOption{}

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

	tQuery := `
		SELECT CAST(tm.pklTalukaId AS CHAR), COALESCE(tm.vsDisplayName, tm.vsTalukaName, '')
		FROM taluka_master tm
		WHERE tm.bEnabled = 1
	`
	tArgs := []interface{}{}
	if districtID != "" {
		tQuery += " AND CAST(tm.fklDistrictId AS CHAR) = ?"
		tArgs = append(tArgs, districtID)
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

	vQuery := `
		SELECT DISTINCT CAST(vm.pklVillageId AS CHAR), COALESCE(vm.vsDisplayName, vm.vsVillageName, '')
		FROM village_master vm
		JOIN grampanchayat_master gm ON gm.pklGramPanchayatId = vm.fklGramPanchayatId
		JOIN taluka_master tm ON tm.pklTalukaId = gm.fklTalukaId
		WHERE vm.bEnabled = 1
	`
	vArgs := []interface{}{}
	if districtID != "" {
		vQuery += " AND CAST(tm.fklDistrictId AS CHAR) = ?"
		vArgs = append(vArgs, districtID)
	}
	if talukaID != "" {
		vQuery += " AND CAST(tm.pklTalukaId AS CHAR) = ?"
		vArgs = append(vArgs, talukaID)
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
