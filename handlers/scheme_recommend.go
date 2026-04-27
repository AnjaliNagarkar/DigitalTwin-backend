package handlers

import (
	"database/sql"
	"fmt"
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
	Source     string                 `json:"source"` // "db" | "fallback"
	Schemes    []SchemeRecommendation `json:"schemes"`
}

// SchemeRecommendHandler holds DB ref and the pre-computed table-existence flag.
type SchemeRecommendHandler struct {
	DB            *sql.DB
	hasSchemeTable bool // set once at construction time
}

func NewSchemeRecommendHandler(db *sql.DB) *SchemeRecommendHandler {
	h := &SchemeRecommendHandler{DB: db}
	// Check once at startup whether scheme_criteria table exists.
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='scheme_criteria'`).Scan(&n)
	h.hasSchemeTable = n > 0
	if h.hasSchemeTable {
		fmt.Println("[SchemeRecommend] scheme_criteria table found — using DB schemes")
	} else {
		fmt.Println("[SchemeRecommend] scheme_criteria table not found — using built-in fallback data")
	}
	return h
}

// GET /schemes/recommend?problem=noIrrigation&land_size=small&occupation=farmer&bpl=yes
func (h *SchemeRecommendHandler) GetRecommendations(c *gin.Context) {
	problemKey  := strings.TrimSpace(c.Query("problem"))
	landSize    := strings.ToLower(strings.TrimSpace(c.Query("land_size")))
	occupation  := strings.ToLower(strings.TrimSpace(c.Query("occupation")))
	bpl         := strings.ToLower(strings.TrimSpace(c.Query("bpl")))

	if problemKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "problem query param is required"})
		return
	}

	if h.hasSchemeTable {
		resp, err := h.fromDB(problemKey, landSize, occupation, bpl)
		if err == nil {
			c.JSON(http.StatusOK, resp)
			return
		}
		// DB error — fall through to hardcoded fallback
		fmt.Printf("[SchemeRecommend] DB query error: %v — using fallback\n", err)
	}

	c.JSON(http.StatusOK, h.fromFallback(problemKey, landSize, occupation, bpl))
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
		return RecommendResponse{}, err
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
	if len(schemes) == 0 {
		// DB had the table but no rows for this problem — use fallback
		fb := h.fromFallback(problemKey, landSize, occupation, bpl)
		fb.Source = "fallback"
		return fb, nil
	}
	return RecommendResponse{
		ProblemKey: problemKey,
		Cause:      causeForProblem(problemKey),
		Source:     "db",
		Schemes:    schemes,
	}, nil
}

// fromFallback returns curated per-problem scheme data when the DB table is absent.
func (h *SchemeRecommendHandler) fromFallback(problemKey, landSize, occupation, bpl string) RecommendResponse {
	type fallbackEntry struct {
		cause   string
		schemes []SchemeRecommendation
	}

	isSmallFarmer := landSize == "small" || landSize == "marginal" || occupation == "farmer"
	isBPL        := bpl == "yes" || bpl == "true" || bpl == "bpl" || bpl == "antyodaya"

	data := map[string]fallbackEntry{
		"noIrrigation": {
			cause: "Household is entirely rain-fed — crop yield is vulnerable to irregular monsoon and drought conditions.",
			schemes: func() []SchemeRecommendation {
				list := []SchemeRecommendation{
					{Name: "PMKSY – Har Khet Ko Pani", Description: "Pradhan Mantri Krishi Sinchai Yojana — irrigation expansion to every field", Benefit: "Canal/pipeline irrigation access", Eligibility: "All farmers with agricultural land", MatchReason: "No irrigation source detected for this household"},
					{Name: "PMKSY – Per Drop More Crop", Description: "Micro-irrigation subsidy (drip & sprinkler systems)", Benefit: "Up to 90% cost subsidy on drip/sprinkler installation", Eligibility: "All landholding farmers", MatchReason: "Rain-fed farming flag detected"},
				}
				if isSmallFarmer {
					list = append(list, SchemeRecommendation{Name: "Nanaji Deshmukh Krishi Sanjivani Yojana", Description: "Climate-resilient agriculture for small and marginal farmers in drought-prone areas", Benefit: "Free soil & water testing, irrigation infrastructure, input support", Eligibility: "Small and marginal farmers in water-stressed districts", MatchReason: "Small/marginal farmer profile matches scheme eligibility"})
				}
				return list
			}(),
		},
		"noSanitation": {
			cause: "No toilet facility — household relies on open defecation or a shared community latrine.",
			schemes: []SchemeRecommendation{
				{Name: "Swachh Bharat Mission (Gramin) Phase II", Description: "Construction of individual household toilets and ODF+ sustainability", Benefit: "₹12,000 cash incentive per toilet constructed", Eligibility: "BPL/APL households without paved toilet facility", MatchReason: "No sanitation facility recorded for this household"},
			},
		},
		"noLand": {
			cause: "Landless or near-zero owned land — limits agriculture-based livelihood and formal credit access.",
			schemes: []SchemeRecommendation{
				{Name: "PM-FME (Formalisation of Micro Food Enterprises)", Description: "Support for landless families to start or join agri-food micro enterprises", Benefit: "Credit-linked subsidy up to ₹10 lakh", Eligibility: "Landless and marginal households involved in food processing", MatchReason: "No owned agricultural land recorded"},
				{Name: "NRLM – DAY-NRLM SHG", Description: "Self Help Group formation for livelihood and collective lease farming", Benefit: "Revolving fund, credit linkage, livelihood support", Eligibility: "Rural women from landless/BPL households", MatchReason: "Landless household eligible for SHG livelihood support"},
			},
		},
		"noRationCard": {
			cause: "Household may be excluded from subsidised food benefits due to absent or invalid ration card.",
			schemes: []SchemeRecommendation{
				{Name: "National Food Security Act (NFSA) Enrollment", Description: "Enrollment drive for ration card under Pradhan Mantri Garib Kalyan Anna Yojana", Benefit: "5 kg free grain per member per month", Eligibility: "Households not holding a valid ration card", MatchReason: "Ration card missing or unrecognised for this household"},
			},
		},
		"unemployed": {
			cause: "Head of household recorded as unemployed — no stable livelihood income source identified.",
			schemes: []SchemeRecommendation{
				{Name: "DDU-GKY (Deen Dayal Upadhyaya Grameen Kaushalya Yojana)", Description: "Skill training and placement for rural youth", Benefit: "Free vocational training + guaranteed placement assistance", Eligibility: "Rural youth aged 15–35", MatchReason: "Unemployed status detected"},
				{Name: "MGNREGA", Description: "Guaranteed 100 days of unskilled wage employment per year", Benefit: "₹220–300/day wage + job card", Eligibility: "All rural households on demand", MatchReason: "Unemployment flag — MGNREGA provides immediate wage work"},
			},
		},
		"bplFamilies": {
			cause: "Household classified under BPL category — economically vulnerable with limited access to mainstream credit and welfare.",
			schemes: func() []SchemeRecommendation {
				list := []SchemeRecommendation{
					{Name: "Ayushman Bharat PM-JAY", Description: "Health insurance for BPL families", Benefit: "₹5 lakh health cover per family per year", Eligibility: "BPL families as per SECC data", MatchReason: "BPL classification match"},
					{Name: "NFSA – Priority Household Ration Card", Description: "Subsidised food grains under PDS", Benefit: "5 kg grain at ₹1–3 per kg per member", Eligibility: "BPL/Antyodaya households", MatchReason: "BPL classification match"},
				}
				if isBPL {
					list = append(list, SchemeRecommendation{Name: "PM-KISAN", Description: "Direct income support to small and marginal farmers", Benefit: "₹6,000/year in 3 instalments", Eligibility: "BPL and small/marginal landholder farmers", MatchReason: "BPL + agricultural household profile"})
				}
				return list
			}(),
		},
		"illiterateMembers": {
			cause: "One or more household members have no formal schooling recorded — limits long-term income mobility.",
			schemes: []SchemeRecommendation{
				{Name: "Saakshar Bharat Mission", Description: "Adult literacy and functional literacy for non-literate citizens", Benefit: "Free literacy classes through Anganwadi/school centers", Eligibility: "Non-literate adults aged 15 and above", MatchReason: "Illiterate member(s) detected in household"},
				{Name: "Anganwadi / ICDS", Description: "Early childhood education and nutrition for children under 6", Benefit: "Free pre-school education + supplementary nutrition", Eligibility: "Children 0–6 years and pregnant/lactating women", MatchReason: "Household may have young children eligible for ICDS"},
			},
		},
		"unemployedMembers": {
			cause: "One or more household members are recorded as unemployed — risk of income shock without social protection.",
			schemes: []SchemeRecommendation{
				{Name: "e-Shram Portal Registration", Description: "National database of unorganised workers — gateway to social security", Benefit: "₹2 lakh accidental insurance + welfare scheme access", Eligibility: "Unorganised sector workers aged 16–59 not in EPFO/ESIC", MatchReason: "Unemployed member(s) detected — e-Shram registration is the first step"},
				{Name: "PM Vishwakarma Yojana", Description: "Support for traditional artisans and craftspeople", Benefit: "₹15,000 toolkit grant + skill training + ₹3 lakh credit at 5%", Eligibility: "Traditional craftspeople in 18 trades", MatchReason: "May apply if occupation involves traditional craft/trade"},
			},
		},
		"divyangMembers": {
			cause: "One or more household members have a disability recorded — entitled to specific disability welfare and pension schemes.",
			schemes: []SchemeRecommendation{
				{Name: "Indira Gandhi National Disability Pension Scheme", Description: "Monthly pension for persons with severe disability", Benefit: "₹300–500/month central share + state top-up", Eligibility: "BPL households with ≥80% disability aged 18–79", MatchReason: "Divyang member detected in household"},
				{Name: "NHFDC – Disability Finance", Description: "Concessional loans for self-employment for persons with disability", Benefit: "Loans at 5% interest for livelihood activities", Eligibility: "Persons with ≥40% disability", MatchReason: "Divyang member eligible for NHFDC livelihood loan"},
				{Name: "ADIP Scheme (Assistive Devices)", Description: "Free/subsidised assistive devices for persons with disability", Benefit: "Hearing aids, wheelchairs, crutches, etc. at subsidised rates", Eligibility: "BPL persons with disability", MatchReason: "BPL + Divyang profile matches ADIP eligibility"},
			},
		},
		"laborers": {
			cause: "Household members are primarily wage laborers — income is irregular and lacks formal social security coverage.",
			schemes: []SchemeRecommendation{
				{Name: "MGNREGA", Description: "100 days of guaranteed unskilled wage employment", Benefit: "₹220–300/day + job card + worksite facilities", Eligibility: "All rural households on demand", MatchReason: "Wage laborer household — MGNREGA provides income floor"},
				{Name: "e-Shram Portal Registration", Description: "Unorganised worker registration — access to welfare schemes", Benefit: "₹2 lakh accidental insurance + priority in PM welfare schemes", Eligibility: "Unorganised sector workers", MatchReason: "Laborer profile — e-Shram is the entry point for worker benefits"},
			},
		},
	}

	entry, ok := data[problemKey]
	if !ok {
		return RecommendResponse{
			ProblemKey: problemKey,
			Cause:      fmt.Sprintf("Data analysis suggests a gap in %s. Please review with the relevant line department.", humaniseProblemKey(problemKey)),
			Source:     "fallback",
			Schemes:    []SchemeRecommendation{},
		}
	}
	return RecommendResponse{
		ProblemKey: problemKey,
		Cause:      entry.cause,
		Source:     "fallback",
		Schemes:    entry.schemes,
	}
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

func causeForProblem(key string) string {
	causes := map[string]string{
		"noIrrigation":      "Household is entirely rain-fed — crop yield is vulnerable to irregular monsoon and drought conditions.",
		"noSanitation":      "No toilet facility — household relies on open defecation or a shared community latrine.",
		"noLand":            "Landless or near-zero owned land — limits agriculture-based livelihood and formal credit access.",
		"noRationCard":      "Household may be excluded from subsidised food benefits due to absent or invalid ration card.",
		"unemployed":        "Head of household recorded as unemployed — no stable livelihood income source identified.",
		"laborers":          "Household members are primarily wage laborers — income is irregular and lacks formal social security.",
		"bplFamilies":       "Household classified under BPL category — economically vulnerable.",
		"illiterateMembers": "One or more household members have no formal schooling recorded.",
		"unemployedMembers": "One or more household members are recorded as unemployed.",
		"divyangMembers":    "One or more household members have a disability recorded.",
	}
	if c, ok := causes[key]; ok {
		return c
	}
	return fmt.Sprintf("Data analysis suggests a gap in %s.", humaniseProblemKey(key))
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
