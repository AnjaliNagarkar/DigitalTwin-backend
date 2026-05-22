# 2D Geo-Intelligence Map — Query & API Documentation

Read-only technical reference for every backend query, API endpoint, and data flow used by the **Geo-Intelligence Map** (2D Map) module.

| Item | Value |
|------|-------|
| **UI page** | Geo-Intelligence Map |
| **Frontend component** | `frontend/src/views/agriculture/MapView.vue` |
| **Route** | `/agriculture/map` (via `AgricultureLayout.vue`) |
| **API client** | `frontend/src/api/index.js`, `frontend/src/views/population/api.js` |
| **Backend entry** | `main.go` (Gin router, port `:8081`) |
| **Auth** | All data routes use `handlers.AuthMiddleware()` (Bearer token) |

There is **no** `frontend/src/services/*` or map-specific composable; all logic lives in `MapView.vue` plus shared API helpers.

---

## Table of contents

1. [Architecture overview](#1-architecture-overview)
2. [API inventory (MapView only)](#2-api-inventory-mapview-only)
3. [APIs defined but not used by MapView](#3-apis-defined-but-not-used-by-mapview)
4. [Initial Maharashtra map load](#4-initial-maharashtra-map-load)
5. [District marker click → taluka drill-down](#5-district-marker-click--taluka-drill-down)
6. [Taluka marker click → village drill-down](#6-taluka-marker-click--village-drill-down)
7. [Village marker click → household markers](#7-village-marker-click--household-markers)
8. [Location filters (dropdowns, Apply, Reset)](#8-location-filters-dropdowns-apply-reset)
9. [View By dropdown (color modes)](#9-view-by-dropdown-color-modes)
10. [View Analytics panel](#10-view-analytics-panel)
11. [GPS mismatch detection](#11-gps-mismatch-detection)
12. [Cluster / marker counts](#12-cluster--marker-counts)
13. [GeoJSON & boundary loading](#13-geojson--boundary-loading)
14. [Household detail panel](#14-household-detail-panel)
15. [Villages view mode (client clusters)](#15-villages-view-mode-client-clusters)
16. [Dynamic filter helpers](#16-dynamic-filter-helpers)
17. [Database tables reference](#17-database-tables-reference)
18. [Startup cache (houses)](#18-startup-cache-houses)

---

## 1. Architecture overview

### End-to-end flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Browser: MapView.vue (Geo-Intelligence Map)                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│  onMounted                                                                  │
│    ├─ loadDistrictOptionsOnce()     → GET /api/districts                    │
│    ├─ Leaflet map init + OSM tiles    (no backend)                          │
│    ├─ addDistrictBorders()          → external GeoJSON (GitHub)             │
│    ├─ addMaharashtraHighlight()     → external GeoJSON (GitHub)           │
│    ├─ refreshDistrictCentroids()    → GET /api/map/district-centroids       │
│    └─ loadFamilyMembers()           → GET /api/population/map-data          │
│                                                                             │
│  Drill-down (centroid markers)                                              │
│    District click → getTalukaCentroids → render taluka markers               │
│    Taluka click   → getVillageCentroids → render village cluster markers    │
│    Village click  → getHouses (filtered) → plot household circle markers    │
│                                                                             │
│  Apply (location filters)                                                   │
│    applyFilters() → getHouses(district/taluka/village) → plotMarkers()       │
│                                                                             │
│  View By / Analytics / GPS mismatch                                         │
│    Client-side only on `houses` + population enrichment maps                 │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  main.go → protected GET routes under AuthMiddleware                        │
│  Handlers: houses.go, district_centroids.go, taluka_village_centroids.go,   │
│            district_survey_counts.go, location_options.go, population.go    │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  MySQL: FAMILY, FAMILY_MEMBER, district_master, taluka_master,              │
│         village_master, grampanchayat_master                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Navigation levels

| `navigationLevel` | Map layers | Primary data source |
|-------------------|------------|---------------------|
| `district` | Maharashtra mask + district polygons + **district centroid markers** | `GET /map/district-centroids` |
| `taluka` | Taluka centroid markers | `GET /map/taluka-centroids?district_id=` |
| `village` | Village markers (`L.markerClusterGroup`) | `GET /map/village-centroids?district_id=&taluka_id=` |
| `household` | Circle markers per house | `GET /houses?…` (Apply or village click) |

### View modes (header toggle)

| `viewMode` | UI label | Behavior |
|------------|----------|----------|
| `points` | Households | Individual `L.circleMarker` per loaded house |
| `villages` | Villages | Client-side `buildVillageClusters()` circles (no extra API) |

---

## 2. API inventory (MapView only)

| # | Frontend function | HTTP | Backend handler | Used when |
|---|-------------------|------|-----------------|-----------|
| 1 | `getDistricts()` | `GET /api/districts` | `LocationHandler.GetDistricts` | Mount: district dropdown |
| 2 | `getLocationOptions({ district_id })` | `GET /api/location-options?district_id=` | `LocationHandler.GetLocationOptions` | `watch(selectedDistrict)` → taluka list |
| 3 | `getLocationOptions({ district_id, taluka_id })` | `GET /api/location-options?…` | same | `watch(selectedTaluka)` → village list |
| 4 | `getDistrictCentroids(params)` | `GET /api/map/district-centroids?…` | `DistrictCentroidsHandler.GetDistrictCentroids` | Mount, reset, back-to-Maharashtra, Apply (no geo filter) |
| 5 | `getTalukaCentroids({ district_id })` | `GET /api/map/taluka-centroids?district_id=` | `TalukaCentroidsHandler.GetTalukaCentroids` | District centroid click, geo-nav back to taluka |
| 6 | `getVillageCentroids({ district_id, taluka_id })` | `GET /api/map/village-centroids?…` | `VillageCentroidsHandler.GetVillageCentroids` | Taluka click, geo-nav back to village |
| 7 | `getHouses(params)` | `GET /api/houses?…` | `HouseHandler.GetHouses` | Apply (district selected), village centroid click |
| 8 | `getPopulationMapData({})` | `GET /api/population/map-data` | `PopulationHandler.GetPopulationMapData` | Mount: global member stats enrichment |

**External (not backend):**

| Source | URL | Purpose |
|--------|-----|---------|
| District polygons | `https://raw.githubusercontent.com/geohacker/india/master/district/india_district.geojson` | Maharashtra district borders (styled) |
| State boundary | `https://raw.githubusercontent.com/geohacker/india/master/state/india_state.geojson` | Maharashtra outline + inverted mask |

---

## 3. APIs defined but not used by MapView

These exist in `frontend/src/api/index.js` and/or `main.go` but **are not called** from `MapView.vue`:

| API | Notes |
|-----|-------|
| `GET /view-options` | `getViewOptions()` — MapView uses **hardcoded** `groupedColorOptions` (i18n), not this endpoint |
| `GET /map/district-survey-counts` | Imported as `getDistrictSurveyCounts` but **never invoked** |
| `GET /houses/map-points` | Used by **3D Digital Twin** (`DigitalTwin.vue`), not 2D map |
| `GET /house/:id` | Used by 3D twin; MapView sets `selectedHouse` from loaded `houses` array |
| `GET /houses/batch-members` | 3D twin population enrichment |
| `GET /houses/summary` | Viewport grid clusters (3D / other flows) |
| `GET /insights/*` | Dashboard / other modules |
| `GET /advisory`, `GET /advisory/cluster` | Advisory panels elsewhere |

---

## 4. Initial Maharashtra map load

### 4.1 Feature summary

| Step | Frontend function | Backend? |
|------|-------------------|----------|
| District dropdown | `loadDistrictOptionsOnce()` | Yes — `GET /districts` |
| Map container + tiles | `onMounted` → `L.map`, `addTiles` | No (OpenStreetMap) |
| District borders | `addDistrictBorders(map)` | No — external GeoJSON |
| State highlight / mask | `addMaharashtraHighlight(map)` | No — external GeoJSON |
| District count markers | `refreshDistrictCentroids()` | Yes — `GET /map/district-centroids` |
| Population enrichment cache | `loadFamilyMembers()` | Yes — `GET /population/map-data` |
| Fit viewport | `fitToMaharashtra()`, `captureInitialMaharashtraFit()` | No |

**Important:** Initial load does **not** call `GET /houses`. Header shows `0 Households` until Apply or village drill-down loads houses.

### 4.2 `GET /api/districts`

| Field | Value |
|-------|-------|
| **Frontend** | `loadDistrictOptionsOnce()` → `getDistricts()` |
| **Route** | `GET /districts` → `LocationHandler.GetDistricts` |
| **File** | `handlers/location_options.go` |

```sql
SELECT
  dm.pklDistrictId,
  COALESCE(dm.vsDistrictName, '')
FROM district_master dm
WHERE dm.bEnabled = 1
ORDER BY dm.pklDistrictId
```

**Response fields used:**

| JSON field | Frontend use |
|------------|--------------|
| `pklDistrictId` | `districtOptions[].value` |
| `vsDistrictName` | `districtOptions[].label` |

Frontend prepends `{ label: 'All', value: null }`.

---

### 4.3 `GET /api/map/district-centroids`

| Field | Value |
|-------|-------|
| **Frontend** | `refreshDistrictCentroids()` → `getDistrictCentroids(getActiveLocationParams())` |
| **Route** | `GET /map/district-centroids` → `DistrictCentroidsHandler.GetDistrictCentroids` |
| **File** | `handlers/district_centroids.go` |
| **Click handler** | `handleDistrictCentroidClick(d, lat, lng)` |

**Query params sent:** `district_id`, `taluka_id`, `village_id` (from `getActiveLocationParams()`).

**Backend behavior:** Handler **ignores all query parameters** — always returns all districts with surveys. Params are sent from frontend but have no effect.

#### SQL (primary path — when `district_master` has lat/lng columns)

Column names are detected once via `INFORMATION_SCHEMA` (`detectDistrictMasterColumns`). Join key defaults to `pklDistrictId` if detection fails.

```sql
SELECT
  f.DISTRICT_ID,
  MAX(COALESCE(dm.vsDistrictName, '')) AS district_name,
  COUNT(*) AS total_count,
  COALESCE(AVG(f.LATITUDE), MAX(dm.<latCol>)) AS lat,
  COALESCE(AVG(f.LONGITUDE), MAX(dm.<lngCol>)) AS lng
FROM FAMILY f
LEFT JOIN district_master dm
  ON f.DISTRICT_ID = dm.<idCol>
WHERE f.DISTRICT_ID IS NOT NULL
GROUP BY f.DISTRICT_ID
```

#### SQL (fallback — no master coordinates)

```sql
SELECT
  f.DISTRICT_ID,
  MAX(COALESCE(dm.vsDistrictName, '')) AS district_name,
  COUNT(*) AS total_count,
  AVG(f.LATITUDE) AS lat,
  AVG(f.LONGITUDE) AS lng
FROM FAMILY f
LEFT JOIN district_master dm
  ON f.DISTRICT_ID = dm.<idCol>
WHERE f.DISTRICT_ID IS NOT NULL
GROUP BY f.DISTRICT_ID
```

**Tables:** `FAMILY`, `district_master`

**Response → UI:**

| JSON field | UI |
|------------|-----|
| `district_id` | Drill selection, tooltip |
| `district_name` | Tooltip label |
| `count` | Orange marker badge HTML |
| `lat`, `lng` | `L.marker` position; invalid coords skipped |

---

### 4.4 `GET /api/population/map-data` (enrichment only)

| Field | Value |
|-------|-------|
| **Frontend** | `loadFamilyMembers()` → `getPopulationMapData({})` |
| **Route** | `GET /population/map-data` → `PopulationHandler.GetPopulationMapData` |
| **File** | `handlers/population.go`, route in `routes/population.go` |

Called once on mount with **no filters** (global dataset for enrichment lookups).

#### Dynamic WHERE (built by `buildPopulationFamilyFilters`)

Base clause always starts as `1=1`, then:

```sql
-- Optional (MapView sends none on mount):
(? IS NULL OR ? = '' OR CAST(f.DISTRICT_ID AS CHAR) = ?)   -- district_id ×3 bind
(? IS NULL OR ? = '' OR CAST(f.TALUKA_ID AS CHAR) = ?)     -- taluka_id ×3 bind
(? IS NULL OR ? = '' OR CAST(f.VILLAGE_ID AS CHAR) = ?)    -- village_id ×3 bind
-- Optional state_id if column STATE_ID exists
```

Full outer filter:

```sql
WHERE f.LATITUDE IS NOT NULL AND f.LONGITUDE IS NOT NULL
  AND f.LATITUDE != 0 AND f.LONGITUDE != 0
  AND <buildPopulationFamilyFilters clauses>
```

#### Core SELECT (abbreviated; member aggregates in subquery)

The handler runs one large query joining `FAMILY` to aggregated `FAMILY_MEMBER` (occupation list, BPL, divyang, aadhaar/caste coverage). See `handlers/population.go` lines 818–918.

**Key CASE / aggregation logic:**

| Metric | SQL pattern |
|--------|-------------|
| Working members | `SUM(CASE WHEN occupation not in housewife/student/unemployed/... THEN 1 END)` |
| Aadhaar coverage | `CASE WHEN all members yes → 'complete'; partial → 'partial'; else 'missing'` |
| Divyang | `SUM(CASE WHEN DIVYANG='YES' OR DISABILITY='YES' THEN 1 END)` (column-gated) |
| Occupation list | `GROUP_CONCAT(DISTINCT … SEPARATOR '|')` excluding non-working values |

**Response:** JSON **array** of marker objects (not wrapped in `{ data: … }`).

**Frontend maps:**

| Backend field | Frontend map |
|---------------|--------------|
| `external_family_id` | `populationStatsByFamily` key |
| `total_members`, `male_members`, `female_members`, … | Enrichment via `buildPopulationStatsMap` |
| `head_name`, `lat`, `lng`, `house_no`, `village_id` | Signature / house-village fallback maps |

`enrichHouseholdForPopulation()` merges `/houses` rows with these stats when member fields are missing.

---

## 5. District marker click → taluka drill-down

### 5.1 Feature flow

| Step | Function | API |
|------|----------|-----|
| 1 | `handleDistrictCentroidClick` | — |
| 2 | Set `selectedDistrict`, `navigationLevel = 'taluka'` | — |
| 3 | `map.flyTo([lat,lng], 9)` | — |
| 4 | `refreshTalukaCentroids()` | `GET /map/taluka-centroids?district_id=` |
| 5 | `renderTalukaCentroids(rows)` | — |

`watch(selectedDistrict)` also loads taluka dropdown via `GET /location-options?district_id=`.

### 5.2 `GET /api/map/taluka-centroids`

| Field | Value |
|-------|-------|
| **Route** | `GET /map/taluka-centroids` |
| **Handler** | `TalukaCentroidsHandler.GetTalukaCentroids` |
| **Required param** | `district_id` (integer) |

```sql
SELECT
  f.TALUKA_ID,
  tm.vsTalukaName AS taluka_name,
  COUNT(*) AS total_count,
  AVG(CAST(f.LATITUDE AS DECIMAL(10,6))) AS lat,
  AVG(CAST(f.LONGITUDE AS DECIMAL(10,6))) AS lng
FROM FAMILY f
LEFT JOIN taluka_master tm
  ON tm.pklTalukaId = f.TALUKA_ID
WHERE f.DISTRICT_ID = ?
  AND f.LATITUDE IS NOT NULL
  AND f.LONGITUDE IS NOT NULL
GROUP BY f.TALUKA_ID, tm.vsTalukaName
```

**Dynamic filters:** Only `district_id` (required). No taluka/village filter.

**Tables:** `FAMILY`, `taluka_master`

**Response fields:** `taluka_id`, `taluka_name`, `count`, `lat`, `lng` → taluka marker badge + `handleTalukaCentroidClick`.

---

## 6. Taluka marker click → village drill-down

### 6.1 Feature flow

| Step | Function | API |
|------|----------|-----|
| 1 | `handleTalukaCentroidClick` | — |
| 2 | Set `selectedTaluka`, `navigationLevel = 'village'` | — |
| 3 | `loadVillageOptionsByTaluka(districtId, talukaId)` | `GET /location-options?district_id=&taluka_id=` |
| 4 | `refreshVillageCentroids()` | `GET /map/village-centroids?…` |
| 5 | `map.flyTo([lat,lng], 11)` | — |

### 6.2 `GET /api/map/village-centroids`

| Field | Value |
|-------|-------|
| **Route** | `GET /map/village-centroids` |
| **Handler** | `VillageCentroidsHandler.GetVillageCentroids` |
| **Required params** | `district_id`, `taluka_id` |

```sql
SELECT
  f.VILLAGE_ID,
  vm.vsVillageName AS village_name,
  COUNT(*) AS total_count,
  AVG(CAST(f.LATITUDE AS DECIMAL(10,6))) AS lat,
  AVG(CAST(f.LONGITUDE AS DECIMAL(10,6))) AS lng
FROM FAMILY f
LEFT JOIN village_master vm
  ON vm.pklVillageId = f.VILLAGE_ID
WHERE f.DISTRICT_ID = ?
  AND f.TALUKA_ID = ?
  AND f.LATITUDE IS NOT NULL
  AND f.LONGITUDE IS NOT NULL
GROUP BY f.VILLAGE_ID, vm.vsVillageName
```

**Tables:** `FAMILY`, `village_master`

**UI:** `L.markerClusterGroup` — cluster icon sums `villageCount` on child markers; click → `handleVillageCentroidClick`.

---

## 7. Village marker click → household markers

### 7.1 Feature flow

| Step | Function | API |
|------|----------|-----|
| 1 | `handleVillageCentroidClick` | — |
| 2 | Set `selectedVillage`, `navigationLevel = 'household'` | — |
| 3 | `loadLiveHouseData(token)` → `fetchAllHouses()` | `GET /houses?…` |
| 4 | `plotMarkers(enriched)` | — |
| 5 | `map.flyTo([lat,lng], 14)` | — |

Filters applied: `district_id`, `taluka_id`, `village_id` from current dropdown selections (set by drill + watchers).

### 7.2 `GET /api/houses`

| Field | Value |
|-------|-------|
| **Frontend** | `fetchAllHouses()` → `getHouses(apiParams)` |
| **Route** | `GET /houses` → `HouseHandler.GetHouses` |
| **File** | `handlers/houses.go` |

#### Query parameters

| Param | Default | Purpose |
|-------|---------|---------|
| `page` | `1` | Pagination |
| `limit` | `500` (max 5000) | Page size; MapView uses `500` with location filter, `2000` without |
| `district_id` | — | Geo filter |
| `taluka_id` | — | Geo filter |
| `village_id` | — | Geo filter |
| `irrigation` | — | Optional `f.SOURCE_WATER_IRRIGATION = ?` |
| `own_land` | — | Optional `f.OWN_AGRICULTURE_LAND = ?` |
| `bbox` or `min_lat`/`max_lat`/`min_lng`/`max_lng` | — | Viewport filter (not used by MapView) |

**MapView `getHouseFilters()`:**

```javascript
{ page: 1, limit: 500|2000, district_id, taluka_id, village_id }
```

#### Dynamic WHERE (coordinate + geo filters)

Base (lat/lng column names from `ColumnChecker`, default `LATITUDE`/`LONGITUDE`):

```sql
WHERE f.<latCol> IS NOT NULL AND f.<lngCol> IS NOT NULL
  AND f.<latCol> != 0 AND f.<lngCol> != 0
```

**Optional filters (appended in order):**

```sql
AND f.SOURCE_WATER_IRRIGATION = ?          -- if irrigation query param
AND f.OWN_AGRICULTURE_LAND = ?             -- if own_land query param
AND f.DISTRICT_ID = ?                      -- if district_id (numeric equality preferred)
AND f.TALUKA_ID = ?                        -- if taluka_id
AND f.VILLAGE_ID = ?                       -- if village_id
AND f.<latCol> > ? AND f.<latCol> < ?      -- bbox / viewport
AND f.<lngCol> > ? AND f.<lngCol> < ?
```

`appendFamilyGeoIDFilter`: numeric string → `column = int`; else `CAST(column AS CHAR) = ?`.

#### Main query structure (CTE pagination)

```sql
WITH page_families AS (
  SELECT f.*
  FROM FAMILY f
  <WHERE above>
  ORDER BY f.FAMILY_ID
  LIMIT <limit> OFFSET <offset>
)
SELECT
  f.FAMILY_ID,
  f.EXTERNAL_FAMILY_ID,
  f.HOUSE_NO,
  f.DISTRICT_ID, district name,
  f.TALUKA_ID, taluka name,
  f.VILLAGE_ID, village name,
  f.<latCol>, f.<lngCol>,
  f.AREA_AGRICULTURE_LAND_ACRES,
  f.LAND_UNDER_CULTIVATION_ACRES,
  f.OWN_AGRICULTURE_LAND,
  f.SOURCE_WATER_IRRIGATION,
  f.CULTIVATING_DURING_KHARIF_SEASON,
  f.CULTIVATING_DURING_RABI_SEASON,
  f.TYPE_HOUSE, f.OWNERSHIP_HOUSE, f.PRADHAN_MANTRI_AWAS,
  f.SANITATION_TOILET_FACILITY_HOME,
  f.A_SOAKPIT_MANAGING_WASTEWATER,
  f.RATION_CARD_COLOR,
  aadhaar_agg.*, caste_agg.*,
  <sanitationExpr>, <lightingExpr>, <rationExpr>,
  <bplExpr>, <incomeExpr>,
  fm_agg.primary_occupation,
  head name,
  fm_agg.total_members, male_members, female_members,
  working_members, illiterate_members, divyang_members, unemployed_members,
  divyang_agg.divyang_members_json
FROM page_families f
LEFT JOIN fm_agg ON EXTERNAL_FAMILY_ID
LEFT JOIN aadhaar_agg …
LEFT JOIN caste_agg …
LEFT JOIN divyang_agg …
LEFT JOIN district_master, taluka_master, village_master
ORDER BY f.FAMILY_ID
```

#### Member aggregation CASE highlights (`fm_agg` subquery)

| Output | Logic |
|--------|--------|
| `primary_occupation` | `GROUP_CONCAT(DISTINCT CASE WHEN occupation is working… THEN TRIM(occupation) END SEPARATOR '|')` |
| `working_members` | Count where occupation not in unemployed/housewife/student/… |
| `divyang_members` | `SUM(CASE DIVYANG='YES' OR DISABILITY='YES')` if columns exist |
| `aadhaar_coverage_status` | `complete` / `partial` / `missing` / `unknown` from `AADHAAR='yes'` counts |
| `caste_certificate_coverage_status` | Same pattern on `CASTE_CERTIFICATE` |

**Tables:** `FAMILY`, `FAMILY_MEMBER`, `district_master`, `taluka_master`, `village_master`

**Response shape:**

```json
{
  "data": [ { "familyId", "latitude", "longitude", "villageId", … } ],
  "total": <rows in this page only>,
  "page": 1,
  "limit": 500
}
```

**Frontend pipeline:**

1. `extractFamiliesFromResponse(res)` → uses `res.data` array
2. `enrichHouseholdForPopulation(family, memberStatsLookup)`
3. `houses.value = enriched`
4. `plotMarkers()` → `L.circleMarker` colored by `getMarkerColor(house)`

**Note:** `total` in response is **page row count**, not full filtered COUNT (by design in handler).

---

## 8. Location filters (dropdowns, Apply, Reset)

### 8.1 Dropdown data sources

| Dropdown | Load trigger | API | SQL section |
|----------|--------------|-----|-------------|
| District | `loadDistrictOptionsOnce()` | `GET /districts` | [§4.2](#42-get-apidistricts) |
| Taluka | `watch(selectedDistrict)` | `GET /location-options?district_id=` | Taluka query below |
| Village | `watch(selectedTaluka)` | `GET /location-options?district_id=&taluka_id=` | Village query below |

Changing village dropdown alone does **not** fetch houses (`watch(selectedVillage)` is empty). Houses load only on **Apply** or **village centroid click**.

### 8.2 `GET /api/location-options`

| Field | Value |
|-------|-------|
| **Handler** | `LocationHandler.GetLocationOptions` |
| **File** | `handlers/location_options.go` |

Always returns three arrays: `districts`, `talukas`, `villages` (taluka/village lists filtered when params present).

#### Districts (always full list)

```sql
SELECT CAST(dm.pklDistrictId AS CHAR), COALESCE(dm.vsDisplayName, dm.vsDistrictName, '')
FROM district_master dm
WHERE dm.bEnabled = 1
ORDER BY COALESCE(dm.vsDisplayName, dm.vsDistrictName)
```

#### Talukas (dynamic)

```sql
SELECT CAST(tm.pklTalukaId AS CHAR), COALESCE(tm.vsDisplayName, tm.vsTalukaName, ''),
       CAST(tm.fklDistrictId AS CHAR)
FROM taluka_master tm
WHERE tm.bEnabled = 1
  AND CAST(tm.fklDistrictId AS CHAR) IN (?)   -- when district_id / district_ids provided
ORDER BY COALESCE(tm.vsDisplayName, tm.vsTalukaName)
```

#### Villages (dynamic)

```sql
SELECT DISTINCT CAST(vm.pklVillageId AS CHAR),
       COALESCE(vm.vsDisplayName, vm.vsVillageName, ''),
       CAST(gm.fklTalukaId AS CHAR)
FROM village_master vm
JOIN grampanchayat_master gm ON gm.pklGramPanchayatId = vm.fklGramPanchayatId
JOIN taluka_master tm ON tm.pklTalukaId = gm.fklTalukaId
WHERE vm.bEnabled = 1
  AND CAST(tm.fklDistrictId AS CHAR) IN (?)   -- optional district filter
  AND CAST(tm.pklTalukaId AS CHAR) IN (?)     -- when taluka_id / taluka_ids provided
ORDER BY COALESCE(vm.vsDisplayName, vm.vsVillageName)
```

**Tables:** `district_master`, `taluka_master`, `village_master`, `grampanchayat_master`

---

### 8.3 Apply button — `applyFilters(autoZoomToResults)`

| Condition | Behavior |
|-----------|----------|
| No `district_id` selected | Clear houses/markers; `navigationLevel = 'district'`; `refreshDistrictCentroids()` only |
| District selected | `navigationLevel = 'household'`; `loadLiveHouseData()` → `GET /houses`; clear centroid layers if any geo filter set |
| `autoZoomToResults === true` | Sets `fitAfterLoad`; `plotMarkers` runs `flyToBounds` on results |

**Does not** call location-options or centroids again except district refresh when filter cleared.

---

### 8.4 Reset button — `resetFilters()`

| Action | API |
|--------|-----|
| Clear selections + houses | — |
| `refreshDistrictCentroids()` | `GET /map/district-centroids` |
| `restoreInitialMaharashtraFit()` | — (stored Leaflet bounds from mount) |

---

### 8.5 Geo-nav back — `handleGeoNavBack()`

Uses `zoomStack` (center, zoom, level). Restores prior level and re-fetches centroids:

| Previous level | APIs |
|----------------|------|
| `district` | `refreshDistrictCentroids()` + Maharashtra fit |
| `taluka` | `loadTalukaOptionsByDistrict` + `refreshTalukaCentroids()` |
| `village` | `loadVillageOptionsByTaluka` + `refreshVillageCentroids()` |

---

## 9. View By dropdown (color modes)

### 9.1 Configuration source

**Not loaded from `GET /view-options`.** Options are hardcoded in `MapView.vue`:

| Group | Modes (`colorMode`) |
|-------|---------------------|
| Population | `population_density`, `bpl_status`, `divyang_presence`, `employment_status` |
| Agriculture | `crops`, `irrigation`, `land` |
| Infrastructure | `housing_quality`, `electricity`, `toilet_access`, `wastewater_management` |
| Document gap | `aadhaar_coverage`, `caste_certificate_coverage` |

`GET /view-options` (`handlers/view_options.go`) builds a **different** schema-driven list at server startup for other screens; 2D map does not consume it.

### 9.2 Data source for coloring

All modes use fields already on each house object from:

1. **`GET /houses`** — primary attributes
2. **`GET /population/map-data`** — fallback member/BPL/occupation stats via `enrichHouseholdForPopulation`

**No additional API call** when user changes View By.

### 9.3 Per-mode logic (client-side)

| View By label | `colorMode` | Color rule (`getMarkerColor`) | Source columns |
|---------------|-------------|-------------------------------|----------------|
| Population Density | `population_density` | Green shades by `getTotalMembers(house)` | `totalMembers` / population map fallback |
| BPL Status | `bpl_status` | Red=BPL, green=non-BPL | `bplCategory`, `FAMILY_BELONG_BPL_CATEGORY`, ration card |
| Divyang Presence | `divyang_presence` | Purple=yes, gray=no | `divyangMembers`, `has_disability` |
| Employment Status | `employment_status` | Amber=working | `workingMembers`, `occupation` |
| Crops / Season | `crops` | By kharif/rabi presence | `kharif`, `rabi` (`CULTIVATING_DURING_*`) |
| Irrigation | `irrigation` | Green=irrigated, red=rain-fed | `SOURCE_WATER_IRRIGATION` / `waterSource` |
| Land Holdings | `land` | Red/amber/green by acres | `AREA_AGRICULTURE_LAND_ACRES`, `totalLand` |
| Housing Quality | `housing_quality` | By `TYPE_HOUSE` PUCCA/KUCHA | `TYPE_HOUSE` |
| Electricity | `electricity` | Blue=yes | `lighting` / `ELECTRICITY_CONNECTION` |
| Toilet Access | `toilet_access` | Green/red | `SANITATION_TOILET_FACILITY_HOME` |
| Wastewater Management | `wastewater_management` | By `A_SOAKPIT_MANAGING_WASTEWATER` | same column |
| Aadhaar Coverage | `aadhaar_coverage` | blue/amber/red/gray by status | `aadhaarCoverageStatus` (SQL CASE in GetHouses) |
| Caste Certificate | `caste_certificate_coverage` | same pattern | `casteCertificateCoverageStatus` |

#### SQL CASE (backend) backing document modes

Executed inside `GET /houses` aggregations:

**Aadhaar:**

```sql
CASE
  WHEN COUNT(*) = 0 THEN 'unknown'
  WHEN <members_with_aadhaar> = COUNT(*) THEN 'complete'
  WHEN <members_with_aadhaar> > 0 THEN 'partial'
  ELSE 'missing'
END AS aadhaar_coverage_status
-- members_with_aadhaar = SUM(CASE WHEN LOWER(TRIM(AADHAAR)) = 'yes' THEN 1 ELSE 0 END)
```

**Caste certificate:** identical pattern on `CASTE_CERTIFICATE`.

**BPL (display):** `FAMILY_BELONG_BPL_CATEGORY` via `ColOrEmpty` in SELECT.

**Employment (display):** `working_members` + filtered `GROUP_CONCAT` occupation list in `fm_agg`.

---

## 10. View Analytics panel

### 10.1 Feature summary

| Item | Value |
|------|-------|
| **Toggle** | `analyticsPanelOpen` |
| **Chart data** | `analyticsChart` computed property |
| **Backend API** | **None** — 100% client-side aggregation on `houses.value` |
| **Prerequisite** | `houses.length > 0` and `stats` computed non-null |

Supporting stats (`stats` computed):

```javascript
// From houses.value only:
total = houses.length
farmers = count(OWN_AGRICULTURE_LAND / ownLand === 'yes')
noIrrigation = count(isNoIrrigationValue(waterSource))
kharif / rabi = count(non-empty crop fields)
```

### 10.2 Per-mode analytics segments

| `colorMode` | Chart title (i18n) | Segments |
|-------------|-------------------|----------|
| `population_density` | Gender distribution | male / female totals (`getMaleMembers`, `getFemaleMembers`) |
| `bpl_status` | BPL distribution | BPL vs non-BPL (`getBplStatus === 'yes'`) |
| `electricity` | Electricity | hasElectricity yes/no |
| `aadhaar_coverage` | Aadhaar | complete / partial / missing / unknown |
| `caste_certificate_coverage` | Caste cert | same buckets |
| `divyang_presence` | Divyang | hasDivyangPresence |
| `toilet_access` | Toilet | hasToilet |
| `wastewater_management` | Wastewater | yes / no / unknown |
| `employment_status` | Employment | hasEmployment |
| `crops` | Crop seasons | kharif count, rabi count (`stats.kharif`, `stats.rabi`) |
| `irrigation` | Irrigation | irrigated vs `stats.noIrrigation` |
| `land` | Land holding | marginal ≤1ac, small ≤2.5ac, medium/large |
| `housing_quality` | Housing | PUCCA / KUCHA / unknown |

**Filters:** Uses whatever households are currently in `houses.value` (from last Apply or village click). Changing district dropdown without Apply does not change analytics.

---

## 11. GPS mismatch detection

### 11.1 Feature summary

| Item | Value |
|-------|-------|
| **UI** | “GPS Mismatches” toggle + anomaly sidebar |
| **Backend API** | **None** |
| **Algorithm** | `anomalies` computed in `MapView.vue` |

### 11.2 Client algorithm

1. Filter `houses` with valid `villageId`, `latitude`, `longitude`.
2. Group by `villageId`.
3. Skip groups with `< 3` households (`MIN_GROUP_SIZE`).
4. Per village: compute centroid (mean lat/lng), distances (Haversine), mean + std dev.
5. Threshold = `max(mean + 2.5 * stdDev, 5 km)`.
6. Flag houses beyond threshold; find nearest **other** village centroid for “plotted village” label.

**Tables (indirect):** Uses coordinates from `FAMILY` originally loaded via `GET /houses`.

**Response fields used on flagged houses:**

| Internal field | UI |
|----------------|-----|
| `_distanceKm` | Sidebar distance |
| `_plottedVillage` | “Likely village” mismatch hint |
| `_centroidLat/Lng` | Debug / enrichment |

Clicking flagged marker sets `selectedHouse` to enriched anomaly row (no `GET /house/:id`).

---

## 12. Cluster / marker counts

### 12.1 District / taluka / village centroid counts

| Level | API | Count field | SQL aggregation |
|-------|-----|-------------|-----------------|
| District | `/map/district-centroids` | `count` | `COUNT(*)` per `DISTRICT_ID` |
| Taluka | `/map/taluka-centroids` | `count` | `COUNT(*)` per `TALUKA_ID` in district |
| Village | `/map/village-centroids` | `count` | `COUNT(*)` per `VILLAGE_ID` in taluka |

### 12.2 Household count (header subtitle)

Displays `houses.length` — count of records returned by last `GET /houses` page (≤ limit), not statewide total.

### 12.3 Village view mode clusters

`buildVillageClusters(houses)` — client-side buckets by `villageId` or rounded lat/lng; **no SQL**.

### 12.4 `GET /map/district-survey-counts` (unused by MapView)

Documented for completeness — available in API client but **not called** from MapView.

| Field | Value |
|-------|-------|
| **Handler** | `DistrictSurveyCountHandler.GetDistrictSurveyCounts` |

```sql
SELECT
  CAST(f.DISTRICT_ID AS UNSIGNED) AS district_id,
  MAX(COALESCE(d.vsDisplayName, d.vsDistrictName, '')) AS district_name,
  COUNT(*) AS survey_count
FROM FAMILY f
JOIN district_master d ON d.pklDistrictId = f.DISTRICT_ID
WHERE 1=1
  -- optional:
  AND CAST(f.DISTRICT_ID AS CHAR) = ?   -- district_id
  AND CAST(f.TALUKA_ID AS CHAR) = ?     -- taluka_id
  AND CAST(f.VILLAGE_ID AS CHAR) = ?    -- village_id
GROUP BY f.DISTRICT_ID
ORDER BY survey_count DESC
```

---

## 13. GeoJSON & boundary loading

### 13.1 No backend GeoJSON

Maharashtra and district boundaries are **not** loaded from MySQL or Digital Twin APIs.

### 13.2 External sources

| Layer | Function | URL |
|-------|----------|-----|
| District polygons | `addDistrictBorders` | `https://raw.githubusercontent.com/geohacker/india/master/district/india_district.geojson` |
| State boundary + mask | `addMaharashtraHighlight` | `https://raw.githubusercontent.com/geohacker/india/master/state/india_state.geojson` |

**Filter (client):** Features where properties contain `MAHARASHTRA` in `ST_NM`, `state`, `STATE`, or `NAME_1`.

**Map constants (client):**

```javascript
MAHARASHTRA_BOUNDS = L.latLngBounds([15.6, 72.6], [22.1, 80.9])
MAHARASHTRA_CENTER = [19.7515, 75.7139]
MAHARASHTRA_INITIAL_ZOOM = 7
```

### 13.3 Tiles

OpenStreetMap raster tiles via Leaflet (standard OSM attribution) — no SQL.

---

## 14. Household detail panel

| Item | Value |
|-------|-------|
| **Trigger** | Click `L.circleMarker` → `selectedHouse = house` |
| **API** | **None** on click — uses in-memory object from `GET /houses` |
| **Template** | `v-if="selectedHouse && viewMode === 'points'"` — fields vary by `colorMode` |

`GET /house/:id` exists (`HouseHandler.GetHouseByID`) and serves from **startup preloaded** `houseCache` (same data as bulk preload query in `buildHouseCacheQuery()`). **MapView does not call it.**

---

## 15. Villages view mode (client clusters)

| Item | Value |
|-------|-------|
| **Toggle** | `setViewMode('villages')` |
| **Function** | `buildVillageClusters` → `drawClusters` |
| **API** | None — uses current `houses.value` |

Cluster aggregates: `count`, `noToilet`, `noElec`, `noIrrig`, `bpl` per village bucket.

---

## 16. Dynamic filter helpers

### 16.1 `appendFamilyGeoIDFilter` (`handlers/houses.go`)

Used by `GetHouses`, `GetHousesMapPoints`:

```
IF raw ID parses as int64:
  AND f.<COLUMN> = ?
ELSE:
  AND CAST(f.<COLUMN> AS CHAR) = ?
```

### 16.2 `buildPopulationFamilyFilters` (`handlers/population.go`)

Nullable triplet pattern per geo ID:

```sql
(? IS NULL OR ? = '' OR CAST(f.DISTRICT_ID AS CHAR) = ?)
```

Binds each ID three times. MapView calls `/population/map-data` with **empty** params → only `1=1` + coordinate filter.

---

## 17. Database tables reference

| Table | Map module usage |
|-------|------------------|
| `FAMILY` | Coordinates, land, crops, irrigation, housing, BPL, all survey fields; join key `EXTERNAL_FAMILY_ID` |
| `FAMILY_MEMBER` | Aggregated counts, occupation, aadhaar, caste, divyang, gender |
| `district_master` | Dropdowns, names, optional centroid fallback coords |
| `taluka_master` | Dropdowns, taluka centroids |
| `village_master` | Village centroids |
| `grampanchayat_master` | Join path village → taluka in location-options |

---

## 18. Startup cache (houses)

On server start (`main.go`):

```go
houseHandler.PreloadHouseCache()
```

Loads all geo-valid families into `houseCache` (in-memory `sync.Map`) for:

- `GET /house/:id` (O(1), header `X-Cache: HIT`)
- `GET /houses/batch-members` (no DB)

**MapView** uses live `GET /houses` with pagination instead of the full cache endpoint.

Preload SQL: `buildHouseCacheQuery()` in `handlers/houses.go` — single CTE scan of `FAMILY` + one `FAMILY_MEMBER` aggregation (`buildPopStatsSQL()`).

---

## Appendix A — `GET /houses/map-points` (3D only)

For cross-reference (not MapView):

```sql
SELECT f.FAMILY_ID, COALESCE(f.<latCol>, 0), COALESCE(f.<lngCol>, 0)
FROM FAMILY f
WHERE f.<latCol> IS NOT NULL AND f.<lngCol> IS NOT NULL
  AND f.<latCol> != 0 AND f.<lngCol> != 0
  -- optional district_id, taluka_id, village_id via appendFamilyGeoIDFilter
ORDER BY f.FAMILY_ID
LIMIT <optional>
```

---

## Appendix B — `GET /view-options` (unused by MapView)

`ViewOptionsHandler.buildGroups()` at startup introspects `FAMILY` / `FAMILY_MEMBER` columns and returns grouped options with `action: "colorMode"` or `"view"`. Example groups: Population, Infrastructure, Document Gap Analysis, Agriculture.

---

## Appendix C — File index

| File | Role |
|------|------|
| `frontend/src/views/agriculture/MapView.vue` | All map UI, drill-down, analytics, GPS anomaly |
| `frontend/src/api/index.js` | HTTP wrappers (`/api` prefix) |
| `frontend/src/views/population/api.js` | `getPopulationMapData` |
| `main.go` | Route registration |
| `handlers/houses.go` | `GET /houses`, map-points, summary, batch-members, house/:id |
| `handlers/district_centroids.go` | District centroids |
| `handlers/taluka_village_centroids.go` | Taluka & village centroids |
| `handlers/district_survey_counts.go` | District survey counts (unused in MapView) |
| `handlers/location_options.go` | Districts & location-options |
| `handlers/view_options.go` | View-options (unused in MapView) |
| `handlers/population.go` | `/population/map-data` |
| `routes/population.go` | Population route group |

---

*Generated from read-only codebase investigation. No application code was modified.*
