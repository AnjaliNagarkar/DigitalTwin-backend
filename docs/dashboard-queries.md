# Dashboard Query Documentation

> **Technical reference** for all backend/API/database queries and data flow used by the **Village Command Center** dashboard (Agriculture module).

| Item | Value |
|------|-------|
| **UI page** | Village Command Center |
| **Frontend component** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Route** | `/agriculture/dashboard` (via `AgricultureLayout.vue`) |
| **Primary data API** | `GET /api/dashboard/summary` |
| **Filter options API** | `GET /api/location-options` |

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Main Dashboard API](#2-main-dashboard-api)
3. [Common Patterns & Shared Logic](#3-common-patterns--shared-logic)
4. [Population Section](#4-population-section)
   - 4.1 Total Households
   - 4.2 Total Population
   - 4.3 BPL Status
5. [Demographics Section](#5-demographics-section)
   - 5.1 Gender Distribution
   - 5.2 Divyang Distribution
   - 5.3 Age-wise Family Income Distribution
6. [Education Section](#6-education-section)
   - 6.1 Education Intelligence
   - 6.2 Qualification Distribution
   - 6.3 Literacy Rate
7. [Employment Section](#7-employment-section)
   - 7.1 Employment Insights
   - 7.2 Occupation Distribution
8. [Agriculture Section](#8-agriculture-section)
   - 8.1 Agriculture Intelligence (Summary Cards)
   - 8.2 Land Holdings Distribution
   - 8.3 Land Utilization
   - 8.4 Season-wise Crops
9. [Location Filters](#9-location-filters)
10. [Parallel Processing](#10-parallel-processing)
11. [Cache Mechanism](#11-cache-mechanism)
12. [Performance Notes](#12-performance-notes)
13. [Important Business Logic](#13-important-business-logic)
14. [File Mapping Summary](#14-file-mapping-summary)

---

## 1. Architecture Overview

### End-to-End Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Browser: Dashboard.vue (Village Command Center)                            │
├─────────────────────────────────────────────────────────────────────────────┤
│  onMounted / Apply / Reset                                                  │
│    ├─ loadLocationOptions()  →  GET /api/location-options                   │
│    └─ fetchDashboardData()   →  GET /api/dashboard/summary                  │
│           │                                                                 │
│           ▼                                                                 │
│  Parse JSON: { population, demographics, education, employment,             │
│                agriculture, partial_errors? }                               │
│           │                                                                 │
│           ▼                                                                 │
│  Vue refs + computed properties + Chart.js / ApexCharts render widgets      │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  frontend/src/api/index.js                                                  │
│    getDashboardSummary(params)  →  fetchJSON('/dashboard/summary?...')      │
│    getLocationOptions(params)   →  fetchJSON('/location-options?...')       │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  main.go (Gin router, protected routes)                                     │
│    GET /dashboard/summary  → DashboardSummaryHandler.GetDashboardSummary    │
│    GET /location-options   → LocationHandler.GetLocationOptions             │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    ▼                               ▼
         ┌──────────────────┐            ┌──────────────────────┐
         │ Cache lookup     │            │ location_options.go  │
         │ (5 min TTL)      │            │ 3 SQL queries on     │
         │ on cache miss:   │            │ master tables        │
         │ 5 parallel       │            └──────────────────────┘
         │ goroutines       │
         └────────┬─────────┘
                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  handlers/dashboard_summary.go                                              │
│    fetchPopulationSection    → 4–5 SQL queries (parallel inside section)    │
│    fetchDemographicsSection  → 5 SQL queries (parallel inside section)      │
│    fetchEducationSection     → 1 SQL query                                  │
│    fetchEmploymentSection    → 1 SQL query                                  │
│    fetchAgricultureSection   → 7 SQL queries (parallel inside section)      │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  MySQL tables: FAMILY, FAMILY_MEMBER, district_master, taluka_master,       │
│                village_master, grampanchayat_master                         │
│  No materialized views. No stored procedures for dashboard.                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Quick Reference

| Component | Detail |
|-----------|--------|
| **Frontend** | `frontend/src/views/agriculture/Dashboard.vue` |
| **API client** | `frontend/src/api/index.js` |
| **Backend handler** | `handlers/dashboard_summary.go` |
| **Route** | `GET /api/dashboard/summary` |
| **Cache TTL** | 5 minutes |
| **Dashboard sections** | 5 (population, demographics, education, employment, agriculture) |
| **Total queries / cache miss** | ~18–19 SQL queries |
| **Concurrency model** | 5 top-level goroutines + nested parallelism per section |

### Section Query Counts

| Section | Handler Function | Internal Queries | Parallelism |
|---------|-----------------|-----------------|-------------|
| Population | `fetchPopulationSection` | 4–5 | 4 goroutines + sequential BPL |
| Demographics | `fetchDemographicsSection` | 5 | 5 goroutines |
| Education | `fetchEducationSection` | 1 | Sequential |
| Employment | `fetchEmploymentSection` | 1 | Sequential |
| Agriculture | `fetchAgricultureSection` | 7 | 7 goroutines |

### Frontend Data Flow

1. **`onMounted`** — calls `fetchDashboardData()` and `loadLocationOptions()` in parallel (`Promise.allSettled`).
2. **`applyFilters`** — builds `district_ids`, `taluka_ids`, `village_ids` from multi-select state; calls `fetchDashboardData(params)`.
3. **`resetFilters`** — clears selections; reloads dashboard and location options.
4. **`fetchDashboardData`** — single `getDashboardSummary(params)` call; maps response into refs:
   - `populationStats` ← `payload.population`
   - `demographics` ← `payload.demographics` (+ `applyDemographicsData` for age/income chart)
   - `education` ← `payload.education`
   - `employment` ← `payload.employment`
   - `agriculture` ← `payload.agriculture`
   - `bplDistribution` ← `payload.demographics.bpl_distribution` **or** `payload.population.bpl_distribution`
5. **Charts** — CSS donut (gender, divyang, BPL), Chart.js (age/income), ApexCharts (land utilization donut, season crops bar).

### API Layer

- Base URL: `/api` (`frontend/src/api/index.js`, `const BASE = '/api'`).
- Auth: `Authorization: Bearer <token>` when present.
- Dashboard timeout: `TIMEOUT_DATA` = 30 seconds.
- `cache: 'no-store'` on fetch (browser does not cache HTTP responses).

### Backend Handler Flow

1. **`GetDashboardSummary`** reads query params (singular and plural forms).
2. **Cache check** — in-memory map keyed by location filter string.
3. On miss: build `whereF, args` via `buildOptionalLocationFilterWithArrays("f", ...)`.
4. Spawn **5 goroutines** (population, demographics, education, employment, agriculture).
5. Merge results; attach `partial_errors` if any section failed.
6. Store in cache (5 min TTL); return JSON.

### Database Notes

- All dashboard metrics are **live aggregations** on `FAMILY` / `FAMILY_MEMBER`.
- Location filters append `AND f.DISTRICT_ID IN (...)`, `f.TALUKA_ID IN (...)`, `f.VILLAGE_ID IN (...)` (alias `f` on `FAMILY`).
- Member-level metrics join: `FAMILY_MEMBER fm JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID`.
- Placeholder `__WHERE_CLAUSE__` in SQL templates is replaced by the built filter clause via `injectWhere()`.

### APIs Not Used by This Dashboard

The following exist elsewhere in the project but are **not** called from `Dashboard.vue`:

- `GET /api/insights/agriculture`
- `GET /api/population/dashboard`
- `GET /api/population/demographics`
- `GET /api/population/education`
- `GET /api/population/employment`

---

## 2. Main Dashboard API

```
GET /api/dashboard/summary
```

**Authentication:** Protected route (Bearer token required).

### Request Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `district_id` | string | Single district ID (legacy) |
| `district_ids` | string | Comma-separated district IDs (used by Dashboard.vue) |
| `taluka_id` | string | Single taluka ID |
| `taluka_ids` | string | Comma-separated taluka IDs |
| `village_id` | string | Single village ID |
| `village_ids` | string | Comma-separated village IDs |

> Empty or invalid values (`0`, `null`, `undefined`) are ignored.

### Response Structure

```json
{
  "population": {
    "total_population": 0,
    "total_households": 0,
    "working_population": 0,
    "dependent_population": 0,
    "bpl_distribution": { "bpl": 0, "non_bpl": 0, "total_households": 0 }
  },
  "demographics": {
    "gender_distribution": { "male": 0, "female": 0, "other": 0 },
    "age_distribution": { "age_0_5": 0, "age_6_18": 0, "age_19_35": 0, "age_36_60": 0, "age_60_plus": 0 },
    "age_income_gender_distribution": [
      { "age_group": "18-30", "families": 0, "avg_income": 0.0 }
    ],
    "total_divyang": 0,
    "disability_distribution": [{ "name": "Visual Disability", "value": 0 }]
  },
  "education": {
    "literate_population": 0,
    "illiterate_population": 0,
    "students_count": 0,
    "dropout_count": 0,
    "graduate_population": 0,
    "literacy_rate": 0,
    "qualification_distribution": {
      "below_10th": 0, "tenth": 0, "twelfth": 0, "graduate_above": 0
    }
  },
  "employment": {
    "employed_members": 0,
    "unemployed_members": 0,
    "daily_wage_workers": 0,
    "skilled_workers": 0,
    "occupation_distribution": {
      "farm_based": 0, "agri_allied": 0, "non_farm": 0, "salaried": 0,
      "wage_workers": 0, "housewife": 0, "students": 0, "unemployed": 0, "other": 0
    }
  },
  "agriculture": {
    "totalFarmers": 0,
    "farmersWithoutIrrigation": 0,
    "landDistribution": [{ "label": "Small", "count": 0 }],
    "landUtilizationRows": [],
    "landUtilizationSummary": {
      "total_land": 0, "cultivated_land": 0, "unused_land": 0,
      "valid_records": 0, "invalid_records": 0,
      "cultivated_percent": 0, "unused_percent": 0
    },
    "seasonCropRows": [{ "season": "Kharif", "crop": "Rice", "count": 0 }],
    "kharifFarmers": 0,
    "rabiFarmers": 0
  },
  "partial_errors": { "section_name": "error message" }
}
```

> `partial_errors` is present only when one or more sections fail; other sections still return data or defaults.

### Handler Reference

| Item | Value |
|------|-------|
| **Frontend file** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Frontend function** | `fetchDashboardData(params)` |
| **API wrapper** | `frontend/src/api/index.js` → `getDashboardSummary(params)` |
| **Route** | `main.go` → `protected.GET("/dashboard/summary", dashboardSummaryHandler.GetDashboardSummary)` |
| **Handler** | `handlers/dashboard_summary.go` → `(*DashboardSummaryHandler).GetDashboardSummary` |
| **Section fetchers** | `fetchPopulationSection`, `fetchDemographicsSection`, `fetchEducationSection`, `fetchEmploymentSection`, `fetchAgricultureSection` |
| **Cache key** | `dashboard_{districtKey}_{talukaKey}_{villageKey}` |
| **Cache TTL** | 5 minutes (`dashboardSummaryCacheTTL`) |
| **Cache storage** | `DashboardSummaryHandler.cache map[string]dashboardCacheItem` |
| **Cache concurrency** | `sync.RWMutex` (`cacheMux`) |

---

## 3. Common Patterns & Shared Logic

### 3.1 Location Filter Placeholder

All section queries use `__WHERE_CLAUSE__`, injected at runtime by `injectWhere()`:

```sql
-- Example injected clause (with district + taluka + village filters):
1=1 AND f.DISTRICT_ID IN (?) AND f.TALUKA_ID IN (?) AND f.VILLAGE_ID IN (?)
```

Without filters, `__WHERE_CLAUSE__` resolves to `1=1`.

### 3.2 JOIN Duplication Problem & Fix

**Root cause:** The `FAMILY` table can contain multiple rows with the same `EXTERNAL_FAMILY_ID`. When `FAMILY_MEMBER` is joined to `FAMILY` on that key, each member row is duplicated once per matching family row. Any `SUM(CASE WHEN … THEN 1 END)` or `COUNT(*)` on the joined result inflates all member-level metrics.

**Fix applied:** All member-level metrics use the duplicate-safe aggregation pattern:

```sql
COUNT(DISTINCT CASE
  WHEN <condition>
  THEN fm.FAMILY_MEMBER_ID
END) AS <metric>
```

This counts each person (`FAMILY_MEMBER_ID`) exactly once regardless of how many `FAMILY` rows share the same `EXTERNAL_FAMILY_ID`.

> **All sections** (Population, Demographics, Education, Employment) use this pattern. Notes in individual metric sections that reference "duplicate-safe" point here.

### 3.3 Common SQL Patterns Reference

| Pattern | Purpose | Used In |
|---------|---------|---------|
| `COUNT(DISTINCT CASE WHEN ... THEN fm.FAMILY_MEMBER_ID END)` | Duplicate-safe member count | Demographics, Education, Employment |
| `COUNT(DISTINCT f.FAMILY_ID)` | Duplicate-safe household count | Population (BPL) |
| `UPPER(TRIM(COALESCE(column, '')))` | Normalize DB strings before comparison | All sections |
| `__WHERE_CLAUSE__` placeholder | Dynamic location filter injection | All section queries |
| `injectWhere(query, whereClause)` | Replaces placeholder at runtime | `dashboard_summary.go` |
| `buildOptionalLocationFilterWithArrays` | Builds `IN (...)` clauses from comma-separated IDs | `dashboard_summary.go` |

### 3.4 Go Handler Patterns

| Pattern | Detail |
|---------|--------|
| Parallel goroutines | `sync.WaitGroup` + channels for each section and sub-query |
| In-memory cache | `map[string]dashboardCacheItem` with `sync.RWMutex` |
| Stale cache fallback | On section timeout/failure, returns last-known-good section data |
| Per-section timeouts | Independent `context.WithTimeout` per section (no shared timeout) |
| Stampede guard | In-flight map (`inflight map[string]chan struct{}`) prevents duplicate DB calls |
| `partial_errors` | Failed sections report error; other successful sections still returned |
| Column detection | `ColumnChecker` used before optional queries (e.g., BPL columns) |

---

## 4. Population Section

**Handler function:** `fetchPopulationSection` (`handlers/dashboard_summary.go`)  
**Parallelism:** 4 goroutines for main queries + 1 sequential BPL query (if columns exist)

---

### 4.1 Total Households

#### Purpose
Count of all surveyed families within the current filter scope.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `populationMetrics` (metric card) |
| **API field** | `response.population.total_households` |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchPopulationSection` |
| **Query var** | `householdsQuery` |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query

```sql
SELECT COUNT(*)
FROM FAMILY f
WHERE __WHERE_CLAUSE__
```

#### Tables Used

- `FAMILY`

#### Notes

- Displayed as the top summary card with home icon.
- `working_population` and `dependent_population` are fetched in the same section but **not currently displayed** in the UI.

---

### 4.2 Total Population

#### Purpose
Count of all family members in scope (not distinct families).

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `populationMetrics` |
| **API field** | `response.population.total_population` |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchPopulationSection` |
| **Query var** | `populationQuery` |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query

```sql
SELECT COUNT(*)
FROM FAMILY_MEMBER fm
JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
WHERE __WHERE_CLAUSE__
```

#### Tables Used

- `FAMILY`
- `FAMILY_MEMBER`

#### Notes

- Counts all family member rows in scope — not distinct families.

---

### 4.3 BPL Status

#### Purpose
Count of households classified as Below Poverty Line (BPL), used for the dashboard donut chart.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `bplSegments`, `bplTotal`, `bplPieStyle` |
| **API field** | `payload.population.bpl_distribution` (primary; fallback: `demographics.bpl_distribution`) |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchPopulationSection` |
| **Query var** | `bplHouseholdQuery` (conditional) |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query

Only runs if `ColumnChecker` reports `FAMILY_BELONG_BPL_CATEGORY` and/or `RATION_CARD_TYPE` exist. The `AND (...)` clause is built from whichever of those columns exist:

```sql
SELECT COUNT(DISTINCT f.FAMILY_ID) AS bpl_households
FROM FAMILY f
WHERE __WHERE_CLAUSE__
  AND (
    UPPER(TRIM(COALESCE(f.FAMILY_BELONG_BPL_CATEGORY, ''))) = 'YES'
    OR UPPER(TRIM(COALESCE(f.RATION_CARD_TYPE, ''))) IN ('BPL', 'AAY')
  )
```

#### Tables Used

- `FAMILY`

#### Notes / Edge Cases

- **Duplicate-safe:** `COUNT(DISTINCT f.FAMILY_ID)` counts each household once. See §3.2.
- `non_bpl` = `total_households - bpl` — computed in Go, not SQL.
- `bpl_distribution` is stored under `population` in the API response (not `demographics`).
- Donut shows BPL vs. Non-BPL families.
- **2D Map:** BPL Status coloring uses client-side logic (`getBplStatus`) on loaded `houses` data — this query applies to the **Dashboard donut only**.

---

## 5. Demographics Section

**Handler function:** `fetchDemographicsSection` (`handlers/dashboard_summary.go`)  
**Parallelism:** 5 goroutines — `genderQuery`, `ageQuery`, `divyangQuery`, `ageIncomeQuery`, `disabilityQuery`

---

### 5.1 Gender Distribution

#### Purpose
Count of male, female, and other members for the gender donut chart.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `genderSegments`, `genderTotal`, `genderPieStyle` |
| **API field** | `response.demographics.gender_distribution` (`male`, `female`, `other`) |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchDemographicsSection` |
| **Query var** | `genderQuery` |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query

```sql
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
JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
WHERE __WHERE_CLAUSE__
```

#### Tables Used

- `FAMILY`
- `FAMILY_MEMBER` (column: `GENDER`)

#### Notes

- Duplicate-safe via `COUNT(DISTINCT CASE … THEN fm.FAMILY_MEMBER_ID END)`. See §3.2.
- Rendered as CSS `conic-gradient` donut (not Chart.js).
- Header total = sum of male + female + other.

---

### 5.2 Divyang Distribution

#### Purpose
Total count and disability-category breakdown of persons with disabilities (Divyang).

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `divyangSegments`, `divyangTotal`, `divyangPieStyle`, `mapDisabilityGroup` |
| **API fields** | `response.demographics.disability_distribution`, `response.demographics.total_divyang` |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchDemographicsSection` |
| **Query vars** | `divyangQuery` (total count), `disabilityQuery` (category breakdown) |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query — Total Count

```sql
SELECT COUNT(*)
FROM FAMILY_MEMBER fm
JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
WHERE UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES'
  AND __WHERE_CLAUSE__
```

#### SQL Query — Distribution (top 8 groups)

```sql
SELECT
  CASE
    WHEN DISABILITY_CATEGORY LIKE '%Blind%'
      OR DISABILITY_CATEGORY LIKE '%Low vision%'
    THEN 'Visual Disability'

    WHEN DISABILITY_CATEGORY LIKE '%Locomotor%'
      OR DISABILITY_CATEGORY LIKE '%Cerebral%'
      OR DISABILITY_CATEGORY LIKE '%Muscular%'
      OR DISABILITY_CATEGORY LIKE '%Dwarf%'
    THEN 'Locomotor Disability'

    WHEN DISABILITY_CATEGORY LIKE '%Mental%'
      OR DISABILITY_CATEGORY LIKE '%Autism%'
      OR DISABILITY_CATEGORY LIKE '%Intellectual%'
      OR DISABILITY_CATEGORY LIKE '%Learning%'
    THEN 'Intellectual Disability'

    WHEN DISABILITY_CATEGORY LIKE '%Hearing%'
    THEN 'Hearing Disability'

    WHEN DISABILITY_CATEGORY LIKE '%Speech%'
    THEN 'Speech Disability'

    WHEN DISABILITY_CATEGORY LIKE '%Multiple%'
    THEN 'Multiple Disabilities'

    WHEN DISABILITY_CATEGORY LIKE '%Parkinson%'
      OR DISABILITY_CATEGORY LIKE '%Sclerosis%'
      OR DISABILITY_CATEGORY LIKE '%Sickle%'
      OR DISABILITY_CATEGORY LIKE '%Thalassemia%'
    THEN 'Chronic Conditions'

    ELSE 'Other'
  END AS disability_group,

  COUNT(DISTINCT fm.FAMILY_MEMBER_ID) AS total

FROM FAMILY_MEMBER fm
JOIN FAMILY f ON fm.EXTERNAL_FAMILY_ID = f.EXTERNAL_FAMILY_ID

WHERE UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES'
  AND __WHERE_CLAUSE__

GROUP BY disability_group
ORDER BY total DESC
LIMIT 8
```

#### Tables Used

- `FAMILY`
- `FAMILY_MEMBER` (columns: `DIVYANG`, `DISABILITY_CATEGORY`)

#### Notes / Edge Cases

- Duplicate-safe via `COUNT(DISTINCT fm.FAMILY_MEMBER_ID)` in breakdown. See §3.2.
- Category `CASE` mapping is unchanged; new disability keywords require a new `WHEN` block.
- Frontend maps API labels to i18n keys via `DISABILITY_AGGREGATE_LABELS` / `mapDisabilityGroup`.
- `divyangTotal` prefers `total_divyang` from API; else sums segment values.
- Unmatched disability names are shown as the raw API `name` value.

---

### 5.3 Age-wise Family Income Distribution

#### Purpose
Distribution of families by head-of-household age group, with average annual income per group. Rendered as a Chart.js bar chart.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `applyDemographicsData`, `ageIncomeGenderSegments`, `syncAgeIncomeGenderChart` |
| **API field** | `response.demographics.age_income_gender_distribution` |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchDemographicsSection` |
| **Query var** | `ageIncomeQuery` |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query

```sql
SELECT
  age_group,
  COUNT(DISTINCT EXTERNAL_FAMILY_ID) AS families,
  AVG(income) AS avg_income
FROM (
  SELECT
    f.EXTERNAL_FAMILY_ID,

    CAST(
      NULLIF(TRIM(f.ANNUAL_INCOME), '')
      AS DECIMAL(15,2)
    ) AS income,

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
        MIN(
          CASE
            WHEN LOWER(TRIM(COALESCE(fm.RELATION_FAMILY_HEAD, '')))
                 IN ('head', 'self', 'head of family')
            THEN STR_TO_DATE(fm.DOB, '%d-%m-%Y')
          END
        ),
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
```

#### Tables Used

- `FAMILY` (column: `ANNUAL_INCOME`)
- `FAMILY_MEMBER` (columns: `DOB`, `RELATION_FAMILY_HEAD`)

#### Notes / Edge Cases

- Duplicate-safe via `COUNT(DISTINCT EXTERNAL_FAMILY_ID)`. See §3.2.
- Age groups: `18-30`, `31-45`, `46-60`, `60+`.
- **DOB priority:** head-of-household DOB preferred; fallback to earliest member DOB in family.
- Only families with a valid `ANNUAL_INCOME` and a resolvable DOB ≥ 18 years are included.
- Chart: bar height = `families`; data label = average income (₹).
- `age_distribution` (0–5, 6–18, etc.) is queried separately in `ageQuery` but **not displayed** in the current UI.

---

## 6. Education Section

**Handler function:** `fetchEducationSection` (`handlers/dashboard_summary.go`)  
**Parallelism:** Single aggregated query — all education metrics in one SQL call.

---

### 6.1 Education Intelligence

#### Purpose
Five summary stat cards: literate population, illiterate population, students, dropouts, graduates.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `educationMetrics` (5 mini-stat cards) |
| **API fields** | `education.literate_population`, `illiterate_population`, `students_count`, `dropout_count`, `graduate_population` |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchEducationSection` |
| **Query var** | Single aggregated `query` |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query

```sql
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
  END) AS qualification_not_available,

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
JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
WHERE __WHERE_CLAUSE__
```

#### Tables Used

- `FAMILY`
- `FAMILY_MEMBER` (columns: `EVER_ATTENDED_SCHOOL`, `CURRENTLY_PURSUING_EDUCATION`, `DROP_OUT`, `QUALIFICATION`)

#### Notes / Edge Cases

- Duplicate-safe via `COUNT(DISTINCT fm.FAMILY_MEMBER_ID)`. See §3.2.
- **One query** powers both Education Intelligence cards and Qualification Distribution (§6.2).
- **Verification (unfiltered scope):** Dashboard total was **69,611** vs. actual distinct members **69,598** — 13 duplicate join rows before the fix.
- See §13 (Important Business Logic) for `dropout_count` and `below_10th` exact logic.

---

### 6.2 Qualification Distribution

#### Purpose
Horizontal bar chart showing member counts across four qualification buckets.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `qualificationSegments`, `qualificationTotal`, `qualificationBarWidth` |
| **API field** | `response.education.qualification_distribution` |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchEducationSection` — same query as §6.1 |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query

Same SQL as **§6.1 Education Intelligence** — fields `qualification_not_available` (`below_10th`), `tenth`, `twelfth`, `graduate_above`.

#### Tables Used

- `FAMILY`
- `FAMILY_MEMBER`

#### Notes

- Uses the same duplicate-safe query as §6.1.
- Bar chart sorted descending by count in frontend.
- Labels: Qualification Not Available, 10th, 12th, Graduation & Above (via `t('agriDashboard.below10th')` etc.).
- See §13.7 for the `below_10th` business logic.

---

### 6.3 Literacy Rate

#### Purpose
Percentage of literate persons among the total member population for the current filter scope.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `literacyRateLabel` |
| **API field** | `response.education.literacy_rate` |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchEducationSection` |
| **Computation** | Go: `(literate_population / total_population) * 100` when `total_population > 0` |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query

No separate SQL — derived from §6.1 Education Intelligence query aggregates.

#### Tables Used

- (derived from the same query as §6.1)

#### Notes

- Formula: `literacy_rate = (literate_population / total_population) × 100`.
- Both values come from the duplicate-safe education query (`COUNT(DISTINCT fm.FAMILY_MEMBER_ID)`).
- Footer text shows literate count vs. `population.total_population` from the Population section (separate query).

---

## 7. Employment Section

**Handler function:** `fetchEmploymentSection` (`handlers/dashboard_summary.go`)  
**Parallelism:** Single aggregated query — all employment metrics in one SQL call.

---

### 7.1 Employment Insights

#### Purpose
Four summary stat cards: employed members, unemployed members, daily wage workers, skilled workers.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `employmentMetrics` (4 mini-stat cards) |
| **API fields** | `employment.employed_members`, `unemployed_members`, `daily_wage_workers`, `skilled_workers` |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchEmploymentSection` |
| **Query var** | Single aggregated `query` |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query

```sql
SELECT
  COUNT(DISTINCT CASE
    WHEN TRIM(COALESCE(fm.OCCUPATION, '')) IN (
      'Salaried Job', 'Self Employed - Farm based', 'Self Employed- Non-farm based',
      'Self Employed-Agri allied', 'Wage Work'
    )
    THEN fm.FAMILY_MEMBER_ID
  END) AS employed_members,

  COUNT(DISTINCT CASE
    WHEN TRIM(COALESCE(fm.OCCUPATION, '')) IN ('Unemployed', 'Not Applicable')
    THEN fm.FAMILY_MEMBER_ID
  END) AS unemployed_members,

  COUNT(DISTINCT CASE
    WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Wage Work'
      OR fm.NATURE_WAGE_WORK IS NOT NULL
    THEN fm.FAMILY_MEMBER_ID
  END) AS daily_wage_workers,

  COUNT(DISTINCT CASE
    WHEN LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%driver%'
      OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%electric%'
      OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%mechanic%'
      OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%tailor%'
      OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%carpenter%'
      OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%computer%'
      OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%bank%'
      OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%shop%'
      OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%company worker%'
      OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%security%'
      OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%painter%'
      OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%civil%'
      OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%technician%'
    THEN fm.FAMILY_MEMBER_ID
  END) AS skilled_workers,

  COUNT(DISTINCT CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Self Employed - Farm based'    THEN fm.FAMILY_MEMBER_ID END) AS farm_based,
  COUNT(DISTINCT CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Self Employed-Agri allied'     THEN fm.FAMILY_MEMBER_ID END) AS agri_allied,
  COUNT(DISTINCT CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Self Employed- Non-farm based' THEN fm.FAMILY_MEMBER_ID END) AS non_farm,
  COUNT(DISTINCT CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Salaried Job'                  THEN fm.FAMILY_MEMBER_ID END) AS salaried,
  COUNT(DISTINCT CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Wage Work'                     THEN fm.FAMILY_MEMBER_ID END) AS wage_workers,
  COUNT(DISTINCT CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Housewife'                     THEN fm.FAMILY_MEMBER_ID END) AS housewife,
  COUNT(DISTINCT CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Studying'                      THEN fm.FAMILY_MEMBER_ID END) AS students,
  COUNT(DISTINCT CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Unemployed'                    THEN fm.FAMILY_MEMBER_ID END) AS unemployed,
  COUNT(DISTINCT CASE
    WHEN fm.OCCUPATION IS NULL
      OR TRIM(COALESCE(fm.OCCUPATION, '')) = ''
    THEN fm.FAMILY_MEMBER_ID
  END) AS other

FROM FAMILY_MEMBER fm
JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
WHERE __WHERE_CLAUSE__
```

#### Tables Used

- `FAMILY`
- `FAMILY_MEMBER` (columns: `OCCUPATION`, `NATURE_WAGE_WORK`)

#### Notes / Edge Cases

- Duplicate-safe via `COUNT(DISTINCT CASE … THEN fm.FAMILY_MEMBER_ID END)`. See §3.2.
- Same query also feeds §7.2 Occupation Distribution.
- See §13 (Important Business Logic) for exact `students` and `other` category definitions.

---

### 7.2 Occupation Distribution

#### Purpose
Horizontal bar chart showing member counts per occupation bucket.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `occupationSegments`, `occupationTotal`, `occupationBarWidth` |
| **API field** | `response.employment.occupation_distribution` |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchEmploymentSection` — same query as §7.1 |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query

Same SQL as **§7.1 Employment Insights**.

#### Tables Used

- `FAMILY`
- `FAMILY_MEMBER`

#### Notes

- Uses the same duplicate-safe query as §7.1.
- Buckets: `farm_based`, `agri_allied`, `non_farm`, `salaried`, `wage_workers`, `housewife`, `students`, `unemployed`, `other`.
- Sorted descending by value in frontend.

---

## 8. Agriculture Section

**Handler function:** `fetchAgricultureSection` (`handlers/dashboard_summary.go`)  
**Parallelism:** 7 goroutines — `totalFarmersQuery`, `noIrrigationQuery`, `landUtilQuery`, `invalidQuery`, `landDistributionQuery`, `cropQuery`, `kharifCountQuery` + `rabiCountQuery`

---

### 8.1 Agriculture Intelligence (Summary Cards)

#### Purpose
Four top-level summary numbers: Total Farmer Households, Farmers Without Irrigation, Kharif Active, Rabi Active.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Template** | `.agri-stats` — four summary numbers |
| **API fields** | `agriculture.totalFarmers`, `farmersWithoutIrrigation`, `kharifFarmers`, `rabiFarmers` |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchAgricultureSection` |
| **Query vars** | `totalFarmersQuery`, `noIrrigationQuery`, `kharifCountQuery`, `rabiCountQuery` |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Queries

**Total Farmer Households:**
```sql
SELECT COUNT(*) FROM FAMILY f
WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
  AND __WHERE_CLAUSE__
```

**Farmers Without Irrigation:**
```sql
SELECT COUNT(*) FROM FAMILY f
WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
  AND (
    f.SOURCE_WATER_IRRIGATION IS NULL
    OR f.SOURCE_WATER_IRRIGATION = ''
    OR f.SOURCE_WATER_IRRIGATION = 'None'
    OR f.SOURCE_WATER_IRRIGATION = 'Rain Fed'
  )
  AND __WHERE_CLAUSE__
```

**Kharif Active Farmers:**
```sql
SELECT COUNT(*) FROM FAMILY f
WHERE f.CULTIVATING_DURING_KHARIF_SEASON IS NOT NULL
  AND f.CULTIVATING_DURING_KHARIF_SEASON != ''
  AND f.CULTIVATING_DURING_KHARIF_SEASON != 'No'
  AND __WHERE_CLAUSE__
```

**Rabi Active Farmers:**
```sql
SELECT COUNT(*) FROM FAMILY f
WHERE f.TAKING_CROPS_RABI_SEASON IS NOT NULL
  AND f.TAKING_CROPS_RABI_SEASON != ''
  AND f.TAKING_CROPS_RABI_SEASON != 'No'
  AND __WHERE_CLAUSE__
```

#### Tables Used

- `FAMILY`

#### Notes

- Section also contains Land Holdings, Land Utilization, and Season-wise Crops (§8.2–8.4).
- See §13.1 (Important Business Logic) for the Farmer definition.

---

### 8.2 Land Holdings Distribution

#### Purpose
Bar chart showing count of farming families grouped by land area: Landless, Small, Medium, Large.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `landDistributionRows`, `landPct` |
| **API field** | `response.agriculture.landDistribution` (`label`, `count`) |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchAgricultureSection` |
| **Query var** | `landDistributionQuery` |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query

```sql
SELECT
  CASE
    WHEN CAST(TRIM(f.AREA_AGRICULTURE_LAND_ACRES) AS DECIMAL(10,2)) = 0
    THEN 'Landless'

    WHEN CAST(TRIM(f.AREA_AGRICULTURE_LAND_ACRES) AS DECIMAL(10,2)) > 0
      AND CAST(TRIM(f.AREA_AGRICULTURE_LAND_ACRES) AS DECIMAL(10,2)) <= 2.5
    THEN 'Small'

    WHEN CAST(TRIM(f.AREA_AGRICULTURE_LAND_ACRES) AS DECIMAL(10,2)) > 2.5
      AND CAST(TRIM(f.AREA_AGRICULTURE_LAND_ACRES) AS DECIMAL(10,2)) <= 10
    THEN 'Medium'

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
```

#### Tables Used

- `FAMILY` (columns: `OWN_AGRICULTURE_LAND`, `AREA_AGRICULTURE_LAND_ACRES`)

#### Notes

- Go merges categories via `canonicalLandDistributionLabel` → Landless, Small, Medium, Large.
- Frontend i18n maps display labels (`landCategoryLandless`, etc.).
- Bar width normalized to the max count in the section.

---

### 8.3 Land Utilization

#### Purpose
ApexCharts donut showing cultivated vs. unused land (in acres), with a footnote for invalid/missing records.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `landUtilizationRows`, `landUtilizationSeries`, `landUtilizationOptions`, `landUtilizationHasData` |
| **API field** | `response.agriculture.landUtilizationSummary` |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchAgricultureSection` |
| **Query vars** | `landUtilQuery`, `invalidQuery` |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query — Utilization Aggregates

```sql
SELECT
  COALESCE(ROUND(SUM(t.total_land), 2), 0)                     AS total_land,
  COALESCE(ROUND(SUM(t.cultivated_land), 2), 0)                AS cultivated_land,
  COALESCE(ROUND(SUM(t.total_land - t.cultivated_land), 2), 0) AS unused_land,
  COUNT(*) AS valid_records
FROM (
  SELECT
    CAST(f.AREA_AGRICULTURE_LAND_ACRES  AS DECIMAL(12,2)) AS total_land,
    CAST(f.LAND_UNDER_CULTIVATION_ACRES AS DECIMAL(12,2)) AS cultivated_land,
    f.DISTRICT_ID, f.TALUKA_ID, f.VILLAGE_ID
  FROM FAMILY f
  WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
    AND f.AREA_AGRICULTURE_LAND_ACRES IS NOT NULL
    AND f.LAND_UNDER_CULTIVATION_ACRES IS NOT NULL
    AND TRIM(f.AREA_AGRICULTURE_LAND_ACRES) <> ''
    AND TRIM(f.LAND_UNDER_CULTIVATION_ACRES) <> ''
    AND f.AREA_AGRICULTURE_LAND_ACRES     REGEXP '^[0-9]*\\.?[0-9]+$'
    AND f.LAND_UNDER_CULTIVATION_ACRES    REGEXP '^[0-9]*\\.?[0-9]+$'
    AND CAST(f.LAND_UNDER_CULTIVATION_ACRES AS DECIMAL(12,2))
        <= CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2))
    AND CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2)) BETWEEN 0 AND 500
    AND __WHERE_CLAUSE__
) t
```

#### SQL Query — Invalid Record Count (footnote)

```sql
SELECT COUNT(*)
FROM FAMILY f
WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
  AND __WHERE_CLAUSE__
  AND (
    f.AREA_AGRICULTURE_LAND_ACRES IS NULL
    OR f.LAND_UNDER_CULTIVATION_ACRES IS NULL
    OR TRIM(f.AREA_AGRICULTURE_LAND_ACRES) = ''
    OR TRIM(f.LAND_UNDER_CULTIVATION_ACRES) = ''
    OR f.AREA_AGRICULTURE_LAND_ACRES     NOT REGEXP '^[0-9]*\\.?[0-9]+$'
    OR f.LAND_UNDER_CULTIVATION_ACRES    NOT REGEXP '^[0-9]*\\.?[0-9]+$'
    OR CAST(f.LAND_UNDER_CULTIVATION_ACRES AS DECIMAL(12,2))
       > CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2))
    OR CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2)) > 500
  )
```

#### Tables Used

- `FAMILY`

#### Notes

- `cultivated_percent` / `unused_percent` are computed in Go, not SQL.
- UI footnote: valid vs. invalid survey record counts.
- Donut: cultivated vs. unused land in acres.

---

### 8.4 Season-wise Crops

#### Purpose
ApexCharts grouped bar showing top 5 crop/season combinations (Kharif + Rabi) by count.

#### Frontend

| Item | Value |
|------|-------|
| **File** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Computed** | `seasonCropCounts`, `seasonCropSeries`, `seasonCropOptions`, `seasonCropHasData` |
| **API field** | `response.agriculture.seasonCropRows` (`season`, `crop`, `count`) |

#### Backend

| Item | Value |
|------|-------|
| **Function** | `fetchAgricultureSection` |
| **Query var** | `cropQuery` (UNION Kharif + Rabi) |
| **File** | `handlers/dashboard_summary.go` |

#### SQL Query

```sql
SELECT season, crop, SUM(cnt) AS cnt
FROM (
  SELECT
    'Kharif' AS season,
    TRIM(f.CULTIVATING_DURING_KHARIF_SEASON) AS crop,
    COUNT(*) AS cnt
  FROM FAMILY f
  WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
    AND f.CULTIVATING_DURING_KHARIF_SEASON IS NOT NULL
    AND TRIM(f.CULTIVATING_DURING_KHARIF_SEASON) != ''
    AND __WHERE_CLAUSE__
  GROUP BY TRIM(f.CULTIVATING_DURING_KHARIF_SEASON)

  UNION ALL

  SELECT
    'Rabi' AS season,
    TRIM(f.CULTIVATING_DURING_RABI_SEASON) AS crop,
    COUNT(DISTINCT f.EXTERNAL_FAMILY_ID) AS cnt
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
```

#### Tables Used

- `FAMILY` (columns: `CULTIVATING_DURING_KHARIF_SEASON`, `CULTIVATING_DURING_RABI_SEASON`, `TAKING_CROPS_RABI_SEASON`)

#### Notes / Edge Cases

- Location filter args are bound **twice** — once per UNION branch.
- Top 5 crop/season pairs by count returned from DB; frontend may further aggregate by crop name.
- Kharif uses `COUNT(*)`; Rabi uses `COUNT(DISTINCT f.EXTERNAL_FAMILY_ID)`.

---

## 9. Location Filters

### Filter Flow

| Step | File / Function | Behavior |
|------|-----------------|----------|
| Load districts | `loadLocationOptions()` | `GET /api/location-options` (no params on initial load) |
| District change | `watch(selectedDistricts)` | Clears taluka/village; reloads options with `district_ids` |
| Taluka change | `watch(selectedTalukas)` | Clears village; reloads with `district_ids` + `taluka_ids` |
| Apply | `applyFilters()` → `buildLocationParams()` | `district_ids`, `taluka_ids`, `village_ids` comma-separated |
| Reset | `resetFilters()` | Clears all selections; reloads dashboard + options |

### API: `GET /api/location-options`

| Item | Value |
|------|-------|
| **Frontend** | `frontend/src/api/index.js` → `getLocationOptions(params)` |
| **Backend route** | `main.go` → `protected.GET("/location-options", locationHandler.GetLocationOptions)` |
| **Handler** | `handlers/location_options.go` → `GetLocationOptions` |

**Query parameters:** `district_id`, `district_ids`, `district_name`, `district_names`, `taluka_id`, `taluka_ids`, `village_id`, `village_ids`

**Response:**
```json
{
  "districts": [{ "id": "1", "name": "..." }],
  "talukas":   [{ "id": "1", "name": "...", "district_id": "1" }],
  "villages":  [{ "id": "1", "name": "...", "taluka_id": "1" }]
}
```

### SQL — Districts (always full list)

```sql
SELECT CAST(dm.pklDistrictId AS CHAR), COALESCE(dm.vsDisplayName, dm.vsDistrictName, '')
FROM district_master dm
WHERE dm.bEnabled = 1
ORDER BY COALESCE(dm.vsDisplayName, dm.vsDistrictName)
```

### SQL — Talukas (optional district filter)

```sql
SELECT CAST(tm.pklTalukaId AS CHAR),
       COALESCE(tm.vsDisplayName, tm.vsTalukaName, ''),
       CAST(tm.fklDistrictId AS CHAR)
FROM taluka_master tm
-- JOIN district_master dm ... when filtering by district name
WHERE tm.bEnabled = 1
  AND CAST(tm.fklDistrictId AS CHAR) IN (?)   -- when district_ids provided
ORDER BY COALESCE(tm.vsDisplayName, tm.vsTalukaName)
```

### SQL — Villages (optional district/taluka filter)

```sql
SELECT DISTINCT CAST(vm.pklVillageId AS CHAR),
       COALESCE(vm.vsDisplayName, vm.vsVillageName, ''),
       CAST(gm.fklTalukaId AS CHAR)
FROM village_master vm
JOIN grampanchayat_master gm ON gm.pklGramPanchayatId = vm.fklGramPanchayatId
JOIN taluka_master tm ON tm.pklTalukaId = gm.fklTalukaId
WHERE vm.bEnabled = 1
  AND CAST(tm.fklDistrictId AS CHAR) IN (?)   -- optional
  AND CAST(tm.pklTalukaId AS CHAR) IN (?)     -- optional
ORDER BY COALESCE(vm.vsDisplayName, vm.vsVillageName)
```

### Master Tables Reference

| Table | Role |
|-------|------|
| `district_master` | District dropdown |
| `taluka_master` | Taluka dropdown (`fklDistrictId`) |
| `grampanchayat_master` | Links village to taluka |
| `village_master` | Village dropdown |

### Notes

- Location options API does **not** affect dashboard metrics — it only populates filter UI.
- Dashboard metrics filtering uses `FAMILY.DISTRICT_ID`, `FAMILY.TALUKA_ID`, `FAMILY.VILLAGE_ID` (not master table joins in summary handler).

---

## 10. Parallel Processing

### Top-level Goroutines (`GetDashboardSummary`)

| Goroutine | Section Key | Handler Function |
|-----------|-------------|-----------------|
| 1 | `population` | `fetchPopulationSection` |
| 2 | `demographics` | `fetchDemographicsSection` |
| 3 | `education` | `fetchEducationSection` |
| 4 | `employment` | `fetchEmploymentSection` |
| 5 | `agriculture` | `fetchAgricultureSection` |

Synchronization: `sync.WaitGroup` with `wg.Add(5)`; results merged under mutex.

### Nested Parallelism

| Section | Sub-queries | Pattern |
|---------|-------------|---------|
| `fetchPopulationSection` | `populationQuery`, `householdsQuery`, `workingQuery`, `dependentQuery` | 4 goroutines; `bplHouseholdQuery` sequential after |
| `fetchDemographicsSection` | `genderQuery`, `ageQuery`, `divyangQuery`, `ageIncomeQuery`, `disabilityQuery` | 5 goroutines |
| `fetchAgricultureSection` | `totalFarmersQuery`, `noIrrigationQuery`, `landUtilQuery`, `invalidQuery`, `landDistributionQuery`, `cropQuery`, `kharifCountQuery`+`rabiCountQuery` | 7 goroutines |
| `fetchEducationSection` | 1 query | Sequential |
| `fetchEmploymentSection` | 1 query | Sequential |

---

## 11. Cache Mechanism

| Property | Detail |
|----------|--------|
| **Implementation** | `DashboardSummaryHandler.cache map[string]dashboardCacheItem` |
| **TTL constant** | `dashboardSummaryCacheTTL = 5 * time.Minute` |
| **Key format** | `dashboard_{districtKey}_{talukaKey}_{villageKey}` |
| **Key parts** | Each part is the query param value or `"all"` if empty |
| **Concurrency** | `cacheMux sync.RWMutex` — `RLock` for read, `Lock` for write |
| **On hit** | Return cached `gin.H` immediately; no DB queries |
| **On miss** | Run all sections; store result with `ExpiresAt = now + TTL` |
| **Invalidation** | None (time-based only; data edits do not clear cache) |
| **Multi-instance** | Each backend process has its own cache (not distributed) |
| **Stale fallback** | On section timeout/failure, returns last-known-good section data |
| **Stampede guard** | In-flight map (`inflight map[string]chan struct{}`) prevents duplicate DB calls for the same key |

### Index Bootstrap (one-time on handler init)

`ensureSummaryIndexes()` may create:

```sql
CREATE INDEX idx_family_member_location
ON FAMILY_MEMBER (DISTRICT_ID, TALUKA_ID, VILLAGE_ID)
```

Only if index is missing and columns exist — logged at startup.

---

## 12. Performance Notes

Observations only — no optimization recommendations implemented.

| Area | Observation |
|------|-------------|
| **Heavy queries** | `ageIncomeQuery` — subquery with per-family DOB aggregation + join to `FAMILY` |
| **Heavy queries** | `disabilityQuery` — `LIKE` patterns on `DISABILITY_CATEGORY`, `GROUP BY`, `LIMIT 8` |
| **Heavy queries** | `cropQuery` — `UNION ALL` of two grouped scans on `FAMILY` |
| **Heavy queries** | `landUtilQuery` / `landDistributionQuery` — `CAST`, `REGEXP`, `DECIMAL` on text acreage fields |
| **Aggregation-heavy** | `fetchEducationSection`, `fetchEmploymentSection` — full member scan with many `COUNT(DISTINCT CASE…)` |
| **Joins** | Almost all member metrics: `FAMILY_MEMBER` ⋈ `FAMILY` on `EXTERNAL_FAMILY_ID` |
| **Large scans** | Unfiltered dashboard (no location params) scans all families/members in scope |
| **Parallelism** | 5 section goroutines reduce latency but increase concurrent DB load |
| **Nested parallelism** | Population (4), Demographics (5), Agriculture (7) add more concurrent queries per request |
| **Timeout** | Per-section context timeout: population 5s, demographics 8s, agriculture 8s, education 12s, employment 12s |
| **Frontend timeout** | 30 seconds (`TIMEOUT_DATA`) |
| **Cache benefit** | Repeated identical filter combinations within 5 minutes avoid ~18+ SQL round-trips |
| **Bottleneck risk** | `STR_TO_DATE`, `TIMESTAMPDIFF` on `DOB` strings — not index-friendly |
| **Bottleneck risk** | `REGEXP` on land acreage columns — table scans for agriculture section |
| **Partial failure** | Section errors collected in `partial_errors`; other sections still returned |

### Approximate SQL Count per Cache Miss

| Section | Query Count |
|---------|-------------|
| Population | 4–5 |
| Demographics | 5 |
| Education | 1 |
| Employment | 1 |
| Agriculture | 7 |
| **Total** | **~18–19** |

---

## 13. Important Business Logic

### 13.1 Farmer Definition

A household is counted as a **farming household** if it owns agricultural land:

- Condition: `f.OWN_AGRICULTURE_LAND = 'Yes'`

Used by: Total Farmer Households, Farmers Without Irrigation, Land Holdings, Land Utilization, Season-wise Crops.

### 13.2 Students Logic (Employment)

**API field:** `employment.occupation_distribution.students`

- Condition: `TRIM(COALESCE(fm.OCCUPATION, '')) = 'Studying'`
- The DB value is `'Studying'` — **not** `'Student'`. Matching `OCCUPATION = 'Student'` returns **0** rows.
- Not to be confused with `education.students_count`, which uses `CURRENTLY_PURSUING_EDUCATION = 'YES'`.

### 13.3 Other Occupation Logic (Employment)

**API field:** `employment.occupation_distribution.other`

- Condition: `fm.OCCUPATION IS NULL OR TRIM(COALESCE(fm.OCCUPATION, '')) = ''`
- Counts **only** members with NULL or empty `OCCUPATION`.
- Does **not** use a `NOT IN` exclusion — any member with a non-empty occupation (including `'Not Applicable'`) is **not** in `other`.

### 13.4 BPL Logic

A household is BPL if any of the following apply:

```sql
UPPER(TRIM(COALESCE(f.FAMILY_BELONG_BPL_CATEGORY, ''))) = 'YES'
OR UPPER(TRIM(COALESCE(f.RATION_CARD_TYPE, ''))) IN ('BPL', 'AAY')
```

- Query runs only when `ColumnChecker` confirms at least one column exists.
- `non_bpl` = `total_households - bpl`, computed in Go.

### 13.5 Literacy Logic

**API field:** `education.literate_population`

- Condition: `UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'YES'`
- Literacy rate: `(literate_population / total_population) × 100` — computed in Go.

### 13.6 Dropout Logic

**API field:** `education.dropout_count`

- `FAMILY_MEMBER.DROP_OUT` stores the dropout class/standard (numeric values 1–10), **not** a YES/NO flag.
- A non-empty `DROP_OUT` value indicates the member dropped out of school.
- Condition: `fm.DROP_OUT IS NOT NULL AND TRIM(fm.DROP_OUT) != ''`
- No `EVER_ATTENDED_SCHOOL` or `CURRENTLY_PURSUING_EDUCATION` guards are applied — those could exclude valid dropout rows.

### 13.7 Qualification Not Available (`below_10th`) Logic

**API key:** `education.qualification_distribution.below_10th`  
**UI label:** Qualification Not Available

- Condition: `fm.QUALIFICATION IS NULL OR TRIM(fm.QUALIFICATION) = ''`
- The source database does **not** contain the value `'Below 10th'`. Stored values are only: `10th`, `12th`, `Graduation & Above`.
- Do **not** match `TRIM(fm.QUALIFICATION) = 'Below 10th'` — that string is absent from the DB.
- API key remains `below_10th`; UI label set via `agriDashboard.below10th` in `en.json` / `mr.json`.

---

## 14. File Mapping Summary

### Section → Code Mapping

| Dashboard Section | Frontend Computed | Backend Function | Query Variable |
|-------------------|------------------|-----------------|---------------|
| Total Households | `populationMetrics` | `fetchPopulationSection` | `householdsQuery` |
| Total Population | `populationMetrics` | `fetchPopulationSection` | `populationQuery` |
| BPL Status | `bplSegments` | `fetchPopulationSection` | `bplHouseholdQuery` |
| Gender Distribution | `genderSegments` | `fetchDemographicsSection` | `genderQuery` |
| Divyang Distribution | `divyangSegments` | `fetchDemographicsSection` | `divyangQuery`, `disabilityQuery` |
| Age-wise Family Income | `syncAgeIncomeGenderChart` | `fetchDemographicsSection` | `ageIncomeQuery` |
| Education Intelligence | `educationMetrics` | `fetchEducationSection` | single `query` |
| Qualification Distribution | `qualificationSegments` | `fetchEducationSection` | same query as Education |
| Literacy Rate | `literacyRateLabel` | `fetchEducationSection` | computed in Go |
| Employment Insights | `employmentMetrics` | `fetchEmploymentSection` | single `query` |
| Occupation Distribution | `occupationSegments` | `fetchEmploymentSection` | same query as Employment |
| Agriculture Cards | `agriculture.*` | `fetchAgricultureSection` | farmer/irrigation/kharif/rabi queries |
| Land Holdings | `landDistributionRows` | `fetchAgricultureSection` | `landDistributionQuery` |
| Land Utilization | ApexCharts donut | `fetchAgricultureSection` | `landUtilQuery`, `invalidQuery` |
| Season-wise Crops | ApexCharts bar | `fetchAgricultureSection` | `cropQuery` |
| Location filters (UI) | `loadLocationOptions` | `handlers/location_options.go` | `GetLocationOptions` (3 queries) |
| Working/dependent population | (stored, not displayed) | `fetchPopulationSection` | `workingQuery`, `dependentQuery` |
| Age distribution buckets | (stored, not displayed) | `fetchDemographicsSection` | `ageQuery` |

### Database Tables Reference

| Table | Used By |
|-------|---------|
| `FAMILY` | All sections — primary source for household-level data |
| `FAMILY_MEMBER` | Population, Demographics, Education, Employment — member-level data |
| `district_master` | Location options (district dropdown) |
| `taluka_master` | Location options (taluka dropdown) |
| `village_master` | Location options (village dropdown, via `grampanchayat_master`) |
| `grampanchayat_master` | Village → taluka join in location options |

### Source File Index

| File | Role |
|------|------|
| `frontend/src/views/agriculture/Dashboard.vue` | All dashboard UI, computed properties, chart config |
| `frontend/src/api/index.js` | HTTP wrappers — `getDashboardSummary`, `getLocationOptions` |
| `main.go` | Route registration (Gin router) |
| `handlers/dashboard_summary.go` | All dashboard SQL + caching + parallel execution |
| `handlers/location_options.go` | `GET /location-options`, `GET /districts` |

---

## Document Metadata

| Field | Value |
|-------|-------|
| **Generated for** | Digital Twin — Agriculture Village Command Center |
| **Source revision** | Reflects `handlers/dashboard_summary.go` and `Dashboard.vue` as of documentation date |
| **Maintainer note** | Update this document when dashboard SQL or API response shape changes |
