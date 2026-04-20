package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HouseRecord struct {
	FamilyID         int     `json:"familyId"`
	ExternalFamilyID string  `json:"externalFamilyId"`
	HouseNo          string  `json:"houseNo"`
	DistrictID       string  `json:"districtId"`
	DistrictName     string  `json:"districtName"`
	TalukaID         string  `json:"talukaId"`
	TalukaName       string  `json:"talukaName"`
	VillageID        string  `json:"villageId"`
	VillageName      string  `json:"villageName"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	TotalLand        string  `json:"totalLand"`
	CultivatedLand   string  `json:"cultivatedLand"`
	OwnLand          string  `json:"ownLand"`
	WaterSource      string  `json:"waterSource"`
	Kharif           string  `json:"kharif"`
	Rabi             string  `json:"rabi"`
	Latrine          string  `json:"latrine"`
	Lighting         string  `json:"lighting"`
	RationCard       string  `json:"rationCard"`
	Occupation       string  `json:"occupation"`
	HeadName         string  `json:"headName"`
}

type HouseDetail struct {
	HouseRecord
	Members []MemberRecord `json:"members"`
}

type MemberRecord struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type HouseHandler struct {
	DB *sql.DB
	CC *ColumnChecker
}

func (h *HouseHandler) GetHouses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "500"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 2000 {
		limit = 500
	}
	offset := (page - 1) * limit

	irrigation := c.Query("irrigation")
	ownLand := c.Query("own_land")
	districtID := c.Query("district_id")
	talukaID := c.Query("taluka_id")
	villageID := c.Query("village_id")

	// Use auto-detected lat/lng columns from ColumnChecker
	latCol := h.CC.LatCol
	lngCol := h.CC.LngCol
	if latCol == "" {
		latCol = "LATITUDE"
	}
	if lngCol == "" {
		lngCol = "LONGITUDE"
	}

	where := fmt.Sprintf(
		"WHERE f.%s IS NOT NULL AND f.%s IS NOT NULL AND f.%s != 0 AND f.%s != 0",
		latCol, lngCol, latCol, lngCol,
	)
	args := []interface{}{}

	if irrigation != "" {
		where += " AND f.SOURCE_WATER_IRRIGATION = ?"
		args = append(args, irrigation)
	}
	if ownLand != "" {
		where += " AND f.OWN_AGRICULTURE_LAND = ?"
		args = append(args, ownLand)
	}
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

	// Use the actual columns present in the FAMILY table.
	query := fmt.Sprintf(`
		SELECT
			f.FAMILY_ID,
				COALESCE(CAST(f.EXTERNAL_FAMILY_ID AS CHAR), ''),
				COALESCE(CAST(f.HOUSE_NO AS CHAR), ''),
			COALESCE(CAST(f.DISTRICT_ID AS CHAR), ''),
			COALESCE(dm.vsDisplayName, dm.vsDistrictName, ''),
			COALESCE(CAST(f.TALUKA_ID AS CHAR), ''),
			COALESCE(tm.vsDisplayName, tm.vsTalukaName, ''),
			COALESCE(CAST(f.VILLAGE_ID AS CHAR), ''),
			COALESCE(vm.vsDisplayName, vm.vsVillageName, ''),
			COALESCE(f.%s, 0),
			COALESCE(f.%s, 0),
			COALESCE(f.AREA_AGRICULTURE_LAND_ACRES, ''),
			COALESCE(f.LAND_UNDER_CULTIVATION_ACRES, ''),
			COALESCE(f.OWN_AGRICULTURE_LAND, ''),
			COALESCE(f.SOURCE_WATER_IRRIGATION, ''),
			COALESCE(f.CULTIVATING_DURING_KHARIF_SEASON, ''),
			COALESCE(f.TAKING_CROPS_RABI_SEASON, ''),
			COALESCE(f.SANITATION_TOILET_FACILITY, ''),
			COALESCE(f.ELECTRICITY_CONNECTION, ''),
			COALESCE(f.RATION_CARD_TYPE, ''),
			COALESCE(occ.primary_occupation, ''),
			COALESCE(TRIM(CONCAT(
				COALESCE(f.FIRST_NAME_HOUSEHOLD_HEAD, ''), ' ',
				COALESCE(f.MIDDLE_NAME_HOUSEHOLD_HEAD, ''), ' ',
				COALESCE(f.LAST_NAME_HOUSEHOLD_HEAD, '')
			)), '')
		FROM FAMILY f
		LEFT JOIN (
			SELECT
				fm.EXTERNAL_FAMILY_ID,
				COALESCE(
					MAX(NULLIF(TRIM(COALESCE(fm.NATURE_WAGE_WORK, '')), '')),
					MAX(NULLIF(TRIM(COALESCE(fm.OCCUPATION, '')), '')),
					''
				) AS primary_occupation
			FROM FAMILY_MEMBER fm
			GROUP BY fm.EXTERNAL_FAMILY_ID
		) occ ON occ.EXTERNAL_FAMILY_ID = f.EXTERNAL_FAMILY_ID
		LEFT JOIN district_master dm ON dm.pklDistrictId = f.DISTRICT_ID
		LEFT JOIN taluka_master tm ON tm.pklTalukaId = f.TALUKA_ID
		LEFT JOIN village_master vm ON vm.pklVillageId = f.VILLAGE_ID
		%s
		ORDER BY f.FAMILY_ID
		LIMIT %d OFFSET %d
	`,
		latCol,
		lngCol,
		where,
		limit,
		offset,
	)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch houses", "detail": err.Error()})
		return
	}
	defer rows.Close()

	var houses []HouseRecord
	for rows.Next() {
		var house HouseRecord
		if err := rows.Scan(
			&house.FamilyID,
			&house.ExternalFamilyID,
			&house.HouseNo,
			&house.DistrictID, &house.DistrictName,
			&house.TalukaID, &house.TalukaName,
			&house.VillageID, &house.VillageName,
			&house.Latitude, &house.Longitude,
			&house.TotalLand, &house.CultivatedLand, &house.OwnLand,
			&house.WaterSource, &house.Kharif, &house.Rabi,
			&house.Latrine, &house.Lighting, &house.RationCard,
			&house.Occupation,
			&house.HeadName,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan house record", "detail": err.Error()})
			return
		}
		houses = append(houses, house)
	}

	if houses == nil {
		houses = []HouseRecord{}
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM FAMILY f %s", where)
	var total int
	h.DB.QueryRow(countQuery, args...).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"data":  houses,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *HouseHandler) GetHouseByID(c *gin.Context) {
	id := c.Param("id")
	cc := h.CC

	latCol := cc.LatCol
	lngCol := cc.LngCol
	if latCol == "" {
		latCol = "LATITUDE"
	}
	if lngCol == "" {
		lngCol = "LONGITUDE"
	}

	query := fmt.Sprintf(`
		SELECT
			f.FAMILY_ID,
			COALESCE(CAST(f.EXTERNAL_FAMILY_ID AS CHAR), ''),
			COALESCE(CAST(f.HOUSE_NO AS CHAR), ''),
			COALESCE(CAST(f.DISTRICT_ID AS CHAR), ''),
			COALESCE(dm.vsDisplayName, dm.vsDistrictName, ''),
			COALESCE(CAST(f.TALUKA_ID AS CHAR), ''),
			COALESCE(tm.vsDisplayName, tm.vsTalukaName, ''),
			COALESCE(CAST(f.VILLAGE_ID AS CHAR), ''),
			COALESCE(vm.vsDisplayName, vm.vsVillageName, ''),
			COALESCE(f.%s, 0),
			COALESCE(f.%s, 0),
			COALESCE(f.AREA_AGRICULTURE_LAND_ACRES, ''),
			COALESCE(f.LAND_UNDER_CULTIVATION_ACRES, ''),
			COALESCE(f.OWN_AGRICULTURE_LAND, ''),
			COALESCE(f.SOURCE_WATER_IRRIGATION, ''),
			COALESCE(f.CULTIVATING_DURING_KHARIF_SEASON, ''),
			COALESCE(f.TAKING_CROPS_RABI_SEASON, ''),
			COALESCE(f.SANITATION_TOILET_FACILITY, ''),
			COALESCE(f.ELECTRICITY_CONNECTION, ''),
			COALESCE(f.RATION_CARD_TYPE, ''),
			COALESCE(occ.primary_occupation, ''),
			COALESCE(
				TRIM(CONCAT(
					COALESCE(f.FIRST_NAME_HOUSEHOLD_HEAD, ''), ' ',
					COALESCE(f.MIDDLE_NAME_HOUSEHOLD_HEAD, ''), ' ',
					COALESCE(f.LAST_NAME_HOUSEHOLD_HEAD, '')
				)), '')
		FROM FAMILY f
		LEFT JOIN (
			SELECT
				fm.EXTERNAL_FAMILY_ID,
				COALESCE(
					MAX(NULLIF(TRIM(COALESCE(fm.NATURE_WAGE_WORK, '')), '')),
					MAX(NULLIF(TRIM(COALESCE(fm.OCCUPATION, '')), '')),
					''
				) AS primary_occupation
			FROM FAMILY_MEMBER fm
			GROUP BY fm.EXTERNAL_FAMILY_ID
		) occ ON occ.EXTERNAL_FAMILY_ID = f.EXTERNAL_FAMILY_ID
		LEFT JOIN district_master dm ON dm.pklDistrictId = f.DISTRICT_ID
		LEFT JOIN taluka_master tm ON tm.pklTalukaId = f.TALUKA_ID
		LEFT JOIN village_master vm ON vm.pklVillageId = f.VILLAGE_ID
		WHERE f.FAMILY_ID = ?
	`,
		latCol,
		lngCol,
	)

	var house HouseRecord
	err := h.DB.QueryRow(query, id).Scan(
		&house.FamilyID,
		&house.ExternalFamilyID,
		&house.HouseNo,
		&house.DistrictID, &house.DistrictName,
		&house.TalukaID, &house.TalukaName,
		&house.VillageID, &house.VillageName,
		&house.Latitude, &house.Longitude,
		&house.TotalLand, &house.CultivatedLand, &house.OwnLand,
		&house.WaterSource, &house.Kharif, &house.Rabi,
		&house.Latrine, &house.Lighting, &house.RationCard,
		&house.Occupation,
		&house.HeadName,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "house not found"})
		return
	}

	memberRows, err := h.DB.Query(`
		SELECT COALESCE(FIRST_NAME, ''), COALESCE(LAST_NAME, '')
		FROM FAMILY_MEMBER
		WHERE EXTERNAL_FAMILY_ID = ?
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch members"})
		return
	}
	defer memberRows.Close()

	var members []MemberRecord
	for memberRows.Next() {
		var m MemberRecord
		memberRows.Scan(&m.FirstName, &m.LastName)
		members = append(members, m)
	}
	if members == nil {
		members = []MemberRecord{}
	}

	c.JSON(http.StatusOK, HouseDetail{
		HouseRecord: house,
		Members:     members,
	})
}
