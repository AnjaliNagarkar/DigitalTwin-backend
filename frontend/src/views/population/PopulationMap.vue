<template>
  <div class="map-page">
    <header class="map-header">
      <div class="map-title-area">
        <h1 class="page-title">Geo-Intelligence Map</h1>
        <p class="page-subtitle">
          {{ summary.total_households.toLocaleString() }} households plotted from survey database
        </p>
      </div>

      <div class="map-controls">
        <div class="view-toggle">
          <button class="toggle-btn" :class="{ active: viewMode === 'points' }" @click="setViewMode('points')">
            Households
          </button>
          <button class="toggle-btn" :class="{ active: viewMode === 'villages' }" @click="setViewMode('villages')">
            Villages
          </button>
        </div>

        <div class="map-control-group">
          <label class="control-label">District</label>
          <div class="custom-select" :class="{ open: openDropdown === 'district' }" @click.stop="toggleDropdown('district')">
            <button class="cs-trigger" type="button">
              <span class="cs-value">{{ selectedDistrictLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'district'" @click.stop>
              <div class="cs-option" :class="{ selected: !selectedDistrict }" @click="selectDistrict('')">All</div>
              <div
                class="cs-option"
                v-for="district in districtOptions"
                :key="district.id"
                :class="{ selected: String(selectedDistrict) === String(district.id) }"
                @click="selectDistrict(district.id)"
              >
                {{ district.name }}
              </div>
            </div>
          </div>
        </div>

        <div class="map-control-group">
          <label class="control-label">Taluka</label>
          <div
            class="custom-select"
            :class="{ open: openDropdown === 'taluka', disabled: !talukaOptions.length }"
            @click.stop="talukaOptions.length && toggleDropdown('taluka')"
          >
            <button class="cs-trigger" type="button" :disabled="!talukaOptions.length">
              <span class="cs-value">{{ selectedTalukaLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'taluka'" @click.stop>
              <div class="cs-option" :class="{ selected: !selectedTaluka }" @click="selectTaluka('')">All</div>
              <div
                class="cs-option"
                v-for="taluka in talukaOptions"
                :key="taluka.id"
                :class="{ selected: String(selectedTaluka) === String(taluka.id) }"
                @click="selectTaluka(taluka.id)"
              >
                {{ taluka.name }}
              </div>
            </div>
          </div>
        </div>

        <div class="map-control-group">
          <label class="control-label">Village</label>
          <div
            class="custom-select"
            :class="{ open: openDropdown === 'village', disabled: !villageOptions.length }"
            @click.stop="villageOptions.length && toggleDropdown('village')"
          >
            <button class="cs-trigger" type="button" :disabled="!villageOptions.length">
              <span class="cs-value">{{ selectedVillageLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'village'" @click.stop>
              <div class="cs-option" :class="{ selected: !selectedVillage }" @click="selectVillage('')">All</div>
              <div
                class="cs-option"
                v-for="village in villageOptions"
                :key="village.id"
                :class="{ selected: String(selectedVillage) === String(village.id) }"
                @click="selectVillage(village.id)"
              >
                {{ village.name }}
              </div>
            </div>
          </div>
        </div>

        <div class="map-control-group">
          <button class="apply-btn" @click="applyFilters">Apply</button>
          <button class="reset-btn" @click="resetFilters">Reset</button>
        </div>

        <div class="map-control-group" v-if="viewMode === 'points'">
          <label class="control-label">Color by</label>
          <div class="custom-select cs-align-right" :class="{ open: openDropdown === 'colorMode' }" @click.stop="toggleDropdown('colorMode')">
            <button class="cs-trigger" type="button">
              <span class="cs-value">{{ selectedColorModeLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown cs-dropdown-right" v-show="openDropdown === 'colorMode'" @click.stop>
              <div class="cs-option" :class="{ selected: colorMode === 'population_density' }" @click="selectColorMode('population_density')">Population Density</div>
              <div class="cs-option" :class="{ selected: colorMode === 'bpl_status' }" @click="selectColorMode('bpl_status')">BPL Status</div>
              <div class="cs-option" :class="{ selected: colorMode === 'divyang_presence' }" @click="selectColorMode('divyang_presence')">Divyang Presence</div>
              <div class="cs-option" :class="{ selected: colorMode === 'employment' }" @click="selectColorMode('employment')">Employment Status</div>
            </div>
          </div>
        </div>

        <label class="gps-toggle" v-if="viewMode === 'points'">
          <input type="checkbox" v-model="showLocationIssues" />
          <span>Show GPS Issues</span>
        </label>

        <div class="map-legend">
          <template v-if="viewMode === 'villages'">
            <div class="legend-item"><span class="legend-dot" style="background:#10b981;"></span>High coverage</div>
            <div class="legend-item"><span class="legend-dot" style="background:#f59e0b;"></span>Medium</div>
            <div class="legend-item"><span class="legend-dot" style="background:#ef4444;"></span>Low coverage</div>
          </template>
          <template v-else>
            <div class="legend-item" v-for="leg in headerLegend" :key="leg.label">
              <span class="legend-dot" :style="{ background: leg.color }"></span>{{ leg.label }}
            </div>
          </template>
        </div>

      </div>
    </header>

    <section class="map-shell">
      <div v-if="!loading && !markers.length" class="empty-state">
        No live population map data returned from the database API.
      </div>

      <div class="map-content" ref="mapContentRef">
        <div class="map-stage">
          <div class="map-container" ref="mapContainer"></div>

          <div class="map-floating-controls">
            <button
              class="analytics-toggle"
              type="button"
              @click="analyticsPanelOpen = !analyticsPanelOpen"
              :aria-expanded="analyticsPanelOpen"
            >
              {{ analyticsPanelOpen ? 'Hide Analytics' : 'View Analytics' }}
            </button>
            <button
              class="fullscreen-toggle"
              type="button"
              @click="toggleFullscreen"
              :aria-pressed="isFullscreen"
            >
              {{ isFullscreen ? 'Exit Fullscreen' : 'Fullscreen' }}
            </button>
          </div>

        <transition name="slide">
          <aside v-if="selectedMarker && viewMode === 'points'" class="detail-panel">
            <button class="panel-close" @click="closeSelectedMarker">×</button>
            <h3 class="panel-title">Head: {{ selectedMarker.head_name || 'N/A' }}</h3>
            <div class="panel-subtitle">House No: {{ selectedMarker.house_no || 'N/A' }}</div>
            <div class="panel-id">
              Members: {{ Number(selectedMarker.total_members || 0).toLocaleString() }}
            </div>
            <div class="panel-grid">
              <div class="panel-stat" v-for="item in detailStats" :key="item.label">
                <div class="panel-stat-label">{{ item.label }}</div>
                <div class="panel-stat-value" :style="item.style || {}">{{ item.value }}</div>
              </div>
            </div>

            <div v-if="showLocationIssues" class="gps-detail-card">
              <div class="gps-detail-title">Location Quality</div>
              <div class="gps-detail-row">
                <span class="gps-detail-icon" :style="{ color: getLocationColor(selectedMarker.location_quality) }">
                  {{ getLocationIcon(selectedMarker.location_quality) }}
                </span>
                <div class="gps-detail-body">
                  <div class="gps-detail-label" :style="{ color: getLocationColor(selectedMarker.location_quality) }">
                    {{ getLocationLabel(selectedMarker.location_quality) }}
                  </div>
                  <div class="gps-detail-reason">Reason: {{ selectedMarker.location_reason || 'N/A' }}</div>
                  <div class="gps-detail-coords">
                    Coordinates: {{ Number(selectedMarker.lat || 0).toFixed(5) }}, {{ Number(selectedMarker.lng || 0).toFixed(5) }}
                  </div>
                </div>
              </div>

              <div
                v-if="Number(selectedMarker.duplicate_count || 0) >= 6 && getDuplicateHousePreview(selectedMarker).length"
                class="gps-dup-card"
              >
                <div class="gps-dup-title">
                  Same Coordinates: {{ Number(selectedMarker.duplicate_count || 0) }} houses
                </div>

                <div v-if="!isExpanded(selectedMarker)">
                  <div v-for="houseNo in getDuplicateHousePreview(selectedMarker)" :key="houseNo" class="gps-dup-item">
                    House No: {{ houseNo }}
                  </div>

                  <button
                    v-if="getRemainingDuplicateHouseCount(selectedMarker) > 0"
                    type="button"
                    class="gps-dup-toggle"
                    @click="toggleDuplicateListForMarker(selectedMarker)"
                  >
                    +{{ getRemainingDuplicateHouseCount(selectedMarker) }} more houses
                  </button>
                </div>

                <div v-else class="duplicate-list-scroll">
                  <div v-for="houseNo in getDuplicateHouseList(selectedMarker)" :key="houseNo" class="gps-dup-item">
                    House No: {{ houseNo }}
                  </div>

                  <button type="button" class="gps-dup-toggle" @click="toggleDuplicateListForMarker(selectedMarker)">
                    Show less
                  </button>
                </div>
              </div>
            </div>
          </aside>
        </transition>

        <transition name="slide">
          <aside v-if="selectedCluster" class="detail-panel village-panel">
            <button class="panel-close" @click="clearClusterSelection">×</button>
            <div class="village-badge">{{ selectedCluster.level }}</div>
            <h3 class="panel-title">{{ selectedCluster.name }}</h3>
            <div class="panel-id">{{ selectedCluster.households.toLocaleString() }} households covered</div>

            <div class="village-stats">
              <div class="vstat" :class="issueClass(selectedCluster.bpl_percent)">
                <div class="vstat-val">{{ selectedCluster.bpl_percent }}%</div>
                <div class="vstat-label">BPL Families</div>
              </div>
              <div class="vstat" :class="issueClass(selectedCluster.literacy_percent, null, true)">
                <div class="vstat-val">{{ selectedCluster.literacy_percent }}%</div>
                <div class="vstat-label">Literacy</div>
              </div>
              <div class="vstat" :class="issueClass(selectedCluster.working_percent, null, true)">
                <div class="vstat-val">{{ selectedCluster.working_percent }}%</div>
                <div class="vstat-label">Working Population</div>
              </div>
              <div class="vstat" :class="issueClass(selectedCluster.divyang_percent)">
                <div class="vstat-val">{{ selectedCluster.divyang_percent }}%</div>
                <div class="vstat-label">Divyang Population</div>
              </div>
            </div>

            <div class="village-bar-section">
              <div class="vbar-row" v-for="item in clusterIssues" :key="item.label">
                <span class="vbar-label">{{ item.label }}</span>
                <div class="vbar-track">
                  <div class="vbar-fill" :style="{ width: item.pct + '%', background: item.color }"></div>
                </div>
                <span class="vbar-pct">{{ item.pct }}%</span>
              </div>
            </div>

            <div class="panel-coords">
              {{ selectedCluster.latitude.toFixed(5) }}, {{ selectedCluster.longitude.toFixed(5) }}
            </div>
          </aside>
        </transition>
        </div>

        <transition name="analytics-panel-slide">
          <aside v-if="analyticsPanelOpen" class="analytics-panel" aria-label="Map analytics">
            <div class="analytics-panel-head">
              <h2 class="analytics-panel-title">Map Analytics</h2>
              <button class="analytics-close" type="button" @click="analyticsPanelOpen = false" aria-label="Close analytics">×</button>
            </div>

            <div class="analytics-scroll" v-if="analyticsCards.length">
              <article class="analytics-card" v-for="card in analyticsCards" :key="card.title">
                <div class="analytics-card-head">
                  <div>
                    <h3 class="analytics-title">{{ card.title }}</h3>
                    <p class="analytics-subtitle">{{ card.subtitle }}</p>
                  </div>
                  <div class="analytics-total">{{ card.totalLabel }}</div>
                </div>
                <div class="chart-layout">
                  <div class="donut" :style="pieStyle(card.segments)">
                    <div class="donut-hole">
                      <div class="donut-label">{{ card.centerLabel }}</div>
                      <div class="donut-value">{{ card.centerValue }}</div>
                    </div>
                  </div>
                  <div class="legend-list">
                    <div class="legend-row" v-for="segment in card.segments" :key="segment.label">
                      <span class="legend-dot" :style="{ background: segment.color }"></span>
                      <span class="legend-name">{{ segment.label }}</span>
                      <span class="legend-value">{{ segment.value.toLocaleString() }}</span>
                    </div>
                  </div>
                </div>
              </article>
            </div>
          </aside>
        </transition>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import L from 'leaflet'
import { getLocationOptions } from '../../api/index.js'
import { getPopulationMapData, getPopulationMapInsights, getPopulationMapSummary } from './api.js'

const mapContainer = ref(null)
const mapContentRef = ref(null)
const loading = ref(true)
const markers = ref([])
const summary = ref({ total_households: 0 })
const insights = ref({
  bpl_distribution: { bpl: 0, non_bpl: 0, total_households: 0 },
  education_status: { literate: 0, illiterate: 0, students: 0, dropouts: 0 },
  working_vs_dependent: { working: 0, dependent: 0, total_population: 0 },
})

const districtOptions = ref([])
const talukaOptions = ref([])
const villageOptions = ref([])

const selectedDistrict = ref('')
const selectedTaluka = ref('')
const selectedVillage = ref('')
const selectedMarker = ref(null)
const selectedCluster = ref(null)
const viewMode = ref('points')
const colorMode = ref('population_density')
const showLocationIssues = ref(false)
const expandedDuplicates = ref({})
const analyticsPanelOpen = ref(false)
const isFullscreen = ref(false)
const openDropdown = ref(null)

let map = null
let markerLayer = null
let clusterLayer = null
let requestToken = 0

const COLOR_MODE_LABELS = {
  population_density: 'Population Density',
  bpl_status: 'BPL Status',
  divyang_presence: 'Divyang Presence',
  employment: 'Employment Status',
}

const selectedDistrictLabel = computed(() => {
  if (!selectedDistrict.value) return 'All'
  return districtOptions.value.find((item) => String(item.id) === String(selectedDistrict.value))?.name || 'All'
})

const selectedTalukaLabel = computed(() => {
  if (!selectedTaluka.value) return 'All'
  return talukaOptions.value.find((item) => String(item.id) === String(selectedTaluka.value))?.name || 'All'
})

const selectedVillageLabel = computed(() => {
  if (!selectedVillage.value) return 'All'
  return villageOptions.value.find((item) => String(item.id) === String(selectedVillage.value))?.name || 'All'
})

const selectedColorModeLabel = computed(() => COLOR_MODE_LABELS[colorMode.value] || 'Population Density')

function toggleDropdown(name) {
  openDropdown.value = openDropdown.value === name ? null : name
}

function closeDropdowns() {
  openDropdown.value = null
}

function selectDistrict(id) {
  selectedDistrict.value = id
  closeDropdowns()
}

function selectTaluka(id) {
  selectedTaluka.value = id
  closeDropdowns()
}

function selectVillage(id) {
  selectedVillage.value = id
  closeDropdowns()
}

function selectColorMode(mode) {
  colorMode.value = mode
  closeDropdowns()
}

function handleMapResize() {
  if (map) {
    map.invalidateSize()
  }
}

function handleFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement
  handleMapResize()
}

async function toggleFullscreen() {
  const target = mapContentRef.value
  if (!target) return

  try {
    if (document.fullscreenElement) {
      await document.exitFullscreen()
      return
    }
    await target.requestFullscreen()
  } catch (error) {
    console.warn('Fullscreen unavailable:', error?.message || error)
  }
}

function pieStyle(segments) {
  const total = segments.reduce((sum, segment) => sum + segment.value, 0)
  if (!total) return { background: 'conic-gradient(#d1d5db 0deg 360deg)' }

  let cursor = 0
  const stops = segments.map((segment) => {
    const start = cursor
    cursor += (segment.value / total) * 360
    return `${segment.color} ${start}deg ${cursor}deg`
  })

  return { background: `conic-gradient(${stops.join(', ')})` }
}

function escapeHtml(value) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function normalizeText(value) {
  return String(value ?? '').trim().toLowerCase()
}

function getLocationLabel(q) {
  const quality = normalizeText(q)
  if (quality === 'valid') return 'Valid GPS location'
  if (quality === 'approximate') return 'Approximate GPS'
  if (quality === 'suspicious') return 'Suspicious location'
  if (quality === 'missing') return 'Missing GPS'
  return 'Unknown'
}

function getLocationColor(q) {
  const quality = normalizeText(q)
  if (quality === 'valid') return '#22c55e'
  if (quality === 'approximate') return '#f59e0b'
  if (quality === 'suspicious') return '#ef4444'
  if (quality === 'missing') return '#6b7280'
  return '#3b82f6'
}

function getLocationIcon(q) {
  const quality = normalizeText(q)
  if (quality === 'valid') return '🟢'
  if (quality === 'approximate') return '🟡'
  if (quality === 'suspicious') return '🔴'
  if (quality === 'missing') return '⚫'
  return '📍'
}

function getDuplicateHousePreview(marker) {
  const list = Array.isArray(marker?.duplicate_houses) ? marker.duplicate_houses : []
  return list.slice(0, 5)
}

function getDuplicateHouseList(marker) {
  const list = Array.isArray(marker?.duplicate_houses) ? marker.duplicate_houses : []
  return isExpanded(marker) ? list : list.slice(0, 5)
}

function getRemainingDuplicateHouseCount(marker) {
  const list = Array.isArray(marker?.duplicate_houses) ? marker.duplicate_houses : []
  if (isExpanded(marker)) return 0
  return Math.max(list.length - 5, 0)
}

function getMarkerKey(marker) {
  return `${marker?.lat}_${marker?.lng}`
}

function isExpanded(marker) {
  return Boolean(expandedDuplicates.value[getMarkerKey(marker)])
}

function toggleDuplicateListForMarker(marker) {
  const key = getMarkerKey(marker)
  expandedDuplicates.value[key] = !expandedDuplicates.value[key]
}

function closeSelectedMarker() {
  selectedMarker.value = null
}

function parseIncomeValue(value) {
  const numeric = Number(String(value ?? '').replace(/[^0-9.-]/g, ''))
  return Number.isFinite(numeric) ? numeric : null
}

function getBplStatusLabel(house) {
  const category = normalizeText(house.FAMILY_BELONG_BPL_CATEGORY)
  if (category) {
    if (category.includes('non-bpl') || category === 'no' || category === 'apl' || category.includes('above poverty')) {
      return 'Non-BPL'
    }
    if (category.includes('bpl') || category === 'yes') {
      return 'BPL'
    }
  }

  const rationCardType = normalizeText(house.RATION_CARD_TYPE)
  if (rationCardType.includes('bpl') || rationCardType.includes('aay') || rationCardType.includes('antyodaya')) {
    return 'BPL'
  }

  const annualIncome = parseIncomeValue(house.ANNUAL_INCOME)
  if (annualIncome !== null && annualIncome < 100000) {
    return 'BPL'
  }

  return 'Non-BPL'
}

function getBplColor(house) {
  return getBplStatusLabel(house) === 'BPL' ? '#e53935' : '#2e7d32'
}

function hasDivyangPresence(house) {
  return Number(house.has_disability || 0) === 1
}

function getDivyangStatusLabel(house) {
  return hasDivyangPresence(house) ? 'Yes' : 'No'
}

function getDivyangColor(house) {
  return hasDivyangPresence(house) ? '#7b1fa2' : '#2e7d32'
}

function getWorkingMembers(house) {
  return Number(house.working_members || 0)
}

function getOccupationList(house) {
  if (Array.isArray(house.occupation_list_array) && house.occupation_list_array.length) {
    return house.occupation_list_array
  }
  if (typeof house.occupation_list === 'string' && house.occupation_list.trim()) {
    return house.occupation_list
      .split('|')
      .map((item) => item.trim())
      .filter(Boolean)
  }
  return []
}

function getEmploymentColor(house) {
  return getWorkingMembers(house) > 0 ? '#2e7d32' : '#f57c00'
}

function getMarkerColor(house) {
  if (showLocationIssues.value) {
    const quality = normalizeText(house.location_quality)
    if (quality === 'valid') return '#22c55e'
    if (quality === 'approximate') return '#f59e0b'
    if (quality === 'suspicious') return '#ef4444'
    if (quality === 'missing') return '#6b7280'
    return '#6b7280'
  }

  if (colorMode.value === 'bpl_status') return getBplColor(house)
  if (colorMode.value === 'divyang_presence') return getDivyangColor(house)
  if (colorMode.value === 'employment') return getEmploymentColor(house)
  return '#2c7a7b'
}

function buildTooltipContent(house) {
  const houseNo = escapeHtml(house.house_no || 'N/A')
  const headName = escapeHtml(house.head_name || '')
  const members = Number(house.total_members || 0)

  if (colorMode.value === 'employment') {
    const workingMembers = getWorkingMembers(house)
    return `House No: ${houseNo}<br/>Head: ${headName}<br/>Members: ${members}<br/>Working Members: ${workingMembers}`
  }

  if (colorMode.value === 'bpl_status') {
    const category = escapeHtml(getBplStatusLabel(house))
    return `House No: ${houseNo}<br/>Head: ${headName}<br/>Members: ${members}<br/>Category: ${category}`
  }
  return `House No: ${houseNo}<br/>Head: ${headName}<br/>Members: ${members}`
}

function buildQueryParams() {
  const params = {}
  if (selectedDistrict.value) params.district_id = selectedDistrict.value
  if (selectedTaluka.value) params.taluka_id = selectedTaluka.value
  if (selectedVillage.value) params.village_id = selectedVillage.value
  if (colorMode.value === 'bpl_status') params.color_by = 'bpl'
  if (colorMode.value === 'divyang_presence') params.color_by = 'divyang'
  if (colorMode.value === 'employment') params.color_by = 'employment'
  return params
}

async function loadLocationOptions() {
  try {
    const res = await getLocationOptions({
      district_id: selectedDistrict.value || undefined,
      taluka_id: selectedTaluka.value || undefined,
    })
    districtOptions.value = res.districts || []
    talukaOptions.value = res.talukas || []
    villageOptions.value = res.villages || []
  } catch (error) {
    console.warn('Population location options unavailable:', error.message)
  }
}

function pct(value, total) {
  if (!total) return 0
  return Math.round((value / total) * 100)
}

function issueClass(value, total = null, higherIsBetter = false) {
  const share = total === null ? Number(value || 0) : pct(value, total)
  if (higherIsBetter) {
    if (share >= 60) return 'vstat-ok'
    if (share >= 30) return 'vstat-warn'
    return 'vstat-bad'
  }
  if (share >= 60) return 'vstat-bad'
  if (share >= 30) return 'vstat-warn'
  return 'vstat-ok'
}

function buildVillageClusters(rows) {
  const buckets = new Map()

  rows.forEach((house) => {
    if (typeof house.lat !== 'number' || typeof house.lng !== 'number') return
    const villageId = String(house.village_id || '').trim()
    const latKey = Math.round(house.lat * 20) / 20
    const lngKey = Math.round(house.lng * 20) / 20
    const key = villageId ? `village:${villageId}` : `${latKey}:${lngKey}`
    const existing = buckets.get(key) || {
      cluster_id: villageId || `cluster-${buckets.size + 1}`,
      village_id: villageId,
      latitude: latKey,
      longitude: lngKey,
      count: 0,
      households: 0,
      total_population: 0,
      bpl_families: 0,
      literate_members: 0,
      working_members: 0,
      divyang_members: 0,
      name: house.village_name || (villageId ? `Village ${villageId}` : `Village Cluster ${buckets.size + 1}`),
      level: villageId ? 'Village ID' : 'Live Cluster',
    }

    existing.count += 1
    existing.households += 1
    existing.total_population += Number(house.total_members || 0)
    existing.literate_members += Number(house.literate_members || 0)
    existing.working_members += Number(house.village_working_members || 0)
    existing.divyang_members += Number(house.divyang_members || 0)

    if (normalizeText(house.FAMILY_BELONG_BPL_CATEGORY) === 'yes') {
      existing.bpl_families += 1
    }

    buckets.set(key, existing)
  })

  return [...buckets.values()]
    .map((cluster) => ({
      ...cluster,
      bpl_percent: pct(cluster.bpl_families, cluster.households),
      literacy_percent: pct(cluster.literate_members, cluster.total_population),
      working_percent: pct(cluster.working_members, cluster.total_population),
      divyang_percent: pct(cluster.divyang_members, cluster.total_population),
    }))
    .sort((a, b) => b.count - a.count)
}

const clusterIssues = computed(() => {
  const cluster = selectedCluster.value
  if (!cluster) return []
  return [
    { label: 'BPL Families', pct: cluster.bpl_percent, color: '#60a5fa' },
    { label: 'Literacy', pct: cluster.literacy_percent, color: '#10b981' },
    { label: 'Working Population', pct: cluster.working_percent, color: '#f59e0b' },
    { label: 'Divyang Population', pct: cluster.divyang_percent, color: '#a78bfa' },
  ]
})

function clearClusterSelection() {
  selectedCluster.value = null
  if (clusterLayer && map) {
    clusterLayer.clearLayers()
  }
}

function clearMarkerLayer() {
  if (markerLayer && map) {
    markerLayer.clearLayers()
  }
}

function renderMarkers(options = {}) {
  const { fitBounds = true } = options
  if (!map || !markerLayer) return

  clearMarkerLayer()
  selectedMarker.value = null

  markers.value.forEach((marker) => {
    const markerColor = getMarkerColor(marker)

    const circle = L.circleMarker([marker.lat, marker.lng], {
      radius: 8,
      fillColor: markerColor,
      color: '#ffffff',
      weight: 2,
      opacity: 1,
      fillOpacity: 0.9,
    })

    circle.bindTooltip(buildTooltipContent(marker), {
      sticky: true,
      direction: 'top',
      opacity: 0.96,
      className: 'map-tooltip',
    })

    circle.on('click', () => {
      selectedMarker.value = marker
      selectedCluster.value = null
    })

    circle.addTo(markerLayer)
  })

  if (fitBounds) {
    fitMapToMarkers()
  }
}

function fitMapToMarkers() {
  if (!map || !markers.value.length) return
  const bounds = L.latLngBounds(markers.value.map((marker) => [marker.lat, marker.lng]))
  if (bounds.isValid()) {
    map.fitBounds(bounds.pad(0.18))
  }
}

function renderClusters() {
  if (!map || !clusterLayer) return

  clusterLayer.clearLayers()
  selectedMarker.value = null

  const clusters = buildVillageClusters(markers.value)
  if (!clusters.length) return

  const maxCount = Math.max(...clusters.map((cluster) => cluster.count))
  clusters.forEach((cluster) => {
    const color = cluster.count >= 100 ? '#10b981' : cluster.count >= 30 ? '#f59e0b' : '#ef4444'
    const radius = Math.max(4000, Math.min(18000, 4000 + (Math.log(cluster.count + 1) / Math.log(maxCount + 1)) * 14000))
    const baseStyle = {
      fillOpacity: 0.18,
      weight: 1.5,
      opacity: 0.55,
    }

    const circle = L.circle([cluster.latitude, cluster.longitude], {
      radius,
      fillColor: color,
      fillOpacity: baseStyle.fillOpacity,
      color,
      weight: baseStyle.weight,
      opacity: baseStyle.opacity,
    }).addTo(clusterLayer)

    const dot = L.circleMarker([cluster.latitude, cluster.longitude], {
      radius: 7,
      fillColor: color,
      color: '#fff',
      weight: 2,
      fillOpacity: 1,
    }).addTo(clusterLayer)

    L.marker([cluster.latitude, cluster.longitude], {
      icon: L.divIcon({
        className: '',
        html: `<div class="cluster-label">${cluster.name}<br/><span>${cluster.count} HH</span></div>`,
        iconSize: [120, 36],
        iconAnchor: [60, -10],
      }),
      interactive: false,
    }).addTo(clusterLayer)

    const selectCluster = () => {
      selectedCluster.value = cluster
      selectedMarker.value = null
      map.flyTo([cluster.latitude, cluster.longitude], 11, { duration: 1 })
    }

    circle.on('click', selectCluster)
    dot.on('click', selectCluster)
    circle.on('mouseover', () => {
      circle.setStyle({ fillOpacity: 0.30, weight: 2.5 })
      circle.bringToFront()
    })
    circle.on('mouseout', () => {
      circle.setStyle({ fillOpacity: baseStyle.fillOpacity, weight: baseStyle.weight, opacity: baseStyle.opacity })
    })
    circle.bindTooltip(`<strong>${cluster.name}</strong><br/>${cluster.count} households`, { sticky: true })
  })

  const bounds = L.latLngBounds(clusters.map((cluster) => [cluster.latitude, cluster.longitude]))
  if (bounds.isValid()) {
    map.fitBounds(bounds, { padding: [60, 60] })
  }
}

async function fetchMapData() {
  const token = ++requestToken
  loading.value = true

  try {
    const params = buildQueryParams()
    const [markerResponse, insightResponse, summaryResponse] = await Promise.all([
      getPopulationMapData(params),
      getPopulationMapInsights(params),
      getPopulationMapSummary(params),
    ])

    if (token !== requestToken) return

    markers.value = Array.isArray(markerResponse) ? markerResponse : []
    insights.value = insightResponse || insights.value
    summary.value = summaryResponse || { total_households: 0 }
    await nextTick()

    if (viewMode.value === 'villages') {
      clearMarkerLayer()
      renderClusters()
    } else {
      if (clusterLayer) clusterLayer.clearLayers()
      renderMarkers()
    }
  } catch (error) {
    if (error?.name !== 'AbortError') {
      console.error('Population map load failed:', error)
    }
    markers.value = []
  } finally {
    if (token === requestToken) {
      loading.value = false
    }
  }
}

async function applyFilters() {
  await fetchMapData()
}

async function resetFilters() {
  selectedDistrict.value = ''
  selectedTaluka.value = ''
  selectedVillage.value = ''
  await loadLocationOptions()
  await fetchMapData()
}

function markerStatsForCard() {
  const total = markers.value.length || 1
  const bpl = insights.value.bpl_distribution || { bpl: 0, non_bpl: 0, total_households: 0 }
  const education = insights.value.education_status || { literate: 0, illiterate: 0, students: 0, dropouts: 0 }
  const working = insights.value.working_vs_dependent || { working: 0, dependent: 0, total_population: 0 }

  return {
    bpl,
    education,
    working,
    total,
  }
}

const analyticsCards = computed(() => {
  const stats = markerStatsForCard()
  return [
    {
      title: 'BPL Distribution',
      subtitle: 'Household economic category',
      totalLabel: `${stats.bpl.total_households.toLocaleString()} households`,
      centerLabel: 'Households',
      centerValue: stats.bpl.total_households.toLocaleString(),
      segments: [
        { label: 'BPL households', value: stats.bpl.bpl, color: '#ef4444' },
        { label: 'Non-BPL households', value: stats.bpl.non_bpl, color: '#0f766e' },
      ],
    },
    {
      title: 'Education Status',
      subtitle: 'Literacy and school participation',
      totalLabel: `${(stats.education.literate + stats.education.illiterate).toLocaleString()} people`,
      centerLabel: 'People',
      centerValue: (stats.education.literate + stats.education.illiterate).toLocaleString(),
      segments: [
        { label: 'Literate', value: stats.education.literate, color: '#0f766e' },
        { label: 'Illiterate', value: stats.education.illiterate, color: '#f59e0b' },
        { label: 'Students', value: stats.education.students, color: '#2563eb' },
        { label: 'Dropouts', value: stats.education.dropouts, color: '#ef4444' },
      ],
    },
    {
      title: 'Working vs Dependent',
      subtitle: 'Population activity profile',
      totalLabel: `${stats.working.total_population.toLocaleString()} people`,
      centerLabel: 'People',
      centerValue: stats.working.total_population.toLocaleString(),
      segments: [
        { label: 'Working population', value: stats.working.working, color: '#16a34a' },
        { label: 'Dependent population', value: stats.working.dependent, color: '#f59e0b' },
      ],
    },
  ]
})

const detailStats = computed(() => {
  const marker = selectedMarker.value
  if (!marker) return []

  if (colorMode.value === 'employment') {
    const totalMembers = Number(marker.total_members || 0)
    const workingMembers = getWorkingMembers(marker)
    const nonWorkingMembers = Math.max(totalMembers - workingMembers, 0)
    const occupations = getOccupationList(marker)

    return [
      { label: 'Working Members', value: workingMembers.toLocaleString(), style: { color: getEmploymentColor(marker) } },
      { label: 'Non-working Members', value: nonWorkingMembers.toLocaleString() },
      { label: 'Occupations', value: occupations.length ? occupations.join(', ') : 'N/A' },
    ]
  }

  if (colorMode.value === 'divyang_presence') {
    return [
      { label: 'Male', value: Number(marker.male_members || 0).toLocaleString() },
      { label: 'Female', value: Number(marker.female_members || 0).toLocaleString() },
      { label: 'Disability', value: getDivyangStatusLabel(marker), style: { color: getDivyangColor(marker) } },
    ]
  }

  return [
    { label: 'Category', value: getBplStatusLabel(marker), style: { color: getBplColor(marker) } },
    { label: 'Members', value: Number(marker.total_members || 0).toLocaleString() },
    { label: 'Male', value: Number(marker.male_members || 0).toLocaleString() },
    { label: 'Female', value: Number(marker.female_members || 0).toLocaleString() },
  ]
})

const headerLegend = computed(() => {
  if (showLocationIssues.value) {
    return [
      { color: '#22c55e', label: 'Valid location' },
      { color: '#f59e0b', label: 'Approximate GPS' },
      { color: '#ef4444', label: 'Suspicious location' },
      { color: '#6b7280', label: 'Missing GPS' },
    ]
  }

  if (colorMode.value === 'population_density') {
    return [
      { color: '#2c7a7b', label: 'Population households' },
    ]
  }

  if (colorMode.value === 'employment') {
    return [
      { color: '#2e7d32', label: 'Working household' },
      { color: '#f57c00', label: 'No earning member' },
    ]
  }

  if (colorMode.value === 'divyang_presence') {
    return [
      { color: '#7b1fa2', label: 'Divyang present' },
      { color: '#2e7d32', label: 'No disability' },
    ]
  }

  if (colorMode.value === 'bpl_status') {
    return [
      { color: '#e53935', label: 'BPL households' },
      { color: '#2e7d32', label: 'Non-BPL households' },
    ]
  }

  return []
})

function setViewMode(mode) {
  viewMode.value = mode
  if (!map) return

  if (mode === 'villages') {
    clearMarkerLayer()
    if (!clusterLayer) clusterLayer = L.layerGroup().addTo(map)
    renderClusters()
  } else {
    if (clusterLayer) clusterLayer.clearLayers()
    clearClusterSelection()
    renderMarkers()
  }
}

// Distinct palette — 36 colours covering all Maharashtra districts
const DISTRICT_PALETTE = [
  '#3b82f6','#10b981','#f59e0b','#8b5cf6','#06b6d4',
  '#f43f5e','#84cc16','#fb923c','#14b8a6','#eab308',
  '#6366f1','#22d3ee','#4ade80','#fb7185','#facc15',
  '#818cf8','#34d399','#f97316','#c084fc','#2dd4bf',
  '#fbbf24','#f87171','#60a5fa','#a78bfa','#e879f9',
  '#38bdf8','#86efac','#fde68a','#fca5a5','#93c5fd',
  '#d8b4fe','#6ee7b7','#fcd34d','#fdba74','#bef264',
  '#67e8f9',
]

async function addDistrictBorders(mapInstance) {
  try {
    const res = await fetch('https://raw.githubusercontent.com/geohacker/india/master/district/india_district.geojson')
    const data = await res.json()

    // Filter to Maharashtra only — geohacker uses ST_NM property
    const mhDistricts = data.features.filter((f) => {
      const props = f.properties || {}
      return (
        String(props.ST_NM || '').toUpperCase().includes('MAHARASHTRA') ||
        String(props.state || '').toUpperCase().includes('MAHARASHTRA') ||
        String(props.STATE || '').toUpperCase().includes('MAHARASHTRA') ||
        String(props.NAME_1 || '').toUpperCase().includes('MAHARASHTRA')
      )
    })

    mhDistricts.forEach((feature, i) => {
      const color = DISTRICT_PALETTE[i % DISTRICT_PALETTE.length]
      const props = feature.properties || {}
      const districtName = props.DISTRICT || props.dtname || props.NAME_2 || props.district || 'District'
      const baseStyle = {
        color,
        weight: 1.8,
        opacity: 0.75,
        fillColor: color,
        fillOpacity: 0.10,
        dashArray: null,
      }

      const layer = L.geoJSON(feature, {
        renderer: L.svg(),
        style: baseStyle,
        onEachFeature: (_f, districtLayer) => {
          districtLayer.on('mouseover', (e) => {
            const target = e.target
            target.setStyle({ fillOpacity: 0.30, weight: 2.5 })
            target.bindTooltip(
              `<div style="font-weight:700;font-size:0.78rem;color:#1e293b;">${districtName}</div>`,
              { sticky: true, className: 'map-tooltip district-tooltip', direction: 'top' },
            ).openTooltip(e.latlng)
          })

          districtLayer.on('mousemove', (e) => {
            e.target.getTooltip()?.setLatLng(e.latlng)
          })

          districtLayer.on('mouseout', (e) => {
            e.target.setStyle(baseStyle)
            e.target.closeTooltip()
          })
        },
      }).addTo(mapInstance).bringToBack()

      layer.bringToBack()
    })
  } catch (e) {
    console.warn('District boundaries unavailable:', e.message)
  }
}

async function addMaharashtraHighlight(mapInstance) {
  try {
    const res = await fetch('https://raw.githubusercontent.com/geohacker/india/master/state/india_state.geojson')
    const data = await res.json()
    const mh = data.features.find((f) =>
      Object.values(f.properties || {}).some((v) => String(v).toUpperCase().includes('MAHARASHTRA')),
    )
    if (mh) {
      // Outer world ring + Maharashtra hole to darken non-Maharashtra regions.
      const worldRing = [[-90, -180], [90, -180], [90, 180], [-90, 180], [-90, -180]]

      const holeRings = []
      if (mh.geometry.type === 'Polygon') {
        mh.geometry.coordinates.forEach((ring) => holeRings.push(ring))
      } else if (mh.geometry.type === 'MultiPolygon') {
        mh.geometry.coordinates.forEach((poly) =>
          poly.forEach((ring) => holeRings.push(ring)),
        )
      }

      const maskFeature = {
        type: 'Feature',
        geometry: {
          type: 'Polygon',
          coordinates: [worldRing, ...holeRings],
        },
        properties: {},
      }

      L.geoJSON(maskFeature, {
        style: {
          fillColor: '#000000',
          fillOpacity: 0.55,
          color: 'transparent',
          weight: 0,
        },
        interactive: false,
      }).addTo(mapInstance)

      L.geoJSON(mh, {
        style: {
          color: '#f59e0b',
          weight: 2.5,
          opacity: 0.9,
          fillColor: 'transparent',
          fillOpacity: 0,
          dashArray: '8,5',
        },
        interactive: false,
      }).addTo(mapInstance)

      const center = L.geoJSON(mh).getBounds().getCenter()
      L.marker(center, {
        icon: L.divIcon({ className: '', html: '<div class="mh-label">Maharashtra</div>', iconSize: [120, 24], iconAnchor: [60, 12] }),
        interactive: false,
      }).addTo(mapInstance)
    }
  } catch (e) {
    console.warn('Maharashtra boundary unavailable:', e.message)
  }
}

watch(selectedDistrict, async () => {
  selectedTaluka.value = ''
  selectedVillage.value = ''
  await loadLocationOptions()
})

watch(selectedTaluka, async () => {
  selectedVillage.value = ''
  await loadLocationOptions()
})

watch(colorMode, () => {
  if (!map) return
  if (viewMode.value === 'points' && !showLocationIssues.value) {
    renderMarkers({ fitBounds: false })
  }
})

watch(showLocationIssues, () => {
  if (!map) return
  if (viewMode.value === 'points') {
    renderMarkers({ fitBounds: false })
  }
})

watch(analyticsPanelOpen, async () => {
  await nextTick()
  handleMapResize()
})

onMounted(async () => {
  await loadLocationOptions()

  map = L.map(mapContainer.value, {
    zoomControl: true,
    preferCanvas: true,
  }).setView([19.7515, 75.7139], 6)

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map)

  addDistrictBorders(map)
  await addMaharashtraHighlight(map)

  markerLayer = L.layerGroup().addTo(map)
  clusterLayer = L.layerGroup().addTo(map)

  await fetchMapData()
  window.addEventListener('resize', handleMapResize)
  window.addEventListener('click', closeDropdowns)
  document.addEventListener('fullscreenchange', handleFullscreenChange)
})

onUnmounted(() => {
  requestToken += 1
  window.removeEventListener('resize', handleMapResize)
  window.removeEventListener('click', closeDropdowns)
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
  if (map) {
    map.remove()
    map = null
  }
})
</script>

<style scoped>
.map-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  position: relative;
}

.map-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.85rem 2rem;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border);
  z-index: 20;
  flex-shrink: 0;
  gap: 1.5rem;
}

.page-title {
  font-family: var(--font-display);
  font-size: 1.5rem;
  color: var(--text-primary);
  font-weight: 400;
}

.page-subtitle {
  color: var(--text-dim);
  font-size: 0.75rem;
  margin-top: 0.2rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.map-controls {
  display: flex;
  align-items: center;
  gap: 0.9rem;
  flex-wrap: wrap;
}

.view-toggle {
  display: flex;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

.toggle-btn {
  padding: 0.3rem 0.85rem;
  font-size: 0.72rem;
  font-family: var(--font-body);
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  white-space: nowrap;
}

.toggle-btn:hover { background: var(--bg-card); color: var(--text-body); }

.toggle-btn.active {
  background: var(--teal);
  color: #fff;
}

.map-control-group {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.custom-select {
  position: relative;
  min-width: 90px;
}

.cs-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.35rem;
  width: 100%;
  background: #ffffff;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  color: #334155;
  font-family: var(--font-body);
  font-size: 0.76rem;
  padding: 0.3rem 0.6rem;
  cursor: pointer;
  outline: none;
  text-align: left;
  white-space: nowrap;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.cs-trigger:hover:not(:disabled) {
  border-color: #94a3b8;
  background: #f9fafb;
}

.custom-select.open .cs-trigger {
  border-color: #14b8a6;
  box-shadow: 0 0 0 2px rgba(20, 184, 166, 0.18);
}

.custom-select.disabled .cs-trigger,
.cs-trigger:disabled {
  background: #f3f4f6;
  color: #9ca3af;
  border-color: #e5e7eb;
  cursor: not-allowed;
}

.cs-value {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
}

.cs-arrow {
  font-size: 0.58rem;
  color: #64748b;
  flex-shrink: 0;
  transition: transform 0.15s;
  line-height: 1;
}

.custom-select.open .cs-arrow {
  transform: rotate(180deg);
}

.cs-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 100%;
  max-height: 220px;
  overflow-y: auto;
  background: #ffffff;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.18), 0 3px 8px rgba(0,0,0,0.10);
  z-index: 9999;
}

.cs-dropdown-right {
  left: auto;
  right: 0;
}

.cs-option {
  padding: 0.42rem 0.75rem;
  font-size: 0.76rem;
  color: #1e293b;
  background: #ffffff;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.1s, color 0.1s;
}

.cs-option:hover {
  background: #f0fdfa;
  color: #0f766e;
}

.cs-option.selected {
  background: #ccfbf1;
  color: #0f766e;
  font-weight: 600;
}

.gps-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.38rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  user-select: none;
}

.gps-toggle input {
  margin: 0;
  accent-color: #14b8a6;
}

.control-label {
  font-size: 0.62rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-dim);
  white-space: nowrap;
}

.control-select {
  background: #ffffff;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  color: #334155;
  color-scheme: light;
  appearance: none;
  font-family: var(--font-body);
  font-size: 0.76rem;
  padding: 0.3rem 0.6rem; outline: none; cursor: pointer;
}

.control-select option {
  color: #334155;
  background: #ffffff;
}

.village-select,
.village-select option {
  color: #334155;
}

.control-select:focus {
  border-color: #14b8a6;
  box-shadow: 0 0 0 2px rgba(20, 184, 166, 0.18);
}

.control-select:disabled {
  background: #f3f4f6;
  color: #9ca3af;
  cursor: not-allowed;
}

.apply-btn {
  border: 1px solid #14b8a6;
  background: #14b8a6;
  color: #ffffff;
  border-radius: 6px;
  font-size: 0.74rem;
  padding: 0.34rem 0.8rem;
  cursor: pointer;
  font-family: var(--font-body);
  transition: background 0.15s, border-color 0.15s;
}

.apply-btn:hover {
  background: #0d9488;
  border-color: #0d9488;
}

.reset-btn {
  border: 1px solid #cbd5e1;
  background: #ffffff;
  color: #475569;
  border-radius: 6px;
  font-size: 0.74rem;
  padding: 0.34rem 0.8rem;
  cursor: pointer;
  font-family: var(--font-body);
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}

.reset-btn:hover {
  background: #f8fafc;
  border-color: #94a3b8;
  color: #334155;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.67rem;
  color: var(--text-muted);
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 999px;
  padding: 0.18rem 0.55rem 0.18rem 0.3rem;
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  box-shadow: inset 0 0 0 1px rgba(0,0,0,0.08);
}

.map-legend {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
  flex-basis: 100%;
  margin-top: 0.2rem;
}

.analytics-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 1rem;
  box-shadow: 0 10px 24px var(--shadow);
}

.analytics-card-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 0.75rem;
}

.analytics-title {
  font-family: var(--font-display);
  font-size: 1rem;
  color: var(--text-primary);
  font-weight: 400;
}

.analytics-subtitle {
  font-size: 0.72rem;
  color: var(--text-dim);
  margin-top: 0.1rem;
}

.analytics-total {
  font-size: 0.65rem;
  color: var(--amber);
  border: 1px solid var(--amber-dim);
  background: var(--amber-dim);
  border-radius: 999px;
  padding: 0.2rem 0.45rem;
  white-space: nowrap;
}

.chart-layout {
  display: grid;
  grid-template-columns: 110px 1fr;
  gap: 1rem;
  align-items: center;
}

.donut {
  width: 110px;
  height: 110px;
  border-radius: 50%;
  position: relative;
  border: 1px solid var(--border);
}

.donut-hole {
  position: absolute;
  inset: 23px;
  border-radius: 50%;
  background: var(--bg-primary);
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.donut-label {
  font-size: 0.6rem;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.donut-value {
  font-family: var(--font-display);
  font-size: 1.1rem;
  color: var(--text-primary);
}

.legend-list { display: flex; flex-direction: column; gap: 0.45rem; }

.legend-row {
  display: grid;
  grid-template-columns: 8px 1fr auto;
  gap: 0.45rem;
  align-items: center;
  font-size: 0.74rem;
}

.legend-name { color: var(--text-muted); }
.legend-value { color: var(--text-body); font-variant-numeric: tabular-nums; }

.family-list-wrap {
  margin-top: 0.65rem;
  padding-top: 0.55rem;
  border-top: 1px solid var(--border);
}

.family-list-title {
  font-size: 0.7rem;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-bottom: 0.35rem;
}

.family-list {
  margin: 0;
  padding-left: 1rem;
  display: grid;
  gap: 0.2rem;
}

.family-list li {
  font-size: 0.78rem;
  color: var(--text-body);
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}

.family-head {
  color: var(--text-primary);
  font-weight: 600;
}

.family-members {
  color: var(--text-muted);
}

.map-shell {
  padding: 1rem 2rem 1.5rem;
  flex: 1;
  min-height: 0;
}

.map-content {
  position: relative;
  display: flex;
  height: 100%;
  min-height: 520px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0,0,0,0.07), 0 1px 4px rgba(0,0,0,0.04);
}

.map-stage {
  position: relative;
  flex: 1;
  min-width: 0;
}

.map-floating-controls {
  position: absolute;
  top: 20px;
  left: 20px;
  z-index: 450;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.analytics-toggle {
  border: 1px solid rgba(20,184,166,0.4);
  background: rgba(255,255,255,0.92);
  color: var(--teal);
  border-radius: 999px;
  font-size: 0.68rem;
  font-weight: 600;
  padding: 0.36rem 0.8rem;
  cursor: pointer;
  font-family: var(--font-body);
  transition: all 0.14s ease;
  box-shadow: 0 2px 10px rgba(0,0,0,0.08);
}

.analytics-toggle:hover {
  background: var(--teal);
  color: #ffffff;
  border-color: var(--teal);
}

.fullscreen-toggle {
  border: 1px solid rgba(0,0,0,0.1);
  background: rgba(255,255,255,0.92);
  color: #475569;
  border-radius: 999px;
  font-size: 0.68rem;
  font-weight: 500;
  padding: 0.36rem 0.8rem;
  cursor: pointer;
  font-family: var(--font-body);
  transition: all 0.14s ease;
  box-shadow: 0 2px 10px rgba(0,0,0,0.08);
}

.fullscreen-toggle:hover {
  border-color: rgba(20,184,166,0.4);
  color: var(--teal);
}

.analytics-panel {
  width: 360px;
  max-width: 42vw;
  border-left: 1px solid var(--border);
  background: color-mix(in srgb, var(--bg-card) 88%, transparent);
  backdrop-filter: blur(6px);
  display: flex;
  flex-direction: column;
  z-index: 5;
}

.analytics-panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.85rem 1rem;
  border-bottom: 1px solid var(--border);
}

.analytics-panel-title {
  font-family: var(--font-display);
  font-size: 1.05rem;
  font-weight: 400;
  color: var(--text-primary);
}

.analytics-close {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  border: 1px solid var(--border);
  background: var(--bg-surface);
  color: var(--text-muted);
  cursor: pointer;
  font-size: 1rem;
  line-height: 1;
}

.analytics-scroll {
  padding: 0.75rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.analytics-panel-slide-enter-active,
.analytics-panel-slide-leave-active {
  transition: all 0.22s ease;
}

.analytics-panel-slide-enter-from,
.analytics-panel-slide-leave-to {
  opacity: 0;
  transform: translateX(16px);
}

.map-container {
  position: absolute;
  inset: 0;
  z-index: 1;
}

.empty-state {
  margin: 0 2rem 1rem;
  padding: 0.85rem 1rem;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--bg-card);
  color: var(--text-muted);
  font-size: 0.8rem;
}

.detail-panel {
  position: absolute;
  z-index: 10;
  right: 1rem;
  top: 1rem;
  width: min(340px, calc(100% - 2rem));
  padding: 1rem;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(21, 48, 35, 0.08);
  box-shadow: 0 18px 40px rgba(16, 27, 21, 0.14);
}

.panel-close {
  position: absolute;
  top: 0.6rem;
  right: 0.75rem;
  border: none;
  background: transparent;
  color: #6a7c70;
  font-size: 1.6rem;
  line-height: 1;
  cursor: pointer;
}

.panel-title {
  margin: 0;
  padding-right: 2rem;
  color: #153023;
  font-size: 1.1rem;
}

.panel-subtitle {
  margin-top: 0.2rem;
  color: #637567;
  font-size: 0.86rem;
  font-weight: 600;
}

.panel-id {
  margin-top: 0.25rem;
  color: #637567;
  font-size: 0.9rem;
  font-weight: 600;
}

.panel-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.65rem;
  margin-top: 1rem;
}

.panel-stat {
  padding: 0.75rem;
  border-radius: 14px;
  background: rgba(15, 118, 110, 0.06);
}

.panel-stat-label {
  color: #65776b;
  font-size: 0.74rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  font-weight: 700;
}

.panel-stat-value {
  margin-top: 0.35rem;
  color: #153023;
  font-size: 1rem;
  font-weight: 800;
}

.panel-coords {
  margin-top: 0.95rem;
  color: #718176;
  font-size: 0.84rem;
  word-break: break-word;
}

.village-panel {
  width: min(360px, calc(100% - 2rem));
}

.village-badge {
  display: inline-flex;
  align-items: center;
  margin-bottom: 0.5rem;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  border: 1px solid rgba(15, 118, 110, 0.28);
  color: #0f766e;
  font-size: 0.68rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  background: rgba(15, 118, 110, 0.08);
}

.village-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.5rem;
  margin-top: 1rem;
}

.vstat {
  padding: 0.75rem 0.65rem;
  border-radius: 14px;
  text-align: center;
}

.vstat-bad {
  background: rgba(239, 68, 68, 0.10);
  border: 1px solid rgba(239, 68, 68, 0.25);
}

.vstat-warn {
  background: rgba(245, 158, 11, 0.10);
  border: 1px solid rgba(245, 158, 11, 0.25);
}

.vstat-ok {
  background: rgba(22, 163, 74, 0.10);
  border: 1px solid rgba(22, 163, 74, 0.25);
}

.vstat-val {
  color: #153023;
  font-size: 1.3rem;
  font-weight: 800;
  line-height: 1;
}

.vstat-label {
  margin-top: 0.25rem;
  color: #65776b;
  font-size: 0.64rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  font-weight: 700;
}

.village-bar-section {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  margin-top: 1rem;
}

.vbar-row {
  display: grid;
  grid-template-columns: 92px 1fr 34px;
  align-items: center;
  gap: 0.5rem;
}

.vbar-label {
  font-size: 0.68rem;
  color: #617166;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.vbar-track {
  height: 6px;
  background: #e5e7eb;
  border-radius: 999px;
  overflow: hidden;
}

.vbar-fill {
  height: 100%;
  border-radius: 999px;
}

.vbar-pct {
  font-size: 0.65rem;
  color: #718176;
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.gps-detail-card {
  margin-top: 0.95rem;
  padding: 0.75rem;
  border-radius: 14px;
  background: rgba(59, 130, 246, 0.06);
  border: 1px solid rgba(59, 130, 246, 0.18);
}

.gps-detail-title {
  color: #374151;
  font-size: 0.62rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-weight: 800;
  margin-bottom: 0.45rem;
}

.gps-detail-row {
  display: flex;
  gap: 0.55rem;
  align-items: flex-start;
}

.gps-detail-icon {
  font-size: 1.05rem;
  line-height: 1;
  margin-top: 0.08rem;
  flex-shrink: 0;
}

.gps-detail-body {
  min-width: 0;
}

.gps-detail-label {
  font-size: 0.82rem;
  font-weight: 800;
}

.gps-detail-reason,
.gps-detail-coords {
  margin-top: 0.2rem;
  color: #4b5563;
  font-size: 0.72rem;
  line-height: 1.45;
  word-break: break-word;
}

.gps-dup-card {
  margin-top: 0.6rem;
  padding-top: 0.55rem;
  border-top: 1px solid rgba(59, 130, 246, 0.22);
}

.gps-dup-title {
  font-size: 0.7rem;
  color: #1f2937;
  font-weight: 700;
}

.gps-dup-item {
  margin-top: 0.35rem;
  color: #4b5563;
  font-size: 0.71rem;
}

.duplicate-list-scroll {
  max-height: 160px;
  overflow-y: auto;
  padding-right: 4px;
}

.duplicate-list-scroll::-webkit-scrollbar {
  width: 6px;
}

.duplicate-list-scroll::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 4px;
}

.gps-dup-toggle {
  margin-top: 0.35rem;
  border: none;
  background: transparent;
  color: #2563eb;
  font-size: 0.68rem;
  font-weight: 700;
  padding: 0;
  cursor: pointer;
}

.gps-dup-toggle:hover {
  text-decoration: underline;
}

.slide-enter-active { transition: all 0.25s ease-out; }
.slide-leave-active { transition: all 0.15s ease-in; }
.slide-enter-from, .slide-leave-to { opacity: 0; transform: translateX(20px); }

@media (max-width: 1100px) {
  .chart-layout {
    grid-template-columns: 96px 1fr;
  }

  .analytics-panel {
    width: 320px;
    max-width: 48vw;
  }
}

@media (max-width: 760px) {
  .map-header,
  .map-shell {
    padding-left: 1rem;
    padding-right: 1rem;
  }

  .map-header { flex-direction: column; align-items: flex-start; gap: 0.75rem; }

  .chart-layout {
    grid-template-columns: 1fr;
  }

  .donut {
    margin: 0 auto;
  }

  .detail-panel {
    width: calc(100% - 2rem);
  }

  .analytics-panel {
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    width: min(88vw, 340px);
    max-width: none;
    border-left: 1px solid var(--border);
    box-shadow: -8px 0 26px var(--shadow);
  }

}
</style>

<style>
.map-tooltip {
  background: var(--bg-card) !important;
  border: 1px solid var(--border) !important;
  border-radius: 8px !important;
  color: var(--text-body) !important;
  font-family: var(--font-body) !important;
  font-size: 0.75rem !important;
  padding: 0.5rem 0.75rem !important;
  box-shadow: 0 4px 16px var(--shadow) !important;
}

.leaflet-popup-content-wrapper.map-tooltip,
.leaflet-popup-tip.map-tooltip {
  background: var(--bg-card) !important;
  color: var(--text-body) !important;
  border: 1px solid var(--border) !important;
}

.leaflet-popup-content-wrapper.map-tooltip {
  border-radius: 8px !important;
}

.mh-label {
  font-family: sans-serif;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: #f59e0b;
  text-shadow: 0 1px 3px rgba(0,0,0,0.5);
  white-space: nowrap;
  pointer-events: none;
}
</style>
