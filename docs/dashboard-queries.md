# Dashboard Query Documentation

Technical reference for all backend/API/database queries and data flow used by the **Village Command Center** dashboard (Agriculture module).

| Item | Value |
|------|-------|
| **UI page** | Village Command Center |
| **Frontend component** | `frontend/src/views/agriculture/Dashboard.vue` |
| **Route** | `/agriculture/dashboard` (via `AgricultureLayout.vue`) |
| **Primary data API** | `GET /api/dashboard/summary` |
| **Filter options API** | `GET /api/location-options` |

---

## 1. Dashboard Architecture Overview

### End-to-end flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Browser: Dashboard.vue (Village Command Center)                            │
├─────────────────────────────────────────────────────────────────────────────┤
│  onMounted / Apply / Reset                                                  │
│    ├─ loadLocationOptions()  →  GET /api/location-options                   │
│    └─ fetchDashboardData()   →  GET /api/dashboard/summary                  │
│           │                                                                 │
│           ▼                                                                 │
│  Parse JSON: { population, demographics, education, employment,           │
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
│    fetchEducationSection     → 1 SQL query                                    │
│    fetchEmploymentSection    → 1 SQL query                                    │
│    fetchAgricultureSection   → 7 SQL queries (parallel inside section)      │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  MySQL tables: FAMILY, FAMILY_MEMBER, district_master, taluka_master,       │
│                village_master, grampanchayat_master                           │
│  No materialized views. No stored procedures for dashboard.                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Frontend (`Dashboard.vue`) flow

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

### API layer flow

- Base URL: `/api` (`frontend/src/api/index.js`, `const BASE = '/api'`).
- Auth: `Authorization: Bearer <token>` when present.
- Dashboard timeout: `TIMEOUT_DATA` = 30 seconds.
- `cache: 'no-store'` on fetch (browser does not cache HTTP responses).

### Backend handler flow

1. **`GetDashboardSummary`** reads query params (singular and plural forms).
2. **Cache check** — in-memory map keyed by location filter string.
3. On miss: build `whereF, args` via `buildOptionalLocationFilterWithArrays("f", ...)`.
4. Spawn **5 goroutines** (population, demographics, education, employment, agriculture).
5. Merge results; attach `partial_errors` if any section failed.
6. Store in cache (5 min TTL); return JSON.

### Database flow

- All dashboard metrics are **live aggregations** on `FAMILY` / `FAMILY_MEMBER`.
- Location filters append `AND f.DISTRICT_ID IN (...)`, `f.TALUKA_ID IN (...)`, `f.VILLAGE_ID IN (...)` (alias `f` on `FAMILY`).
- Member-level metrics join: `FAMILY_MEMBER fm JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID`.
- Placeholder `__WHERE_CLAUSE__` in SQL templates is replaced by the built filter clause via `injectWhere()`.

### Cache usage

- **Type:** In-process `map[string]dashboardCacheItem` on `DashboardSummaryHandler`.
- **TTL:** `dashboardSummaryCacheTTL` = 5 minutes.
- **Scope:** Per server process (not shared across instances).
- **Invalidation:** Time-based only (no explicit invalidation on data change).

### Parallel goroutines (top level)

| Goroutine | Handler function | Internal parallelism |
|-----------|------------------|--------------------|
| 1 | `fetchPopulationSection` | 4 goroutines + optional BPL query |
| 2 | `fetchDemographicsSection` | 5 goroutines |
| 3 | `fetchEducationSection` | Single query |
| 4 | `fetchEmploymentSection` | Single query |
| 5 | `fetchAgricultureSection` | 7 goroutines |

### Filter flow

1. User selects districts → talukas reload via `watch(selectedDistricts)` → `loadLocationOptions({ district_ids })`.
2. User selects talukas → villages reload via `watch(selectedTalukas)` → `loadLocationOptions({ district_ids, taluka_ids })`.
3. **Apply** sends comma-separated IDs: `district_ids=1,2`, `taluka_ids=3`, `village_ids=4,5`.
4. Backend `buildOptionalLocationFilterWithArrays` produces `IN (...)` clauses on `f.DISTRICT_ID`, `f.TALUKA_ID`, `f.VILLAGE_ID`.

### APIs not used by this dashboard

The following exist elsewhere in the project but are **not** called from `Dashboard.vue`:

- `GET /api/insights/agriculture`
- `GET /api/population/dashboard`
- `GET /api/population/demographics`
- `GET /api/population/education`
- `GET /api/population/employment`

---

## 2. Main Dashboard API

### API

```
GET /api/dashboard/summary
```

**Authentication:** Protected route (Bearer token required).

### Request parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `district_id` | string | Single district ID (legacy) |
| `district_ids` | string | Comma-separated district IDs (used by Dashboard.vue) |
| `taluka_id` | string | Single taluka ID |
| `taluka_ids` | string | Comma-separated taluka IDs |
| `village_id` | string | Single village ID |
| `village_ids` | string | Comma-separated village IDs |

Empty or invalid values (`0`, `null`, `undefined`) are ignored.

### Response structure

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

`partial_errors` is present only when one or more sections fail; other sections still return data or defaults.

### Frontend

| Item | Value |
|------|-------|
| **file** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function** | `fetchDashboardData(params)` |
| **API wrapper** | `frontend/src/api/index.js` → `getDashboardSummary(params)` |

### Backend

| Item | Value |
|------|-------|
| **route** | `main.go` → `protected.GET("/dashboard/summary", dashboardSummaryHandler.GetDashboardSummary)` |
| **handler** | `handlers/dashboard_summary.go` → `(*DashboardSummaryHandler).GetDashboardSummary` |
| **service/repository** | Same file — section fetchers: `fetchPopulationSection`, `fetchDemographicsSection`, `fetchEducationSection`, `fetchEmploymentSection`, `fetchAgricultureSection` |

### Cache

| Item | Value |
|------|-------|
| **cache key format** | `dashboard_{districtKey}_{talukaKey}_{villageKey}` — each key is param value or `"all"` |
| **TTL** | 5 minutes (`dashboardSummaryCacheTTL`) |
| **storage** | `DashboardSummaryHandler.cache map[string]dashboardCacheItem` |
| **concurrency** | `sync.RWMutex` (`cacheMux`) |

---

## 3. Dashboard Sections

Location filter placeholder used in all section queries below:

```sql
-- Injected as __WHERE_CLAUSE__ (example with filters):
-- 1=1 AND f.DISTRICT_ID IN (?) AND f.TALUKA_ID IN (?) AND f.VILLAGE_ID IN (?)
```

---

## Total Households

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `populationMetrics` (metric card) |
| **API field** | `response.population.total_households` |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchPopulationSection` |
| **query** | `householdsQuery` |
| **file** | `handlers/dashboard_summary.go` |

### Database

```sql
SELECT COUNT(*)
FROM FAMILY f
WHERE __WHERE_CLAUSE__
```

### Tables used

- `FAMILY`

### Notes

- Displayed as top summary card with home icon.
- `working_population` and `dependent_population` are fetched in the same section but **not shown** in current UI.

---

## Total Population

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `populationMetrics` |
| **API field** | `response.population.total_population` |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchPopulationSection` |
| **query** | `populationQuery` |
| **file** | `handlers/dashboard_summary.go` |

### Database

```sql
SELECT COUNT(*)
FROM FAMILY_MEMBER fm
JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
WHERE __WHERE_CLAUSE__
```

### Tables used

- `FAMILY`
- `FAMILY_MEMBER`

### Notes

- Counts all family members in scope (not distinct families).

---

## Gender Distribution

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `genderSegments`, `genderTotal`, `genderPieStyle` |
| **API field** | `response.demographics.gender_distribution` (`male`, `female`, `other`) |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchDemographicsSection` |
| **query** | `genderQuery` |
| **file** | `handlers/dashboard_summary.go` |

### Database

```sql
SELECT
  COALESCE(SUM(CASE WHEN LOWER(TRIM(fm.GENDER)) = 'male' THEN 1 ELSE 0 END), 0) AS male,
  COALESCE(SUM(CASE WHEN LOWER(TRIM(fm.GENDER)) = 'female' THEN 1 ELSE 0 END), 0) AS female,
  COALESCE(SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER, ''))) NOT IN ('male', 'female') THEN 1 ELSE 0 END), 0) AS other
FROM FAMILY_MEMBER fm
JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
WHERE __WHERE_CLAUSE__
```

### Tables used

- `FAMILY`
- `FAMILY_MEMBER`

### Notes

- Rendered as CSS `conic-gradient` donut (not Chart.js).
- Total in header = sum of male + female + other.

---

## Divyang Distribution

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `divyangSegments`, `divyangTotal`, `divyangPieStyle`, `mapDisabilityGroup` |
| **API fields** | `response.demographics.disability_distribution`, `response.demographics.total_divyang` |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchDemographicsSection` |
| **queries** | `divyangQuery` (total), `disabilityQuery` (breakdown) |
| **file** | `handlers/dashboard_summary.go` |

### Database — total count

```sql
SELECT COUNT(*)
FROM FAMILY_MEMBER fm
JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
WHERE UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES'
  AND __WHERE_CLAUSE__
```

### Database — distribution (top 8 groups)

```sql
SELECT
  CASE
    WHEN DISABILITY_CATEGORY LIKE '%Blind%' OR DISABILITY_CATEGORY LIKE '%Low vision%' THEN 'Visual Disability'
    WHEN DISABILITY_CATEGORY LIKE '%Locomotor%' OR DISABILITY_CATEGORY LIKE '%Cerebral%'
      OR DISABILITY_CATEGORY LIKE '%Muscular%' OR DISABILITY_CATEGORY LIKE '%Dwarf%' THEN 'Locomotor Disability'
    WHEN DISABILITY_CATEGORY LIKE '%Mental%' OR DISABILITY_CATEGORY LIKE '%Autism%'
      OR DISABILITY_CATEGORY LIKE '%Intellectual%' OR DISABILITY_CATEGORY LIKE '%Learning%' THEN 'Intellectual Disability'
    WHEN DISABILITY_CATEGORY LIKE '%Hearing%' THEN 'Hearing Disability'
    WHEN DISABILITY_CATEGORY LIKE '%Speech%' THEN 'Speech Disability'
    WHEN DISABILITY_CATEGORY LIKE '%Multiple%' THEN 'Multiple Disabilities'
    WHEN DISABILITY_CATEGORY LIKE '%Parkinson%' OR DISABILITY_CATEGORY LIKE '%Sclerosis%'
      OR DISABILITY_CATEGORY LIKE '%Sickle%' OR DISABILITY_CATEGORY LIKE '%Thalassemia%' THEN 'Chronic Conditions'
    ELSE 'Other'
  END AS disability_group,
  COUNT(*) AS total
FROM FAMILY_MEMBER fm
JOIN FAMILY f ON fm.EXTERNAL_FAMILY_ID = f.EXTERNAL_FAMILY_ID
WHERE UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES'
  AND __WHERE_CLAUSE__
GROUP BY disability_group
ORDER BY total DESC
LIMIT 8
```

### Tables used

- `FAMILY`
- `FAMILY_MEMBER` (columns: `DIVYANG`, `DISABILITY_CATEGORY`)

### Notes

- Frontend maps API labels to i18n keys via `DISABILITY_AGGREGATE_LABELS` / `mapDisabilityGroup`.
- `divyangTotal` prefers `total_divyang` from API; else sums segment values.
- Unmatched disability names shown as raw API `name`.

---

## BPL Status

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `bplSegments`, `bplTotal`, `bplPieStyle` |
| **API field** | `payload.population.bpl_distribution` (primary; fallback from `demographics.bpl_distribution` in `fetchDashboardData`) |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchPopulationSection` |
| **query** | `bplHouseholdQuery` (conditional) |
| **file** | `handlers/dashboard_summary.go` |

### Database

Only runs if `ColumnChecker` reports `FAMILY_BELONG_BPL_CATEGORY` and/or `RATION_CARD_TYPE` exist:

```sql
SELECT COUNT(*)
FROM FAMILY f
WHERE __WHERE_CLAUSE__
  AND (
    UPPER(TRIM(COALESCE(f.FAMILY_BELONG_BPL_CATEGORY, ''))) = 'YES'
    OR UPPER(TRIM(COALESCE(f.RATION_CARD_TYPE, ''))) IN ('BPL', 'AAY')
  )
```

### Tables used

- `FAMILY`

### Notes

- `non_bpl` = `total_households - bpl` (computed in Go, not SQL).
- `bpl_distribution` is stored under `population` in API response (not `demographics`).
- Donut shows BPL vs Non-BPL families.

---

## Age-wise Family Income Distribution

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `applyDemographicsData`, `ageIncomeGenderSegments`, `syncAgeIncomeGenderChart` (Chart.js bar) |
| **API field** | `response.demographics.age_income_gender_distribution` |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchDemographicsSection` |
| **query** | `ageIncomeQuery` |
| **file** | `handlers/dashboard_summary.go` |

### Database

```sql
SELECT
  age_group,
  COUNT(*) AS families,
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
```

### Tables used

- `FAMILY` (`ANNUAL_INCOME`)
- `FAMILY_MEMBER` (`DOB`, `RELATION_FAMILY_HEAD`)

### Notes

- Backend normalizes to age groups: `18-30`, `31-45`, `46-60`, `60+`.
- Chart: bar height = `families`; data label = average income (₹).
- Head-of-household DOB preferred for age; else earliest member DOB.
- `age_distribution` (0–5, 6–18, etc.) is queried separately but **not displayed** in current UI.

---

## Education Intelligence

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `educationMetrics` (5 mini-stat cards) |
| **API fields** | `education.literate_population`, `illiterate_population`, `students_count`, `dropout_count`, `graduate_population` |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchEducationSection` |
| **query** | Single aggregated `query` |
| **file** | `handlers/dashboard_summary.go` |

### Database

```sql
SELECT
  COUNT(*) AS total_population,
  SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'YES' THEN 1 ELSE 0 END) AS literate_population,
  SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'NO' OR fm.EVER_ATTENDED_SCHOOL IS NULL THEN 1 ELSE 0 END) AS illiterate_population,
  SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.CURRENTLY_PURSUING_EDUCATION, ''))) = 'YES' THEN 1 ELSE 0 END) AS students_count,
  SUM(CASE WHEN UPPER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'YES'
    AND UPPER(TRIM(COALESCE(fm.CURRENTLY_PURSUING_EDUCATION, ''))) != 'YES'
    AND fm.DROP_OUT IS NOT NULL AND TRIM(fm.DROP_OUT) != '' THEN 1 ELSE 0 END) AS dropout_count,
  SUM(CASE WHEN TRIM(COALESCE(fm.QUALIFICATION, '')) = 'Graduation & Above' THEN 1 ELSE 0 END) AS graduate_population,
  SUM(CASE WHEN fm.QUALIFICATION IS NULL OR TRIM(fm.QUALIFICATION) = '' THEN 1 ELSE 0 END) AS below_10th,
  SUM(CASE WHEN TRIM(fm.QUALIFICATION) = '10th' THEN 1 ELSE 0 END) AS tenth,
  SUM(CASE WHEN TRIM(fm.QUALIFICATION) = '12th' THEN 1 ELSE 0 END) AS twelfth,
  SUM(CASE WHEN TRIM(fm.QUALIFICATION) = 'Graduation & Above' THEN 1 ELSE 0 END) AS graduate_above
FROM FAMILY_MEMBER fm
JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
WHERE __WHERE_CLAUSE__
```

### Tables used

- `FAMILY`
- `FAMILY_MEMBER` (`EVER_ATTENDED_SCHOOL`, `CURRENTLY_PURSUING_EDUCATION`, `DROP_OUT`, `QUALIFICATION`)

### Notes

- One query powers both Education Intelligence cards and Qualification Distribution.

---

## Qualification Distribution

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `qualificationSegments`, `qualificationTotal`, `qualificationBarWidth` |
| **API field** | `response.education.qualification_distribution` |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchEducationSection` (same query as Education Intelligence) |
| **file** | `handlers/dashboard_summary.go` |

### Database

Same SQL as **Education Intelligence** — fields `below_10th`, `tenth`, `twelfth`, `graduate_above`.

### Tables used

- `FAMILY`
- `FAMILY_MEMBER`

### Notes

- Horizontal bar chart; sorted descending by count in frontend.
- Labels: Below 10th, 10th, 12th, Graduation & Above.

---

## Literacy Rate

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `literacyRateLabel` |
| **API field** | `response.education.literacy_rate` |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchEducationSection` |
| **computation** | Go: `(literate_population / total_population) * 100` when `total_population > 0` |
| **file** | `handlers/dashboard_summary.go` |

### Database

No separate SQL — derived from Education Intelligence query aggregates.

### Tables used

- (derived from same query as Education Intelligence)

### Notes

- Footer text also shows literate count vs `population.total_population` from population section.

---

## Employment Insights

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `employmentMetrics` (4 mini-stat cards) |
| **API fields** | `employment.employed_members`, `unemployed_members`, `daily_wage_workers`, `skilled_workers` |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchEmploymentSection` |
| **query** | Single aggregated `query` |
| **file** | `handlers/dashboard_summary.go` |

### Database

```sql
SELECT
  SUM(CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) IN (
    'Salaried Job','Self Employed - Farm based','Self Employed- Non-farm based',
    'Self Employed-Agri allied','Wage Work'
  ) THEN 1 ELSE 0 END) AS employed_members,
  SUM(CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) IN ('Unemployed','Not Applicable') THEN 1 ELSE 0 END) AS unemployed_members,
  SUM(CASE WHEN TRIM(COALESCE(fm.OCCUPATION, '')) = 'Wage Work' OR fm.NATURE_WAGE_WORK IS NOT NULL THEN 1 ELSE 0 END) AS daily_wage_workers,
  SUM(CASE WHEN LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%driver%'
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
    OR LOWER(COALESCE(fm.NATURE_WAGE_WORK, '')) LIKE '%technician%' THEN 1 ELSE 0 END) AS skilled_workers,
  -- occupation_distribution buckets: farm_based, agri_allied, non_farm, salaried,
  -- wage_workers, housewife, students, unemployed, other
  ...
FROM FAMILY_MEMBER fm
JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
WHERE __WHERE_CLAUSE__
```

(Full SELECT includes all `occupation_distribution` SUM cases — see `handlers/dashboard_summary.go` lines 886–904.)

### Tables used

- `FAMILY`
- `FAMILY_MEMBER` (`OCCUPATION`, `NATURE_WAGE_WORK`)

### Notes

- Same query also feeds Occupation Distribution.

---

## Occupation Distribution

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `occupationSegments`, `occupationTotal`, `occupationBarWidth` |
| **API field** | `response.employment.occupation_distribution` |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchEmploymentSection` |
| **file** | `handlers/dashboard_summary.go` |

### Database

Same single query as **Employment Insights**.

### Tables used

- `FAMILY`
- `FAMILY_MEMBER`

### Notes

- Buckets: farm_based, agri_allied, non_farm, salaried, wage_workers, housewife, students, unemployed, other.
- Sorted descending by value in frontend.

---

## Agriculture Intelligence

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **template** | `.agri-stats` — four summary numbers |
| **API fields** | `agriculture.totalFarmers`, `farmersWithoutIrrigation`, `kharifFarmers`, `rabiFarmers` |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchAgricultureSection` |
| **queries** | `totalFarmersQuery`, `noIrrigationQuery`, `kharifCountQuery`, `rabiCountQuery` |
| **file** | `handlers/dashboard_summary.go` |

### Database

**Total farmers:**
```sql
SELECT COUNT(*) FROM FAMILY f
WHERE f.OWN_AGRICULTURE_LAND = 'Yes' AND __WHERE_CLAUSE__
```

**No irrigation:**
```sql
SELECT COUNT(*) FROM FAMILY f
WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
  AND (f.SOURCE_WATER_IRRIGATION IS NULL OR f.SOURCE_WATER_IRRIGATION = ''
       OR f.SOURCE_WATER_IRRIGATION = 'None' OR f.SOURCE_WATER_IRRIGATION = 'Rain Fed')
  AND __WHERE_CLAUSE__
```

**Kharif active:**
```sql
SELECT COUNT(*) FROM FAMILY f
WHERE f.CULTIVATING_DURING_KHARIF_SEASON IS NOT NULL
  AND f.CULTIVATING_DURING_KHARIF_SEASON != ''
  AND f.CULTIVATING_DURING_KHARIF_SEASON != 'No'
  AND __WHERE_CLAUSE__
```

**Rabi active:**
```sql
SELECT COUNT(*) FROM FAMILY f
WHERE f.TAKING_CROPS_RABI_SEASON IS NOT NULL
  AND f.TAKING_CROPS_RABI_SEASON != ''
  AND f.TAKING_CROPS_RABI_SEASON != 'No'
  AND __WHERE_CLAUSE__
```

### Tables used

- `FAMILY`

### Notes

- Section also contains Land Holdings, Land Utilization, and Season-wise Crops (below).

---

## Land Holdings Distribution

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `landDistributionRows`, `landPct` |
| **API field** | `response.agriculture.landDistribution` (`label`, `count`) |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchAgricultureSection` |
| **query** | `landDistributionQuery` |
| **file** | `handlers/dashboard_summary.go` |

### Database

```sql
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
```

### Tables used

- `FAMILY` (`OWN_AGRICULTURE_LAND`, `AREA_AGRICULTURE_LAND_ACRES`)

### Notes

- Go merges categories via `canonicalLandDistributionLabel` → Landless, Small, Medium, Large.
- Frontend i18n maps display labels (`landCategoryLandless`, etc.).
- Bar width normalized to max count in section.

---

## Land Utilization

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `landUtilizationRows`, `landUtilizationSeries`, `landUtilizationOptions`, `landUtilizationHasData` (ApexCharts donut) |
| **API field** | `response.agriculture.landUtilizationSummary` |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchAgricultureSection` |
| **queries** | `landUtilQuery`, `invalidQuery` |
| **file** | `handlers/dashboard_summary.go` |

### Database — utilization aggregates

```sql
SELECT
  COALESCE(ROUND(SUM(t.total_land), 2), 0) AS total_land,
  COALESCE(ROUND(SUM(t.cultivated_land), 2), 0) AS cultivated_land,
  COALESCE(ROUND(SUM(t.total_land - t.cultivated_land), 2), 0) AS unused_land,
  COUNT(*) AS valid_records
FROM (
  SELECT
    CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2)) AS total_land,
    CAST(f.LAND_UNDER_CULTIVATION_ACRES AS DECIMAL(12,2)) AS cultivated_land,
    f.DISTRICT_ID, f.TALUKA_ID, f.VILLAGE_ID
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
```

### Database — invalid record count (footnote)

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
    OR f.AREA_AGRICULTURE_LAND_ACRES NOT REGEXP '^[0-9]*\\.?[0-9]+$'
    OR f.LAND_UNDER_CULTIVATION_ACRES NOT REGEXP '^[0-9]*\\.?[0-9]+$'
    OR CAST(f.LAND_UNDER_CULTIVATION_ACRES AS DECIMAL(12,2)) > CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2))
    OR CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2)) > 500
  )
```

### Tables used

- `FAMILY`

### Notes

- `cultivated_percent` / `unused_percent` computed in Go.
- UI footnote: valid vs invalid survey record counts.
- Donut: cultivated vs unused land (acres).

---

## Season-wise Crops

### Frontend

| Item | Value |
|------|-------|
| **file path** | `frontend/src/views/agriculture/Dashboard.vue` |
| **function/computed** | `seasonCropCounts`, `seasonCropSeries`, `seasonCropOptions`, `seasonCropHasData` (ApexCharts grouped bar) |
| **API field** | `response.agriculture.seasonCropRows` (`season`, `crop`, `count`) |

### Backend

| Item | Value |
|------|-------|
| **handler function** | `fetchAgricultureSection` |
| **query** | `cropQuery` (UNION Kharif + Rabi) |
| **file** | `handlers/dashboard_summary.go` |

### Database

```sql
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

  SELECT 'Rabi' AS season, TRIM(f.CULTIVATING_DURING_RABI_SEASON) AS crop,
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

### Tables used

- `FAMILY` (`CULTIVATING_DURING_KHARIF_SEASON`, `CULTIVATING_DURING_RABI_SEASON`, `TAKING_CROPS_RABI_SEASON`)

### Notes

- Location filter args bound **twice** (one set per UNION branch).
- Top 5 crop/season pairs by count returned from DB; frontend may further aggregate by crop name.
- Kharif counts rows; Rabi counts distinct `EXTERNAL_FAMILY_ID`.

---

## 4. Location Filters

### Frontend flow

| Step | File / function | Behavior |
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
  "talukas": [{ "id": "1", "name": "...", "district_id": "1" }],
  "villages": [{ "id": "1", "name": "...", "taluka_id": "1" }]
}
```

### SQL — districts (always all enabled)

```sql
SELECT CAST(dm.pklDistrictId AS CHAR), COALESCE(dm.vsDisplayName, dm.vsDistrictName, '')
FROM district_master dm
WHERE dm.bEnabled = 1
ORDER BY COALESCE(dm.vsDisplayName, dm.vsDistrictName)
```

### SQL — talukas (optional district filter)

```sql
SELECT CAST(tm.pklTalukaId AS CHAR), COALESCE(tm.vsDisplayName, tm.vsTalukaName, ''),
       CAST(tm.fklDistrictId AS CHAR)
FROM taluka_master tm
-- JOIN district_master dm ... when filtering by district name
WHERE tm.bEnabled = 1
  AND CAST(tm.fklDistrictId AS CHAR) IN (?)  -- when district_ids provided
ORDER BY COALESCE(tm.vsDisplayName, tm.vsTalukaName)
```

### SQL — villages (optional district/taluka filter)

```sql
SELECT DISTINCT CAST(vm.pklVillageId AS CHAR), COALESCE(vm.vsDisplayName, vm.vsVillageName, ''),
       CAST(gm.fklTalukaId AS CHAR)
FROM village_master vm
JOIN grampanchayat_master gm ON gm.pklGramPanchayatId = vm.fklGramPanchayatId
JOIN taluka_master tm ON tm.pklTalukaId = gm.fklTalukaId
WHERE vm.bEnabled = 1
  AND CAST(tm.fklDistrictId AS CHAR) IN (?)   -- optional
  AND CAST(tm.pklTalukaId AS CHAR) IN (?)    -- optional
ORDER BY COALESCE(vm.vsDisplayName, vm.vsVillageName)
```

### Master tables

| Table | Role |
|-------|------|
| `district_master` | District dropdown |
| `taluka_master` | Taluka dropdown (`fklDistrictId`) |
| `grampanchayat_master` | Links village to taluka |
| `village_master` | Village dropdown |

### Notes

- Location options API does **not** affect dashboard metrics directly; it only populates filter UI.
- Dashboard metrics filtering uses `FAMILY.DISTRICT_ID`, `FAMILY.TALUKA_ID`, `FAMILY.VILLAGE_ID` (not master table joins in summary handler).

---

## 5. Parallel Processing

### Top-level (`GetDashboardSummary`)

| Goroutine # | Section key | Function |
|-------------|-------------|----------|
| 1 | `population` | `fetchPopulationSection` |
| 2 | `demographics` | `fetchDemographicsSection` |
| 3 | `education` | `fetchEducationSection` |
| 4 | `employment` | `fetchEmploymentSection` |
| 5 | `agriculture` | `fetchAgricultureSection` |

Synchronization: `sync.WaitGroup` with `wg.Add(5)`; results merged under mutex.

### Inside `fetchPopulationSection` (4 parallel + BPL)

- `populationQuery`, `householdsQuery`, `workingQuery`, `dependentQuery` (4 goroutines)
- `bplHouseholdQuery` (sequential after wg, if columns exist)

### Inside `fetchDemographicsSection` (5 parallel)

- `genderQuery`, `ageQuery`, `divyangQuery`, `ageIncomeQuery`, `disabilityQuery`

### Inside `fetchAgricultureSection` (7 parallel)

- `totalFarmersQuery`, `noIrrigationQuery`, `landUtilQuery`, `invalidQuery`, `landDistributionQuery`, `cropQuery`, `kharifCountQuery` + `rabiCountQuery` (same goroutine)

### Sequential sections

- `fetchEducationSection` — 1 query
- `fetchEmploymentSection` — 1 query

---

## 6. Cache Mechanism

| Property | Detail |
|----------|--------|
| **Implementation** | `DashboardSummaryHandler.cache map[string]dashboardCacheItem` |
| **TTL constant** | `dashboardSummaryCacheTTL = 5 * time.Minute` |
| **Key format** | `dashboard_{districtKey}_{talukaKey}_{villageKey}` |
| **Key parts** | Each part is query param value or `"all"` if empty |
| **Concurrency** | `cacheMux sync.RWMutex` — RLock for read, Lock for write |
| **On hit** | Return cached `gin.H` immediately; no DB queries |
| **On miss** | Run all sections; store result with `ExpiresAt = now + TTL` |
| **Invalidation** | None (except expiry). Data edits do not clear cache. |
| **Multi-instance** | Each backend process has its own cache (not distributed) |

### Index bootstrap (one-time)

On handler init, `ensureSummaryIndexes()` may create:

```sql
CREATE INDEX idx_family_member_location ON FAMILY_MEMBER (DISTRICT_ID, TALUKA_ID, VILLAGE_ID)
```

(if index missing and columns exist — logged in startup).

---

## 7. Performance Notes

Observations only — no optimization recommendations implemented.

| Area | Observation |
|------|-------------|
| **Heavy queries** | `ageIncomeQuery` — subquery with per-family DOB aggregation + join to `FAMILY` |
| **Heavy queries** | `disabilityQuery` — `LIKE` patterns on `DISABILITY_CATEGORY`, `GROUP BY`, `LIMIT 8` |
| **Heavy queries** | `cropQuery` — `UNION ALL` of two grouped scans on `FAMILY` |
| **Heavy queries** | `landUtilQuery` / `landDistributionQuery` — `CAST`, `REGEXP`, `DECIMAL` on text acreage fields |
| **Aggregation-heavy** | `fetchEducationSection`, `fetchEmploymentSection` — full member scan with many `SUM(CASE...)` |
| **Joins** | Almost all member metrics: `FAMILY_MEMBER` ⋈ `FAMILY` on `EXTERNAL_FAMILY_ID` |
| **Large scans** | Unfiltered dashboard (no location params) scans all families/members in scope |
| **Parallelism** | 5 section goroutines reduce latency but increase concurrent DB load |
| **Nested parallelism** | Population (4), Demographics (5), Agriculture (7) add more concurrent queries per request |
| **Timeout** | Request context timeout: 30 seconds (`context.WithTimeout`) |
| **Frontend timeout** | 30 seconds (`TIMEOUT_DATA`) |
| **Cache benefit** | Repeated identical filter combinations within 5 minutes avoid ~18+ SQL round-trips |
| **Bottleneck risk** | `STR_TO_DATE`, `TIMESTAMPDIFF` on `DOB` strings — not index-friendly |
| **Bottleneck risk** | `REGEXP` on land acreage columns — table scans for agriculture section |
| **Partial failure** | Section errors collected in `partial_errors`; other sections still returned |

### Approximate SQL count per cache miss

| Section | Query count |
|---------|-------------|
| population | 4–5 |
| demographics | 5 |
| education | 1 |
| employment | 1 |
| agriculture | 7 |
| **Total** | **~18–19** |

---

## 8. File Mapping Summary

| Dashboard Section | Frontend File | Backend File | SQL Location |
|-------------------|---------------|--------------|--------------|
| All sections (data) | `frontend/src/views/agriculture/Dashboard.vue` | `handlers/dashboard_summary.go` | `GetDashboardSummary` + `fetch*Section` |
| API client | `frontend/src/api/index.js` | — | — |
| Route registration | — | `main.go` | — |
| Total Households | `Dashboard.vue` → `populationMetrics` | `dashboard_summary.go` | `fetchPopulationSection` → `householdsQuery` |
| Total Population | `Dashboard.vue` → `populationMetrics` | `dashboard_summary.go` | `fetchPopulationSection` → `populationQuery` |
| Gender Distribution | `Dashboard.vue` → `genderSegments` | `dashboard_summary.go` | `fetchDemographicsSection` → `genderQuery` |
| Divyang Distribution | `Dashboard.vue` → `divyangSegments` | `dashboard_summary.go` | `fetchDemographicsSection` → `divyangQuery`, `disabilityQuery` |
| BPL Status | `Dashboard.vue` → `bplSegments` | `dashboard_summary.go` | `fetchPopulationSection` → `bplHouseholdQuery` |
| Age-wise Family Income | `Dashboard.vue` → `syncAgeIncomeGenderChart` | `dashboard_summary.go` | `fetchDemographicsSection` → `ageIncomeQuery` |
| Education Intelligence | `Dashboard.vue` → `educationMetrics` | `dashboard_summary.go` | `fetchEducationSection` |
| Qualification Distribution | `Dashboard.vue` → `qualificationSegments` | `dashboard_summary.go` | `fetchEducationSection` (same query) |
| Literacy Rate | `Dashboard.vue` → `literacyRateLabel` | `dashboard_summary.go` | `fetchEducationSection` (computed in Go) |
| Employment Insights | `Dashboard.vue` → `employmentMetrics` | `dashboard_summary.go` | `fetchEmploymentSection` |
| Occupation Distribution | `Dashboard.vue` → `occupationSegments` | `dashboard_summary.go` | `fetchEmploymentSection` (same query) |
| Agriculture summary cards | `Dashboard.vue` → `agriculture.*` | `dashboard_summary.go` | `fetchAgricultureSection` → farmer/irrigation/kharif/rabi queries |
| Land Holdings Distribution | `Dashboard.vue` → `landDistributionRows` | `dashboard_summary.go` | `fetchAgricultureSection` → `landDistributionQuery` |
| Land Utilization | `Dashboard.vue` → ApexCharts donut | `dashboard_summary.go` | `fetchAgricultureSection` → `landUtilQuery`, `invalidQuery` |
| Season-wise Crops | `Dashboard.vue` → ApexCharts bar | `dashboard_summary.go` | `fetchAgricultureSection` → `cropQuery` |
| Location filters (UI) | `Dashboard.vue` → `loadLocationOptions` | `handlers/location_options.go` | `GetLocationOptions` (3 queries) |
| Working / dependent population | (stored, not displayed) | `dashboard_summary.go` | `workingQuery`, `dependentQuery` |
| Age distribution buckets | (stored, not displayed) | `dashboard_summary.go` | `ageQuery` |

---

## Document metadata

| Field | Value |
|-------|-------|
| **Generated for** | Digital Twin — Agriculture Village Command Center |
| **Source revision** | Reflects `handlers/dashboard_summary.go` and `Dashboard.vue` as of documentation date |
| **Maintainer note** | Update this doc when dashboard SQL or API response shape changes |
