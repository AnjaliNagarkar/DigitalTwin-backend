<template>
  <div class="map-page">
    <header class="map-header">
      <div class="map-title-area">
        <h1 class="page-title">Population Geo-Intelligence Map</h1>
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
          <select v-model="selectedDistrict" class="control-select">
            <option value="">All</option>
            <option v-for="district in districtOptions" :key="district.id" :value="district.id">
              {{ district.name }}
            </option>
          </select>
        </div>

        <div class="map-control-group">
          <label class="control-label">Taluka</label>
          <select v-model="selectedTaluka" class="control-select" :disabled="!talukaOptions.length">
            <option value="">All</option>
            <option v-for="taluka in talukaOptions" :key="taluka.id" :value="taluka.id">
              {{ taluka.name }}
            </option>
          </select>
        </div>

        <div class="map-control-group">
          <label class="control-label">Village</label>
          <select v-model="selectedVillage" class="control-select village-select" :disabled="!villageOptions.length">
            <option value="">All</option>
            <option v-for="village in villageOptions" :key="village.id" :value="village.id">
              {{ village.name }}
            </option>
          </select>
        </div>

        <div class="map-control-group">
          <button class="apply-btn" @click="applyFilters">Apply</button>
          <button class="reset-btn" @click="resetFilters">Reset</button>
        </div>

        <div class="map-control-group" v-if="viewMode === 'points'">
          <label class="control-label">Color by</label>
          <select v-model="colorMode" class="control-select">
            <option value="population_density">Population Density</option>
            <option value="bpl_status">BPL Status</option>
            <option value="literacy">Literacy</option>
            <option value="working_population">Working Population</option>
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
            <div class="legend-row" v-for="segment in card.segments" :key="segment.label">
              <span class="legend-dot" :style="{ background: segment.color }"></span>
              <span class="legend-name">{{ segment.label }}</span>
              <span class="legend-value">{{ segment.value.toLocaleString() }}</span>
            </div>
          </div>
        </div>
      </article>
    </section>

    <section class="map-shell">
      <div v-if="!loading && !markers.length" class="empty-state">
        No live population map data returned from the database API.
      </div>

      <div class="map-content">
        <div class="map-container" ref="mapContainer"></div>

        <transition name="slide">
          <aside v-if="selectedMarker && viewMode === 'points'" class="detail-panel">
            <button class="panel-close" @click="selectedMarker = null">×</button>
            <h3 class="panel-title">{{ selectedMarker.head_name || 'Household' }}</h3>
            <div class="panel-id">House No: {{ selectedMarker.house_no || 'N/A' }}</div>
            <div class="panel-grid">
              <div class="panel-stat" v-for="item in detailStats" :key="item.label">
                <div class="panel-stat-label">{{ item.label }}</div>
                <div class="panel-stat-value" :style="item.style || {}">{{ item.value }}</div>
              </div>
            </div>
            <div class="panel-coords">
              {{ selectedMarker.lat.toFixed(6) }}, {{ selectedMarker.lng.toFixed(6) }}
            </div>
          </aside>
        </transition>

        <transition name="slide">
          <aside v-if="selectedCluster" class="detail-panel village-panel">
            <button class="panel-close" @click="clearClusterSelection">×</button>
            <div class="village-badge">{{ selectedCluster.level }}</div>
            <h3 class="panel-title">{{ selectedCluster.name }}</h3>
            <div class="panel-id">{{ selectedCluster.count.toLocaleString() }} households covered</div>

            <div class="village-stats">
              <div class="vstat" :class="issueClass(selectedCluster.bpl, selectedCluster.count)">
                <div class="vstat-val">{{ pct(selectedCluster.bpl, selectedCluster.count) }}%</div>
                <div class="vstat-label">BPL Families</div>
              </div>
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
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import L from 'leaflet'
import { getLocationOptions } from '../../api/index.js'
import { getPopulationMapData, getPopulationMapInsights, getPopulationMapSummary } from './api.js'

const mapContainer = ref(null)
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

let map = null
let markerLayer = null
let clusterLayer = null
let requestToken = 0

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

function markerRadius(totalMembers) {
  return Math.max(5, Math.min(12, 4 + totalMembers * 0.35))
}

function householdColor(totalMembers) {
  if (totalMembers >= 8) return '#ef4444'
  if (totalMembers >= 4) return '#f59e0b'
  return '#0f766e'
}

function buildQueryParams() {
  const params = {}
  if (selectedDistrict.value) params.district_id = selectedDistrict.value
  if (selectedTaluka.value) params.taluka_id = selectedTaluka.value
  if (selectedVillage.value) params.village_id = selectedVillage.value
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

function issueClass(value, total) {
  const share = pct(value, total)
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
      latitude: latKey,
      longitude: lngKey,
      count: 0,
      bpl: 0,
      noToilet: 0,
      noElec: 0,
      noIrrig: 0,
      name: house.village_name || (villageId ? `Village ${villageId}` : `Village Cluster ${buckets.size + 1}`),
      level: villageId ? 'Village ID' : 'Live Cluster',
    }

    existing.count += 1
    const latrine = (house.latrine || '').toLowerCase()
    const lighting = (house.lighting || '').toLowerCase()
    const water = (house.water_source || '').toLowerCase()
    const ration = (house.ration_card || '').toLowerCase()

    if (!latrine || latrine === 'no latrine' || latrine === 'none') existing.noToilet += 1
    if (!lighting || lighting === 'kerosene' || lighting === 'none') existing.noElec += 1
    if (!water || water === 'rain fed' || water === 'none') existing.noIrrig += 1
    if (ration.includes('bpl') || ration.includes('antyodaya')) existing.bpl += 1

    buckets.set(key, existing)
  })

  return [...buckets.values()].sort((a, b) => b.count - a.count)
}

const clusterIssues = computed(() => {
  const cluster = selectedCluster.value
  if (!cluster) return []
  return [
    { label: 'BPL Families', pct: pct(cluster.bpl, cluster.count), color: '#60a5fa' },
    { label: 'No Sanitation', pct: pct(cluster.noToilet, cluster.count), color: '#ef4444' },
    { label: 'No Electricity', pct: pct(cluster.noElec, cluster.count), color: '#f59e0b' },
    { label: 'No Irrigation', pct: pct(cluster.noIrrig, cluster.count), color: '#a78bfa' },
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

function renderMarkers() {
  if (!map || !markerLayer) return

  clearMarkerLayer()
  selectedMarker.value = null

  markers.value.forEach((marker) => {
    const circle = L.circleMarker([marker.lat, marker.lng], {
      radius: markerRadius(marker.total_members),
      color: '#0f172a',
      weight: 1,
      fillColor: householdColor(marker.total_members),
      fillOpacity: 0.82,
    })

    circle.bindTooltip([
      `<strong>${marker.head_name || 'Household'}</strong>`,
      `House No: ${marker.house_no || 'N/A'}`,
      `Family Members: ${marker.total_members.toLocaleString()}`,
    ].join('<br/>'), { sticky: true, direction: 'top', opacity: 0.96 })

    circle.on('click', () => {
      selectedMarker.value = marker
      selectedCluster.value = null
    })

    circle.addTo(markerLayer)
  })

  fitMapToMarkers()
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

    const circle = L.circle([cluster.latitude, cluster.longitude], {
      radius,
      fillColor: color,
      fillOpacity: 0.18,
      color,
      weight: 1.5,
      opacity: 0.55,
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
    console.error('Population map load failed:', error)
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
  return [
    { label: 'Members', value: marker.total_members.toLocaleString() },
    { label: 'Latitude', value: marker.lat.toFixed(5) },
    { label: 'Longitude', value: marker.lng.toFixed(5) },
  ]
})

const headerLegend = computed(() => ([
  { color: '#ef4444', label: 'BPL households' },
  { color: '#0f766e', label: 'Literate population' },
  { color: '#16a34a', label: 'Working population' },
]))

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

function handleResize() {
  if (map) {
    map.invalidateSize()
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

onMounted(async () => {
  await loadLocationOptions()

  map = L.map(mapContainer.value, {
    zoomControl: true,
    preferCanvas: true,
  }).setView([19.7515, 75.7139], 6)

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map)

  markerLayer = L.layerGroup().addTo(map)
  clusterLayer = L.layerGroup().addTo(map)

  await fetchMapData()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  requestToken += 1
  window.removeEventListener('resize', handleResize)
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
  padding: 1rem 2rem;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border);
  z-index: 20;
  flex-shrink: 0;
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
  gap: 1.25rem;
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

.map-legend {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.7rem;
  color: var(--text-muted);
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
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
