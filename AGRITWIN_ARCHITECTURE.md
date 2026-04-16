# AgriTwin: Agricultural Data Visualization & Analysis Platform
## Comprehensive Architecture & Features Documentation

---

## 1. Executive Summary

**AgriTwin** is an advanced agricultural data visualization and analysis platform designed to provide real-time insights into household-level agricultural data, infrastructure conditions, and welfare metrics across rural Maharashtra. The platform integrates geospatial intelligence with digital twin technology, enabling government agencies and agricultural organizations to identify at-risk households, optimize resource allocation, and monitor rural development initiatives.

### Key Capabilities
- **Household Registry**: Browse and filter 10,000+ household records with farming profiles
- **Geospatial Intelligence**: Interactive 2D mapping with village-level clustering and heatmaps
- **3D Digital Twin**: Cesium-based terrain visualization with problem highlighting
- **Insights Engine**: Automated analysis of governance, agriculture, and welfare metrics
- **PDF Reporting**: Generate customized reports on selected regions with problem filtering
- **Multi-level Filtering**: Navigate the hierarchy (District → Taluka → Village) with real-time filtering

---

## 2. System Architecture

### 2.1 Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Frontend (Vue 3 + Vite)                      │
│  ┌──────────────┬──────────────┬──────────────┬──────────────────┐  │
│  │  Dashboard   │   Farmers    │   2D Map     │   3D Digital Twin│  │
│  │              │   Registry   │   (Leaflet)  │   (Cesium.js)    │  │
│  └──────────────┴──────────────┴──────────────┴──────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────┐ │
│  │              API Client Layer (fetch with timeout)               │ │
│  └──────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    │                           │
                    ▼                           ▼
        ┌──────────────────────┐    ┌──────────────────────┐
        │  Go Backend (Gin)    │    │   SQL Database       │
        │  Port 8081           │    │   MySQL (IVDP)       │
        │  Read-Only Mode      │    │                      │
        │  ┌────────────────┐  │    │  ┌────────────────┐ │
        │  │ /houses        │  │    │  │ FAMILY         │ │
        │  │ /farmers       │  │    │  │ FAMILY_MEMBER  │ │
        │  │ /insights/*    │  │    │  │ Master Tables  │ │
        │  │ /pdf/report    │  │    │  │ (district,     │ │
        │  │ /location-opts │  │    │  │  taluka,       │ │
        │  └────────────────┘  │    │  │  village)      │ │
        └──────────────────────┘    │  └────────────────┘ │
                                    └──────────────────────┘
```

### 2.2 Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Frontend Framework** | Vue 3 (Composition API) | Reactive UI components |
| **Build Tool** | Vite | Fast dev server & production builds |
| **2D Mapping** | Leaflet.js + Canvas Heatmap | Interactive geo-visualization |
| **3D Visualization** | Cesium.js | Globe-based digital twin rendering |
| **Backend Framework** | Go + Gin | High-performance REST API |
| **Database** | MySQL (IVDP Schema) | Household & agricultural data |
| **PDF Generation** | gofpdf | Server-side report generation |
| **Data Transport** | JSON | API request/response format |
| **Styling** | CSS (variables, dark/light theme) | Responsive UI with theme support |

### 2.3 Data Flow Architecture

```
User Action (e.g., "Download PDF for filtered villages")
    │
    ▼
Frontend Component (DigitalTwin.vue)
    │
    ├─► User selects filters (District, Taluka, Village)
    ├─► Applies problem filter checkboxes (Multiple AND logic)
    ├─► Clicks "PDF Report"
    │
    ▼
Frontend API Client (api/index.js)
    │
    ├─► POST /pdf/report with:
    │   - selectedDistrictID, talukaID, villageID
    │   - Frontend-rendered chart images (base64 PNG)
    │   - Problem filter metadata
    │
    ▼
Backend PDF Handler (handlers/pdf.go)
    │
    ├─► Parse JSON request
    ├─► Query FAMILY table with:
    │   - WHERE latitude IS NOT NULL AND longitude IS NOT NULL
    │   - AND district_id = ? AND taluka_id = ? AND village_id = ?
    │   - LIMIT 5000
    ├─► Compute statistics (irrigated, no sanitation, etc.)
    ├─► Build PDF document with:
    │   - Header (region metadata)
    │   - Embedded chart images
    │   - Problem filter statistics
    │   - Household-level issue flags
    │
    ▼
MySQL Database
    │
    ├─► Tables: FAMILY (19,832 records)
    ├─► Columns: latitude, longitude, sanitation, lighting,
    │            water_source, land_acres, crops, ration_card
    │
    ▼
Backend Streams PDF
    │
    ▼
Browser Downloads File
    │
    └─► AgriTwin_[District]_[Date].pdf
```

---

## 3. Core Features & Their Purposes

### 3.1 Farmer Registry

**Purpose**: Comprehensive household-level farmer record browsing with advanced filtering and sorting.

**UI Location**: `/farmers` route

**Key Capabilities**:
- Browse all household heads in the database
- Full-text search by first/last name
- Multi-filter system:
  - **Land Ownership**: Filter by "Land Owners" vs "No Land" households
  - **Irrigation Type**: "Irrigated" vs "Rain-fed" water sources
  - **Ration Card Status**: "BPL" (Below Poverty Line), "APL" (Above Poverty Line), or "None"
- **Sortable Columns**: 
  - First Name (alphabetical)
  - Last Name (alphabetical)
  - Total Land (numeric)
  - Water Source (alphabetical)
- **Pagination**: 50 records per page with "Previous/Next" navigation

**Data Displayed**:
```
┌─────┬──────────────┬──────────────┬────────────┬──────────────┬────────────┬───────────┐
│ #   │ First Name   │ Last Name    │ Land (ac)  │ Water Source │ Ration     │ Owns Land │
├─────┼──────────────┼──────────────┼────────────┼──────────────┼────────────┼───────────┤
│ 1   │ Ramesh       │ Kumar        │ 2.5 ac    │ Irrigated    │ BPL        │ Yes       │
│ 2   │ Priya        │ Sharma       │ 0.8 ac    │ Rain-fed     │ APL        │ Yes       │
│ 3   │ Ajay         │ Patel        │ —         │ —            │ None       │ No        │
└─────┴──────────────┴──────────────┴────────────┴──────────────┴────────────┴───────────┘
```

**Backend Implementation**:
- **Endpoint**: `GET /farmers`
- **Handler**: `FarmerHandler.GetFarmers()` in `handlers/farmers.go`
- **Query**: Joins FAMILY table with FAMILY_MEMBER table
- **Returns**: Array of ~10,000 farmer records with no pagination (all records at once)

**Frontend Filtering Logic** (computed in Vue):
```javascript
// Filters applied sequentially (AND logic)
1. Search text: case-insensitive substring match on firstName or lastName
2. Land filter: ownAgricultureLand === "Yes" or "No"
3. Irrigation filter: matches "irrigated" or "rain-fed" strings
4. Ration filter: matches "BPL", "APL", or "None"
5. Sort: by selected column in asc/desc order
6. Paginate: slice 50 records per page
```

---

### 3.2 2D Geo-Intelligence Map

**Purpose**: Interactive 2D map visualization of household geospatial data with village-level clustering and heatmap analysis.

**UI Location**: `/map` route

**Key Capabilities**:
- **Mapping Library**: Leaflet.js with OpenStreetMap/satellite tiles
- **View Modes**:
  - **Points Mode**: Display individual households as small dots
  - **Villages Mode**: Cluster households by village/GP boundary
  
- **Hierarchical Filtering**:
  - Select District → updates Taluka options
  - Select Taluka → updates Village options
  - Select Village → filters to specific cluster
  - "Apply" and "Reset" buttons for state management

- **Coloring Strategies** (in Points mode):
  - **Sanitation**: Green (has toilet) / Red (no sanitation)
  - **Crops/Season**: Color-coded by Kharif or Rabi cultivation
  - **Land Holdings**: Gradient based on total acres
  
- **Village Coverage Heatmap** (in Villages mode):
  - **Green**: High data coverage (>80% households with geo coords)
  - **Orange**: Medium coverage (50-80%)
  - **Red**: Low coverage (<50%)

- **Analytics Cards**: Interactive donut charts showing:
  - Sanitation coverage percentages
  - Irrigation distribution (irrigated vs rain-fed)
  - Land holding categories (marginal, small, medium, large)
  - BPL/APL distribution

- **Detail Panels**:
  - Click household dot → right panel slides in with:
    - Household head name, village, taluka
    - Total/cultivated land, crops
    - Water source, sanitation, lighting, ration card
    - Geo coordinates
  - Click village cluster → aggregated stats:
    - No. of households without sanitation
    - No. without electricity
    - No. without irrigation
    - % BPL families

**Backend Implementation**:
- **Endpoint**: `GET /houses?page=1&limit=500&district_id=1&taluka_id=2&village_id=3`
- **Handler**: `HouseHandler.GetHouses()` in `handlers/houses.go`
- **Pagination**: Default 500 records per page (max 2000)
- **Filtering**: AND logic on district, taluka, village, irrigation, own_land
- **Returns**: 
  ```json
  {
    "data": [
      {
        "familyId": 101,
        "districtName": "Satara",
        "talukaName": "Karad",
        "villageName": "Karadpeth",
        "latitude": 17.3045,
        "longitude": 73.8814,
        "totalLand": "2.5",
        "waterSource": "Irrigated",
        "latrine": "Pit Latrine",
        "lighting": "Electricity",
        "rationCard": "APL"
      }
    ],
    "total": 5234,
    "page": 1,
    "limit": 500
  }
  ```

**Frontend Heatmap Implementation** (Canvas-based Gaussian Blur):
```javascript
// Render to off-screen canvas
1. Create canvas at 2x resolution (e.g., 1280x800 → 2560x1600)
2. For each household:
   - Convert lat/lng to pixel coords using map projection
   - Draw Gaussian blur circle at that point
   - Alpha/intensity based on selected metric (sanitation, crops, etc.)
3. Apply canvas.filter = "blur(X px)" for smooth heatmap
4. Convert canvas to image and overlay on Leaflet map
5. Use gradient colormap (red→yellow→green)
```

---

### 3.3 3D Digital Twin Visualization

**Purpose**: Immersive 3D terrain visualization with household data overlaid as 3D building models with problem highlighting.

**UI Location**: `/twin` route

**Technologies**:
- **Cesium.js**: Web-based 3D geospatial globe
- **3D Models**: Procedurally generated rectangular building shapes
- **Coloring**: Roof color matches condition metric (sanitation, irrigation, etc.)
- **Terrain**: High-resolution elevation data for realistic landscape

**Key Capabilities**:
- **Hierarchical Filtering**: Same as 2D map (District → Taluka → Village)
- **Color-by Modes**:
  - **Sanitation**: Green (has latrine) / Red (no latrine) roofs
  - **Irrigation**: Green (irrigated) / Red (rain-fed)
  - **Lighting**: Green (electricity) / Amber (kerosene/none)
  - **Crops/Season**: Color-coded Kharif vs Rabi crops
  - **Land Holdings**: Gradient from marginal to large farms
  - **Ration Card**: BPL (red) / APL (green) / None (gray)

- **Problem Filter Panel** (Left Sidebar):
  - Checkboxes for multiple problem criteria (AND logic):
    - "No Sanitation" (red roofs)
    - "No Irrigation" (red roofs)
    - "No Electricity" (amber roofs)
    - "BPL Households" (red roofs)
    - "Small Farm, No Irrigation" (specialized)
  - Real-time count of matching households
  - Problem details drawer (click to expand):
    - Root cause explanation
    - Recommended solution
    - Applicable government scheme
  - Visual highlighting: matched houses glow brighter on map

- **Legend Panel**:
  - Mini house icons showing roof color for each condition
  - "Roof color = [metric] status" explanatory text
  
- **Statistics Bar** (Top):
  - Total households visible / filtered total
  - Farmer count in selection
  - Current zoom level (District / Taluka / Village)
  - Filter status indicator

- **Hover Tooltip**:
  - Appears on mouse hover over buildings
  - Shows household head name, village, taluka
  - Quick status grid: Irrigation, Ration, Land, Power
  - Crops summary
  - "Click for full details" hint

- **Detail Panel** (Right Sidebar):
  - Triggered by clicking a building
  - Full household information
  - "Zoom to Location" button (fly camera to house)
  - Agricultural section: land, crops, irrigation
  - Infrastructure section: sanitation, electricity, ration card
  - Farm Advisory section: highlighted issues with schemes

- **Tile Style Toggle**:
  - Toggle between street view (OpenStreetMap) and satellite imagery
  - Updates map tiles in real-time

- **PDF Report Generation**:
  - Download PDF for currently filtered households
  - Includes frontend-rendered donut charts
  - Problem filter statistics embedded
  - File naming: `AgriTwin_[District]_[Date].pdf`

**Frontend Data Aggregation**:
```javascript
// Compute statistics for sidebar from filtered houses
stats = {
  farmers: filteredHouses.filter(h => h.ownLand === "Yes").length,
  noSanitation: filteredHouses.filter(h => !hasLatrine(h)).length,
  noIrrigation: filteredHouses.filter(h => isRainFed(h)).length,
  noElectricity: filteredHouses.filter(h => !hasElectricity(h)).length,
  bpl: filteredHouses.filter(h => isBPL(h)).length,
}
```

**Cesium 3D Rendering Logic**:
```javascript
// For each visible household (up to 2000)
1. Get latitude, longitude (from database)
2. Create rectangular prism (building shape)
3. Set height based on land area (scale: 1 acre = 10m)
4. Set roof color based on condition metric:
   - Green: Good condition
   - Red: Problem detected
   - Amber: Warning/partial
5. Add outline glow if selected by problem filter
6. Attach click handler → show detail panel
7. Attach hover handler → show tooltip
```

**Backend Support**:
- **Endpoint**: `GET /houses?district_id=...&taluka_id=...&village_id=...`
- Same backend as 2D map, but filtered by geospatial hierarchy
- Returns up to 2000 household records at a time

---

### 3.4 PDF Report Generation

**Purpose**: Create comprehensive regional reports with aggregated household data, problem statistics, and visual charts.

**Triggering**: 
- Button in 3D Digital Twin interface: "Download PDF Report"
- Available when ≥1 household is in filtered selection

**Request Flow**:
```
Frontend:
  1. Render donut charts to canvas
  2. Convert charts to base64 PNG images
  3. Capture problem filter UI state (active filters + counts)
  4. POST /pdf/report with:
     {
       "districtId": "1",
       "talukaId": "2",
       "villageId": "3",
       "charts": [
         {
           "title": "Sanitation Coverage",
           "image": "iVBORw0KGgoAAAANS...",
           "segments": [
             {"label": "With Toilet", "pct": 75, "color": "#16a34a"},
             {"label": "No Latrine", "pct": 25, "color": "#ef4444"}
           ]
         }
       ],
       "problemFilters": [
         {"key": "noToilet", "label": "No Sanitation", "count": 123, "active": true},
         {"key": "noIrrigation", "label": "No Irrigation", "count": 456, "active": false}
       ],
       "problemMatchTotal": 123
     }

Backend (pdf.go):
  1. Parse JSON request
  2. Query FAMILY with WHERE filters (up to 5000 households)
  3. Compute statistics:
     - Total households
     - Irrigated count
     - No latrine count
     - No electricity count
     - BPL count
     - Crop combinations
  4. Build PDF with gofpdf:
     - Page 1: Header (region name, date, summary stats)
     - Page 2: Embedded donut charts
     - Page 3+: Tabular data (up to 50 rows per page)
     - Final: Problem summary with recommended schemes
  5. Stream PDF as binary to client
  6. Set Content-Disposition header for download
```

**PDF Document Structure**:
1. **Title Page**:
   - Region hierarchy: "District > Taluka > Village"
   - Generation date
   - AgriTwin branding

2. **Executive Summary**:
   - Total households analyzed
   - Key metrics (% with sanitation, % irrigated, % BPL)
   - Critical issue count

3. **Visual Analytics** (Embedded Charts):
   - Donut charts from frontend (sanitation, irrigation, land, ration cards)
   - Color-coded segments with legend
   - Percentage labels

4. **Problem Filter Summary**:
   - Checkbox summary showing which filters were active
   - Count of households matching combined criteria
   - Estimated population affected

5. **Household Table** (Sample):
   ```
   ID  | Head Name      | Land (ac) | Water    | Sanitation | Lighting    | Ration
   ─────────────────────────────────────────────────────────────────────────────────
   101 | Ramesh Kumar   | 2.5       | Irrigated| Pit Latrine| Electricity | APL
   102 | Priya Sharma   | 0.8       | Rain-fed | No Latrine | Kerosene    | BPL
   103 | Ajay Patel     | 0         | None     | —          | None        | None
   ```

6. **Recommendations** (Scheme Mapping):
   - Issues detected → Suggested government schemes
   - Example: "No Sanitation" → "Swachh Bharat Mission" + "PMAY-G"

**Backend Statistics Computation** (`computePDFStats` in `handlers/pdf.go`):
```go
type pdfStats struct {
  Total         int  // Total households in query
  Irrigated     int  // Has non-rain-fed water source
  NoLatrine     int  // Missing or "No Latrine"
  NoElec        int  // Missing electricity or kerosene only
  BPL           int  // Ration card contains "BPL" or "Antyodaya"
  Both          int  // Growing both Kharif and Rabi
  KharifOnly    int  // Kharif only
  RabiOnly      int  // Rabi only
  NoCrop        int  // Neither
  Marginal      int  // <1 acre
  Small         int  // 1-2.5 acres
  MedLarge      int  // >2.5 acres
}

// Logic examples:
if waterSource.Lower() contains "rain" or "none":
  stats.Irrigated++
if latrine.Lower() == "" or contains "no latrine":
  stats.NoLatrine++
if rationCard contains "BPL" or "Antyodaya":
  stats.BPL++
```

**File Naming Convention**:
```
AgriTwin_[DistrictName]_[TalukaName]_[YYYY-MM-DD].pdf
Example: AgriTwin_Satara_Karad_2024-02-15.pdf
```

---

### 3.5 Detail Panels (Household Information)

**Purpose**: Display comprehensive household information when a single record is selected.

**Available In**:
- 2D Map view (click household dot)
- 3D Digital Twin (click building or hover)

**Information Displayed**:

| Section | Field | Example |
|---------|-------|---------|
| **Header** | Condition Badge | "Fair" / "Critical" (color-coded) |
| | Household Head Name | "Ramesh Kumar" |
| | Family ID | "ID 101" |
| | Village + Taluka | "Karadpeth · Karad" |
| **Agriculture** | Total Land | "2.5 ac" |
| | Cultivated Land | "2.0 ac" |
| | Kharif Crop | "Sugarcane" |
| | Rabi Crop | "Wheat" |
| | Irrigation Source | "Well (Irrigated)" or "Rain-fed" |
| **Infrastructure** | Sanitation/Latrine | "Pit Latrine" / "No Latrine" |
| | Electricity/Lighting | "Electricity" / "Kerosene" / "None" |
| | Ration Card Type | "BPL" / "APL" / "None" |
| **Location** | Coordinates | "17.3045°, 73.8814°" |
| **Farm Advisory** | Issues Detected | List of problems with solutions |

**Condition Badge Logic**:
```javascript
function getConditionLabel(house) {
  // Based on aggregated risk factors
  if (isRainFed(house) && !hasLatrine(house) && isBPL(house)) return "Critical"
  if (isRainFed(house) || !hasLatrine(house) || !hasElectricity(house)) return "At-Risk"
  return "Fair" or "Good"
}

function getConditionColor(house) {
  // Maps to CSS color variable
  if (condition === "Critical") return "#ef4444" (red)
  if (condition === "At-Risk") return "#f59e0b" (amber)
  return "#16a34a" (green)
}
```

---

## 4. Data Flow & Query Patterns

### 4.1 Household Data Query Flow

```
Frontend Request:
  GET /houses?page=1&limit=500&district_id=1&taluka_id=2&village_id=3

Go Handler (houses.go):
  ├─► Parse query params: page, limit, district_id, taluka_id, village_id
  ├─► Validate: limit between 1 and 2000
  ├─► Calculate: offset = (page - 1) * limit
  ├─► Build WHERE clause:
  │   WHERE latitude IS NOT NULL
  │    AND longitude IS NOT NULL
  │    AND latitude != 0 AND longitude != 0
  │    AND (district_id = ? if provided)
  │    AND (taluka_id = ? if provided)
  │    AND (village_id = ? if provided)
  ├─► Execute SELECT query with 18 columns:
  │   FAMILY_ID, DISTRICT_ID, DISTRICT_NAME,
  │   TALUKA_ID, TALUKA_NAME, VILLAGE_ID, VILLAGE_NAME,
  │   LATITUDE, LONGITUDE,
  │   TOTAL_LAND, CULTIVATED_LAND, OWN_LAND,
  │   WATER_SOURCE, KHARIF_CROP, RABI_CROP,
  │   LATRINE, LIGHTING, RATION_CARD, HEAD_NAME
  ├─► Scan rows into HouseRecord structs
  ├─► Query total count (for pagination)
  └─► Return JSON response

Frontend (Vue Component):
  ├─► Receive JSON response
  ├─► Update reactive array: houses.value = response.data
  ├─► Render map markers at each lat/lng
  ├─► Show pagination controls
  └─► Cache for filter operations
```

### 4.2 Filtered Query Pattern (AND Logic)

**Example**: Find BPL households with no sanitation in a specific taluka

```sql
SELECT * FROM FAMILY f
WHERE f.TALUKA_ID = 2
  AND f.RATION_CARD_TYPE LIKE '%BPL%'
  AND (f.SANITATION_TOILET_FACILITY IS NULL 
       OR f.SANITATION_TOILET_FACILITY = 'No Latrine')
  AND f.LATITUDE IS NOT NULL AND f.LONGITUDE IS NOT NULL
LIMIT 5000
```

**Frontend Implementation** (Farmer Registry):
```javascript
// Computed property applies filters sequentially
filteredFarmers = farmers
  .filter(f => !search.value || matchesSearchText(f, search.value))
  .filter(f => !activeFilter.value || f.ownAgricultureLand === activeFilter.value)
  .filter(f => !irrigationFilter.value || matchesIrrigation(f, irrigationFilter.value))
  .filter(f => !rationFilter.value || matchesRation(f, rationFilter.value))
  .sort((a, b) => compareByField(a, b, sortKey.value, sortDir.value))

// Then paginate
paginatedFarmers = filteredFarmers.slice(offset, offset + pageSize)
```

### 4.3 Location Options Query

**Purpose**: Populate dropdown menus with hierarchical location data.

```
GET /location-options?district_id=1&taluka_id=2

Query for Talukas (if district_id provided):
  SELECT DISTINCT taluka_id, taluka_name FROM taluka_master
  WHERE district_id = ?
  ORDER BY taluka_name

Query for Villages (if taluka_id provided):
  SELECT DISTINCT village_id, village_name FROM village_master
  WHERE taluka_id = ?
  ORDER BY village_name

Response:
  {
    "districts": [
      {"id": 1, "name": "Satara"},
      {"id": 2, "name": "Sangli"}
    ],
    "talukas": [
      {"id": 1, "name": "Karad"},
      {"id": 2, "name": "Phaltan"}
    ],
    "villages": [
      {"id": 1, "name": "Karadpeth"},
      {"id": 2, "name": "Karadi"}
    ]
  }
```

---

## 5. Key Technical Implementations

### 5.1 Heatmap Visualization (Canvas-Based Gaussian Blur)

**Algorithm** (MapView.vue):
```javascript
function renderHeatmap(houses, colorMode) {
  // 1. Create off-screen canvas at 2x resolution
  const canvas = document.createElement('canvas')
  const ctx = canvas.getContext('2d')
  canvas.width = mapWidth * 2    // e.g., 2560px
  canvas.height = mapHeight * 2  // e.g., 1600px
  
  // 2. Clear with transparent background
  ctx.fillStyle = 'rgba(0,0,0,0)'
  ctx.fillRect(0, 0, canvas.width, canvas.height)
  
  // 3. For each household, draw Gaussian blob
  for (const house of houses) {
    const [px, py] = mapProject(house.latitude, house.longitude)
    const pixelX = px * 2, pixelY = py * 2
    
    // Color based on mode (sanitation, crops, irrigation, etc.)
    const color = getHeatmapColor(house, colorMode)
    const alpha = 0.6  // Intensity
    
    // Draw Gaussian circle (larger blur = hotter)
    ctx.fillStyle = color
    ctx.globalAlpha = alpha
    ctx.beginPath()
    ctx.arc(pixelX, pixelY, 20, 0, Math.PI * 2)  // 20px radius
    ctx.fill()
  }
  
  // 4. Apply Gaussian blur filter
  ctx.filter = 'blur(15px)'
  
  // 5. Convert to ImageData and create overlay
  const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height)
  
  // 6. Apply colormap gradient (red → yellow → green)
  applyColormapGradient(imageData)
  
  // 7. Create Leaflet image overlay
  const url = canvas.toDataURL('image/png')
  L.imageOverlay(url, bounds).addTo(map)
}

function getHeatmapColor(house, mode) {
  if (mode === 'sanitation') {
    return hasLatrine(house) ? '#16a34a' : '#ef4444'
  } else if (mode === 'irrigation') {
    return isIrrigated(house) ? '#16a34a' : '#ef4444'
  }
  // ... etc for crops, land, ration
}
```

**Performance Optimization**:
- Render to off-screen canvas (no immediate DOM reflow)
- Use 2x resolution for crisp rendering on high-DPI displays
- Limit to visible area using map bounds
- Cache canvas when filters don't change
- Debounce zoom/pan events to avoid excessive redraws

---

### 5.2 Problem Filtering System (AND Logic)

**Frontend Problem Filter** (DigitalTwin.vue):
```javascript
const PROBLEM_FILTER_META = [
  {
    key: 'noSanitation',
    label: 'No Sanitation',
    color: '#ef4444',
    check: (h) => !hasLatrine(h),
    cause: 'Open defecation due to lack of toilet facilities',
    solution: 'Build household latrine with PMAY-G and SBM support',
    scheme: 'Pradhan Mantri Awas Yojana - Gramin'
  },
  {
    key: 'noIrrigation',
    label: 'No Irrigation',
    color: '#ef4444',
    check: (h) => isRainFed(h),
    cause: 'Dependency on rainfall; limited water access',
    solution: 'Install bore well or drip irrigation with PMKSY support',
    scheme: 'Pradhan Mantri Krishi Sinchayee Yojana'
  },
  {
    key: 'noElectricity',
    label: 'No Electricity',
    color: '#f59e0b',
    check: (h) => !hasElectricity(h),
    cause: 'Limited grid connectivity in remote areas',
    solution: 'Apply for solar home systems or grid extension',
    scheme: 'KUSUM Scheme / PM-KUSUM'
  },
  {
    key: 'bpl',
    label: 'BPL Households',
    color: '#ef4444',
    check: (h) => isBPL(h),
    cause: 'Below poverty line status',
    solution: 'Enroll in public distribution and livelihood schemes',
    scheme: 'Public Distribution System (PDS)'
  }
]

// Compute matching count (AND logic)
function computeProblemMatches() {
  const activeChecks = activeProblemFilters.map(key => 
    PROBLEM_FILTER_META.find(p => p.key === key).check
  )
  
  const matchingHouses = filteredHouses.filter(house =>
    activeChecks.every(check => check(house))  // ALL checks must pass
  )
  
  return matchingHouses.length
}
```

**Highlight on Map**:
```javascript
// When problem filters change, re-color buildings
function updateProblemHighlight() {
  for (const building of cesiumBuildings) {
    const house = houseLookup[building.id]
    const hasMatch = activeProblemFilters.every(key => {
      const check = PROBLEM_FILTER_META.find(p => p.key === key).check
      return check(house)
    })
    
    // Glow/brighten matching buildings
    building.material.color.brightness = hasMatch ? 1.5 : 1.0
    building.outlineColor = hasMatch ? '#ffff00' : 'rgba(0,0,0,0)'
  }
}
```

---

### 5.3 Pagination & Lazy Loading

**Farmer Registry** (client-side pagination):
```javascript
// All farmers loaded at once
const farmers = ref([])
const pageSize = 50

// Computed pagination
const totalPages = computed(() => 
  Math.ceil(filteredFarmers.value.length / pageSize)
)

const paginatedFarmers = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredFarmers.value.slice(start, start + pageSize)
})

// Reset to page 1 when filters change
watch([search, activeFilter, irrigationFilter, rationFilter], () => {
  currentPage.value = 1
})
```

**2D/3D Map** (server-side pagination with limit):
```
Frontend:
  const houses = await getHouses({
    page: 1,
    limit: 500,
    district_id: selectedDistrict.value,
    taluka_id: selectedTaluka.value,
    village_id: selectedVillage.value
  })

Backend:
  offset := (page - 1) * limit  // (1-1) * 500 = 0
  query := SELECT ... FROM FAMILY ... LIMIT 500 OFFSET 0
  
Response:
  {
    "data": [...500 houses...],
    "total": 19832,
    "page": 1,
    "limit": 500
  }

Frontend:
  totalPages = Math.ceil(19832 / 500) = 40 pages
  user can navigate with nextPage = (page < totalPages) -> page++
```

**Max Records Protection**:
```go
// In handlers/houses.go
if limit < 1 || limit > 2000 {
  limit = 500
}
```

---

## 6. Database Schema Overview

### 6.1 Core Tables

**FAMILY** (Main household table - 19,832 records)
```sql
CREATE TABLE FAMILY (
  FAMILY_ID INT PRIMARY KEY,
  
  -- Geolocation
  LATITUDE DECIMAL(10,8),
  LONGITUDE DECIMAL(11,8),
  
  -- Hierarchy
  DISTRICT_ID INT,
  TALUKA_ID INT,
  VILLAGE_ID INT,
  
  -- Demographics
  FIRST_NAME_HOUSEHOLD_HEAD VARCHAR(100),
  MIDDLE_NAME_HOUSEHOLD_HEAD VARCHAR(100),
  LAST_NAME_HOUSEHOLD_HEAD VARCHAR(100),
  
  -- Agricultural Data
  OWN_AGRICULTURE_LAND ENUM('Yes', 'No'),
  AREA_AGRICULTURE_LAND_ACRES DECIMAL(8,2),
  LAND_UNDER_CULTIVATION_ACRES DECIMAL(8,2),
  
  -- Irrigation & Crops
  SOURCE_WATER_IRRIGATION VARCHAR(100),  -- 'Irrigated', 'Rain Fed', 'None'
  CULTIVATING_DURING_KHARIF_SEASON VARCHAR(100),  -- Crop name
  TAKING_CROPS_RABI_SEASON VARCHAR(100),  -- Crop name
  
  -- Infrastructure
  SANITATION_TOILET_FACILITY VARCHAR(100),  -- 'Pit Latrine', 'No Latrine', etc.
  ELECTRICITY_CONNECTION VARCHAR(100),  -- 'Electricity', 'Kerosene', 'None'
  
  -- Welfare
  RATION_CARD_TYPE VARCHAR(50),  -- 'BPL', 'APL', 'Antyodaya', 'None'
  
  -- Alternative column names (schema-aware)
  TYPE_OF_LATRINE VARCHAR(100),  -- Alternative to SANITATION_TOILET_FACILITY
  SOURCE_OF_LIGHTING VARCHAR(100),  -- Alternative to ELECTRICITY_CONNECTION
  TYPE_OF_RATION_CARD VARCHAR(50),  -- Alternative to RATION_CARD_TYPE
  
  INDEX idx_district (DISTRICT_ID),
  INDEX idx_taluka (TALUKA_ID),
  INDEX idx_village (VILLAGE_ID),
  INDEX idx_geo (LATITUDE, LONGITUDE),
  INDEX idx_ration (RATION_CARD_TYPE)
)
```

**FAMILY_MEMBER** (Household members - ~50k+ records)
```sql
CREATE TABLE FAMILY_MEMBER (
  MEMBER_ID INT PRIMARY KEY,
  EXTERNAL_FAMILY_ID INT,
  FIRST_NAME VARCHAR(100),
  LAST_NAME VARCHAR(100),
  -- ... other fields
  INDEX idx_family (EXTERNAL_FAMILY_ID)
)
```

**Master Tables** (Reference data)
```sql
CREATE TABLE district_master (
  pklDistrictId INT PRIMARY KEY,
  vsDistrictName VARCHAR(100),
  vsDisplayName VARCHAR(100)
)

CREATE TABLE taluka_master (
  pklTalukaId INT PRIMARY KEY,
  pklDistrictId INT,
  vsTalukaName VARCHAR(100),
  vsDisplayName VARCHAR(100),
  INDEX idx_district (pklDistrictId)
)

CREATE TABLE village_master (
  pklVillageId INT PRIMARY KEY,
  pklTalukaId INT,
  pklDistrictId INT,
  vsVillageName VARCHAR(100),
  vsDisplayName VARCHAR(100),
  INDEX idx_taluka (pklTalukaId),
  INDEX idx_district (pklDistrictId)
)
```

### 6.2 Column Detection Strategy

**ColumnChecker** (handlers/columns.go):
```go
type ColumnChecker struct {
  LatCol string  // Auto-detected: "LATITUDE" or alternative
  LngCol string  // Auto-detected: "LONGITUDE" or alternative
  hasColumn map[string]bool  // Tracks optional columns
}

func (cc *ColumnChecker) Has(col string) bool {
  return cc.hasColumn[col]
}

// On startup, query INFORMATION_SCHEMA to find available columns
// This allows flexibility across different database schema versions
```

**Adaptive Queries**:
```go
// Example from houses.go
latCol := h.CC.LatCol
if latCol == "" {
  latCol = "LATITUDE"  // Fallback
}

// Use in WHERE clause dynamically
query := fmt.Sprintf(
  "WHERE f.%s IS NOT NULL AND f.%s IS NOT NULL",
  latCol, lngCol,
)
```

---

## 7. API Endpoints Reference

### 7.1 Household Data Endpoints

| Method | Endpoint | Purpose | Returns |
|--------|----------|---------|---------|
| GET | `/houses` | List households with pagination & filters | `{data: [...], total: int, page: int, limit: int}` |
| GET | `/house/:id` | Get single household + family members | `{...HouseRecord, members: [...MemberRecord]}` |

**Query Parameters** (GET /houses):
- `page` (default: 1)
- `limit` (default: 500, max: 2000)
- `district_id` (filter by district)
- `taluka_id` (filter by taluka)
- `village_id` (filter by village)
- `irrigation` (filter: "Irrigated" or "Rain Fed")
- `own_land` (filter: "Yes" or "No")

### 7.2 Farmer Registry Endpoint

| Method | Endpoint | Purpose | Returns |
|--------|----------|---------|---------|
| GET | `/farmers` | List all farmer records | `[{firstName, lastName, totalLand, waterSource, rationCard, ...}]` |

**Note**: Returns all records (no pagination) - frontend handles filtering and pagination

### 7.3 Insights Endpoints

| Method | Endpoint | Purpose | Returns |
|--------|----------|---------|---------|
| GET | `/insights/governance` | Government metrics (sanitation, lighting, geo coverage) | `{totalHouseholds, householdsWithGeoData, householdsWithoutToilet, latrineDistribution, ...}` |
| GET | `/insights/agriculture` | Agricultural metrics (land distribution, irrigation, crops) | `{totalFarmers, farmersWithoutIrrigation, landDistribution, waterSourceDistribution, ...}` |
| GET | `/insights/welfare` | Welfare metrics (BPL, ration cards, vulnerable groups) | `{bplHouseholds, rationCardDistribution, bplWithoutToilet, smallFarmersWithoutIrrigation, ...}` |

### 7.4 Location Options Endpoint

| Method | Endpoint | Purpose | Returns |
|--------|----------|---------|---------|
| GET | `/location-options` | Get hierarchical location data | `{districts: [...], talukas: [...], villages: [...]}` |

**Query Parameters**:
- `district_id` (returns talukas for this district)
- `taluka_id` (returns villages for this taluka)

### 7.5 PDF Report Endpoint

| Method | Endpoint | Purpose | Returns |
|--------|----------|---------|---------|
| POST | `/pdf/report` | Generate customized PDF report | Binary PDF file (attachment download) |

**Request Body**:
```json
{
  "districtId": "1",
  "talukaId": "2",
  "villageId": "3",
  "charts": [
    {
      "title": "Sanitation Coverage",
      "image": "iVBORw0KGgoAAAANS...",
      "segments": [
        {"label": "With Toilet", "pct": 75, "color": "#16a34a"}
      ]
    }
  ],
  "problemFilters": [
    {
      "key": "noToilet",
      "label": "No Sanitation",
      "count": 123,
      "active": true
    }
  ],
  "problemMatchTotal": 123
}
```

**Response Headers**:
```
Content-Type: application/pdf
Content-Disposition: attachment; filename="AgriTwin_Satara_Karad_2024-02-15.pdf"
```

### 7.6 Health Check Endpoint

| Method | Endpoint | Purpose | Returns |
|--------|----------|---------|---------|
| GET | `/ping` | API health status | `{message: "pong", mode: "read-only"}` |

---

## 8. User Workflows & Common Tasks

### 8.1 Workflow: Identify BPL Households Without Sanitation

**Goal**: Find all Below Poverty Line households lacking proper sanitation in a specific taluka to recommend government schemes.

**Steps**:

1. **Navigate to 3D Digital Twin**
   - Click "3D Twin" in sidebar navigation
   - System loads all household data (may take 5-10 seconds)

2. **Filter by Geography**
   - Click District dropdown → Select "Satara"
   - Dropdown updates to show Talukas in Satara
   - Click Taluka dropdown → Select "Karad"
   - Dropdown updates to show Villages in Karad
   - Click "Apply" button
   - System filters visualization to ~500-800 households in Karad taluka

3. **Apply Problem Filters**
   - Left sidebar shows "Problem Filter" section
   - Check "BPL Households" checkbox
   - Check "No Sanitation" checkbox
   - System highlights matching households in red (AND logic: must be BOTH BPL AND no sanitation)
   - Display shows: "45 flagged households"

4. **Download PDF Report**
   - Click "⬇ PDF Report" button (top right)
   - System collects:
     - Frontend renders donut charts to canvas
     - Captures problem filter state (which filters active)
     - Sends POST /pdf/report with region + charts + filters
   - Backend queries FAMILY where:
     - TALUKA_ID = 2 (Karad)
     - RATION_CARD_TYPE LIKE '%BPL%'
     - SANITATION_TOILET_FACILITY = 'No Latrine'
   - Generates PDF with:
     - Title: "AgriTwin Report - Satara > Karad"
     - Summary: "45 BPL households without sanitation identified"
     - Charts: embedded donut charts
     - Table: household-level data
     - Recommendation: "Swachh Bharat Mission (SBM) + PMAY-G scheme recommended"
   - Browser downloads: `AgriTwin_Satara_Karad_2024-02-15.pdf`

5. **Share & Action**
   - Distribute PDF to Taluka-level officials
   - Officials use data to target SBM programs
   - Track implementation with future data updates

---

### 8.2 Workflow: Analyze Irrigation Patterns by Village

**Goal**: Understand which villages are rain-fed vs irrigated to plan water resource allocation.

**Steps**:

1. **Navigate to 2D Geo-Intelligence Map**
   - Click "2D Map" in sidebar
   - System loads Leaflet-based interactive map

2. **Switch to Village View Mode**
   - Top control bar has two buttons: "Households" and "Villages"
   - Click "Villages" button
   - Map clusters households by village
   - Each cluster (circle) is colored by data coverage:
     - Green: High coverage (80%+ have geo data)
     - Orange: Medium coverage (50-80%)
     - Red: Low coverage (<50%)

3. **Apply Irrigation Filter**
   - Dropdown "Color by" → Select "Irrigation"
   - Map recolors to show:
     - Each village cluster shows irrigation status
     - Green clusters: mostly irrigated farms
     - Red clusters: mostly rain-fed
   - Legend updates to show "Irrigated vs Rain-fed"

4. **Click Village Cluster**
   - Click on orange village cluster
   - Right panel opens with aggregated stats:
     - Village name, district, taluka
     - Total households: 234
     - Farmers with irrigation: 156 (67%)
     - Farmers rain-fed only: 78 (33%)
     - Average land holding: 1.8 acres
     - Recommended water scheme: "PMKSY - Micro Irrigation"

5. **Drill Down**
   - Click "View Households" or zoom in on map
   - Switch back to "Households" view mode
   - Map shows individual dots for each farm
   - Can further filter by taluka or view specific family records

---

### 8.3 Workflow: Browse Farmer Registry with Sorting

**Goal**: Sort farmers by land size to identify top 50 large-farm holders for crop insurance program.

**Steps**:

1. **Navigate to Farmer Registry**
   - Click "Farmers" in sidebar
   - Page loads ~10,000 farmer records

2. **Sort by Land**
   - Click "Land (ac)" column header
   - Sorts ascending: 0 → max acres
   - Click again to sort descending: max acres → 0
   - Arrow indicator shows sort direction

3. **Filter by Land Ownership**
   - Click "Land Owners" chip to show only households with OWN_AGRICULTURE_LAND = 'Yes'
   - Filters farmers with agricultural land (excludes landless)

4. **Further Refine**
   - Click "Irrigated" water chip to show only farmers with water sources
   - Excludes rain-fed only farmers
   - Table now shows: "Showing 342 of 10,000 records"

5. **Review Top 50**
   - Top 50 records on current page are farmers with:
     - Land ownership: Yes
     - Irrigation: Yes (not rain-fed)
     - Largest land holdings (sorted descending)
   - Pagination shows: "Page 1 of 7"
   - Could navigate to next page for larger farmers

6. **Export/Action**
   - Manually note top farmers for crop insurance enrollment
   - Or use data for targeted agricultural extension programs

---

### 8.4 Workflow: Identify Small Farms Without Irrigation

**Goal**: Find marginal farmers (< 1 acre) without irrigation to recommend drip irrigation schemes.

**Steps**:

1. **Navigate to 3D Digital Twin**
   - Loads all households with 3D buildings

2. **Open Problem Filter**
   - Left sidebar: "Problem Filter" section
   - Check "No Irrigation" checkbox
   - System highlights red buildings (rain-fed farms)
   - Display: "2,145 flagged households"

3. **Apply Color Mode**
   - Top control: "Color by" dropdown → Select "Land Holdings"
   - Roof colors now indicate land size:
     - Green: Large (>5 acres)
     - Orange: Medium (2.5-5 acres)
     - Red: Small/Marginal (<2.5 acres)

4. **Visual Analysis**
   - Red-roofed buildings = small farms
   - Highlighted outline = no irrigation
   - Intersection = small farms without irrigation (highest priority)
   - Sidebar updates to show breakdown:
     - No Irrigation: 2,145 (11%)
     - Distribution by size:
       - Landless: 234
       - Marginal (0-1 ac): 1,245
       - Small (1-2.5 ac): 666

5. **Generate Targeted Report**
   - Change "Color by" → "Land Holdings"
   - Keep "No Irrigation" problem filter active
   - Click "PDF Report"
   - Backend queries:
     - OWN_AGRICULTURE_LAND = 'Yes'
     - AREA_AGRICULTURE_LAND_ACRES <= 2.5
     - SOURCE_WATER_IRRIGATION = 'Rain Fed' or NULL
   - PDF includes:
     - Chart: Land size distribution
     - Chart: Irrigation status
     - Table: 1,245 marginal farmers needing irrigation
     - Recommendation: "PMKSY - Drip Irrigation (60% subsidy) + NABARD loans"

6. **Implementation**
   - Share report with block-level agriculture officer
   - Target PMKSY subsidies to these 1,245 marginal farmers
   - Follow-up monitoring in next data cycle

---

### 8.5 Workflow: Dashboard Insights Overview

**Goal**: Get a quick executive summary of regional agricultural and welfare status.

**Steps**:

1. **Navigate to Dashboard**
   - Click "Dashboard" in sidebar (default landing page)
   - System loads three insight API calls in parallel:
     - `/insights/governance` (sanitation, lighting, geo coverage)
     - `/insights/agriculture` (land distribution, irrigation, crops)
     - `/insights/welfare` (BPL, ration cards)

2. **Review Key Metrics**
   - **Governance Insights** card:
     - Total households: 19,832
     - With geo data: 18,645 (94%)
     - Without sanitation: 3,456 (17%)
     - Without electricity: 5,123 (26%)
   
   - **Agriculture Insights** card:
     - Total farmers: 12,450
     - Without irrigation: 4,200 (34%)
     - Land distribution:
       - Landless: 3,400
       - Marginal (0-1 ac): 5,200
       - Small (1-2.5 ac): 2,100
       - Medium (2.5-5 ac): 1,200
       - Large (>5 ac): 400
   
   - **Welfare Insights** card:
     - Total BPL households: 6,789
     - BPL without sanitation: 2,345
     - BPL without electricity: 3,102
     - Small farmers without irrigation: 1,890

3. **Drill Down**
   - Click on any insight card to navigate to detailed view
   - Example: Click "Without Sanitation" → Opens 3D Twin with "No Sanitation" filter active
   - Example: Click "Land Distribution" → Opens 2D Map with "Land Holdings" color mode

4. **Share Insights**
   - Screenshot or export key metrics
   - Share with government stakeholder meetings
   - Inform policy decisions on subsidy allocation

---

## 9. Deployment & Configuration

### 9.1 Frontend Setup

**Development**:
```bash
cd frontend
npm install
npm run dev    # Vite dev server on localhost:5173
```

**Production Build**:
```bash
npm run build   # Generates dist/ folder
npm run preview # Test production build locally
```

**Vite Configuration** (vite.config.js):
```javascript
export default {
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
}
```

### 9.2 Backend Setup

**Build**:
```bash
cd DT-backend
go mod download
go build -o agritwin-api .
```

**Run**:
```bash
./agritwin-api
# Starts server on :8081
# Outputs: "[STARTUP] Agriculture Digital Twin backend starting on :8081"
```

**Database Connection** (db/db.go):
```go
dsn := "playground:Pl@Ygr0und@tcp(10.15.20.235:3306)/ivdp_db?parseTime=true"
```
- **Host**: 10.15.20.235:3306 (remote MySQL server)
- **User**: playground
- **Database**: ivdp_db
- **Connection**: TLS not used (local network)

### 9.3 Read-Only Mode Enforcement

**Backend middleware** (main.go):
```go
// Block all non-GET requests EXCEPT POST /pdf/*
r.Use(func(c *gin.Context) {
  method := c.Request.Method
  path := c.Request.URL.Path
  isPDFPost := method == "POST" && strings.HasPrefix(path, "/pdf/")
  if isPDFPost {
    c.Next()
    return
  }
  if method != "GET" && method != "OPTIONS" {
    c.AbortWithStatusJSON(http.StatusMethodNotAllowed, 
      gin.H{"error": "read-only mode — only GET requests permitted"})
    return
  }
  c.Next()
})
```

**CORS Configuration**:
```go
c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
// Allows frontend to make GET and POST /pdf requests
```

---

## 10. Performance Considerations

### 10.1 Frontend Optimization

| Optimization | Implementation |
|--------------|-----------------|
| **Pagination** | Server-side (max 2000 records per request) or client-side (farmer registry: all records, frontend filters) |
| **Lazy Loading** | Maps/3D viewers load data on-demand with spinner |
| **Canvas Caching** | Heatmap canvas rendered once, cleared on filter change |
| **Debouncing** | Zoom/pan events debounced to avoid excessive redraws |
| **Code Splitting** | Vue Router lazy-loads views (Farmers, Map, Twin) |
| **Theme Persistence** | Dark/light mode saved to localStorage |

### 10.2 Backend Query Optimization

| Query Type | Optimization |
|-----------|--------------|
| **Index Usage** | Indexes on LATITUDE, LONGITUDE, FAMILY_ID, DISTRICT_ID, TALUKA_ID, VILLAGE_ID |
| **Column Selection** | SELECT only needed columns (18 for house, 6 for farmer) |
| **Pagination** | LIMIT and OFFSET in household queries |
| **Filtering** | Push filters to database WHERE clause (not frontend post-processing) |
| **Read-Only** | SELECT-only mode — no locks or transaction overhead |
| **Connection Pool** | Go sql.DB auto-manages connection pooling |

### 10.3 API Timeout Handling

```javascript
// frontend/src/api/index.js
const TIMEOUT_DEFAULT = 5000       // 5 seconds for quick queries
const TIMEOUT_DATA = 30000         // 30 seconds for large dataset queries

// Applied to /houses, /farmers, /location-options calls
```

---

## 11. Error Handling & Fallbacks

### 11.1 Frontend Error Handling

| Scenario | Behavior |
|----------|----------|
| **API Timeout** | Abort request, show "Loading..." → "Error: Request timeout" |
| **API 500 Error** | Show error toast with backend error message |
| **Invalid Filters** | Reset to default filters, reload data |
| **Map Load Failure** | Show "Unable to load map. Check your connection." |
| **PDF Generation Fails** | Show "PDF generation failed. Try again." |
| **Missing Data** | Show "—" (em dash) for null/undefined fields |

### 11.2 Backend Error Handling

| Scenario | Response |
|----------|----------|
| **DB Connection Fails** | `[FATAL] Failed to reach DB` → shutdown |
| **Query Error** | HTTP 500 with `{"error": "failed to fetch...", "detail": "..."}` |
| **Invalid Params** | Coerce invalid values (page < 1 → 1, limit > 2000 → 500) |
| **No Results** | Return empty array `[]` with total count = 0 |
| **PDF Rendering Fails** | HTTP 500 with error detail |

---

## 12. Future Enhancements & Roadmap

### Potential Features

1. **User Authentication & Role-Based Access**
   - Admin: Full system access
   - Block Officer: Filtered by block/taluka
   - Gram Panchayat: Filtered by village
   - Read-only mode for public portal

2. **Data Refresh & Sync**
   - Automated periodic data sync from source IVDP database
   - Incremental updates (only changed records)
   - Version tracking (data as of: 2024-02-15)

3. **Advanced Analytics**
   - Trend analysis (YoY comparison of metrics)
   - Predictive modeling (identify at-risk areas before crisis)
   - Correlation analysis (e.g., "BPL households with low irrigation")

4. **Customizable Reports**
   - User-defined report templates
   - Scheduled report generation (weekly/monthly)
   - Email delivery of reports

5. **Mobile App**
   - React Native version for field officers
   - Offline mode with local data sync
   - QR code scanning for household lookup

6. **Real-time Data Updates**
   - WebSocket push for live data changes
   - Activity feed showing recent updates
   - Collaborative editing & annotations

7. **Extended Spatial Analysis**
   - Accessibility analysis (distance to nearest market/hospital)
   - Weather correlation (rainfall vs crop outcomes)
   - Livestock & animal husbandry data integration

---

## 13. Glossary & Key Terms

| Term | Definition |
|------|-----------|
| **Kharif** | Monsoon season (June-October) planting cycle |
| **Rabi** | Winter season (October-March) planting cycle |
| **BPL** | Below Poverty Line; households with income below state poverty threshold |
| **APL** | Above Poverty Line; households with income above state poverty threshold |
| **Antyodaya** | Poorest of the poor; targeted ration card category |
| **Irrigated** | Water source includes well, bore well, canal, or river |
| **Rain-fed** | Dependent on monsoon rainfall; no irrigation infrastructure |
| **Sanitation Facility** | Type of latrine/toilet used (pit, septic, none, etc.) |
| **Electricity Connection** | Main power source (grid electricity, solar, kerosene, none) |
| **Land Holding** | Total agricultural land owned or cultivated (in acres) |
| **PMKSY** | Pradhan Mantri Krishi Sinchayee Yojana (Prime Minister Irrigation Scheme) |
| **PMAY-G** | Pradhan Mantri Awas Yojana - Gramin (Rural Housing Scheme) |
| **SBM** | Swachh Bharat Mission (Clean India Mission) |
| **Digital Twin** | 3D virtual replica of real-world geography and infrastructure |

---

## 14. Support & Troubleshooting

### 14.1 Common Issues

**Q: Map shows no households**
- A: Ensure data has valid latitude/longitude (not NULL or 0)
- Check that filter selections narrow results too much
- Verify database connection is active (check `/ping` endpoint)

**Q: PDF download fails**
- A: Check that at least one household is in the filtered selection
- Ensure backend has write permissions (for gofpdf)
- Check browser console for error details
- Try with a smaller selection (< 1000 households)

**Q: 3D viewer is slow**
- A: Reduce data by filtering to smaller region
- Lower 3D quality settings in browser
- Check browser GPU support (WebGL)
- Try a different browser (Chrome/Firefox recommended)

**Q: Filters not working**
- A: Clear browser cache and reload
- Check that filter values exist in database
- Verify column names if schema was updated
- Check browser console for JavaScript errors

---

## 15. Technical Contact & Support

For deployment, integration, or technical questions:

- **Backend**: Go/Gin API on port 8081, MySQL database integration
- **Frontend**: Vue 3 Composition API with Leaflet & Cesium.js
- **Database**: MySQL IVDP schema with 19,832 household records
- **PDF Engine**: gofpdf for server-side report generation

---

## Conclusion

AgriTwin provides a comprehensive, user-friendly platform for agricultural data analysis and visualization. By combining geospatial intelligence (2D/3D mapping), advanced filtering, and automated report generation, it enables government agencies and agricultural stakeholders to make data-driven decisions for rural development. The read-only architecture ensures data security while supporting real-time insights into household farming conditions, infrastructure gaps, and welfare needs.

