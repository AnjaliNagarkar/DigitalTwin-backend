<template>
  <div class="twin-page">
    <div class="cesium-wrap" ref="cesiumContainer"></div>

    <!-- ── TOP BAR ── -->
    <div class="topbar">
      <div class="topbar-brand">
        <span class="brand-dot"></span>
        <span class="brand-name">AgriTwin</span>
        <span class="brand-sub">3D Digital Twin</span>
      </div>

      <!-- FILTERS: District → Taluka → Village -->
      <div class="filter-bar">

        <!-- District -->
        <div class="filter-group">
          <label class="filter-label">District</label>
          <div class="custom-select" :class="{ open: openDropdown === 'district' }"
               @click.stop="toggleDropdown('district')">
            <button class="cs-trigger" type="button">
              <span class="cs-value">{{ selectedDistrictLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'district'" @click.stop>
              <div class="cs-option" :class="{ selected: !pendingDistrict }"
                   @click="selectDistrict('')">All Districts</div>
              <div class="cs-option" v-for="d in districtOptions" :key="d.id"
                   :class="{ selected: String(pendingDistrict) === String(d.id) }"
                   @click="selectDistrict(d.id)">{{ d.name }}</div>
            </div>
          </div>
        </div>

        <span class="filter-arrow">›</span>

        <!-- Taluka -->
        <div class="filter-group">
          <label class="filter-label">Taluka</label>
          <div class="custom-select" :class="{ open: openDropdown === 'taluka', disabled: !pendingDistrict }"
               @click.stop="pendingDistrict && toggleDropdown('taluka')">
            <button class="cs-trigger" type="button" :disabled="!pendingDistrict">
              <span class="cs-value">{{ selectedTalukaLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'taluka'" @click.stop>
              <div class="cs-option" :class="{ selected: !pendingTaluka }"
                   @click="selectTaluka('')">All Talukas</div>
              <div class="cs-option" v-for="t in talukaOptions" :key="t.id"
                   :class="{ selected: String(pendingTaluka) === String(t.id) }"
                   @click="selectTaluka(t.id)">{{ t.name }}</div>
            </div>
          </div>
        </div>

        <span class="filter-arrow">›</span>

        <!-- Village -->
        <div class="filter-group">
          <label class="filter-label">Village</label>
          <div class="custom-select" :class="{ open: openDropdown === 'village', disabled: !pendingTaluka }"
               @click.stop="pendingTaluka && toggleDropdown('village')">
            <button class="cs-trigger" type="button" :disabled="!pendingTaluka">
              <span class="cs-value">{{ selectedVillageLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'village'" @click.stop>
              <div class="cs-option" :class="{ selected: !pendingVillage }"
                   @click="selectVillage('')">All Villages</div>
              <div class="cs-option" v-for="v in villageOptions" :key="v.id"
                   :class="{ selected: String(pendingVillage) === String(v.id) }"
                   @click="selectVillage(v.id)">{{ v.name }}</div>
            </div>
          </div>
        </div>

        <button class="apply-btn" @click="applyFilters" :disabled="!filtersDirty">
          Apply
        </button>
        <button class="reset-btn" @click="resetFilters" v-if="pendingDistrict || pendingTaluka || pendingVillage || filterDistrict || filterTaluka || filterVillage">
          ✕ Reset
        </button>
      </div>

      <!-- RIGHT CONTROLS -->
      <div class="topbar-right">
        <div class="ctrl-group">
          <label class="filter-label">Color by</label>
          <div class="custom-select cs-align-right" :class="{ open: openDropdown === 'colorMode' }"
               @click.stop="toggleDropdown('colorMode')">
            <button class="cs-trigger" type="button">
              <span class="cs-value">{{ selectedColorModeLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown cs-dropdown-right" v-show="openDropdown === 'colorMode'" @click.stop>
              <div class="cs-option" :class="{ selected: colorMode === 'sanitation' }" @click="selectColorMode('sanitation')">Sanitation</div>
              <div class="cs-option" :class="{ selected: colorMode === 'irrigation' }" @click="selectColorMode('irrigation')">Irrigation</div>
              <div class="cs-option" :class="{ selected: colorMode === 'lighting' }"   @click="selectColorMode('lighting')">Electricity</div>
              <div class="cs-option" :class="{ selected: colorMode === 'crops' }"      @click="selectColorMode('crops')">Crops / Season</div>
              <div class="cs-option" :class="{ selected: colorMode === 'land' }"       @click="selectColorMode('land')">Land Holdings</div>
              <div class="cs-option" :class="{ selected: colorMode === 'ration' }"     @click="selectColorMode('ration')">Ration Card</div>
            </div>
          </div>
        </div>
        <button class="ctrl-btn" :class="{ active: tileStyle === 'satellite' }" @click="toggleTile">
          {{ tileStyle === 'satellite' ? '🛰 Satellite' : '🗺 Street' }}
        </button>

        <!-- DOWNLOAD PDF -->
        <div class="dl-wrap" v-if="!loadingLiveData && filteredHouses.length">
          <button class="dl-btn" @click="downloadPDF" :disabled="pdfLoading"
                  :title="`Download PDF report for ${filteredHouses.length} households`">
            {{ pdfLoading ? '⏳ Generating…' : '⬇ PDF Report' }}
          </button>
          <span class="dl-count">{{ filteredHouses.length.toLocaleString() }} rows</span>
        </div>
      </div>
    </div>

    <!-- LOADING STATE -->
    <div class="loading-overlay" v-if="loadingLiveData">
      <div class="loading-spinner"></div>
      <div class="loading-text">Loading village data…</div>
    </div>

    <!-- STATS BAR -->
    <div class="stats-bar" v-if="!loadingLiveData">
      <span class="stat-item">
        <span class="stat-dot" style="background:#16a34a"></span>
        <strong>{{ filteredHouses.length.toLocaleString() }}</strong> households
        <span v-if="filterVillage || filterTaluka || filterDistrict" class="stat-filter-note">
          (filtered from {{ houses.length.toLocaleString() }})
        </span>
      </span>
      <span class="stat-sep">·</span>
      <span class="stat-item"><strong>{{ stats?.farmers.toLocaleString() || 0 }}</strong> farmers</span>
      <span class="stat-sep">·</span>
      <span class="stat-item">Maharashtra</span>
      <span class="stat-sep" v-if="zoomLabel">·</span>
      <span class="stat-item zoom-label" v-if="zoomLabel">{{ zoomLabel }}</span>
    </div>

    <!-- ── LEFT SIDEBAR ── -->
    <div class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <button class="sidebar-toggle" @click="sidebarCollapsed = !sidebarCollapsed"
              :title="sidebarCollapsed ? 'Open panel' : 'Close panel'">
        {{ sidebarCollapsed ? '›' : '‹' }}
      </button>

      <div class="sidebar-body" v-if="!sidebarCollapsed && !loadingLiveData">

        <!-- LEGEND — mini 3D house icons match the map buildings -->
        <div class="panel-card">
          <div class="card-title">{{ legendTitle }}</div>
          <div class="legend-item" v-for="leg in currentLegend" :key="leg.label">
            <!-- Mini house: roof triangle (condition color) + wall block (sandstone) -->
            <span class="mini-house" :style="{ '--mh-roof': leg.color }">
              <span class="mh-roof"></span>
              <span class="mh-wall"></span>
            </span>
            <span class="legend-text">{{ leg.label }}</span>
          </div>
          <div class="legend-note">Roof color = {{ legendTitle.toLowerCase() }} status</div>
        </div>

        <!-- FIELD ISSUES -->
        <div class="panel-card">
          <div class="card-title">Field Issues
            <span class="card-title-sub">tap to recolor map</span>
          </div>
          <div v-if="!issueList.length" class="empty-note">No data for current selection.</div>
          <div v-for="issue in issueList" :key="issue.label">
            <div class="issue-row" :class="{ active: colorMode === issue.mode }"
                 @click="colorMode = issue.mode; activeIssue = activeIssue === issue.mode ? null : issue.mode">
              <!-- Mini house pip — roof = issue color, activates when row selected -->
              <span class="mini-house mini-house-sm" :style="{ '--mh-roof': issue.color }">
                <span class="mh-roof"></span>
                <span class="mh-wall"></span>
              </span>
              <div class="issue-body">
                <div class="issue-top">
                  <span class="issue-name">{{ issue.label }}</span>
                  <span class="issue-count">{{ issue.count.toLocaleString() }}</span>
                  <span class="issue-pct" :style="{ color: issue.color }">{{ issue.pct }}%</span>
                </div>
                <div class="issue-track">
                  <div class="issue-fill" :style="{ width: issue.pct + '%', background: issue.color }"></div>
                </div>
              </div>
              <span class="issue-chevron" :class="{ open: activeIssue === issue.mode }">›</span>
            </div>
            <transition name="drawer">
              <div v-if="activeIssue === issue.mode" class="issue-drawer" :style="{ borderLeftColor: issue.color }">
                <p class="drawer-cause"><strong>Cause:</strong> {{ issue.cause }}</p>
                <p class="drawer-solution"><strong>Solution:</strong> {{ issue.solution }}</p>
                <span class="drawer-scheme" :style="{ color: issue.color }">{{ issue.scheme }}</span>
              </div>
            </transition>
          </div>
        </div>

        <!-- PROBLEM FILTER -->
        <div class="panel-card">
          <div class="card-title">Problem Filter
            <span class="card-title-sub">highlight on map</span>
          </div>
          <label class="pf-item" v-for="pf in PROBLEM_FILTER_META" :key="pf.key">
            <input class="pf-check" type="checkbox" :value="pf.key" v-model="activeProblemFilters" />
            <span class="mini-house mini-house-sm" :style="{ '--mh-roof': pf.color }">
              <span class="mh-roof"></span>
              <span class="mh-wall"></span>
            </span>
            <span class="pf-label">{{ pf.label }}</span>
            <span class="pf-count">{{ problemFilterStats[pf.key] }}</span>
          </label>
          <div class="pf-summary" v-if="activeProblemFilters.length">
            <span><strong>{{ problemMatchCount }}</strong> flagged</span>
            <button class="pf-clear-btn" @click="activeProblemFilters = []">✕ Clear</button>
          </div>
          <div class="pf-hint" v-else>
            Select filters to highlight at-risk households and find high-need clusters on map
          </div>
        </div>


        <!-- AGRICULTURE DONUT CHARTS -->
        <div class="panel-card" v-if="pieCharts.length">
          <div class="card-title">Agriculture Overview</div>
          <div class="agri-chart" v-for="chart in pieCharts" :key="chart.title">
            <div class="agri-label">{{ chart.title }}</div>
            <div class="pie-row">
              <div class="pie-donut" :style="pieStyle(chart.segments)"></div>
              <div class="pie-legend">
                <div class="pie-item" v-for="seg in chart.segments" :key="seg.label">
                  <span class="pie-dot" :style="{ background: seg.color }"></span>
                  <span class="pie-name">{{ seg.label }}</span>
                  <span class="pie-pct">{{ seg.pct }}%</span>
                </div>
              </div>
            </div>
          </div>
        </div>

      </div>
    </div>

    <!-- HOVER TOOLTIP -->
    <div v-if="hoveredHouse" class="hover-card"
         :style="{ left: mouseX + 16 + 'px', top: mouseY - 8 + 'px' }">
      <!-- Header: name + condition badge -->
      <div class="hc-head">
        <div class="hc-name">{{ hoveredHouse.headName || 'Household #' + hoveredHouse.familyId }}</div>
        <span class="hc-badge" :style="{
          background: getConditionColor(hoveredHouse) + '22',
          color: getConditionColor(hoveredHouse),
          borderColor: getConditionColor(hoveredHouse) + '55'
        }">{{ getConditionLabel(hoveredHouse) }}</span>
      </div>
      <div class="hc-loc">{{ hoveredHouse.villageName || '—' }}{{ hoveredHouse.talukaName ? ' · ' + hoveredHouse.talukaName : '' }}</div>

      <!-- Status grid: 4 key indicators with color-coded dot -->
      <div class="hc-grid">
        <div class="hc-cell">
          <span class="hc-dot" :style="{ background: isRainFed(hoveredHouse) ? '#ef4444' : '#16a34a' }"></span>
          <span class="hc-ck">Irrigation</span>
          <span class="hc-cv">{{ isRainFed(hoveredHouse) ? 'Rain-fed' : 'Irrigated' }}</span>
        </div>
        <div class="hc-cell">
          <span class="hc-dot" :style="{
            background: (() => { const r=(hoveredHouse.rationCard||'').toLowerCase(); return r.includes('bpl')||r.includes('antyodaya') ? '#ef4444' : r.includes('apl') ? '#16a34a' : '#94a3b8' })()
          }"></span>
          <span class="hc-ck">Ration</span>
          <span class="hc-cv">{{ hoveredHouse.rationCard || '—' }}</span>
        </div>
        <div class="hc-cell">
          <span class="hc-dot" :style="{
            background: parseFloat(hoveredHouse.totalLand) > 0 ? '#16a34a' : '#ef4444'
          }"></span>
          <span class="hc-ck">Land</span>
          <span class="hc-cv">{{ parseFloat(hoveredHouse.totalLand) > 0 ? (hoveredHouse.totalLand + ' ac') : 'None' }}</span>
        </div>
        <div class="hc-cell">
          <span class="hc-dot" :style="{
            background: (hoveredHouse.lighting||'').toLowerCase() === 'electricity' ? '#16a34a' : '#f59e0b'
          }"></span>
          <span class="hc-ck">Power</span>
          <span class="hc-cv">{{ hoveredHouse.lighting || '—' }}</span>
        </div>
      </div>

      <!-- Crops row -->
      <div class="hc-crops" v-if="hoveredHouse.kharif || hoveredHouse.rabi">
        <span class="hc-ck">Crops</span>
        <span class="hc-cv">
          {{ [hoveredHouse.kharif, hoveredHouse.rabi].filter(Boolean).join(' · ') || '—' }}
        </span>
      </div>

      <div class="hc-hint">Click for full details</div>
    </div>

    <!-- DETAIL PANEL -->
    <transition name="slide">
      <div v-if="selectedHouse" class="detail-panel">

        <!-- ── Header ── -->
        <div class="detail-header">
          <div class="detail-header-info">
            <div class="detail-badge"
                 :style="{ background: getConditionColor(selectedHouse) + '18',
                           borderColor: getConditionColor(selectedHouse) + '55',
                           color: getConditionColor(selectedHouse) }">
              {{ getConditionLabel(selectedHouse) }}
            </div>
            <div class="detail-name">{{ selectedHouse.headName || 'Household' }}</div>
            <div class="detail-sub">
              <span class="detail-id-chip">ID {{ selectedHouse.familyId }}</span>
              <span>{{ selectedHouse.villageName }}</span>
              <span v-if="selectedHouse.talukaName"> · {{ selectedHouse.talukaName }}</span>
            </div>
          </div>
          <button class="detail-close" @click="selectedHouse = null" title="Close">×</button>
        </div>

        <button class="focus-btn" @click="flyToHouse(selectedHouse)">📍 Zoom to Location</button>

        <!-- ── Land & Crops ── -->
        <div class="dp-section-label">
          <span class="dp-section-icon">🌾</span> Agriculture
        </div>

        <div class="dp-stat-row">
          <div class="dp-stat">
            <div class="dp-stat-val">{{ selectedHouse.totalLand || '0' }} <small>ac</small></div>
            <div class="dp-stat-key">Total Land</div>
          </div>
          <div class="dp-stat">
            <div class="dp-stat-val">{{ selectedHouse.cultivatedLand || '0' }} <small>ac</small></div>
            <div class="dp-stat-key">Cultivated</div>
          </div>
        </div>

        <div class="dp-chip-row">
          <div class="dp-chip-block">
            <div class="dp-chip-label">Kharif Crop</div>
            <div class="dp-chip dp-chip-kharif">{{ selectedHouse.kharif || '—' }}</div>
          </div>
          <div class="dp-chip-block">
            <div class="dp-chip-label">Rabi Crop</div>
            <div class="dp-chip dp-chip-rabi">{{ selectedHouse.rabi || '—' }}</div>
          </div>
        </div>

        <!-- Irrigation source full-width -->
        <div class="dp-field-row">
          <span class="dp-field-icon">💧</span>
          <span class="dp-field-key">Irrigation Source</span>
          <span class="dp-field-val"
                :class="isIrrigated(selectedHouse) ? 'dp-ok' : 'dp-warn'">
            {{ selectedHouse.waterSource || '—' }}
          </span>
        </div>

        <!-- ── Infrastructure ── -->
        <div class="dp-section-label">
          <span class="dp-section-icon">🏠</span> Infrastructure
        </div>

        <div class="dp-field-row">
          <span class="dp-field-icon">🚽</span>
          <span class="dp-field-key">Latrine / Sanitation</span>
          <span class="dp-field-val" :style="{ color: getConditionColor(selectedHouse) }">
            {{ selectedHouse.latrine || '—' }}
          </span>
        </div>

        <div class="dp-field-row">
          <span class="dp-field-icon">⚡</span>
          <span class="dp-field-key">Lighting / Electricity</span>
          <span class="dp-field-val"
                :class="(selectedHouse.lighting || '').toLowerCase() === 'electricity' ? 'dp-ok' : 'dp-warn'">
            {{ selectedHouse.lighting || '—' }}
          </span>
        </div>

        <div class="dp-field-row">
          <span class="dp-field-icon">🪪</span>
          <span class="dp-field-key">Ration Card</span>
          <span class="dp-field-val">{{ selectedHouse.rationCard || '—' }}</span>
        </div>

        <!-- ── Farm Advisory ── -->
        <div v-if="getIssues(selectedHouse).length" class="detail-issues">
          <div class="dp-section-label">
            <span class="dp-section-icon">⚠️</span> Farm Advisory
          </div>
          <div class="advisory-card" v-for="iss in getIssues(selectedHouse)" :key="iss.label"
               :style="{ borderLeftColor: iss.color }">
            <div class="advisory-title" :style="{ color: iss.color }">{{ iss.label }}</div>
            <div class="advisory-row">
              <span class="advisory-tag cause">Cause</span>
              <span class="advisory-text">{{ iss.cause }}</span>
            </div>
            <div class="advisory-row">
              <span class="advisory-tag solution">Solution</span>
              <span class="advisory-text">{{ iss.solution }}</span>
            </div>
            <div class="advisory-scheme" :style="{ color: iss.color }">{{ iss.scheme }}</div>
          </div>
        </div>
        <div v-else class="all-good">
          <span>✓</span> This household looks well-resourced
        </div>

      </div>
    </transition>

    <!-- CLUSTER SOLUTION PANEL -->
    <transition name="slide">
      <div v-if="selectedCluster" class="cluster-panel">
        <button class="detail-close cluster-close" @click="selectedCluster = null">×</button>

        <!-- Header -->
        <div class="cluster-header">
          <div class="cluster-badge">⚠ High Need Area</div>
          <div class="cluster-count">
            <strong>{{ selectedCluster.count }}</strong> households in this zone
          </div>
        </div>

        <!-- Problems -->
        <div class="cluster-section-title" v-if="selectedCluster.problems.length">
          🔍 Main Issues Detected
        </div>

        <div class="cp-card" v-for="p in selectedCluster.problems" :key="p.key">
          <div class="cp-top">
            <span class="cp-emoji">{{ p.emoji }}</span>
            <div class="cp-info">
              <span class="cp-label">{{ p.label }}</span>
              <span class="cp-stat">{{ p.count }} of {{ selectedCluster.count }} families ({{ p.pct }}%)</span>
            </div>
          </div>
          <div class="cp-bar-track">
            <div class="cp-bar-fill" :style="{ width: p.pct + '%' }"></div>
          </div>
          <div class="cp-action">💡 {{ p.action }}</div>
          <p class="cp-solution">{{ p.solution }}</p>
          <div class="cp-scheme">📋 {{ p.scheme }}</div>
        </div>

        <div class="cluster-ok" v-if="!selectedCluster.problems.length">
          ✅ No major issues detected in this cluster based on current filters.
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { getHouses, getAgricultureInsights } from '../../api/index.js'
import * as Cesium from 'cesium'
import 'cesium/Build/Cesium/Widgets/widgets.css'

Cesium.Ion.defaultAccessToken = ''

// ── Core state ────────────────────────────────────────────────────────────────
const houses              = ref([])
const selectedHouse       = ref(null)
const hoveredHouse        = ref(null)
const mouseX              = ref(0)
const mouseY              = ref(0)
const colorMode           = ref('sanitation')
const activeIssue         = ref(null)
const tileStyle           = ref('satellite')   // default satellite view
const cesiumContainer     = ref(null)
const agricultureInsights = ref(null)
const loadingLiveData     = ref(true)
const sidebarCollapsed    = ref(false)
const cameraHeight        = ref(120000)

// Location filters — applied state (drives filteredHouses + map)
const filterDistrict = ref('')
const filterTaluka   = ref('')
const filterVillage  = ref('')

// Pending state — bound to UI dropdowns, only committed on Apply
const pendingDistrict = ref('')
const pendingTaluka   = ref('')
const pendingVillage  = ref('')

let viewer      = null
const entityMap  = new Map()   // entityId → house
const buildingIds = new Set()  // 3D box entity IDs
const pointIds    = new Set()  // point beacon entity IDs
const clusterIds   = new Set()  // High-Need cluster entity IDs
const clusterMap   = new Map()  // clusterEntityId → { count, lat, lng, problems[] }
let retryTimer    = null
let twinLoadSeq   = 0

// Zoom thresholds (meters)
const THRESHOLD_BUILDINGS = 3500   // below: show 3D boxes
                                   // above: show point beacons (always visible)

// ── Location filter options (derived from loaded data) ────────────────────────
const districtOptions = computed(() => {
  const seen = new Map()
  houses.value.forEach(h => {
    if (h.districtId != null && !seen.has(h.districtId)) {
      seen.set(h.districtId, { id: h.districtId, name: h.districtName || 'District ' + h.districtId })
    }
  })
  return [...seen.values()].sort((a, b) => a.name.localeCompare(b.name))
})

// Cascade options use PENDING refs so they update before Apply is clicked
const talukaOptions = computed(() => {
  if (!pendingDistrict.value) return []
  const seen = new Map()
  houses.value
    .filter(h => String(h.districtId) === String(pendingDistrict.value))
    .forEach(h => {
      if (h.talukaId != null && !seen.has(h.talukaId)) {
        seen.set(h.talukaId, { id: h.talukaId, name: h.talukaName || 'Taluka ' + h.talukaId })
      }
    })
  return [...seen.values()].sort((a, b) => a.name.localeCompare(b.name))
})

const villageOptions = computed(() => {
  if (!pendingTaluka.value) return []
  const seen = new Map()
  houses.value
    .filter(h => String(h.talukaId) === String(pendingTaluka.value))
    .forEach(h => {
      if (h.villageId != null && !seen.has(h.villageId)) {
        seen.set(h.villageId, { id: h.villageId, name: h.villageName || 'Village ' + h.villageId })
      }
    })
  return [...seen.values()].sort((a, b) => a.name.localeCompare(b.name))
})

// Active filtered subset displayed on map
const filteredHouses = computed(() => {
  let result = houses.value
  if (filterDistrict.value) result = result.filter(h => String(h.districtId) === String(filterDistrict.value))
  if (filterTaluka.value)   result = result.filter(h => String(h.talukaId)   === String(filterTaluka.value))
  if (filterVillage.value)  result = result.filter(h => String(h.villageId)  === String(filterVillage.value))
  return result
})

// ── Filter handlers ───────────────────────────────────────────────────────────
// Reset child pending selections when a parent changes
function onDistrictChange() {
  pendingTaluka.value  = ''
  pendingVillage.value = ''
}
function onTalukaChange() {
  pendingVillage.value = ''
}

// Apply: copy pending → applied; watcher on filteredHouses rebuilds map + flies
function applyFilters() {
  filterDistrict.value = pendingDistrict.value
  filterTaluka.value   = pendingTaluka.value
  filterVillage.value  = pendingVillage.value
}

function resetFilters() {
  pendingDistrict.value = filterDistrict.value = ''
  pendingTaluka.value   = filterTaluka.value   = ''
  pendingVillage.value  = filterVillage.value  = ''
}

// Check if pending state differs from applied state
const filtersDirty = computed(() =>
  pendingDistrict.value !== filterDistrict.value ||
  pendingTaluka.value   !== filterTaluka.value   ||
  pendingVillage.value  !== filterVillage.value
)

// ── Custom dropdown state ─────────────────────────────────────────────────────
const openDropdown = ref(null)

function toggleDropdown(name) {
  openDropdown.value = openDropdown.value === name ? null : name
}

function closeDropdowns() {
  openDropdown.value = null
}

// Selection handlers — each closes the dropdown and cascades resets downward
function selectDistrict(id) {
  pendingDistrict.value = id
  pendingTaluka.value   = ''
  pendingVillage.value  = ''
  closeDropdowns()
}

function selectTaluka(id) {
  pendingTaluka.value  = id
  pendingVillage.value = ''
  closeDropdowns()
}

function selectVillage(id) {
  pendingVillage.value = id
  closeDropdowns()
}

const COLOR_MODE_LABELS = {
  sanitation: 'Sanitation',
  irrigation: 'Irrigation',
  lighting:   'Electricity',
  crops:      'Crops / Season',
  land:       'Land Holdings',
  ration:     'Ration Card',
}

function selectColorMode(mode) {
  colorMode.value = mode
  closeDropdowns()
}

// Human-readable labels for current selections (shown in trigger button)
const selectedDistrictLabel = computed(() => {
  if (!pendingDistrict.value) return 'All Districts'
  return districtOptions.value.find(d => String(d.id) === String(pendingDistrict.value))?.name || 'All Districts'
})
const selectedTalukaLabel = computed(() => {
  if (!pendingTaluka.value) return 'All Talukas'
  return talukaOptions.value.find(t => String(t.id) === String(pendingTaluka.value))?.name || 'All Talukas'
})
const selectedVillageLabel = computed(() => {
  if (!pendingVillage.value) return 'All Villages'
  return villageOptions.value.find(v => String(v.id) === String(pendingVillage.value))?.name || 'All Villages'
})
const selectedColorModeLabel = computed(() => COLOR_MODE_LABELS[colorMode.value] || 'Sanitation')

// ── Problem Filter state ──────────────────────────────────────────────────────
// Array of active problem keys; v-model on checkboxes drives buildEntities()
const activeProblemFilters = ref([])

// Static metadata — one entry per problem type
const PROBLEM_FILTER_META = [
  { key: 'noRationCard', label: 'No Ration Card', color: '#f97316' },
  { key: 'noIrrigation', label: 'No Irrigation',  color: '#a78bfa' },
  { key: 'noLand',       label: 'No Own Land',    color: '#ef4444' },
]

// Returns true if the house matches the given problem key
function matchesProblemFilter(house, key) {
  if (key === 'noRationCard') {
    const r = (house.rationCard || '').toLowerCase().trim()
    return !r || r === 'none' || r === 'na' || r === 'no ration card'
  }
  if (key === 'noIrrigation') return isRainFed(house)
  if (key === 'noLand') {
    const land = parseFloat(house.totalLand) || 0
    const own  = (house.ownLand || '').toLowerCase()
    return land <= 0 || own !== 'yes'
  }
  return false
}

// Returns true if house satisfies ALL active problem filters (AND logic)
function matchesAllProblems(house) {
  return activeProblemFilters.value.every(k => matchesProblemFilter(house, k))
}

// Per-key counts for the sidebar display
const problemFilterStats = computed(() => {
  const list = filteredHouses.value
  return {
    noRationCard: list.filter(h => matchesProblemFilter(h, 'noRationCard')).length,
    noIrrigation: list.filter(h => matchesProblemFilter(h, 'noIrrigation')).length,
    noLand:       list.filter(h => matchesProblemFilter(h, 'noLand')).length,
  }
})

// Total households matching ALL active problem filters simultaneously
const problemMatchCount = computed(() => {
  if (!activeProblemFilters.value.length) return 0
  return filteredHouses.value.filter(matchesAllProblems).length
})

// ── Cluster solution panel state ──────────────────────────────────────────────
const selectedCluster = ref(null)  // { count, lat, lng, problems[] }


// Problem metadata used for cluster analysis (emoji + plain-language text)
const CLUSTER_PROBLEM_META = [
  {
    key:      'noRationCard',
    label:    'No Ration Card',
    emoji:    '🪪',
    action:   'Organise a ration card enrollment camp in this area',
    solution: 'Families can visit the local Gram Panchayat office with their Aadhaar card to enroll under the National Food Security Act and get subsidised food grains.',
    scheme:   'National Food Security Act (NFSA)',
  },
  {
    key:      'noIrrigation',
    label:    'No Irrigation',
    emoji:    '💧',
    action:   'Run an irrigation subsidy camp for this cluster',
    solution: 'Households can apply for free drip or sprinkler irrigation at the District Agriculture Office. Government pays up to 90% of the cost — families pay very little.',
    scheme:   'PMKSY – Pradhan Mantri Krishi Sinchai Yojana',
  },
  {
    key:      'noLand',
    label:    'No Own Land',
    emoji:    '🌾',
    action:   'Connect families with lease farming and livelihood groups',
    solution: 'Landless families can join Farmer Producer Organisations (FPOs) for shared lease farming, pooled resources, and stable income — no land ownership required.',
    scheme:   'PM-FME / NRLM via Gram Panchayat',
  },
]

// Analyse the houses inside a cluster → top 2 problems with counts + solutions
function analyzeCluster(houseList) {
  const total = houseList.length
  return CLUSTER_PROBLEM_META
    .map(meta => ({
      ...meta,
      count: houseList.filter(h => matchesProblemFilter(h, meta.key)).length,
    }))
    .filter(p => p.count > 0)
    .sort((a, b) => b.count - a.count)
    .slice(0, 2)
    .map(p => ({ ...p, pct: Math.round((p.count / total) * 100) }))
}

// When filter changes, rebuild entities and fly to filtered data
watch(filteredHouses, (newHouses) => {
  if (!viewer || loadingLiveData.value) return
  buildEntities()
  if (newHouses.length) {
    setTimeout(() => flyToPoints(newHouses), 150)
  }
}, { flush: 'post' })

// ── Zoom label ────────────────────────────────────────────────────────────────
const zoomLabel = computed(() => {
  const h = cameraHeight.value
  if (h < THRESHOLD_BUILDINGS) return '3D buildings visible'
  if (h < 15000)  return 'Zoom in to see buildings'
  return null
})

// ── Issue analysis ────────────────────────────────────────────────────────────
const stats = computed(() => {
  if (!filteredHouses.value.length) return null
  const list  = filteredHouses.value
  const total = list.length
  return {
    total,
    farmers:  list.filter(h => (h.ownLand || '').toLowerCase() === 'yes').length,
    noToilet: list.filter(h => !h.latrine || h.latrine === 'No Latrine' || h.latrine === 'None').length,
    noElec:   list.filter(h => !h.lighting || h.lighting === 'Kerosene' || h.lighting === 'None').length,
    noIrrig:  list.filter(h => isRainFed(h)).length,
    bpl:      list.filter(h => (h.rationCard || '').toLowerCase().includes('bpl') || (h.rationCard || '').toLowerCase().includes('antyodaya')).length,
  }
})

function isRainFed(house) {
  const w = (house.waterSource || '').toLowerCase()
  return !w || w === 'none' || w === 'rain fed' || w.includes('no source')
}

function isIrrigated(house) { return !isRainFed(house) }

const ISSUE_META = {
  sanitation: {
    cause: 'No toilet facility — household relies on open defecation or shared community latrine',
    solution: 'Apply for toilet construction under Swachh Bharat Mission at the Gram Panchayat',
    scheme: 'Swachh Bharat Mission — ₹12,000 subsidy per toilet',
  },
  lighting: {
    cause: 'No grid connection — household uses kerosene or has no lighting source',
    solution: 'Apply under Saubhagya Scheme for free connection, or KUSUM for a solar home system',
    scheme: 'PM Saubhagya / KUSUM Solar Scheme',
  },
  irrigation: {
    cause: 'Entirely rain-fed farming — high risk during drought or irregular monsoon',
    solution: 'Apply for micro-irrigation (drip/sprinkler) and check dug-well subsidy eligibility',
    scheme: 'PMKSY — Pradhan Mantri Krishi Sinchai Yojana',
  },
  ration: {
    cause: 'Household classified as BPL/Antyodaya — economically vulnerable',
    solution: 'Enroll in PM-KISAN for ₹6,000/year income support; verify NFSA food grain entitlement',
    scheme: 'PM-KISAN + National Food Security Act (NFSA)',
  },
}

const issueList = computed(() => {
  if (!stats.value) return []
  const { total, noToilet, noElec, noIrrig, bpl } = stats.value
  return [
    { label: 'No Sanitation',  count: noToilet, pct: Math.round(noToilet / total * 100), color: '#ef4444', mode: 'sanitation', ...ISSUE_META.sanitation },
    { label: 'No Electricity', count: noElec,   pct: Math.round(noElec   / total * 100), color: '#f59e0b', mode: 'lighting',   ...ISSUE_META.lighting   },
    { label: 'No Irrigation',  count: noIrrig,  pct: Math.round(noIrrig  / total * 100), color: '#a78bfa', mode: 'irrigation', ...ISSUE_META.irrigation  },
    { label: 'BPL Households', count: bpl,      pct: Math.round(bpl      / total * 100), color: '#60a5fa', mode: 'ration',     ...ISSUE_META.ration      },
  ]
})

// ── Legend ────────────────────────────────────────────────────────────────────
const legendTitle = computed(() => ({
  sanitation: 'Sanitation', irrigation: 'Irrigation',
  lighting:   'Electricity', ration: 'Ration Card',
  crops:      'Crops / Season', land: 'Land Holdings',
})[colorMode.value] || 'Legend')

const currentLegend = computed(() => {
  if (colorMode.value === 'sanitation') return [
    { color: '#16a34a', label: 'Has toilet facility' },
    { color: '#f59e0b', label: 'Pit / open latrine' },
    { color: '#ef4444', label: 'No sanitation' },
  ]
  if (colorMode.value === 'irrigation') return [
    { color: '#16a34a', label: 'Irrigated source' },
    { color: '#ef4444', label: 'Rain-fed / no irrigation' },
  ]
  if (colorMode.value === 'lighting') return [
    { color: '#16a34a', label: 'Grid electricity' },
    { color: '#f59e0b', label: 'Kerosene lamp' },
    { color: '#ef4444', label: 'No lighting' },
  ]
  if (colorMode.value === 'crops') return [
    { color: '#16a34a', label: 'Both Kharif & Rabi' },
    { color: '#f59e0b', label: 'Kharif only' },
    { color: '#38bdf8', label: 'Rabi only' },
    { color: '#94a3b8', label: 'No crop data' },
  ]
  if (colorMode.value === 'land') return [
    { color: '#16a34a', label: '> 5 acres (Large)' },
    { color: '#4ade80', label: '2.5–5 acres (Medium)' },
    { color: '#f59e0b', label: '1–2.5 acres (Small)' },
    { color: '#ef4444', label: '≤ 1 acre (Marginal)' },
    { color: '#94a3b8', label: 'No land' },
  ]
  return [
    { color: '#16a34a', label: 'APL — Above Poverty Line' },
    { color: '#ef4444', label: 'BPL / Antyodaya' },
    { color: '#94a3b8', label: 'No ration data' },
  ]
})

// ── Agriculture bar charts ────────────────────────────────────────────────────
function pct(v, t) { return t ? Math.round(v / t * 100) : 0 }

const pieCharts = computed(() => {
  const list  = filteredHouses.value
  const total = list.length
  if (!total) return []

  let both = 0, kOnly = 0, rOnly = 0, none = 0
  let marginal = 0, small = 0, medLarge = 0
  const irrigated = list.filter(h => isIrrigated(h)).length

  list.forEach(h => {
    const k = (h.kharif || '').toLowerCase() === 'yes'
    const r = (h.rabi   || '').toLowerCase() === 'yes'
    if (k && r) both++; else if (k) kOnly++; else if (r) rOnly++; else none++
    const land = parseFloat(h.totalLand) || 0
    if (land <= 1) marginal++; else if (land <= 2.5) small++; else medLarge++
  })

  return [
    {
      title: 'Irrigation Coverage',
      segments: [
        { label: 'Irrigated',     pct: pct(irrigated,      total), color: '#16a34a' },
        { label: 'Rain-fed',      pct: pct(total-irrigated, total), color: '#ef4444' },
      ],
    },
    {
      title: 'Crop Seasons',
      segments: [
        { label: 'Both seasons',  pct: pct(both,  total), color: '#16a34a' },
        { label: 'Kharif only',   pct: pct(kOnly, total), color: '#f59e0b' },
        { label: 'Rabi only',     pct: pct(rOnly, total), color: '#38bdf8' },
        { label: 'No crop data',  pct: pct(none,  total), color: '#94a3b8' },
      ],
    },
    {
      title: 'Land Holdings',
      segments: [
        { label: 'Marginal ≤1ac', pct: pct(marginal, total), color: '#ef4444' },
        { label: 'Small 1–2.5ac', pct: pct(small,    total), color: '#f59e0b' },
        { label: 'Med/Large >2.5',pct: pct(medLarge, total), color: '#16a34a' },
      ],
    },
  ]
})

// ── Donut chart helper ────────────────────────────────────────────────────────
function pieStyle(segments) {
  const total = segments.reduce((sum, s) => sum + s.pct, 0)
  if (!total) return { background: '#e5e7eb' }
  let start = 0
  const stops = segments.map(seg => {
    const span = (seg.pct / total) * 360
    const end  = start + span
    const val  = `${seg.color} ${start.toFixed(1)}deg ${end.toFixed(1)}deg`
    start = end
    return val
  })
  return { background: `conic-gradient(${stops.join(', ')})` }
}

// ── Color helpers ────────────────────────────────────────────────────────────
function getConditionColor(house) {
  if (colorMode.value === 'sanitation') {
    const l = (house.latrine || '').toLowerCase()
    if (!l || l === 'no latrine' || l === 'none') return '#ef4444'
    if (l.includes('pit') || l.includes('open'))  return '#f59e0b'
    return '#16a34a'
  }
  if (colorMode.value === 'irrigation') {
    return isRainFed(house) ? '#ef4444' : '#16a34a'
  }
  if (colorMode.value === 'lighting') {
    const l = (house.lighting || '').toLowerCase()
    return l === 'electricity' ? '#16a34a' : l === 'kerosene' ? '#f59e0b' : '#ef4444'
  }
  if (colorMode.value === 'crops') {
    const k = (house.kharif || '').toLowerCase() === 'yes'
    const r = (house.rabi   || '').toLowerCase() === 'yes'
    if (k && r) return '#16a34a'
    if (k)      return '#f59e0b'
    if (r)      return '#38bdf8'
    return '#94a3b8'
  }
  if (colorMode.value === 'land') {
    const a = parseFloat(house.totalLand) || 0
    if (a === 0)  return '#94a3b8'
    if (a <= 1)   return '#ef4444'
    if (a <= 2.5) return '#f59e0b'
    if (a <= 5)   return '#4ade80'
    return '#16a34a'
  }
  const r = (house.rationCard || '').toLowerCase()
  if (r.includes('bpl') || r.includes('antyodaya')) return '#ef4444'
  if (r.includes('apl')) return '#16a34a'
  return '#94a3b8'
}

function getConditionLabel(house) {
  const color = getConditionColor(house)
  if (colorMode.value === 'crops') {
    if (color === '#16a34a') return 'Double Crop'
    if (color === '#f59e0b') return 'Kharif Only'
    if (color === '#38bdf8') return 'Rabi Only'
    return 'No Crop Data'
  }
  if (colorMode.value === 'land') {
    const a = parseFloat(house.totalLand) || 0
    if (a === 0)  return 'Landless'
    if (a <= 1)   return 'Marginal Farmer'
    if (a <= 2.5) return 'Small Farmer'
    if (a <= 5)   return 'Medium Holding'
    return 'Large Holding'
  }
  if (color === '#ef4444') return 'High Risk'
  if (color === '#f59e0b') return 'Needs Attention'
  return 'Good Standing'
}

function getIssues(house) {
  const issues = []
  const totalLand      = parseFloat(house.totalLand) || 0
  const cultivatedLand = parseFloat(house.cultivatedLand) || 0
  const ownLand        = (house.ownLand || '').toLowerCase()
  const k = normalizeCrop(house.kharif)
  const r = normalizeCrop(house.rabi)

  function normalizeCrop(v) {
    const s = String(v || '').trim().toLowerCase()
    return s && s !== 'no' && s !== 'none' && s !== 'na' ? s : ''
  }

  if (ownLand !== 'yes' || totalLand <= 0) {
    issues.push({ label: 'No Own Land', color: '#ef4444',
      cause: 'Household has no cultivable land — farm income is unstable.',
      solution: 'Explore lease farming and enroll in FPO/NRLM livelihood programs.',
      scheme: 'PM-FME / NRLM via Gram Panchayat' })
  } else if (totalLand <= 1) {
    issues.push({ label: 'Marginal Holding', color: '#f59e0b',
      cause: `Only ${totalLand.toFixed(2)} acres — limits single-crop income.`,
      solution: 'Adopt high-value short-duration crops and kitchen horticulture.',
      scheme: 'MIDH + ATMA farm advisory' })
  }

  if (cultivatedLand <= 0 && totalLand > 0) {
    issues.push({ label: 'Uncultivated Land', color: '#ef4444',
      cause: 'Cultivated area is zero despite owning land — land is fallow.',
      solution: 'Create a seasonal crop plan with seed-kit and extension support.',
      scheme: 'State Agriculture Dept seed kit + KVK guidance' })
  }

  if (isRainFed(house)) {
    issues.push({ label: 'No Irrigation', color: '#a78bfa',
      cause: 'Fully rain-fed — high crop-loss risk during irregular monsoon.',
      solution: 'Register for drip/sprinkler subsidy and community water support.',
      scheme: 'PMKSY micro-irrigation subsidy' })
  }

  if (!k && !r) {
    issues.push({ label: 'No Crop Record', color: '#60a5fa',
      cause: 'No kharif or rabi crop recorded.',
      solution: 'Prepare a two-season crop calendar with extension workers.',
      scheme: 'ATMA + mKisan advisories' })
  } else if (!k || !r) {
    issues.push({ label: 'Single Season Only', color: '#38bdf8',
      cause: 'Active in only one season — reduces annual income stability.',
      solution: 'Introduce a second-season crop with short-duration seed support.',
      scheme: 'RKVY seasonal diversification' })
  }

  return issues
}

// ── Cesium helpers ─────────────────────────────────────────────────────────────
function cesiumColor(house) {
  const base  = Cesium.Color.fromCssColorString(getConditionColor(house))
  return new Cesium.Color(base.red * 0.8, base.green * 0.8, base.blue * 0.8, 1.0)
}

function landHeight(house) {
  return Math.max(8, Math.min(8 + (parseFloat(house.totalLand) || 0) * 2.4, 18))
}

// ── Imagery providers ─────────────────────────────────────────────────────────
function buildImageryProvider(style) {
  const s = style || tileStyle.value
  if (s === 'street') {
    return new Cesium.UrlTemplateImageryProvider({
      url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Street_Map/MapServer/tile/{z}/{y}/{x}',
      credit: 'Tiles © Esri',
      maximumLevel: 19, tileWidth: 256, tileHeight: 256,
    })
  }
  return new Cesium.UrlTemplateImageryProvider({
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
    credit: 'Tiles © Esri',
    maximumLevel: 19, tileWidth: 256, tileHeight: 256,
  })
}

function toggleTile() {
  tileStyle.value = tileStyle.value === 'satellite' ? 'street' : 'satellite'
  if (!viewer) return
  viewer.imageryLayers.removeAll()
  viewer.imageryLayers.addImageryProvider(buildImageryProvider())
}

// ── Camera ────────────────────────────────────────────────────────────────────
function flyToMaharashtra() {
  if (!viewer) return
  viewer.camera.flyTo({
    destination: Cesium.Cartesian3.fromDegrees(76.0, 19.5, 150000),
    orientation: { heading: 0, pitch: Cesium.Math.toRadians(-40), roll: 0 },
    duration: 1.8,
  })
}

function flyToPoints(list) {
  if (!viewer || !list.length) return
  const pts = list
    .filter(h => Number.isFinite(h.longitude) && Number.isFinite(h.latitude))
    .map(h => Cesium.Cartesian3.fromDegrees(h.longitude, h.latitude, 0))
  if (!pts.length) return
  const sphere = Cesium.BoundingSphere.fromPoints(pts)
  const range  = Math.max(sphere.radius * 2.6, 300)
  viewer.camera.flyToBoundingSphere(sphere, {
    duration: 2,
    offset: new Cesium.HeadingPitchRange(
      Cesium.Math.toRadians(5),
      Cesium.Math.toRadians(-42),
      range,
    ),
  })
}

function flyToVillage() { flyToPoints(filteredHouses.value) }

function flyToHouse(house) {
  if (!viewer || !house) return
  viewer.camera.flyTo({
    destination: Cesium.Cartesian3.fromDegrees(house.longitude, house.latitude, 200),
    orientation: { heading: 0, pitch: Cesium.Math.toRadians(-55), roll: 0 },
    duration: 1.5,
  })
}

// ── Zoom-based entity visibility ──────────────────────────────────────────────
// KEY FIX: entities are ALWAYS visible — just switch between buildings and points.
// There is NO "hide all" state. At every zoom level, something is shown.
function updateZoomVisibility() {
  if (!viewer || viewer.isDestroyed()) return
  const pos = viewer.camera.positionCartographic
  if (!pos) return
  const h = pos.height
  cameraHeight.value = Math.round(h)

  const showBuildings = h < THRESHOLD_BUILDINGS

  viewer.entities.values.forEach(entity => {
    if (buildingIds.has(entity.id)) {
      entity.show = showBuildings
    } else if (pointIds.has(entity.id)) {
      entity.show = !showBuildings   // points always on when buildings are off
    }
  })

}

function setupZoomListener() {
  if (!viewer) return
  viewer.camera.percentageChanged = 0.03
  viewer.camera.changed.addEventListener(updateZoomVisibility)
}

// ── Jitter: spread houses that share the same coordinate ─────────────────────
// Groups duplicates by rounded lat/lng key, then places each in a small circle.
// Radius scales with group size (capped at 20 m) so clusters stay tight.
function computeJitteredPositions(houseList) {
  const posMap = new Map()  // 'lat6,lng6' → array of list-indices
  houseList.forEach((h, i) => {
    if (!Number.isFinite(h.latitude) || !Number.isFinite(h.longitude)) return
    const key = `${h.latitude.toFixed(6)},${h.longitude.toFixed(6)}`
    if (!posMap.has(key)) posMap.set(key, [])
    posMap.get(key).push(i)
  })

  const out = houseList.map(h => ({ lat: h.latitude, lng: h.longitude }))

  posMap.forEach(indices => {
    if (indices.length < 2) return
    const count    = indices.length
    const radiusM  = Math.min(7 + count * 1.2, 20)  // 8–20 m depending on density
    const refH     = houseList[indices[0]]
    const cosLat   = Math.cos((refH.latitude * Math.PI) / 180)

    indices.forEach((listIdx, slot) => {
      const angle = (2 * Math.PI * slot) / count
      out[listIdx] = {
        lat: refH.latitude  + (radiusM * Math.cos(angle)) / 111_000,
        lng: refH.longitude + (radiusM * Math.sin(angle)) / (111_000 * cosLat),
      }
    })
  })

  return out
}

// ── Haversine distance (metres) ───────────────────────────────────────────────
function haversineM(lat1, lng1, lat2, lng2) {
  const R  = 6_371_000
  const φ1 = (lat1 * Math.PI) / 180
  const φ2 = (lat2 * Math.PI) / 180
  const Δφ = ((lat2 - lat1) * Math.PI) / 180
  const Δλ = ((lng2 - lng1) * Math.PI) / 180
  const a  = Math.sin(Δφ / 2) ** 2 + Math.cos(φ1) * Math.cos(φ2) * Math.sin(Δλ / 2) ** 2
  return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
}

// Group problem houses that lie within RADIUS_M of each other.
// Returns clusters with ≥ MIN_SIZE members only (centroid + count).
function computeProblemClusters(houseList) {
  const RADIUS_M = 300
  const MIN_SIZE = 5
  const visited  = new Set()
  const clusters = []

  houseList.forEach((h, i) => {
    if (!Number.isFinite(h.latitude) || !Number.isFinite(h.longitude)) return
    if (visited.has(i)) return
    const group = [i]
    visited.add(i)
    houseList.forEach((h2, j) => {
      if (visited.has(j) || !Number.isFinite(h2.latitude) || !Number.isFinite(h2.longitude)) return
      if (haversineM(h.latitude, h.longitude, h2.latitude, h2.longitude) <= RADIUS_M) {
        group.push(j)
        visited.add(j)
      }
    })
    if (group.length >= MIN_SIZE) {
      const lat    = group.reduce((s, idx) => s + houseList[idx].latitude,  0) / group.length
      const lng    = group.reduce((s, idx) => s + houseList[idx].longitude, 0) / group.length
      const houses = group.map(idx => houseList[idx])   // actual house objects for analysis
      clusters.push({ lat, lng, count: group.length, houses })
    }
  })

  return clusters
}

// Draw a translucent red circle + "High Need Area" label for each cluster.
// Also registers each entity in clusterMap so clicks can open the solution panel.
function addClusterEntities(problemHouses) {
  clusterMap.clear()
  const clusters = computeProblemClusters(problemHouses)

  clusters.forEach(({ lat, lng, count, houses }) => {
    const pos = Cesium.Cartesian3.fromDegrees(lng, lat, 0)

    // Pre-analyse once so the click handler doesn't redo work
    const problems = analyzeCluster(houses)
    const clusterData = { count, lat, lng, problems }

    const circleEnt = viewer.entities.add({
      position: pos,
      show: true,
      ellipse: {
        semiMajorAxis:   130,
        semiMinorAxis:   130,
        material:        Cesium.Color.fromCssColorString('#ef4444').withAlpha(0.15),
        outline:         true,
        outlineColor:    Cesium.Color.fromCssColorString('#ef4444').withAlpha(0.75),
        outlineWidth:    2,
        heightReference: Cesium.HeightReference.CLAMP_TO_GROUND,
      },
    })

    const labelEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, 35),
      show: true,
      label: {
        text:             `⚠ High Need Area\n${count} households`,
        font:             '600 12px system-ui, sans-serif',
        fillColor:        Cesium.Color.WHITE,
        outlineColor:     Cesium.Color.fromCssColorString('#7f1d1d'),
        outlineWidth:     2,
        style:            Cesium.LabelStyle.FILL_AND_OUTLINE,
        verticalOrigin:   Cesium.VerticalOrigin.BOTTOM,
        horizontalOrigin: Cesium.HorizontalOrigin.CENTER,
        pixelOffset:      new Cesium.Cartesian2(0, -6),
        disableDepthTestDistance: Number.POSITIVE_INFINITY,
        showBackground:   true,
        backgroundColor:  Cesium.Color.fromCssColorString('#ef4444').withAlpha(0.88),
        backgroundPadding: new Cesium.Cartesian2(8, 5),
      },
    })

    // Register both entities so either can be clicked to open the panel
    clusterIds.add(circleEnt.id)
    clusterIds.add(labelEnt.id)
    clusterMap.set(circleEnt.id, clusterData)
    clusterMap.set(labelEnt.id,  clusterData)
  })
}

// ── Build Cesium entities ─────────────────────────────────────────────────────
function buildEntities() {
  if (!viewer) return
  viewer.entities.removeAll()
  entityMap.clear()
  buildingIds.clear()
  pointIds.clear()
  clusterIds.clear()
  clusterMap.clear()
  selectedCluster.value = null

  const selectedId       = selectedHouse.value?.familyId
  const camH             = viewer.camera.positionCartographic?.height ?? cameraHeight.value
  const showBuildings    = camH < THRESHOLD_BUILDINGS
  const houseList        = filteredHouses.value
  const jittered         = computeJitteredPositions(houseList)
  const hasProblemFilter = activeProblemFilters.value.length > 0

  // Collect problem-matched houses for cluster analysis after the loop
  const problemHouses = []

  houseList.forEach((house, idx) => {
    const { lat, lng } = jittered[idx]
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) return

    const isSelected    = house.familyId === selectedId
    const isProblem     = hasProblemFilter && matchesAllProblems(house)
    // Non-flagged houses are still rendered when a problem filter is active so
    // the colorMode theme stays readable as context.  They are visually dimmed
    // (lower alpha on wall) and de-emphasized in outline weight.
    const isBackground  = hasProblemFilter && !isProblem && !isSelected

    if (isProblem) problemHouses.push(house)

    // ── Colour logic ──────────────────────────────────────────────────────────
    // DESIGN RULE: colorMode ALWAYS controls roof color — it is never overridden.
    // Problem filter adds a VISUAL OVERLAY (bright outlines + alert ring) on top
    // so both signals are simultaneously readable:
    //   Green roof  = good irrigation  |  red outline = also has no ration card
    //   Orange roof = kharif-only crop |  red outline = flagged by problem filter
    const conditionColor = cesiumColor(house)   // 80% of getConditionColor hex

    // Roof: ALWAYS the colorMode condition color (selected = gold override only).
    // Background (non-flagged) houses are dimmed to 40% alpha so flagged ones pop.
    const roofAlpha = isSelected ? 1.0 : isBackground ? 0.35 : 1.0
    const roofColor = isSelected
      ? Cesium.Color.fromCssColorString('#facc15').withAlpha(1.0)
      : conditionColor.withAlpha(roofAlpha)   // problem filter does NOT change hue

    // Wall: sandstone base; flagged houses get a vivid pale-red wall tint.
    // Background houses are also dimmed to keep flagged ones dominant.
    const wallColor = isSelected
      ? Cesium.Color.fromCssColorString('#fef3c7').withAlpha(1.0)
      : isProblem
        ? Cesium.Color.fromCssColorString('#f4b8b8').withAlpha(0.95)  // pale red — flagged
        : Cesium.Color.fromCssColorString('#c8a97e').withAlpha(isBackground ? 0.3 : 1.0)

    // Outline: flagged → vivid red; background → nearly invisible; normal → mortar
    const wallOutline = isSelected
      ? Cesium.Color.fromCssColorString('#f59e0b').withAlpha(1.0)
      : isProblem
        ? Cesium.Color.fromCssColorString('#dc2626').withAlpha(1.0)   // vivid red ring
        : Cesium.Color.fromCssColorString('#7a6040').withAlpha(isBackground ? 0.2 : 1.0)

    const footprint = 10
    const baseH     = 7
    const roofH     = Math.max(2.5, Math.min(landHeight(house) * 0.22, 5))

    // ── 3D building (low zoom) ──
    const baseEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, baseH / 2),
      show: showBuildings,
      box: {
        dimensions:   new Cesium.Cartesian3(footprint, footprint, baseH),
        material:     wallColor,
        outline:      true,
        outlineColor: wallOutline,
        outlineWidth: isSelected ? 2 : (isProblem ? 2 : 1.5),
      },
    })

    const roofEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, baseH + roofH / 2),
      show: showBuildings,
      box: {
        dimensions:   new Cesium.Cartesian3(footprint * 0.88, footprint * 0.88, roofH),
        material:     roofColor,
        outline:      true,
        // Flagged roof gets a bright white edge so it stands out from background
        outlineColor: isSelected
          ? Cesium.Color.WHITE
          : isProblem
            ? Cesium.Color.fromCssColorString('#fff5f5').withAlpha(0.95)
            : roofColor.darken(0.25, new Cesium.Color()),
        outlineWidth: isSelected ? 2.5 : (isProblem ? 2.5 : (isBackground ? 0.5 : 1.5)),
      },
    })

    // ── Point beacon (high zoom) ──
    const ptEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, 1),
      show: !showBuildings,
      point: {
        pixelSize:    isSelected ? 13 : isProblem ? 11 : isBackground ? 5 : 8,
        color:        roofColor,
        outlineColor: isSelected
          ? Cesium.Color.WHITE
          : isProblem
            ? Cesium.Color.fromCssColorString('#dc2626').withAlpha(0.9)
            : Cesium.Color.fromCssColorString('#1a1a1a').withAlpha(isBackground ? 0.25 : 0.7),
        outlineWidth:    isSelected ? 2 : (isProblem ? 2.5 : 1.5),
        heightReference: Cesium.HeightReference.CLAMP_TO_GROUND,
      },
    })

    buildingIds.add(baseEnt.id)
    buildingIds.add(roofEnt.id)
    pointIds.add(ptEnt.id)

    entityMap.set(baseEnt.id, house)
    entityMap.set(roofEnt.id, house)
    entityMap.set(ptEnt.id,   house)
  })

  // Add High-Need Area cluster markers when problem filters are active
  if (hasProblemFilter && problemHouses.length >= 5) {
    addClusterEntities(problemHouses)
  }

  // Map stays clean by default — no permanent icons above houses.
  // Data is surfaced through building color (colorMode) and click/hover tooltips.
}

// ── PDF download ──────────────────────────────────────────────────────────────
const pdfLoading = ref(false)

// Draws a donut chart onto an off-screen Canvas and returns a base64 PNG string.
// This is purely client-side — no external library needed.
function renderChartToBase64(segments, size = 160) {
  const canvas = document.createElement('canvas')
  canvas.width  = size * 2   // 2× for retina sharpness
  canvas.height = size * 2
  const ctx = canvas.getContext('2d')
  ctx.scale(2, 2)

  const cx = size / 2, cy = size / 2
  const outerR = size / 2 - 4
  const innerR = outerR * 0.52   // donut hole radius

  const total = segments.reduce((sum, s) => sum + (s.pct || 0), 0)
  if (!total) return null

  // White background so the PNG has an opaque background
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, size, size)

  let startAngle = -Math.PI / 2  // 12 o'clock
  segments.forEach(seg => {
    if (!seg.pct) return
    const sliceAngle = (seg.pct / total) * 2 * Math.PI
    ctx.beginPath()
    ctx.moveTo(cx, cy)
    ctx.arc(cx, cy, outerR, startAngle, startAngle + sliceAngle)
    ctx.closePath()
    ctx.fillStyle = seg.color
    ctx.fill()
    startAngle += sliceAngle
  })

  // Punch donut hole
  ctx.beginPath()
  ctx.arc(cx, cy, innerR, 0, 2 * Math.PI)
  ctx.fillStyle = '#ffffff'
  ctx.fill()

  // Strip the data-URL prefix — backend expects raw base64
  return canvas.toDataURL('image/png').replace('data:image/png;base64,', '')
}

// POST /api/pdf/report → download PDF blob
async function downloadPDF() {
  if (pdfLoading.value) return
  pdfLoading.value = true
  try {
    // Resolve human-readable names from the currently applied filter IDs
    const districtName = filterDistrict.value
      ? (houses.value.find(h => String(h.districtId) === String(filterDistrict.value))?.districtName || '')
      : ''
    const talukaName = filterTaluka.value
      ? (houses.value.find(h => String(h.talukaId) === String(filterTaluka.value))?.talukaName || '')
      : ''
    const villageName = filterVillage.value
      ? (houses.value.find(h => String(h.villageId) === String(filterVillage.value))?.villageName || '')
      : ''

    // Render all sidebar donut charts to PNG and ship them to the backend
    const charts = pieCharts.value
      .map(chart => ({
        title:    chart.title,
        image:    renderChartToBase64(chart.segments) || '',
        segments: chart.segments.map(s => ({ label: s.label, pct: s.pct, color: s.color })),
      }))
      .filter(c => c.image)

    // Build problem filter summary to embed in PDF
    const problemFilters = PROBLEM_FILTER_META.map(pf => ({
      key:        pf.key,
      label:      pf.label,
      count:      problemFilterStats.value[pf.key] ?? 0,
      total:      filteredHouses.value.length,
      active:     activeProblemFilters.value.includes(pf.key),
    }))
    const problemMatchTotal = problemMatchCount.value

    const body = {
      districtId:   filterDistrict.value ? String(filterDistrict.value) : '',
      districtName,
      talukaId:     filterTaluka.value   ? String(filterTaluka.value)   : '',
      talukaName,
      villageId:    filterVillage.value  ? String(filterVillage.value)  : '',
      villageName,
      charts,
      problemFilters,
      problemMatchTotal,
    }

    const res = await fetch('/api/pdf/report', {
      method:  'POST',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify(body),
    })

    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      console.error('[PDF] Generation failed:', err)
      return
    }

    const blob = await res.blob()
    const url  = URL.createObjectURL(blob)
    const a    = document.createElement('a')
    a.href     = url
    const stem = ['AgriTwin',
      districtName.replace(/\s+/g, '_') || '',
      talukaName.replace(/\s+/g, '_')   || '',
      villageName.replace(/\s+/g, '_')  || '',
      new Date().toISOString().slice(0, 10),
    ].filter(Boolean).join('_')
    a.download = stem + '.pdf'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (err) {
    console.error('[PDF] Download error:', err)
  } finally {
    pdfLoading.value = false
  }
}

// ── Watchers ──────────────────────────────────────────────────────────────────
watch(colorMode,            () => { if (viewer) buildEntities() })
watch(selectedHouse,        () => { if (viewer) buildEntities() })
watch(activeProblemFilters, () => { if (viewer) buildEntities() }, { deep: true })

// ── Data loading ──────────────────────────────────────────────────────────────
function clearRetryTimer() {
  if (retryTimer) { clearTimeout(retryTimer); retryTimer = null }
}

async function fetchAllHouses() {
  const limit = 2000
  const all = []
  let page = 1, total = null
  while (true) {
    const res   = await getHouses({ page, limit })
    const chunk = res.data || []
    if (!chunk.length) break
    all.push(...chunk)
    if (typeof res.total === 'number') total = res.total
    if (chunk.length < limit) break
    if (total !== null && all.length >= total) break
    if (page >= 20) break
    page++
  }
  return all
}

async function loadData(attempt = 0) {
  const seq     = ++twinLoadSeq
  const results = await Promise.allSettled([fetchAllHouses(), getAgricultureInsights()])
  if (seq !== twinLoadSeq) return

  if (results[0].status === 'fulfilled') {
    const real = results[0].value || []
    if (real.length > 0) {
      clearRetryTimer()
      houses.value = real
      if (viewer) {
        buildEntities()
        // Intro: fly from state level → village data
        setTimeout(() => flyToPoints(real), 300)
      }
    } else if (attempt < 10) {
      retryTimer = setTimeout(() => { if (seq === twinLoadSeq) loadData(attempt + 1) }, 3000)
    }
  } else if (attempt < 10) {
    retryTimer = setTimeout(() => { if (seq === twinLoadSeq) loadData(attempt + 1) }, 3000)
  }

  if (results[1].status === 'fulfilled') agricultureInsights.value = results[1].value
  loadingLiveData.value = false
}

// ── Lifecycle ─────────────────────────────────────────────────────────────────
function handleResize() {
  if (!viewer || viewer.isDestroyed()) return
  viewer.resize()
  viewer.scene.requestRender()
}

onMounted(async () => {
  try {
    viewer = new Cesium.Viewer(cesiumContainer.value, {
      imageryProvider:      buildImageryProvider('satellite'),   // satellite by default
      terrainProvider:      new Cesium.EllipsoidTerrainProvider(),
      baseLayerPicker:      false,
      navigationHelpButton: false,
      homeButton:           false,
      sceneModePicker:      false,
      geocoder:             false,
      animation:            false,
      timeline:             false,
      fullscreenButton:     false,
      infoBox:              false,
      selectionIndicator:   false,
      skyBox:               false,
      skyAtmosphere:        false,
    })

    viewer.scene.backgroundColor              = Cesium.Color.fromCssColorString('#0c1a2e')
    viewer.scene.globe.baseColor              = Cesium.Color.fromCssColorString('#4a7c59')
    viewer.scene.globe.enableLighting         = false
    viewer.scene.globe.showGroundAtmosphere   = false
    viewer.scene.fog.enabled                  = false
    viewer.scene.globe.depthTestAgainstTerrain = false

    // Restrict zoom: prevent users from getting closer than 500 m above the surface
    viewer.scene.screenSpaceCameraController.minimumZoomDistance = 500

    // Start at Maharashtra state level — no globe view
    viewer.camera.setView({
      destination: Cesium.Cartesian3.fromDegrees(76.0, 19.5, 120000),
      orientation: { heading: 0, pitch: Cesium.Math.toRadians(-42), roll: 0 },
    })

    // Click → select house OR open cluster solution panel
    viewer.screenSpaceEventHandler.setInputAction((e) => {
      const picked = viewer.scene.pick(e.position)
      if (Cesium.defined(picked) && picked.id) {
        const entityId = picked.id.id

        // House entity → open detail panel
        const house = entityMap.get(entityId)
        if (house) {
          selectedHouse.value   = house
          selectedCluster.value = null
          return
        }

        // Cluster entity → open solution panel
        const cluster = clusterMap.get(entityId)
        if (cluster) {
          selectedCluster.value = cluster
          selectedHouse.value   = null
          return
        }
      }
      // Clicked empty space → clear both
      selectedHouse.value   = null
      selectedCluster.value = null
    }, Cesium.ScreenSpaceEventType.LEFT_CLICK)

    // Double-click → zoom
    viewer.screenSpaceEventHandler.setInputAction((e) => {
      const picked = viewer.scene.pick(e.position)
      if (Cesium.defined(picked) && picked.id) {
        const house = entityMap.get(picked.id.id)
        if (house) flyToHouse(house)
      }
    }, Cesium.ScreenSpaceEventType.LEFT_DOUBLE_CLICK)

    // Hover
    viewer.screenSpaceEventHandler.setInputAction((e) => {
      mouseX.value = e.endPosition.x + 16
      mouseY.value = e.endPosition.y + 12
      const picked = viewer.scene.pick(e.endPosition)
      hoveredHouse.value = (Cesium.defined(picked) && picked.id)
        ? (entityMap.get(picked.id.id) || null)
        : null
    }, Cesium.ScreenSpaceEventType.MOUSE_MOVE)

    setupZoomListener()

  } catch (err) {
    console.warn('Cesium init failed:', err)
  }

  loadData()
  setTimeout(handleResize, 60)
  setTimeout(handleResize, 300)
  window.addEventListener('resize', handleResize)
  // Close any open custom dropdown when clicking outside
  window.addEventListener('click', closeDropdowns)
})

onUnmounted(() => {
  clearRetryTimer()
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('click', closeDropdowns)
  if (viewer && !viewer.isDestroyed()) {
    viewer.camera.changed.removeEventListener(updateZoomVisibility)
    viewer.destroy()
  }
  viewer = null
})
</script>

<style scoped>
/* ═══════════════════════════════════════════════
   DESIGN TOKENS
   NOTE: Must live on .twin-page, NOT :root —
   Vue scoped styles break :root selectors.
═══════════════════════════════════════════════ */
.twin-page {
  /* Surfaces */
  --c-bg:        #f3f4f6;
  --c-surface:   #ffffff;
  --c-surface-2: #f9fafb;

  /* Borders — visible against white */
  --c-border:    #d1d5db;
  --c-border-2:  #e5e7eb;

  /* Text — strong contrast on white */
  --c-text:      #111827;
  --c-text2:     #374151;
  --c-text3:     #6b7280;

  /* Accent */
  --c-accent:    #16a34a;
  --c-accent-lt: #f0fdf4;

  /* Shadows */
  --c-shadow:    0 1px 3px rgba(0,0,0,0.08), 0 1px 2px rgba(0,0,0,0.05);
  --c-shadow-md: 0 4px 12px rgba(0,0,0,0.10), 0 2px 4px rgba(0,0,0,0.06);
  --c-shadow-lg: 0 8px 24px rgba(0,0,0,0.12), 0 3px 6px rgba(0,0,0,0.07);

  /* Radii */
  --radius:    8px;
  --radius-sm: 5px;

  /* ── Page shell ── */
  position: relative;
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background: #0c1a2e;
  font-family: system-ui, -apple-system, 'Segoe UI', sans-serif;
  font-size: 13px;
  color: var(--c-text);
}

.cesium-wrap {
  position: absolute;
  inset: 0;
}
.cesium-wrap :deep(.cesium-widget),
.cesium-wrap :deep(.cesium-widget canvas) {
  width: 100% !important;
  height: 100% !important;
}

/* ═══════════════════════════════════════════════
   TOP BAR
═══════════════════════════════════════════════ */
.topbar {
  position: absolute;
  top: 0; left: 0; right: 0;
  z-index: 200;
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0 1rem;
  height: 50px;
  background: var(--c-surface);
  border-bottom: 1px solid var(--c-border);
  box-shadow: var(--c-shadow);
}

/* Brand */
.topbar-brand {
  display: flex; align-items: center; gap: 0.45rem;
  flex-shrink: 0; white-space: nowrap;
}
.brand-dot {
  width: 9px; height: 9px; border-radius: 50%;
  background: #16a34a;
  box-shadow: 0 0 0 2px #bbf7d0;
  flex-shrink: 0;
}
.brand-name { font-weight: 700; font-size: 0.92rem; color: #111827; letter-spacing: -0.01em; }
.brand-sub  { font-size: 0.68rem; color: #6b7280; padding-left: 0.1rem; }

/* Filter bar */
.filter-bar {
  display: flex; align-items: center; gap: 0.5rem;
  flex: 1; min-width: 0;
}
.filter-group {
  display: flex; flex-direction: column; gap: 2px;
}
.filter-label {
  font-size: 0.58rem; text-transform: uppercase;
  letter-spacing: 0.07em; color: #374151; font-weight: 700;
  line-height: 1;
}

/* ── Custom Select — replaces native <select> to avoid OS dark-mode override ── */
.custom-select {
  position: relative;
  min-width: 110px;
  /* Clicking anywhere on .custom-select opens the dropdown; stop propagation
     at the wrapper level prevents the window click-outside handler from
     immediately closing it */
}

/* Trigger button — looks identical to the old .filter-select */
.cs-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.35rem;
  width: 100%;
  background: #ffffff;
  border: 1.5px solid #9ca3af;
  border-radius: var(--radius-sm);
  color: #111827;
  font-size: 0.76rem;
  font-weight: 500;
  padding: 0.28rem 0.55rem;
  cursor: pointer;
  outline: none;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
  transition: border-color 0.15s, box-shadow 0.15s;
  text-align: left;
  white-space: nowrap;
  /* No OS dark-mode interference — pure div, not a <select> */
}
.cs-trigger:hover:not(:disabled) {
  border-color: #6b7280;
  box-shadow: 0 1px 4px rgba(0,0,0,0.12);
}
.custom-select.open .cs-trigger {
  border-color: #16a34a;
  box-shadow: 0 0 0 3px rgba(22,163,74,0.15);
}
.custom-select.disabled .cs-trigger,
.cs-trigger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: #f9fafb;
  border-color: #e5e7eb;
}

.cs-value {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
}
.cs-arrow {
  font-size: 0.6rem;
  color: #6b7280;
  flex-shrink: 0;
  transition: transform 0.15s;
  line-height: 1;
}
.custom-select.open .cs-arrow {
  transform: rotate(180deg);
}

/* Dropdown panel */
.cs-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 100%;
  max-height: 220px;
  overflow-y: auto;
  background: #ffffff;
  border: 1.5px solid #d1d5db;
  border-radius: var(--radius);
  box-shadow: 0 8px 24px rgba(0,0,0,0.13), 0 3px 8px rgba(0,0,0,0.08);
  z-index: 600;           /* well above topbar z-index:200 */
  scrollbar-width: thin;
  scrollbar-color: #d1d5db transparent;
}
/* Right-aligned variant for "Color by" which sits near the right edge */
.cs-dropdown-right {
  left: auto;
  right: 0;
}

/* Individual option rows */
.cs-option {
  padding: 0.42rem 0.75rem;
  font-size: 0.76rem;
  color: #111827;
  cursor: pointer;
  transition: background 0.1s, color 0.1s;
  white-space: nowrap;
  background: #ffffff;   /* explicit — never inherits dark-mode */
}
.cs-option:first-child { border-radius: var(--radius) var(--radius) 0 0; }
.cs-option:last-child  { border-radius: 0 0 var(--radius) var(--radius); }
.cs-option:hover {
  background: #f0fdf4;
  color: #15803d;
}
.cs-option.selected {
  background: #dcfce7;
  color: #15803d;
  font-weight: 600;
}

.filter-arrow { color: #9ca3af; font-size: 0.9rem; flex-shrink: 0; font-weight: 600; }

.apply-btn {
  background: #16a34a; border: 1.5px solid #15803d;
  border-radius: var(--radius-sm);
  color: #ffffff; font-size: 0.72rem; font-weight: 700;
  padding: 0.3rem 0.85rem;
  cursor: pointer; white-space: nowrap; transition: all 0.15s; flex-shrink: 0;
  box-shadow: 0 1px 3px rgba(22,163,74,0.3);
}
.apply-btn:hover:not(:disabled) { background: #15803d; border-color: #14532d; }
.apply-btn:disabled {
  background: #d1fae5; border-color: #a7f3d0; color: #6ee7b7;
  cursor: not-allowed; box-shadow: none;
}

.reset-btn {
  background: #fff1f2; border: 1.5px solid #fca5a5;
  border-radius: var(--radius-sm);
  color: #dc2626; font-size: 0.72rem; font-weight: 600;
  padding: 0.3rem 0.7rem;
  cursor: pointer; white-space: nowrap; transition: all 0.15s; flex-shrink: 0;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}
.reset-btn:hover { background: #dc2626; color: #fff; border-color: #dc2626; }

/* Right controls */
.topbar-right {
  display: flex; align-items: center; gap: 0.5rem; flex-shrink: 0;
}
.ctrl-group { display: flex; flex-direction: column; gap: 2px; }

.ctrl-btn {
  background: #ffffff;
  border: 1.5px solid #d1d5db;
  border-radius: var(--radius-sm);
  color: #374151; font-size: 0.73rem; font-weight: 500;
  padding: 0.32rem 0.8rem;
  cursor: pointer; white-space: nowrap; transition: all 0.15s;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}
.ctrl-btn:hover          { border-color: #16a34a; color: #16a34a; background: #f0fdf4; }
.ctrl-btn.active         { border-color: #16a34a; background: #dcfce7; color: #15803d; font-weight: 600; }
.ctrl-btn.primary        { background: #16a34a; border-color: #15803d; color: #fff; font-weight: 600; box-shadow: 0 1px 3px rgba(22,163,74,0.3); }
.ctrl-btn.primary:hover  { background: #15803d; border-color: #14532d; }

/* ── Download PDF widget ── */
.dl-wrap {
  display: flex; flex-direction: column; align-items: flex-end; gap: 2px;
  flex-shrink: 0;
}
.dl-btn {
  background: #2563eb; color: #ffffff;
  border: 1.5px solid #1d4ed8;
  border-radius: var(--radius-sm);
  font-size: 0.73rem; font-weight: 600;
  padding: 0.32rem 0.9rem;
  cursor: pointer; white-space: nowrap;
  transition: background 0.15s, opacity 0.15s;
  box-shadow: 0 1px 3px rgba(37,99,235,0.22);
  line-height: 1;
}
.dl-btn:hover:not(:disabled) { background: #1d4ed8; border-color: #1e40af; }
.dl-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.dl-count {
  font-size: 0.6rem; color: #6b7280;
  font-variant-numeric: tabular-nums;
  text-align: right; line-height: 1;
}

/* ═══════════════════════════════════════════════
   STATS BAR
═══════════════════════════════════════════════ */
.stats-bar {
  position: absolute;
  top: 50px; left: 50%; transform: translateX(-50%);
  z-index: 100;
  display: flex; align-items: center; gap: 0.55rem;
  background: rgba(255, 255, 255, 0.96);
  border: 1.5px solid #d1d5db;
  border-top: none;
  border-radius: 0 0 var(--radius) var(--radius);
  padding: 0.25rem 1.1rem;
  font-size: 0.73rem; color: #374151;
  backdrop-filter: blur(8px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.10);
  white-space: nowrap;
}
.stat-item          { display: flex; align-items: center; gap: 0.3rem; }
.stat-item strong   { color: #111827; font-weight: 700; }
.stat-dot           { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.stat-sep           { color: #9ca3af; font-weight: 300; }
.stat-filter-note   { color: #6b7280; font-size: 0.66rem; }
.zoom-label         { color: #6b7280; font-style: italic; }

/* ═══════════════════════════════════════════════
   LOADING OVERLAY
═══════════════════════════════════════════════ */
.loading-overlay {
  position: absolute;
  inset: 0; z-index: 500;
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(4px);
  gap: 0.8rem;
}
.loading-spinner {
  width: 36px; height: 36px; border-radius: 50%;
  border: 3px solid #e2e8f0;
  border-top-color: #16a34a;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.loading-text { font-size: 0.84rem; color: #374151; font-weight: 500; }

/* ═══════════════════════════════════════════════
   LEFT SIDEBAR
═══════════════════════════════════════════════ */
.sidebar {
  position: absolute;
  left: 0.75rem;
  top: 70px;
  z-index: 100;
  width: 240px;
  max-height: calc(100vh - 82px);
  overflow-y: auto;
  overflow-x: visible;
  scrollbar-width: thin;
  scrollbar-color: var(--c-border) transparent;
  transition: width 0.22s ease;
}
.sidebar.collapsed { width: 20px; overflow: visible; }

.sidebar-toggle {
  position: absolute;
  right: -14px; top: 0;
  width: 26px; height: 26px;
  background: #ffffff;
  border: 1.5px solid #d1d5db;
  border-radius: 50%;
  color: #374151; font-size: 0.9rem; line-height: 1;
  cursor: pointer; z-index: 110;
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 2px 6px rgba(0,0,0,0.10);
  transition: all 0.15s;
}
.sidebar-toggle:hover { border-color: #16a34a; color: #16a34a; background: #f0fdf4; }

.sidebar-body {
  display: flex; flex-direction: column; gap: 0.55rem;
  padding-right: 18px; /* room for toggle btn */
}

/* ── Panel cards ── */
.panel-card {
  background: #ffffff;
  border: 1.5px solid #d1d5db;
  border-radius: var(--radius);
  padding: 0.8rem 0.85rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08), 0 1px 3px rgba(0,0,0,0.05);
}
.card-title {
  font-size: 0.62rem; font-weight: 800;
  text-transform: uppercase; letter-spacing: 0.1em;
  color: #374151;   /* strong — was #94a3b8 via broken var */
  margin-bottom: 0.6rem;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 0.4rem;
}
.card-title-sub {
  font-size: 0.56rem; font-weight: 400;
  text-transform: none; letter-spacing: 0;
  color: #9ca3af; margin-left: 0.4rem;
}
.empty-note { font-size: 0.73rem; color: #6b7280; line-height: 1.5; padding: 0.2rem 0; }

/* ── Legend ── */
.legend-item   { display: flex; align-items: center; gap: 0.55rem; margin-bottom: 0.38rem; }
.legend-text   { font-size: 0.7rem; color: #111827; font-weight: 500; }
.legend-note   { font-size: 0.62rem; color: #9ca3af; margin-top: 0.45rem; font-style: italic; }

/* ── Mini 3D House Icon ─────────────────────────────────────────────────────
   Mirrors the map's 3D block aesthetic: triangle roof (condition color) +
   rectangular wall block (sandstone).  --mh-roof CSS variable drives color.
   Two sizes: default (legend) and .mini-house-sm (issue pips, filter labels).
────────────────────────────────────────────────────────────────────────── */
.mini-house {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
  filter: drop-shadow(0 1px 1.5px rgba(0,0,0,0.28));
}
.mh-roof {
  width: 0;
  height: 0;
  border-left:  8px solid transparent;
  border-right: 8px solid transparent;
  border-bottom: 6px solid var(--mh-roof, #94a3b8);
  transition: border-bottom-color 0.2s;
}
.mh-wall {
  width: 12px;
  height: 8px;
  background: #c8a97e;   /* sandstone — matches map wall color */
  border: 1px solid rgba(0,0,0,0.18);
  box-shadow: inset 1px 0 0 rgba(255,255,255,0.18), 1px 1px 0 rgba(0,0,0,0.12);
}
/* Smaller variant for issue rows and problem filter labels */
.mini-house-sm .mh-roof {
  border-left-width:   6px;
  border-right-width:  6px;
  border-bottom-width: 5px;
}
.mini-house-sm .mh-wall {
  width: 9px;
  height: 6px;
}

/* ── Field Issues ── */
.issue-row {
  display: flex; align-items: center; gap: 0.45rem;
  padding: 0.32rem 0.35rem; border-radius: var(--radius-sm);
  cursor: pointer; transition: background 0.12s; margin-bottom: 0.35rem;
  border: 1px solid transparent;
}
.issue-row:hover  { background: #f9fafb; border-color: #e5e7eb; }
.issue-row.active { background: #f0fdf4; border-color: #bbf7d0; }
/* .issue-pip replaced by .mini-house-sm — see mini-house CSS block above */
.issue-body  { flex: 1; min-width: 0; }
.issue-top   { display: flex; align-items: baseline; gap: 0.3rem; margin-bottom: 0.2rem; }
.issue-name  { font-size: 0.7rem; color: #111827; font-weight: 600; flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.issue-count { font-size: 0.63rem; color: #6b7280; font-variant-numeric: tabular-nums; }
.issue-pct   { font-size: 0.7rem; font-weight: 700; font-variant-numeric: tabular-nums; }
/* Thicker, more visible progress bar */
.issue-track { height: 5px; background: #e5e7eb; border-radius: 3px; overflow: hidden; }
.issue-fill  { height: 100%; border-radius: 3px; transition: width 0.5s ease; }
.issue-chevron { font-size: 0.95rem; color: #9ca3af; transition: transform 0.2s; flex-shrink: 0; line-height: 1; }
.issue-chevron.open { transform: rotate(90deg); color: #374151; }

.issue-drawer {
  margin: 0 0 0.45rem 1.1rem;
  padding: 0.55rem 0.65rem;
  border-left: 3px solid;
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  background: #f9fafb;
  border-top: 1px solid #e5e7eb;
  border-right: 1px solid #e5e7eb;
  border-bottom: 1px solid #e5e7eb;
}
.drawer-cause, .drawer-solution {
  font-size: 0.72rem; color: #374151;
  line-height: 1.55; margin: 0 0 0.3rem;
}
.drawer-cause strong, .drawer-solution strong { color: #111827; }
.drawer-scheme {
  font-size: 0.68rem; font-weight: 700;
  padding: 0.2rem 0.5rem; border-radius: 4px;
  border: 1.5px solid currentColor;
  display: inline-block; background: #ffffff;
  margin-top: 0.15rem;
}

/* ── Agriculture donut charts ── */
.agri-chart { margin-bottom: 1rem; }
.agri-chart:last-child { margin-bottom: 0; }
.agri-label {
  font-size: 0.6rem; color: #374151; font-weight: 700;
  margin-bottom: 0.45rem; text-transform: uppercase; letter-spacing: 0.04em;
}
.pie-row {
  display: flex; align-items: center; gap: 0.75rem;
}
.pie-donut {
  width: 52px; height: 52px; border-radius: 50%; flex-shrink: 0;
  position: relative;
}
.pie-donut::after {
  content: ''; position: absolute;
  inset: 13px; border-radius: 50%;
  background: #ffffff;
  box-shadow: inset 0 1px 3px rgba(0,0,0,0.08);
}
.pie-legend { flex: 1; min-width: 0; }
.pie-item {
  display: flex; align-items: center; gap: 0.32rem;
  margin-bottom: 0.22rem;
}
.pie-item:last-child { margin-bottom: 0; }
.pie-dot  { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; box-shadow: 0 1px 2px rgba(0,0,0,0.15); }
.pie-name { font-size: 0.63rem; color: #374151; font-weight: 500; flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.pie-pct  { font-size: 0.65rem; color: #111827; font-variant-numeric: tabular-nums; font-weight: 700; flex-shrink: 0; }

/* ═══════════════════════════════════════════════
   HOVER TOOLTIP (rich data popup)
═══════════════════════════════════════════════ */
.hover-card {
  position: fixed; z-index: 300;
  background: #ffffff;
  border: 1.5px solid #e5e7eb;
  border-radius: 12px;
  padding: 0.65rem 0.8rem 0.5rem;
  box-shadow: 0 12px 32px rgba(0,0,0,0.16), 0 2px 8px rgba(0,0,0,0.08);
  pointer-events: none;
  min-width: 210px;
  max-width: 260px;
}
/* Header */
.hc-head  { display: flex; align-items: flex-start; justify-content: space-between; gap: 0.4rem; margin-bottom: 0.16rem; }
.hc-name  { font-size: 0.82rem; font-weight: 700; color: #111827; line-height: 1.25; flex: 1; }
.hc-badge {
  font-size: 0.58rem; font-weight: 700; letter-spacing: 0.03em;
  padding: 2px 6px; border-radius: 999px; border: 1px solid; white-space: nowrap;
  flex-shrink: 0;
}
.hc-loc { font-size: 0.63rem; color: #6b7280; margin-bottom: 0.42rem; }
/* 2×2 status grid */
.hc-grid {
  display: grid; grid-template-columns: 1fr 1fr;
  gap: 0.28rem 0.5rem; margin-bottom: 0.35rem;
}
.hc-cell { display: flex; align-items: center; gap: 0.28rem; min-width: 0; }
.hc-dot  { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; box-shadow: 0 1px 2px rgba(0,0,0,0.22); }
.hc-ck   { font-size: 0.59rem; color: #9ca3af; white-space: nowrap; }
.hc-cv   { font-size: 0.65rem; font-weight: 600; color: #1f2937; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
/* Crops row */
.hc-crops { display: flex; align-items: baseline; gap: 0.3rem; font-size: 0.65rem; margin-bottom: 0.32rem; }
.hc-crops .hc-ck { color: #9ca3af; }
.hc-crops .hc-cv { color: #1f2937; font-weight: 600; }
/* Footer hint */
.hc-hint { font-size: 0.59rem; color: #9ca3af; border-top: 1px solid #f3f4f6; padding-top: 0.28rem; font-style: italic; }

/* ═══════════════════════════════════════════════
   DETAIL PANEL (right slide-in)
═══════════════════════════════════════════════ */
/* ═══════════════════════════════════════════════
   DETAIL PANEL (farmer click popup)
═══════════════════════════════════════════════ */
.detail-panel {
  position: absolute;
  right: 0.75rem; top: 60px;
  width: 320px;
  max-height: calc(100vh - 72px);
  overflow-y: auto;
  z-index: 100;
  background: #ffffff;
  border: 1.5px solid #e2e8f0;
  border-radius: 12px;
  box-shadow: 0 12px 40px rgba(0,0,0,0.16), 0 4px 12px rgba(0,0,0,0.08);
  scrollbar-width: thin;
  scrollbar-color: #cbd5e1 transparent;
}

/* ── Header ── */
.detail-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: 0.5rem; padding: 1rem 1rem 0.8rem;
  border-bottom: 1px solid #e2e8f0;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 12px 12px 0 0;
}
.detail-header-info { flex: 1; min-width: 0; }
.detail-badge {
  display: inline-block; padding: 0.18rem 0.55rem;
  border-radius: 20px; border: 1.5px solid;
  font-size: 0.62rem; font-weight: 700; letter-spacing: 0.04em;
  margin-bottom: 0.38rem; text-transform: uppercase;
}
.detail-name {
  font-size: 1rem; font-weight: 800; color: #0f172a;
  line-height: 1.25; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.detail-sub {
  display: flex; align-items: center; flex-wrap: wrap; gap: 0.3rem;
  font-size: 0.66rem; color: #64748b; margin-top: 0.28rem; font-weight: 500;
}
.detail-id-chip {
  background: #1e293b; color: #f8fafc;
  border-radius: 4px; padding: 0.08rem 0.38rem;
  font-size: 0.6rem; font-weight: 700; letter-spacing: 0.04em;
}
.detail-close {
  background: #f1f5f9; border: 1px solid #e2e8f0;
  border-radius: 50%; color: #64748b;
  font-size: 1.1rem; line-height: 1; cursor: pointer;
  width: 26px; height: 26px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; transition: all 0.15s;
}
.detail-close:hover { background: #ef4444; border-color: #ef4444; color: #fff; }

/* ── Zoom button ── */
.focus-btn {
  display: block; width: calc(100% - 2rem);
  margin: 0.8rem 1rem 0;
  background: #f0fdf4; border: 1.5px solid #86efac;
  border-radius: 8px;
  color: #15803d; font-size: 0.73rem; font-weight: 600;
  padding: 0.42rem 0.7rem; cursor: pointer; text-align: center;
  transition: all 0.15s; box-shadow: 0 1px 3px rgba(22,163,74,0.12);
}
.focus-btn:hover { background: #16a34a; border-color: #15803d; color: #fff; }

/* ── Section labels ── */
.dp-section-label {
  display: flex; align-items: center; gap: 0.4rem;
  font-size: 0.6rem; text-transform: uppercase; letter-spacing: 0.09em;
  color: #475569; font-weight: 800;
  padding: 0.85rem 1rem 0.35rem;
  border-top: 1px solid #f1f5f9;
}
.dp-section-icon { font-size: 0.85rem; }

/* ── Big stat row (Land) ── */
.dp-stat-row {
  display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem;
  padding: 0 1rem;
}
.dp-stat {
  background: #f8fafc; border: 1.5px solid #e2e8f0;
  border-radius: 8px; padding: 0.6rem 0.75rem; text-align: center;
}
.dp-stat-val {
  font-size: 1.25rem; font-weight: 800; color: #0f172a; line-height: 1.1;
}
.dp-stat-val small { font-size: 0.65rem; color: #64748b; font-weight: 600; }
.dp-stat-key { font-size: 0.57rem; text-transform: uppercase; letter-spacing: 0.06em; color: #94a3b8; font-weight: 600; margin-top: 0.18rem; }

/* ── Crop season chips ── */
.dp-chip-row {
  display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem;
  padding: 0.5rem 1rem 0;
}
.dp-chip-block { display: flex; flex-direction: column; gap: 0.22rem; }
.dp-chip-label { font-size: 0.57rem; text-transform: uppercase; letter-spacing: 0.06em; color: #94a3b8; font-weight: 600; }
.dp-chip {
  padding: 0.3rem 0.55rem; border-radius: 6px;
  font-size: 0.72rem; font-weight: 600; text-align: center;
}
.dp-chip-kharif { background: #fef9c3; color: #854d0e; border: 1px solid #fde68a; }
.dp-chip-rabi   { background: #dbeafe; color: #1e40af; border: 1px solid #bfdbfe; }

/* ── Full-width field rows ── */
.dp-field-row {
  display: flex; align-items: center; gap: 0.55rem;
  padding: 0.55rem 1rem;
  border-bottom: 1px solid #f8fafc;
}
.dp-field-icon { font-size: 0.85rem; flex-shrink: 0; width: 1.2rem; text-align: center; }
.dp-field-key  { font-size: 0.68rem; color: #64748b; font-weight: 600; flex: 1; }
.dp-field-val  { font-size: 0.76rem; color: #0f172a; font-weight: 700; text-align: right; max-width: 55%; }
.dp-ok   { color: #15803d !important; }
.dp-warn { color: #b45309 !important; }

/* ── Advisory cards ── */
.detail-issues { padding: 0 1rem 1rem; }
.advisory-card {
  border-left: 3px solid; border-radius: 0 8px 8px 0;
  background: #f8fafc;
  border-top: 1px solid #e2e8f0;
  border-right: 1px solid #e2e8f0;
  border-bottom: 1px solid #e2e8f0;
  padding: 0.65rem 0.75rem;
  margin-bottom: 0.55rem;
}
.advisory-title { font-size: 0.8rem; font-weight: 700; margin-bottom: 0.42rem; }
.advisory-row   { display: flex; gap: 0.48rem; margin-bottom: 0.32rem; align-items: flex-start; }
.advisory-tag {
  font-size: 0.59rem; text-transform: uppercase; letter-spacing: 0.05em;
  padding: 0.14rem 0.4rem; border-radius: 3px;
  font-weight: 800; flex-shrink: 0; margin-top: 0.04rem;
}
.advisory-tag.cause    { background: #fef2f2; color: #dc2626; border: 1px solid #fca5a5; }
.advisory-tag.solution { background: #f0fdf4; color: #15803d; border: 1px solid #86efac; }
.advisory-text  { font-size: 0.73rem; color: #374151; line-height: 1.55; }
.advisory-scheme {
  font-size: 0.68rem; font-weight: 700;
  padding: 0.2rem 0.5rem; border-radius: 4px;
  border: 1.5px solid currentColor; display: inline-block;
  background: #ffffff; margin-top: 0.22rem;
}

.all-good {
  display: flex; align-items: center; gap: 0.45rem;
  font-size: 0.73rem; color: #15803d; font-weight: 600;
  padding: 0.55rem 0.9rem 0.9rem;
  background: #f0fdf4; margin: 0.6rem 1rem 1rem;
  border-radius: 8px; border: 1px solid #bbf7d0;
}

/* ═══════════════════════════════════════════════
   TRANSITIONS
═══════════════════════════════════════════════ */
.slide-enter-active { transition: all 0.22s ease-out; }
.slide-leave-active { transition: all 0.15s ease-in; }
.slide-enter-from, .slide-leave-to { opacity: 0; transform: translateX(16px); }

.drawer-enter-active { transition: all 0.18s ease-out; }
.drawer-leave-active { transition: all 0.1s ease-in; }
.drawer-enter-from, .drawer-leave-to { opacity: 0; transform: translateY(-4px); }

/* ═══════════════════════════════════════════════
   CLUSTER SOLUTION PANEL
═══════════════════════════════════════════════ */
.cluster-panel {
  position: absolute;
  right: 0.75rem;
  top: 60px;
  width: 292px;
  max-height: calc(100vh - 72px);
  overflow-y: auto;
  z-index: 100;
  background: #ffffff;
  border: 1.5px solid #fca5a5;
  border-radius: var(--radius);
  box-shadow: 0 8px 28px rgba(239,68,68,0.14), 0 3px 8px rgba(0,0,0,0.07);
  scrollbar-width: thin;
  scrollbar-color: #fca5a5 transparent;
}

.cluster-close {
  position: absolute;
  top: 0.7rem;
  right: 0.7rem;
}

.cluster-header {
  padding: 1rem 1rem 0.7rem;
  background: linear-gradient(135deg, #fef2f2 0%, #fff7f7 100%);
  border-bottom: 1.5px solid #fecaca;
  border-radius: var(--radius) var(--radius) 0 0;
}

.cluster-badge {
  display: inline-block;
  background: #ef4444;
  color: #fff;
  font-size: 0.67rem;
  font-weight: 700;
  padding: 0.2rem 0.6rem;
  border-radius: 20px;
  letter-spacing: 0.04em;
  margin-bottom: 0.42rem;
}

.cluster-count {
  font-size: 0.82rem;
  color: #374151;
  line-height: 1.3;
}
.cluster-count strong { color: #ef4444; font-size: 1.1rem; }

.cluster-section-title {
  font-size: 0.62rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #374151;
  padding: 0.75rem 1rem 0.3rem;
}

/* Individual problem card */
.cp-card {
  margin: 0 0.75rem 0.65rem;
  padding: 0.65rem 0.7rem;
  background: #fafafa;
  border: 1.5px solid #e5e7eb;
  border-radius: var(--radius-sm);
  border-left: 3px solid #ef4444;
}

.cp-top {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  margin-bottom: 0.45rem;
}
.cp-emoji { font-size: 1.25rem; flex-shrink: 0; line-height: 1.2; }
.cp-info  { flex: 1; min-width: 0; }
.cp-label { display: block; font-size: 0.8rem; font-weight: 700; color: #111827; }
.cp-stat  { font-size: 0.66rem; color: #6b7280; margin-top: 0.1rem; display: block; }

.cp-bar-track {
  height: 5px;
  background: #f3f4f6;
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 0.52rem;
}
.cp-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #ef4444, #f87171);
  border-radius: 3px;
  transition: width 0.5s ease;
}

.cp-action {
  font-size: 0.72rem;
  font-weight: 700;
  color: #dc2626;
  margin-bottom: 0.3rem;
}
.cp-solution {
  font-size: 0.72rem;
  color: #374151;
  line-height: 1.55;
  margin: 0 0 0.38rem;
}
.cp-scheme {
  font-size: 0.65rem;
  font-weight: 700;
  color: #6b7280;
  padding: 0.18rem 0.48rem;
  border: 1.5px solid #d1d5db;
  border-radius: 4px;
  display: inline-block;
  background: #ffffff;
}

.cluster-ok {
  font-size: 0.73rem;
  color: #15803d;
  padding: 0.75rem 1rem 1rem;
  font-weight: 500;
}

/* ═══════════════════════════════════════════════
   PROBLEM FILTER PANEL
═══════════════════════════════════════════════ */
.pf-item {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.3rem 0.25rem;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background 0.1s;
  margin-bottom: 0.18rem;
}
.pf-item:hover { background: #f9fafb; }

.pf-check {
  width: 13px; height: 13px;
  accent-color: #ef4444;
  cursor: pointer;
  flex-shrink: 0;
}
.pf-pip {
  width: 9px; height: 9px;
  border-radius: 50%;
  flex-shrink: 0;
  box-shadow: 0 1px 2px rgba(0,0,0,0.2);
}
.pf-label {
  flex: 1;
  font-size: 0.72rem;
  color: #111827;
  font-weight: 500;
}
.pf-count {
  font-size: 0.65rem;
  color: #6b7280;
  font-variant-numeric: tabular-nums;
  background: #f3f4f6;
  padding: 0.06rem 0.4rem;
  border-radius: 10px;
  flex-shrink: 0;
}

.pf-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 0.55rem;
  padding: 0.35rem 0.5rem;
  background: #fef2f2;
  border: 1px solid #fca5a5;
  border-radius: var(--radius-sm);
  font-size: 0.7rem;
  color: #dc2626;
}
.pf-summary strong { color: #dc2626; }

.pf-clear-btn {
  background: #ef4444;
  color: #fff;
  border: none;
  border-radius: 4px;
  font-size: 0.64rem;
  font-weight: 600;
  padding: 0.18rem 0.5rem;
  cursor: pointer;
  transition: background 0.15s;
}
.pf-clear-btn:hover { background: #dc2626; }

.pf-hint {
  font-size: 0.64rem;
  color: #9ca3af;
  line-height: 1.5;
  margin-top: 0.45rem;
  font-style: italic;
}


/* ═══════════════════════════════════════════════
   RESPONSIVE
═══════════════════════════════════════════════ */
@media (max-width: 700px) {
  .sidebar       { display: none; }
  .filter-bar    { display: none; }
  .detail-panel  { width: calc(100vw - 1.5rem); right: 0.75rem; }
  .topbar        { gap: 0.5rem; }
  .brand-sub     { display: none; }
}
</style>
