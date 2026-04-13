<template>
  <div class="map-page">
    <header class="map-header">
      <div class="map-title-area">
        <h1 class="page-title">Geo-Intelligence Map</h1>
        <p class="page-subtitle">
          {{ houses.length.toLocaleString() }} households plotted from the live database
        </p>
      </div>
      <div class="map-controls">
        <!-- View mode toggle -->
        <div class="view-toggle">
          <button class="toggle-btn" :class="{ active: viewMode === 'points' }" @click="setViewMode('points')">
            Households
          </button>
          <button class="toggle-btn" :class="{ active: viewMode === 'villages' }" @click="setViewMode('villages')">
            Villages
          </button>
        </div>

        <!-- District -->
        <div class="map-control-group">
          <label class="control-label">District</label>
          <div class="custom-select" :class="{ open: openDropdown === 'district' }"
               @click.stop="toggleDropdown('district')">
            <button class="cs-trigger" type="button">
              <span class="cs-value">{{ selectedDistrictLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'district'" @click.stop>
              <div class="cs-option" :class="{ selected: !selectedDistrict }" @click="selectDistrict('')">All</div>
              <div class="cs-option" v-for="d in districtOptions" :key="d.id"
                   :class="{ selected: String(selectedDistrict) === String(d.id) }"
                   @click="selectDistrict(d.id)">{{ d.name }}</div>
            </div>
          </div>
        </div>

        <!-- Taluka -->
        <div class="map-control-group">
          <label class="control-label">Taluka</label>
          <div class="custom-select" :class="{ open: openDropdown === 'taluka', disabled: !talukaOptions.length }"
               @click.stop="talukaOptions.length && toggleDropdown('taluka')">
            <button class="cs-trigger" type="button" :disabled="!talukaOptions.length">
              <span class="cs-value">{{ selectedTalukaLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'taluka'" @click.stop>
              <div class="cs-option" :class="{ selected: !selectedTaluka }" @click="selectTaluka('')">All</div>
              <div class="cs-option" v-for="t in talukaOptions" :key="t.id"
                   :class="{ selected: String(selectedTaluka) === String(t.id) }"
                   @click="selectTaluka(t.id)">{{ t.name }}</div>
            </div>
          </div>
        </div>

        <!-- Village -->
        <div class="map-control-group">
          <label class="control-label">Village</label>
          <div class="custom-select" :class="{ open: openDropdown === 'village', disabled: !villageOptions.length }"
               @click.stop="villageOptions.length && toggleDropdown('village')">
            <button class="cs-trigger" type="button" :disabled="!villageOptions.length">
              <span class="cs-value">{{ selectedVillageLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'village'" @click.stop>
              <div class="cs-option" :class="{ selected: !selectedVillage }" @click="selectVillage('')">All</div>
              <div class="cs-option" v-for="v in villageOptions" :key="v.id"
                   :class="{ selected: String(selectedVillage) === String(v.id) }"
                   @click="selectVillage(v.id)">{{ v.name }}</div>
            </div>
          </div>
        </div>

        <div class="map-control-group">
          <button class="apply-btn" @click="applyFilters">Apply</button>
          <button class="reset-btn" @click="resetFilters">Reset</button>
        </div>

        <!-- Color by (only in points mode) -->
        <div class="map-control-group" v-if="viewMode === 'points'">
          <label class="control-label">Color by</label>
          <div class="custom-select cs-align-right" :class="{ open: openDropdown === 'colorMode' }"
               @click.stop="toggleDropdown('colorMode')">
            <button class="cs-trigger" type="button">
              <span class="cs-value">{{ selectedColorModeLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown cs-dropdown-right" v-show="openDropdown === 'colorMode'" @click.stop>
              <div class="cs-option" :class="{ selected: colorMode === 'sanitation' }" @click="selectColorMode('sanitation')">Sanitation</div>
              <div class="cs-option" :class="{ selected: colorMode === 'crops' }"      @click="selectColorMode('crops')">Crops / Season</div>
              <div class="cs-option" :class="{ selected: colorMode === 'land' }"       @click="selectColorMode('land')">Land Holdings</div>
            </div>
          </div>
        </div>

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

    <section class="analytics-grid" v-if="analyticsCards.length">
      <article class="analytics-card" v-for="card in analyticsCards" :key="card.title">
        <div class="analytics-card-head">
          <div>
            <h2 class="analytics-title">{{ card.title }}</h2>
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
            <div class="legend-row" v-for="item in card.segments" :key="item.label">
              <span class="legend-dot" :style="{ background: item.color }"></span>
              <span class="legend-name">{{ item.label }}</span>
              <span class="legend-value">{{ item.value.toLocaleString() }}</span>
            </div>
          </div>
        </div>
      </article>
    </section>

    <section class="map-shell">
      <div v-if="!loading && !houses.length" class="empty-state">
        No live household data returned from the database API.
      </div>

      <div class="map-content">
        <div class="map-container" ref="mapContainer"></div>

        <!-- Household Detail Panel -->
        <transition name="slide">
          <aside v-if="selectedHouse && viewMode === 'points'" class="detail-panel">

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
                  <span v-if="selectedHouse.villageName">{{ selectedHouse.villageName }}</span>
                  <span v-if="selectedHouse.talukaName"> · {{ selectedHouse.talukaName }}</span>
                </div>
              </div>
              <button class="detail-close" @click="selectedHouse = null" title="Close">×</button>
            </div>

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

            <div class="dp-field-row">
              <span class="dp-field-icon">💧</span>
              <span class="dp-field-key">Irrigation Source</span>
              <span class="dp-field-val"
                    :style="{ color: (selectedHouse.waterSource || '').toLowerCase().includes('rain') ? '#b45309' : '#15803d' }">
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
                    :style="{ color: (selectedHouse.lighting || '').toLowerCase() === 'electricity' ? '#15803d' : '#b45309' }">
                {{ selectedHouse.lighting || '—' }}
              </span>
            </div>

            <div class="dp-field-row">
              <span class="dp-field-icon">🪪</span>
              <span class="dp-field-key">Ration Card</span>
              <span class="dp-field-val">{{ selectedHouse.rationCard || '—' }}</span>
            </div>

            <div class="panel-coords">
              {{ selectedHouse.latitude.toFixed(6) }}, {{ selectedHouse.longitude.toFixed(6) }}
            </div>

          </aside>
        </transition>

        <!-- Village/GP Detail Panel -->
        <transition name="slide">
          <aside v-if="selectedCluster" class="detail-panel village-panel">
            <button class="panel-close" @click="clearClusterSelection">×</button>
            <div class="village-badge">{{ selectedCluster.level }}</div>
            <h3 class="panel-title">{{ selectedCluster.name }}</h3>
            <div class="panel-id">{{ selectedCluster.count.toLocaleString() }} households covered</div>

            <div class="village-stats">
              <div class="vstat" :class="issueClass(selectedCluster.noToilet, selectedCluster.count)">
                <div class="vstat-val">{{ pct(selectedCluster.noToilet, selectedCluster.count) }}%</div>
                <div class="vstat-label">No Sanitation</div>
              </div>
              <div class="vstat" :class="issueClass(selectedCluster.noElec, selectedCluster.count)">
                <div class="vstat-val">{{ pct(selectedCluster.noElec, selectedCluster.count) }}%</div>
                <div class="vstat-label">No Electricity</div>
              </div>
              <div class="vstat" :class="issueClass(selectedCluster.noIrrig, selectedCluster.count)">
                <div class="vstat-val">{{ pct(selectedCluster.noIrrig, selectedCluster.count) }}%</div>
                <div class="vstat-label">No Irrigation</div>
              </div>
              <div class="vstat" :class="issueClass(selectedCluster.bpl, selectedCluster.count)">
                <div class="vstat-val">{{ pct(selectedCluster.bpl, selectedCluster.count) }}%</div>
                <div class="vstat-label">BPL Families</div>
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
    </section>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { getHouses, getLocationOptions } from '../../api/index.js'
import L from 'leaflet'

const loading       = ref(true)
const houses        = ref([])
const selectedHouse  = ref(null)
const selectedCluster = ref(null)
const mapContainer  = ref(null)
const colorMode     = ref('sanitation')
const viewMode      = ref('points')   // 'points' | 'villages'
const districtOptions = ref([])
const talukaOptions = ref([])
const villageOptions = ref([])
const selectedDistrict = ref('')
const selectedTaluka = ref('')
const selectedVillage = ref('')

// ── Custom dropdown state ─────────────────────────────────────────────────────
const openDropdown = ref(null)

function toggleDropdown(name) {
  openDropdown.value = openDropdown.value === name ? null : name
}

function closeDropdowns() {
  openDropdown.value = null
}

// Selection handlers — the existing watchers handle cascade reset + API refetch
function selectDistrict(id) {
  selectedDistrict.value = id   // watcher fires: resets taluka/village + reloads options
  closeDropdowns()
}

function selectTaluka(id) {
  selectedTaluka.value = id     // watcher fires: resets village + reloads options
  closeDropdowns()
}

function selectVillage(id) {
  selectedVillage.value = id
  closeDropdowns()
}

const COLOR_MODE_LABELS_MAP = {
  sanitation: 'Sanitation',
  crops:      'Crops / Season',
  land:       'Land Holdings',
}

function selectColorMode(mode) {
  colorMode.value = mode
  closeDropdowns()
}

// Human-readable labels shown in the trigger button
const selectedDistrictLabel = computed(() => {
  if (!selectedDistrict.value) return 'All'
  return districtOptions.value.find(d => String(d.id) === String(selectedDistrict.value))?.name || 'All'
})
const selectedTalukaLabel = computed(() => {
  if (!selectedTaluka.value) return 'All'
  return talukaOptions.value.find(t => String(t.id) === String(selectedTaluka.value))?.name || 'All'
})
const selectedVillageLabel = computed(() => {
  if (!selectedVillage.value) return 'All'
  return villageOptions.value.find(v => String(v.id) === String(selectedVillage.value))?.name || 'All'
})
const selectedColorModeLabel = computed(() => COLOR_MODE_LABELS_MAP[colorMode.value] || 'Sanitation')

let map = null
const markerRefs    = []   // { marker, house }
let clusterGroup    = null // L.layerGroup for village circles
let highlightCircle = null // currently highlighted village circle
let retryTimer = null

function handleMapResize() {
  if (!map) return
  map.invalidateSize()
}

function clearMarkers() {
  markerRefs.forEach(({ marker }) => {
    if (map && map.hasLayer(marker)) map.removeLayer(marker)
  })
  markerRefs.length = 0
}

async function loadLocationDropdowns() {
  try {
    const res = await getLocationOptions({
      district_id: selectedDistrict.value,
      taluka_id: selectedTaluka.value,
    })
    districtOptions.value = res.districts || []
    talukaOptions.value = res.talukas || []
    villageOptions.value = res.villages || []
  } catch (e) {
    console.warn('Location options unavailable:', e.message)
  }
}

function getHouseFilters() {
  return {
    limit: 2000,
    district_id: selectedDistrict.value || undefined,
    taluka_id: selectedTaluka.value || undefined,
    village_id: selectedVillage.value || undefined,
  }
}

async function fetchAllHouses() {
  const base = getHouseFilters()
  const pageLimit = Number(base.limit) || 2000
  let page = 1
  const all = []
  let total = null

  while (true) {
    const res = await getHouses({ ...base, page, limit: pageLimit })
    const chunk = res.data || []

    if (!chunk.length) break
    all.push(...chunk)

    if (typeof res.total === 'number') total = res.total

    if (chunk.length < pageLimit) break
    if (total !== null && all.length >= total) break
    if (page >= 20) break // hard guard against runaway paging

    page += 1
  }

  return all
}

function applyFilters() {
  houses.value = []
  selectedHouse.value = null
  selectedCluster.value = null
  clearClusterSelection()
  if (clusterGroup) { clusterGroup.remove(); clusterGroup = null }
  clearMarkers()
  if (map) {
    loading.value = true
    loadLiveHouseData(0)
  }
}

async function resetFilters() {
  selectedDistrict.value = ''
  selectedTaluka.value = ''
  selectedVillage.value = ''
  await loadLocationDropdowns()
  applyFilters()
}

const detailStats = computed(() => {
  const h = selectedHouse.value
  if (!h) return []
  return [
    { label: 'Total Land',  value: `${h.totalLand || '0'} acres` },
    { label: 'Cultivated',  value: `${h.cultivatedLand || '0'} acres` },
    { label: 'Irrigation',  value: h.waterSource || 'None' },
    { label: 'Latrine',     value: h.latrine || 'None', style: { color: getConditionColor(h) } },
    { label: 'Lighting',    value: h.lighting || 'None' },
    { label: 'Ration Card', value: h.rationCard || 'Unknown' },
    { label: 'Kharif Crop', value: h.kharif || 'No' },
    { label: 'Rabi Crop',   value: h.rabi || 'No' },
  ]
})

const stats = computed(() => {
  if (!houses.value.length) return null
  const total = houses.value.length
  const farmers = houses.value.filter(h => (h.ownLand || '').toLowerCase() === 'yes').length
  const noIrrigation = houses.value.filter(h => !h.waterSource || h.waterSource === 'Rain Fed' || h.waterSource === 'None').length
  const kharif = houses.value.filter(h => (h.kharif || '').toLowerCase() === 'yes').length
  const rabi = houses.value.filter(h => (h.rabi || '').toLowerCase() === 'yes').length

  return { total, farmers, noIrrigation, kharif, rabi }
})

const analyticsCards = computed(() => {
  if (!stats.value) return []
  const total = stats.value.total || 1
  const landless = houses.value.filter(h => (parseFloat(h.totalLand) || 0) <= 1).length
  const small = houses.value.filter(h => {
    const land = parseFloat(h.totalLand) || 0
    return land > 1 && land <= 2.5
  }).length
  const mediumLarge = Math.max(stats.value.farmers - landless - small, 0)

  return [
    {
      title: 'Irrigation Coverage',
      subtitle: 'Water access across plotted households',
      totalLabel: `${stats.value.total.toLocaleString()} HH`,
      centerLabel: 'Irrigated',
      centerValue: `${Math.max(total - stats.value.noIrrigation, 0).toLocaleString()}`,
      segments: [
        { label: 'Irrigated', value: Math.max(total - stats.value.noIrrigation, 0), color: '#22c55e' },
        { label: 'No Irrigation', value: stats.value.noIrrigation, color: '#ef4444' },
      ],
    },
    {
      title: 'Crop Seasons',
      subtitle: 'Kharif and rabi participation',
      totalLabel: `${stats.value.farmers.toLocaleString()} farmers`,
      centerLabel: 'Active',
      centerValue: `${(stats.value.kharif + stats.value.rabi).toLocaleString()}`,
      segments: [
        { label: 'Kharif', value: stats.value.kharif, color: '#f59e0b' },
        { label: 'Rabi', value: stats.value.rabi, color: '#38bdf8' },
      ],
    },
    {
      title: 'Land Holding Mix',
      subtitle: 'Agriculture footprint by holding size',
      totalLabel: `${stats.value.farmers.toLocaleString()} farmers`,
      centerLabel: 'Holding',
      centerValue: `${stats.value.farmers.toLocaleString()}`,
      segments: [
        { label: 'Marginal', value: landless, color: '#fb7185' },
        { label: 'Small', value: small, color: '#eab308' },
        { label: 'Medium/Large', value: mediumLarge, color: '#14b8a6' },
      ],
    },
  ]
})

function pieStyle(segments) {
  const total = segments.reduce((sum, seg) => sum + seg.value, 0)
  if (!total) return { background: 'conic-gradient(#334155 0deg 360deg)' }

  let start = 0
  const stops = segments.map((segment) => {
    const span = (segment.value / total) * 360
    const end = start + span
    const stop = `${segment.color} ${start}deg ${end}deg`
    start = end
    return stop
  })

  return { background: `conic-gradient(${stops.join(', ')})` }
}

function getConditionColor(house) {
  const lat = (house.latrine || '').toLowerCase()
  if (!lat || lat === 'no latrine' || lat === 'none') return '#ef4444'
  if (lat.includes('pit') || lat.includes('open'))   return '#f59e0b'
  return '#22c55e'
}

function getConditionLabel(house) {
  const color = getConditionColor(house)
  if (color === '#ef4444') return 'High Risk'
  if (color === '#f59e0b') return 'Needs Attention'
  return 'Good Standing'
}

function getMarkerColor(house) {
  if (colorMode.value === 'crops') {
    const k = (house.kharif || '').toLowerCase() === 'yes'
    const r = (house.rabi   || '').toLowerCase() === 'yes'
    if (k && r)  return '#10b981'
    if (k)       return '#f59e0b'
    if (r)       return '#38bdf8'
    return '#64748b'
  }
  if (colorMode.value === 'land') {
    const acres = parseFloat(house.totalLand) || 0
    if (acres === 0)   return '#64748b'
    if (acres <= 1)    return '#ef4444'
    if (acres <= 2.5)  return '#f59e0b'
    if (acres <= 5)    return '#22c55e'
    return '#10b981'
  }
  // sanitation (default)
  const lat   = (house.latrine  || '').toLowerCase()
  const light = (house.lighting || '').toLowerCase()
  const hasToilet = lat   && lat   !== 'no latrine' && lat   !== 'none'
  const hasElec   = light && light !== 'kerosene'   && light !== 'none'
  if (!hasToilet && !hasElec) return '#ef4444'
  if (!hasToilet || !hasElec) return '#f59e0b'
  return '#22c55e'
}

const headerLegend = computed(() => {
  if (colorMode.value === 'crops') return [
    { color: '#10b981', label: 'Both Seasons' },
    { color: '#f59e0b', label: 'Kharif Only' },
    { color: '#38bdf8', label: 'Rabi Only' },
    { color: '#64748b', label: 'No Crops' },
  ]
  if (colorMode.value === 'land') return [
    { color: '#10b981', label: 'Large >5ac' },
    { color: '#22c55e', label: 'Medium 2.5-5ac' },
    { color: '#f59e0b', label: 'Small 1-2.5ac' },
    { color: '#ef4444', label: 'Marginal ≤1ac' },
  ]
  return [
    { color: '#ef4444', label: 'No Sanitation' },
    { color: '#f59e0b', label: 'Partial' },
    { color: '#22c55e', label: 'Good' },
  ]
})

watch(colorMode, () => {
  markerRefs.forEach(({ marker, house }) => {
    marker.setStyle({ fillColor: getMarkerColor(house) })
  })
})

// ── Village cluster helpers ─────────────────────────────────────────────────
function pct(val, total) {
  if (!total) return 0
  return Math.round((val / total) * 100)
}

function clusterColor(cluster) {
  // Color by household density/coverage level
  if (cluster.count >= 100) return '#10b981'
  if (cluster.count >= 30)  return '#f59e0b'
  return '#ef4444'
}

function clusterRadius(cluster, maxCount) {
  // Scale circle radius: min 4km, max 18km, logarithmically
  const ratio = Math.log(cluster.count + 1) / Math.log(maxCount + 1)
  return 4000 + ratio * 14000  // meters
}

function issueClass(val, total) {
  const p = pct(val, total)
  if (p >= 60) return 'vstat-bad'
  if (p >= 30) return 'vstat-warn'
  return 'vstat-ok'
}

function buildVillageClusters(rows) {
  const buckets = new Map()

  rows.forEach((house) => {
    if (typeof house.latitude !== 'number' || typeof house.longitude !== 'number') return
    const villageId = String(house.villageId || '').trim()
    const latKey = Math.round(house.latitude * 20) / 20
    const lngKey = Math.round(house.longitude * 20) / 20
    const key = villageId ? `village:${villageId}` : `${latKey}:${lngKey}`
    const existing = buckets.get(key) || {
      latitude: latKey,
      longitude: lngKey,
      count: 0,
      noToilet: 0,
      noElec: 0,
      noIrrig: 0,
      bpl: 0,
      name: house.villageName || (villageId ? `Village ${villageId}` : `Village Cluster ${buckets.size + 1}`),
      level: villageId ? 'Village ID' : 'Live Cluster',
    }

    existing.count += 1
    const latrine = (house.latrine || '').toLowerCase()
    const lighting = (house.lighting || '').toLowerCase()
    const water = (house.waterSource || '').toLowerCase()
    const ration = (house.rationCard || '').toLowerCase()

    if (!latrine || latrine === 'no latrine' || latrine === 'none') existing.noToilet += 1
    if (!lighting || lighting === 'kerosene' || lighting === 'none') existing.noElec += 1
    if (!water || water === 'rain fed' || water === 'none') existing.noIrrig += 1
    if (ration.includes('bpl') || ration.includes('antyodaya')) existing.bpl += 1

    buckets.set(key, existing)
  })

  return [...buckets.values()]
    .filter(cluster => cluster.count > 0)
    .sort((a, b) => b.count - a.count)
}

const clusterIssues = computed(() => {
  const cl = selectedCluster.value
  if (!cl) return []
  return [
    { label: 'No Sanitation', pct: pct(cl.noToilet, cl.count), color: '#ef4444' },
    { label: 'No Electricity', pct: pct(cl.noElec, cl.count),  color: '#f59e0b' },
    { label: 'No Irrigation',  pct: pct(cl.noIrrig, cl.count), color: '#a78bfa' },
    { label: 'BPL Households', pct: pct(cl.bpl, cl.count),     color: '#60a5fa' },
  ]
})

function clearClusterSelection() {
  selectedCluster.value = null
  if (highlightCircle) {
    highlightCircle.remove()
    highlightCircle = null
  }
}

function drawClusters(clusters) {
  if (clusterGroup) { clusterGroup.remove(); clusterGroup = null }
  clearClusterSelection()
  if (!map || !clusters.length) return

  clusterGroup = L.layerGroup().addTo(map)
  const maxCount = Math.max(...clusters.map(c => c.count))

  clusters.forEach(cluster => {
    const color  = clusterColor(cluster)
    const radius = clusterRadius(cluster, maxCount)

    // Filled translucent area circle
    const areaCircle = L.circle([cluster.latitude, cluster.longitude], {
      radius,
      fillColor: color,
      fillOpacity: 0.18,
      color: color,
      weight: 1.5,
      opacity: 0.55,
    }).addTo(clusterGroup)

    // Center dot
    const dot = L.circleMarker([cluster.latitude, cluster.longitude], {
      radius: 7,
      fillColor: color,
      color: '#fff',
      weight: 2,
      fillOpacity: 1,
    }).addTo(clusterGroup)

    // Label
    const label = L.marker([cluster.latitude, cluster.longitude], {
      icon: L.divIcon({
        className: '',
        html: `<div class="cluster-label">${cluster.name}<br/><span>${cluster.count} HH</span></div>`,
        iconSize: [120, 36],
        iconAnchor: [60, -10],
      }),
      interactive: false,
    }).addTo(clusterGroup)

    // Click: highlight + show detail
    const onClick = () => {
      selectedCluster.value = cluster

      // Remove old highlight
      if (highlightCircle) { highlightCircle.remove() }

      highlightCircle = L.circle([cluster.latitude, cluster.longitude], {
        radius: radius * 1.05,
        fillColor: color,
        fillOpacity: 0.1,
        color: color,
        weight: 3,
        opacity: 0.9,
        dashArray: '8,5',
      }).addTo(map)

      map.flyTo([cluster.latitude, cluster.longitude], 11, { duration: 1 })
    }

    areaCircle.on('click', onClick)
    dot.on('click', onClick)
    areaCircle.bindTooltip(`<strong>${cluster.name}</strong><br/>${cluster.count} households`, { className: 'map-tooltip' })
  })

  // Fit map to all clusters
  const bounds = L.latLngBounds(clusters.map(c => [c.latitude, c.longitude]))
  map.fitBounds(bounds, { padding: [60, 60] })
}

function showPointLayer() {
  if (clusterGroup) { clusterGroup.remove(); clusterGroup = null }
  clearClusterSelection()
  // Markers are already on the map from plotMarkers; just need to show them
  markerRefs.forEach(({ marker }) => marker.addTo(map))
}

function hidePointLayer() {
  markerRefs.forEach(({ marker }) => map.removeLayer(marker))
}

async function setViewMode(mode) {
  viewMode.value = mode
  if (mode === 'villages') {
    hidePointLayer()
    drawClusters(buildVillageClusters(houses.value))
  } else {
    if (clusterGroup) { clusterGroup.remove(); clusterGroup = null }
    clearClusterSelection()
    showPointLayer()
  }
}

function isDarkTheme() {
  return document.documentElement.getAttribute('data-theme') !== 'light'
}

function addTiles(mapInstance) {
  const dark = isDarkTheme()
  if (dark) {
    L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
      attribution: '© <a href="https://www.openstreetmap.org/copyright">OSM</a> © <a href="https://carto.com/">CARTO</a>',
      subdomains: 'abcd', maxZoom: 19,
    }).addTo(mapInstance)
  } else {
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      subdomains: 'abc', maxZoom: 19,
    }).addTo(mapInstance)
  }
}

async function addMaharashtraHighlight(mapInstance) {
  try {
    const res = await fetch('https://raw.githubusercontent.com/geohacker/india/master/state/india_state.geojson')
    const data = await res.json()
    const mh = data.features.find(f =>
      Object.values(f.properties || {}).some(v => String(v).toUpperCase().includes('MAHARASHTRA'))
    )
    if (mh) {
      L.geoJSON(mh, {
        style: { color: '#f59e0b', weight: 2.5, opacity: 0.9, fillColor: '#f59e0b', fillOpacity: 0.05, dashArray: '8,5' },
      }).addTo(mapInstance).bringToBack()

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

function plotMarkers(data) {
  clearMarkers()
  data.forEach(house => {
    const color  = getMarkerColor(house)
    const marker = L.circleMarker([house.latitude, house.longitude], {
      radius: 6, fillColor: color, color: '#fff',
      weight: 1.5, opacity: 1, fillOpacity: 0.88,
    }).addTo(map)
    markerRefs.push({ marker, house })
    marker.on('click', (e) => { L.DomEvent.stopPropagation(e); selectedHouse.value = house })
    marker.bindTooltip(`
      <strong>${house.headName || 'Household'}</strong><br/>
      Land: ${house.totalLand || '0'} acres · Kharif: ${house.kharif || '—'} · Rabi: ${house.rabi || '—'}
    `, { className: 'map-tooltip' })
  })
  const bounds = L.latLngBounds(data.map(h => [h.latitude, h.longitude]))
  map.fitBounds(bounds, { padding: [40, 40] })
}

function clearRetryTimer() {
  if (retryTimer) {
    clearTimeout(retryTimer)
    retryTimer = null
  }
}

async function loadLiveHouseData(attempt = 0) {
  try {
    const real = await fetchAllHouses()
    if (real.length > 0) {
      clearRetryTimer()
      houses.value = real
      plotMarkers(real)
      if (viewMode.value === 'villages') {
        drawClusters(buildVillageClusters(real))
      }
      loading.value = false
      return
    }

    if (attempt < 10 && !selectedDistrict.value && !selectedTaluka.value && !selectedVillage.value) {
      retryTimer = setTimeout(() => loadLiveHouseData(attempt + 1), 3000)
      return
    }
  } catch (e) {
    if (attempt < 10 && !selectedDistrict.value && !selectedTaluka.value && !selectedVillage.value) {
      retryTimer = setTimeout(() => loadLiveHouseData(attempt + 1), 3000)
      return
    }
    console.warn('Houses API not available:', e.message)
  }

  loading.value = false
}

onMounted(async () => {
  await nextTick()

  if (mapContainer.value) {
    map = L.map(mapContainer.value, { center: [19.75, 75.71], zoom: 7, zoomControl: false, doubleClickZoom: false })
    L.control.zoom({ position: 'topright' }).addTo(map)
    addTiles(map)
    addMaharashtraHighlight(map)
    setTimeout(handleMapResize, 60)
    setTimeout(handleMapResize, 250)
    window.addEventListener('resize', handleMapResize)
    window.addEventListener('click', closeDropdowns)
  }

  loading.value = true
  await loadLocationDropdowns()
  applyFilters()
})

onUnmounted(() => {
  clearRetryTimer()
  window.removeEventListener('resize', handleMapResize)
  window.removeEventListener('click', closeDropdowns)
  if (map) { map.remove(); map = null }
})

watch(selectedDistrict, async () => {
  selectedTaluka.value = ''
  selectedVillage.value = ''
  await loadLocationDropdowns()
})

watch(selectedTaluka, async () => {
  selectedVillage.value = ''
  await loadLocationDropdowns()
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
  padding: 1rem 2rem;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border);
  z-index: 20;
  flex-shrink: 0;
}

.page-title { font-family: var(--font-display); font-size: 1.5rem; color: var(--text-primary); font-weight: 400; }
.page-subtitle { color: var(--text-dim); font-size: 0.75rem; margin-top: 0.2rem; display: flex; align-items: center; gap: 0.5rem; }

.map-controls { display: flex; align-items: center; gap: 1.25rem; flex-wrap: wrap; }
.map-control-group { display: flex; align-items: center; gap: 0.45rem; }
.control-label { font-size: 0.62rem; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-dim); white-space: nowrap; }
/* ── Custom Select Dropdowns — no native <select>, immune to OS dark mode ── */
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
  background: #ffffff !important;      /* defeat dark-theme inheritance */
  border: 1px solid #d1d5db !important;
  border-radius: 6px;
  color: #334155 !important;
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
  border-color: #94a3b8 !important;
  background: #f9fafb !important;
}
.custom-select.open .cs-trigger {
  border-color: #14b8a6 !important;
  box-shadow: 0 0 0 2px rgba(20, 184, 166, 0.18);
}
.custom-select.disabled .cs-trigger,
.cs-trigger:disabled {
  background: #f3f4f6 !important;
  color: #9ca3af !important;
  border-color: #e5e7eb !important;
  cursor: not-allowed;
  opacity: 0.7;
}

.cs-value {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #334155 !important;
}
.cs-arrow {
  font-size: 0.58rem;
  color: #64748b !important;
  flex-shrink: 0;
  transition: transform 0.15s;
  line-height: 1;
}
.custom-select.open .cs-arrow { transform: rotate(180deg); }

/* Dropdown panel — hardcoded white, defeats dark theme entirely */
.cs-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 100%;
  max-height: 220px;
  overflow-y: auto;
  background: #ffffff !important;
  border: 1px solid #d1d5db !important;
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.18), 0 3px 8px rgba(0,0,0,0.10);
  z-index: 9999;          /* above everything — Leaflet, headers, overlays */
  scrollbar-width: thin;
  scrollbar-color: #e2e8f0 transparent;
  /* Ensure it is never clipped by parent overflow */
  isolation: isolate;
}
.cs-dropdown-right { left: auto; right: 0; }

/* Option rows — all hardcoded, zero CSS variable inheritance */
.cs-option {
  padding: 0.42rem 0.75rem;
  font-size: 0.76rem;
  color: #1e293b !important;
  background: #ffffff !important;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.1s, color 0.1s;
  user-select: none;
}
.cs-option:first-child { border-radius: 8px 8px 0 0; }
.cs-option:last-child  { border-radius: 0 0 8px 8px; }
.cs-option:hover {
  background: #f0fdfa !important;
  color: #0f766e !important;
}
.cs-option.selected {
  background: #ccfbf1 !important;
  color: #0f766e !important;
  font-weight: 600;
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

.map-legend { display: flex; gap: 0.75rem; flex-wrap: wrap; }
.legend-item { display: flex; align-items: center; gap: 0.35rem; font-size: 0.7rem; color: var(--text-muted); }
.legend-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }

.empty-state {
  margin: 0 2rem 1rem;
  padding: 0.85rem 1rem;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--bg-card);
  color: var(--text-muted);
  font-size: 0.8rem;
}

.analytics-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
  padding: 1rem 2rem 0.75rem;
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

.map-shell {
  padding: 0 2rem 1.5rem;
  flex: 1;
  min-height: 0;
}

.map-content {
  position: relative;
  height: 100%;
  min-height: 520px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 12px 32px var(--shadow);
}

.map-container { position: absolute; inset: 0; z-index: 1; }

/* ═══════════════════════════════════════════════
   DETAIL PANEL — household click popup
═══════════════════════════════════════════════ */
.detail-panel {
  position: absolute;
  top: 1rem; right: 1rem;
  width: 320px;
  max-height: calc(100% - 2rem);
  overflow-y: auto;
  z-index: 500;
  background: #ffffff;
  border: 1.5px solid #e2e8f0;
  border-radius: 12px;
  box-shadow: 0 12px 40px rgba(0,0,0,0.16), 0 4px 12px rgba(0,0,0,0.08);
  scrollbar-width: thin;
  scrollbar-color: #cbd5e1 transparent;
}

/* Header */
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

/* Section labels */
.dp-section-label {
  display: flex; align-items: center; gap: 0.4rem;
  font-size: 0.6rem; text-transform: uppercase; letter-spacing: 0.09em;
  color: #475569; font-weight: 800;
  padding: 0.85rem 1rem 0.35rem;
  border-top: 1px solid #f1f5f9;
}
.dp-section-icon { font-size: 0.85rem; }

/* Big stat row */
.dp-stat-row {
  display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem;
  padding: 0 1rem;
}
.dp-stat {
  background: #f8fafc; border: 1.5px solid #e2e8f0;
  border-radius: 8px; padding: 0.6rem 0.75rem; text-align: center;
}
.dp-stat-val { font-size: 1.25rem; font-weight: 800; color: #0f172a; line-height: 1.1; }
.dp-stat-val small { font-size: 0.65rem; color: #64748b; font-weight: 600; }
.dp-stat-key { font-size: 0.57rem; text-transform: uppercase; letter-spacing: 0.06em; color: #94a3b8; font-weight: 600; margin-top: 0.18rem; }

/* Crop chips */
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

/* Field rows */
.dp-field-row {
  display: flex; align-items: center; gap: 0.55rem;
  padding: 0.55rem 1rem;
  border-bottom: 1px solid #f8fafc;
}
.dp-field-icon { font-size: 0.85rem; flex-shrink: 0; width: 1.2rem; text-align: center; }
.dp-field-key  { font-size: 0.68rem; color: #64748b; font-weight: 600; flex: 1; }
.dp-field-val  { font-size: 0.76rem; color: #0f172a; font-weight: 700; text-align: right; max-width: 55%; }

.panel-coords {
  margin-top: 0.9rem;
  padding-top: 0.6rem;
  border-top: 1px solid var(--border);
  font-size: 0.68rem;
  color: var(--text-dim);
  font-variant-numeric: tabular-nums;
  text-align: center;
}

/* View mode toggle */
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

/* Village detail panel extras */
.village-panel { min-width: 280px; }
.village-badge {
  display: inline-block;
  font-size: 0.6rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--teal);
  border: 1px solid var(--teal);
  border-radius: 999px;
  padding: 0.1rem 0.5rem;
  margin-bottom: 0.5rem;
}

.village-stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
  margin-bottom: 1rem;
}
.vstat {
  border-radius: 8px;
  padding: 0.55rem 0.6rem;
  text-align: center;
}
.vstat-bad  { background: rgba(239,68,68,0.12); border: 1px solid rgba(239,68,68,0.3); }
.vstat-warn { background: rgba(245,158,11,0.12); border: 1px solid rgba(245,158,11,0.3); }
.vstat-ok   { background: rgba(34,197,94,0.10); border: 1px solid rgba(34,197,94,0.25); }
.vstat-val {
  font-family: var(--font-display);
  font-size: 1.3rem;
  font-weight: 500;
  line-height: 1;
  color: var(--text-primary);
}
.vstat-label {
  font-size: 0.6rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--text-dim);
  margin-top: 0.2rem;
}

/* Issue bar charts */
.village-bar-section {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  margin-bottom: 0.75rem;
}
.vbar-row {
  display: grid;
  grid-template-columns: 90px 1fr 34px;
  align-items: center;
  gap: 0.5rem;
}
.vbar-label { font-size: 0.68rem; color: var(--text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.vbar-track {
  height: 6px;
  background: var(--bg-surface);
  border-radius: 999px;
  overflow: hidden;
}
.vbar-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.4s ease;
}
.vbar-pct { font-size: 0.65rem; color: var(--text-dim); text-align: right; font-variant-numeric: tabular-nums; }

.slide-enter-active { transition: all 0.25s ease-out; }
.slide-leave-active { transition: all 0.15s ease-in; }
.slide-enter-from, .slide-leave-to { opacity: 0; transform: translateX(20px); }

@media (max-width: 1100px) {
  .analytics-grid {
    grid-template-columns: 1fr;
  }

  .chart-layout {
    grid-template-columns: 96px 1fr;
  }
}

@media (max-width: 760px) {
  .map-header,
  .analytics-grid,
  .map-shell {
    padding-left: 1rem;
    padding-right: 1rem;
  }

  .map-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .map-legend {
    flex-wrap: wrap;
  }

  .chart-layout {
    grid-template-columns: 1fr;
  }

  .donut {
    margin: 0 auto;
  }

  .detail-panel {
    width: calc(100% - 2rem);
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
.leaflet-control-zoom a {
  background: var(--bg-card) !important;
  color: var(--text-body) !important;
  border-color: var(--border) !important;
}
.leaflet-control-zoom a:hover { background: var(--bg-surface) !important; }
.leaflet-control-attribution {
  background: var(--bg-card) !important;
  color: var(--text-dim) !important;
  font-size: 0.6rem !important;
}
.leaflet-control-attribution a { color: var(--text-muted) !important; }
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
.cluster-label {
  font-family: sans-serif;
  font-size: 0.7rem;
  font-weight: 600;
  color: #f1f5f9;
  text-shadow: 0 1px 4px rgba(0,0,0,0.7);
  white-space: nowrap;
  pointer-events: none;
  text-align: center;
  line-height: 1.3;
}
.cluster-label span {
  font-size: 0.6rem;
  font-weight: 400;
  color: #94a3b8;
}
</style>
