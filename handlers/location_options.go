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
// - district_id: narrows taluka/village options
// - taluka_id: narrows village options
func (h *LocationHandler) GetLocationOptions(c *gin.Context) {
	districtIDs := parseIDs(c.Query("district_ids"), c.Query("district_id"))
	talukaIDs := parseIDs(c.Query("taluka_ids"), c.Query("taluka_id"))

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
	if len(districtIDs) > 0 {
		tQuery += " AND CAST(tm.fklDistrictId AS CHAR) IN (" + placeholders(len(districtIDs)) + ")"
		for _, id := range districtIDs {
			tArgs = append(tArgs, id)
		}
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
	if len(districtIDs) > 0 {
		vQuery += " AND CAST(tm.fklDistrictId AS CHAR) IN (" + placeholders(len(districtIDs)) + ")"
		for _, id := range districtIDs {
			vArgs = append(vArgs, id)
		}
	}
	if len(talukaIDs) > 0 {
		vQuery += " AND CAST(tm.pklTalukaId AS CHAR) IN (" + placeholders(len(talukaIDs)) + ")"
		for _, id := range talukaIDs {
			vArgs = append(vArgs, id)
		}
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

func parseIDs(csv string, single string) []string {
	src := strings.TrimSpace(csv)
	if src == "" {
		src = strings.TrimSpace(single)
	}
	if src == "" {
		return nil
	}
	parts := strings.Split(src, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		id := strings.TrimSpace(part)
		if id == "" || id == "0" || strings.EqualFold(id, "null") || strings.EqualFold(id, "undefined") {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", n), ",")
}
