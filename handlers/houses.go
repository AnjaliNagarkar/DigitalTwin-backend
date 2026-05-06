package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
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
	FamilyID                     int     `json:"familyId"`
	ExternalFamilyID             string  `json:"externalFamilyId"`
	HouseNo                      string  `json:"houseNo"`
	DistrictID                   string  `json:"districtId"`
	DistrictName                 string  `json:"districtName"`
	TalukaID                     string  `json:"talukaId"`
	TalukaName                   string  `json:"talukaName"`
	VillageID                    string  `json:"villageId"`
	VillageName                  string  `json:"villageName"`
	Latitude                     float64 `json:"latitude"`
	Longitude                    float64 `json:"longitude"`
	TotalLand                    string  `json:"totalLand"`
	CultivatedLand               string  `json:"cultivatedLand"`
	OwnLand                      string  `json:"ownLand"`
	WaterSource                  string  `json:"waterSource"`
	Kharif                       string  `json:"kharif"`
	Rabi                         string  `json:"rabi"`
	TypeHouse                    string  `json:"TYPE_HOUSE"`
	OwnershipHouse               string  `json:"OWNERSHIP_HOUSE"`
	PradhanMantriAwas            string  `json:"PRADHAN_MANTRI_AWAS"`
	SanitationToiletFacilityHome string  `json:"SANITATION_TOILET_FACILITY_HOME"`
	ASoakpitManagingWastewater   string  `json:"A_SOAKPIT_MANAGING_WASTEWATER"`
	RationCardColor              string  `json:"RATION_CARD_COLOR"`
	AadhaarCoverageStatus        string  `json:"aadhaarCoverageStatus"`
	MembersWithAadhaar           int     `json:"membersWithAadhaar"`
	TotalFamilyMembers           int     `json:"totalFamilyMembers"`
	Latrine                      string  `json:"latrine"`
	Lighting                     string  `json:"lighting"`
	RationCard                   string  `json:"rationCard"`
	Occupation                   string  `json:"occupation"`
	HeadName                     string  `json:"headName"`

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

var houseCache sync.Map

func (h *HouseHandler) firstNonEmptyExpr(columns ...string) string {
	parts := make([]string, 0, len(columns))
	for _, col := range columns {
		if h.CC != nil && h.CC.Has(col) {
			parts = append(parts, fmt.Sprintf("NULLIF(TRIM(COALESCE(f.%s, '')), '')", col))
		}
	}
	if len(parts) == 0 {
		return "''"
	}
	return fmt.Sprintf("COALESCE(%s, '')", strings.Join(parts, ", "))
}

func (h *HouseHandler) yesNoExpr(columns ...string) string {
	parts := make([]string, 0, len(columns))
	for _, col := range columns {
		if h.CC != nil && h.CC.Has(col) {
			parts = append(parts, fmt.Sprintf("NULLIF(TRIM(COALESCE(f.%s, '')), '')", col))
		}
	}
	if len(parts) == 0 {
		return "''"
	}
	raw := fmt.Sprintf("LOWER(TRIM(COALESCE(%s, '')))", strings.Join(parts, ", "))
	return fmt.Sprintf(`CASE
		WHEN %s IN ('', '0', 'false', 'no', 'none', 'na', 'n/a', 'kerosene', 'no electricity', 'no lighting') THEN 'No'
		ELSE 'Yes'
	END`, raw)
}

func (h *HouseHandler) memberFirstNonEmptyExpr(columns ...string) string {
	parts := make([]string, 0, len(columns))
	for _, col := range columns {
		if h.memberColExists(col) {
			parts = append(parts, fmt.Sprintf("NULLIF(TRIM(COALESCE(fm.%s, '')), '')", col))
		}
	}
	if len(parts) == 0 {
		return "''"
	}
	return fmt.Sprintf("COALESCE(%s, '')", strings.Join(parts, ", "))
}

func (h *HouseHandler) memberYesNoExpr(columns ...string) string {
	parts := make([]string, 0, len(columns))
	for _, col := range columns {
		if h.memberColExists(col) {
			parts = append(parts, fmt.Sprintf("NULLIF(TRIM(COALESCE(fm.%s, '')), '')", col))
		}
	}
	if len(parts) == 0 {
		return "''"
	}
	raw := fmt.Sprintf("LOWER(TRIM(COALESCE(%s, '')))", strings.Join(parts, ", "))
	return fmt.Sprintf(`CASE
		WHEN %s IN ('', '0', 'false', 'no', 'none', 'na', 'n/a') THEN 'No'
		ELSE 'Yes'
	END`, raw)
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
	familyJoinExpr := "CAST(fm.EXTERNAL_FAMILY_ID AS CHAR)"
	if !hasExternalFamilyID {
		familyJoinExpr = "''"
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

func (h *HouseHandler) buildAadhaarCoverageSQL() string {
	hasExternalFamilyID := h.memberColExists("EXTERNAL_FAMILY_ID")
	familyJoinExpr := "CAST(fm.EXTERNAL_FAMILY_ID AS CHAR)"
	if !hasExternalFamilyID {
		familyJoinExpr = "''"
	}

	aadhaarAvailableExpr := "SUM(CASE WHEN fm.AADHAAR IS NOT NULL AND TRIM(fm.AADHAAR) <> '' THEN 1 ELSE 0 END)"

	return fmt.Sprintf(`
		SELECT %s AS family_join_id,
			COUNT(*) AS total_family_members,
			%s AS members_with_aadhaar,
			CASE
				WHEN COUNT(*) = 0 THEN 'unknown'
				WHEN %s = COUNT(*) THEN 'complete'
				WHEN %s > 0 THEN 'partial'
				ELSE 'missing'
			END AS aadhaar_coverage_status
		FROM FAMILY_MEMBER fm
		GROUP BY %s`,
		familyJoinExpr,
		aadhaarAvailableExpr,
		aadhaarAvailableExpr,
		aadhaarAvailableExpr,
		familyJoinExpr)
}

// appendFamilyGeoIDFilter uses equality on numeric ID columns when possible so indexes on
// DISTRICT_ID / TALUKA_ID / VILLAGE_ID can be used (avoids CAST(... AS CHAR) = ?).
func appendFamilyGeoIDFilter(where *string, args *[]interface{}, alias, column, raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}
	if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
		*where += fmt.Sprintf(" AND %s.%s = ?", alias, column)
		*args = append(*args, id)
		return
	}
	*where += fmt.Sprintf(" AND CAST(%s.%s AS CHAR) = ?", alias, column)
	*args = append(*args, raw)
}
func (h *HouseHandler) buildHouseCacheQuery() string {
	latCol := h.CC.LatCol
	lngCol := h.CC.LngCol
	if latCol == "" {
		latCol = "LATITUDE"
	}
	if lngCol == "" {
		lngCol = "LONGITUDE"
	}

	sanitationExpr := h.firstNonEmptyExpr("TYPE_OF_LATRINE", "SANITATION_TOILET_FACILITY")
	lightingExpr := h.yesNoExpr("SOURCE_OF_LIGHTING", "ELECTRICITY_CONNECTION")
	rationExpr := h.firstNonEmptyExpr("TYPE_OF_RATION_CARD", "RATION_CARD_TYPE")
	bplExpr := h.CC.ColOrEmpty("FAMILY_BELONG_BPL_CATEGORY", "bpl_category")
	incomeExpr := h.CC.ColOrEmpty("ANNUAL_INCOME", "annual_income")

	// Build head name from household head columns (matches GetHouses)
	headNameExpr := `COALESCE(TRIM(CONCAT(
		COALESCE(f.FIRST_NAME_HOUSEHOLD_HEAD, ''), ' ',
		COALESCE(f.MIDDLE_NAME_HOUSEHOLD_HEAD, ''), ' ',
		COALESCE(f.LAST_NAME_HOUSEHOLD_HEAD, '')
	)), '')`

	// Build family member aggregation (matches GetHouses dual-join strategy)
	popStatsSQL := h.buildPopStatsSQL()
	aadhaarCoverageSQL := h.buildAadhaarCoverageSQL()

	return fmt.Sprintf(`
		SELECT
			f.FAMILY_ID,
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
			COALESCE(f.DRINKING_WATER_SOURCE, ''),
			COALESCE(f.CULTIVATING_DURING_KHARIF_SEASON, ''),
			COALESCE(f.TAKING_CROPS_RABI_SEASON, ''),
			COALESCE(f.TYPE_HOUSE, ''),
			COALESCE(f.OWNERSHIP_HOUSE, ''),
			COALESCE(f.PRADHAN_MANTRI_AWAS, ''),
			COALESCE(f.SANITATION_TOILET_FACILITY_HOME, ''),
			COALESCE(f.A_SOAKPIT_MANAGING_WASTEWATER, ''),
			COALESCE(f.RATION_CARD_COLOR, ''),
			COALESCE(aadhaar_agg.aadhaar_coverage_status, 'unknown'),
			COALESCE(aadhaar_agg.members_with_aadhaar, 0),
			COALESCE(aadhaar_agg.total_family_members, 0),
			%s,
			%s,
			%s,
			COALESCE(fm_agg_ext.primary_occupation, fm_agg_fid.primary_occupation, ''),
			%s,
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
		LEFT JOIN (%s) aadhaar_agg ON aadhaar_agg.family_join_id = CAST(f.EXTERNAL_FAMILY_ID AS CHAR)
		LEFT JOIN district_master dm ON dm.pklDistrictId = f.DISTRICT_ID
		LEFT JOIN taluka_master tm ON tm.pklTalukaId = f.TALUKA_ID
		LEFT JOIN village_master vm ON vm.pklVillageId = f.VILLAGE_ID
		ORDER BY f.FAMILY_ID
	`, latCol, lngCol, sanitationExpr, lightingExpr, rationExpr, headNameExpr, bplExpr, incomeExpr, popStatsSQL, popStatsSQL, aadhaarCoverageSQL)
}

func (h *HouseHandler) PreloadHouseCache() error {
	houseCache = sync.Map{}

	rows, err := h.DB.Query(h.buildHouseCacheQuery())
	if err != nil {
		return err
	}
	defer rows.Close()

	details := make(map[int]*HouseDetail, 4096)
	for rows.Next() {
		var detail HouseDetail
		if err := rows.Scan(
			&detail.FamilyID,
			&detail.DistrictID, &detail.DistrictName,
			&detail.TalukaID, &detail.TalukaName,
			&detail.VillageID, &detail.VillageName,
			&detail.Latitude, &detail.Longitude,
			&detail.TotalLand, &detail.CultivatedLand, &detail.OwnLand,
			&detail.WaterSource, &detail.Kharif, &detail.Rabi,
			&detail.TypeHouse,
			&detail.OwnershipHouse, &detail.PradhanMantriAwas, &detail.SanitationToiletFacilityHome, &detail.ASoakpitManagingWastewater, &detail.RationCardColor,
			&detail.AadhaarCoverageStatus, &detail.MembersWithAadhaar, &detail.TotalFamilyMembers,
			&detail.Latrine, &detail.Lighting, &detail.RationCard,
			&detail.Occupation, &detail.HeadName,
			&detail.TotalMembers, &detail.MaleMembers, &detail.FemaleMembers,
			&detail.WorkingMembers, &detail.IlliterateMembers, &detail.DivyangMembers,
			&detail.UnemployedMembers,
			&detail.BplCategory, &detail.AnnualIncome,
		); err != nil {
			return err
		}
		detail.Members = []MemberRecord{}
		details[detail.FamilyID] = &detail
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Query individual members using dual-join strategy (EXTERNAL_FAMILY_ID preferred, fallback to FAMILY_ID)
	memberFirstNameExpr := "COALESCE(fm.FIRST_NAME, '')"
	memberLastNameExpr := "COALESCE(fm.LAST_NAME, '')"
	if !h.memberColExists("FIRST_NAME") {
		memberFirstNameExpr = "''"
	}
	if !h.memberColExists("LAST_NAME") {
		memberLastNameExpr = "''"
	}

	memberQuery := fmt.Sprintf(`
		SELECT
			CAST(fm.EXTERNAL_FAMILY_ID AS CHAR),
			%s,
			%s
		FROM FAMILY_MEMBER fm
		WHERE fm.EXTERNAL_FAMILY_ID IS NOT NULL
		ORDER BY fm.EXTERNAL_FAMILY_ID, fm.FAMILY_MEMBER_ID
	`, memberFirstNameExpr, memberLastNameExpr)

	memberRows, err := h.DB.Query(memberQuery)
	if err != nil {
		return err
	}
	defer memberRows.Close()

	for memberRows.Next() {
		var familyIDRaw string
		var firstName string
		var lastName string
		if err := memberRows.Scan(&familyIDRaw, &firstName, &lastName); err != nil {
			continue
		}

		familyID, convErr := strconv.Atoi(strings.TrimSpace(familyIDRaw))
		if convErr != nil {
			continue
		}

		detail := details[familyID]
		if detail == nil {
			continue
		}

		detail.Members = append(detail.Members, MemberRecord{
			FirstName: strings.TrimSpace(firstName),
			LastName:  strings.TrimSpace(lastName),
		})
	}
	if err := memberRows.Err(); err != nil {
		return err
	}

	for id, detail := range details {
		if detail.Members == nil {
			detail.Members = []MemberRecord{}
		}
		houseCache.Store(id, detail)
	}

	return nil
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
	appendFamilyGeoIDFilter(&where, &args, "f", "DISTRICT_ID", districtID)
	appendFamilyGeoIDFilter(&where, &args, "f", "TALUKA_ID", talukaID)
	appendFamilyGeoIDFilter(&where, &args, "f", "VILLAGE_ID", villageID)

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

	log.Println("REQUEST PARAMS:",
		"district_id =", districtID,
		"taluka_id =", talukaID,
		"village_id =", villageID,
	)
	log.Println("WHERE:", where)
	log.Println("ARGS:", args)

	// FAMILY-level optional columns (no FAMILY_MEMBER aggregation in this handler — MapView enriches from population API).
	bplExpr := h.CC.ColOrEmpty("FAMILY_BELONG_BPL_CATEGORY", "bpl_category")
	incomeExpr := h.CC.ColOrEmpty("ANNUAL_INCOME", "annual_income")
	sanitationExpr := h.firstNonEmptyExpr("TYPE_OF_LATRINE", "SANITATION_TOILET_FACILITY")
	lightingExpr := h.yesNoExpr("SOURCE_OF_LIGHTING", "ELECTRICITY_CONNECTION")
	rationExpr := h.firstNonEmptyExpr("TYPE_OF_RATION_CARD", "RATION_CARD_TYPE")
	aadhaarCoverageSQL := h.buildAadhaarCoverageSQL()

	// Build member aggregation expressions with optional-column safety.
	memberOccExpr := "COALESCE(MAX(NULLIF(TRIM(COALESCE(fm.OCCUPATION,'')),'')), '')"
	if h.memberColExists("NATURE_WAGE_WORK") {
		memberOccExpr = "COALESCE(MAX(NULLIF(TRIM(COALESCE(fm.NATURE_WAGE_WORK,'')),'')), MAX(NULLIF(TRIM(COALESCE(fm.OCCUPATION,'')),'')), '')"
	}

	memberWorkingExpr := "SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.OCCUPATION, ''))) NOT IN ('','UNEMPLOYED','NOT WORKING','NO WORK','HOUSEWIFE','HOMEMAKER','NOT APPLICABLE','STUDYING') THEN 1 ELSE 0 END)"
	memberUnemployedExpr := "SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.OCCUPATION, ''))) IN ('','UNEMPLOYED','NOT WORKING','NO WORK','HOUSEWIFE','HOMEMAKER','NOT APPLICABLE','STUDYING') THEN 1 ELSE 0 END)"
	memberIlliterateExpr := "0"
	if h.memberColExists("EVER_ATTENDED_SCHOOL") {
		memberIlliterateExpr = "SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL,'')))='NO' THEN 1 ELSE 0 END)"
	}
	memberDivyangExpr := "0"
	if h.memberColExists("DIVYANG") {
		if h.memberColExists("DISABILITY") {
			memberDivyangExpr = "SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.DIVYANG,'')))='YES' OR UPPER(TRIM(COALESCE(fm.DISABILITY,'')))='YES' THEN 1 ELSE 0 END)"
		} else {
			memberDivyangExpr = "SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.DIVYANG,'')))='YES' THEN 1 ELSE 0 END)"
		}
	}

	// Reuse the exact FAMILY filter logic for aggregation pre-filtering.
	aggFamilyWhere := strings.ReplaceAll(where, "f.", "f2.")

	// Single query: FAMILY + location masters + one filtered FAMILY_MEMBER aggregate join.
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
			COALESCE(f.DRINKING_WATER_SOURCE, ''),
			COALESCE(f.CULTIVATING_DURING_KHARIF_SEASON, ''),
			COALESCE(f.TAKING_CROPS_RABI_SEASON, ''),
			COALESCE(f.TYPE_HOUSE, ''),
			COALESCE(f.OWNERSHIP_HOUSE, ''),
			COALESCE(f.PRADHAN_MANTRI_AWAS, ''),
			COALESCE(f.SANITATION_TOILET_FACILITY_HOME, ''),
			COALESCE(f.A_SOAKPIT_MANAGING_WASTEWATER, ''),
			COALESCE(f.RATION_CARD_COLOR, ''),
			COALESCE(aadhaar_agg.aadhaar_coverage_status, 'unknown'),
			COALESCE(aadhaar_agg.members_with_aadhaar, 0),
			COALESCE(aadhaar_agg.total_family_members, 0),
			%s,
			%s,
			%s,
			COALESCE(fm_agg.primary_occupation, ''),
			COALESCE(TRIM(CONCAT(
				COALESCE(f.FIRST_NAME_HOUSEHOLD_HEAD, ''), ' ',
				COALESCE(f.MIDDLE_NAME_HOUSEHOLD_HEAD, ''), ' ',
				COALESCE(f.LAST_NAME_HOUSEHOLD_HEAD, '')
			)), ''),
			COALESCE(fm_agg.total_members, 0),
			COALESCE(fm_agg.male_members, 0),
			COALESCE(fm_agg.female_members, 0),
			COALESCE(fm_agg.working_members, 0),
			COALESCE(fm_agg.illiterate_members, 0),
			COALESCE(fm_agg.divyang_members, 0),
			COALESCE(fm_agg.unemployed_members, 0),
			%s,
			%s
		FROM FAMILY f
		LEFT JOIN (
			SELECT
				CAST(fm.EXTERNAL_FAMILY_ID AS CHAR) AS family_join_id,
				%s AS primary_occupation,
				COUNT(*) AS total_members,
				SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER,''))) IN ('male','m') THEN 1 ELSE 0 END) AS male_members,
				SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER,''))) IN ('female','f') THEN 1 ELSE 0 END) AS female_members,
				%s AS working_members,
				%s AS illiterate_members,
				%s AS divyang_members,
				%s AS unemployed_members
			FROM FAMILY_MEMBER fm
			WHERE CAST(fm.EXTERNAL_FAMILY_ID AS CHAR) IN (
				SELECT CAST(COALESCE(f2.EXTERNAL_FAMILY_ID, '') AS CHAR)
				FROM FAMILY f2
				%s
			)
			GROUP BY CAST(fm.EXTERNAL_FAMILY_ID AS CHAR)
		) fm_agg ON fm_agg.family_join_id = CAST(f.EXTERNAL_FAMILY_ID AS CHAR)
		LEFT JOIN (%s) aadhaar_agg ON aadhaar_agg.family_join_id = CAST(f.EXTERNAL_FAMILY_ID AS CHAR)
		LEFT JOIN district_master dm ON dm.pklDistrictId = f.DISTRICT_ID
		LEFT JOIN taluka_master tm ON tm.pklTalukaId = f.TALUKA_ID
		LEFT JOIN village_master vm ON vm.pklVillageId = f.VILLAGE_ID
		%s
		ORDER BY f.FAMILY_ID
		LIMIT %d OFFSET %d
	`,
		latCol,
		lngCol,
		sanitationExpr,
		lightingExpr,
		rationExpr,
		bplExpr,
		incomeExpr,
		memberOccExpr,
		memberWorkingExpr,
		memberIlliterateExpr,
		memberDivyangExpr,
		memberUnemployedExpr,
		aggFamilyWhere,
		aadhaarCoverageSQL,
		where,
		limit,
		offset,
	)
	queryArgs := append([]interface{}{}, args...)
	queryArgs = append(queryArgs, args...)

	if os.Getenv("HOUSES_DIAGNOSTIC") == "1" {
		log.Println("LAT COL:", latCol, "LNG COL:", lngCol)
		log.Println("Params:",
			"district_id=", districtID,
			"taluka_id=", talukaID,
			"village_id=", villageID,
		)
		log.Println("WHERE:", where)
		log.Println("ARGS (filters only):", args)
		log.Println("QUERY ARGS:", queryArgs)
		log.Println("MAIN SQL:", query)

		// 6. COUNT before coordinate filter — DISTRICT_ID only (same bind pattern as request)
		if strings.TrimSpace(districtID) != "" {
			districtOnlyWhere := "WHERE 1=1"
			districtOnlyArgs := []interface{}{}
			appendFamilyGeoIDFilter(&districtOnlyWhere, &districtOnlyArgs, "f", "DISTRICT_ID", districtID)
			var countBeforeCoord int
			if err := h.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM FAMILY f %s", districtOnlyWhere), districtOnlyArgs...).Scan(&countBeforeCoord); err != nil {
				log.Println("COUNT before coordinate filter error:", err)
			} else {
				log.Println("COUNT before coordinate filter (DISTRICT_ID only):", countBeforeCoord)
			}
		} else {
			log.Println("COUNT before coordinate filter (DISTRICT_ID only): skipped (empty district_id)")
		}

		// 7. COUNT after coordinate filter — identical WHERE + args as main list query
		var countAfterCoord int
		if err := h.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM FAMILY f %s", where), args...).Scan(&countAfterCoord); err != nil {
			log.Println("COUNT after coordinate filter error:", err)
		} else {
			log.Println("COUNT after coordinate filter (full WHERE):", countAfterCoord)
		}

		// 8. Sample 5 rows without coordinate filter (resolved lat/lng column names)
		if strings.TrimSpace(districtID) != "" {
			sampleNoGeoWhere := "WHERE 1=1"
			sampleNoGeoArgs := []interface{}{}
			appendFamilyGeoIDFilter(&sampleNoGeoWhere, &sampleNoGeoArgs, "f", "DISTRICT_ID", districtID)
			sampleNoGeoSQL := fmt.Sprintf(
				"SELECT f.FAMILY_ID, f.%s, f.%s FROM FAMILY f %s LIMIT 5",
				latCol, lngCol, sampleNoGeoWhere,
			)
			log.Println("sample (no coord filter) SQL:", sampleNoGeoSQL)
			sr8, err := h.DB.Query(sampleNoGeoSQL, sampleNoGeoArgs...)
			if err != nil {
				log.Println("sample (no coord filter) error:", err)
			} else {
				i := 0
				for sr8.Next() {
					var fid, la, ln sql.NullString
					if scanErr := sr8.Scan(&fid, &la, &ln); scanErr != nil {
						log.Println("sample (no coord filter) scan error:", scanErr)
						break
					}
					log.Println("sample (no coord filter) row:", i+1, "FAMILY_ID=", fid.String, latCol+"=", la.String, lngCol+"=", ln.String)
					i++
				}
				sr8.Close()
				log.Println("sample (no coord filter) row count:", i)
			}
		}

		// 9. Sample 5 rows WITH same WHERE as main query (coordinate filter applied)
		sampleFullSQL := fmt.Sprintf(
			"SELECT f.FAMILY_ID, f.%s, f.%s FROM FAMILY f %s LIMIT 5",
			latCol, lngCol, where,
		)
		log.Println("sample (with coord filter) SQL:", sampleFullSQL)
		sr9, err9 := h.DB.Query(sampleFullSQL, args...)
		if err9 != nil {
			log.Println("sample (with coord filter) error:", err9)
		} else {
			j := 0
			for sr9.Next() {
				var fid, la, ln sql.NullString
				if scanErr := sr9.Scan(&fid, &la, &ln); scanErr != nil {
					log.Println("sample (with coord filter) scan error:", scanErr)
					break
				}
				log.Println("sample (with coord filter) row:", j+1, "FAMILY_ID=", fid.String, latCol+"=", la.String, lngCol+"=", ln.String)
				j++
			}
			sr9.Close()
			log.Println("sample (with coord filter) row count:", j)
		}
	}

	if os.Getenv("DIGITAL_TWIN_DEBUG_SQL") == "1" {
		log.Println("HOUSES QUERY:", query)
	}

	if os.Getenv("HOUSES_EXPLAIN") == "1" {
		explainRows, exErr := h.DB.Query("EXPLAIN ANALYZE "+query, queryArgs...)
		if exErr != nil {
			log.Println("HOUSES EXPLAIN ANALYZE error:", exErr)
		} else {
			cols, _ := explainRows.Columns()
			for explainRows.Next() {
				raw := make([]sql.RawBytes, len(cols))
				scan := make([]interface{}, len(cols))
				for i := range raw {
					scan[i] = &raw[i]
				}
				if err := explainRows.Scan(scan...); err != nil {
					log.Println("HOUSES EXPLAIN scan error:", err)
					break
				}
				var b strings.Builder
				for i, c := range cols {
					if i > 0 {
						b.WriteString(" | ")
					}
					b.WriteString(c)
					b.WriteString("=")
					if i < len(raw) && raw[i] != nil {
						b.Write(raw[i])
					}
				}
				log.Println("HOUSES EXPLAIN ANALYZE:", b.String())
			}
			explainRows.Close()
		}
	}

	rows, err := h.DB.Query(query, queryArgs...)
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
			&house.TypeHouse,
			&house.OwnershipHouse, &house.PradhanMantriAwas, &house.SanitationToiletFacilityHome, &house.ASoakpitManagingWastewater, &house.RationCardColor,
			&house.AadhaarCoverageStatus, &house.MembersWithAadhaar, &house.TotalFamilyMembers,
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

	if os.Getenv("HOUSES_DIAGNOSTIC") == "1" {
		log.Println("Rows returned:", len(houses))
	}

	// Avoid COUNT(*) over filtered FAMILY (expensive at scale). Clients use pagination by page/limit; total reflects this page only.
	total := len(houses)

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

	appendFamilyGeoIDFilter(&where, &args, "f", "DISTRICT_ID", strings.TrimSpace(c.Query("district_id")))
	appendFamilyGeoIDFilter(&where, &args, "f", "TALUKA_ID", strings.TrimSpace(c.Query("taluka_id")))
	appendFamilyGeoIDFilter(&where, &args, "f", "VILLAGE_ID", strings.TrimSpace(c.Query("village_id")))

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
	idStr := strings.TrimSpace(c.Param("id"))

	numericID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
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
	aadhaarCoverageSQL := h.buildAadhaarCoverageSQL()
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
			COALESCE(f.OWNERSHIP_HOUSE, ''),
			COALESCE(f.PRADHAN_MANTRI_AWAS, ''),
			COALESCE(f.SANITATION_TOILET_FACILITY_HOME, ''),
			COALESCE(f.A_SOAKPIT_MANAGING_WASTEWATER, ''),
			COALESCE(f.RATION_CARD_COLOR, ''),
			COALESCE(aadhaar_agg.aadhaar_coverage_status, 'unknown'),
			COALESCE(aadhaar_agg.members_with_aadhaar, 0),
			COALESCE(aadhaar_agg.total_family_members, 0),
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
		LEFT JOIN (%s) aadhaar_agg ON aadhaar_agg.family_join_id = CAST(f.EXTERNAL_FAMILY_ID AS CHAR)
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
		aadhaarCoverageSQL,
	)

	var house HouseRecord
	err = h.DB.QueryRow(query, numericID).Scan(
		&house.FamilyID,
		&house.ExternalFamilyID,
		&house.HouseNo,
		&house.DistrictID, &house.DistrictName,
		&house.TalukaID, &house.TalukaName,
		&house.VillageID, &house.VillageName,
		&house.Latitude, &house.Longitude,
		&house.TotalLand, &house.CultivatedLand, &house.OwnLand,
		&house.WaterSource, &house.Kharif, &house.Rabi,
		&house.OwnershipHouse, &house.PradhanMantriAwas, &house.SanitationToiletFacilityHome, &house.ASoakpitManagingWastewater, &house.RationCardColor,
		&house.AadhaarCoverageStatus, &house.MembersWithAadhaar, &house.TotalFamilyMembers,
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

	// ✅ CACHE FIRST (must be at top)
	value, ok := houseCache.Load(numericID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "house not found"})
		return
	}

	houseDetail := value.(*HouseDetail)

	// ✅ MEMBER FILTER LOGIC
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

	// ✅ use idStr (string) for member matching
	memberExternalID := idStr
	memberFamilyID := idStr

	if hasExternalFamilyID && h.CC != nil && h.CC.Has("EXTERNAL_FAMILY_ID") {
		_ = h.DB.QueryRow(
			"SELECT CAST(COALESCE(EXTERNAL_FAMILY_ID, FAMILY_ID) AS CHAR) FROM FAMILY WHERE FAMILY_ID = ?",
			numericID,
		).Scan(&memberExternalID)
	}

	if hasExternalFamilyID {
		memberArgs = append(memberArgs, memberExternalID, memberFamilyID)
	}
	if hasFamilyID {
		memberArgs = append(memberArgs, memberFamilyID)
	}

	// ✅ HANDLE NO MEMBER CASE
	if len(memberWhere) == 0 {
		c.JSON(http.StatusOK, HouseDetail{
			HouseRecord: houseDetail.HouseRecord,
			Members:     []MemberRecord{},
		})
		return
	}

	c.JSON(http.StatusOK, value)
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
