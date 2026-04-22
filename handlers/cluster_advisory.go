package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ClusterProblemInput is one problem entry sent by the frontend.
// { key, count, total } — derived from the cluster's house list.
type ClusterProblemInput struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
	Total int    `json:"total"`
}

// GroupActionItem is a single community-level action recommendation.
type GroupActionItem struct {
	ProblemKey    string `json:"problemKey"`
	ProblemLabel  string `json:"problemLabel"`
	AffectedPct   int    `json:"affectedPct"`
	Count         int    `json:"count"`
	Total         int    `json:"total"`
	IsMassIssue   bool   `json:"isMassIssue"` // true when >= MASS_THRESHOLD of cluster
	MassHeading   string `json:"massHeading"` // "Mass Issue Detected: Lack of Irrigation"
	Cause         string `json:"cause"`
	GroupAction   string `json:"groupAction"` // community-level recommended action
	SchemeName    string `json:"schemeName"`
	SchemeBenefit string `json:"schemeBenefit"`
	SchemeType    string `json:"schemeType"` // "community_scheme" | "government_scheme"
	Source        string `json:"source"`     // "scheme_criteria" | "curated"
}

// ClusterAdvisoryResponse is returned for /advisory/cluster
type ClusterAdvisoryResponse struct {
	TotalHouseholds int               `json:"totalHouseholds"`
	PriorityLabel   string            `json:"priorityLabel"` // "High Priority Cluster" | "Medium" etc.
	Actions         []GroupActionItem `json:"actions"`
}

// MASS_THRESHOLD: % of cluster that must share a problem for "Mass Issue" heading
const MASS_THRESHOLD = 60

type ClusterAdvisoryHandler struct {
	DB                *sql.DB
	hasSchemeCriteria bool
}

func NewClusterAdvisoryHandler(db *sql.DB) *ClusterAdvisoryHandler {
	h := &ClusterAdvisoryHandler{DB: db}
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='scheme_criteria'`).Scan(&n)
	h.hasSchemeCriteria = n > 0
	return h
}

// GET /advisory/cluster?problems=noIrrigation:8:10,noSanitation:5:10&total=10
// problems param: comma-separated key:count:total triplets
func (h *ClusterAdvisoryHandler) GetClusterAdvisory(c *gin.Context) {
	problemsRaw := strings.TrimSpace(c.Query("problems"))
	totalStr := strings.TrimSpace(c.Query("total"))

	total, _ := strconv.Atoi(totalStr)
	if total <= 0 {
		total = 1
	}
	if problemsRaw == "" {
		c.JSON(http.StatusOK, ClusterAdvisoryResponse{
			TotalHouseholds: total,
			PriorityLabel:   "No Issues Detected",
			Actions:         []GroupActionItem{},
		})
		return
	}

	var inputs []ClusterProblemInput
	for _, seg := range strings.Split(problemsRaw, ",") {
		parts := strings.Split(strings.TrimSpace(seg), ":")
		if len(parts) < 2 {
			continue
		}
		count, _ := strconv.Atoi(parts[1])
		t := total
		if len(parts) >= 3 {
			t, _ = strconv.Atoi(parts[2])
		}
		inputs = append(inputs, ClusterProblemInput{Key: parts[0], Count: count, Total: t})
	}

	// Sort descending by count so highest-impact problem comes first
	sort.Slice(inputs, func(i, j int) bool { return inputs[i].Count > inputs[j].Count })

	var actions []GroupActionItem
	for _, inp := range inputs {
		action := h.resolveGroupAction(inp)
		// Only include actions with actual cause AND groupAction — filter out empty/incomplete advisories
		if strings.TrimSpace(action.Cause) != "" && strings.TrimSpace(action.GroupAction) != "" {
			actions = append(actions, action)
		}
	}
	if actions == nil {
		actions = []GroupActionItem{}
	}

	// Priority label based on top problem severity
	priority := "Moderate Priority Cluster"
	for _, a := range actions {
		if a.IsMassIssue {
			priority = "High Priority Cluster"
			break
		}
	}

	c.JSON(http.StatusOK, ClusterAdvisoryResponse{
		TotalHouseholds: total,
		PriorityLabel:   priority,
		Actions:         actions,
	})
}

func (h *ClusterAdvisoryHandler) resolveGroupAction(inp ClusterProblemInput) GroupActionItem {
	pct := int(float64(inp.Count) / float64(max1(inp.Total, 1)) * 100)
	isMass := pct >= MASS_THRESHOLD
	meta := clusterProblemMeta(inp.Key)

	massHeading := ""
	if isMass {
		massHeading = fmt.Sprintf("Mass Issue Detected: %s", meta.massLabel)
	}

	// Only try scheme_criteria (database only) — no curated fallback.
	if h.hasSchemeCriteria {
		if item, ok := h.fromSchemeCriteria(inp, pct, isMass, massHeading, meta); ok {
			return item
		}
	}

	// Not found in database — return empty group action (nothing shown to user)
	return GroupActionItem{}
}

func (h *ClusterAdvisoryHandler) fromSchemeCriteria(
	inp ClusterProblemInput, pct int, isMass bool, massHeading string, meta clusterMeta,
) (GroupActionItem, bool) {
	query := `
		SELECT COALESCE(scheme_name,''), COALESCE(cause,''), 
		       COALESCE(benefit,''), COALESCE(description,'')
		FROM scheme_criteria
		WHERE FIND_IN_SET(?, REPLACE(problem_key,' ','')) > 0
		   OR problem_key = ?
		LIMIT 1`
	var name, cause, benefit, desc string
	err := h.DB.QueryRow(query, inp.Key, inp.Key).Scan(&name, &cause, &benefit, &desc)
	if err != nil || name == "" {
		// Not found in database — return false
		return GroupActionItem{}, false
	}

	groupAction := desc
	if benefit != "" && desc == "" {
		groupAction = benefit
	}

	// Only return if BOTH cause and groupAction exist in database
	if strings.TrimSpace(cause) == "" || strings.TrimSpace(groupAction) == "" {
		return GroupActionItem{}, false
	}

	return GroupActionItem{
		ProblemKey:    inp.Key,
		ProblemLabel:  meta.label,
		AffectedPct:   pct,
		Count:         inp.Count,
		Total:         inp.Total,
		IsMassIssue:   isMass,
		MassHeading:   massHeading,
		Cause:         cause,
		GroupAction:   groupAction,
		SchemeName:    name,
		SchemeBenefit: benefit,
		SchemeType:    "government_scheme",
		Source:        "scheme_criteria",
	}, true
}

// DEPRECATED: curatedGroupAction removed. Only database-driven group actions are used.

type clusterMeta struct {
	label     string
	massLabel string
}

func clusterProblemMeta(key string) clusterMeta {
	table := map[string]clusterMeta{
		"noIrrigation":      {"No Irrigation", "Lack of Irrigation"},
		"noSanitation":      {"No Sanitation", "Open Sanitation Crisis"},
		"noLand":            {"No Own Land", "Landlessness"},
		"noRationCard":      {"No Ration Card", "Food Security Gap"},
		"unemployed":        {"Unemployed", "Unemployment Crisis"},
		"laborers":          {"Laborers", "Wage Labour Dependency"},
		"bplFamilies":       {"BPL Families", "Poverty Concentration"},
		"illiterateMembers": {"Illiterate Members", "Literacy Crisis"},
		"unemployedMembers": {"Unemployed Members", "Member Unemployment"},
		"divyangMembers":    {"Divyang Members", "Disability Support Gap"},
	}
	if m, ok := table[key]; ok {
		return m
	}
	return clusterMeta{key, key}
}

func max1(a, b int) int {
	if a > b {
		return a
	}
	return b
}
