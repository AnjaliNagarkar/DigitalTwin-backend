<template>
  <div class="twin-page">
    <div class="cesium-wrap" ref="cesiumContainer"></div>

    <div class="topbar">
      <div class="topbar-brand">
        <span class="brand-dot"></span>
        <span class="brand-name">PopTwin</span>
        <span class="brand-sub">3D Digital Twin</span>
      </div>

      <div class="filter-bar">
        <div class="filter-group">
          <label class="filter-label">District</label>
          <div class="custom-select" :class="{ open: openDropdown === 'district' }" @click.stop="toggleDropdown('district')">
            <button class="cs-trigger" type="button">
              <span class="cs-value">{{ selectedDistrictLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'district'" @click.stop>
              <div class="cs-option" :class="{ selected: !pendingDistrict }" @click="selectDistrict('')">All Districts</div>
              <div class="cs-option" v-for="d in districtOptions" :key="d.id" :class="{ selected: String(pendingDistrict) === String(d.id) }" @click="selectDistrict(d.id)">{{ d.name }}</div>
            </div>
          </div>
        </div>

        <span class="filter-arrow">›</span>

        <div class="filter-group">
          <label class="filter-label">Taluka</label>
          <div class="custom-select" :class="{ open: openDropdown === 'taluka', disabled: !pendingDistrict }" @click.stop="pendingDistrict && toggleDropdown('taluka')">
            <button class="cs-trigger" type="button" :disabled="!pendingDistrict">
              <span class="cs-value">{{ selectedTalukaLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'taluka'" @click.stop>
              <div class="cs-option" :class="{ selected: !pendingTaluka }" @click="selectTaluka('')">All Talukas</div>
              <div class="cs-option" v-for="t in talukaOptions" :key="t.id" :class="{ selected: String(pendingTaluka) === String(t.id) }" @click="selectTaluka(t.id)">{{ t.name }}</div>
            </div>
          </div>
        </div>

        <span class="filter-arrow">›</span>

        <div class="filter-group">
          <label class="filter-label">Village</label>
          <div class="custom-select" :class="{ open: openDropdown === 'village', disabled: !pendingTaluka }" @click.stop="pendingTaluka && toggleDropdown('village')">
            <button class="cs-trigger" type="button" :disabled="!pendingTaluka">
              <span class="cs-value">{{ selectedVillageLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'village'" @click.stop>
              <div class="cs-option" :class="{ selected: !pendingVillage }" @click="selectVillage('')">All Villages</div>
              <div class="cs-option" v-for="v in villageOptions" :key="v.id" :class="{ selected: String(pendingVillage) === String(v.id) }" @click="selectVillage(v.id)">{{ v.name }}</div>
            </div>
          </div>
        </div>

        <button class="apply-btn" @click="applyFilters" :disabled="!filtersDirty">Apply</button>
        <button class="reset-btn" @click="resetFilters" v-if="pendingDistrict || pendingTaluka || pendingVillage || filterDistrict || filterTaluka || filterVillage">✕ Reset</button>
      </div>

      <div class="topbar-right">
        <div class="ctrl-group">
          <label class="filter-label">Color by</label>
          <div class="custom-select cs-align-right" :class="{ open: openDropdown === 'colorMode' }" @click.stop="toggleDropdown('colorMode')">
            <button class="cs-trigger" type="button">
              <span class="cs-value">{{ selectedColorModeLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown cs-dropdown-right" v-show="openDropdown === 'colorMode'" @click.stop>
              <div class="cs-option" :class="{ selected: colorMode === 'population_density' }" @click="selectColorMode('population_density')">Population Density</div>
              <div class="cs-option" :class="{ selected: colorMode === 'bpl_status' }" @click="selectColorMode('bpl_status')">BPL Status</div>
              <div class="cs-option" :class="{ selected: colorMode === 'education_status' }" @click="selectColorMode('education_status')">Education Status</div>
              <div class="cs-option" :class="{ selected: colorMode === 'employment_status' }" @click="selectColorMode('employment_status')">Employment Status</div>
              <div class="cs-option" :class="{ selected: colorMode === 'divyang_presence' }" @click="selectColorMode('divyang_presence')">Divyang Presence</div>
            </div>
          </div>
        </div>

        <button class="ctrl-btn" :class="{ active: tileStyle === 'satellite' }" @click="toggleTile">
          {{ tileStyle === 'satellite' ? '🛰 Satellite' : '🗺 Street' }}
        </button>

        <div class="dl-wrap" v-if="!loadingLiveData && houses.length">
          <button class="dl-btn" type="button" title="PDF report placeholder">⬇ PDF Report</button>
          <span class="dl-count">{{ houses.length.toLocaleString() }} rows</span>
        </div>
      </div>
    </div>

    <div class="loading-overlay" v-if="loadingLiveData">
      <div class="loading-spinner"></div>
      <div class="loading-text">Loading population twin data…</div>
    </div>

    <div class="stats-bar" v-if="!loadingLiveData">
      <span class="stat-item">
        <span class="stat-dot" style="background:#16a34a"></span>
        <strong>{{ houses.length.toLocaleString() }}</strong> households
      </span>
      <span class="stat-sep">·</span>
      <span class="stat-item"><strong>{{ totalMembers.toLocaleString() }}</strong> population</span>
      <span class="stat-sep">·</span>
      <span class="stat-item">Maharashtra</span>
      <span class="stat-sep" v-if="zoomLabel">·</span>
      <span class="stat-item zoom-label" v-if="zoomLabel">{{ zoomLabel }}</span>
    </div>

    <div class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <button class="sidebar-toggle" @click="sidebarCollapsed = !sidebarCollapsed" :title="sidebarCollapsed ? 'Open panel' : 'Close panel'">
        {{ sidebarCollapsed ? '›' : '‹' }}
      </button>

      <div class="sidebar-body" v-if="!sidebarCollapsed && !loadingLiveData">
        <div class="panel-card">
          <div class="card-title">{{ legendTitle }}</div>
          <div class="legend-item" v-for="leg in currentLegend" :key="leg.label">
            <span class="legend-swatch" :style="{ background: leg.color }"></span>
            <span class="legend-text">{{ leg.label }}</span>
          </div>
          <div class="legend-note">Building roof color follows selected filter</div>
        </div>

        <div class="panel-card">
          <div class="card-title">Population Overview</div>
          <div class="issue-row">
            <span class="issue-pip" style="background:#14b8a6"></span>
            <div class="issue-body">
              <div class="issue-top"><span class="issue-name">Total Households</span><span class="issue-count">{{ houses.length.toLocaleString() }}</span></div>
              <div class="issue-track"><div class="issue-fill" style="width:100%;background:#14b8a6"></div></div>
            </div>
          </div>
          <div class="issue-row">
            <span class="issue-pip" style="background:#2563eb"></span>
            <div class="issue-body">
              <div class="issue-top"><span class="issue-name">Total Population</span><span class="issue-count">{{ totalMembers.toLocaleString() }}</span></div>
              <div class="issue-track"><div class="issue-fill" style="width:100%;background:#2563eb"></div></div>
            </div>
          </div>
        </div>

        <div class="panel-card">
          <div class="card-title">Gender Ratio</div>
          <div class="legend-item"><span class="legend-swatch" style="background:#2563eb"></span><span class="legend-text">Male {{ malePct }}%</span></div>
          <div class="legend-item"><span class="legend-swatch" style="background:#ec4899"></span><span class="legend-text">Female {{ femalePct }}%</span></div>
        </div>

        <div class="panel-card">
          <div class="card-title">Employment Summary</div>
          <div class="legend-item"><span class="legend-swatch" style="background:#16a34a"></span><span class="legend-text">Working households {{ workingHouseholds.toLocaleString() }}</span></div>
          <div class="legend-item"><span class="legend-swatch" style="background:#f59e0b"></span><span class="legend-text">Dependent households {{ dependentHouseholds.toLocaleString() }}</span></div>
        </div>

        <div class="panel-card">
          <div class="card-title">Education Summary</div>
          <div class="legend-item"><span class="legend-swatch" style="background:#16a34a"></span><span class="legend-text">Literacy Rate {{ literacyRate }}%</span></div>
        </div>

        <div class="panel-card">
          <div class="card-title">Divyang Summary</div>
          <div class="legend-item"><span class="legend-swatch" style="background:#7b1fa2"></span><span class="legend-text">Households with disability {{ divyangHouseholds.toLocaleString() }}</span></div>
        </div>
      </div>
    </div>

    <div v-if="hoveredHouse" class="hover-card" :style="{ left: mouseX + 'px', top: mouseY + 'px' }">
      <div class="hover-name">{{ hoveredHouse.head_name || 'Household' }}</div>
      <div class="hover-row"><span class="hr-key">House No</span><span class="hr-val">{{ hoveredHouse.house_no || '—' }}</span></div>
      <div class="hover-row"><span class="hr-key">Members</span><span class="hr-val">{{ Number(hoveredHouse.total_members || 0) }}</span></div>
      <div class="hover-hint">Click for details · Double-click to zoom</div>
    </div>

    <transition name="slide">
      <div v-if="selectedHouse" class="detail-panel">
        <div class="detail-header">
          <div>
            <div class="detail-badge" :style="{ background: getConditionColor(selectedHouse) + '18', borderColor: getConditionColor(selectedHouse) + '60', color: getConditionColor(selectedHouse) }">
              {{ selectedColorModeLabel }}
            </div>
            <div class="detail-name">{{ selectedHouse.head_name || 'Household' }}</div>
            <div class="detail-sub">House {{ selectedHouse.house_no || 'N/A' }}</div>
          </div>
          <button class="detail-close" @click="selectedHouse = null">×</button>
        </div>

        <button class="focus-btn" @click="flyToHouse(selectedHouse)">📍 Zoom to House</button>

        <div class="detail-section">Population</div>
        <div class="kv-grid">
          <div class="kv"><span class="kv-k">Members</span><span class="kv-v">{{ Number(selectedHouse.total_members || 0) }}</span></div>
          <div class="kv"><span class="kv-k">Male</span><span class="kv-v">{{ Number(selectedHouse.male_members || 0) }}</span></div>
          <div class="kv"><span class="kv-k">Female</span><span class="kv-v">{{ Number(selectedHouse.female_members || 0) }}</span></div>
          <div class="kv" v-if="colorMode === 'employment_status'"><span class="kv-k">Working Members</span><span class="kv-v">{{ Number(selectedHouse.working_members || 0) }}</span></div>
          <div class="kv" v-if="colorMode === 'employment_status'"><span class="kv-k">Non-working</span><span class="kv-v">{{ Math.max(Number(selectedHouse.total_members || 0) - Number(selectedHouse.working_members || 0), 0) }}</span></div>
          <div class="kv" v-if="colorMode === 'employment_status'"><span class="kv-k">Occupations</span><span class="kv-v">{{ selectedHouse.working_occupations || selectedHouse.occupation_list || 'N/A' }}</span></div>
          <div class="kv" v-if="colorMode === 'divyang_presence'"><span class="kv-k">Disability</span><span class="kv-v">{{ Number(selectedHouse.has_disability || 0) === 1 ? 'Yes' : 'No' }}</span></div>
          <div class="kv" v-if="colorMode === 'bpl_status'"><span class="kv-k">BPL Status</span><span class="kv-v">{{ getBplStatusLabel(selectedHouse) }}</span></div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import * as Cesium from 'cesium'
import 'cesium/Build/Cesium/Widgets/widgets.css'
import { getLocationOptions } from '../../api/index.js'
import { getPopulationMapData, getPopulationMapInsights } from './api.js'

Cesium.Ion.defaultAccessToken = ''

const cesiumContainer = ref(null)
const houses = ref([])
const insights = ref(null)
const loadingLiveData = ref(true)
const selectedHouse = ref(null)
const hoveredHouse = ref(null)
const mouseX = ref(0)
const mouseY = ref(0)
const tileStyle = ref('satellite')
const sidebarCollapsed = ref(false)
const cameraHeight = ref(120000)
const colorMode = ref('population_density')
const openDropdown = ref(null)

const districtOptions = ref([])
const talukaOptions = ref([])
const villageOptions = ref([])

const filterDistrict = ref('')
const filterTaluka = ref('')
const filterVillage = ref('')
const pendingDistrict = ref('')
const pendingTaluka = ref('')
const pendingVillage = ref('')

let viewer = null
const entityMap = new Map()

const THRESHOLD_BUILDINGS = 3500

const COLOR_MODE_LABELS = {
  population_density: 'Population Density',
  bpl_status: 'BPL Status',
  education_status: 'Education Status',
  employment_status: 'Employment Status',
  divyang_presence: 'Divyang Presence',
}

const legendTitle = computed(() => COLOR_MODE_LABELS[colorMode.value] || 'Legend')

const currentLegend = computed(() => {
  if (colorMode.value === 'population_density') {
    return [
      { color: '#16a34a', label: '1-2 members' },
      { color: '#f59e0b', label: '3-5 members' },
      { color: '#ef4444', label: '6+ members' },
    ]
  }
  if (colorMode.value === 'bpl_status') {
    return [
      { color: '#ef4444', label: 'BPL household' },
      { color: '#16a34a', label: 'Non-BPL household' },
    ]
  }
  if (colorMode.value === 'education_status') {
    return [
      { color: '#16a34a', label: 'Literate dominant' },
      { color: '#f59e0b', label: 'Needs literacy support' },
    ]
  }
  if (colorMode.value === 'employment_status') {
    return [
      { color: '#16a34a', label: 'Working member present' },
      { color: '#f59e0b', label: 'No earning member' },
    ]
  }
  return [
    { color: '#7b1fa2', label: 'Disability present' },
    { color: '#16a34a', label: 'No disability' },
  ]
})

const totalMembers = computed(() => houses.value.reduce((sum, h) => sum + Number(h.total_members || 0), 0))
const maleMembers = computed(() => houses.value.reduce((sum, h) => sum + Number(h.male_members || 0), 0))
const femaleMembers = computed(() => houses.value.reduce((sum, h) => sum + Number(h.female_members || 0), 0))
const malePct = computed(() => (totalMembers.value ? Math.round((maleMembers.value / totalMembers.value) * 100) : 0))
const femalePct = computed(() => (totalMembers.value ? Math.round((femaleMembers.value / totalMembers.value) * 100) : 0))
const workingHouseholds = computed(() => houses.value.filter((h) => Number(h.working_members || 0) >= 1).length)
const dependentHouseholds = computed(() => Math.max(houses.value.length - workingHouseholds.value, 0))
const divyangHouseholds = computed(() => houses.value.filter((h) => Number(h.has_disability || 0) === 1).length)
const literacyRate = computed(() => {
  const edu = insights.value?.education_status || {}
  const literate = Number(edu.literate || 0)
  const illiterate = Number(edu.illiterate || 0)
  const total = literate + illiterate
  return total ? Math.round((literate / total) * 100) : 0
})

const filtersDirty = computed(() => pendingDistrict.value !== filterDistrict.value || pendingTaluka.value !== filterTaluka.value || pendingVillage.value !== filterVillage.value)
const selectedDistrictLabel = computed(() => districtOptions.value.find((d) => String(d.id) === String(pendingDistrict.value))?.name || 'All Districts')
const selectedTalukaLabel = computed(() => talukaOptions.value.find((t) => String(t.id) === String(pendingTaluka.value))?.name || 'All Talukas')
const selectedVillageLabel = computed(() => villageOptions.value.find((v) => String(v.id) === String(pendingVillage.value))?.name || 'All Villages')
const selectedColorModeLabel = computed(() => COLOR_MODE_LABELS[colorMode.value])
const zoomLabel = computed(() => {
  if (cameraHeight.value < THRESHOLD_BUILDINGS) return '3D buildings visible'
  if (cameraHeight.value < 15000) return 'Zoom in to see buildings'
  return null
})

function toggleDropdown(name) {
  openDropdown.value = openDropdown.value === name ? null : name
}

function closeDropdowns() {
  openDropdown.value = null
}

function selectDistrict(id) {
  pendingDistrict.value = id
  pendingTaluka.value = ''
  pendingVillage.value = ''
  closeDropdowns()
  loadLocationOptions()
}

function selectTaluka(id) {
  pendingTaluka.value = id
  pendingVillage.value = ''
  closeDropdowns()
  loadLocationOptions()
}

function selectVillage(id) {
  pendingVillage.value = id
  closeDropdowns()
}

function selectColorMode(mode) {
  colorMode.value = mode
  closeDropdowns()
}

function applyFilters() {
  filterDistrict.value = pendingDistrict.value
  filterTaluka.value = pendingTaluka.value
  filterVillage.value = pendingVillage.value
  loadTwinData()
}

function resetFilters() {
  pendingDistrict.value = ''
  pendingTaluka.value = ''
  pendingVillage.value = ''
  filterDistrict.value = ''
  filterTaluka.value = ''
  filterVillage.value = ''
  loadLocationOptions()
  loadTwinData()
}

function buildImageryProvider(style) {
  const selected = style || tileStyle.value
  if (selected === 'street') {
    return new Cesium.UrlTemplateImageryProvider({
      url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Street_Map/MapServer/tile/{z}/{y}/{x}',
      credit: 'Tiles © Esri',
      maximumLevel: 19,
    })
  }
  return new Cesium.UrlTemplateImageryProvider({
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
    credit: 'Tiles © Esri',
    maximumLevel: 19,
  })
}

function toggleTile() {
  tileStyle.value = tileStyle.value === 'satellite' ? 'street' : 'satellite'
  if (!viewer) return
  viewer.imageryLayers.removeAll()
  viewer.imageryLayers.addImageryProvider(buildImageryProvider())
}

function normalizeText(value) {
  return String(value ?? '').trim().toLowerCase()
}

function getBplStatusLabel(house) {
  const category = normalizeText(house.FAMILY_BELONG_BPL_CATEGORY)
  if (category.includes('non-bpl') || category === 'no' || category === 'apl') return 'Non-BPL'
  if (category.includes('bpl') || category === 'yes') return 'BPL'
  return 'Non-BPL'
}

function getConditionColor(house) {
  const members = Number(house.total_members || 0)

  if (colorMode.value === 'population_density') {
    if (members <= 2) return '#16a34a'
    if (members <= 5) return '#f59e0b'
    return '#ef4444'
  }

  if (colorMode.value === 'bpl_status') {
    return getBplStatusLabel(house) === 'BPL' ? '#ef4444' : '#16a34a'
  }

  if (colorMode.value === 'education_status') {
    const literate = Number(house.literate_members || 0)
    const illiterate = Number(house.illiterate_members || 0)
    return literate > illiterate ? '#16a34a' : '#f59e0b'
  }

  if (colorMode.value === 'employment_status') {
    return Number(house.working_members || 0) >= 1 ? '#16a34a' : '#f59e0b'
  }

  return Number(house.has_disability || 0) === 1 ? '#7b1fa2' : '#16a34a'
}

function buildingHeight(house) {
  return Math.max(Number(house.total_members || 0) * 2, 4)
}

function updateZoomVisibility() {
  if (!viewer || viewer.isDestroyed()) return
  const pos = viewer.camera.positionCartographic
  if (!pos) return
  cameraHeight.value = Math.round(pos.height)
}

function flyToPoints(list) {
  if (!viewer || !list.length) return
  const pts = list
    .filter((h) => Number.isFinite(Number(h.lng)) && Number.isFinite(Number(h.lat)))
    .map((h) => Cesium.Cartesian3.fromDegrees(Number(h.lng), Number(h.lat), 0))

  if (!pts.length) return

  const sphere = Cesium.BoundingSphere.fromPoints(pts)
  const range = Math.max(sphere.radius * 2.2, 300)
  viewer.camera.flyToBoundingSphere(sphere, {
    duration: 1.8,
    offset: new Cesium.HeadingPitchRange(
      Cesium.Math.toRadians(5),
      Cesium.Math.toRadians(-42),
      range,
    ),
  })
}

function flyToHouse(house) {
  if (!viewer || !house) return
  viewer.camera.flyTo({
    destination: Cesium.Cartesian3.fromDegrees(Number(house.lng), Number(house.lat), 220),
    orientation: { heading: 0, pitch: Cesium.Math.toRadians(-55), roll: 0 },
    duration: 1.2,
  })
}

function buildEntities() {
  if (!viewer) return
  viewer.entities.removeAll()
  entityMap.clear()

  houses.value.forEach((house) => {
    const lat = Number(house.lat)
    const lng = Number(house.lng)
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) return

    const height = buildingHeight(house)
    const color = Cesium.Color.fromCssColorString(getConditionColor(house)).withAlpha(0.98)

    const entity = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, height / 2),
      box: {
        dimensions: new Cesium.Cartesian3(10, 10, height),
        material: color,
        outline: true,
        outlineColor: Cesium.Color.fromCssColorString('#7a6040'),
        outlineWidth: 1.5,
      },
    })

    entityMap.set(entity.id, house)
  })
}

function buildQueryParams() {
  const params = {}
  if (filterDistrict.value) params.district_id = filterDistrict.value
  if (filterTaluka.value) params.taluka_id = filterTaluka.value
  if (filterVillage.value) params.village_id = filterVillage.value
  return params
}

async function loadLocationOptions() {
  const res = await getLocationOptions({ district_id: pendingDistrict.value || undefined, taluka_id: pendingTaluka.value || undefined })
  districtOptions.value = res?.districts || []
  talukaOptions.value = res?.talukas || []
  villageOptions.value = res?.villages || []
}

async function loadTwinData() {
  loadingLiveData.value = true
  try {
    const params = buildQueryParams()
    const [mapData, mapInsights] = await Promise.all([
      getPopulationMapData(params),
      getPopulationMapInsights(params),
    ])
    houses.value = Array.isArray(mapData) ? mapData : []
    insights.value = mapInsights || null

    if (viewer) {
      buildEntities()
      flyToPoints(houses.value)
    }
  } catch (error) {
    houses.value = []
    insights.value = null
    console.error('Population 3D twin load failed:', error)
  } finally {
    loadingLiveData.value = false
  }
}

watch(colorMode, () => {
  if (viewer) buildEntities()
})

onMounted(async () => {
  await loadLocationOptions()

  viewer = new Cesium.Viewer(cesiumContainer.value, {
    animation: false,
    timeline: false,
    sceneModePicker: false,
    baseLayerPicker: false,
    geocoder: false,
    homeButton: false,
    navigationHelpButton: false,
    fullscreenButton: false,
    selectionIndicator: false,
    infoBox: false,
    shouldAnimate: false,
    imageryProvider: buildImageryProvider(),
  })

  viewer.scene.globe.enableLighting = true
  viewer.screenSpaceEventHandler.removeInputAction(Cesium.ScreenSpaceEventType.LEFT_DOUBLE_CLICK)

  viewer.screenSpaceEventHandler.setInputAction((event) => {
    const picked = viewer.scene.pick(event.position)
    if (!picked?.id) return
    const house = entityMap.get(picked.id.id || picked.id)
    if (house) selectedHouse.value = house
  }, Cesium.ScreenSpaceEventType.LEFT_CLICK)

  viewer.screenSpaceEventHandler.setInputAction((event) => {
    const picked = viewer.scene.pick(event.endPosition)
    const house = picked?.id ? entityMap.get(picked.id.id || picked.id) : null
    hoveredHouse.value = house || null
    if (house) {
      mouseX.value = event.endPosition.x + 18
      mouseY.value = event.endPosition.y + 16
    }
  }, Cesium.ScreenSpaceEventType.MOUSE_MOVE)

  viewer.screenSpaceEventHandler.setInputAction((event) => {
    const picked = viewer.scene.pick(event.position)
    if (!picked?.id) return
    const house = entityMap.get(picked.id.id || picked.id)
    if (house) flyToHouse(house)
  }, Cesium.ScreenSpaceEventType.LEFT_DOUBLE_CLICK)

  viewer.camera.percentageChanged = 0.03
  viewer.camera.changed.addEventListener(updateZoomVisibility)

  await loadTwinData()
})

onUnmounted(() => {
  if (viewer && !viewer.isDestroyed()) {
    viewer.destroy()
    viewer = null
  }
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
.legend-item   { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.32rem; }
.legend-swatch { width: 12px; height: 12px; border-radius: 3px; flex-shrink: 0; box-shadow: 0 1px 2px rgba(0,0,0,0.15); }
.legend-text   { font-size: 0.7rem; color: #111827; font-weight: 500; }   /* dark, readable */
.legend-note   { font-size: 0.62rem; color: #9ca3af; margin-top: 0.45rem; font-style: italic; }

/* ── Field Issues ── */
.issue-row {
  display: flex; align-items: center; gap: 0.45rem;
  padding: 0.32rem 0.35rem; border-radius: var(--radius-sm);
  cursor: pointer; transition: background 0.12s; margin-bottom: 0.35rem;
  border: 1px solid transparent;
}
.issue-row:hover  { background: #f9fafb; border-color: #e5e7eb; }
.issue-row.active { background: #f0fdf4; border-color: #bbf7d0; }
.issue-pip   { width: 9px; height: 9px; border-radius: 50%; flex-shrink: 0; box-shadow: 0 1px 2px rgba(0,0,0,0.2); }
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
   HOVER CARD
═══════════════════════════════════════════════ */
.hover-card {
  position: fixed; z-index: 300;
  background: #ffffff;
  border: 1.5px solid #d1d5db;
  border-radius: var(--radius);
  padding: 0.6rem 0.85rem;
  box-shadow: 0 8px 24px rgba(0,0,0,0.14), 0 2px 6px rgba(0,0,0,0.08);
  pointer-events: none;
  min-width: 170px;
}
.hover-name { font-size: 0.82rem; font-weight: 700; color: #111827; margin-bottom: 0.35rem; }
.hover-row  { display: flex; justify-content: space-between; gap: 0.8rem; margin-bottom: 0.12rem; }
.hr-key     { font-size: 0.66rem; color: #6b7280; }
.hr-val     { font-size: 0.66rem; color: #111827; font-weight: 600; }
.hover-hint { font-size: 0.61rem; color: #9ca3af; margin-top: 0.38rem; font-style: italic; border-top: 1px solid #e5e7eb; padding-top: 0.32rem; }

/* ═══════════════════════════════════════════════
   DETAIL PANEL (right slide-in)
═══════════════════════════════════════════════ */
.detail-panel {
  position: absolute;
  right: 0.75rem; top: 60px;
  width: 288px;
  max-height: calc(100vh - 72px);
  overflow-y: auto;
  z-index: 100;
  background: #ffffff;
  border: 1.5px solid #d1d5db;
  border-radius: var(--radius);
  box-shadow: 0 8px 28px rgba(0,0,0,0.14), 0 3px 8px rgba(0,0,0,0.07);
  scrollbar-width: thin;
  scrollbar-color: #d1d5db transparent;
}

.detail-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: 0.5rem; padding: 0.9rem 0.9rem 0.65rem;
  border-bottom: 1.5px solid #e5e7eb;
  background: #f9fafb;
  border-radius: var(--radius) var(--radius) 0 0;
}
.detail-badge {
  display: inline-block; padding: 0.2rem 0.6rem;
  border-radius: 20px; border: 1.5px solid;
  font-size: 0.66rem; font-weight: 700;
  margin-bottom: 0.32rem;
}
.detail-name  { font-size: 0.97rem; font-weight: 800; color: #111827; line-height: 1.2; }
.detail-sub   { font-size: 0.66rem; color: #6b7280; margin-top: 0.22rem; font-weight: 500; }
.detail-close {
  background: #f3f4f6; border: 1px solid #e5e7eb;
  border-radius: 50%; color: #6b7280;
  font-size: 1.1rem; line-height: 1; cursor: pointer;
  width: 24px; height: 24px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; transition: all 0.15s;
}
.detail-close:hover { background: #ef4444; border-color: #ef4444; color: #fff; }

.focus-btn {
  display: block; width: calc(100% - 1.8rem);
  margin: 0.65rem 0.9rem 0;
  background: #f0fdf4; border: 1.5px solid #86efac;
  border-radius: var(--radius-sm);
  color: #15803d; font-size: 0.73rem; font-weight: 600;
  padding: 0.38rem 0.6rem; cursor: pointer; text-align: center;
  transition: all 0.15s; box-shadow: 0 1px 3px rgba(22,163,74,0.15);
}
.focus-btn:hover { background: #16a34a; border-color: #15803d; color: #fff; }

.detail-section {
  font-size: 0.59rem; text-transform: uppercase; letter-spacing: 0.1em;
  color: #374151; font-weight: 800;            /* was text3 = invisible */
  padding: 0.75rem 0.9rem 0.32rem;
  border-top: 1px solid #f3f4f6;
}

.kv-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.38rem; padding: 0 0.9rem; }
.kv {
  background: #f9fafb; border: 1.5px solid #e5e7eb;
  border-radius: var(--radius-sm); padding: 0.38rem 0.52rem;
}
.kv-k { font-size: 0.57rem; text-transform: uppercase; letter-spacing: 0.06em; color: #6b7280; display: block; font-weight: 600; }
.kv-v { font-size: 0.76rem; color: #111827; font-weight: 600; margin-top: 0.1rem; display: block; }

/* Advisory cards */
.detail-issues { padding: 0 0.9rem 0.9rem; }
.advisory-card {
  border-left: 3px solid; border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  background: #f9fafb;
  border-top: 1px solid #e5e7eb;
  border-right: 1px solid #e5e7eb;
  border-bottom: 1px solid #e5e7eb;
  padding: 0.6rem 0.7rem;
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
  background: #f0fdf4; margin: 0.5rem 0.9rem 0.9rem;
  border-radius: var(--radius-sm); border: 1px solid #bbf7d0;
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
