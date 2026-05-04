# Multi-Select Hierarchical Filters Implementation Guide

## Overview

The Dashboard now supports **hierarchical multi-select filters** with checkboxes for District → Taluka → Village filtering. This guide explains the implementation, API changes, and how to use the new features.

## Features

✅ **Multi-select Checkboxes**: Users can select multiple districts, talukas, and villages  
✅ **Hierarchical Dependencies**: Child filters automatically clear and reload when parent changes  
✅ **Search Functionality**: Each dropdown has a search box to filter options  
✅ **Scrollable Dropdowns**: Options are scrollable with max-height of 280px  
✅ **Smart Display**: Shows count for multiple selections or name for single selection  
✅ **Array-based Backend**: Backend filters using SQL IN clauses  
✅ **Backward Compatible**: Existing API still supports single-value queries  

## Frontend Implementation

### State Management (Vue Composition API)

```javascript
// Array-based state for multi-select
const selectedDistricts = ref([])      // Array of district IDs
const selectedTalukas = ref([])        // Array of taluka IDs
const selectedVillages = ref([])       // Array of village IDs

// UI state
const openDropdown = ref(null)         // 'district' | 'taluka' | 'village' | null
const districtSearchText = ref('')
const talukaSearchText = ref('')
const villageSearchText = ref('')
```

### Computed Properties

#### Display Text
Shows selected values or placeholder:
```javascript
selectedDistrictDisplay // "2 Districts" or "Nashik"
selectedTalukaDisplay   // "3 Talukas" or "Nashik"
selectedVillageDisplay  // "5 Villages" or "Nashik"
```

#### Filtered Options
```javascript
filteredDistricts // Districts matching search text
filteredTalukas   // Talukas based on selected districts + search
filteredVillages  // Villages based on selected talukas + search
```

### Functions

#### Toggle Functions
```javascript
toggleDistrict(districtId)  // Add/remove from selectedDistricts
toggleTaluka(talukaId)      // Add/remove from selectedTalukas
toggleVillage(villageId)    // Add/remove from selectedVillages
toggleDropdown(dropdown)    // Open/close dropdown menu
```

#### Filter Management
```javascript
applyFilters()     // Send selected filters to backend
resetFilters()     // Clear all selections and reload
```

### Watch Functions (Hierarchical Dependencies)

When **selectedDistricts** changes:
- Clears selectedTalukas and selectedVillages
- Clears search text
- Calls loadLocationOptions() with new district IDs

When **selectedTalukas** changes:
- Clears selectedVillages
- Clears village search text
- Calls loadLocationOptions() with new taluka IDs

## Backend API Changes

### 1. Location Options Endpoint
**Endpoint**: `GET /api/location-options`

**Old Parameters** (Single Select):
```
?district_id=1&taluka_id=2&village_id=3
```

**New Parameters** (Multi-Select):
```
?district_ids=1,2,3&taluka_ids=4,5,6&village_ids=7,8,9
```

**Backward Compatibility**: Both formats are supported. Plural (`_ids`) parameters are preferred.

**Response**:
```json
{
  "districts": [
    {"id": "1", "name": "Nashik"},
    {"id": "2", "name": "Pune"}
  ],
  "talukas": [
    {"id": "4", "name": "Nashik Taluka"},
    {"id": "5", "name": "Deola"}
  ],
  "villages": [
    {"id": "7", "name": "Village A"},
    {"id": "8", "name": "Village B"}
  ]
}
```

### 2. Dashboard Summary Endpoint
**Endpoint**: `GET /api/dashboard/summary`

**New Query Format**:
```
?district_ids=1,2&taluka_ids=3,4&village_ids=5,6
```

**Backend Processing**:
The handler now uses SQL IN clauses:
```sql
WHERE DISTRICT_ID IN (1, 2)
  AND TALUKA_ID IN (3, 4)
  AND VILLAGE_ID IN (5, 6)
```

**Response**: Same as before, but filtered by multiple locations

## API Payload Structure

### Apply Filters Request
When user clicks "Apply", the frontend sends:

```javascript
{
  districtIds: [1, 2, 3],        // Array of selected district IDs
  talukaIds: [4, 5],             // Array of selected taluka IDs  
  villageIds: [6, 7, 8]          // Array of selected village IDs
}
```

Converted to query string:
```
?district_ids=1,2,3&taluka_ids=4,5&village_ids=6,7,8
```

## UI/UX Components

### Checkbox Dropdown Structure
```
┌─────────────────────────┐
│ Trigger Button          │ ◀ Shows selected count/name
└────┬────────────────────┘
     │ Click to toggle
     │
     ▼
┌─────────────────────────┐
│ Search Box              │ ◀ Search within options
├─────────────────────────┤
│ ☑ Nashik               │ ◀ Checkboxes for selection
│ ☐ Pune                 │
│ ☐ Nagpur               │
│ ☐ Aurangabad           │
│          (scrollable)   │
└─────────────────────────┘
```

### Styling Classes
- `.filter-dropdown-wrapper`: Container for each filter
- `.filter-dropdown-trigger`: Button to open/close
- `.filter-dropdown-menu`: Dropdown container
- `.filter-dropdown-options`: Scrollable options list
- `.filter-option`: Individual checkbox + label
- `.filter-checkbox`: Checkbox input
- `.filter-option-label`: Option text
- `.filter-search-input`: Search input box

### Disabled State
The Taluka and Village filters are disabled (grayed out) when their parent filter has no selections:
- Taluka is disabled if no districts are selected
- Village is disabled if no talukas are selected

## Data Flow Diagram

```
User Interface
    │
    ├─ Select Districts [1, 2]
    │     │
    │     ▼
    │  Watch: selectedDistricts
    │     │ Clear talukas & villages
    │     │ Load talukas for districts [1, 2]
    │     │
    │     ▼
    │  API: /location-options?district_ids=1,2
    │     │
    │     ▼
    │  Backend: Build WHERE with IN clause
    │     │ WHERE DISTRICT_ID IN (1, 2)
    │     │
    │     ▼
    │  Response: [{id: "3", name: "Taluka A"}, ...]
    │
    ├─ Select Talukas [3, 4]
    │     │
    │     ▼
    │  Watch: selectedTalukas
    │     │ Clear villages
    │     │ Load villages for talukas [3, 4]
    │     │
    │     ▼
    │  API: /location-options?taluka_ids=3,4
    │     │
    │     ▼
    │  Backend: Build WHERE with IN clause
    │     │ WHERE TALUKA_ID IN (3, 4)
    │
    ├─ Select Villages [5, 6, 7]
    │
    ├─ Click "Apply"
    │     │
    │     ▼
    │  buildLocationParams()
    │     │ Returns: {district_ids: "1,2", taluka_ids: "3,4", village_ids: "5,6,7"}
    │     │
    │     ▼
    │  fetchDashboardData(params)
    │     │
    │     ▼
    │  API: /dashboard/summary?district_ids=1,2&taluka_ids=3,4&village_ids=5,6,7
    │     │
    │     ▼
    │  Backend: WHERE DISTRICT_ID IN (1,2) AND TALUKA_ID IN (3,4) AND VILLAGE_ID IN (5,6,7)
    │
    ▼
Dashboard displays filtered data
```

## Backend GO Code Changes

### 1. location_options.go

**Updated Handler**:
- Parses comma-separated IDs from `district_ids`, `taluka_ids` parameters
- Supports both singular (`district_id`) and plural (`district_ids`) for backward compatibility
- Uses SQL `IN` clauses for multiple IDs
- Returns full list of districts, filtered talukas, and filtered villages

**Key Functions**:
```go
func (h *LocationHandler) GetLocationOptions(c *gin.Context)
```

### 2. dashboard_summary.go

**Updated Handler**:
- `GetDashboardSummary()`: Accepts both `district_id` and `district_ids` parameters
- New function: `buildOptionalLocationFilterWithArrays()`: Builds WHERE clauses with IN operators

**SQL Example**:
```sql
SELECT * FROM FAMILY_MEMBER f
WHERE f.DISTRICT_ID IN (1, 2)
  AND f.TALUKA_ID IN (3, 4)
  AND f.VILLAGE_ID IN (5, 6, 7)
```

## Frontend API Functions

### New Helper Functions in api/index.js

```javascript
// Fetch talukas for multiple districts
getTalukasByDistricts(districtIds)  // accepts: [1, 2, 3] or "1,2,3"

// Fetch villages for multiple talukas
getVillagesByTalukas(talukaIds)    // accepts: [3, 4] or "3,4"
```

Both functions use the `/location-options` endpoint with array parameters.

## Usage Example

### Frontend Code
```javascript
// Select multiple districts
selectedDistricts.value = [1, 2, 3]  // Nashik, Pune, Nagpur

// Watch automatically clears talukas/villages and loads new options
// User selects talukas
selectedTalukas.value = [4, 5]

// Watch automatically clears villages and loads new options
// User selects villages
selectedVillages.value = [6, 7, 8]

// Click Apply
applyFilters()  // Sends: ?district_ids=1,2,3&taluka_ids=4,5&village_ids=6,7,8
```

### Backend Query Generation
```go
// When district_ids=1,2,3 is provided:
WHERE f.DISTRICT_ID IN (1, 2, 3)

// When multiple parameters:
WHERE f.DISTRICT_ID IN (1, 2, 3)
  AND f.TALUKA_ID IN (4, 5)
  AND f.VILLAGE_ID IN (6, 7, 8)
```

## Important Rules

⚠️ **Always clear child filters when parent changes**
- Changing districts must clear talukas and villages
- Changing talukas must clear villages
- This is handled automatically by watch() functions

⚠️ **Never show all options by default**
- Talukas are only shown for selected districts
- Villages are only shown for selected talukas
- If no parent is selected, child filter is disabled

⚠️ **Arrays must be comma-separated in URL**
- `district_ids=1,2,3` NOT `district_ids[]=1&district_ids[]=2`
- This is handled by frontend's `buildLocationParams()` function

⚠️ **Backend must use IN clauses**
- Not: `WHERE district_id = ?`
- But: `WHERE district_id IN (?, ?, ?)`
- Implemented in `buildOptionalLocationFilterWithArrays()`

## Testing

### Test Scenarios
1. ✅ Select 1 district → verify talukas load
2. ✅ Select 2 districts → verify talukas from both load
3. ✅ Select districts → change selection → verify old talukas cleared
4. ✅ Select taluka → change districts → verify talukas/villages cleared
5. ✅ Click Apply → verify dashboard data filtered correctly
6. ✅ Click Reset → verify all selections cleared
7. ✅ Search in dropdown → verify filtering works
8. ✅ Disable state → verify can't click when parent not selected

### API Testing
```bash
# Test location options with multiple districts
curl "http://localhost:8080/api/location-options?district_ids=1,2"

# Test dashboard summary with arrays
curl "http://localhost:8080/api/dashboard/summary?district_ids=1,2&taluka_ids=3,4&village_ids=5,6"

# Test backward compatibility (single values)
curl "http://localhost:8080/api/location-options?district_id=1&taluka_id=2"
```

## Migration Notes

- ✅ Old single-select queries still work (backward compatible)
- ✅ Both `district_id` and `district_ids` parameters accepted
- ✅ Existing API endpoints updated in-place (no new endpoints needed)
- ✅ No database schema changes required
- ✅ Frontend uses new multi-select, but backend handles both formats

## Summary

The implementation provides a complete hierarchical multi-select filter system with:
- **Frontend**: Vue.js checkboxes with automatic dependency management
- **Backend**: SQL IN clauses for array-based filtering
- **API**: Support for comma-separated IDs with backward compatibility
- **UX**: Smooth dropdown interactions, search, and smart display

All requirements have been met without external libraries and with proper hierarchical dependency handling.
