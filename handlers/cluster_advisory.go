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
	ProblemKey   string `json:"problemKey"`
	ProblemLabel string `json:"problemLabel"`
	AffectedPct  int    `json:"affectedPct"`
	Count        int    `json:"count"`
	Total        int    `json:"total"`
	IsMassIssue  bool   `json:"isMassIssue"` // true when >= MASS_THRESHOLD of cluster
	MassHeading  string `json:"massHeading"` // "Mass Issue Detected: Lack of Irrigation"
	Cause        string `json:"cause"`
	GroupAction  string `json:"groupAction"` // community-level recommended action
	SchemeName   string `json:"schemeName"`
	SchemeBenefit string `json:"schemeBenefit"`
	SchemeType   string `json:"schemeType"` // "community_scheme" | "government_scheme"
	Source       string `json:"source"`     // "scheme_criteria" | "curated"
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
	DB              *sql.DB
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
	totalStr    := strings.TrimSpace(c.Query("total"))

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
		actions = append(actions, action)
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
	pct     := int(float64(inp.Count) / float64(max1(inp.Total, 1)) * 100)
	isMass  := pct >= MASS_THRESHOLD
	meta    := clusterProblemMeta(inp.Key)

	massHeading := ""
	if isMass {
		massHeading = fmt.Sprintf("Mass Issue Detected: %s", meta.massLabel)
	}

	// 1. Try scheme_criteria for community-level scheme
	if h.hasSchemeCriteria {
		if item, ok := h.fromSchemeCriteria(inp, pct, isMass, massHeading, meta); ok {
			return item
		}
	}

	// 2. Curated community fallback
	return h.curatedGroupAction(inp, pct, isMass, massHeading, meta)
}

func (h *ClusterAdvisoryHandler) fromSchemeCriteria(
	inp ClusterProblemInput, pct int, isMass bool, massHeading string, meta clusterMeta,
) (GroupActionItem, bool) {
	query := `
		SELECT COALESCE(scheme_name,''), COALESCE(benefit,''), COALESCE(description,'')
		FROM scheme_criteria
		WHERE FIND_IN_SET(?, REPLACE(problem_key,' ','')) > 0
		   OR problem_key = ?
		LIMIT 1`
	var name, benefit, desc string
	err := h.DB.QueryRow(query, inp.Key, inp.Key).Scan(&name, &benefit, &desc)
	if err != nil || name == "" {
		return GroupActionItem{}, false
	}
	groupAction := desc
	if benefit != "" {
		groupAction = benefit
	}
	return GroupActionItem{
		ProblemKey:    inp.Key,
		ProblemLabel:  meta.label,
		AffectedPct:   pct,
		Count:         inp.Count,
		Total:         inp.Total,
		IsMassIssue:   isMass,
		MassHeading:   massHeading,
		Cause:         fmt.Sprintf("This area lacks %s, affecting %d of %d households (%d%%).", meta.resourceLabel, inp.Count, inp.Total, pct),
		GroupAction:   groupAction,
		SchemeName:    name,
		SchemeBenefit: benefit,
		SchemeType:    "government_scheme",
		Source:        "scheme_criteria",
	}, true
}

func (h *ClusterAdvisoryHandler) curatedGroupAction(
	inp ClusterProblemInput, pct int, isMass bool, massHeading string, meta clusterMeta,
) GroupActionItem {
	type communityEntry struct {
		cause       string
		groupAction string
		scheme      string
		benefit     string
		schemeType  string
	}

	communityData := map[string]communityEntry{
		"noIrrigation": {
			cause:       fmt.Sprintf("This area lacks a primary water source, affecting %d of %d households (%d%%).", inp.Count, inp.Total, pct),
			groupAction: "Apply for a Cluster-based Irrigation Project — the District Agriculture Office can process a group application for a Community Farm Pond or Group Well covering multiple fields in one approval.",
			scheme:      "PMKSY – Community Farm Pond Scheme / Group Drip Irrigation Project",
			benefit:     "Government funds up to 90% of shared irrigation infrastructure for clusters of 5+ households",
			schemeType:  "community_scheme",
		},
		"noSanitation": {
			cause:       fmt.Sprintf("Open defecation remains prevalent — %d of %d households (%d%%) in this area have no toilet facility.", inp.Count, inp.Total, pct),
			groupAction: "Organise a Gram Sabha sanitation drive — the Gram Panchayat can apply for a cluster toilet-construction grant covering all ODF-gap households in one ward simultaneously.",
			scheme:      "Swachh Bharat Mission (Gramin) – ODF+ Ward Sanitation Drive",
			benefit:     "₹12,000 per toilet; cluster applications processed as a single ward sanitation plan",
			schemeType:  "community_scheme",
		},
		"noLand": {
			cause:       fmt.Sprintf("%d of %d households (%d%%) in this cluster are landless — collective livelihood action is needed.", inp.Count, inp.Total, pct),
			groupAction: "Form a Farmer Producer Organisation (FPO) or Self Help Group (SHG) to access collective lease farming, shared inputs, and institutional credit as a group.",
			scheme:      "NRLM / FPO Formation under SFAC – Collective Land Access Programme",
			benefit:     "Revolving fund, crop loan linkage, and marketing support for organised landless groups",
			schemeType:  "community_scheme",
		},
		"noRationCard": {
			cause:       fmt.Sprintf("%d of %d households (%d%%) lack a valid ration card — food security is at risk in this area.", inp.Count, inp.Total, pct),
			groupAction: "Organise a Ration Card Enrollment Camp at the Gram Panchayat office — bulk Aadhaar-linked applications can be submitted together, reducing processing time.",
			scheme:      "National Food Security Act (NFSA) – Cluster Enrollment Drive",
			benefit:     "5 kg free grain/member/month; group applications prioritised in Gram Panchayat allocation",
			schemeType:  "government_scheme",
		},
		"unemployed": {
			cause:       fmt.Sprintf("%d of %d households (%d%%) have unemployed members — a targeted livelihood camp is recommended.", inp.Count, inp.Total, pct),
			groupAction: "Coordinate with the District Employment Mission to run a skill assessment and placement camp for this cluster — group registrations on e-Shram enable faster social security linkage.",
			scheme:      "DDU-GKY + MGNREGA + e-Shram Cluster Registration Camp",
			benefit:     "MGNREGA job cards + 100 days guaranteed work; DDU-GKY skill training and placement support",
			schemeType:  "community_scheme",
		},
		"bplFamilies": {
			cause:       fmt.Sprintf("%d of %d households (%d%%) are BPL-classified — a multi-scheme convergence camp is needed.", inp.Count, inp.Total, pct),
			groupAction: "Run a BPL Welfare Convergence Camp covering Ayushman Bharat registration, NFSA ration card enrollment, and PM-KISAN form collection for all eligible households in one session.",
			scheme:      "Ayushman Bharat PM-JAY + NFSA + PM-KISAN – BPL Convergence Camp",
			benefit:     "₹5 lakh health cover + subsidised grain + ₹6,000/year farmer income support per household",
			schemeType:  "community_scheme",
		},
		"illiterateMembers": {
			cause:       fmt.Sprintf("%d of %d households (%d%%) have illiterate members — this cluster needs a literacy mission outreach.", inp.Count, inp.Total, pct),
			groupAction: "Establish a Saakshar Bharat literacy circle at the Anganwadi centre serving this cluster — identify a local facilitator and register non-literate adults as a group.",
			scheme:      "Saakshar Bharat Mission – Cluster Literacy Circle",
			benefit:     "Free adult literacy classes; government funds facilitator and materials for every registered circle",
			schemeType:  "community_scheme",
		},
		"unemployedMembers": {
			cause:       fmt.Sprintf("%d of %d households (%d%%) have unemployed members — group social security registration is the priority.", inp.Count, inp.Total, pct),
			groupAction: "Conduct a cluster-level e-Shram registration drive — bulk registration enables faster portal onboarding and unlocks accidental insurance immediately for all registered workers.",
			scheme:      "e-Shram Cluster Registration + SHG Formation under DAY-NRLM",
			benefit:     "₹2 lakh accidental insurance per registered worker; SHG credit linkage within 3 months of formation",
			schemeType:  "community_scheme",
		},
		"divyangMembers": {
			cause:       fmt.Sprintf("%d of %d households (%d%%) include divyang (disabled) members — certification and pension access must be coordinated.", inp.Count, inp.Total, pct),
			groupAction: "Organise a Disability Certification Camp with a visiting CDMO/DHO team — group certification drives reduce travel barrier for affected families and speed up pension enrollment.",
			scheme:      "Indira Gandhi National Disability Pension + ADIP Scheme – Cluster Certification Camp",
			benefit:     "Monthly pension ₹300–500 per person + free assistive devices for BPL/divyang members",
			schemeType:  "community_scheme",
		},
		"laborers": {
			cause:       fmt.Sprintf("%d of %d households (%d%%) are primarily wage-laborer households — collective social security registration is the first step.", inp.Count, inp.Total, pct),
			groupAction: "Run a MGNREGA Job Card + e-Shram cluster enrollment camp — all wage-laborer households can apply together, reducing admin time and ensuring no household is left out.",
			scheme:      "MGNREGA + e-Shram Cluster Worker Registration",
			benefit:     "Job cards for 100 days guaranteed work + ₹2 lakh accidental insurance per worker",
			schemeType:  "community_scheme",
		},
	}

	if e, ok := communityData[inp.Key]; ok {
		return GroupActionItem{
			ProblemKey:    inp.Key,
			ProblemLabel:  meta.label,
			AffectedPct:   pct,
			Count:         inp.Count,
			Total:         inp.Total,
			IsMassIssue:   isMass,
			MassHeading:   massHeading,
			Cause:         e.cause,
			GroupAction:   e.groupAction,
			SchemeName:    e.scheme,
			SchemeBenefit: e.benefit,
			SchemeType:    e.schemeType,
			Source:        "curated",
		}
	}

	return GroupActionItem{
		ProblemKey:   inp.Key,
		ProblemLabel: meta.label,
		AffectedPct:  pct,
		Count:        inp.Count,
		Total:        inp.Total,
		IsMassIssue:  isMass,
		MassHeading:  massHeading,
		Cause:        fmt.Sprintf("Data analysis suggests %d households (%d%%) are affected in this area.", inp.Count, pct),
		GroupAction:  "Consult the District Collector or Block Development Officer for a coordinated intervention.",
		SchemeName:   "District Administration",
		SchemeType:   "government_scheme",
		Source:       "curated",
	}
}

type clusterMeta struct {
	label         string
	massLabel     string
	resourceLabel string
}

func clusterProblemMeta(key string) clusterMeta {
	table := map[string]clusterMeta{
		"noIrrigation":      {"No Irrigation",      "Lack of Irrigation",      "a primary water source"},
		"noSanitation":      {"No Sanitation",       "Open Sanitation Crisis",  "toilet facilities"},
		"noLand":            {"No Own Land",          "Landlessness",            "owned agricultural land"},
		"noRationCard":      {"No Ration Card",       "Food Security Gap",       "valid ration cards"},
		"unemployed":        {"Unemployed",           "Unemployment Crisis",     "stable employment"},
		"laborers":          {"Laborers",             "Wage Labour Dependency",  "formal employment"},
		"bplFamilies":       {"BPL Families",         "Poverty Concentration",   "welfare scheme access"},
		"illiterateMembers": {"Illiterate Members",   "Literacy Crisis",         "formal education"},
		"unemployedMembers": {"Unemployed Members",   "Member Unemployment",     "employment opportunities"},
		"divyangMembers":    {"Divyang Members",      "Disability Support Gap",  "disability services"},
	}
	if m, ok := table[key]; ok {
		return m
	}
	return clusterMeta{key, key, key}
}

func max1(a, b int) int {
	if a > b {
		return a
	}
	return b
}
