package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ViewOption represents a single selectable item in the VIEW BY dropdown.
type ViewOption struct {
	Value  string `json:"value"`
	Label  string `json:"label"`
	Action string `json:"action"` // "colorMode" or "view"
}

// ViewOptionGroup is a named group of related options (e.g. "Population").
type ViewOptionGroup struct {
	Label   string       `json:"label"`
	Options []ViewOption `json:"options"`
}

// ViewOptionsHandler pre-builds the VIEW BY option list at startup based on
// what columns actually exist in the database. All DB introspection runs once
// in the constructor; GetViewOptions serves the cached result per-request.
type ViewOptionsHandler struct {
	DB     *sql.DB
	CC     *ColumnChecker
	groups []ViewOptionGroup
}

// NewViewOptionsHandler runs schema introspection once at startup and caches
// the resulting option list. Must be called after NewColumnChecker().
func NewViewOptionsHandler(db *sql.DB, cc *ColumnChecker) *ViewOptionsHandler {
	h := &ViewOptionsHandler{DB: db, CC: cc}
	h.groups = h.buildGroups()
	fmt.Printf("[ViewOptions] built %d groups\n", len(h.groups))
	return h
}

func (h *ViewOptionsHandler) buildGroups() []ViewOptionGroup {
	// colsExist is defined in unified_registry.go (same package).
	fmCols := colsExist(h.DB, "FAMILY_MEMBER", []string{
		"EDUCATION", "DIVYANG", "OCCUPATION", "AADHAAR", "CASTE_CERTIFICATE",
	})

	// Agriculture data lives in FAMILY columns, not separate tables.
	hasCrops := h.CC.Has("CULTIVATING_DURING_KHARIF_SEASON") || h.CC.Has("TAKING_CROPS_RABI_SEASON") || h.CC.Has("CULTIVATING_DURING_RABI_SEASON")
	hasIrrigation := h.CC.Has("SOURCE_WATER_IRRIGATION") || h.CC.Has("A_PUMP_IRRIGATION")
	hasLand := h.CC.Has("AREA_AGRICULTURE_LAND_ACRES") || h.CC.Has("LAND_UNDER_CULTIVATION_ACRES")

	var groups []ViewOptionGroup

	// Population group — population_density is always available
	popOpts := []ViewOption{
		{Value: "population_density", Label: "Population Density", Action: "colorMode"},
	}
	if fmCols["EDUCATION"] {
		popOpts = append(popOpts, ViewOption{"education_level", "Education Level", "colorMode"})
	}
	if fmCols["DIVYANG"] {
		popOpts = append(popOpts, ViewOption{"divyang_presence", "Divyang Presence", "colorMode"})
	}
	if fmCols["OCCUPATION"] {
		popOpts = append(popOpts, ViewOption{"occupation", "Occupation", "colorMode"})
	}
	groups = append(groups, ViewOptionGroup{"Population", popOpts})

	// Infrastructure group
	var infraOpts []ViewOption
	if h.CC.Has("SANITATION_TOILET_FACILITY") || h.CC.Has("TYPE_OF_LATRINE") {
		infraOpts = append(infraOpts, ViewOption{"sanitation", "Sanitation / Toilet", "colorMode"})
	}
	if h.CC.Has("ELECTRICITY_CONNECTION") || h.CC.Has("SOURCE_OF_LIGHTING") {
		infraOpts = append(infraOpts, ViewOption{"lighting", "Electricity", "colorMode"})
	}
	if h.CC.Has("RATION_CARD_TYPE") || h.CC.Has("TYPE_OF_RATION_CARD") {
		infraOpts = append(infraOpts, ViewOption{"ration", "Ration Card", "colorMode"})
	}
	if len(infraOpts) > 0 {
		groups = append(groups, ViewOptionGroup{"Infrastructure", infraOpts})
	}

	// Document Gap Analysis group
	var docOpts []ViewOption
	if h.CC.Has("FAMILY_BELONG_BPL_CATEGORY") {
		docOpts = append(docOpts, ViewOption{"bpl_ration_status", "BPL / Ration Card Status", "colorMode"})
	}
	if fmCols["AADHAAR"] {
		docOpts = append(docOpts, ViewOption{"aadhaar_coverage", "Aadhaar Coverage", "colorMode"})
	}
	if fmCols["CASTE_CERTIFICATE"] {
		docOpts = append(docOpts, ViewOption{"caste_certificate_coverage", "Caste Certificate Coverage", "colorMode"})
	}
	if fmCols["OCCUPATION"] {
		docOpts = append(docOpts, ViewOption{"unemployed_gap", "Unemployed Adults", "colorMode"})
	}
	if fmCols["DIVYANG"] {
		docOpts = append(docOpts, ViewOption{"divyang_gap", "Divyang — Certificate Gap", "colorMode"})
	}
	if len(docOpts) > 0 {
		groups = append(groups, ViewOptionGroup{"Document Gap Analysis", docOpts})
	}

	// Agriculture group
	var agriOpts []ViewOption
	if hasCrops {
		agriOpts = append(agriOpts, ViewOption{"crop", "Crop Type", "view"})
	}
	if hasIrrigation {
		agriOpts = append(agriOpts, ViewOption{"irrigation", "Irrigation", "view"})
	}
	if hasLand {
		agriOpts = append(agriOpts, ViewOption{"land", "Land Holdings", "view"})
	}
	if len(agriOpts) > 0 {
		groups = append(groups, ViewOptionGroup{"Agriculture", agriOpts})
	}

	return groups
}

// GetViewOptions — GET /view-options
// Returns the pre-computed, schema-driven VIEW BY option groups.
func (h *ViewOptionsHandler) GetViewOptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"groups": h.groups})
}
