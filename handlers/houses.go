package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

func parseBBoxParam(raw string) (minLng, minLat, maxLng, maxLat float64, ok bool, err error) {
	parts := strings.Split(raw, ",")
	if len(parts) != 4 {
		return 0, 0, 0, 0, false, fmt.Errorf("bbox must be minLng,minLat,maxLng,maxLat")
	}

	vals := make([]float64, 4)
	for i := range parts {
		v, convErr := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		if convErr != nil {
			return 0, 0, 0, 0, false, fmt.Errorf("invalid bbox value %q", parts[i])
		}
		vals[i] = v
	}

	minLng, minLat, maxLng, maxLat = vals[0], vals[1], vals[2], vals[3]
	if minLng > maxLng {
		minLng, maxLng = maxLng, minLng
	}
	if minLat > maxLat {
		minLat, maxLat = maxLat, minLat
	}

	return minLng, minLat, maxLng, maxLat, true, nil
}

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

	// Population aggregate fields
	TotalMembers      int    `json:"totalMembers"`
	MaleMembers       int    `json:"maleMembers"`
	FemaleMembers     int    `json:"femaleMembers"`
	WorkingMembers    int    `json:"workingMembers"`
	IlliterateMembers int    `json:"illiterateMembers"`
	DivyangMembers    int    `json:"divyangMembers"`
	UnemployedMembers int    `json:"unemployedMembers"`
	BplCategory       string `json:"bplCategory"`
	AnnualIncome      string `json:"annualIncome"`
}

type HouseDetail struct {
	HouseRecord
	Members []MemberRecord `json:"members"`
}

type HouseMapPoint struct {
	ID  int     `json:"id"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type MemberRecord struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type HouseHandler struct {
	DB *sql.DB
	CC *ColumnChecker

	memberColCacheMu sync.RWMutex
	memberColCache   map[string]bool
}

func (h *HouseHandler) memberColExists(col string) bool {
	h.memberColCacheMu.RLock()
	if h.memberColCache != nil {
		if v, ok := h.memberColCache[col]; ok {
			h.memberColCacheMu.RUnlock()
			return v
		}
	}
	h.memberColCacheMu.RUnlock()

	var n int
	_ = h.DB.QueryRow(`SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='FAMILY_MEMBER' AND COLUMN_NAME=?`, col).Scan(&n)
	exists := n > 0

	h.memberColCacheMu.Lock()
	if h.memberColCache == nil {
		h.memberColCache = make(map[string]bool)
	}
	h.memberColCache[col] = exists
	h.memberColCacheMu.Unlock()

	return exists
}

// buildPopStatsSQL returns a SQL subquery that aggregates per-family member stats.
// It detects optional FAMILY_MEMBER columns (DIVYANG, DISABILITY, EVER_ATTENDED_SCHOOL)
// and falls back to 0 when they are absent, so the main query never errors.
func (h *HouseHandler) buildPopStatsSQL() string {
	hasExternalFamilyID := h.memberColExists("EXTERNAL_FAMILY_ID")
	hasFamilyID := h.memberColExists("FAMILY_ID")

	familyJoinExpr := "CAST(fm.EXTERNAL_FAMILY_ID AS CHAR)"
	switch {
	case hasExternalFamilyID && hasFamilyID:
		familyJoinExpr = "CAST(COALESCE(fm.EXTERNAL_FAMILY_ID, fm.FAMILY_ID) AS CHAR)"
	case hasFamilyID:
		familyJoinExpr = "CAST(fm.FAMILY_ID AS CHAR)"
	case hasExternalFamilyID:
		familyJoinExpr = "CAST(fm.EXTERNAL_FAMILY_ID AS CHAR)"
	}

	illiterateExpr := "0"
	if h.memberColExists("EVER_ATTENDED_SCHOOL") {
		illiterateExpr = "SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL,'')))='NO' THEN 1 ELSE 0 END)"
	}

	divyangExpr := "0"
	if h.memberColExists("DIVYANG") {
		if h.memberColExists("DISABILITY") {
			divyangExpr = "SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.DIVYANG,'')))='YES' OR UPPER(TRIM(COALESCE(fm.DISABILITY,'')))='YES' THEN 1 ELSE 0 END)"
		} else {
			divyangExpr = "SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.DIVYANG,'')))='YES' THEN 1 ELSE 0 END)"
		}
	}

	// Use NATURE_WAGE_WORK first, fall back to OCCUPATION
	workExprBase := "COALESCE(fm.NATURE_WAGE_WORK, fm.OCCUPATION, '')"
	if !h.memberColExists("NATURE_WAGE_WORK") {
		workExprBase = "COALESCE(fm.OCCUPATION, '')"
	}

	workingExpr := fmt.Sprintf(
		"SUM(CASE WHEN UPPER(TRIM(%s)) NOT IN ('','UNEMPLOYED','NOT WORKING','NO WORK','HOUSEWIFE','HOMEMAKER') THEN 1 ELSE 0 END)",
		workExprBase)
	unemployedExpr := fmt.Sprintf(
		"SUM(CASE WHEN UPPER(TRIM(%s)) IN ('','UNEMPLOYED','NOT WORKING','NO WORK') THEN 1 ELSE 0 END)",
		workExprBase)

	var occExpr string
	if h.memberColExists("NATURE_WAGE_WORK") {
		occExpr = "COALESCE(MAX(NULLIF(TRIM(COALESCE(fm.NATURE_WAGE_WORK,'')),'')), MAX(NULLIF(TRIM(COALESCE(fm.OCCUPATION,'')),'')), '')"
	} else {
		occExpr = "COALESCE(MAX(NULLIF(TRIM(COALESCE(fm.OCCUPATION,'')),'')), '')"
	}

	return fmt.Sprintf(`
		SELECT %s AS family_join_id,
			%s AS primary_occupation,
			COUNT(*) AS total_members,
			SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER,''))) IN ('male','m') THEN 1 ELSE 0 END) AS male_members,
			SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER,''))) IN ('female','f') THEN 1 ELSE 0 END) AS female_members,
			%s AS working_members,
			%s AS illiterate_members,
			%s AS divyang_members,
			%s AS unemployed_members
		FROM FAMILY_MEMBER fm
		GROUP BY %s`,
		familyJoinExpr,
		occExpr, workingExpr, illiterateExpr, divyangExpr, unemployedExpr,
		familyJoinExpr)
}

func (h *HouseHandler) GetHouses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "500"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 5000 {
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

	bboxRaw := strings.TrimSpace(c.Query("bbox"))
	if bboxRaw != "" {
		minLng, minLat, maxLng, maxLat, ok, parseErr := parseBBoxParam(bboxRaw)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bbox", "detail": parseErr.Error()})
			return
		}
		if ok {
			where += fmt.Sprintf(" AND f.%s > ? AND f.%s < ? AND f.%s > ? AND f.%s < ?", latCol, latCol, lngCol, lngCol)
			args = append(args, minLat, maxLat, minLng, maxLng)
		}
	} else {
		minLatRaw := strings.TrimSpace(c.Query("min_lat"))
		maxLatRaw := strings.TrimSpace(c.Query("max_lat"))
		minLngRaw := strings.TrimSpace(c.Query("mLalit Sir pan bye-byein_lng"))
		maxLngRaw := strings.TrimSpace(c.Query("max_lng"))
		if minLatRaw != "" && maxLatRaw != "" && minLngRaw != "" && maxLngRaw != "" {
			minLat, errMinLat := strconv.ParseFloat(minLatRaw, 64)
			maxLat, errMaxLat := strconv.ParseFloat(maxLatRaw, 64)
			minLng, errMinLng := strconv.ParseFloat(minLngRaw, 64)
			maxLng, errMaxLng := strconv.ParseFloat(maxLngRaw, 64)
			if errMinLat != nil || errMaxLat != nil || errMinLng != nil || errMaxLng != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid viewport bounds"})
				return
			}
			if minLat > maxLat {
				minLat, maxLat = maxLat, minLat
			}
			if minLng > maxLng {
				minLng, maxLng = maxLng, minLng
			}
			where += fmt.Sprintf(" AND f.%s > ? AND f.%s < ? AND f.%s > ? AND f.%s < ?", latCol, latCol, lngCol, lngCol)
			args = append(args, minLat, maxLat, minLng, maxLng)
		}
	}

	// Build population stats subquery (detects optional FAMILY_MEMBER columns at runtime)
	popStatsSQL := h.buildPopStatsSQL()

	// FAMILY-level optional columns for population context
	bplExpr := h.CC.ColOrEmpty("FAMILY_BELONG_BPL_CATEGORY", "bpl_category")
	incomeExpr := h.CC.ColOrEmpty("ANNUAL_INCOME", "annual_income")

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
			COALESCE(fm_agg_ext.primary_occupation, fm_agg_fid.primary_occupation, ''),
			COALESCE(TRIM(CONCAT(
				COALESCE(f.FIRST_NAME_HOUSEHOLD_HEAD, ''), ' ',
				COALESCE(f.MIDDLE_NAME_HOUSEHOLD_HEAD, ''), ' ',
				COALESCE(f.LAST_NAME_HOUSEHOLD_HEAD, '')
			)), ''),
			COALESCE(fm_agg_ext.total_members, fm_agg_fid.total_members, 0),
			COALESCE(fm_agg_ext.male_members, fm_agg_fid.male_members, 0),
			COALESCE(fm_agg_ext.female_members, fm_agg_fid.female_members, 0),
			COALESCE(fm_agg_ext.working_members, fm_agg_fid.working_members, 0),
			COALESCE(fm_agg_ext.illiterate_members, fm_agg_fid.illiterate_members, 0),
			COALESCE(fm_agg_ext.divyang_members, fm_agg_fid.divyang_members, 0),
			COALESCE(fm_agg_ext.unemployed_members, fm_agg_fid.unemployed_members, 0),
			%s,
			%s
		FROM FAMILY f
		LEFT JOIN (%s) fm_agg_fid ON fm_agg_fid.family_join_id = CAST(f.FAMILY_ID AS CHAR)
		LEFT JOIN (%s) fm_agg_ext ON fm_agg_ext.family_join_id = CAST(COALESCE(f.EXTERNAL_FAMILY_ID, f.FAMILY_ID) AS CHAR)
		LEFT JOIN district_master dm ON dm.pklDistrictId = f.DISTRICT_ID
		LEFT JOIN taluka_master tm ON tm.pklTalukaId = f.TALUKA_ID
		LEFT JOIN village_master vm ON vm.pklVillageId = f.VILLAGE_ID
		%s
		ORDER BY f.FAMILY_ID
		LIMIT %d OFFSET %d
	`,
		latCol,
		lngCol,
		bplExpr,
		incomeExpr,
		popStatsSQL,
		popStatsSQL,
		where,
		limit,
		offset,
	)

	log.Println("HOUSES QUERY:", query)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		log.Println("HOUSES QUERY ERROR:", err)
		c.JSON(http.StatusOK, gin.H{
			"data":  []HouseRecord{},
			"total": 0,
			"page":  page,
			"limit": limit,
		})
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
			&house.Occupation, &house.HeadName,
			&house.TotalMembers, &house.MaleMembers, &house.FemaleMembers,
			&house.WorkingMembers, &house.IlliterateMembers,
			&house.DivyangMembers, &house.UnemployedMembers,
			&house.BplCategory, &house.AnnualIncome,
		); err != nil {
			log.Println("SCAN ERROR:", err)
			c.JSON(http.StatusOK, gin.H{
				"data":  []HouseRecord{},
				"total": 0,
				"page":  page,
				"limit": limit,
			})
			return
		}
		houses = append(houses, house)
	}
	if err := rows.Err(); err != nil {
		log.Println("ROWS ERROR:", err)
		c.JSON(http.StatusOK, gin.H{
			"data":  []HouseRecord{},
			"total": 0,
			"page":  page,
			"limit": limit,
		})
		return
	}

	if houses == nil {
		houses = []HouseRecord{}
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM FAMILY f %s", where)
	var total int
	if err := h.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		log.Println("COUNT QUERY ERROR:", err)
		total = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  houses,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetHousesMapPoints — GET /houses/map-points
// Returns lightweight household coordinates for fast client-side clustering.
func (h *HouseHandler) GetHousesMapPoints(c *gin.Context) {
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

	if districtID := strings.TrimSpace(c.Query("district_id")); districtID != "" {
		where += " AND CAST(f.DISTRICT_ID AS CHAR) = ?"
		args = append(args, districtID)
	}
	if talukaID := strings.TrimSpace(c.Query("taluka_id")); talukaID != "" {
		where += " AND CAST(f.TALUKA_ID AS CHAR) = ?"
		args = append(args, talukaID)
	}
	if villageID := strings.TrimSpace(c.Query("village_id")); villageID != "" {
		where += " AND CAST(f.VILLAGE_ID AS CHAR) = ?"
		args = append(args, villageID)
	}

	query := fmt.Sprintf(`
		SELECT
			f.FAMILY_ID,
			COALESCE(f.%s, 0),
			COALESCE(f.%s, 0)
		FROM FAMILY f
		%s
		ORDER BY f.FAMILY_ID
	`, latCol, lngCol, where)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch map points", "detail": err.Error()})
		return
	}
	defer rows.Close()

	points := make([]HouseMapPoint, 0, 4096)
	for rows.Next() {
		var point HouseMapPoint
		if err := rows.Scan(&point.ID, &point.Lat, &point.Lng); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan map point", "detail": err.Error()})
			return
		}
		points = append(points, point)
	}
	if points == nil {
		points = []HouseMapPoint{}
	}

	c.JSON(http.StatusOK, points)
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

	popStatsSQL := h.buildPopStatsSQL()
	bplExpr := h.CC.ColOrEmpty("FAMILY_BELONG_BPL_CATEGORY", "bpl_category")
	incomeExpr := h.CC.ColOrEmpty("ANNUAL_INCOME", "annual_income")

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
			COALESCE(fm_agg_ext.primary_occupation, fm_agg_fid.primary_occupation, ''),
			COALESCE(
				TRIM(CONCAT(
					COALESCE(f.FIRST_NAME_HOUSEHOLD_HEAD, ''), ' ',
					COALESCE(f.MIDDLE_NAME_HOUSEHOLD_HEAD, ''), ' ',
					COALESCE(f.LAST_NAME_HOUSEHOLD_HEAD, '')
				)), ''),
			COALESCE(fm_agg_ext.total_members, fm_agg_fid.total_members, 0),
			COALESCE(fm_agg_ext.male_members, fm_agg_fid.male_members, 0),
			COALESCE(fm_agg_ext.female_members, fm_agg_fid.female_members, 0),
			COALESCE(fm_agg_ext.working_members, fm_agg_fid.working_members, 0),
			COALESCE(fm_agg_ext.illiterate_members, fm_agg_fid.illiterate_members, 0),
			COALESCE(fm_agg_ext.divyang_members, fm_agg_fid.divyang_members, 0),
			COALESCE(fm_agg_ext.unemployed_members, fm_agg_fid.unemployed_members, 0),
			%s,
			%s
		FROM FAMILY f
		LEFT JOIN (%s) fm_agg_fid ON fm_agg_fid.family_join_id = CAST(f.FAMILY_ID AS CHAR)
		LEFT JOIN (%s) fm_agg_ext ON fm_agg_ext.family_join_id = CAST(COALESCE(f.EXTERNAL_FAMILY_ID, f.FAMILY_ID) AS CHAR)
		LEFT JOIN district_master dm ON dm.pklDistrictId = f.DISTRICT_ID
		LEFT JOIN taluka_master tm ON tm.pklTalukaId = f.TALUKA_ID
		LEFT JOIN village_master vm ON vm.pklVillageId = f.VILLAGE_ID
		WHERE f.FAMILY_ID = ?
	`,
		latCol,
		lngCol,
		bplExpr,
		incomeExpr,
		popStatsSQL,
		popStatsSQL,
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
		&house.Occupation, &house.HeadName,
		&house.TotalMembers, &house.MaleMembers, &house.FemaleMembers,
		&house.WorkingMembers, &house.IlliterateMembers,
		&house.DivyangMembers, &house.UnemployedMembers,
		&house.BplCategory, &house.AnnualIncome,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "house not found"})
		return
	}

	memberWhere := []string{}
	memberArgs := []interface{}{}
	hasExternalFamilyID := h.memberColExists("EXTERNAL_FAMILY_ID")
	hasFamilyID := h.memberColExists("FAMILY_ID")

	if hasExternalFamilyID {
		memberWhere = append(memberWhere, "CAST(EXTERNAL_FAMILY_ID AS CHAR) = ?")
	}
	if hasFamilyID {
		memberWhere = append(memberWhere, "CAST(FAMILY_ID AS CHAR) = ?")
	}

	memberExternalID := id
	memberFamilyID := id
	if hasExternalFamilyID && h.CC != nil && h.CC.Has("EXTERNAL_FAMILY_ID") {
		_ = h.DB.QueryRow("SELECT CAST(COALESCE(EXTERNAL_FAMILY_ID, FAMILY_ID) AS CHAR) FROM FAMILY WHERE FAMILY_ID = ?", id).Scan(&memberExternalID)
	}

	if hasExternalFamilyID {
		memberArgs = append(memberArgs, memberExternalID, memberFamilyID)
	}
	if hasFamilyID {
		memberArgs = append(memberArgs, memberFamilyID)
	}

	if len(memberWhere) == 0 {
		c.JSON(http.StatusOK, HouseDetail{
			HouseRecord: house,
			Members:     []MemberRecord{},
		})
		return
	}

	memberQuery := fmt.Sprintf(`
		SELECT COALESCE(FIRST_NAME, ''), COALESCE(LAST_NAME, '')
		FROM FAMILY_MEMBER
		WHERE %s
	`, strings.Join(memberWhere, " OR "))

	memberRows, err := h.DB.Query(memberQuery, memberArgs...)
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

// GetHousesSummary — GET /houses/summary
// Returns a grid-aggregated count of households within a bounding box.
// Query params: min_lat, max_lat, min_lng, max_lng (required), grid (degrees, default 0.01)
func (h *HouseHandler) GetHousesSummary(c *gin.Context) {
	minLat := c.Query("min_lat")
	maxLat := c.Query("max_lat")
	minLng := c.Query("min_lng")
	maxLng := c.Query("max_lng")
	grid := c.DefaultQuery("grid", "0.01")

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

	if minLat != "" && maxLat != "" && minLng != "" && maxLng != "" {
		where += fmt.Sprintf(" AND f.%s BETWEEN ? AND ? AND f.%s BETWEEN ? AND ?", latCol, lngCol)
		args = append(args, minLat, maxLat, minLng, maxLng)
	}

	query := fmt.Sprintf(`
		SELECT
			ROUND(f.%s / %s) * %s AS cell_lat,
			ROUND(f.%s / %s) * %s AS cell_lng,
			COUNT(*) AS cnt
		FROM FAMILY f
		%s
		GROUP BY cell_lat, cell_lng
		ORDER BY cnt DESC
		LIMIT 5000
	`, latCol, grid, grid, lngCol, grid, grid, where)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "summary query failed", "detail": err.Error()})
		return
	}
	defer rows.Close()

	type SummaryCell struct {
		Lat   float64 `json:"lat"`
		Lng   float64 `json:"lng"`
		Count int     `json:"count"`
	}
	var cells []SummaryCell
	for rows.Next() {
		var cell SummaryCell
		rows.Scan(&cell.Lat, &cell.Lng, &cell.Count)
		cells = append(cells, cell)
	}
	if cells == nil {
		cells = []SummaryCell{}
	}
	c.JSON(http.StatusOK, gin.H{"cells": cells})
}
