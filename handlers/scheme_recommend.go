package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SchemeRecommendation is what the API returns for each matching scheme.
type SchemeRecommendation struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Benefit     string `json:"benefit"`
	Eligibility string `json:"eligibility"`
	MatchReason string `json:"matchReason"`
}

// RecommendResponse wraps schemes + a cause string for the problem.
type RecommendResponse struct {
	ProblemKey string                 `json:"problemKey"`
	Cause      string                 `json:"cause"`
	Source     string                 `json:"source,omitempty"`
	Schemes    []SchemeRecommendation `json:"schemes"`
}

// SchemeRecommendHandler holds DB ref and the pre-computed table-existence flag.
type SchemeRecommendHandler struct {
	DB *sql.DB
}

func NewSchemeRecommendHandler(db *sql.DB) *SchemeRecommendHandler {
	return &SchemeRecommendHandler{DB: db}
}

// GET /schemes/recommend?problem=noIrrigation&land_size=small&occupation=farmer&bpl=yes
func (h *SchemeRecommendHandler) GetRecommendations(c *gin.Context) {
	problemKey := strings.TrimSpace(c.Query("problem"))
	landSize := strings.ToLower(strings.TrimSpace(c.Query("land_size")))
	occupation := strings.ToLower(strings.TrimSpace(c.Query("occupation")))
	bpl := strings.ToLower(strings.TrimSpace(c.Query("bpl")))

	if problemKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "problem query param is required"})
		return
	}

	resp, _ := h.fromDB(problemKey, landSize, occupation, bpl)
	if resp.Schemes == nil {
		resp.Schemes = []SchemeRecommendation{}
	}
	c.JSON(http.StatusOK, resp)
}

// fromDB queries scheme_criteria table.
// Expected columns: id, scheme_name, problem_key, eligibility_criteria, benefit, description, applicable_occupation
// All are VARCHAR; problem_key is comma-separated when a scheme covers multiple problems.
func (h *SchemeRecommendHandler) fromDB(problemKey, landSize, occupation, bpl string) (RecommendResponse, error) {
	query := `
		SELECT scheme_name,
		       COALESCE(description, ''),
		       COALESCE(benefit, ''),
		       COALESCE(eligibility_criteria, ''),
		       COALESCE(applicable_occupation, '')
		FROM scheme_criteria
		WHERE FIND_IN_SET(?, REPLACE(problem_key, ' ', '')) > 0
		   OR problem_key = ?
		ORDER BY scheme_name
		LIMIT 10`

	rows, err := h.DB.Query(query, problemKey, problemKey)
	if err != nil {
		return RecommendResponse{
			ProblemKey: problemKey,
			Cause:      "",
			Schemes:    []SchemeRecommendation{},
		}, err
	}
	defer rows.Close()

	var schemes []SchemeRecommendation
	for rows.Next() {
		var name, desc, benefit, eligibility, appOcc string
		if err := rows.Scan(&name, &desc, &benefit, &eligibility, &appOcc); err != nil {
			continue
		}
		reason := buildMatchReason(problemKey, landSize, occupation, bpl, eligibility, appOcc)
		schemes = append(schemes, SchemeRecommendation{
			Name:        name,
			Description: desc,
			Benefit:     benefit,
			Eligibility: eligibility,
			MatchReason: reason,
		})
	}
	if err := rows.Err(); err != nil {
		return RecommendResponse{
			ProblemKey: problemKey,
			Cause:      "",
			Schemes:    []SchemeRecommendation{},
		}, err
	}
	if len(schemes) == 0 {
		return RecommendResponse{
			ProblemKey: problemKey,
			Cause:      "",
			Schemes:    []SchemeRecommendation{},
		}, nil
	}
	return RecommendResponse{
		ProblemKey: problemKey,
		Cause:      "",
		Source:     "database",
		Schemes:    schemes,
	}, nil
}

// buildMatchReason constructs a plain-language reason string from DB row + citizen profile.
func buildMatchReason(problemKey, landSize, occupation, bpl, eligibility, appOcc string) string {
	parts := []string{}
	if problemKey != "" {
		parts = append(parts, humaniseProblemKey(problemKey)+" detected")
	}
	if appOcc != "" && occupation != "" && strings.Contains(strings.ToLower(appOcc), occupation) {
		parts = append(parts, "occupation ("+occupation+") matches scheme target")
	}
	if bpl == "yes" && strings.Contains(strings.ToLower(eligibility), "bpl") {
		parts = append(parts, "BPL household matches eligibility")
	}
	if landSize != "" && strings.Contains(strings.ToLower(eligibility), landSize) {
		parts = append(parts, landSize+" farmer profile matches eligibility")
	}
	if len(parts) == 0 {
		return "Scheme matches based on identified household problem"
	}
	return strings.Join(parts, " · ")
}

func humaniseProblemKey(key string) string {
	m := map[string]string{
		"noIrrigation":      "Irrigation Access",
		"noSanitation":      "Sanitation",
		"noLand":            "Land Ownership",
		"noRationCard":      "Ration Card Coverage",
		"unemployed":        "Employment",
		"laborers":          "Stable Employment",
		"bplFamilies":       "BPL Welfare Access",
		"illiterateMembers": "Literacy",
		"unemployedMembers": "Member Employment",
		"divyangMembers":    "Disability Support",
	}
	if h, ok := m[key]; ok {
		return h
	}
	return key
}
