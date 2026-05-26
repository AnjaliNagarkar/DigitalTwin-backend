package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type dashboardCacheItem struct {
	Data      gin.H
	ExpiresAt time.Time
}

// sectionCacheItem holds the last successful result for one dashboard section.
// It has no TTL — it is used only as a stale-value fallback when the live query
// times out or errors, preventing zero-filled responses.
type sectionCacheItem struct {
	Data gin.H
}

const dashboardSummaryCacheTTL = 5 * time.Minute

// Per-section query timeout budgets.
// Education and Employment run a single heavy aggregation query each; other
// sections are lighter or parallelise several smaller queries internally.
const (
	timeoutPopulation   = 5 * time.Second
	timeoutDemographics = 8 * time.Second
	timeoutAgriculture  = 8 * time.Second
	timeoutEducation    = 12 * time.Second
	timeoutEmployment   = 12 * time.Second
)

var dashboardSummaryIndexInit sync.Once

type DashboardSummaryHandler struct {
	DB *sql.DB
	CC *ColumnChecker

	// Full-response cache (TTL-based; keyed by location filter params).
	cache    map[string]dashboardCacheItem
	cacheMux sync.RWMutex

	// Per-section stale cache: key = "<cacheKey>:<sectionName>".
	// Updated on every successful section fetch; never expires.
	// Used to return last-good data instead of zeros when a live query times out.
	sectionCache    map[string]sectionCacheItem
	sectionCacheMux sync.RWMutex

	// Stampede prevention: tracks in-flight full-response fetches per cache key.
	// Concurrent requests for the same key wait for the first fetch to complete
	// instead of all hammering the DB simultaneously.
	inflightMu sync.Mutex
	inflight   map[string]chan struct{}
}

func NewDashboardSummaryHandler(db *sql.DB, cc *ColumnChecker) *DashboardSummaryHandler {
	h := &DashboardSummaryHandler{
		DB:           db,
		CC:           cc,
		cache:        map[string]dashboardCacheItem{},
		sectionCache: map[string]sectionCacheItem{},
		inflight:     map[string]chan struct{}{},
	}
	dashboardSummaryIndexInit.Do(func() {
		h.ensureSummaryIndexes()
	})
	return h
}

func (h *DashboardSummaryHandler) ensureSummaryIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const targetIndex = "idx_family_member_location"

	var existing int
	if err := h.DB.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM INFORMATION_SCHEMA.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'FAMILY_MEMBER'
		  AND INDEX_NAME = ?
	`, targetIndex).Scan(&existing); err != nil {
		log.Printf("[dashboard/summary] index check failed: %v", err)
		return
	}

	if existing > 0 {
		return
	}

	var columnCount int
	if err := h.DB.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'FAMILY_MEMBER'
		  AND COLUMN_NAME IN ('DISTRICT_ID', 'TALUKA_ID', 'VILLAGE_ID')
	`).Scan(&columnCount); err != nil {
		log.Printf("[dashboard/summary] FAMILY_MEMBER column check failed: %v", err)
		return
	}

	if columnCount != 3 {
		log.Printf("[dashboard/summary] skipping index creation: FAMILY_MEMBER location columns not found")
		return
	}

	if _, err := h.DB.ExecContext(ctx, `CREATE INDEX idx_family_member_location ON FAMILY_MEMBER (DISTRICT_ID, TALUKA_ID, VILLAGE_ID)`); err != nil {
		log.Printf("[dashboard/summary] index creation failed: %v", err)
		return
	}

	log.Printf("[dashboard/summary] created index idx_family_member_location")
}

func (h *DashboardSummaryHandler) GetDashboardSummary(c *gin.Context) {
	started := time.Now()

	// Support both singular and plural parameter names
	districtID  := normalizeFilterValue(c.Query("district_id"))
	districtIDs := c.Query("district_ids")
	talukaID    := normalizeFilterValue(c.Query("taluka_id"))
	talukaIDs   := c.Query("taluka_ids")
	villageID   := normalizeFilterValue(c.Query("village_id"))
	villageIDs  := c.Query("village_ids")

	log.Printf("[dashboard] request district=%q districts=%q taluka=%q talukas=%q village=%q villages=%q",
		districtID, districtIDs, talukaID, talukaIDs, villageID, villageIDs)

	districtCacheKey := nonEmpty(districtID, districtIDs)
	if districtCacheKey == "" {
		districtCacheKey = "all"
	}
	talukaCacheKey := nonEmpty(talukaID, talukaIDs)
	if talukaCacheKey == "" {
		talukaCacheKey = "all"
	}
	villageCacheKey := nonEmpty(villageID, villageIDs)
	if villageCacheKey == "" {
		villageCacheKey = "all"
	}
	cacheKey := fmt.Sprintf("dashboard_%s_%s_%s", districtCacheKey, talukaCacheKey, villageCacheKey)

	// ── 1. Full-response cache fast path ─────────────────────────────────────
	h.cacheMux.RLock()
	if item, ok := h.cache[cacheKey]; ok && time.Now().Before(item.ExpiresAt) {
		h.cacheMux.RUnlock()
		log.Printf("[dashboard] cache-HIT key=%s elapsed=%s", cacheKey, time.Since(started))
		c.JSON(http.StatusOK, item.Data)
		return
	}
	h.cacheMux.RUnlock()
	log.Printf("[dashboard] cache-MISS key=%s", cacheKey)

	// ── 2. Stampede prevention ───────────────────────────────────────────────
	// Only one goroutine fetches from DB per cache key; all others wait and
	// reuse the result, avoiding concurrent identical queries under burst load.
	h.inflightMu.Lock()
	if ch, exists := h.inflight[cacheKey]; exists {
		h.inflightMu.Unlock()
		log.Printf("[dashboard] in-flight wait key=%s", cacheKey)
		select {
		case <-ch:
		case <-time.After(32 * time.Second):
			log.Printf("[dashboard] in-flight wait timeout key=%s", cacheKey)
		}
		// Serve from cache if the lead goroutine populated it
		h.cacheMux.RLock()
		if item, ok := h.cache[cacheKey]; ok && time.Now().Before(item.ExpiresAt) {
			h.cacheMux.RUnlock()
			log.Printf("[dashboard] served after wait key=%s elapsed=%s", cacheKey, time.Since(started))
			c.JSON(http.StatusOK, item.Data)
			return
		}
		h.cacheMux.RUnlock()
		// Lead fetch did not produce a usable cache entry; serve stale/defaults
		c.JSON(http.StatusOK, h.buildStaleOrDefaultResponse(cacheKey))
		return
	}
	// Register this goroutine as the sole in-flight fetcher for this key
	done := make(chan struct{})
	h.inflight[cacheKey] = done
	h.inflightMu.Unlock()
	defer func() {
		h.inflightMu.Lock()
		delete(h.inflight, cacheKey)
		h.inflightMu.Unlock()
		close(done)
	}()

	// ── 3. DB health check ───────────────────────────────────────────────────
	_, pingCancel, ok := ensureDBReady(c, h.DB, "/dashboard/summary")
	if !ok {
		result := defaultDashboardSummaryResponse()
		result["partial_errors"] = gin.H{"database": "database connection unavailable"}
		c.JSON(http.StatusOK, result)
		return
	}
	defer pingCancel()

	whereF, args := buildOptionalLocationFilterWithArrays("f", districtID, districtIDs, talukaID, talukaIDs, villageID, villageIDs)

	// ── 4. Fetch all sections concurrently with independent per-section timeouts
	// Each section owns its context so a slow section cannot cancel others.
	type sectionResult struct {
		name    string
		data    gin.H
		err     error
		elapsed time.Duration
	}
	type sectionTask struct {
		name    string
		timeout time.Duration
		fetch   func(context.Context) (gin.H, error)
	}

	tasks := []sectionTask{
		{"population", timeoutPopulation, func(ctx context.Context) (gin.H, error) {
			return h.fetchPopulationSection(ctx, whereF, args)
		}},
		{"demographics", timeoutDemographics, func(ctx context.Context) (gin.H, error) {
			return h.fetchDemographicsSection(ctx, whereF, args)
		}},
		{"education", timeoutEducation, func(ctx context.Context) (gin.H, error) {
			return h.fetchEducationSection(ctx, whereF, args)
		}},
		{"employment", timeoutEmployment, func(ctx context.Context) (gin.H, error) {
			return h.fetchEmploymentSection(ctx, whereF, args)
		}},
		{"agriculture", timeoutAgriculture, func(ctx context.Context) (gin.H, error) {
			return h.fetchAgricultureSection(ctx, whereF, args)
		}},
	}

	ch := make(chan sectionResult, len(tasks))
	for _, t := range tasks {
		t := t
		go func() {
			t0 := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
			defer cancel()
			data, err := t.fetch(ctx)
			ch <- sectionResult{name: t.name, data: data, err: err, elapsed: time.Since(t0)}
		}()
	}

	// ── 5. Collect results; apply stale fallback on timeout/error ────────────
	final := defaultDashboardSummaryResponse()
	errorsBySection := map[string]string{}

	for range tasks {
		sr := <-ch
		if sr.err != nil {
			isTimeout := strings.Contains(sr.err.Error(), "context deadline exceeded") ||
				strings.Contains(sr.err.Error(), "context canceled")
			tag := ""
			if isTimeout {
				tag = " [timeout]"
			}
			log.Printf("[dashboard] section=%s error%s elapsed=%s: %v", sr.name, tag, sr.elapsed, sr.err)

			// Prefer last-good stale value over default zeros
			sectionKey := cacheKey + ":" + sr.name
			h.sectionCacheMux.RLock()
			stale, hasStale := h.sectionCache[sectionKey]
			h.sectionCacheMux.RUnlock()
			if hasStale {
				log.Printf("[dashboard] section=%s stale-fallback used", sr.name)
				final[sr.name] = stale.Data
				// Stale data is real — do not add to partial_errors
			} else {
				final[sr.name] = sr.data // section default (zeros)
				errorsBySection[sr.name] = sr.err.Error()
			}
		} else {
			log.Printf("[dashboard] section=%s ok elapsed=%s", sr.name, sr.elapsed)
			final[sr.name] = sr.data
			// Persist successful result for future stale-fallback use
			sectionKey := cacheKey + ":" + sr.name
			h.sectionCacheMux.Lock()
			h.sectionCache[sectionKey] = sectionCacheItem{Data: sr.data}
			h.sectionCacheMux.Unlock()
		}
	}

	if len(errorsBySection) > 0 {
		final["partial_errors"] = errorsBySection
	}

	// Write to full-response cache even if some sections used stale fallbacks.
	// The TTL ensures a fresh DB round-trip will be attempted after expiry,
	// and all requests within the TTL window benefit from the fast path.
	h.cacheMux.Lock()
	h.cache[cacheKey] = dashboardCacheItem{Data: final, ExpiresAt: time.Now().Add(dashboardSummaryCacheTTL)}
	h.cacheMux.Unlock()

	log.Printf("[dashboard] complete key=%s errors=%d elapsed=%s", cacheKey, len(errorsBySection), time.Since(started))
	c.JSON(http.StatusOK, final)
}

// buildStaleOrDefaultResponse returns the best available data per section from
// the per-section stale cache, falling back to zeros for sections not yet seen.
// Called when a concurrent in-flight fetch did not populate the full-response cache.
func (h *DashboardSummaryHandler) buildStaleOrDefaultResponse(cacheKey string) gin.H {
	result := defaultDashboardSummaryResponse()
	sections := []string{"population", "demographics", "education", "employment", "agriculture"}
	h.sectionCacheMux.RLock()
	defer h.sectionCacheMux.RUnlock()
	for _, name := range sections {
		if stale, ok := h.sectionCache[cacheKey+":"+name]; ok {
			result[name] = stale.Data
		}
	}
	return result
}

func buildOptionalLocationFilter(alias, districtID, talukaID, villageID string) (string, []interface{}) {
	where := "1=1"
	args := []interface{}{}

	if district := normalizeFilterValue(districtID); district != "" {
		where += fmt.Sprintf(" AND %s.DISTRICT_ID = ?", alias)
		args = append(args, district)
	}

	if taluka := normalizeFilterValue(talukaID); taluka != "" {
		where += fmt.Sprintf(" AND %s.TALUKA_ID = ?", alias)
		args = append(args, taluka)
	}

	if village := normalizeFilterValue(villageID); village != "" {
		where += fmt.Sprintf(" AND %s.VILLAGE_ID = ?", alias)
		args = append(args, village)
	}

	return where, args
}

// buildOptionalLocationFilterWithArrays builds WHERE clause supporting both single values and comma-separated arrays
func buildOptionalLocationFilterWithArrays(alias, districtID, districtIDs, talukaID, talukaIDs, villageID, villageIDs string) (string, []interface{}) {
	where := "1=1"
	args := []interface{}{}

	// Handle district IDs - support both single and multiple
	var districtIDList []string
	if districtIDs != "" {
		for _, id := range strings.Split(districtIDs, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				districtIDList = append(districtIDList, trimmed)
			}
		}
	} else if district := normalizeFilterValue(districtID); district != "" {
		districtIDList = []string{district}
	}

	if len(districtIDList) > 0 {
		placeholders := make([]string, len(districtIDList))
		for i, id := range districtIDList {
			placeholders[i] = "?"
			args = append(args, id)
		}
		where += fmt.Sprintf(" AND %s.DISTRICT_ID IN (%s)", alias, strings.Join(placeholders, ","))
	}

	// Handle taluka IDs - support both single and multiple
	var talukaIDList []string
	if talukaIDs != "" {
		for _, id := range strings.Split(talukaIDs, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				talukaIDList = append(talukaIDList, trimmed)
			}
		}
	} else if taluka := normalizeFilterValue(talukaID); taluka != "" {
		talukaIDList = []string{taluka}
	}

	if len(talukaIDList) > 0 {
		placeholders := make([]string, len(talukaIDList))
		for i, id := range talukaIDList {
			placeholders[i] = "?"
			args = append(args, id)
		}
		where += fmt.Sprintf(" AND %s.TALUKA_ID IN (%s)", alias, strings.Join(placeholders, ","))
	}

	// Handle village IDs - support both single and multiple
	var villageIDList []string
	if villageIDs != "" {
		for _, id := range strings.Split(villageIDs, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				villageIDList = append(villageIDList, trimmed)
			}
		}
	} else if village := normalizeFilterValue(villageID); village != "" {
		villageIDList = []string{village}
	}

	if len(villageIDList) > 0 {
		placeholders := make([]string, len(villageIDList))
		for i, id := range villageIDList {
			placeholders[i] = "?"
			args = append(args, id)
		}
		where += fmt.Sprintf(" AND %s.VILLAGE_ID IN (%s)", alias, strings.Join(placeholders, ","))
	}

	return where, args
}

func isValidFilterValue(val string) bool {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return false
	}
	if strings.EqualFold(trimmed, "0") || strings.EqualFold(trimmed, "null") || strings.EqualFold(trimmed, "undefined") {
		return false
	}
	return true
}

func normalizeFilterValue(val string) string {
	trimmed := strings.TrimSpace(val)
	if !isValidFilterValue(trimmed) {
		return ""
	}
	return trimmed
}

func nonEmpty(value, fallback string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return fallback
	}
	return v
}

func normalizeLandDistributionLabel(label string) string {
	hyphenNormalizer := strings.NewReplacer(
		"‐", "-", // hyphen
		"‑", "-", // non-breaking hyphen
		"‒", "-", // figure dash
		"–", "-", // en dash
		"—", "-", // em dash
		"−", "-", // minus sign
	)
	normalized := hyphenNormalizer.Replace(strings.TrimSpace(label))
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}

func canonicalLandDistributionLabel(label string) string {
	normalized := normalizeLandDistributionLabel(label)
	switch strings.ToLower(normalized) {
	case "landless":
		return "Landless"
	case "small",
		"marginal (0-1 acre)", "marginal (0-1 acres)",
		"small (1-2.5 acre)", "small (1-2.5 acres)":
		return "Small"
	case "medium",
		"semi medium (2.5-5 acres)", "semi-medium (2.5-5 acres)", "semi medium (2.5-5 acre)", "semi-medium (2.5-5 acre)",
		"medium (5-10 acre)", "medium (5-10 acres)":
		return "Medium"
	case "large", "large (>10 acre)", "large (>10 acres)":
		return "Large"
	default:
		return normalized
	}
}

func withInvalidConnectionRetry(op func() error) error {
	err := op()
	if err == nil {
		return nil
	}
	if !isInvalidDBConnection(err) {
		return err
	}
	return op()
}

func queryRowsWithRetry(ctx context.Context, db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	if strings.Contains(query, "__WHERE_CLAUSE__") {
		return nil, fmt.Errorf("unresolved SQL placeholder __WHERE_CLAUSE__")
	}

	var rows *sql.Rows
	err := withInvalidConnectionRetry(func() error {
		var opErr error
		rows, opErr = db.QueryContext(ctx, query, args...)
		return opErr
	})
	return rows, err
}

func queryRowWithRetry(ctx context.Context, db *sql.DB, query string, args []interface{}, dest ...interface{}) error {
	if strings.Contains(query, "__WHERE_CLAUSE__") {
		return fmt.Errorf("unresolved SQL placeholder __WHERE_CLAUSE__")
	}

	return withInvalidConnectionRetry(func() error {
		return db.QueryRowContext(ctx, query, args...).Scan(dest...)
	})
}

func injectWhere(template, whereClause string) string {
	return strings.ReplaceAll(template, "__WHERE_CLAUSE__", whereClause)
}

func defaultDashboardSummaryResponse() gin.H {
	return gin.H{
		"population":   defaultPopulationSummary(),
		"demographics": defaultDemographicsSummary(),
		"education":    defaultEducationSummary(),
		"employment":   defaultEmploymentSummary(),
		"agriculture":  defaultAgricultureSummary(),
	}
}

func defaultPopulationSummary() gin.H {
	return gin.H{
		"total_population":     0,
		"total_households":     0,
		"working_population":   0,
		"dependent_population": 0,
		"bpl_distribution":     gin.H{"bpl": 0, "non_bpl": 0, "total_households": 0},
	}
}

func defaultDemographicsSummary() gin.H {
	return gin.H{
		"gender_distribution":            gin.H{"male": 0, "female": 0, "other": 0},
		"age_distribution":               gin.H{"age_0_5": 0, "age_6_18": 0, "age_19_35": 0, "age_36_60": 0, "age_60_plus": 0},
		"age_income_gender_distribution": defaultAgeIncomeGenderDistribution(),
		"total_divyang":                  0,
		"disability_distribution":        []gin.H{},
	}
}

func defaultAgeIncomeGenderDistribution() []gin.H {
	ageGroups := []string{"18-30", "31-45", "46-60", "60+"}
	rows := make([]gin.H, 0, len(ageGroups))
	for _, ageGroup := range ageGroups {
		rows = append(rows, gin.H{"age_group": ageGroup, "families": 0, "avg_income": 0.0})
	}
	return rows
}

func defaultEducationSummary() gin.H {
	return gin.H{
		"literate_population":   0,
		"illiterate_population": 0,
		"students_count":        0,
		"dropout_count":         0,
		"graduate_population":   0,
		"literacy_rate":         0,
		"qualification_distribution": gin.H{
			"below_10th": 0, "tenth": 0, "twelfth": 0, "graduate_above": 0,
		},
	}
}

func defaultEmploymentSummary() gin.H {
	return gin.H{
		"employed_members":   0,
		"unemployed_members": 0,
		"daily_wage_workers": 0,
		"skilled_workers":    0,
		"occupation_distribution": gin.H{
			"farm_based": 0, "agri_allied": 0, "non_farm": 0, "salaried": 0,
			"wage_workers": 0, "housewife": 0, "students": 0, "unemployed": 0, "other": 0,
		},
	}
}

func defaultAgricultureSummary() gin.H {
	return gin.H{
		"totalFarmers":             0,
		"farmersWithoutIrrigation": 0,
		"landDistribution":         []gin.H{},
		"landUtilizationRows":      []gin.H{},
		"landUtilizationSummary":   gin.H{"total_land": 0, "cultivated_land": 0, "unused_land": 0, "valid_records": 0, "invalid_records": 0, "cultivated_percent": 0, "unused_percent": 0},
		"seasonCropRows":           []gin.H{},
		"kharifFarmers":            0,
		"rabiFarmers":              0,
	}
}

func (h *DashboardSummaryHandler) fetchPopulationSection(ctx context.Context, whereF string, args []interface{}) (gin.H, error) {
	section := defaultPopulationSummary()

	populationQuery := injectWhere(`
		SELECT COUNT(DISTINCT fm.FAMILY_MEMBER_ID)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f
		ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE __WHERE_CLAUSE__
	`, whereF)
	householdsQuery := injectWhere(`
		SELECT COUNT(*)
		FROM FAMILY f
		WHERE __WHERE_CLAUSE__
	`, whereF)
	workingQuery := injectWhere(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE fm.OCCUPATION IS NOT NULL
		  AND TRIM(fm.OCCUPATION) != ''
		  AND UPPER(TRIM(fm.OCCUPATION)) NOT IN ('HOUSEWIFE','STUDYING','STUDENT','UNEMPLOYED','NOT APPLICABLE','NA')
		  AND __WHERE_CLAUSE__
	`, whereF)
	dependentQuery := injectWhere(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE (
			(STR_TO_DATE(fm.DOB, '%%Y-%%m-%%d') IS NOT NULL AND (TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%Y-%%m-%%d'), CURDATE()) < 18 OR TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%Y-%%m-%%d'), CURDATE()) > 60))
			OR
			(STR_TO_DATE(fm.DOB, '%%Y-%%m-%%d') IS NOT NULL AND TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%Y-%%m-%%d'), CURDATE()) BETWEEN 18 AND 60 AND UPPER(TRIM(COALESCE(fm.OCCUPATION, ''))) IN ('HOUSEWIFE','STUDYING','STUDENT','UNEMPLOYED','NOT APPLICABLE'))
		)
		AND __WHERE_CLAUSE__
	`, whereF)
	bplConditions := []string{}
	if h.CC != nil && h.CC.Has("FAMILY_BELONG_BPL_CATEGORY") {
		bplConditions = append(bplConditions, "UPPER(TRIM(COALESCE(f.FAMILY_BELONG_BPL_CATEGORY, ''))) = 'YES'")
	}
	if h.CC != nil && h.CC.Has("RATION_CARD_TYPE") {
		bplConditions = append(bplConditions, "UPPER(TRIM(COALESCE(f.RATION_CARD_TYPE, ''))) IN ('BPL', 'AAY')")
	}

	var totalPop, totalHouseholds, working, dependent, bplHouseholds int
	var firstErr error
	errMux := sync.Mutex{}
	recordErr := func(err error) {
		if err == nil {
			return
		}
		errMux.Lock()
		if firstErr == nil {
			firstErr = err
		}
		errMux.Unlock()
	}

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		if err := queryRowWithRetry(ctx, h.DB, populationQuery, args, &totalPop); err != nil {
			recordErr(fmt.Errorf("population total query failed: %w", err))
		}
	}()

	go func() {
		defer wg.Done()
		if err := queryRowWithRetry(ctx, h.DB, householdsQuery, args, &totalHouseholds); err != nil {
			recordErr(fmt.Errorf("households query failed: %w", err))
		}
	}()

	go func() {
		defer wg.Done()
		if err := queryRowWithRetry(ctx, h.DB, workingQuery, args, &working); err != nil {
			recordErr(fmt.Errorf("working query failed: %w", err))
		}
	}()

	go func() {
		defer wg.Done()
		if err := queryRowWithRetry(ctx, h.DB, dependentQuery, args, &dependent); err != nil {
			recordErr(fmt.Errorf("dependent query failed: %w", err))
		}
	}()

	wg.Wait()

	if len(bplConditions) > 0 {
		bplHouseholdQuery := injectWhere(fmt.Sprintf(`
			SELECT COUNT(DISTINCT f.FAMILY_ID)
			FROM FAMILY f
			WHERE __WHERE_CLAUSE__
			  AND (%s)
		`, strings.Join(bplConditions, " OR ")), whereF)
		if err := queryRowWithRetry(ctx, h.DB, bplHouseholdQuery, args, &bplHouseholds); err != nil {
			recordErr(fmt.Errorf("bpl query failed: %w", err))
		}
	}

	section["total_population"] = totalPop
	section["total_households"] = totalHouseholds
	section["working_population"] = working
	section["dependent_population"] = dependent
	nonBplHouseholds := totalHouseholds - bplHouseholds
	if nonBplHouseholds < 0 {
		nonBplHouseholds = 0
	}
	section["bpl_distribution"] = gin.H{"bpl": bplHouseholds, "non_bpl": nonBplHouseholds, "total_households": totalHouseholds}
	return section, firstErr
}

func (h *DashboardSummaryHandler) fetchDemographicsSection(ctx context.Context, whereF string, args []interface{}) (gin.H, error) {
	section := defaultDemographicsSummary()

	genderQuery := injectWhere(`
		SELECT
			COUNT(DISTINCT CASE
				WHEN LOWER(TRIM(fm.GENDER)) = 'male'
				THEN fm.FAMILY_MEMBER_ID
			END) AS male,
			COUNT(DISTINCT CASE
				WHEN LOWER(TRIM(fm.GENDER)) = 'female'
				THEN fm.FAMILY_MEMBER_ID
			END) AS female,
			COUNT(DISTINCT CASE
				WHEN LOWER(TRIM(COALESCE(fm.GENDER, ''))) NOT IN ('male', 'female')
				THEN fm.FAMILY_MEMBER_ID
			END) AS other
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f
		ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE __WHERE_CLAUSE__
	`, whereF)
	ageQuery := injectWhere(`
		SELECT
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) BETWEEN 0 AND 5 THEN 1 ELSE 0 END),
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) BETWEEN 6 AND 18 THEN 1 ELSE 0 END),
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) BETWEEN 19 AND 35 THEN 1 ELSE 0 END),
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) BETWEEN 36 AND 60 THEN 1 ELSE 0 END),
			SUM(CASE WHEN TIMESTAMPDIFF(YEAR, STR_TO_DATE(fm.DOB, '%%d-%%m-%%Y'), CURDATE()) > 60 THEN 1 ELSE 0 END)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE __WHERE_CLAUSE__
	`, whereF)
	divyangQuery := injectWhere(`
		SELECT COUNT(*)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES'
		  AND __WHERE_CLAUSE__
	`, whereF)
	ageIncomeQuery := injectWhere(`
		SELECT
			age_group,
			COUNT(DISTINCT EXTERNAL_FAMILY_ID) AS families,
			AVG(income) AS avg_income
		FROM (
			SELECT
				f.EXTERNAL_FAMILY_ID,
				CAST(NULLIF(TRIM(f.ANNUAL_INCOME), '') AS DECIMAL(15,2)) AS income,
				CASE
					WHEN TIMESTAMPDIFF(YEAR, m.selected_dob, CURDATE()) BETWEEN 18 AND 30 THEN '18-30'
					WHEN TIMESTAMPDIFF(YEAR, m.selected_dob, CURDATE()) BETWEEN 31 AND 45 THEN '31-45'
					WHEN TIMESTAMPDIFF(YEAR, m.selected_dob, CURDATE()) BETWEEN 46 AND 60 THEN '46-60'
					ELSE '60+'
				END AS age_group
			FROM FAMILY f
			JOIN (
				SELECT
					fm.EXTERNAL_FAMILY_ID,
					COALESCE(
						MIN(CASE
							WHEN LOWER(TRIM(COALESCE(fm.RELATION_FAMILY_HEAD, ''))) IN ('head', 'self', 'head of family')
							THEN STR_TO_DATE(fm.DOB, '%d-%m-%Y')
						END),
						MIN(STR_TO_DATE(fm.DOB, '%d-%m-%Y'))
					) AS selected_dob
				FROM FAMILY_MEMBER fm
				WHERE fm.DOB IS NOT NULL
				  AND TRIM(fm.DOB) != ''
				  AND STR_TO_DATE(fm.DOB, '%d-%m-%Y') IS NOT NULL
				GROUP BY fm.EXTERNAL_FAMILY_ID
			) m ON m.EXTERNAL_FAMILY_ID = f.EXTERNAL_FAMILY_ID
			WHERE NULLIF(TRIM(f.ANNUAL_INCOME), '') IS NOT NULL
			  AND CAST(NULLIF(TRIM(f.ANNUAL_INCOME), '') AS DECIMAL(15,2)) IS NOT NULL
			  AND TIMESTAMPDIFF(YEAR, m.selected_dob, CURDATE()) >= 18
			  AND __WHERE_CLAUSE__
		) t
		GROUP BY age_group
	`, whereF)
	disabilityQuery := injectWhere(`
		SELECT
			CASE
				WHEN DISABILITY_CATEGORY LIKE '%%Blind%%' OR DISABILITY_CATEGORY LIKE '%%Low vision%%' THEN 'Visual Disability'
				WHEN DISABILITY_CATEGORY LIKE '%%Locomotor%%' OR DISABILITY_CATEGORY LIKE '%%Cerebral%%' OR DISABILITY_CATEGORY LIKE '%%Muscular%%' OR DISABILITY_CATEGORY LIKE '%%Dwarf%%' THEN 'Locomotor Disability'
				WHEN DISABILITY_CATEGORY LIKE '%%Mental%%' OR DISABILITY_CATEGORY LIKE '%%Autism%%' OR DISABILITY_CATEGORY LIKE '%%Intellectual%%' OR DISABILITY_CATEGORY LIKE '%%Learning%%' THEN 'Intellectual Disability'
				WHEN DISABILITY_CATEGORY LIKE '%%Hearing%%' THEN 'Hearing Disability'
				WHEN DISABILITY_CATEGORY LIKE '%%Speech%%' THEN 'Speech Disability'
				WHEN DISABILITY_CATEGORY LIKE '%%Multiple%%' THEN 'Multiple Disabilities'
				WHEN DISABILITY_CATEGORY LIKE '%%Parkinson%%' OR DISABILITY_CATEGORY LIKE '%%Sclerosis%%' OR DISABILITY_CATEGORY LIKE '%%Sickle%%' OR DISABILITY_CATEGORY LIKE '%%Thalassemia%%' THEN 'Chronic Conditions'
				ELSE 'Other'
			END AS disability_group,
			COUNT(DISTINCT fm.FAMILY_MEMBER_ID) AS total
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f
		ON fm.EXTERNAL_FAMILY_ID = f.EXTERNAL_FAMILY_ID
		WHERE UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES'
		  AND __WHERE_CLAUSE__
		GROUP BY disability_group
		ORDER BY total DESC
		LIMIT 8
	`, whereF)

	var firstErr error
	errMux := sync.Mutex{}
	sectionMux := sync.Mutex{}
	recordErr := func(err error) {
		if err == nil {
			return
		}
		errMux.Lock()
		if firstErr == nil {
			firstErr = err
		}
		errMux.Unlock()
	}

	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		defer wg.Done()
		var male, female, other int
		if err := queryRowWithRetry(ctx, h.DB, genderQuery, args, &male, &female, &other); err != nil {
			recordErr(fmt.Errorf("gender query failed: %w", err))
			return
		}
		sectionMux.Lock()
		section["gender_distribution"] = gin.H{"male": male, "female": female, "other": other}
		sectionMux.Unlock()
	}()

	go func() {
		defer wg.Done()
		var age0_5, age6_18, age19_35, age36_60, age60Plus int
		if err := queryRowWithRetry(ctx, h.DB, ageQuery, args, &age0_5, &age6_18, &age19_35, &age36_60, &age60Plus); err != nil {
			recordErr(fmt.Errorf("age query failed: %w", err))
			return
		}
		sectionMux.Lock()
		section["age_distribution"] = gin.H{"age_0_5": age0_5, "age_6_18": age6_18, "age_19_35": age19_35, "age_36_60": age36_60, "age_60_plus": age60Plus}
		sectionMux.Unlock()
	}()

	go func() {
		defer wg.Done()
		var totalDivyang int
		if err := queryRowWithRetry(ctx, h.DB, divyangQuery, args, &totalDivyang); err != nil {
			recordErr(fmt.Errorf("divyang query failed: %w", err))
			return
		}
		sectionMux.Lock()
		section["total_divyang"] = totalDivyang
		sectionMux.Unlock()
	}()

	go func() {
		defer wg.Done()
		rows, err := queryRowsWithRetry(ctx, h.DB, ageIncomeQuery, args...)
		if err != nil {
			log.Println("AGE-GENDER QUERY FAILED:", err)
			recordErr(fmt.Errorf("age-income query failed: %w", err))
			return
		}
		defer rows.Close()

		ageGroupOrder := []string{"18-30", "31-45", "46-60", "60+"}
		pivot := map[string]map[string]interface{}{}
		for _, ageGroup := range ageGroupOrder {
			pivot[ageGroup] = map[string]interface{}{"families": 0, "avg_income": 0.0}
		}

		for rows.Next() {
			var ageGroup string
			var families int
			var avgIncome sql.NullFloat64
			if scanErr := rows.Scan(&ageGroup, &families, &avgIncome); scanErr != nil {
				recordErr(fmt.Errorf("age-income scan failed: %w", scanErr))
				return
			}

			bucket, ok := pivot[ageGroup]
			if !ok {
				continue
			}
			bucket["families"] = families
			if avgIncome.Valid {
				bucket["avg_income"] = avgIncome.Float64
			}
		}

		if err := rows.Err(); err != nil {
			recordErr(fmt.Errorf("age-income rows failed: %w", err))
			return
		}

		ageIncome := make([]gin.H, 0, len(ageGroupOrder))
		for _, ageGroup := range ageGroupOrder {
			bucket := pivot[ageGroup]
			ageIncome = append(ageIncome, gin.H{
				"age_group":  ageGroup,
				"families":   bucket["families"],
				"avg_income": bucket["avg_income"],
			})
		}
		sectionMux.Lock()
		section["age_income_gender_distribution"] = ageIncome
		sectionMux.Unlock()
	}()

	go func() {
		defer wg.Done()
		dRows, err := queryRowsWithRetry(ctx, h.DB, disabilityQuery, args...)
		if err != nil {
			recordErr(fmt.Errorf("disability query failed: %w", err))
			return
		}
		defer dRows.Close()

		disability := []gin.H{}
		for dRows.Next() {
			var name string
			var value int
			if scanErr := dRows.Scan(&name, &value); scanErr != nil {
				recordErr(fmt.Errorf("disability scan failed: %w", scanErr))
				return
			}
			disability = append(disability, gin.H{"name": name, "value": value})
		}
		sectionMux.Lock()
		section["disability_distribution"] = disability
		sectionMux.Unlock()
	}()

	wg.Wait()
	return section, firstErr
}

func (h *DashboardSummaryHandler) fetchEducationSection(ctx context.Context, whereF string, args []interface{}) (gin.H, error) {
	section := defaultEducationSummary()

	// below_10th: NULL or empty QUALIFICATION only (UI label: "Qualification Not Available") — DB stores 10th, 12th, Graduation & Above only.
	query := injectWhere(`
		SELECT
			COUNT(DISTINCT fm.FAMILY_MEMBER_ID) AS total_population,
			COUNT(DISTINCT CASE
				WHEN UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'YES'
				THEN fm.FAMILY_MEMBER_ID
			END) AS literate_population,
			COUNT(DISTINCT CASE
				WHEN UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'NO'
					OR fm.EVER_ATTENDED_SCHOOL IS NULL
				THEN fm.FAMILY_MEMBER_ID
			END) AS illiterate_population,
			COUNT(DISTINCT CASE
				WHEN UPPER(TRIM(COALESCE(fm.CURRENTLY_PURSUING_EDUCATION, ''))) = 'YES'
				THEN fm.FAMILY_MEMBER_ID
			END) AS students_count,
			COUNT(DISTINCT CASE
				WHEN fm.DROP_OUT IS NOT NULL
					AND TRIM(fm.DROP_OUT) != ''
				THEN fm.FAMILY_MEMBER_ID
			END) AS dropout_count,
			COUNT(DISTINCT CASE
				WHEN TRIM(COALESCE(fm.QUALIFICATION, '')) = 'Graduation & Above'
				THEN fm.FAMILY_MEMBER_ID
			END) AS graduate_population,
			COUNT(DISTINCT CASE
				WHEN fm.QUALIFICATION IS NULL
					OR TRIM(fm.QUALIFICATION) = ''
				THEN fm.FAMILY_MEMBER_ID
			END) AS below_10th,
			COUNT(DISTINCT CASE
				WHEN TRIM(fm.QUALIFICATION) = '10th'
				THEN fm.FAMILY_MEMBER_ID
			END) AS tenth,
			COUNT(DISTINCT CASE
				WHEN TRIM(fm.QUALIFICATION) = '12th'
				THEN fm.FAMILY_MEMBER_ID
			END) AS twelfth,
			COUNT(DISTINCT CASE
				WHEN TRIM(fm.QUALIFICATION) = 'Graduation & Above'
				THEN fm.FAMILY_MEMBER_ID
			END) AS graduate_above
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f
		ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE __WHERE_CLAUSE__
	`, whereF)

	var totalPopulation, literate, illiterate, students, dropout, graduate, below10th, tenth, twelfth, graduateAbove int
	if err := queryRowWithRetry(ctx, h.DB, query, args, &totalPopulation, &literate, &illiterate, &students, &dropout, &graduate, &below10th, &tenth, &twelfth, &graduateAbove); err != nil {
		return section, err
	}

	literacyRate := 0.0
	if totalPopulation > 0 {
		literacyRate = (float64(literate) / float64(totalPopulation)) * 100
	}

	section["literate_population"] = literate
	section["illiterate_population"] = illiterate
	section["students_count"] = students
	section["dropout_count"] = dropout
	section["graduate_population"] = graduate
	section["literacy_rate"] = literacyRate
	section["qualification_distribution"] = gin.H{"below_10th": below10th, "tenth": tenth, "twelfth": twelfth, "graduate_above": graduateAbove}

	return section, nil
}

func (h *DashboardSummaryHandler) fetchEmploymentSection(ctx context.Context, whereF string, args []interface{}) (gin.H, error) {
	section := defaultEmploymentSummary()

	query := injectWhere(`
		SELECT
			COUNT(DISTINCT CASE
				WHEN TRIM(COALESCE(fm.OCCUPATION, '')) IN ('Salaried Job','Self Employed - Farm based','Self Employed- Non-farm based','Self Employed-Agri allied','Wage Work')
				THEN fm.FAMILY_MEMBER_ID
			END) AS employed_members,
			COUNT(DISTINCT CASE
				WHEN TRIM(COALESCE(fm.OCCUPATION, '')) IN ('Unemployed','Not Applicable')
				THEN fm.FAMILY_MEMBER_ID
			END) AS unemployed_members,
			COUNT(DISTINCT CASE
				WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Wage Work' OR fm.NATURE_WAGE_WORK IS NOT NULL
				THEN fm.FAMILY_MEMBER_ID
			END) AS daily_wage_workers,
			COUNT(DISTINCT CASE
				WHEN LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%driver%%' OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%electric%%' OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%mechanic%%' OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%tailor%%' OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%carpenter%%' OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%computer%%' OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%bank%%' OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%shop%%' OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%company worker%%' OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%security%%' OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%painter%%' OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%civil%%' OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%%technician%%'
				THEN fm.FAMILY_MEMBER_ID
			END) AS skilled_workers,
			COUNT(DISTINCT CASE
				WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Self Employed - Farm based'
				THEN fm.FAMILY_MEMBER_ID
			END) AS farm_based,
			COUNT(DISTINCT CASE
				WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Self Employed-Agri allied'
				THEN fm.FAMILY_MEMBER_ID
			END) AS agri_allied,
			COUNT(DISTINCT CASE
				WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Self Employed- Non-farm based'
				THEN fm.FAMILY_MEMBER_ID
			END) AS non_farm,
			COUNT(DISTINCT CASE
				WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Salaried Job'
				THEN fm.FAMILY_MEMBER_ID
			END) AS salaried,
			COUNT(DISTINCT CASE
				WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Wage Work'
				THEN fm.FAMILY_MEMBER_ID
			END) AS wage_workers,
			COUNT(DISTINCT CASE
				WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Housewife'
				THEN fm.FAMILY_MEMBER_ID
			END) AS housewife,
			COUNT(DISTINCT CASE
				WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Studying'
				THEN fm.FAMILY_MEMBER_ID
			END) AS students,
			COUNT(DISTINCT CASE
				WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Unemployed'
				THEN fm.FAMILY_MEMBER_ID
			END) AS unemployed,
			COUNT(DISTINCT CASE
				WHEN fm.OCCUPATION IS NULL OR TRIM(COALESCE(fm.OCCUPATION, '')) = ''
				THEN fm.FAMILY_MEMBER_ID
			END) AS other
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		WHERE __WHERE_CLAUSE__
	`, whereF)

	var employed, unemployed, dailyWage, skilled int
	var farmBased, agriAllied, nonFarm, salaried, wageWorkers, housewife, students, unemployedDist, other int
	if err := queryRowWithRetry(ctx, h.DB, query, args, &employed, &unemployed, &dailyWage, &skilled, &farmBased, &agriAllied, &nonFarm, &salaried, &wageWorkers, &housewife, &students, &unemployedDist, &other); err != nil {
		return section, err
	}

	section["employed_members"] = employed
	section["unemployed_members"] = unemployed
	section["daily_wage_workers"] = dailyWage
	section["skilled_workers"] = skilled
	section["occupation_distribution"] = gin.H{
		"farm_based": farmBased, "agri_allied": agriAllied, "non_farm": nonFarm, "salaried": salaried,
		"wage_workers": wageWorkers, "housewife": housewife, "students": students, "unemployed": unemployedDist, "other": other,
	}

	return section, nil
}

func (h *DashboardSummaryHandler) fetchAgricultureSection(ctx context.Context, whereF string, args []interface{}) (gin.H, error) {
	section := defaultAgricultureSummary()

	totalFarmersQuery := injectWhere(`
		SELECT COUNT(*) FROM FAMILY f WHERE f.OWN_AGRICULTURE_LAND = 'Yes' AND __WHERE_CLAUSE__
	`, whereF)
	noIrrigationQuery := injectWhere(`
		SELECT COUNT(*) FROM FAMILY f
		WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
		  AND (f.SOURCE_WATER_IRRIGATION IS NULL OR f.SOURCE_WATER_IRRIGATION = '' OR f.SOURCE_WATER_IRRIGATION = 'None' OR f.SOURCE_WATER_IRRIGATION = 'Rain Fed')
		  AND __WHERE_CLAUSE__
	`, whereF)
	landUtilQuery := injectWhere(`
		SELECT
			COALESCE(ROUND(SUM(t.total_land), 2), 0) AS total_land,
			COALESCE(ROUND(SUM(t.cultivated_land), 2), 0) AS cultivated_land,
			COALESCE(ROUND(SUM(t.total_land - t.cultivated_land), 2), 0) AS unused_land,
			COUNT(*) AS valid_records
		FROM (
			SELECT
				CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2)) AS total_land,
				CAST(f.LAND_UNDER_CULTIVATION_ACRES AS DECIMAL(12,2)) AS cultivated_land,
				f.DISTRICT_ID,
				f.TALUKA_ID,
				f.VILLAGE_ID
			FROM FAMILY f
			WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
			  AND f.AREA_AGRICULTURE_LAND_ACRES IS NOT NULL
			  AND f.LAND_UNDER_CULTIVATION_ACRES IS NOT NULL
			  AND TRIM(f.AREA_AGRICULTURE_LAND_ACRES) <> ''
			  AND TRIM(f.LAND_UNDER_CULTIVATION_ACRES) <> ''
			  AND f.AREA_AGRICULTURE_LAND_ACRES REGEXP '^[0-9]*\\.?[0-9]+$'
			  AND f.LAND_UNDER_CULTIVATION_ACRES REGEXP '^[0-9]*\\.?[0-9]+$'
			  AND CAST(f.LAND_UNDER_CULTIVATION_ACRES AS DECIMAL(12,2)) <= CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2))
			  AND CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2)) BETWEEN 0 AND 500
			  AND __WHERE_CLAUSE__
		) t
	`, whereF)
	invalidQuery := injectWhere(`
		SELECT COUNT(*)
		FROM FAMILY f
		WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
		  AND __WHERE_CLAUSE__
		  AND (
			f.AREA_AGRICULTURE_LAND_ACRES IS NULL
			OR f.LAND_UNDER_CULTIVATION_ACRES IS NULL
			OR TRIM(f.AREA_AGRICULTURE_LAND_ACRES) = ''
			OR TRIM(f.LAND_UNDER_CULTIVATION_ACRES) = ''
			OR f.AREA_AGRICULTURE_LAND_ACRES NOT REGEXP '^[0-9]*\\.?[0-9]+$'
			OR f.LAND_UNDER_CULTIVATION_ACRES NOT REGEXP '^[0-9]*\\.?[0-9]+$'
			OR CAST(f.LAND_UNDER_CULTIVATION_ACRES AS DECIMAL(12,2)) > CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2))
			OR CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2)) > 500
		  )
	`, whereF)
	landDistributionQuery := injectWhere(`
		SELECT
			CASE
				WHEN CAST(TRIM(f.AREA_AGRICULTURE_LAND_ACRES) AS DECIMAL(10,2)) = 0 THEN 'Landless'
				WHEN CAST(TRIM(f.AREA_AGRICULTURE_LAND_ACRES) AS DECIMAL(10,2)) > 0
				     AND CAST(TRIM(f.AREA_AGRICULTURE_LAND_ACRES) AS DECIMAL(10,2)) <= 2.5 THEN 'Small'
				WHEN CAST(TRIM(f.AREA_AGRICULTURE_LAND_ACRES) AS DECIMAL(10,2)) > 2.5
				     AND CAST(TRIM(f.AREA_AGRICULTURE_LAND_ACRES) AS DECIMAL(10,2)) <= 10 THEN 'Medium'
				ELSE 'Large'
			END AS category,
			COUNT(*) AS cnt
		FROM FAMILY f
		WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
		  AND f.AREA_AGRICULTURE_LAND_ACRES IS NOT NULL
		  AND TRIM(f.AREA_AGRICULTURE_LAND_ACRES) <> ''
		  AND TRIM(f.AREA_AGRICULTURE_LAND_ACRES) REGEXP '^[0-9]*\\.?[0-9]+$'
		  AND __WHERE_CLAUSE__
		GROUP BY category
	`, whereF)
	cropQuery := injectWhere(`
		SELECT season, crop, SUM(cnt) AS cnt
		FROM (
			SELECT 'Kharif' AS season, TRIM(f.CULTIVATING_DURING_KHARIF_SEASON) AS crop, COUNT(*) AS cnt
			FROM FAMILY f
			WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
			  AND f.CULTIVATING_DURING_KHARIF_SEASON IS NOT NULL
			  AND TRIM(f.CULTIVATING_DURING_KHARIF_SEASON) != ''
			  AND __WHERE_CLAUSE__
			GROUP BY TRIM(f.CULTIVATING_DURING_KHARIF_SEASON)
			UNION ALL
			SELECT 'Rabi' AS season, TRIM(f.CULTIVATING_DURING_RABI_SEASON) AS crop, COUNT(DISTINCT f.EXTERNAL_FAMILY_ID) AS cnt
			FROM FAMILY f
			WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
			  AND LOWER(TRIM(COALESCE(f.TAKING_CROPS_RABI_SEASON, ''))) = 'yes'
			  AND f.CULTIVATING_DURING_RABI_SEASON IS NOT NULL
			  AND TRIM(f.CULTIVATING_DURING_RABI_SEASON) != ''
			  AND LOWER(TRIM(f.CULTIVATING_DURING_RABI_SEASON)) NOT IN ('yes', 'no')
			  AND __WHERE_CLAUSE__
			GROUP BY TRIM(f.CULTIVATING_DURING_RABI_SEASON)
		) merged
		GROUP BY season, crop
		ORDER BY cnt DESC
		LIMIT 5
	`, whereF)
	kharifCountQuery := injectWhere(`
		SELECT COUNT(*) FROM FAMILY f
		WHERE f.CULTIVATING_DURING_KHARIF_SEASON IS NOT NULL
		  AND f.CULTIVATING_DURING_KHARIF_SEASON != ''
		  AND f.CULTIVATING_DURING_KHARIF_SEASON != 'No'
		  AND __WHERE_CLAUSE__
	`, whereF)
	rabiCountQuery := injectWhere(`
		SELECT COUNT(*) FROM FAMILY f
		WHERE f.TAKING_CROPS_RABI_SEASON IS NOT NULL
		  AND f.TAKING_CROPS_RABI_SEASON != ''
		  AND f.TAKING_CROPS_RABI_SEASON != 'No'
		  AND __WHERE_CLAUSE__
	`, whereF)

	var totalFarmers, noIrrigation int
	var totalLand, cultivatedLand, unusedLand float64
	var validRecords, invalidRecords int64
	var landDistribution []gin.H
	var seasonRows []gin.H
	var kharifCount, rabiCount int

	var firstErr error
	errMux := sync.Mutex{}
	recordErr := func(err error) {
		if err == nil {
			return
		}
		errMux.Lock()
		if firstErr == nil {
			firstErr = err
		}
		errMux.Unlock()
	}

	var wg sync.WaitGroup
	wg.Add(7)

	go func() {
		defer wg.Done()
		if err := queryRowWithRetry(ctx, h.DB, totalFarmersQuery, args, &totalFarmers); err != nil {
			recordErr(fmt.Errorf("total farmers query failed: %w", err))
		}
	}()

	go func() {
		defer wg.Done()
		if err := queryRowWithRetry(ctx, h.DB, noIrrigationQuery, args, &noIrrigation); err != nil {
			recordErr(fmt.Errorf("irrigation query failed: %w", err))
		}
	}()

	go func() {
		defer wg.Done()
		if err := queryRowWithRetry(ctx, h.DB, landUtilQuery, args, &totalLand, &cultivatedLand, &unusedLand, &validRecords); err != nil {
			recordErr(fmt.Errorf("land utilization query failed: %w", err))
		}
	}()

	go func() {
		defer wg.Done()
		if err := queryRowWithRetry(ctx, h.DB, invalidQuery, args, &invalidRecords); err != nil {
			recordErr(fmt.Errorf("invalid records query failed: %w", err))
		}
	}()

	go func() {
		defer wg.Done()
		landRows, err := queryRowsWithRetry(ctx, h.DB, landDistributionQuery, args...)
		if err != nil {
			recordErr(fmt.Errorf("land distribution query failed: %w", err))
			return
		}
		defer landRows.Close()

		labelTotals := map[string]int{}
		for landRows.Next() {
			var category string
			var cnt int
			if scanErr := landRows.Scan(&category, &cnt); scanErr != nil {
				recordErr(fmt.Errorf("land distribution scan failed: %w", scanErr))
				return
			}
			canonical := canonicalLandDistributionLabel(category)
			labelTotals[canonical] += cnt
		}
		canonicalOrder := []string{
			"Landless",
			"Small",
			"Medium",
			"Large",
		}
		tmp := make([]gin.H, 0, len(canonicalOrder))
		for _, label := range canonicalOrder {
			tmp = append(tmp, gin.H{"label": label, "count": labelTotals[label]})
		}
		landDistribution = tmp
	}()

	go func() {
		defer wg.Done()
		// cropQuery injects the same location filter twice (Kharif + Rabi subqueries),
		// so bind args must be provided twice to match all placeholders.
		cropArgs := append([]interface{}{}, args...)
		cropArgs = append(cropArgs, args...)
		cropRows, err := queryRowsWithRetry(ctx, h.DB, cropQuery, cropArgs...)
		if err != nil {
			recordErr(fmt.Errorf("season crop query failed: %w", err))
			return
		}
		defer cropRows.Close()

		tmp := []gin.H{}
		for cropRows.Next() {
			var season, crop string
			var cnt int
			if scanErr := cropRows.Scan(&season, &crop, &cnt); scanErr != nil {
				recordErr(fmt.Errorf("season crop scan failed: %w", scanErr))
				return
			}
			tmp = append(tmp, gin.H{"season": season, "crop": crop, "count": cnt})
		}
		seasonRows = tmp
	}()

	go func() {
		defer wg.Done()
		if err := queryRowWithRetry(ctx, h.DB, kharifCountQuery, args, &kharifCount); err != nil {
			recordErr(fmt.Errorf("kharif count query failed: %w", err))
		}
		if err := queryRowWithRetry(ctx, h.DB, rabiCountQuery, args, &rabiCount); err != nil {
			recordErr(fmt.Errorf("rabi count query failed: %w", err))
		}
	}()

	wg.Wait()

	cultivatedPercent := 0.0
	unusedPercent := 0.0
	if totalLand > 0 {
		cultivatedPercent = (cultivatedLand * 100) / totalLand
		unusedPercent = (unusedLand * 100) / totalLand
	}

	section["totalFarmers"] = totalFarmers
	section["farmersWithoutIrrigation"] = noIrrigation
	section["landUtilizationSummary"] = gin.H{
		"total_land": totalLand, "cultivated_land": cultivatedLand, "unused_land": unusedLand,
		"valid_records": validRecords, "invalid_records": invalidRecords,
		"cultivated_percent": cultivatedPercent, "unused_percent": unusedPercent,
	}
	section["landUtilizationRows"] = []gin.H{{
		"AREA_AGRICULTURE_LAND_ACRES":  totalLand,
		"LAND_UNDER_CULTIVATION_ACRES": cultivatedLand,
		"unused_land_acres":            unusedLand,
		"valid_records":                validRecords,
		"invalid_records":              invalidRecords,
	}}
	section["landDistribution"] = landDistribution
	section["seasonCropRows"] = seasonRows
	section["kharifFarmers"] = kharifCount
	section["rabiFarmers"] = rabiCount

	return section, firstErr
}
