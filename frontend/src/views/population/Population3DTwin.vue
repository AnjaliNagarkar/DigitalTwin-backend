<template>
  <div class="twin-page">
    <div class="cesium-wrap" ref="cesiumContainer">
      <button
        class="map-fs-btn"
        :class="{ shifted: selectedHouse || selectedCluster }"
        @click="toggleTwinFullscreen"
        :title="isTwinFullscreen ? t('twin.exitFullscreen') : t('twin.fullscreen')"
        aria-label="Toggle fullscreen"
      >
        {{ isTwinFullscreen ? '⤡' : '⤢' }}
      </button>

    </div>

    <div class="topbar">
      <div class="topbar-brand">
        <span class="brand-dot"></span>
        <span class="brand-name">PopTwin</span>
        <span class="brand-sub">3D Digital Twin</span>
      </div>

      <div class="filter-bar">
        <div class="filter-group">
          <label class="filter-label">{{ t('map.district') }}</label>
          <div class="custom-select" :class="{ open: openDropdown === 'district' }" @click.stop="toggleDropdown('district')">
            <button class="cs-trigger" type="button">
              <span class="cs-value">{{ selectedDistrictLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'district'" @click.stop>
              <div class="cs-option" :class="{ selected: !pendingDistrict }" @click="selectDistrict('')">{{ t('map.allDistricts') }}</div>
              <div class="cs-option" v-for="d in districtOptions" :key="d.id" :class="{ selected: String(pendingDistrict) === String(d.id) }" @click="selectDistrict(d.id)">{{ d.name }}</div>
            </div>
          </div>
        </div>

        <span class="filter-arrow">›</span>

        <div class="filter-group">
          <label class="filter-label">{{ t('map.taluka') }}</label>
          <div class="custom-select" :class="{ open: openDropdown === 'taluka', disabled: !pendingDistrict }" @click.stop="pendingDistrict && toggleDropdown('taluka')">
            <button class="cs-trigger" type="button" :disabled="!pendingDistrict">
              <span class="cs-value">{{ selectedTalukaLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'taluka'" @click.stop>
              <div class="cs-option" :class="{ selected: !pendingTaluka }" @click="selectTaluka('')">{{ t('map.allTalukas') }}</div>
              <div class="cs-option" v-for="t in talukaOptions" :key="t.id" :class="{ selected: String(pendingTaluka) === String(t.id) }" @click="selectTaluka(t.id)">{{ t.name }}</div>
            </div>
          </div>
        </div>

        <span class="filter-arrow">›</span>

        <div class="filter-group">
          <label class="filter-label">{{ t('map.village') }}</label>
          <div class="custom-select" :class="{ open: openDropdown === 'village', disabled: !pendingTaluka }" @click.stop="pendingTaluka && toggleDropdown('village')">
            <button class="cs-trigger" type="button" :disabled="!pendingTaluka">
              <span class="cs-value">{{ selectedVillageLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'village'" @click.stop>
              <div class="cs-option" :class="{ selected: !pendingVillage }" @click="selectVillage('')">{{ t('map.allVillages') }}</div>
              <div class="cs-option" v-for="v in villageOptions" :key="v.id" :class="{ selected: String(pendingVillage) === String(v.id) }" @click="selectVillage(v.id)">{{ v.name }}</div>
            </div>
          </div>
        </div>

        <button class="apply-btn" @click="applyFilters" :disabled="!filtersDirty">{{ t('common.apply') }}</button>
        <button class="reset-btn" @click="resetFilters" v-if="pendingDistrict || pendingTaluka || pendingVillage || filterDistrict || filterTaluka || filterVillage">✕ {{ t('common.reset') }}</button>
      </div>

      <div class="topbar-right">
        <div class="ctrl-group">
          <label class="filter-label">{{ t('popMap.colorBy') }}</label>
          <div class="custom-select cs-align-right" :class="{ open: openDropdown === 'colorMode' }" @click.stop="toggleDropdown('colorMode')">
            <button class="cs-trigger" type="button">
              <span class="cs-value">{{ selectedColorModeLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown cs-dropdown-right" v-show="openDropdown === 'colorMode'" @click.stop>
              <div class="cs-option" :class="{ selected: colorMode === 'population_density' }" @click="selectColorMode('population_density')">{{ t('viewBy.populationDensity') }}</div>
              <div class="cs-option" :class="{ selected: colorMode === 'bpl_status' }" @click="selectColorMode('bpl_status')">{{ t('viewBy.bplStatus') }}</div>
              <div class="cs-option" :class="{ selected: colorMode === 'divyang_presence' }" @click="selectColorMode('divyang_presence')">{{ t('viewBy.divyangPresence') }}</div>
              <div class="cs-option" :class="{ selected: colorMode === 'employment_status' }" @click="selectColorMode('employment_status')">{{ t('viewBy.employmentStatus') }}</div>
              <div class="cs-option-group-label">— {{ t('viewBy.documentGapAnalysis') }} —</div>
              <div class="cs-option" :class="{ selected: colorMode === 'aadhaar_coverage' }" @click="selectColorMode('aadhaar_coverage')">{{ t('viewBy.aadhaarCoverage') }}</div>
              <div class="cs-option" :class="{ selected: colorMode === 'caste_certificate_coverage' }" @click="selectColorMode('caste_certificate_coverage')">{{ t('viewBy.casteCertificate') }}</div>
            </div>
          </div>
        </div>

        <button class="ctrl-btn" :class="{ active: tileStyle === 'satellite' }" @click="toggleTile">
          {{ tileStyle === 'satellite' ? `🛰 ${t('twin.satellite')}` : `🗺 ${t('twin.street')}` }}
        </button>

        <div class="dl-wrap" v-if="!loadingLiveData && houses.length">
          <button class="dl-btn" type="button" :disabled="pdfLoading" @click="downloadPDF" :title="`${t('twin.downloadPdfFor')} ${houses.length} ${t('map.households')}`">{{ pdfLoading ? `⏳ ${t('twin.generating')}` : `⬇ ${t('twin.pdfReport')}` }}</button>
          <span class="dl-count">{{ houses.length.toLocaleString() }} {{ t('twin.rows') }}</span>
        </div>
      </div>
    </div>

    <div class="loading-overlay" v-if="loadingLiveData">
      <div class="loading-spinner"></div>
      <div class="loading-text">{{ t('twin.loadingPopTwin') }}</div>
    </div>

    <div class="stats-bar" v-if="!loadingLiveData">
      <span class="stat-item">
        <span class="stat-dot" style="background:#16a34a"></span>
        <strong>{{ houses.length.toLocaleString() }}</strong> {{ t('map.households') }}
      </span>
      <span class="stat-sep">·</span>
      <span class="stat-item"><strong>{{ totalMembers.toLocaleString() }}</strong> {{ t('common.population') }}</span>
      <span class="stat-sep">·</span>
      <span class="stat-item">Maharashtra</span>
      <span class="stat-sep" v-if="zoomLabel">·</span>
      <span class="stat-item zoom-label" v-if="zoomLabel">{{ zoomLabel }}</span>
    </div>

    <div class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <button class="sidebar-toggle" @click="sidebarCollapsed = !sidebarCollapsed" :title="sidebarCollapsed ? t('twin.openPanel') : t('twin.closePanel')">
        {{ sidebarCollapsed ? '›' : '‹' }}
      </button>

      <div class="sidebar-body" v-if="!sidebarCollapsed && !loadingLiveData">
        <div class="panel-card">
          <div class="card-title">{{ legendTitle }}</div>
          <div class="legend-item" v-for="leg in currentLegend" :key="leg.label">
            <span class="mini-house" :style="{ '--mh-roof': leg.color }">
              <span class="mh-roof"></span>
              <span class="mh-wall"></span>
            </span>
            <span class="legend-text">{{ leg.label }}</span>
          </div>
          <div class="legend-note">{{ legendNote }}</div>
        </div>

        <div class="panel-card">
          <div class="card-title">{{ t('twin.problemFilter') }}
            <span class="card-title-sub">{{ t('twin.highlightOnMap') }}</span>
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
            <button class="pf-clear-btn" @click="activeProblemFilters = []">✕ {{ t('common.reset') }}</button>
          </div>
          <div class="pf-hint" v-else>
            {{ t('twin.problemFilterHint') }}
          </div>
        </div>

        <div class="panel-card">
          <div class="card-title">{{ t('twin.populationOverview') }}</div>
          <div class="issue-row">
            <span class="issue-pip" style="background:#14b8a6"></span>
            <div class="issue-body">
              <div class="issue-top"><span class="issue-name">{{ t('twin.totalHouseholds') }}</span><span class="issue-count">{{ houses.length.toLocaleString() }}</span></div>
              <div class="issue-track"><div class="issue-fill" style="width:100%;background:#14b8a6"></div></div>
            </div>
          </div>
          <div class="issue-row">
            <span class="issue-pip" style="background:#2563eb"></span>
            <div class="issue-body">
              <div class="issue-top"><span class="issue-name">{{ t('twin.totalPopulation') }}</span><span class="issue-count">{{ totalMembers.toLocaleString() }}</span></div>
              <div class="issue-track"><div class="issue-fill" style="width:100%;background:#2563eb"></div></div>
            </div>
          </div>
        </div>

        <div class="panel-card">
          <div class="card-title">{{ t('twin.genderRatio') }}</div>
          <div class="legend-item"><span class="legend-swatch" style="background:#2563eb"></span><span class="legend-text">{{ t('analytics.male') }} {{ malePct }}%</span></div>
          <div class="legend-item"><span class="legend-swatch" style="background:#ec4899"></span><span class="legend-text">{{ t('analytics.female') }} {{ femalePct }}%</span></div>
        </div>

        <div class="panel-card">
          <div class="card-title">{{ t('twin.employmentSummary') }}</div>
          <div class="legend-item"><span class="legend-swatch" style="background:#16a34a"></span><span class="legend-text">{{ t('twin.workingHouseholds') }} {{ workingHouseholds.toLocaleString() }}</span></div>
          <div class="legend-item"><span class="legend-swatch" style="background:#f59e0b"></span><span class="legend-text">{{ t('twin.dependentHouseholds') }} {{ dependentHouseholds.toLocaleString() }}</span></div>
        </div>

        <div class="panel-card">
          <div class="card-title">{{ t('twin.educationSummary') }}</div>
          <div class="legend-item"><span class="legend-swatch" style="background:#16a34a"></span><span class="legend-text">{{ t('twin.literacyRate') }} {{ literacyRate }}%</span></div>
        </div>

        <div class="panel-card">
          <div class="card-title">{{ t('twin.divyangSummary') }}</div>
          <div class="legend-item"><span class="legend-swatch" style="background:#7b1fa2"></span><span class="legend-text">{{ t('twin.householdsWithDisability') }} {{ divyangHouseholds.toLocaleString() }}</span></div>
        </div>
      </div>
    </div>

    <div v-if="hoveredHouse" class="hover-card" :style="{ left: mouseX + 'px', top: mouseY + 'px' }">
      <div class="hover-name">{{ hoveredHouse.head_name || t('map.households') }}</div>
      <div class="hover-row"><span class="hr-key">{{ t('mapView.houseNo') }}</span><span class="hr-val">{{ hoveredHouse.house_no || '—' }}</span></div>
      <div class="hover-row"><span class="hr-key">{{ t('map.totalMembers') }}</span><span class="hr-val">{{ Number(hoveredHouse.total_members || 0) }}</span></div>
      <div class="hover-hint">{{ t('twin.hoverHint') }}</div>
    </div>

    <transition name="slide">
      <div v-if="selectedHouse" class="detail-panel">
        <div class="detail-header">
          <div>
            <div class="detail-badge" :style="{ background: getConditionColor(selectedHouse) + '18', borderColor: getConditionColor(selectedHouse) + '60', color: getConditionColor(selectedHouse) }">
              {{ selectedColorModeLabel }}
            </div>
            <div class="detail-name">{{ selectedHouse.head_name || t('map.households') }}</div>
            <div class="detail-sub">{{ t('mapView.houseNo') }} {{ selectedHouse.house_no || 'N/A' }}</div>
          </div>
          <button class="detail-close" @click="selectedHouse = null">×</button>
        </div>

        <button class="focus-btn" @click="flyToHouse(selectedHouse)">📍 {{ t('twin.zoomToHouse') }}</button>

        <!-- Context-Aware Field Filter Dropdown -->
        <div class="filter-section">
          <label class="filter-label">{{ t('twin.viewFields') }}</label>
          <div class="custom-select" :class="{ open: openDropdown === 'drawerFilter' }" @click.stop="openDropdown = openDropdown === 'drawerFilter' ? null : 'drawerFilter'">
            <button class="cs-trigger" type="button">
              <span class="cs-value">{{ DRAWER_FILTER_OPTIONS.find(f => f.value === activeDrawerFilter)?.label || t('twin.allFields') }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'drawerFilter'" @click.stop>
              <div v-for="filter in DRAWER_FILTER_OPTIONS" :key="filter.value"
                   class="cs-option" :class="{ selected: activeDrawerFilter === filter.value }"
                   @click="activeDrawerFilter = filter.value; openDropdown = null">
                {{ filter.label }}
              </div>
            </div>
          </div>
        </div>

        <!-- Dynamic Field Sections Based on Active Filter -->
        <template v-for="section in displayedSections" :key="section.title">
          <div class="detail-section">{{ section.title }}</div>
          <div class="kv-grid">
            <div v-for="field in section.fields" :key="field.key" class="kv">
              <span class="kv-k">{{ field.label }}</span>
              <span class="kv-v">{{ field.value }}</span>
            </div>
          </div>
        </template>

        <!-- Empty State -->
        <div v-if="!displayedSections.length || displayedSections.every(s => s.fields.length === 0)" class="detail-empty">
          {{ t('twin.noDataForFilter') }}
        </div>
      </div>
    </transition>

    <transition name="slide">
      <div v-if="selectedCluster" class="cluster-panel">
        <button class="detail-close cluster-close" @click="selectedCluster = null">×</button>

        <div class="cluster-header">
          <div class="cluster-badge">⚠ {{ t('twin.highNeedArea') }}</div>
          <div class="cluster-count">
            <strong>{{ selectedCluster.count }}</strong> {{ t('twin.householdsInZone') }}
          </div>
        </div>

        <div class="cluster-section-title" v-if="selectedCluster.problems.length">
          🔍 {{ t('twin.mainIssuesDetected') }}
        </div>

        <div class="cp-card" v-for="p in selectedCluster.problems" :key="p.key">
          <div class="cp-top">
            <span class="cp-emoji">{{ p.emoji }}</span>
            <div class="cp-info">
              <span class="cp-label">{{ p.label }}</span>
              <span class="cp-stat">{{ p.count }} of {{ selectedCluster.count }} households ({{ p.pct }}%)</span>
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
          ✅ {{ t('twin.noIssuesDetected') }}
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import * as Cesium from 'cesium'
import 'cesium/Build/Cesium/Widgets/widgets.css'
import { getLocationOptions } from '../../api/index.js'
import { getPopulationMapData, getPopulationMapInsights } from './api.js'

Cesium.Ion.defaultAccessToken = ''

const { t } = useI18n()

const cesiumContainer = ref(null)
const houses = ref([])
const insights = ref(null)
const loadingLiveData = ref(true)
const selectedHouse = ref(null)
const previouslySelectedHouseNo = ref(null)
const pdfLoading = ref(false)
const tileStyle = ref('street')
const sidebarCollapsed = ref(false)
const isTwinFullscreen = ref(false)
const cameraHeight = ref(120000)
const selectedColorBy = ref('population_density')
const colorMode = selectedColorBy
const openDropdown = ref(null)
const activeProblemFilters = ref([])
const selectedCluster = ref(null)
const activeDrawerFilter = ref('')

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
const buildingIds = new Set()
const pointIds = new Set()
const clusterIds = new Set()
const clusterMap = new Map()
const clusterEntityGroups = []
const clusterIconCache = new Map()
const houseClusterMap = new Map()
const clusterEntities = []
const houseEntities = []
const clusterLabels = []

function handleTwinFullscreenChange() {
  isTwinFullscreen.value = !!document.fullscreenElement
}


const THRESHOLD_BUILDINGS = 3500
const MIN_PIXEL_DISTANCE = 40
const SHOW_HOUSES_ZOOM = 14

const COLOR_MODE_LABELS = {
  population_density: 'Population Density',
  bpl_status: 'BPL Status',
  education_status: 'Education Status',
  employment_status: 'Employment Status',
  divyang_presence: 'Divyang Presence',
  aadhaar_coverage: 'Aadhaar Coverage',
  caste_certificate_coverage: 'Caste Certificate Coverage',
}

/**
 * ════════════════════════════════════════════════════════════════════════════════
 * CONTEXT-AWARE DRAWER SYSTEM
 * ════════════════════════════════════════════════════════════════════════════════
 * Provides modular field filtering for detail panel based on active filter.
 * Maps each filter (BPL, Student, Divyang) to relevant household data fields.
 * ════════════════════════════════════════════════════════════════════════════════
 */

const DRAWER_FILTER_OPTIONS = [
  { label: 'All Fields', value: '' },
  { label: 'BPL & Welfare', value: 'bpl' },
  { label: 'Education & Students', value: 'student' },
  { label: 'Disability & Support', value: 'divyang' },
]

const FIELD_LABELS = {
  total_members: 'Total Members',
  male_members: 'Male',
  female_members: 'Female',
  working_members: 'Working Members',
  unemployed_members: 'Unemployed Members',
  literate_members: 'Literate Members',
  illiterate_members: 'Illiterate Members',
  divyang_members: 'Divyang Members',
  has_disability: 'Disability Status',
  working_occupations: 'Occupations',
  FAMILY_BELONG_BPL_CATEGORY: 'BPL Category',
  RATION_CARD_TYPE: 'Ration Card Type',
  ANNUAL_INCOME: 'Annual Income',
  aadhaarCoverageStatus: 'Aadhaar Coverage',
  casteCertificateCoverageStatus: 'Caste Certificate Coverage',
}

const FIELD_MAPPINGS = {
  '': {
    sections: [
      {
        title: 'Population',
        fields: ['total_members', 'male_members', 'female_members'],
      },
      {
        title: 'Employment & Occupation',
        fields: ['working_members', 'unemployed_members', 'working_occupations'],
      },
      {
        title: 'Education & Literacy',
        fields: ['literate_members', 'illiterate_members'],
      },
      {
        title: 'Welfare & Income',
        fields: ['FAMILY_BELONG_BPL_CATEGORY', 'RATION_CARD_TYPE', 'ANNUAL_INCOME'],
      },
      {
        title: 'Documents & Coverage',
        fields: ['aadhaarCoverageStatus', 'casteCertificateCoverageStatus'],
      },
      {
        title: 'Disability & Health',
        fields: ['divyang_members', 'has_disability'],
      },
    ],
  },
  bpl: {
    sections: [
      {
        title: 'BPL Status & Welfare',
        fields: ['FAMILY_BELONG_BPL_CATEGORY', 'RATION_CARD_TYPE'],
      },
      {
        title: 'Economic Indicators',
        fields: ['ANNUAL_INCOME', 'working_members', 'unemployed_members'],
      },
      {
        title: 'Family Composition',
        fields: ['total_members', 'male_members', 'female_members'],
      },
    ],
  },
  student: {
    sections: [
      {
        title: 'Education Status',
        fields: ['literate_members', 'illiterate_members', 'total_members'],
      },
      {
        title: 'Employment Context',
        fields: ['working_members', 'unemployed_members', 'working_occupations'],
      },
    ],
  },
  divyang: {
    sections: [
      {
        title: 'Disability Information',
        fields: ['divyang_members', 'has_disability'],
      },
      {
        title: 'Family Context',
        fields: ['total_members', 'male_members', 'female_members'],
      },
      {
        title: 'Economic Support Eligibility',
        fields: ['FAMILY_BELONG_BPL_CATEGORY', 'ANNUAL_INCOME'],
      },
    ],
  },
}

const legendTitle = computed(() => COLOR_MODE_LABELS[colorMode.value] || 'Legend')

const legendNote = computed(() => {
  if (colorMode.value === 'population_density') {
    return 'Building roof color = family size'
  }
  return `Roof color = ${legendTitle.value.toLowerCase()} status`
})

const currentLegend = computed(() => {
  if (colorMode.value === 'population_density') {
    return [
      { color: '#22c55e', label: '1-2 members' },
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
  if (colorMode.value === 'aadhaar_coverage') {
    return [
      { color: '#2563eb', label: 'Complete — all members covered' },
      { color: '#f59e0b', label: 'Partial — some members missing' },
      { color: '#dc2626', label: 'Missing — no Aadhaar recorded' },
      { color: '#9ca3af', label: 'Unknown' },
    ]
  }
  if (colorMode.value === 'caste_certificate_coverage') {
    return [
      { color: '#2563eb', label: 'Complete — all members covered' },
      { color: '#f59e0b', label: 'Partial — some members missing' },
      { color: '#dc2626', label: 'Missing — no certificate recorded' },
      { color: '#9ca3af', label: 'Unknown' },
    ]
  }
  return [
    { color: '#7b1fa2', label: 'Disability present' },
    { color: '#16a34a', label: 'No disability' },
  ]
})

const PROBLEM_FILTER_META = [
  { key: 'bplFamilies', label: 'BPL Families', color: '#60a5fa' },
  { key: 'illiterateMembers', label: 'Illiterate Members', color: '#f59e0b' },
  { key: 'unemployedMembers', label: 'Unemployed Members', color: '#ef4444' },
  { key: 'divyangMembers', label: 'Divyang Members', color: '#7b1fa2' },
  { key: 'missingAadhaar', label: 'Missing Aadhaar', color: '#dc2626' },
  { key: 'missingCasteCertificate', label: 'Missing Caste Certificate', color: '#b91c1c' },
]

function matchesProblemFilter(house, key) {
  if (key === 'bplFamilies') {
    return normalizeText(house.FAMILY_BELONG_BPL_CATEGORY) === 'yes'
  }
  if (key === 'illiterateMembers') {
    return Number(house.illiterate_members || 0) > 0
  }
  if (key === 'unemployedMembers') {
    return Number(house.unemployed_members || 0) > 0
  }
  if (key === 'divyangMembers') {
    return Number(house.divyang_members || 0) > 0 || Number(house.has_disability || 0) === 1
  }
  if (key === 'missingAadhaar') {
    return getAadhaarCoverageStatus(house) === 'missing' || getAadhaarCoverageStatus(house) === 'partial'
  }
  if (key === 'missingCasteCertificate') {
    return getCasteCertificateCoverageStatus(house) === 'missing' || getCasteCertificateCoverageStatus(house) === 'partial'
  }
  return false
}

function matchesAllProblems(house) {
  return activeProblemFilters.value.every((key) => matchesProblemFilter(house, key))
}

const problemFilterStats = computed(() => ({
  bplFamilies: houses.value.filter((h) => matchesProblemFilter(h, 'bplFamilies')).length,
  illiterateMembers: houses.value.filter((h) => matchesProblemFilter(h, 'illiterateMembers')).length,
  unemployedMembers: houses.value.filter((h) => matchesProblemFilter(h, 'unemployedMembers')).length,
  divyangMembers: houses.value.filter((h) => matchesProblemFilter(h, 'divyangMembers')).length,
  missingAadhaar: houses.value.filter((h) => matchesProblemFilter(h, 'missingAadhaar')).length,
  missingCasteCertificate: houses.value.filter((h) => matchesProblemFilter(h, 'missingCasteCertificate')).length,
}))

const problemMatchCount = computed(() => {
  if (!activeProblemFilters.value.length) return 0
  return houses.value.filter(matchesAllProblems).length
})

const CLUSTER_PROBLEM_META = [
  {
    key: 'bplFamilies',
    label: 'BPL Families',
    emoji: '🧾',
    action: 'Ensure food and health entitlement access across this zone',
    solution: 'Ensure access to food security and health support benefits.',
    scheme: 'NFSA Ration Card eligibility · Ayushman Bharat support',
  },
  {
    key: 'illiterateMembers',
    label: 'Illiterate Members',
    emoji: '📚',
    action: 'Run school and Anganwadi enrollment outreach for this cluster',
    solution: 'Encourage enrollment in school or Anganwadi education services.',
    scheme: 'Anganwadi education support',
  },
  {
    key: 'unemployedMembers',
    label: 'Unemployed Members',
    emoji: '🧰',
    action: 'Mobilize e-Shram and SHG registration camp in this area',
    solution: 'Encourage registration on e-Shram portal and participation in Self Help Groups.',
    scheme: 'e-Shram employment support',
  },
  {
    key: 'divyangMembers',
    label: 'Divyang Members',
    emoji: '♿',
    action: 'Verify disability certification and pension enrollment status',
    solution: 'Ensure disability certificate and enrollment in disability pension schemes.',
    scheme: 'Disability Pension Support',
  },
  {
    key: 'missingAadhaar',
    label: 'Missing Aadhaar',
    emoji: '🪪',
    action: 'Organise Aadhaar enrollment camp for households with incomplete coverage',
    solution: 'Facilitate Aadhaar enrollment for members without an Aadhaar card.',
    scheme: 'Aadhaar Enrollment Drive',
  },
  {
    key: 'missingCasteCertificate',
    label: 'Missing Caste Certificate',
    emoji: '📄',
    action: 'Assist households in obtaining caste certificates through local administration',
    solution: 'Guide eligible households to apply for caste certificates at the district office.',
    scheme: 'Caste Certificate Issuance Support',
  },
]

function analyzeCluster(houseList) {
  const total = houseList.length
  const selectedKeys = activeProblemFilters.value.length
    ? activeProblemFilters.value
    : CLUSTER_PROBLEM_META.map((p) => p.key)

  return CLUSTER_PROBLEM_META
    .filter((meta) => selectedKeys.includes(meta.key))
    .map((meta) => ({
      ...meta,
      count: houseList.filter((h) => matchesProblemFilter(h, meta.key)).length,
    }))
    .filter((p) => p.count > 0)
    .sort((a, b) => b.count - a.count)
    .map((p) => ({ ...p, pct: Math.round((p.count / total) * 100) }))
}

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

async function toggleTwinFullscreen() {
  const el = cesiumContainer.value
  if (!el) return

  try {
    if (document.fullscreenElement) {
      await document.exitFullscreen()
      return
    }
    await el.requestFullscreen()
  } catch (e) {
    console.warn('Fullscreen unavailable:', e?.message || e)
  }
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

/**
 * Formats field values for display in the detail drawer.
 * Handles boolean values, null/undefined, and special cases.
 */
function formatFieldValue(field, value) {
  // Handle null/undefined
  if (value === null || value === undefined || value === '') return '—'

  // Handle disability status (boolean as number)
  if (field === 'has_disability') {
    return Number(value) === 1 ? 'Yes' : 'No'
  }

  // Handle calculated field: non-working members
  if (field === 'non_working_members' && selectedHouse.value) {
    const total = Number(selectedHouse.value.total_members || 0)
    const working = Number(selectedHouse.value.working_members || 0)
    return Math.max(0, total - working)
  }

  // Default: return as-is
  return String(value)
}

/**
 * Returns the sections and fields to display based on active drawer filter.
 * Provides a clean API for the template to render filtered field groups.
 */
const displayedSections = computed(() => {
  if (!selectedHouse.value) return []

  const house = selectedHouse.value
  const filter = activeDrawerFilter.value
  const mapping = FIELD_MAPPINGS[filter] || FIELD_MAPPINGS['']

  return mapping.sections.map((section) => ({
    title: section.title,
    fields: section.fields.map((fieldKey) => ({
      key: fieldKey,
      label: FIELD_LABELS[fieldKey] || fieldKey,
      value: formatFieldValue(fieldKey, house[fieldKey]),
    })),
  }))
})

/**
 * ════════════════════════════════════════════════════════════════════════════════
 * COLOR PERSISTENCE & DATA READINESS FIX
 * ════════════════════════════════════════════════════════════════════════════════
 * Problem: Houses were displaying gap indicators (yellow) initially, then changing
 * color after data loaded or on user interaction, because:
 *   1. Colors were calculated before document coverage fields were populated
 *   2. No validation of data completeness before rendering
 *   3. Color changes on every click/watcher due to cache misses
 *
 * Solution:
 *   • isHouseDataReady() - Validates required fields exist before color calculation
 *   • getCachedConditionColor() - Prevents thrashing via stable color cache
 *   • clearColorCache() - Invalidates cache only on mode changes, not on clicks
 *   • Fallback gray color for incomplete document data
 * ════════════════════════════════════════════════════════════════════════════════
 */

function getAadhaarCoverageStatus(house) {
  return String(house?.aadhaarCoverageStatus || '').toLowerCase().trim()
}

function getCasteCertificateCoverageStatus(house) {
  return String(house?.casteCertificateCoverageStatus || '').toLowerCase().trim()
}

function getDocCoverageColor(status) {
  if (status === 'complete') return '#2563eb'
  if (status === 'partial') return '#f59e0b'
  if (status === 'missing') return '#dc2626'
  return '#9ca3af'
}

/**
 * Validates that house data is fully loaded and has required fields for accurate color calculation.
 * Prevents color calculation on incomplete data that would yield misleading gap indicators.
 */
function isHouseDataReady(house) {
  if (!house) return false

  // Basic fields required for any color mode
  if (house.total_members === undefined || house.total_members === null) return false

  // For document gap modes, ensure coverage status fields exist AND have meaningful values
  if (colorMode.value === 'aadhaar_coverage') {
    const status = String(house.aadhaarCoverageStatus || '').toLowerCase().trim()
    // Only consider ready if status is explicitly set to a real value, not empty/default
    return status === 'complete' || status === 'partial' || status === 'missing'
  }

  if (colorMode.value === 'caste_certificate_coverage') {
    const status = String(house.casteCertificateCoverageStatus || '').toLowerCase().trim()
    return status === 'complete' || status === 'partial' || status === 'missing'
  }

  // For employment_status, ensure working_members is present (may be 0)
  if (colorMode.value === 'employment_status') {
    return house.working_members !== undefined && house.working_members !== null
  }

  // For education_status, ensure both fields exist
  if (colorMode.value === 'education_status') {
    return house.literate_members !== undefined && house.illiterate_members !== undefined
  }

  // For BPL status, check for bpl indicator field
  if (colorMode.value === 'bpl_status') {
    return house.bpl_status !== undefined || house.bpl !== undefined || house.bpl_category !== undefined
  }

  // For population_density, must have at least 1 member to assign a meaningful color
  if (colorMode.value === 'population_density') {
    return Number(house.total_members || 0) > 0
  }

  return true
}

/**
 * Cached color calculation to ensure consistency within a single render cycle.
 * Prevents color thrashing when data updates trigger multiple render calls.
 */
const colorCache = new Map()
function getCachedConditionColor(house) {
  if (!house) return '#9ca3af'

  const cacheKey = `${house.house_no || house.id}|${colorMode.value}`
  if (colorCache.has(cacheKey)) {
    return colorCache.get(cacheKey)
  }

  const color = getConditionColor(house)
  colorCache.set(cacheKey, color)
  return color
}

function clearColorCache() {
  colorCache.clear()
}

function getConditionColor(house) {
  const members = Number(house.total_members || 0)

  if (colorMode.value === 'population_density') {
    // Guard: 0 members means data not loaded — show neutral, not green (0 ≤ 2 was firing)
    if (members === 0) return '#9ca3af'
    if (members <= 2) return '#22c55e'
    if (members <= 5) return '#f59e0b'
    return '#ef4444'
  }

  if (colorMode.value === 'bpl_status') {
    return getBplStatusLabel(house) === 'BPL' ? '#ef4444' : '#16a34a'
  }

  if (colorMode.value === 'education_status') {
    const literate   = Number(house.literate_members   ?? -1)
    const illiterate = Number(house.illiterate_members ?? -1)
    // Guard: both -1 (field absent) or both 0 (column may not exist) → no data
    if (literate < 0 || (literate === 0 && illiterate === 0)) return '#9ca3af'
    return literate > illiterate ? '#16a34a' : '#f59e0b'
  }

  if (colorMode.value === 'employment_status') {
    // Guard: field absent means no data — distinguish from genuinely 0 working members
    if (house.working_members === undefined || house.working_members === null) return '#9ca3af'
    return Number(house.working_members) >= 1 ? '#16a34a' : '#f59e0b'
  }

  if (colorMode.value === 'aadhaar_coverage') {
    return getDocCoverageColor(getAadhaarCoverageStatus(house))
  }

  if (colorMode.value === 'caste_certificate_coverage') {
    return getDocCoverageColor(getCasteCertificateCoverageStatus(house))
  }

  return Number(house.has_disability || 0) === 1 ? '#7b1fa2' : '#16a34a'
}

function cesiumColor(house) {
  // Use cached color to prevent recalculation thrashing
  const colorHex = getCachedConditionColor(house)
  const base = Cesium.Color.fromCssColorString(colorHex)
  return new Cesium.Color(base.red * 0.8, base.green * 0.8, base.blue * 0.8, 1.0)
}

function buildingHeight(house) {
  return Math.max(Number(house.total_members || 0) * 2, 4)
}

function getHouseId(house) {
  return house.house_id || house.household_id || house.id || house.house_no || `${house.lat},${house.lng}`
}

function getHouseKey(house) {
  const lat = Number(house.lat)
  const lng = Number(house.lng)
  return `${getHouseId(house)}|${lat.toFixed(6)}|${lng.toFixed(6)}`
}

function normalizeClusterSelection(cluster) {
  if (!cluster) return null

  const issues = cluster.issues || {
    bpl: Number(cluster.problems?.find((item) => item.key === 'bplFamilies')?.count || 0),
    illiterate: Number(cluster.problems?.find((item) => item.key === 'illiterateMembers')?.count || 0),
    unemployed: Number(cluster.problems?.find((item) => item.key === 'unemployedMembers')?.count || 0),
    divyang: Number(cluster.problems?.find((item) => item.key === 'divyangMembers')?.count || 0),
  }

  return {
    ...cluster,
    household_count: Number(cluster.household_count ?? cluster.count ?? 0),
    count: Number(cluster.count ?? cluster.household_count ?? 0),
    issues,
    problems: cluster.problems || analyzeCluster(cluster.households || []),
    households: cluster.households || [],
    lat: Number(cluster.lat),
    lng: Number(cluster.lng),
  }
}

function getClusterCenter(cluster) {
  const points = Array.isArray(cluster?.households) ? cluster.households : []
  if (!points.length) {
    return {
      lat: Number(cluster?.lat || 0),
      lng: Number(cluster?.lng || 0),
    }
  }

  const sums = points.reduce((acc, house) => {
    acc.lat += Number(house.lat || 0)
    acc.lng += Number(house.lng || 0)
    return acc
  }, { lat: 0, lng: 0 })

  return {
    lat: sums.lat / points.length,
    lng: sums.lng / points.length,
  }
}

function updateZoomVisibility() {
  if (!viewer || viewer.isDestroyed()) return
  const pos = viewer.camera.positionCartographic
  if (!pos) return
  const h = pos.height
  cameraHeight.value = Math.round(h)

  const zoom = getZoomLevel(h)
  const hasClusterOverlay = activeProblemFilters.value.length > 0 && clusterEntities.length > 0
  const showClusters = hasClusterOverlay && zoom < SHOW_HOUSES_ZOOM
  const showHouses = !hasClusterOverlay || zoom >= SHOW_HOUSES_ZOOM
  const showBuildings = h < THRESHOLD_BUILDINGS

  houseEntities.forEach((entity) => {
    if (!entity) return
    if (buildingIds.has(entity.id)) {
      entity.show = showHouses && showBuildings
    } else if (pointIds.has(entity.id)) {
      entity.show = showHouses && !showBuildings
    } else {
      entity.show = showHouses
    }
  })

  clusterEntities.forEach((entity) => {
    if (!entity) return
    entity.show = showClusters
  })

  clusterLabels.forEach((entity) => {
    if (!entity) return
    entity.show = hasClusterOverlay && showHouses
  })

  applyClusterVisualization(h)
}

function getClusterRadius(count) {
  if (count > 40) return 30
  if (count > 20) return 22
  if (count > 10) return 16
  return 10
}

function getClusterColor(count) {
  if (count > 30) return '#dc2626'
  if (count > 10) return '#f97316'
  return '#eab308'
}

function getClusterText(count, showText) {
  if (!showText) return ''
  if (count > 99) return '100+'
  return String(count)
}

function generateClusterIcon(count, showText) {
  const color = getClusterColor(count)
  const text = getClusterText(count, showText)
  const cacheKey = `${color}:${text || 'dot'}`
  const cached = clusterIconCache.get(cacheKey)
  if (cached) return cached

  const size = 64
  const canvas = document.createElement('canvas')
  canvas.width = size
  canvas.height = size
  const ctx = canvas.getContext('2d')
  if (!ctx) return ''

  const cx = size / 2
  const cy = size / 2
  const r = 24

  ctx.save()
  ctx.shadowColor = 'rgba(0, 0, 0, 0.35)'
  ctx.shadowBlur = 8
  ctx.shadowOffsetY = 2
  ctx.beginPath()
  ctx.arc(cx, cy, r, 0, Math.PI * 2)
  ctx.fillStyle = color
  ctx.fill()
  ctx.lineWidth = 4
  ctx.strokeStyle = '#ffffff'
  ctx.stroke()
  ctx.restore()

  if (text) {
    ctx.fillStyle = '#ffffff'
    ctx.font = text.length > 2 ? '700 17px system-ui, sans-serif' : '700 22px system-ui, sans-serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(text, cx, cy)
  }

  const dataUrl = canvas.toDataURL('image/png')
  clusterIconCache.set(cacheKey, dataUrl)
  return dataUrl
}

function getClusterZoomStage(height) {
  if (height < THRESHOLD_BUILDINGS) return 'very-high'
  if (height < 25000) return 'high'
  if (height < 120000) return 'medium'
  return 'out'
}

function getZoomLevel(height = cameraHeight.value) {
  if (height > 2000000) return 7
  if (height > 1000000) return 9
  if (height > 600000) return 11
  if (height > 300000) return 13
  return 15
}

function getClusterScale(count, zoom) {
  let scale = 0.7

  if (zoom < 12) scale = 0.55
  if (zoom < 10) scale = 0.45
  if (zoom < 8) scale = 0.35

  if (count <= 5) scale *= 0.85

  return scale
}

function offsetCartesianPosition(position, eastMeters, northMeters) {
  const cartographic = Cesium.Cartographic.fromCartesian(position)
  if (!cartographic) return position

  const lat = cartographic.latitude
  const metersPerDegLat = 111320
  const metersPerDegLng = Math.max(Math.cos(lat) * 111320, 1)

  const latDeg = Cesium.Math.toDegrees(lat) + (northMeters / metersPerDegLat)
  const lngDeg = Cesium.Math.toDegrees(cartographic.longitude) + (eastMeters / metersPerDegLng)

  return Cesium.Cartesian3.fromDegrees(lngDeg, latDeg, cartographic.height || 0)
}

function adjustClusterPosition(position, existingPositions) {
  if (!viewer || viewer.isDestroyed()) return position

  let adjusted = position
  const shiftPattern = [
    [8, 8],
    [-8, 8],
    [8, -8],
    [-8, -8],
    [14, 0],
    [0, 14],
  ]

  for (let i = 0; i < shiftPattern.length; i += 1) {
    const screen1 = Cesium.SceneTransforms.worldToWindowCoordinates(viewer.scene, adjusted)
    if (!screen1) break

    const tooClose = existingPositions.some((existingPosition) => {
      const screen2 = Cesium.SceneTransforms.worldToWindowCoordinates(viewer.scene, existingPosition)
      if (!screen2) return false
      const dx = screen1.x - screen2.x
      const dy = screen1.y - screen2.y
      const distance = Math.sqrt((dx * dx) + (dy * dy))
      return distance < MIN_PIXEL_DISTANCE
    })

    if (!tooClose) return adjusted

    const [east, north] = shiftPattern[i]
    adjusted = offsetCartesianPosition(adjusted, east, north)
  }

  return adjusted
}

function applyClusterVisualization(height = cameraHeight.value) {
  if (!viewer || viewer.isDestroyed() || !clusterEntityGroups.length) return

  const zoom = getZoomLevel(height)
  const hasClusterOverlay = activeProblemFilters.value.length > 0 && clusterEntities.length > 0
  const showClusters = hasClusterOverlay && zoom < SHOW_HOUSES_ZOOM
  const showCount = zoom >= 9
  const renderedPositions = []

  clusterEntityGroups.forEach((group) => {
    const count = Number(group.count || 0)
    const scale = getClusterScale(count, zoom)

    const shouldShow = showClusters

    if (group.icon?.billboard) {
      group.icon.show = shouldShow
      if (shouldShow && group.basePosition) {
        const adjusted = adjustClusterPosition(group.basePosition, renderedPositions)
        group.icon.position = adjusted
        renderedPositions.push(adjusted)
      }
      group.icon.billboard.image = generateClusterIcon(count, showCount)
      group.icon.billboard.scale = scale
    }
  })
}

function flyToPoints(list) {
  if (!viewer || !list.length) return
  const pts = list
    .filter((h) => Number.isFinite(Number(h.lng)) && Number.isFinite(Number(h.lat)))
    .map((h) => Cesium.Cartesian3.fromDegrees(Number(h.lng), Number(h.lat), 0))

  if (!pts.length) return

  const sphere = Cesium.BoundingSphere.fromPoints(pts)
  const range = Math.max(sphere.radius * 2.6, 300)
  viewer.camera.flyToBoundingSphere(sphere, {
    duration: 2,
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

function zoomToCluster(cluster) {
  if (!viewer || !cluster) return

  const points = (cluster.households || [])
    .map((h) => ({ lat: Number(h.lat), lng: Number(h.lng) }))
    .filter((p) => Number.isFinite(p.lat) && Number.isFinite(p.lng))
    .map((p) => Cesium.Cartesian3.fromDegrees(p.lng, p.lat, 0))

  if (points.length > 1) {
    const sphere = Cesium.BoundingSphere.fromPoints(points)
    const range = Math.max(sphere.radius * 3.2, 1500)
    viewer.camera.flyToBoundingSphere(sphere, {
      duration: 1.3,
      offset: new Cesium.HeadingPitchRange(
        Cesium.Math.toRadians(0),
        Cesium.Math.toRadians(-55),
        range,
      ),
    })
    return
  }

  viewer.camera.flyTo({
    destination: Cesium.Cartesian3.fromDegrees(Number(cluster.lng), Number(cluster.lat), 1500),
    orientation: { heading: 0, pitch: Cesium.Math.toRadians(-58), roll: 0 },
    duration: 1.2,
  })
}

function openIssuePanel(house) {
  selectedHouse.value = house
  const clusterForHouse = houseClusterMap.get(getHouseKey(house))
  selectedCluster.value = normalizeClusterSelection(clusterForHouse)
}

function computeJitteredPositions(list) {
  const posMap = new Map()

  list.forEach((house, idx) => {
    const lat = Number(house.lat)
    const lng = Number(house.lng)
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) return
    const key = `${lat.toFixed(6)},${lng.toFixed(6)}`
    if (!posMap.has(key)) posMap.set(key, [])
    posMap.get(key).push(idx)
  })

  const out = list.map((house) => ({
    lat: Number(house.lat),
    lng: Number(house.lng),
  }))

  posMap.forEach((indices) => {
    if (indices.length < 2) return

    const count = indices.length
    const radiusM = Math.min(7 + count * 1.2, 20)
    const ref = list[indices[0]]
    const refLat = Number(ref.lat)
    const refLng = Number(ref.lng)
    const cosLat = Math.cos((refLat * Math.PI) / 180)

    indices.forEach((listIdx, slot) => {
      const angle = (2 * Math.PI * slot) / count
      out[listIdx] = {
        lat: refLat + (radiusM * Math.cos(angle)) / 111000,
        lng: refLng + (radiusM * Math.sin(angle)) / (111000 * cosLat),
      }
    })
  })

  return out
}

function haversineM(lat1, lng1, lat2, lng2) {
  const R = 6371000
  const p1 = (lat1 * Math.PI) / 180
  const p2 = (lat2 * Math.PI) / 180
  const dp = ((lat2 - lat1) * Math.PI) / 180
  const dl = ((lng2 - lng1) * Math.PI) / 180
  const a = Math.sin(dp / 2) ** 2 + Math.cos(p1) * Math.cos(p2) * Math.sin(dl / 2) ** 2
  return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
}

function computeProblemClusters(houseList) {
  const RADIUS_M = 300
  const MIN_SIZE = 5
  const visited = new Set()
  const clusters = []

  houseList.forEach((h, i) => {
    const lat = Number(h.lat)
    const lng = Number(h.lng)
    if (!Number.isFinite(lat) || !Number.isFinite(lng) || visited.has(i)) return

    const group = [i]
    visited.add(i)

    houseList.forEach((h2, j) => {
      const lat2 = Number(h2.lat)
      const lng2 = Number(h2.lng)
      if (visited.has(j) || !Number.isFinite(lat2) || !Number.isFinite(lng2)) return
      if (haversineM(lat, lng, lat2, lng2) <= RADIUS_M) {
        group.push(j)
        visited.add(j)
      }
    })

    if (group.length >= MIN_SIZE) {
      const clLat = group.reduce((sum, idx) => sum + Number(houseList[idx].lat), 0) / group.length
      const clLng = group.reduce((sum, idx) => sum + Number(houseList[idx].lng), 0) / group.length
      const clHouses = group.map((idx) => houseList[idx])
      clusters.push({ lat: clLat, lng: clLng, count: group.length, houses: clHouses })
    }
  })

  return clusters
}

function addClusterEntities(problemHouses) {
  clusterMap.clear()
  houseClusterMap.clear()
  clusterEntityGroups.length = 0
  clusterEntities.length = 0
  clusterLabels.length = 0
  const clusters = computeProblemClusters(problemHouses)

  clusters.forEach(({ lat, lng, count, houses: clusterHouses }) => {
    const basePosition = Cesium.Cartesian3.fromDegrees(Number(lng), Number(lat), 0)
    const problems = analyzeCluster(clusterHouses)
    const clusterData = {
      household_count: count,
      count,
      lat,
      lng,
      problems,
      issues: {
        bpl: Number(problems.find((item) => item.key === 'bplFamilies')?.count || 0),
        illiterate: Number(problems.find((item) => item.key === 'illiterateMembers')?.count || 0),
        unemployed: Number(problems.find((item) => item.key === 'unemployedMembers')?.count || 0),
        divyang: Number(problems.find((item) => item.key === 'divyangMembers')?.count || 0),
      },
      households: clusterHouses.map((h) => ({
        houseId: getHouseId(h),
        lat: Number(h.lat),
        lng: Number(h.lng),
      })),
    }

    const iconEnt = viewer.entities.add({
      position: basePosition,
      show: true,
      billboard: {
        image: generateClusterIcon(count, true),
        scale: 0.8,
        verticalOrigin: Cesium.VerticalOrigin.CENTER,
        horizontalOrigin: Cesium.HorizontalOrigin.CENTER,
        disableDepthTestDistance: Number.POSITIVE_INFINITY,
        heightReference: Cesium.HeightReference.CLAMP_TO_GROUND,
      },
    })
    iconEnt.clusterData = clusterData

    const center = getClusterCenter(clusterData)
    const labelEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(Number(center.lng), Number(center.lat), 20),
      show: false,
      label: {
        text: `High Need Area\n${Number(clusterData.household_count || 0)} households`,
        font: 'bold 14px sans-serif',
        fillColor: Cesium.Color.WHITE,
        outlineColor: Cesium.Color.RED,
        outlineWidth: 3,
        showBackground: true,
        backgroundColor: Cesium.Color.RED.withAlpha(0.85),
        pixelOffset: new Cesium.Cartesian2(0, -25),
        scale: 0.7,
        disableDepthTestDistance: Number.POSITIVE_INFINITY,
      },
    })
    labelEnt.clusterData = clusterData

    clusterIds.add(iconEnt.id)
    clusterIds.add(labelEnt.id)
    clusterEntities.push(iconEnt)
    clusterLabels.push(labelEnt)
    clusterMap.set(iconEnt.id, clusterData)
    clusterMap.set(labelEnt.id, clusterData)
    clusterHouses.forEach((house) => {
      houseClusterMap.set(getHouseKey(house), clusterData)
    })
    clusterEntityGroups.push({
      count,
      lat: Number(lat),
      lng: Number(lng),
      basePosition,
      icon: iconEnt,
    })
  })

  applyClusterVisualization(viewer.camera.positionCartographic?.height ?? cameraHeight.value)
}

/**
 * Efficiently updates only the visual states (selected/problem highlighting) of affected houses
 * without recalculating all colors or rebuilding entities.
 * This prevents the yellow selection color from masking the actual color temporarily.
 */
function updateHouseSelectionState() {
  if (!viewer) return

  const newSelectedNo = selectedHouse.value?.house_no
  const prevSelectedNo = previouslySelectedHouseNo.value

  // Only rebuild if selection actually changed and data is ready
  if (newSelectedNo === prevSelectedNo) return
  if (!isDatasetReady()) return

  previouslySelectedHouseNo.value = newSelectedNo

  // Update visual properties of affected house entities
  const updateHouseVisuals = (houseNo, isSelected) => {
    const house = houses.value.find(h => String(h.house_no || '') === String(houseNo || ''))
    if (!house) return

    houseEntities.forEach((entity) => {
      if (entity.houseData?.house_no === houseNo) {
        if (entity.box) {
          // Update materials for building entities
          if (entity.box.material && isSelected) {
            entity.box.material = Cesium.Color.fromCssColorString('#facc15').withAlpha(1.0)
            entity.box.outlineColor = Cesium.Color.fromCssColorString('#f59e0b')
            entity.box.outlineWidth = 4
          } else if (entity.box.material && !isSelected) {
            // Restore original color based on data
            entity.box.material = cesiumColor(house).withAlpha(1.0)
            const roofColor = cesiumColor(house)
            entity.box.outlineColor = roofColor.darken(0.25, new Cesium.Color())
            entity.box.outlineWidth = 1.5
          }
        }

        if (entity.point) {
          // Update point entities (far zoom)
          if (isSelected) {
            entity.point.pixelSize = 18
            entity.point.color = Cesium.Color.fromCssColorString('#facc15').withAlpha(1.0)
            entity.point.outlineColor = Cesium.Color.WHITE
            entity.point.outlineWidth = 4
          } else {
            entity.point.pixelSize = 8
            entity.point.color = cesiumColor(house).withAlpha(1.0)
            entity.point.outlineColor = Cesium.Color.fromCssColorString('#1a1a1a').withAlpha(0.7)
            entity.point.outlineWidth = 1.5
          }
        }
      }
    })
  }

  // Update previous selection (turn off yellow)
  if (prevSelectedNo) {
    updateHouseVisuals(prevSelectedNo, false)
  }

  // Update new selection (turn on yellow)
  if (newSelectedNo) {
    updateHouseVisuals(newSelectedNo, true)
  }
}

/**
 * Validates coverage status completeness across entire dataset.
 * Uses statistical sampling for large datasets (2000+ houses).
 * Requires meaningful values, not just field existence.
 */
function isCoverageCoverageReady() {
  if (!houses.value.length) return true

  const needsCoverage = colorMode.value === 'aadhaar_coverage' || colorMode.value === 'caste_certificate_coverage'
  if (!needsCoverage) return true

  // For large datasets, use statistical sampling to avoid checking all 2000+ houses
  const sampleSize = Math.min(Math.max(Math.ceil(houses.value.length * 0.05), 20), 100) // 5-20% or 20-100 houses
  const sampleIndices = []
  for (let i = 0; i < sampleSize; i++) {
    sampleIndices.push(Math.floor((i * houses.value.length) / sampleSize))
  }

  if (colorMode.value === 'aadhaar_coverage') {
    // ALL sampled houses must have MEANINGFUL coverage status (not empty string)
    return sampleIndices.every(idx => {
      const h = houses.value[idx]
      const status = String(h?.aadhaarCoverageStatus || '').toLowerCase().trim()
      // Explicitly require valid status values - empty string means data not calculated yet
      return status === 'complete' || status === 'partial' || status === 'missing'
    })
  }

  if (colorMode.value === 'caste_certificate_coverage') {
    return sampleIndices.every(idx => {
      const h = houses.value[idx]
      const status = String(h?.casteCertificateCoverageStatus || '').toLowerCase().trim()
      return status === 'complete' || status === 'partial' || status === 'missing'
    })
  }

  return true
}

/**
 * Check if dataset has required fields for current color mode.
 * Prevents rendering incomplete data that would show wrong colors.
 */
function isDatasetReady() {
  if (!houses.value.length) return true // Empty dataset is "ready"

  const sampleHouses = houses.value.slice(0, 10) // Check first 10 for efficiency

  if (colorMode.value === 'employment_status') {
    // Ensure working_members is present (may be 0)
    return sampleHouses.every(h => h.working_members !== undefined && h.working_members !== null)
  }

  if (colorMode.value === 'education_status') {
    return sampleHouses.every(h => h.literate_members !== undefined && h.illiterate_members !== undefined)
  }

  // For coverage modes, use the more thorough statistical validation
  if (colorMode.value === 'aadhaar_coverage' || colorMode.value === 'caste_certificate_coverage') {
    return isCoverageCoverageReady()
  }

  if (colorMode.value === 'bpl_status') {
    return sampleHouses.every(h => h.FAMILY_BELONG_BPL_CATEGORY !== undefined || h.bpl_status !== undefined)
  }

  // population_density and divyang_presence only need total_members and has_disability
  return true
}

function buildEntities() {
  if (!viewer || !houses.value.length) return

  // Skip rendering if critical data fields are missing
  // This prevents flashing wrong colors during initial load
  if (!isDatasetReady()) {
    console.debug('[buildEntities] Skipped: Dataset not ready for color mode', colorMode.value)
    return
  }

  // Clear color cache to ensure fresh calculations
  clearColorCache()

  // Log render trigger for debugging
  console.debug('[buildEntities] Render triggered:', {
    houseCount: houses.value.length,
    selectedHouseNo: selectedHouse.value?.house_no,
    colorMode: colorMode.value,
    hasProblemFilter: activeProblemFilters.value.length > 0,
    timestamp: new Date().toLocaleTimeString(),
  })

  viewer.entities.removeAll()
  entityMap.clear()
  buildingIds.clear()
  pointIds.clear()
  clusterIds.clear()
  clusterMap.clear()
  houseClusterMap.clear()
  clusterEntityGroups.length = 0
  clusterEntities.length = 0
  clusterLabels.length = 0
  houseEntities.length = 0
  selectedCluster.value = null

  const selectedNo = selectedHouse.value?.house_no
  const camH = viewer.camera.positionCartographic?.height ?? cameraHeight.value
  const showBuildings = camH < THRESHOLD_BUILDINGS
  const jittered = computeJitteredPositions(houses.value)
  const hasProblemFilter = activeProblemFilters.value.length > 0
  const problemHouses = []

  houses.value.forEach((house, idx) => {
    const { lat, lng } = jittered[idx]
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) return

    const isSelected = selectedNo && String(house.house_no || '') === String(selectedNo)
    const isProblem = hasProblemFilter && matchesAllProblems(house)
    const isBackground = hasProblemFilter && !isProblem && !isSelected

    if (isProblem) {
      problemHouses.push(house)
    }

    // Determine roof color with data validation
    let roofColor
    if (isSelected) {
      roofColor = Cesium.Color.fromCssColorString('#facc15').withAlpha(1.0)
    } else {
      // For incomplete data, use neutral gray instead of potentially misleading colors
      const needsDataValidation =
        (colorMode.value === 'aadhaar_coverage' || colorMode.value === 'caste_certificate_coverage' || colorMode.value === 'employment_status')
        && !isHouseDataReady(house)

      if (needsDataValidation) {
        roofColor = Cesium.Color.fromCssColorString('#9ca3af').withAlpha(isBackground ? 0.35 : 1.0)
      } else {
        roofColor = cesiumColor(house).withAlpha(isBackground ? 0.35 : 1.0)
      }
    }

    const wallColor = isSelected
      ? Cesium.Color.fromCssColorString('#fef3c7').withAlpha(1.0)
      : isProblem
        ? Cesium.Color.fromCssColorString('#f4b8b8').withAlpha(0.95)
        : Cesium.Color.fromCssColorString('#c8a97e').withAlpha(isBackground ? 0.3 : 1.0)

    const wallOutline = isSelected
      ? Cesium.Color.fromCssColorString('#f59e0b')
      : isProblem
        ? Cesium.Color.fromCssColorString('#dc2626').withAlpha(1.0)
        : Cesium.Color.fromCssColorString('#7a6040').withAlpha(isBackground ? 0.2 : 1.0)

    const footprint = 10
    const baseH = 7
    const roofH = Math.max(2.5, Math.min(buildingHeight(house) * 0.22, 5))

    const baseEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, baseH / 2),
      show: showBuildings,
      box: {
        dimensions: new Cesium.Cartesian3(footprint, footprint, baseH),
        material: wallColor,
        outline: true,
        outlineColor: wallOutline,
        outlineWidth: isSelected ? 4 : 1.5,
      },
    })
    baseEnt.houseData = house
    houseEntities.push(baseEnt)

    const roofEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, baseH + roofH / 2),
      show: showBuildings,
      box: {
        dimensions: new Cesium.Cartesian3(footprint * 0.88, footprint * 0.88, roofH),
        material: roofColor,
        outline: true,
        outlineColor: isSelected
          ? Cesium.Color.WHITE
          : roofColor.darken(0.25, new Cesium.Color()),
        outlineWidth: isSelected ? 4 : 1.5,
      },
    })
    roofEnt.houseData = house
    houseEntities.push(roofEnt)

    const ptEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, 1),
      show: !showBuildings,
      point: {
        pixelSize: isSelected ? 13 : isProblem ? 11 : isBackground ? 5 : 8,
        color: roofColor,
        outlineColor: isSelected
          ? Cesium.Color.WHITE
          : isProblem
            ? Cesium.Color.fromCssColorString('#dc2626').withAlpha(0.9)
            : Cesium.Color.fromCssColorString('#1a1a1a').withAlpha(isBackground ? 0.25 : 0.7),
        outlineWidth: isSelected ? 2 : isProblem ? 2.5 : 1.5,
        heightReference: Cesium.HeightReference.CLAMP_TO_GROUND,
      },
    })
    ptEnt.houseData = house
    houseEntities.push(ptEnt)

    buildingIds.add(baseEnt.id)
    buildingIds.add(roofEnt.id)
    pointIds.add(ptEnt.id)

    entityMap.set(baseEnt.id, house)
    entityMap.set(roofEnt.id, house)
    entityMap.set(ptEnt.id, house)
  })

  if (hasProblemFilter && problemHouses.length) {
    addClusterEntities(problemHouses)
  }

  // Track current selection state for efficient future updates
  previouslySelectedHouseNo.value = selectedHouse.value?.house_no || null

  updateZoomVisibility()
}

function buildQueryParams() {
  let colorBy = 'population_density'
  if (selectedColorBy.value === 'bpl_status') colorBy = 'bpl'
  else if (selectedColorBy.value === 'employment_status') colorBy = 'employment'
  else if (selectedColorBy.value === 'divyang_presence') colorBy = 'divyang'
  else if (selectedColorBy.value === 'education_status') colorBy = 'education'

  return {
    district_id: filterDistrict.value || '',
    taluka_id: filterTaluka.value || '',
    village_id: filterVillage.value || '',
    color_by: colorBy,
  }
}

function renderChartToBase64(segments) {
  const valid = (segments || []).filter((s) => Number(s?.count || 0) > 0)
  const total = valid.reduce((sum, s) => sum + Number(s.count || 0), 0)
  if (!total) return ''

  const size = 280
  const canvas = document.createElement('canvas')
  canvas.width = size
  canvas.height = size
  const ctx = canvas.getContext('2d')
  if (!ctx) return ''

  const cx = size / 2
  const cy = size / 2
  const outerR = 110
  const innerR = 58
  let startAngle = -Math.PI / 2

  valid.forEach((seg) => {
    const sliceAngle = (Number(seg.count || 0) / total) * 2 * Math.PI
    ctx.beginPath()
    ctx.moveTo(cx, cy)
    ctx.arc(cx, cy, outerR, startAngle, startAngle + sliceAngle)
    ctx.closePath()
    ctx.fillStyle = seg.color
    ctx.fill()
    startAngle += sliceAngle
  })

  ctx.beginPath()
  ctx.arc(cx, cy, innerR, 0, 2 * Math.PI)
  ctx.fillStyle = '#ffffff'
  ctx.fill()

  return canvas.toDataURL('image/png').replace('data:image/png;base64,', '')
}

function getFilterName(options, id) {
  if (!id) return ''
  return options.find((item) => String(item.id) === String(id))?.name || ''
}

function buildPopulationCharts() {
  const male = Number(maleMembers.value || 0)
  const female = Number(femaleMembers.value || 0)

  const working = Number(houses.value.reduce((sum, h) => sum + Number(h.working_members || 0), 0))
  const dependent = Math.max(Number(totalMembers.value || 0) - working, 0)

  const edu = insights.value?.education_status || {}
  const literate = Number(edu.literate || 0)
  const illiterate = Number(edu.illiterate || 0)
  const students = Number(edu.students || 0)
  const dropouts = Number(edu.dropouts || 0)

  const bpl = houses.value.filter((h) => getBplStatusLabel(h) === 'BPL').length
  const nonBpl = Math.max(houses.value.length - bpl, 0)

  const fs1to2 = houses.value.filter((h) => Number(h.total_members || 0) <= 2).length
  const fs3to5 = houses.value.filter((h) => {
    const m = Number(h.total_members || 0)
    return m >= 3 && m <= 5
  }).length
  const fs6plus = houses.value.filter((h) => Number(h.total_members || 0) >= 6).length

  const chartDefs = [
    {
      title: 'Gender Ratio Chart',
      segments: [
        { count: male, color: '#2563eb' },
        { count: female, color: '#ec4899' },
      ],
    },
    {
      title: 'Employment Distribution Chart',
      segments: [
        { count: working, color: '#16a34a' },
        { count: dependent, color: '#f59e0b' },
      ],
    },
    {
      title: 'Education Distribution Chart',
      segments: [
        { count: literate, color: '#3b82f6' },
        { count: illiterate, color: '#6b7280' },
        { count: students, color: '#14b8a6' },
        { count: dropouts, color: '#ef4444' },
      ],
    },
    {
      title: 'BPL Distribution Chart',
      segments: [
        { count: bpl, color: '#ef4444' },
        { count: nonBpl, color: '#16a34a' },
      ],
    },
    {
      title: 'Family Size Distribution Chart',
      segments: [
        { count: fs1to2, color: '#22c55e' },
        { count: fs3to5, color: '#f59e0b' },
        { count: fs6plus, color: '#ef4444' },
      ],
    },
  ]

  return chartDefs
    .map((c) => ({ title: c.title, image: renderChartToBase64(c.segments) }))
    .filter((c) => c.image)
}

async function downloadPDF() {
  if (pdfLoading.value) return
  pdfLoading.value = true
  try {
    const districtName = filterDistrict.value ? getFilterName(districtOptions.value, filterDistrict.value) : ''
    const talukaName = filterTaluka.value ? getFilterName(talukaOptions.value, filterTaluka.value) : ''
    const villageName = filterVillage.value ? getFilterName(villageOptions.value, filterVillage.value) : ''

    const body = {
      districtId: filterDistrict.value ? String(filterDistrict.value) : '',
      districtName,
      talukaId: filterTaluka.value ? String(filterTaluka.value) : '',
      talukaName,
      villageId: filterVillage.value ? String(filterVillage.value) : '',
      villageName,
      charts: buildPopulationCharts(),
    }

    const res = await fetch('/api/pdf/population-report', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })

    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      console.error('[Population PDF] Generation failed:', err)
      return
    }

    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    const stem = ['PopTwin', districtName.replace(/\s+/g, '_') || '', talukaName.replace(/\s+/g, '_') || '', villageName.replace(/\s+/g, '_') || '', new Date().toISOString().slice(0, 10)]
      .filter(Boolean)
      .join('_')
    a.download = `${stem}.pdf`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (error) {
    console.error('[Population PDF] Download error:', error)
  } finally {
    pdfLoading.value = false
  }
}

async function loadLocationOptions() {
  try {
    const res = await getLocationOptions({ district_id: pendingDistrict.value || undefined, taluka_id: pendingTaluka.value || undefined })
    districtOptions.value = res?.districts || []
    talukaOptions.value = res?.talukas || []
    villageOptions.value = res?.villages || []
  } catch (error) {
    districtOptions.value = []
    talukaOptions.value = []
    villageOptions.value = []
    console.warn('Population 3D twin location options unavailable:', error?.message || error)
  }
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
  clearColorCache()
  if (viewer) buildEntities()
})

watch(activeProblemFilters, () => {
  if (viewer) buildEntities()
}, { deep: true })

/**
 * When house is selected, ONLY update visual selection state (yellow highlight)
 * without recalculating all colors. This prevents color thrashing and color-flip bugs.
 */
watch(selectedHouse, (newHouse) => {
  if (!viewer) return

  // Log for debugging color-flip issues
  if (newHouse) {
    console.debug('[Selection] House selected:', {
      house_no: newHouse.house_no,
      total_members: newHouse.total_members,
      working_members: newHouse.working_members,
      aadhaarCoverageStatus: newHouse.aadhaarCoverageStatus,
      colorMode: colorMode.value,
      dataReady: isDatasetReady(),
    })
  }

  // Use efficient state update instead of full rebuild
  updateHouseSelectionState()
})

onMounted(async () => {
  isTwinFullscreen.value = !!document.fullscreenElement
  document.addEventListener('fullscreenchange', handleTwinFullscreenChange)

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
    imageryProvider: buildImageryProvider('street'),
  })

  viewer.scene.globe.enableLighting = true
  viewer.screenSpaceEventHandler.removeInputAction(Cesium.ScreenSpaceEventType.LEFT_DOUBLE_CLICK)

  viewer.screenSpaceEventHandler.setInputAction((event) => {
    const picked = viewer.scene.pick(event.position)
    if (!picked?.id) return
    const pickedEntity = picked.id
    const entityId = pickedEntity.id || pickedEntity
    const cluster = pickedEntity.clusterData || clusterMap.get(entityId)
    if (cluster) {
      zoomToCluster(cluster)
      selectedHouse.value = null
      selectedCluster.value = normalizeClusterSelection(cluster)
      return
    }
    const house = pickedEntity.houseData || entityMap.get(entityId)
    if (house) {
      openIssuePanel(house)
    }
  }, Cesium.ScreenSpaceEventType.LEFT_CLICK)


  viewer.screenSpaceEventHandler.setInputAction((event) => {
    const picked = viewer.scene.pick(event.position)
    if (!picked?.id) return
    const house = entityMap.get(picked.id.id || picked.id)
    if (house) flyToHouse(house)
  }, Cesium.ScreenSpaceEventType.LEFT_DOUBLE_CLICK)

  viewer.camera.percentageChanged = 0.03
  viewer.camera.changed.addEventListener(updateZoomVisibility)

  await loadTwinData()
  loadLocationOptions()
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', handleTwinFullscreenChange)

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

.map-fs-btn {
  position: absolute;
  top: 58px;
  right: 12px;
  z-index: 210;
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.96);
  border: 1.5px solid #d1d5db;
  border-radius: 8px;
  color: #374151;
  font-size: 0.95rem;
  line-height: 1;
  cursor: pointer;
  transition: all 0.15s ease;
  box-shadow: 0 2px 8px rgba(0,0,0,0.12);
}
.map-fs-btn.shifted {
  right: 340px;
}
.map-fs-btn:hover {
  border-color: #16a34a;
  color: #16a34a;
  background: #f0fdf4;
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
.cs-option-group-label {
  padding: 0.3rem 0.75rem 0.15rem;
  font-size: 0.68rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: #9ca3af;
  background: #f9fafb;
  border-top: 1px solid #e5e7eb;
  pointer-events: none;
  user-select: none;
}
.cs-option-group-label:first-child { border-top: none; }

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

/* Mini 3D house icon to match Agriculture twin legend style */
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
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-bottom: 6px solid var(--mh-roof, #94a3b8);
  transition: border-bottom-color 0.2s;
}

.mh-wall {
  width: 12px;
  height: 8px;
  background: #c8a97e;
  border: 1px solid rgba(0,0,0,0.18);
  box-shadow: inset 1px 0 0 rgba(255,255,255,0.18), 1px 1px 0 rgba(0,0,0,0.12);
}

.mini-house-sm .mh-roof {
  border-left-width: 6px;
  border-right-width: 6px;
  border-bottom-width: 5px;
}

.mini-house-sm .mh-wall {
  width: 9px;
  height: 6px;
}

.pf-item {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.32rem 0.35rem;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  margin-bottom: 0.3rem;
  cursor: pointer;
  transition: background 0.12s, border-color 0.12s;
}

.pf-item:hover {
  background: #f9fafb;
  border-color: #e5e7eb;
}

.pf-check {
  margin: 0;
  accent-color: #16a34a;
}

.pf-label {
  flex: 1;
  font-size: 0.7rem;
  color: #111827;
  font-weight: 600;
}

.pf-count {
  font-size: 0.65rem;
  color: #6b7280;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
}

.pf-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  margin-top: 0.35rem;
  padding-top: 0.45rem;
  border-top: 1px solid #e5e7eb;
  font-size: 0.68rem;
  color: #374151;
}

.pf-clear-btn {
  background: #fff1f2;
  border: 1px solid #fca5a5;
  color: #dc2626;
  border-radius: 4px;
  font-size: 0.62rem;
  font-weight: 700;
  padding: 0.16rem 0.45rem;
  cursor: pointer;
}

.pf-clear-btn:hover {
  background: #dc2626;
  border-color: #dc2626;
  color: #ffffff;
}

.pf-hint {
  font-size: 0.66rem;
  color: #6b7280;
  line-height: 1.45;
  margin-top: 0.35rem;
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

.filter-section {
  padding: 0.65rem 0.9rem;
  border-bottom: 1.5px solid #e5e7eb;
}
.filter-label {
  display: block; font-size: 0.66rem; color: #6b7280;
  text-transform: uppercase; letter-spacing: 0.05em;
  font-weight: 600; margin-bottom: 0.4rem;
}
.filter-section .custom-select {
  position: relative;
}
.filter-section .cs-trigger {
  width: 100%; padding: 0.4rem 0.6rem;
  border: 1.5px solid #e5e7eb;
  border-radius: var(--radius-sm);
  background: #f9fafb;
  cursor: pointer;
  font-size: 0.73rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: all 0.2s;
}
.filter-section .cs-trigger:hover {
  border-color: #d1d5db;
}
.filter-section .cs-dropdown {
  position: absolute; top: 100%; left: 0; right: 0;
  background: #ffffff;
  border: 1.5px solid #d1d5db;
  border-radius: var(--radius-sm);
  margin-top: 0.25rem;
  z-index: 200;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  max-height: 200px;
  overflow-y: auto;
}
.filter-section .cs-option {
  padding: 0.5rem 0.6rem;
  font-size: 0.73rem;
  cursor: pointer;
  border-bottom: 1px solid #f3f4f6;
  transition: background 0.15s;
}
.filter-section .cs-option:last-child {
  border-bottom: none;
}
.filter-section .cs-option:hover {
  background: #f3f4f6;
}
.filter-section .cs-option.selected {
  background: #dbeafe;
  color: #1e40af;
  font-weight: 600;
}

.detail-empty {
  padding: 1rem 0.9rem;
  text-align: center;
  font-size: 0.73rem;
  color: #9ca3af;
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

.cluster-panel {
  position: absolute;
  right: 0.75rem;
  top: 60px;
  width: 300px;
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

.cluster-close {
  position: absolute;
  right: 0.65rem;
  top: 0.65rem;
  z-index: 1;
}

.cluster-header {
  padding: 0.95rem 0.9rem 0.7rem;
  border-bottom: 1.5px solid #e5e7eb;
  background: #fff1f2;
  border-radius: var(--radius) var(--radius) 0 0;
}

.cluster-badge {
  display: inline-block;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  background: #fee2e2;
  border: 1px solid #fca5a5;
  color: #b91c1c;
  font-size: 0.64rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.cluster-count {
  margin-top: 0.45rem;
  font-size: 0.74rem;
  color: #374151;
}

.cluster-count strong {
  color: #ef4444;
  font-size: 1.1rem;
}

.cluster-section-title {
  font-size: 0.62rem;
  font-weight: 800;
  color: #374151;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 0.7rem 0.9rem 0.35rem;
}

.cp-card {
  margin: 0.5rem 0.9rem;
  border: 1.5px solid #e5e7eb;
  border-radius: var(--radius-sm);
  background: #f9fafb;
  padding: 0.6rem 0.65rem;
}

.cp-top {
  display: flex;
  gap: 0.45rem;
  align-items: flex-start;
  margin-bottom: 0.35rem;
}

.cp-emoji {
  font-size: 0.95rem;
  line-height: 1;
  margin-top: 0.05rem;
}

.cp-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 0;
}

.cp-label {
  font-size: 0.76rem;
  font-weight: 700;
  color: #111827;
}

.cp-stat {
  font-size: 0.64rem;
  color: #6b7280;
}

.cp-bar-track {
  height: 5px;
  background: #e5e7eb;
  border-radius: 3px;
  overflow: hidden;
}

.cp-bar-fill {
  height: 100%;
  background: #ef4444;
  border-radius: 3px;
}

.cp-action {
  margin-top: 0.45rem;
  font-size: 0.68rem;
  font-weight: 700;
  color: #b91c1c;
}

.cp-solution {
  margin: 0.25rem 0 0;
  font-size: 0.71rem;
  color: #374151;
  line-height: 1.5;
}

.cp-scheme {
  margin-top: 0.38rem;
  display: inline-block;
  font-size: 0.66rem;
  font-weight: 700;
  color: #b91c1c;
  border: 1.5px solid #fca5a5;
  border-radius: 4px;
  padding: 0.18rem 0.48rem;
  background: #fff1f2;
}

.cluster-ok {
  margin: 0.7rem 0.9rem 0.9rem;
  font-size: 0.72rem;
  color: #15803d;
  font-weight: 600;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: var(--radius-sm);
  padding: 0.5rem 0.6rem;
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
  .map-fs-btn.shifted { right: 12px; }
}
</style>
