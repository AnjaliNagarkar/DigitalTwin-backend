package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdvisoryItem is one cause+solution pair returned for a field issue.
type AdvisoryItem struct {
	ProblemKey  string `json:"problemKey"`
	ProblemLabel string `json:"problemLabel"`
	Color       string `json:"color"`
	Cause       string `json:"cause"`
	Solution    string `json:"solution"`
	SchemeName  string `json:"schemeName"`
	SchemeType  string `json:"schemeType"` // "government_scheme" | "technical_advisory" | "fallback"
	Source      string `json:"source"`     // "advisory_master" | "scheme_criteria" | "curated"
	CropContext string `json:"cropContext"` // which crop this applies to (empty = all)
}

// AdvisoryResponse is the full response for /advisory
type AdvisoryResponse struct {
	FamilyID int            `json:"familyId"`
	Issues   []AdvisoryItem `json:"issues"`
}

type AdvisoryHandler struct {
	DB                  *sql.DB
	hasAdvisoryTable    bool
	hasSchemeCriteria   bool
}

func NewAdvisoryHandler(db *sql.DB) *AdvisoryHandler {
	h := &AdvisoryHandler{DB: db}
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='advisory_master'`).Scan(&n)
	h.hasAdvisoryTable = n > 0

	n = 0
	_ = db.QueryRow(`SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='scheme_criteria'`).Scan(&n)
	h.hasSchemeCriteria = n > 0

	fmt.Printf("[Advisory] advisory_master=%v  scheme_criteria=%v\n", h.hasAdvisoryTable, h.hasSchemeCriteria)
	return h
}

// GET /advisory?problems=noIrrigation,noLand&crop=cotton&land_size=small&bpl=yes&family_id=123
func (h *AdvisoryHandler) GetAdvisory(c *gin.Context) {
	problemsRaw := strings.TrimSpace(c.Query("problems"))
	crop        := strings.ToLower(strings.TrimSpace(c.Query("crop")))
	landSize    := strings.ToLower(strings.TrimSpace(c.Query("land_size")))
	bpl         := strings.ToLower(strings.TrimSpace(c.Query("bpl")))
	familyIDStr := strings.TrimSpace(c.Query("family_id"))

	if problemsRaw == "" {
		c.JSON(http.StatusOK, AdvisoryResponse{Issues: []AdvisoryItem{}})
		return
	}

	problemKeys := strings.Split(problemsRaw, ",")
	var familyID int
	fmt.Sscanf(familyIDStr, "%d", &familyID)

	var items []AdvisoryItem
	for _, key := range problemKeys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		item := h.resolveAdvisory(key, crop, landSize, bpl)
		items = append(items, item)
	}
	if items == nil {
		items = []AdvisoryItem{}
	}

	c.JSON(http.StatusOK, AdvisoryResponse{FamilyID: familyID, Issues: items})
}

// resolveAdvisory tries advisory_master → scheme_criteria → curated fallback (in order).
func (h *AdvisoryHandler) resolveAdvisory(problemKey, crop, landSize, bpl string) AdvisoryItem {
	// 1. Try advisory_master (technical farming advice)
	if h.hasAdvisoryTable {
		if item, ok := h.fromAdvisoryMaster(problemKey, crop, landSize); ok {
			return item
		}
	}
	// 2. Try scheme_criteria (government scheme match)
	if h.hasSchemeCriteria {
		if item, ok := h.fromSchemeCriteria(problemKey, crop, landSize, bpl); ok {
			return item
		}
	}
	// 3. Curated fallback
	return h.curatedAdvisory(problemKey, crop, landSize)
}

// fromAdvisoryMaster queries advisory_master table.
// Expected schema: problem_key, crop_type (nullable), land_size (nullable),
//                  cause, solution, scheme_name, scheme_type
func (h *AdvisoryHandler) fromAdvisoryMaster(problemKey, crop, landSize string) (AdvisoryItem, bool) {
	// Try exact crop+problem match first, then any-crop match
	query := `
		SELECT COALESCE(cause,''), COALESCE(solution,''),
		       COALESCE(scheme_name,''), COALESCE(scheme_type,'technical_advisory'),
		       COALESCE(crop_type,'')
		FROM advisory_master
		WHERE problem_key = ?
		  AND (crop_type IS NULL OR crop_type = '' OR LOWER(crop_type) = ?)
		  AND (land_size  IS NULL OR land_size  = '' OR LOWER(land_size)  = ?)
		ORDER BY (crop_type IS NOT NULL AND crop_type != '') DESC,
		         (land_size IS NOT NULL AND land_size  != '') DESC
		LIMIT 1`

	var cause, solution, schemeName, schemeType, cropCtx string
	err := h.DB.QueryRow(query, problemKey, crop, landSize).Scan(
		&cause, &solution, &schemeName, &schemeType, &cropCtx)
	if err != nil {
		return AdvisoryItem{}, false
	}
	meta := problemMeta(problemKey)
	return AdvisoryItem{
		ProblemKey:   problemKey,
		ProblemLabel: meta.label,
		Color:        meta.color,
		Cause:        nonEmpty(cause, meta.defaultCause),
		Solution:     nonEmpty(solution, "Consult the local Agriculture Department for guidance."),
		SchemeName:   schemeName,
		SchemeType:   schemeType,
		Source:       "advisory_master",
		CropContext:  cropCtx,
	}, true
}

// fromSchemeCriteria queries scheme_criteria for government scheme suggestions.
func (h *AdvisoryHandler) fromSchemeCriteria(problemKey, _, landSize, _ string) (AdvisoryItem, bool) {
	query := `
		SELECT COALESCE(scheme_name,''), COALESCE(description,''),
		       COALESCE(benefit,''), COALESCE(eligibility_criteria,'')
		FROM scheme_criteria
		WHERE FIND_IN_SET(?, REPLACE(problem_key,' ','')) > 0
		   OR problem_key = ?
		ORDER BY (LOWER(eligibility_criteria) LIKE CONCAT('%',?,'%')) DESC
		LIMIT 1`

	var name, desc, benefit, eligibility string
	err := h.DB.QueryRow(query, problemKey, problemKey, landSize).Scan(
		&name, &desc, &benefit, &eligibility)
	if err != nil {
		return AdvisoryItem{}, false
	}
	meta := problemMeta(problemKey)
	solution := benefit
	if desc != "" {
		solution = desc
	}
	return AdvisoryItem{
		ProblemKey:   problemKey,
		ProblemLabel: meta.label,
		Color:        meta.color,
		Cause:        meta.defaultCause,
		Solution:     nonEmpty(solution, "Contact your local Gram Panchayat for scheme enrollment."),
		SchemeName:   name,
		SchemeType:   "government_scheme",
		Source:       "scheme_criteria",
		CropContext:  "",
	}, true
}

// curatedAdvisory returns crop-aware hardcoded advice as last resort.
func (h *AdvisoryHandler) curatedAdvisory(problemKey, crop, landSize string) AdvisoryItem {
	meta := problemMeta(problemKey)

	type entry struct {
		cause      string
		solution   string
		scheme     string
		schemeType string
	}

	// Crop-specific overrides for common problems
	cropSpecific := map[string]map[string]entry{
		"noIrrigation": {
			"cotton":      {cause: "Cotton is highly moisture-sensitive — rain-fed cultivation leads to boll drop and fibre quality loss.", solution: "Install drip irrigation to maintain consistent soil moisture at 60–70% field capacity during boll formation.", scheme: "PMKSY – Per Drop More Crop", schemeType: "government_scheme"},
			"sugarcane":   {cause: "Sugarcane requires 1500–2500 mm of water — rain-fed growing drastically reduces sugar recovery rate.", solution: "Switch to furrow irrigation or drip-fertigation system with PMKSY subsidy support.", scheme: "PMKSY + State Sugarcane Development Corp", schemeType: "government_scheme"},
			"wheat":       {cause: "Wheat is irrigated at 5 critical growth stages — missing any stage cuts yield by 15–30%.", solution: "Apply 6 cm irrigation at tillering, crown root initiation, and grain filling stages.", scheme: "PMKSY + KVK Wheat Advisory", schemeType: "technical_advisory"},
			"rice":        {cause: "Paddy requires standing water in early growth — rain-fed paddies miss critical flooding windows.", solution: "Use Alternate Wetting and Drying (AWD) method to save 25–30% water while maintaining yield.", scheme: "PMKSY + ICAR AWD Protocol", schemeType: "technical_advisory"},
			"soybean":     {cause: "Soybean is sensitive to drought at flowering and pod-fill stages — rain-fed crop loses 30–40% yield.", solution: "Install sprinkler irrigation and apply 25mm irrigation at R1 and R5 growth stages.", scheme: "PMKSY Micro-irrigation Subsidy", schemeType: "government_scheme"},
		},
		"singleSeason": {
			"cotton":      {cause: "Cotton is a Kharif-only crop — off-season land lies fallow, losing 4–5 months of productive income.", solution: "Intercrop with chickpea or mustard in Rabi season to maximise land use efficiency.", scheme: "RKVY Crop Diversification + ATMA Advisory", schemeType: "technical_advisory"},
			"rice":        {cause: "Mono-cropping rice depletes soil nitrogen and increases pest pressure in subsequent seasons.", solution: "Rotate with pulses (lentil/chickpea) in Rabi to restore soil fertility and diversify income.", scheme: "RKVY + KVK Crop Rotation Module", schemeType: "technical_advisory"},
		},
		"marginalHolding": {
			"cotton":      {cause: "Marginal cotton holding (≤1 acre) — high input cost per unit area, low mechanisation benefit.", solution: "Form FPO with neighbouring farmers for bulk input purchase and shared machinery; target Bt-cotton hybrid.", scheme: "PM-KISAN + FPO Formation under SFAC", schemeType: "government_scheme"},
			"default":     {cause: "Marginal holding limits single-crop income — household vulnerable to one bad season.", solution: "Adopt high-value crops (vegetables, spices) and kitchen horticulture on a portion of the land.", scheme: "MIDH + ATMA Farm Advisory", schemeType: "government_scheme"},
		},
	}

	// Check for crop-specific override
	if cropOverrides, ok := cropSpecific[problemKey]; ok {
		if e, ok2 := cropOverrides[crop]; ok2 {
			return AdvisoryItem{
				ProblemKey:   problemKey,
				ProblemLabel: meta.label,
				Color:        meta.color,
				Cause:        e.cause,
				Solution:     e.solution,
				SchemeName:   e.scheme,
				SchemeType:   e.schemeType,
				Source:       "curated",
				CropContext:  crop,
			}
		}
		// Use generic default for this problem if exists
		if e, ok2 := cropOverrides["default"]; ok2 {
			return AdvisoryItem{
				ProblemKey:   problemKey,
				ProblemLabel: meta.label,
				Color:        meta.color,
				Cause:        e.cause,
				Solution:     e.solution,
				SchemeName:   e.scheme,
				SchemeType:   e.schemeType,
				Source:       "curated",
				CropContext:  "",
			}
		}
	}

	// Generic fallbacks by problem key
	generic := map[string]entry{
		"noLand":          {cause: "No cultivable land — household relies entirely on wage labour for livelihood.", solution: "Enroll in NRLM SHG for collective lease farming and join the nearest FPO for shared land access.", scheme: "NRLM / DAY-NRLM SHG Formation", schemeType: "government_scheme"},
		"uncultivated":    {cause: "Owned land is entirely fallow — possible reasons include labour shortage, credit gap, or water unavailability.", solution: "Apply for a Kisan Credit Card and obtain seed kit + extension support from KVK to restart cultivation.", scheme: "Kisan Credit Card (KCC) + KVK Seed Kit", schemeType: "government_scheme"},
		"noIrrigation":    {cause: "Household is entirely rain-fed — crop yield is highly vulnerable to irregular monsoon and drought.", solution: "Register for drip/sprinkler irrigation at District Agriculture Office — government covers 55–90% of cost.", scheme: "PMKSY – Per Drop More Crop", schemeType: "government_scheme"},
		"noCropRecord":    {cause: "No crop record in either season — land may be fallow or data not captured during survey.", solution: "Prepare a two-season crop calendar with local KVK and Agriculture Extension Worker support.", scheme: "ATMA + mKisan Crop Advisories", schemeType: "technical_advisory"},
		"singleSeason":    {cause: "Active in only one crop season — doubles income risk and leaves land unproductive half the year.", solution: "Introduce a second-season short-duration crop (mung, chickpea, mustard) with KVK guidance and RKVY seed support.", scheme: "RKVY Seasonal Diversification", schemeType: "government_scheme"},
		"marginalHolding": {cause: "Marginal holding (≤1 acre) — high input cost per unit output, limited mechanisation scope.", solution: "Adopt high-value short-duration vegetables or kitchen horticulture; link with MIDH for infrastructure support.", scheme: "MIDH + ATMA Farm Advisory", schemeType: "government_scheme"},
	}
	if e, ok := generic[problemKey]; ok {
		return AdvisoryItem{
			ProblemKey:   problemKey,
			ProblemLabel: meta.label,
			Color:        meta.color,
			Cause:        e.cause,
			Solution:     e.solution,
			SchemeName:   e.scheme,
			SchemeType:   e.schemeType,
			Source:       "curated",
			CropContext:  "",
		}
	}

	// Ultimate fallback
	return AdvisoryItem{
		ProblemKey:   problemKey,
		ProblemLabel: meta.label,
		Color:        meta.color,
		Cause:        fmt.Sprintf("Data analysis suggests a gap in %s for this household.", meta.label),
		Solution:     "Consult the local Agriculture Department or Krishi Vigyan Kendra (KVK) for tailored guidance.",
		SchemeName:   "District Agriculture Office",
		SchemeType:   "technical_advisory",
		Source:       "curated",
		CropContext:  "",
	}
}

// problemMeta returns display metadata for a problem key.
type problemMetaT struct {
	label        string
	color        string
	defaultCause string
}

func problemMeta(key string) problemMetaT {
	table := map[string]problemMetaT{
		"noLand":          {"No Own Land", "#ef4444", "Household has no cultivable land — farm income is unstable."},
		"marginalHolding": {"Marginal Holding", "#f59e0b", "Holding ≤1 acre — limits single-crop income potential."},
		"uncultivated":    {"Uncultivated Land", "#ef4444", "Owned land is fallow — cultivated area is zero."},
		"noIrrigation":    {"No Irrigation", "#a78bfa", "Fully rain-fed farming — high crop-loss risk during irregular monsoon."},
		"noCropRecord":    {"No Crop Record", "#60a5fa", "No kharif or rabi crop recorded in survey."},
		"singleSeason":    {"Single Season Only", "#38bdf8", "Active in only one season — reduces annual income stability."},
	}
	if m, ok := table[key]; ok {
		return m
	}
	return problemMetaT{key, "#94a3b8", "Issue identified for this household."}
}

func nonEmpty(s, fallback string) string {
	if strings.TrimSpace(s) != "" {
		return s
	}
	return fallback
}
