# Graph Report - frontend/src + handlers  (2026-04-17)

## Corpus Check
- 34 files · ~62,061 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 193 nodes · 234 edges · 27 communities detected
- Extraction: 89% EXTRACTED · 11% INFERRED · 0% AMBIGUOUS · INFERRED: 26 edges (avg confidence: 0.83)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Population 3D Twin & API|Population 3D Twin & API]]
- [[_COMMUNITY_Agri Twin & Layout Shell|Agri Twin & Layout Shell]]
- [[_COMMUNITY_Population PDF Generator|Population PDF Generator]]
- [[_COMMUNITY_Agri PDF Generator|Agri PDF Generator]]
- [[_COMMUNITY_Geo Map & Clustering|Geo Map & Clustering]]
- [[_COMMUNITY_Population REST Handlers|Population REST Handlers]]
- [[_COMMUNITY_Population Dashboard Handler|Population Dashboard Handler]]
- [[_COMMUNITY_Unified Registry|Unified Registry]]
- [[_COMMUNITY_Column Checker Utility|Column Checker Utility]]
- [[_COMMUNITY_Citizen Registry API|Citizen Registry API]]
- [[_COMMUNITY_House Detail Handler|House Detail Handler]]
- [[_COMMUNITY_Farmer Handler|Farmer Handler]]
- [[_COMMUNITY_Crop Handler|Crop Handler]]
- [[_COMMUNITY_Irrigation Handler|Irrigation Handler]]
- [[_COMMUNITY_Land Handler|Land Handler]]
- [[_COMMUNITY_Location Options Handler|Location Options Handler]]
- [[_COMMUNITY_Market Data|Market Data]]
- [[_COMMUNITY_Population Registry Handler|Population Registry Handler]]
- [[_COMMUNITY_Soil Data|Soil Data]]
- [[_COMMUNITY_Scheme Data|Scheme Data]]
- [[_COMMUNITY_App Entry Point|App Entry Point]]
- [[_COMMUNITY_Age Distribution|Age Distribution]]
- [[_COMMUNITY_Registry Filter UI|Registry Filter UI]]
- [[_COMMUNITY_Registry Pagination UI|Registry Pagination UI]]
- [[_COMMUNITY_Population PDF Report|Population PDF Report]]
- [[_COMMUNITY_Population Map Summary API|Population Map Summary API]]
- [[_COMMUNITY_Router Config|Router Config]]

## God Nodes (most connected - your core abstractions)
1. `PopulationHandler` - 16 edges
2. `buildPopulationPDF()` - 11 edges
3. `Agriculture Dashboard (Village Command Center)` - 9 edges
4. `AgricultureLayout (top-nav shell / AgriTwin)` - 8 edges
5. `PopulationMap (Geo-Intelligence 2D Map - population)` - 8 edges
6. `DigitalTwin (3D Cesium Twin)` - 7 edges
7. `fetchJSON() - population API fetch wrapper` - 7 edges
8. `buildAgriPDF()` - 7 edges
9. `UnifiedRegistry (multi-category citizen table)` - 6 edges
10. `MapView (Geo-Intelligence 2D Map - agriculture)` - 6 edges

## Surprising Connections (you probably didn't know these)
- `PopulationMap (Geo-Intelligence 2D Map - population)` --semantically_similar_to--> `Geo-Intelligence Map (household geospatial view)`  [INFERRED] [semantically similar]
  frontend/src/views/population/PopulationMap.vue → frontend/src/views/agriculture/MapView.vue
- `App (root component)` --references--> `AgricultureLayout (top-nav shell / AgriTwin)`  [INFERRED]
  frontend/src/App.vue → frontend/src/views/agriculture/AgricultureLayout.vue
- `Farmers (Citizen Registry view - agriculture)` --semantically_similar_to--> `UnifiedRegistry (multi-category citizen table)`  [INFERRED] [semantically similar]
  frontend/src/views/agriculture/Farmers.vue → frontend/src/components/UnifiedRegistry.vue
- `classifyIncome() helper (income bucketing)` --shares_data_with--> `Farmers (Citizen Registry view - agriculture)`  [INFERRED]
  frontend/src/components/UnifiedRegistry.vue → frontend/src/views/agriculture/Farmers.vue
- `Farmers (Citizen Registry view - agriculture)` --implements--> `Citizen Registry (unified multi-category view)`  [INFERRED]
  frontend/src/views/agriculture/Farmers.vue → frontend/src/components/UnifiedRegistry.vue

## Hyperedges (group relationships)
- **Household Data Consumers (getHouses API)** — mapview_mapview, digitaltwin_digitaltwin, api_index_gethouses [EXTRACTED 1.00]
- **AgriTwin Navigation Module Group** — agriculturelayout_agriculturelayout, dashboard_dashboard, farmers_farmers, mapview_mapview, digitaltwin_digitaltwin [EXTRACTED 1.00]
- **Leaflet-based Geo-Intelligence Map Stack** — mapview_mapview, populationmap_populationmap, ext_leaflet, ext_openstreetmap, ext_geohacker_geojson [INFERRED 0.85]
- **PopulationDashboard parallel API call pattern (dashboard + demographics + education + employment)** — populationdashboard_populationdashboard, population_api_getpopulationdashboard, population_api_getpopulationdemographics, population_api_getpopulationeducation, population_api_getpopulationemployment [EXTRACTED 1.00]
- **3D Twin problem filter → cluster analysis → color-coded Cesium building highlight pipeline** — population3dtwin_population3dtwin, population3dtwin_problem_filter_meta, population3dtwin_cluster_analysis, population3dtwin_cesium_viewer, concept_government_schemes [INFERRED 0.85]

## Communities

### Community 0 - "Population 3D Twin & API"
Cohesion: 0.11
Nodes (19): Cesium.js viewer - 3D population twin rendering, analyzeCluster() - cluster-level problem analysis, COLOR_MODE_LABELS - population density/BPL/divyang/employment color modes, District/Taluka/Village hierarchical filter in Population3DTwin, PROBLEM_FILTER_META - population problem filters, fetchJSON() - population API fetch wrapper, getPopulationDashboard(), getPopulationDemographics() (+11 more)

### Community 1 - "Agri Twin & Layout Shell"
Cohesion: 0.14
Nodes (18): AgricultureLayout (top-nav shell / AgriTwin), API Status Indicator (ping check), Theme Toggle (dark/light mode), getAgricultureInsights() API call, getGovernanceInsights() API call, getWelfareInsights() API call, App (root component), Digital Twin Platform (AgriTwin) (+10 more)

### Community 2 - "Population PDF Generator"
Cohesion: 0.19
Nodes (15): occupationStat, populationMetricBox, populationPDFRow, populationPDFStats, buildPopulationPDF(), chartPercentCaption(), computePopulationPDFStats(), drawKVMetrics() (+7 more)

### Community 3 - "Agri PDF Generator"
Cohesion: 0.18
Nodes (14): PDFChartImage, PDFChartSegment, PDFHandler, PDFProblemFilter, PDFRequest, pdfStats, rgbColor, buildAgriPDF() (+6 more)

### Community 4 - "Geo Map & Clustering"
Cohesion: 0.18
Nodes (13): getHouses() API call, getLocationOptions() API call, Geo-Intelligence Map (household geospatial view), GPS Anomaly / Mismatch Detector (MapView), MapView (Geo-Intelligence 2D Map - agriculture), Village Cluster Builder (MapView), Analytics Panel (donut charts - PopulationMap), getPopulationMapData() (population/api.js) (+5 more)

### Community 5 - "Population REST Handlers"
Cohesion: 0.19
Nodes (11): PopulationDemographicsResponse, PopulationEducationResponse, PopulationEmploymentResponse, PopulationMapInsightsResponse, PopulationMapMarker, PopulationMapSummaryResponse, populationVillageGeoStats, isMissingPopulationCoordinate() (+3 more)

### Community 6 - "Population Dashboard Handler"
Cohesion: 0.26
Nodes (1): PopulationHandler

### Community 7 - "Unified Registry"
Cohesion: 0.31
Nodes (7): UnifiedRecord, UnifiedRegistryHandler, capitalizeFullName(), capitalizeWords(), deriveIrrigationType(), deriveSourceOfIncome(), normaliseEducationLevel()

### Community 8 - "Column Checker Utility"
Cohesion: 0.22
Nodes (3): ColumnChecker, CountItem, InsightHandler

### Community 9 - "Citizen Registry API"
Cohesion: 0.32
Nodes (8): getCitizens() API call, getUnifiedRegistry() API call, Citizen Registry (unified multi-category view), Farmers (Citizen Registry view - agriculture), CATEGORY_CONFIG (farmer/student/disabled/housewife/senior), classifyIncome() helper (income bucketing), renderCell() (v-html badge renderer), UnifiedRegistry (multi-category citizen table)

### Community 10 - "House Detail Handler"
Cohesion: 0.29
Nodes (4): HouseDetail, HouseHandler, HouseRecord, MemberRecord

### Community 11 - "Farmer Handler"
Cohesion: 0.5
Nodes (2): FarmerHandler, FarmerRecord

### Community 12 - "Crop Handler"
Cohesion: 0.5
Nodes (2): CropHandler, CropRecord

### Community 13 - "Irrigation Handler"
Cohesion: 0.5
Nodes (2): IrrigationHandler, IrrigationRecord

### Community 14 - "Land Handler"
Cohesion: 0.5
Nodes (2): LandHandler, LandRecord

### Community 15 - "Location Options Handler"
Cohesion: 0.5
Nodes (2): LocationHandler, LocationOption

### Community 16 - "Market Data"
Cohesion: 0.67
Nodes (1): MarketPrice

### Community 17 - "Population Registry Handler"
Cohesion: 0.67
Nodes (2): PopulationRegistryRecord, populationRegistryRow

### Community 18 - "Soil Data"
Cohesion: 0.67
Nodes (1): Soil

### Community 19 - "Scheme Data"
Cohesion: 0.67
Nodes (1): Scheme

### Community 20 - "App Entry Point"
Cohesion: 1.0
Nodes (0): 

### Community 21 - "Age Distribution"
Cohesion: 1.0
Nodes (1): Age distribution bar chart

### Community 22 - "Registry Filter UI"
Cohesion: 1.0
Nodes (1): Population registry filter chips (BPL/Student/Divyang)

### Community 23 - "Registry Pagination UI"
Cohesion: 1.0
Nodes (1): Population registry pagination (50 records/page)

### Community 24 - "Population PDF Report"
Cohesion: 1.0
Nodes (1): PDF Report download in Population3DTwin

### Community 25 - "Population Map Summary API"
Cohesion: 1.0
Nodes (0): 

### Community 26 - "Router Config"
Cohesion: 1.0
Nodes (0): 

## Knowledge Gaps
- **62 isolated node(s):** `CATEGORY_CONFIG (farmer/student/disabled/housewife/senior)`, `renderCell() (v-html badge renderer)`, `getUnifiedRegistry() API call`, `getGovernanceInsights() API call`, `getWelfareInsights() API call` (+57 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `App Entry Point`** (1 nodes): `main.js`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Age Distribution`** (1 nodes): `Age distribution bar chart`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Registry Filter UI`** (1 nodes): `Population registry filter chips (BPL/Student/Divyang)`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Registry Pagination UI`** (1 nodes): `Population registry pagination (50 records/page)`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Population PDF Report`** (1 nodes): `PDF Report download in Population3DTwin`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Population Map Summary API`** (1 nodes): `getPopulationMapSummary()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Router Config`** (1 nodes): `index.js`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `PopulationHandler` connect `Population Dashboard Handler` to `Population PDF Generator`, `Population REST Handlers`?**
  _High betweenness centrality (0.059) - this node is a cross-community bridge._
- **Why does `buildPopulationPDF()` connect `Population PDF Generator` to `Agri PDF Generator`?**
  _High betweenness centrality (0.049) - this node is a cross-community bridge._
- **Why does `AgricultureLayout (top-nav shell / AgriTwin)` connect `Agri Twin & Layout Shell` to `Citizen Registry API`, `Geo Map & Clustering`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Are the 3 inferred relationships involving `buildPopulationPDF()` (e.g. with `pdfSectionTitle()` and `ppct()`) actually correct?**
  _`buildPopulationPDF()` has 3 INFERRED edges - model-reasoned connections that need verification._
- **Are the 2 inferred relationships involving `AgricultureLayout (top-nav shell / AgriTwin)` (e.g. with `App (root component)` and `Digital Twin Platform (AgriTwin)`) actually correct?**
  _`AgricultureLayout (top-nav shell / AgriTwin)` has 2 INFERRED edges - model-reasoned connections that need verification._
- **What connects `CATEGORY_CONFIG (farmer/student/disabled/housewife/senior)`, `renderCell() (v-html badge renderer)`, `getUnifiedRegistry() API call` to the rest of the system?**
  _62 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Population 3D Twin & API` be split into smaller, more focused modules?**
  _Cohesion score 0.11 - nodes in this community are weakly interconnected._