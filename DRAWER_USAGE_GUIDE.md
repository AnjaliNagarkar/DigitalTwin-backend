# Context-Aware Drawer - Usage Guide

## Quick Start

### Step 1: Open the 3D Twin Map
Navigate to the Population 3D Twin view in your application.

### Step 2: Click on a Household
Click on any 3D building or point marker on the map. The detail panel will open on the right side.

### Step 3: Use the Filter Dropdown
At the top of the detail panel, you'll see a "View Fields" dropdown.

```
┌─────────────────────────────────────┐
│  Head: John Smith                   │ [×]
│  House 123                          │
├─────────────────────────────────────┤
│ View Fields  [All Fields ▾]        │ ← Click here to filter
├─────────────────────────────────────┤
│ POPULATION                          │
│ ┌──────────────────────────────────┐│
│ │ Total Members      │   5          ││
│ │ Male              │   3          ││
│ └──────────────────────────────────┘│
│ ...
└─────────────────────────────────────┘
```

### Step 4: Select a Filter
Click the dropdown to see available filters:

- **All Fields** (Default) - Shows complete household information
- **BPL & Welfare** - Shows welfare and economic status
- **Education & Students** - Shows education and employment details
- **Disability & Support** - Shows disability and support eligibility

### Step 5: View Filtered Data
Once you select a filter, the detail panel immediately updates to show only relevant fields organized in labeled sections.

---

## Filter Examples

### Example 1: Checking BPL Status
**Use Case:** Verifying welfare eligibility for a household

1. Click on household
2. Select **"BPL & Welfare"** filter
3. See:
   - BPL Status & Ration Card section
   - Economic Indicators section
   - Family Composition section

**Sample Output:**
```
BPL STATUS & RATION CARD
┌─────────────────────────────────────┐
│ BPL Category        │  BPL           │
│ Ration Card Type    │  AAY           │
└─────────────────────────────────────┘

ECONOMIC INDICATORS
┌─────────────────────────────────────┐
│ Annual Income       │  ₹45,000       │
│ Working Members     │  2             │
│ Unemployed Members  │  1             │
└─────────────────────────────────────┘

FAMILY COMPOSITION
┌─────────────────────────────────────┐
│ Total Members       │  5             │
│ Male                │  3             │
│ Female              │  2             │
└─────────────────────────────────────┘
```

### Example 2: Student Education Tracking
**Use Case:** Checking education status and employment context

1. Click on household
2. Select **"Education & Students"** filter
3. See:
   - Education Status section (literacy levels)
   - Employment Context section

**Sample Output:**
```
EDUCATION STATUS
┌─────────────────────────────────────┐
│ Literate Members    │  4             │
│ Illiterate Members  │  1             │
│ Total Members       │  5             │
└─────────────────────────────────────┘

EMPLOYMENT CONTEXT
┌─────────────────────────────────────┐
│ Working Members     │  2             │
│ Unemployed Members  │  1             │
│ Occupations         │  Farmer, Labor │
└─────────────────────────────────────┘
```

### Example 3: Disability Support Eligibility
**Use Case:** Identifying support needs for persons with disability

1. Click on household
2. Select **"Disability & Support"** filter
3. See:
   - Disability Information section
   - Family Context section
   - Economic Support Eligibility section

**Sample Output:**
```
DISABILITY INFORMATION
┌─────────────────────────────────────┐
│ Divyang Members     │  1             │
│ Disability Status   │  Yes           │
└─────────────────────────────────────┘

FAMILY CONTEXT
┌─────────────────────────────────────┐
│ Total Members       │  4             │
│ Male                │  2             │
│ Female              │  2             │
└─────────────────────────────────────┘

ECONOMIC SUPPORT ELIGIBILITY
┌─────────────────────────────────────┐
│ BPL Category        │  BPL           │
│ Annual Income       │  ₹35,000       │
└─────────────────────────────────────┘
```

---

## Field Descriptions

### Population Fields
- **Total Members** - All household members
- **Male** - Number of male members
- **Female** - Number of female members

### Employment Fields
- **Working Members** - Members with active employment
- **Unemployed Members** - Members without employment
- **Occupations** - List of occupations/professions

### Education Fields
- **Literate Members** - Members who can read/write
- **Illiterate Members** - Members who cannot read/write

### Welfare Fields
- **BPL Category** - Below Poverty Line classification (Yes/No)
- **Ration Card Type** - Type of ration card (AAY, BPL, APL, etc.)
- **Annual Income** - Estimated household annual income

### Document Coverage Fields
- **Aadhaar Coverage** - Aadhaar enrollment status (Complete/Partial/Missing)
- **Caste Certificate Coverage** - Caste certificate status

### Disability Fields
- **Divyang Members** - Count of members with disability
- **Disability Status** - Whether household has disability (Yes/No)

---

## Tips & Tricks

### Switching Between Filters
- Filter selection **persists** while the panel is open
- Close and reopen the panel to **reset to "All Fields"**
- You can switch filters multiple times without closing the panel

### Missing Data
- If a field shows **"—"** (dash), that data is not available
- Empty state message appears if no relevant fields exist for the filter

### Data Accuracy
- All data comes from the database
- Fields are formatted consistently
- Calculations are done server-side before display

### Comparing Households
- Click different households while keeping the same filter
- Useful for comparing welfare status, education levels, etc.

---

## Keyboard Shortcuts (Coming Soon)

| Shortcut | Action |
|----------|--------|
| `F` | Focus on next field group |
| `C` | Copy field value |
| `E` | Export filtered data |

---

## Troubleshooting

### Dropdown Not Opening
- Ensure you've clicked a household first
- Panel must be visible and not scrolled up

### Data Shows Dashes (—)
- Field not recorded in database
- This is expected for optional fields

### Filter Selection Lost
- Close and reopen the detail panel to reset
- This is by design to avoid accidental filter persistence

### Performance Issues
- Filtering is instant (computed property cached)
- No performance impact from using filters

---

## Use Cases by Role

### Field Worker / Surveyor
- Use **BPL & Welfare** to verify income status
- Use **Disability & Support** to identify assistance needs
- Use **Education & Students** for school enrollment tracking

### Welfare Officer
- Use **BPL & Welfare** to check ration card eligibility
- Use **Disability & Support** for disability pension processing
- Use **All Fields** for comprehensive household profile

### Education Officer
- Use **Education & Students** to track literacy levels
- Use **All Fields** to see complete family context
- Filter helps identify households needing education support

### Health Worker
- Use **Disability & Support** for health intervention needs
- Use **All Fields** for complete health assessment
- Use **BPL & Welfare** for nutrition program eligibility

---

## Data Privacy & Security

- **No sensitive data is exported** when using filters
- **Filters run locally** - selections not logged
- **Data access** follows existing database permissions
- **No additional tracking** is added by the filter feature

---

## Version Info

- **Implementation Date:** 2026-05-08
- **Component:** Population3DTwin.vue
- **Fields Supported:** 15 key household indicators
- **Filters Available:** 4 filter categories

---

## Questions or Feedback?

For issues or suggestions regarding the context-aware drawer feature, please refer to the implementation documentation at:
`CONTEXT_AWARE_DRAWER_IMPLEMENTATION.md`
