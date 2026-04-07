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

        <div class="map-control-group">
          <label class="control-label">District</label>
          <select v-model="selectedDistrict" class="control-select">
            <option value="">All</option>
            <option v-for="d in districtOptions" :key="d.id" :value="d.id">{{ d.name }}</option>
          </select>
        </div>

        <div class="map-control-group">
          <label class="control-label">Taluka</label>
          <select v-model="selectedTaluka" class="control-select" :disabled="!talukaOptions.length">
            <option value="">All</option>
            <option v-for="t in talukaOptions" :key="t.id" :value="t.id">{{ t.name }}</option>
          </select>
        </div>

        <div class="map-control-group">
          <label class="control-label">Village</label>
          <select v-model="selectedVillage" class="control-select village-select" :disabled="!villageOptions.length">
            <option value="">All</option>
            <option v-for="v in villageOptions" :key="v.id" :value="v.id">{{ v.name }}</option>
          </select>
        </div>

        <div class="map-control-group">
          <button class="apply-btn" @click="applyFilters">Apply</button>
          <button class="reset-btn" @click="resetFilters">Reset</button>
        </div>

        <!-- Color by (only in points mode) -->
        <div class="map-control-group" v-if="viewMode === 'points'">
          <label class="control-label">Color by</label>
          <select v-model="colorMode" class="control-select">
            <option value="sanitation">Sanitation</option>
            <option value="crops">Crops / Season</option>
            <option value="land">Land Holdings</option>
          </select>
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
            <button class="panel-close" @click="selectedHouse = null">×</button>
            <h3 class="panel-title">{{ selectedHouse.headName || 'Household' }}</h3>
            <div class="panel-id">ID: {{ selectedHouse.familyId }}</div>
            <div class="panel-grid">
              <div class="panel-stat" v-for="s in detailStats" :key="s.label">
                <div class="panel-stat-label">{{ s.label }}</div>
                <div class="panel-stat-value" :style="s.style || {}">{{ s.value }}</div>
              </div>
            </div>
            <div class="panel-coords">{{ selectedHouse.latitude.toFixed(6) }}, {{ selectedHouse.longitude.toFixed(6) }}</div>
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
import { getHouses, getLocationOptions } from '../api/index.js'
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
    marker.on('click', () => { selectedHouse.value = house })
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
    map = L.map(mapContainer.value, { center: [19.75, 75.71], zoom: 7, zoomControl: false })
    L.control.zoom({ position: 'topright' }).addTo(map)
    addTiles(map)
    addMaharashtraHighlight(map)
    setTimeout(handleMapResize, 60)
    setTimeout(handleMapResize, 250)
    window.addEventListener('resize', handleMapResize)
  }

  loading.value = true
  await loadLocationDropdowns()
  applyFilters()
})

onUnmounted(() => {
  clearRetryTimer()
  window.removeEventListener('resize', handleMapResize)
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
