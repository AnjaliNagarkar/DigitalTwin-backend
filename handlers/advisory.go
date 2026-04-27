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
	ProblemKey   string `json:"problemKey"`
	ProblemLabel string `json:"problemLabel"`
	Color        string `json:"color"`
	Cause        string `json:"cause"`
	Solution     string `json:"solution"`
	SchemeName   string `json:"schemeName"`
	SchemeType   string `json:"schemeType"`  // "government_scheme" | "technical_advisory" | "fallback"
	Source       string `json:"source"`      // "advisory_master" | "scheme_criteria" | "curated"
	CropContext  string `json:"cropContext"` // which crop this applies to (empty = all)
}

// AdvisoryResponse is the full response for /advisory
type AdvisoryResponse struct {
	FamilyID int            `json:"familyId"`
	Issues   []AdvisoryItem `json:"issues"`
}

type AdvisoryHandler struct {
	DB                *sql.DB
	hasAdvisoryTable  bool
	hasSchemeCriteria bool
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
	crop := strings.ToLower(strings.TrimSpace(c.Query("crop")))
	landSize := strings.ToLower(strings.TrimSpace(c.Query("land_size")))
	bpl := strings.ToLower(strings.TrimSpace(c.Query("bpl")))
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
		// Only include items with actual cause AND solution — filter out empty/incomplete advisories
		if strings.TrimSpace(item.Cause) != "" && strings.TrimSpace(item.Solution) != "" {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []AdvisoryItem{}
	}

	c.JSON(http.StatusOK, AdvisoryResponse{FamilyID: familyID, Issues: items})
}

// resolveAdvisory only uses scheme_criteria from database.
// Returns empty advisory item if not found (no curated fallback).
func (h *AdvisoryHandler) resolveAdvisory(problemKey, crop, landSize, bpl string) AdvisoryItem {
	// Only try scheme_criteria (government scheme match) — database only.
	if h.hasSchemeCriteria {
		if item, ok := h.fromSchemeCriteria(problemKey, crop, landSize, bpl); ok {
			return item
		}
	}
	// No data found in database — return empty advisory (nothing shown to user)
	return AdvisoryItem{}
}

// fromAdvisoryMaster queries advisory_master table.
// DEPRECATED: No longer used. Only scheme_criteria is used for database-driven advisories.

// fromSchemeCriteria queries scheme_criteria for government scheme suggestions.
// Expected schema: scheme_name, cause, solution (description or benefit), eligibility_criteria
// Returns false if not found — no fallback used.
func (h *AdvisoryHandler) fromSchemeCriteria(problemKey, _, landSize, _ string) (AdvisoryItem, bool) {
	query := `
		SELECT COALESCE(scheme_name,''), COALESCE(cause,''),
		       COALESCE(description,''), COALESCE(benefit,''),
		       COALESCE(eligibility_criteria,'')
		FROM scheme_criteria
		WHERE FIND_IN_SET(?, REPLACE(problem_key,' ','')) > 0
		   OR problem_key = ?
		ORDER BY (LOWER(eligibility_criteria) LIKE CONCAT('%',?,'%')) DESC
		LIMIT 1`

	var name, cause, desc, benefit, eligibility string
	err := h.DB.QueryRow(query, problemKey, problemKey, landSize).Scan(
		&name, &cause, &desc, &benefit, &eligibility)
	if err != nil || name == "" {
		// Not found in database — don't show anything
		return AdvisoryItem{}, false
	}

	meta := problemMeta(problemKey)
	solution := desc
	if benefit != "" && desc == "" {
		solution = benefit
	}

	// Only return if BOTH cause and solution exist in database
	if strings.TrimSpace(cause) == "" || strings.TrimSpace(solution) == "" {
		return AdvisoryItem{}, false
	}

	// Only populate fields that come from database
	return AdvisoryItem{
		ProblemKey:   problemKey,
		ProblemLabel: meta.label,
		Color:        meta.color,
		Cause:        cause,
		Solution:     solution,
		SchemeName:   name,
		SchemeType:   "government_scheme",
		Source:       "scheme_criteria",
		CropContext:  "",
	}, true
}

// problemMeta returns display metadata for a problem key (minimal — label + color only).
type problemMetaT struct {
	label string
	color string
}

func problemMeta(key string) problemMetaT {
	table := map[string]problemMetaT{
		"noLand":          {"No Own Land", "#ef4444"},
		"marginalHolding": {"Marginal Holding", "#f59e0b"},
		"uncultivated":    {"Uncultivated Land", "#ef4444"},
		"noIrrigation":    {"No Irrigation", "#a78bfa"},
		"noCropRecord":    {"No Crop Record", "#60a5fa"},
		"singleSeason":    {"Single Season Only", "#38bdf8"},
	}
	if m, ok := table[key]; ok {
		return m
	}
	return problemMetaT{key, "#94a3b8"}
}
