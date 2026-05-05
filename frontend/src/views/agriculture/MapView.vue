<template>
  <div class="map-page">
    <header class="map-header">
      <div class="map-title-area">
        <h1 class="page-title">Geo-Intelligence Map</h1>
        <p class="page-subtitle">
          <template v-if="loading && houses.length === 0">
            Loading household data from the live database…
          </template>
          <template v-else>
            {{ houses.length.toLocaleString() }} households plotted from the live database
          </template>
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
          <div class="custom-select" :class="{ open: openDropdown === 'district', disabled: isDistrictLoading }"
               @click.stop="!isDistrictLoading && toggleDropdown('district')">
            <button class="cs-trigger" type="button" :disabled="isDistrictLoading">
              <span class="cs-value">{{ selectedDistrictLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown" v-show="openDropdown === 'district'" @click.stop>
              <div class="cs-option" v-for="d in districtOptions" :key="`${d.value ?? ''}`"
                   :class="{ selected: isDistrictRowSelected(selectedDistrict, d) }"
                   @click="selectDistrict(d)">{{ d.label }}</div>
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
              <div class="cs-option" :class="{ selected: !selectedTaluka }" @click="selectTaluka(null)">All</div>
              <div class="cs-option" v-for="t in talukaOptions" :key="t.value"
                   :class="{ selected: isGeoOptionSelected(selectedTaluka, t) }"
                   @click="selectTaluka(t)">{{ t.label }}</div>
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
              <div class="cs-option" :class="{ selected: !selectedVillage }" @click="selectVillage(null)">All</div>
              <div class="cs-option" v-for="v in villageOptions" :key="v.value"
                   :class="{ selected: isGeoOptionSelected(selectedVillage, v) }"
                   @click="selectVillage(v)">{{ v.label }}</div>
            </div>
          </div>
        </div>

        <div class="map-control-group">
          <button class="apply-btn" @click="() => applyFilters(true)">Apply</button>
          <button class="reset-btn" @click="resetFilters">Reset</button>
        </div>

        <!-- Anomaly toggle button -->
        <div class="map-control-group">
          <button
            class="anomaly-toggle-btn"
            :class="{ active: showAnomalies, 'no-anomalies': !anomalies.length }"
            @click="toggleAnomalies"
            :disabled="!anomalies.length"
            :title="anomalies.length ? `${anomalies.length} houses have an incorrect location recorded in the survey` : 'No location errors found'"
          >
            <svg class="anomaly-btn-icon" viewBox="0 0 20 20" fill="currentColor" width="13" height="13">
              <path fill-rule="evenodd" d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 5zm0 9a1 1 0 100-2 1 1 0 000 2z" clip-rule="evenodd"/>
            </svg>
            <span>{{ anomalies.length }} GPS Mismatches</span>
            <span v-if="showAnomalies" class="anomaly-active-dot"></span>
          </button>
        </div>

        <!-- View by (only in points mode) -->
        <div class="map-control-group" v-if="viewMode === 'points'">
          <label class="control-label">VIEW BY</label>
          <div class="custom-select cs-align-right" :class="{ open: openDropdown === 'colorMode' }"
               @click.stop="toggleDropdown('colorMode')">
            <button class="cs-trigger view-by-btn" type="button" :class="{ 'cs-trigger-placeholder': !colorMode }">
              <span class="cs-value">{{ selectedColorModeLabel || 'Select a view...' }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown cs-dropdown-right" v-show="openDropdown === 'colorMode'" @click.stop>
              <template v-for="group in groupedColorOptions" :key="group.label">
                <div class="cs-option-group-label">— {{ group.label }} —</div>
                <div
                  class="cs-option"
                  v-for="option in group.options"
                  :key="option.value"
                  :class="{ selected: colorMode === option.value }"
                  @click="selectColorMode(option.value)"
                >
                  {{ option.label }}
                </div>
              </template>
            </div>
          </div>
        </div>

        <div class="map-legend">
          <template v-if="viewMode === 'villages'">
            <div class="legend-item"><span class="legend-dot" style="background:#16a34a;"></span>Irrigation available</div>
            <div class="legend-item"><span class="legend-dot" style="background:#ef4444;"></span>No irrigation</div>
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
      <div v-if="hasAppliedFilters && !loading && houses.length === 0" class="empty-state">
        No live household data returned from the database API.
      </div>

      <div class="map-content" ref="mapContentRef">
        <div class="map-stage">
          <div class="map-container" :class="{ 'map-container--hidden': !isMapVisualReady }" ref="mapContainer"></div>

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

          <!-- Household Detail Panel -->
          <transition name="slide">
            <aside v-if="selectedHouse && viewMode === 'points'" class="detail-panel">

            <!-- ── Header ── -->
            <div class="detail-header">
              <div class="detail-header-info">
                <div class="detail-badge"
                     :style="{ background: (isPopulationMode ? getMarkerColor(selectedHouse) : getConditionColor(selectedHouse)) + '18',
                               borderColor: (isPopulationMode ? getMarkerColor(selectedHouse) : getConditionColor(selectedHouse)) + '55',
                               color: (isPopulationMode ? getMarkerColor(selectedHouse) : getConditionColor(selectedHouse)) }">
                  {{ isPopulationMode ? selectedColorModeLabel : getConditionLabel(selectedHouse) }}
                </div>
                <div class="detail-name">{{ selectedHouse.headName || getHouseHeadName(selectedHouse) || 'Household' }}</div>
                <div class="detail-sub">
                  <span class="detail-id-chip">ID {{ selectedHouse.familyId || selectedHouse.FAMILY_ID || selectedHouse.EXTERNAL_FAMILY_ID || '—' }}</span>
                  <span v-if="selectedHouse.villageName">{{ selectedHouse.villageName }}</span>
                  <span v-if="selectedHouse.talukaName"> · {{ selectedHouse.talukaName }}</span>
                </div>
              </div>
              <button class="detail-close" @click="selectedHouse = null" title="Close">×</button>
            </div>

            <template v-if="isPopulationMode">
              <div class="dp-section-label">
                <span class="dp-section-icon">👪</span> Population
              </div>

              <div class="dp-field-row">
                <span class="dp-field-icon">👤</span>
                <span class="dp-field-key">Household Head</span>
                <span class="dp-field-val">{{ selectedHouse.headName || getHouseHeadName(selectedHouse) || '—' }}</span>
              </div>

              <div class="dp-field-row">
                <span class="dp-field-icon">🏠</span>
                <span class="dp-field-key">House Number</span>
                <span class="dp-field-val">{{ getHouseNumber(selectedHouse) || '—' }}</span>
              </div>

              <div class="dp-field-row">
                <span class="dp-field-icon">👥</span>
                <span class="dp-field-key">Total Members</span>
                <span class="dp-field-val">{{ getTotalMembers(selectedHouse).toLocaleString() }}</span>
              </div>

              <div class="dp-field-row">
                <span class="dp-field-icon">♂</span>
                <span class="dp-field-key">Male Count</span>
                <span class="dp-field-val">{{ getMaleMembers(selectedHouse).toLocaleString() }}</span>
              </div>

              <div class="dp-field-row">
                <span class="dp-field-icon">♀</span>
                <span class="dp-field-key">Female Count</span>
                <span class="dp-field-val">{{ getFemaleMembers(selectedHouse).toLocaleString() }}</span>
              </div>

              <template v-if="colorMode === 'employment_status' || colorMode === 'employment'">
                <div class="dp-field-row">
                  <span class="dp-field-icon">🧰</span>
                  <span class="dp-field-key">Working Members</span>
                  <span class="dp-field-val">{{ getWorkingMembers(selectedHouse).toLocaleString() }}</span>
                </div>

                <div class="dp-field-row">
                  <span class="dp-field-icon">📋</span>
                  <span class="dp-field-key">Occupation</span>
                  <span class="dp-field-val">{{ getWorkingOccupations(selectedHouse) }}</span>
                </div>
              </template>
            </template>

            <template v-else>
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
            </template>

            <!-- ── Village Mismatch (shown for GPS anomaly markers) ── -->
            <template v-if="selectedHouse._distanceKm != null">
              <div class="dp-section-label gps-section-label">
                <span class="dp-section-icon">⚠️</span> Village Mismatch
              </div>
              <div class="gps-mismatch-card">

                <!-- Mismatch description -->
                <div class="gps-mismatch-headline">
                  Mismatch Detected: This household belongs to
                  <strong class="gps-village-db">{{ selectedHouse.villageName || 'Unknown' }}</strong>,
                  but the survey was recorded in
                  <strong class="gps-village-plotted">{{ selectedHouse._plottedVillage || 'another village' }}</strong> area.
                </div>

                <!-- Distance offset — bold red -->
                <div class="gps-offset-row">
                  <span class="gps-offset-label">Distance offset</span>
                  <span class="gps-offset-value">{{ selectedHouse._distanceKm }} km away</span>
                </div>

                <!-- Village comparison table -->
                <div class="gps-village-compare">
                  <div class="gps-vc-col gps-vc-db">
                    <div class="gps-vc-badge">Database</div>
                    <div class="gps-vc-name">{{ selectedHouse.villageName || '—' }}</div>
                    <div class="gps-vc-sub">{{ selectedHouse.talukaName || '' }}</div>
                  </div>
                  <div class="gps-vc-arrow">→</div>
                  <div class="gps-vc-col gps-vc-plotted">
                    <div class="gps-vc-badge">Plotted At</div>
                    <div class="gps-vc-name">{{ selectedHouse._plottedVillage || 'Unknown area' }}</div>
                    <div class="gps-vc-sub">GPS location</div>
                  </div>
                </div>

                <!-- Helpful tip — no jargon -->
                <div class="gps-mismatch-tip">
                  💡 <strong>Tip:</strong> This usually happens if the GPS was captured at the wrong location
                  during the survey. Please re-verify on-ground.
                </div>
              </div>
            </template>

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

          <!-- ── Anomaly side panel — right-side floating, collapsible ── -->
        </div>

        <transition name="analytics-panel-slide">
          <aside v-if="analyticsPanelOpen" class="analytics-panel" aria-label="Map analytics">
            <div class="analytics-panel-head">
              <h2 class="analytics-panel-title">Map Analytics</h2>
              <button class="analytics-close" type="button" @click="analyticsPanelOpen = false" aria-label="Close analytics">×</button>
            </div>

            <div class="analytics-scroll" v-if="analyticsChart">
              <article class="analytics-card">
                <div class="analytics-card-head">
                  <div>
                    <h3 class="analytics-title">{{ analyticsChart.title }}</h3>
                    <p class="analytics-subtitle">{{ analyticsChart.subtitle }}</p>
                  </div>
                  <div class="analytics-total">{{ analyticsChart.totalLabel }}</div>
                </div>
                <div class="chart-layout">
                  <div class="donut" :style="pieStyle(analyticsChart.segments)">
                    <div class="donut-hole">
                      <div class="donut-label">{{ analyticsChart.centerLabel }}</div>
                      <div class="donut-value">{{ analyticsChart.centerValue }}</div>
                    </div>
                  </div>
                  <div class="legend-list">
                    <div class="legend-row" v-for="item in analyticsChart.segments" :key="item.label">
                      <span class="legend-dot" :style="{ background: item.color }"></span>
                      <span class="legend-name">{{ item.label }}</span>
                      <span class="legend-value">{{ item.value.toLocaleString() }}</span>
                    </div>
                  </div>
                </div>
              </article>
            </div>
          </aside>
        </transition>

        <!-- ── GPS Mismatch Sidebar — right-side collapsible, flex sibling to map-stage ── -->
        <div
          v-if="anomalyDrawerOpen && anomalies.length"
          class="anomaly-sidebar"
          :class="{ 'asb-collapsed': alpCollapsed }"
        >
          <!-- Toggle button lives on the left edge, same as 3D Twin sidebar-toggle -->
          <button
            class="asb-toggle-btn"
            @click="toggleAlpCollapse"
            :title="alpCollapsed ? 'Expand GPS Mismatches' : 'Collapse panel'"
          >
            <span class="asb-badge" v-if="alpCollapsed">{{ anomalies.length }}</span>
            {{ alpCollapsed ? '‹' : '›' }}
          </button>

          <!-- Panel body — slides away when collapsed -->
          <div class="asb-body">

            <!-- Header -->
            <div class="asb-head">
              <div class="asb-head-icon">
                <svg viewBox="0 0 20 20" fill="currentColor" width="13" height="13">
                  <path fill-rule="evenodd" d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 5zm0 9a1 1 0 100-2 1 1 0 000 2z" clip-rule="evenodd"/>
                </svg>
              </div>
              <div class="asb-head-text">
                <div class="asb-title">GPS Mismatches</div>
                <div class="asb-subtitle">{{ anomalies.length }} incorrect locations found</div>
              </div>
              <button class="asb-close" @click="closeAnomalySidebar()" title="Close">×</button>
            </div>

            <!-- Hint -->
            <div class="asb-hint">
              <span class="asb-hint-dot"></span>
              Red dots = village mismatch · click a name to zoom in
            </div>

            <!-- Scrollable list -->
            <div class="asb-list" ref="alpListRef">
              <button
                v-for="(house, i) in anomalies"
                :key="house.familyId"
                :data-fid="house.familyId"
                class="asb-item"
                :class="{ 'asb-item--active': selectedAnomalyId === house.familyId }"
                @click="flyToAnomaly(house)"
              >
                <div class="asb-item-num">{{ i + 1 }}</div>
                <div class="asb-item-body">
                  <div class="asb-item-name">{{ house.headName || 'Unknown' }}</div>
                  <div class="asb-item-meta">
                    <span class="asb-item-village">{{ house.villageName || '—' }}</span>
                    <span class="asb-item-dist">{{ house._distanceKm }} km off</span>
                  </div>
                </div>
                <svg class="asb-item-arrow" viewBox="0 0 16 16" fill="currentColor" width="11" height="11">
                  <path fill-rule="evenodd" d="M1 8a.5.5 0 01.5-.5h11.793l-3.147-3.146a.5.5 0 01.708-.708l4 4a.5.5 0 010 .708l-4 4a.5.5 0 01-.708-.708L13.293 8.5H1.5A.5.5 0 011 8z" clip-rule="evenodd"/>
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { getDistrictCentroids, getDistrictSurveyCounts, getDistricts, getHouses, getLocationOptions } from '../../api/index.js'
import { getPopulationMapData } from '../population/api.js'
import L from 'leaflet'

const loading       = ref(true)
const hasAppliedFilters = ref(false)
const houses        = ref([])
const familyMembers = ref([])
const populationStatsByFamily = ref(new Map())
const populationRowsBySignature = ref(new Map())
const populationRowsByHouseVillage = ref(new Map())
const selectedHouse  = ref(null)
const selectedCluster = ref(null)
const mapContainer  = ref(null)
const mapContentRef = ref(null)
const isMapVisualReady = ref(false)
const colorMode     = ref(null)
const viewMode      = ref('points')   // 'points' | 'villages'
const analyticsPanelOpen = ref(false)
const isFullscreen = ref(false)
const districtOptions = ref([])
const isDistrictLoading = ref(true)
const talukaOptions = ref([])
const villageOptions = ref([])
const selectedDistrict = ref(null)
const selectedTaluka = ref(null)
const selectedVillage = ref(null)
const showAnomalies        = ref(false)
const anomalyDrawerOpen    = ref(false)   // panel visible/hidden
const alpCollapsed         = ref(false)   // panel body collapsed (header-only mode)
const selectedAnomalyId    = ref(null)
const alpListRef           = ref(null)    // ref to the scrollable anomaly list

// ── Custom dropdown state ─────────────────────────────────────────────────────
const openDropdown = ref(null)

const populationFilters = ['population_density', 'bpl_status', 'divyang_presence', 'employment_status']

const colorOptions = [
  { label: 'Irrigation', value: 'irrigation' },
  { label: 'Crops / Season', value: 'crops' },
  { label: 'Land Holdings', value: 'land' },
  { label: 'Population Density', value: 'population_density' },
  { label: 'BPL Status', value: 'bpl_status' },
  { label: 'Divyang Presence', value: 'divyang_presence' },
  { label: 'Employment Status', value: 'employment_status' },
]

const groupedColorOptions = computed(() => {
  const optionByValue = new Map(colorOptions.map(option => [option.value, option]))
  return [
    {
      label: 'Population',
      options: [
        optionByValue.get('population_density'),
        optionByValue.get('bpl_status'),
        optionByValue.get('divyang_presence'),
        optionByValue.get('employment_status'),
      ].filter(Boolean),
    },
    {
      label: 'Agriculture',
      options: [
        optionByValue.get('crops'),
        optionByValue.get('irrigation'),
        optionByValue.get('land'),
      ].filter(Boolean),
    },
  ]
})

function toggleDropdown(name) {
  openDropdown.value = openDropdown.value === name ? null : name
}

function closeDropdowns() {
  openDropdown.value = null
}

let cachedDistrictOptions = null
let districtOptionsRequest = null
let talukaLoadToken = 0
let villageLoadToken = 0

/** Highlight helper: district list includes "All" row with value null. */
function isDistrictRowSelected(sel, row) {
  const isAllRow = row.value === undefined || row.value === null || row.value === ''
  if (isAllRow) {
    return sel == null || sel.value === undefined || sel.value === null || sel.value === ''
  }
  return sel != null && String(sel.value) === String(row.value)
}

/** Highlight helper for taluka/village rows (no "All" in list). */
function isGeoOptionSelected(sel, row) {
  if (row == null || row.value === undefined || row.value === null || row.value === '') return false
  return sel != null && String(sel.value) === String(row.value)
}

// Selection handlers — store full { label, value }; watchers cascade reset + API refetch
function selectDistrict(option) {
  selectedDistrict.value = option
    ? { label: String(option.label ?? ''), value: option.value }
    : null
  closeDropdowns()
}

function selectTaluka(option) {
  selectedTaluka.value = option
    ? { label: String(option.label ?? ''), value: option.value }
    : null
  closeDropdowns()
}

function selectVillage(option) {
  selectedVillage.value = option
    ? { label: String(option.label ?? ''), value: option.value }
    : null
  closeDropdowns()
}

const COLOR_MODE_LABELS_MAP = {
  irrigation: 'Irrigation',
  crops: 'Crops / Season',
  land: 'Land Holdings',
  population_density: 'Population Density',
  bpl_status: 'BPL Status',
  divyang_presence: 'Divyang Presence',
  employment_status: 'Employment Status',
}

function selectColorMode(mode) {
  colorMode.value = mode
  closeDropdowns()
}

// Human-readable labels shown in the trigger button
const selectedDistrictLabel = computed(() => selectedDistrict.value?.label || 'All')
const selectedTalukaLabel = computed(() => selectedTaluka.value?.label || 'All')
const selectedVillageLabel = computed(() => selectedVillage.value?.label || 'All')
const selectedColorModeLabel = computed(() => COLOR_MODE_LABELS_MAP[colorMode.value] || '')
const isPopulationMode = computed(() => populationFilters.includes(colorMode.value))

function normalizeText(value) {
  return String(value ?? '').trim().toLowerCase()
}

function isYesValue(value) {
  const normalized = normalizeText(value)
  return normalized === 'yes' || normalized === 'y' || normalized === 'true' || normalized === '1'
}

function getHouseHeadName(house) {
  const first = String(house?.FIRST_NAME_HOUSEHOLD_HEAD || house?.first_name_household_head || '').trim()
  const middle = String(house?.MIDDLE_NAME_HOUSEHOLD_HEAD || house?.middle_name_household_head || '').trim()
  const last = String(house?.LAST_NAME_HOUSEHOLD_HEAD || house?.last_name_household_head || '').trim()
  return `${first} ${middle} ${last}`.replace(/\s+/g, ' ').trim()
}

function getHouseNumber(house) {
  return house?.houseNo || house?.HOUSE_NO || house?.house_no || ''
}

function getPopulationFallbackForHouse(house) {
  const key = normalizeId(
    house?.EXTERNAL_FAMILY_ID ??
    house?.external_family_id ??
    house?.externalFamilyId
  )
  if (!key) return null
  return populationStatsByFamily.value.get(key) || null
}

function getTotalMembers(house) {
  const direct = toFiniteNumber(house?.totalMembers ?? house?.total_members ?? house?.member_count ?? house?.TOTAL_MEMBERS)
  if (direct !== null) return direct
  const fallback = getPopulationFallbackForHouse(house)
  return Number(fallback?.total_members || 0)
}

function getMaleMembers(house) {
  const direct = toFiniteNumber(house?.maleMembers ?? house?.male_members ?? house?.male_count ?? house?.male)
  if (direct !== null) return direct
  const fallback = getPopulationFallbackForHouse(house)
  return Number(fallback?.male_count || 0)
}

function getFemaleMembers(house) {
  const direct = toFiniteNumber(house?.femaleMembers ?? house?.female_members ?? house?.female_count ?? house?.female)
  if (direct !== null) return direct
  const fallback = getPopulationFallbackForHouse(house)
  return Number(fallback?.female_count || 0)
}

function getWorkingMembers(house) {
  const direct = toFiniteNumber(house?.working_members ?? house?.workingMembers)
  if (direct !== null) return direct
  const fallback = getPopulationFallbackForHouse(house)
  return Number(fallback?.working_members || 0)
}

function getWorkingOccupations(house) {
  const direct = String(house?.working_occupations || house?.occupation_list || '').trim()
  if (direct) return direct
  const fallback = getPopulationFallbackForHouse(house)
  return String(fallback?.working_occupations || fallback?.occupation_list || '').trim() || 'N/A'
}

function getBplStatus(house) {
  const bpl = normalizeText(house?.FAMILY_BELONG_BPL_CATEGORY || house?.familyBelongBplCategory)
  if (bpl.includes('non-bpl') || bpl === 'no' || bpl === 'apl' || bpl.includes('above poverty')) return 'no'
  if (bpl.includes('bpl') || bpl === 'yes') return 'yes'

  const ration = normalizeText(house?.rationCard || house?.ration_card_type || house?.RATION_CARD_TYPE)
  if (ration.includes('bpl') || ration.includes('antyodaya') || ration.includes('aay')) return 'yes'
  return 'no'
}

function hasDivyangPresence(house) {
  if (isYesValue(house?.DIVYANG || house?.divyang)) return true
  if (Number(house?.has_disability || 0) === 1) return true
  if (Number(house?.divyang_members || 0) > 0) return true
  return false
}

function getOccupationValues(house) {
  const invalidOccupationValues = new Set([
    '',
    'N/A',
    'NA',
    'NOT APPLICABLE',
    'UNEMPLOYED',
    'NOT WORKING',
    'NO WORK',
    'HOUSEWIFE',
    'HOMEMAKER',
    'STUDYING',
  ])
  const isValidOccupation = (value) => {
    const normalized = String(value ?? '').trim().toUpperCase()
    return normalized !== '' && !invalidOccupationValues.has(normalized)
  }

  if (Array.isArray(house?.occupation_list_array) && house.occupation_list_array.length) {
    return house.occupation_list_array
      .map(value => String(value).trim())
      .filter(isValidOccupation)
  }

  const raw = String(
    house?.OCCUPATION ||
    house?.occupation ||
    house?.occupation_list ||
    house?.working_occupations ||
    ''
  )

  if (!raw.trim()) return []
  return raw
    .split(/[|,;]+/)
    .map(value => value.trim())
    .filter(isValidOccupation)
}

function hasEmployment(house) {
  if (getWorkingMembers(house) > 0) return true
  return getOccupationValues(house).length > 0
}

function toFiniteNumber(value) {
  if (value === null || value === undefined) return null
  if (typeof value === 'string' && value.trim() === '') return null
  const num = Number(value)
  return Number.isFinite(num) ? num : null
}

const normalizeId = id => String(id ?? '').trim()

function resolveFamilyId(record) {
  const resolved =
    record?.EXTERNAL_FAMILY_ID ??
    record?.external_family_id ??
    record?.externalFamilyId ??
    record?.family_id ??
    record?.FAMILY_ID ??
    record?.familyId ??
    ''
  return normalizeId(resolved)
}

function groupMembersByFamily(memberRows = []) {
  const membersByFamily = {}

  memberRows.forEach(member => {
    const key = normalizeId(member?.EXTERNAL_FAMILY_ID ?? member?.external_family_id ?? member?.family_id ?? member?.FAMILY_ID)
    if (!key) return

    if (!membersByFamily[key]) {
      membersByFamily[key] = []
    }
    membersByFamily[key].push(member)
  })

  return membersByFamily
}

function extractFamiliesFromResponse(pageResponse) {
  const candidates = [
    pageResponse?.data?.families,
    pageResponse?.data?.familyData,
    pageResponse?.families,
    pageResponse?.familyData,
    pageResponse?.data,
  ]

  for (const candidate of candidates) {
    if (Array.isArray(candidate)) return candidate
  }

  return []
}

function collectFamilyMembers(pageResponse, familyRows) {
  const allMembers = []

  const topLevelCandidates = [
    pageResponse?.data?.members,
    pageResponse?.data?.familyMembers,
    pageResponse?.data?.family_members,
    pageResponse?.data?.familyMemberData,
    pageResponse?.familyMembers,
    pageResponse?.family_members,
    pageResponse?.familyMemberData,
    pageResponse?.members,
    pageResponse?.FAMILY_MEMBER,
  ]

  topLevelCandidates.forEach(candidate => {
    if (Array.isArray(candidate)) {
      allMembers.push(...candidate)
    }
  })

  familyRows.forEach(family => {
    const nestedCandidates = [
      family?.members,
      family?.familyMembers,
      family?.family_members,
      family?.FAMILY_MEMBER,
    ]

    nestedCandidates.forEach(candidate => {
      if (Array.isArray(candidate)) {
        allMembers.push(...candidate.map(member => ({
          ...member,
          EXTERNAL_FAMILY_ID:
            member?.EXTERNAL_FAMILY_ID ||
            member?.external_family_id ||
            family?.EXTERNAL_FAMILY_ID ||
            family?.external_family_id,
        })))
      }
    })
  })

  return allMembers
}

function getPopulationStats(family, familyMembers) {
  const familyExternalId = normalizeId(family?.EXTERNAL_FAMILY_ID ?? family?.external_family_id ?? family?.externalFamilyId)

  const members = familyMembers.filter(
    member => normalizeId(member?.EXTERNAL_FAMILY_ID ?? member?.external_family_id) === familyExternalId
  )

  if (members.length > 0) {
    return {
      hasData: true,
      total_members: members.length,
      male_count: members.filter(m => normalizeText(m?.GENDER || m?.gender) === 'male').length,
      female_count: members.filter(m => normalizeText(m?.GENDER || m?.gender) === 'female').length,
      divyang_members: members.filter(m => isYesValue(m?.DIVYANG || m?.divyang)).length,
      working_members: members.filter(m => String(m?.OCCUPATION || m?.occupation || '').trim() !== '').length,
    }
  }

  const fallback = populationStatsByFamily.value.get(familyExternalId)
  if (fallback) {
    return {
      hasData: true,
      total_members: Number(fallback.total_members || 0),
      male_count: Number(fallback.male_count || 0),
      female_count: Number(fallback.female_count || 0),
      divyang_members: Number(fallback.divyang_members || 0),
      working_members: Number(fallback.working_members || 0),
    }
  }

  return {
    hasData: false,
    total_members: 0,
    male_count: 0,
    female_count: 0,
    divyang_members: 0,
    working_members: 0,
  }
}

function getHouseholdHeadLabel(family) {
  const first = String(family?.FIRST_NAME_HOUSEHOLD_HEAD || family?.first_name_household_head || '').trim()
  const middle = String(family?.MIDDLE_NAME_HOUSEHOLD_HEAD || family?.middle_name_household_head || '').trim()
  const last = String(family?.LAST_NAME_HOUSEHOLD_HEAD || family?.last_name_household_head || '').trim()
  const fullName = `${first} ${middle} ${last}`.replace(/\s+/g, ' ').trim()
  return fullName || family?.headName || family?.head_name || ''
}

function buildPopulationSignature(headName, lat, lng) {
  const normalizedHead = normalizeText(headName)
  const latNum = Number(lat)
  const lngNum = Number(lng)
  if (!normalizedHead || !Number.isFinite(latNum) || !Number.isFinite(lngNum)) return ''
  return `${normalizedHead}|${latNum.toFixed(6)}|${lngNum.toFixed(6)}`
}

function buildHouseVillageKey(houseNo, villageId) {
  const normalizedHouseNo = normalizeText(houseNo)
  const normalizedVillageId = normalizeId(villageId)
  if (!normalizedHouseNo || !normalizedVillageId) return ''
  return `${normalizedVillageId}|${normalizedHouseNo}`
}

function inferExternalFamilyId(family) {
  const existing = normalizeId(family?.EXTERNAL_FAMILY_ID ?? family?.external_family_id ?? family?.externalFamilyId)
  if (existing) return existing

  const signature = buildPopulationSignature(
    getHouseholdHeadLabel(family),
    family?.LATITUDE ?? family?.latitude ?? family?.lat,
    family?.LONGITUDE ?? family?.longitude ?? family?.lng,
  )

  const matchedBySignature = signature ? populationRowsBySignature.value.get(signature) : null
  if (matchedBySignature) {
    return normalizeId(matchedBySignature?.external_family_id ?? matchedBySignature?.EXTERNAL_FAMILY_ID)
  }

  const houseVillageKey = buildHouseVillageKey(
    family?.HOUSE_NO ?? family?.houseNo ?? family?.house_no,
    family?.VILLAGE_ID ?? family?.villageId ?? family?.village_id,
  )
  if (!houseVillageKey) return ''

  const matchedByHouseVillage = populationRowsByHouseVillage.value.get(houseVillageKey)
  return normalizeId(matchedByHouseVillage?.external_family_id ?? matchedByHouseVillage?.EXTERNAL_FAMILY_ID)
}

function enrichHouseholdForPopulation(family, familyMembers) {
  // Wall time for all households is aggregated as console.time("enrichLoop") in fetchAllHouses()
  const familyId = resolveFamilyId(family)
  const inferredExternalFamilyId = inferExternalFamilyId(family)
  const stats = getPopulationStats(
    {
      ...family,
      EXTERNAL_FAMILY_ID: family?.EXTERNAL_FAMILY_ID || family?.externalFamilyId || inferredExternalFamilyId,
      external_family_id: family?.external_family_id || family?.externalFamilyId || inferredExternalFamilyId,
    },
    familyMembers
  )
  const hasMemberRows = stats.hasData

  const existingTotal = toFiniteNumber(family?.total_members ?? family?.totalMembers)
  const existingMale = toFiniteNumber(family?.male_count ?? family?.male_members ?? family?.maleMembers)
  const existingFemale = toFiniteNumber(family?.female_count ?? family?.female_members ?? family?.femaleMembers)
  const existingDivyang = toFiniteNumber(family?.divyang_members)
  const existingWorking = toFiniteNumber(family?.working_members)

  const latitude = toFiniteNumber(family?.LATITUDE ?? family?.lat ?? family?.latitude)
  const longitude = toFiniteNumber(family?.LONGITUDE ?? family?.lng ?? family?.longitude)

  const normalizedHouseNo = String(family?.HOUSE_NO || family?.houseNo || family?.house_no || '').trim()
  const normalizedHeadName = getHouseholdHeadLabel(family)
  const externalFamilyId = normalizeId(family?.EXTERNAL_FAMILY_ID ?? family?.external_family_id ?? family?.externalFamilyId ?? inferredExternalFamilyId)
  const fallback = populationStatsByFamily.value.get(externalFamilyId)

  return {
    ...family,
    family_id: familyId || String(family?.familyId || family?.FAMILY_ID || ''),
    EXTERNAL_FAMILY_ID: externalFamilyId,
    head_name: normalizedHeadName,
    headName: family?.headName || normalizedHeadName,
    house_no: normalizedHouseNo,
    houseNo: family?.houseNo || normalizedHouseNo,
    lat: latitude ?? family?.lat,
    lng: longitude ?? family?.lng,
    latitude: latitude ?? family?.latitude,
    longitude: longitude ?? family?.longitude,
    total_members: hasMemberRows ? stats.total_members : (existingTotal ?? 0),
    totalMembers: hasMemberRows ? stats.total_members : (existingTotal ?? 0),
    male_count: hasMemberRows ? stats.male_count : (existingMale ?? 0),
    male_members: hasMemberRows ? stats.male_count : (existingMale ?? 0),
    maleMembers: hasMemberRows ? stats.male_count : (existingMale ?? 0),
    female_count: hasMemberRows ? stats.female_count : (existingFemale ?? 0),
    female_members: hasMemberRows ? stats.female_count : (existingFemale ?? 0),
    femaleMembers: hasMemberRows ? stats.female_count : (existingFemale ?? 0),
    divyang_members: hasMemberRows ? stats.divyang_members : (existingDivyang ?? Number(fallback?.divyang_members || 0)),
    working_members: hasMemberRows ? stats.working_members : (existingWorking ?? Number(fallback?.working_members || 0)),
    FAMILY_BELONG_BPL_CATEGORY: family?.FAMILY_BELONG_BPL_CATEGORY || fallback?.FAMILY_BELONG_BPL_CATEGORY || '',
    occupation_list: family?.occupation_list || fallback?.occupation_list || '',
    working_occupations: family?.working_occupations || fallback?.working_occupations || '',
  }
}

function extractPopulationRows(response) {
  if (Array.isArray(response)) return response
  if (Array.isArray(response?.data)) return response.data
  if (Array.isArray(response?.markers)) return response.markers
  if (Array.isArray(response?.data?.markers)) return response.data.markers
  return []
}

function buildPopulationStatsMap(rows = []) {
  const statsMap = new Map()

  rows.forEach(row => {
    const key = normalizeId(row?.external_family_id ?? row?.EXTERNAL_FAMILY_ID)
    if (!key) return

    const nextValue = {
      total_members: Number(row?.total_members || 0),
      male_count: Number((row?.male_members ?? row?.male_count) || 0),
      female_count: Number((row?.female_members ?? row?.female_count) || 0),
      divyang_members: Number(row?.divyang_members || 0),
      working_members: Number(row?.working_members || 0),
      FAMILY_BELONG_BPL_CATEGORY: row?.FAMILY_BELONG_BPL_CATEGORY || '',
      occupation_list: row?.occupation_list || '',
      working_occupations: row?.working_occupations || '',
    }

    const currentValue = statsMap.get(key)
    if (!currentValue) {
      statsMap.set(key, nextValue)
      return
    }

    // Keep the richest member stats when duplicate family rows exist.
    if (nextValue.total_members > currentValue.total_members) {
      statsMap.set(key, nextValue)
    }
  })

  return statsMap
}

function buildPopulationRowsBySignature(rows = []) {
  const signatureMap = new Map()

  rows.forEach(row => {
    const signature = buildPopulationSignature(row?.head_name, row?.lat, row?.lng)
    if (!signature) return
    signatureMap.set(signature, row)
  })

  return signatureMap
}

function buildPopulationRowsByHouseVillage(rows = []) {
  const houseVillageMap = new Map()

  rows.forEach(row => {
    const key = buildHouseVillageKey(row?.house_no, row?.village_id)
    if (!key) return
    houseVillageMap.set(key, row)
  })

  return houseVillageMap
}

async function loadFamilyMembers() {
  try {
    // Load ALL family members globally (no filters) so count computation is always complete.
    // This data is used to compute population stats during marker enrichment.
    const res = await getPopulationMapData({})
    const rows = extractPopulationRows(res)
    familyMembers.value = Array.isArray(res?.data?.members)
      ? res.data.members
      : Array.isArray(res?.members)
        ? res.members
        : []
    populationStatsByFamily.value = buildPopulationStatsMap(rows)
    populationRowsBySignature.value = buildPopulationRowsBySignature(rows)
    populationRowsByHouseVillage.value = buildPopulationRowsByHouseVillage(rows)
  } catch (error) {
    console.warn('Population member stats unavailable:', error?.message || error)
    familyMembers.value = []
    populationStatsByFamily.value = new Map()
    populationRowsBySignature.value = new Map()
    populationRowsByHouseVillage.value = new Map()
  }
}

const MAHARASHTRA_BOUNDS = L.latLngBounds(
  [15.6, 72.6],
  [22.1, 80.9]
)
const MAHARASHTRA_CENTER = [19.7515, 75.7139]
const MAHARASHTRA_INITIAL_ZOOM = 7

let map = null
const markerRefs    = []   // { marker, house }
let clusterGroup    = null // L.layerGroup for village circles
let highlightCircle = null // currently highlighted village circle
let retryTimer      = null
let fitAfterLoad    = false  // set true by applyFilters; consumed once by plotMarkers
let activeHouseLoadToken = 0
/** Profiling for Apply button only (applyFilters with autoZoomToResults true). Matched by request token. */
let applyClickProfile = null

function logApplyProfileSummary(p, t8) {
  if (!p || p.t0 == null || p.t1 == null || p.t7 == null) return
  const apiMs = p.t2 - p.t1
  const mappingMs = p.t4 - p.t3
  const enrichmentMs = p.t6 - p.t5
  const renderingMs = t8 - p.t7
  console.log('=== Apply profile summary ===')
  console.log('API time:', apiMs, 'ms')
  console.log('Enrichment time:', enrichmentMs, 'ms')
  console.log('Mapping time:', mappingMs, 'ms')
  console.log('Rendering time:', renderingMs, 'ms')
  console.log('Total:', t8 - p.t0, 'ms')
}
let districtCentroidMarkerLayer = null  // L.layerGroup for district centroid markers
const MARKER_RENDER_CHUNK_SIZE = 100
let markerRenderToken = 0
let markerRenderFrame = null

function clearDistrictCentroids() {
  if (map && districtCentroidMarkerLayer) {
    map.removeLayer(districtCentroidMarkerLayer)
  }
  districtCentroidMarkerLayer = null
}

function renderDistrictCentroids(centroidRows) {
  clearDistrictCentroids()
  districtCentroidMarkerLayer = L.layerGroup()
  console.log('District count:', Array.isArray(centroidRows) ? centroidRows.length : 0)

  if (Array.isArray(centroidRows) && centroidRows.length < 30) {
    console.warn('District count lower than expected for Maharashtra (~36):', centroidRows.length)
  }

  centroidRows.forEach((d) => {
    if (!d.lat || !d.lng) {
      return
    }

    const lat = d.lat + (Math.random() * 0.02)
    const lng = d.lng + (Math.random() * 0.02)

    const marker = L.marker([lat, lng], {
      icon: L.divIcon({
        className: 'district-marker',
        html: `<div class="marker-count">${d.count}</div>`,
        iconSize: [32, 32],
        iconAnchor: [16, 16],
      }),
    })

    marker.bindTooltip(`District ID: ${d.district_id} | Count: ${d.count}`, {
      permanent: false,
      direction: 'top',
    })

    marker.bindPopup(`District ID: ${d.district_id}<br/>Count: ${d.count}`)

    marker.addTo(districtCentroidMarkerLayer)
  })

  if (districtCentroidMarkerLayer.getLayers().length > 0) {
    districtCentroidMarkerLayer.addTo(map)
  } else {
    districtCentroidMarkerLayer = null
  }
}

async function refreshDistrictCentroids() {
  if (!map) return

  try {
    const centroidRows = await getDistrictCentroids(getActiveLocationParams())
    renderDistrictCentroids(Array.isArray(centroidRows) ? centroidRows : [])
  } catch (error) {
    console.warn('District centroids unavailable:', error?.message || error)
  }
  console.log('AFTER CENTROIDS:', houses.value.length)
}

function handleFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement
  handleMapResize()
}

function handleMapResize() {
  if (!map) return
  try {
    map.invalidateSize(false)
  } catch (e) {
    console.warn('Map resize error:', e.message)
  }
}

function ensureMapReady() {
  if (!map) return
  try {
    map.invalidateSize(false)
    if (map._canvas) map._canvas.style.cursor = 'grab'
  } catch (e) {
    console.warn('Map render recovery:', e.message)
  }
}

function getMaharashtraFitPadding() {
  const width = window.innerWidth || 0
  if (width < 1100) return [40, 40]
  if (width > 1700) return [24, 24]
  return [32, 32]
}

function fitToMaharashtra() {
  if (!map) return
  map.fitBounds(MAHARASHTRA_BOUNDS, { padding: getMaharashtraFitPadding() })
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

function clearMarkers() {
  markerRenderToken += 1
  if (markerRenderFrame !== null) {
    cancelAnimationFrame(markerRenderFrame)
    markerRenderFrame = null
  }
  markerRefs.forEach(({ marker }) => {
    if (map && map.hasLayer(marker)) map.removeLayer(marker)
  })
  markerRefs.length = 0
}

async function loadDistrictOptionsOnce() {
  if (cachedDistrictOptions) {
    districtOptions.value = [...cachedDistrictOptions]
    return
  }

  if (!districtOptionsRequest) {
    districtOptionsRequest = getDistricts()
      .then((apiResponse) => {
        const normalizedDistrictOptions = (apiResponse || []).map(d => ({
          label: d.vsDistrictName,
          value: d.pklDistrictId,
        }))
        normalizedDistrictOptions.unshift({ label: 'All', value: null })
        cachedDistrictOptions = normalizedDistrictOptions
        return normalizedDistrictOptions
      })
      .catch((error) => {
        districtOptionsRequest = null
        throw error
      })
  }

  districtOptions.value = [...await districtOptionsRequest]
}

async function loadTalukaOptionsByDistrict(districtId) {
  const requestToken = ++talukaLoadToken
  talukaOptions.value = []
  villageOptions.value = []

  if (!districtId) return

  try {
    const res = await getLocationOptions({ district_id: districtId })
    if (requestToken !== talukaLoadToken) return
    talukaOptions.value = (res?.talukas || []).map(t => ({
      label: t.name ?? t.taluka_name ?? t.TALUKA ?? '',
      // API returns LocationOption { id, name } — prefer id, not label fields.
      value: t.id ?? t.pklTalukaId ?? t.taluka_id ?? t.value,
    }))
  } catch (e) {
    if (requestToken !== talukaLoadToken) return
    console.warn('Taluka options unavailable:', e.message)
  }
}

async function loadVillageOptionsByTaluka(districtId, talukaId) {
  const requestToken = ++villageLoadToken
  villageOptions.value = []

  if (!talukaId) return

  try {
    const res = await getLocationOptions({
      district_id: districtId || undefined,
      taluka_id: talukaId,
    })
    if (requestToken !== villageLoadToken) return
    villageOptions.value = (res?.villages || []).map(v => ({
      label: v.name ?? v.village_name ?? v.VILLAGE ?? '',
      value: v.id ?? v.pklVillageId ?? v.village_id ?? v.value,
    }))
  } catch (e) {
    if (requestToken !== villageLoadToken) return
    console.warn('Village options unavailable:', e.message)
  }
}

function geoFilterParam(sel) {
  if (!sel) return undefined

  // IMPORTANT: treat null / empty as no filter (e.g. Village = "All")
  if (sel.value === null || sel.value === undefined || sel.value === '') {
    return undefined
  }

  return sel.value
}

function getHouseFilters() {
  console.log('Selected values:', selectedDistrict.value, selectedTaluka.value, selectedVillage.value)

  const district_id = geoFilterParam(selectedDistrict.value)
  const taluka_id = geoFilterParam(selectedTaluka.value)
  const village_id = geoFilterParam(selectedVillage.value)

  console.log('FINAL FILTERS:', { district_id, taluka_id, village_id })
  console.log('Filters sent:', { district_id, taluka_id, village_id })

  return {
    limit: 2000,
    district_id,
    taluka_id,
    village_id,
  }
}

function getActiveLocationParams() {
  return {
    district_id: geoFilterParam(selectedDistrict.value),
    taluka_id: geoFilterParam(selectedTaluka.value),
    village_id: geoFilterParam(selectedVillage.value),
  }
}

async function fetchAllHouses(requestToken = activeHouseLoadToken) {
  const profile =
    applyClickProfile &&
    applyClickProfile.token === requestToken

  if (profile) {
    applyClickProfile.t1 = performance.now()
    console.log('Before API call')
  }

  // Single filtered fetch per apply/reset cycle to keep marker rendering responsive.
  const hasLocationFilter = Boolean(
    geoFilterParam(selectedDistrict.value) ||
      geoFilterParam(selectedTaluka.value) ||
      geoFilterParam(selectedVillage.value),
  )
  const base = getHouseFilters()
  const pageLimit = hasLocationFilter ? 500 : 2000
  const apiParams = { ...base, page: 1, limit: pageLimit }
  const cleanedEntries = Object.entries(apiParams).filter(([, value]) => {
    if (value === undefined || value === null) return false
    if (typeof value === 'string' && value.trim() === '') return false
    return true
  })
  const queryString = new URLSearchParams(cleanedEntries).toString()
  console.log('API CALL:', queryString ? `/houses?${queryString}` : '/houses')
  const res = await getHouses(apiParams)

  if (profile) {
    applyClickProfile.t2 = performance.now()
    console.log('API response time:', applyClickProfile.t2 - applyClickProfile.t1, 'ms')
  }

  if (profile) {
    applyClickProfile.t3 = performance.now()
    console.log('Before marker mapping')
  }
  const families = extractFamiliesFromResponse(res)
  const houseResponseMembers = collectFamilyMembers(res, families)
  const sourceMembers = houseResponseMembers.length ? houseResponseMembers : familyMembers.value
  if (profile) {
    applyClickProfile.t4 = performance.now()
    console.log('Mapping time:', applyClickProfile.t4 - applyClickProfile.t3, 'ms')
  }

  if (profile) {
    applyClickProfile.t5 = performance.now()
    console.log('Before enrichment')
    console.time('enrichLoop')
  }
  const enriched = families.map(family => enrichHouseholdForPopulation(family, sourceMembers))
  if (profile) console.timeEnd('enrichLoop')
  if (profile) {
    applyClickProfile.t6 = performance.now()
    console.log('Enrichment time:', applyClickProfile.t6 - applyClickProfile.t5, 'ms')
  }

  return enriched
}

function applyFilters(autoZoomToResults = true) {
  console.log('APPLY CLICKED')

  hasAppliedFilters.value = true

  clearRetryTimer()
  const requestToken = ++activeHouseLoadToken

  const useProfile = autoZoomToResults === true
  if (useProfile) {
    const t0 = performance.now()
    setTimeout(() => console.log('UI free'), 0)
    applyClickProfile = { t0, token: requestToken }
  } else {
    applyClickProfile = null
  }

  const district_id = geoFilterParam(selectedDistrict.value)
  if (!district_id) {
    console.log('No district selected, skipping API')
    loading.value = false
    fitAfterLoad = false
    houses.value = []
    selectedHouse.value = null
    selectedCluster.value = null
    clearClusterSelection()
    if (clusterGroup) { clusterGroup.remove(); clusterGroup = null }
    clearMarkers()
    refreshDistrictCentroids()
    console.log('AFTER APPLY:', houses.value.length)
    return
  }

  if (!map) {
    loading.value = false
    houses.value = []
    selectedHouse.value = null
    selectedCluster.value = null
    clearClusterSelection()
    if (clusterGroup) { clusterGroup.remove(); clusterGroup = null }
    clearMarkers()
    console.log('AFTER APPLY:', houses.value.length)
    return
  }

  // House fetch path: set loading before clearing houses so empty-state never flashes
  loading.value = true
  houses.value = []
  selectedHouse.value = null
  selectedCluster.value = null
  clearClusterSelection()
  if (clusterGroup) { clusterGroup.remove(); clusterGroup = null }
  clearMarkers()

  fitAfterLoad = useProfile
  loadLiveHouseData(requestToken)
  if (
    geoFilterParam(selectedDistrict.value) ||
    geoFilterParam(selectedTaluka.value) ||
    geoFilterParam(selectedVillage.value)
  ) {
    clearDistrictCentroids()
  } else {
    refreshDistrictCentroids()
  }
  console.log('AFTER APPLY:', houses.value.length)
}

async function resetFilters() {
  clearRetryTimer()
  ++activeHouseLoadToken
  hasAppliedFilters.value = false
  selectedDistrict.value = null
  selectedTaluka.value = null
  selectedVillage.value = null
  talukaOptions.value = []
  villageOptions.value = []
  fitAfterLoad = false          // reset will fly back to Maharashtra, not fitBounds
  houses.value = []
  selectedHouse.value = null
  clearMarkers()
  loading.value = false
  if (map) {
    refreshDistrictCentroids()
    fitToMaharashtra()
  }
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

const analyticsChart = computed(() => {
  const rows = houses.value
  if (!rows.length || !stats.value) return null

  const mode = colorMode.value
  if (mode === 'population_density') {
    const male = rows.reduce((sum, house) => sum + getMaleMembers(house), 0)
    const female = rows.reduce((sum, house) => sum + getFemaleMembers(house), 0)
    const total = male + female
    return {
      title: 'Gender Distribution',
      subtitle: 'Population gender split',
      totalLabel: `${total.toLocaleString()} members`,
      centerLabel: 'Population',
      centerValue: total.toLocaleString(),
      segments: [
        { label: 'Male', value: male, color: '#2563eb' },
        { label: 'Female', value: female, color: '#ec4899' },
      ],
    }
  }

  if (mode === 'bpl_status') {
    const bpl = rows.filter(house => getBplStatus(house) === 'yes').length
    const nonBpl = Math.max(rows.length - bpl, 0)
    return {
      title: 'BPL Distribution',
      subtitle: 'Household economic category',
      totalLabel: `${rows.length.toLocaleString()} households`,
      centerLabel: 'Households',
      centerValue: rows.length.toLocaleString(),
      segments: [
        { label: 'BPL', value: bpl, color: '#ef4444' },
        { label: 'Non-BPL', value: nonBpl, color: '#16a34a' },
      ],
    }
  }

  if (mode === 'divyang_presence') {
    const divyang = rows.filter(house => hasDivyangPresence(house)).length
    const nonDivyang = Math.max(rows.length - divyang, 0)
    return {
      title: 'Divyang Distribution',
      subtitle: 'Household disability presence',
      totalLabel: `${rows.length.toLocaleString()} households`,
      centerLabel: 'Households',
      centerValue: rows.length.toLocaleString(),
      segments: [
        { label: 'Divyang', value: divyang, color: '#a855f7' },
        { label: 'Non-divyang', value: nonDivyang, color: '#9ca3af' },
      ],
    }
  }

  if (mode === 'employment_status' || mode === 'employment') {
    const working = rows.filter(house => hasEmployment(house)).length
    const nonWorking = Math.max(rows.length - working, 0)
    return {
      title: 'Employment Distribution',
      subtitle: 'Household occupation status',
      totalLabel: `${rows.length.toLocaleString()} households`,
      centerLabel: 'Households',
      centerValue: rows.length.toLocaleString(),
      segments: [
        { label: 'Working', value: working, color: '#f59e0b' },
        { label: 'Non-working', value: nonWorking, color: '#9ca3af' },
      ],
    }
  }

  const total = stats.value.total || 1
  const landless = houses.value.filter(h => (parseFloat(h.totalLand) || 0) <= 1).length
  const small = houses.value.filter(h => {
    const land = parseFloat(h.totalLand) || 0
    return land > 1 && land <= 2.5
  }).length
  const mediumLarge = Math.max(stats.value.farmers - landless - small, 0)

  if (mode === 'crop_type' || mode === 'crops') {
    return {
      title: 'Crop Distribution',
      subtitle: 'Kharif and rabi participation',
      totalLabel: `${stats.value.farmers.toLocaleString()} farmers`,
      centerLabel: 'Active',
      centerValue: `${(stats.value.kharif + stats.value.rabi).toLocaleString()}`,
      segments: [
        { label: 'Kharif', value: stats.value.kharif, color: '#f59e0b' },
        { label: 'Rabi', value: stats.value.rabi, color: '#38bdf8' },
      ],
    }
  }

  if (mode === 'irrigation') {
    const irrigated = Math.max(total - stats.value.noIrrigation, 0)
    return {
      title: 'Irrigation Distribution',
      subtitle: 'Household irrigation coverage',
      totalLabel: `${stats.value.total.toLocaleString()} HH`,
      centerLabel: 'Irrigated',
      centerValue: `${irrigated.toLocaleString()}`,
      segments: [
        { label: 'Irrigated', value: irrigated, color: '#22c55e' },
        { label: 'No Irrigation', value: stats.value.noIrrigation, color: '#ef4444' },
      ],
    }
  }

  if (mode === 'land_holding' || mode === 'land') {
    return {
      title: 'Land Holding Distribution',
      subtitle: 'Agriculture footprint by holding size',
      totalLabel: `${stats.value.farmers.toLocaleString()} farmers`,
      centerLabel: 'Holding',
      centerValue: `${stats.value.farmers.toLocaleString()}`,
      segments: [
        { label: 'Marginal', value: landless, color: '#fb7185' },
        { label: 'Small', value: small, color: '#eab308' },
        { label: 'Medium/Large', value: mediumLarge, color: '#14b8a6' },
      ],
    }
  }

  return null
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
  if (!lat || lat === 'no latrine' || lat === 'none') return '#a855f7'
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
  // GPS anomaly override — red, clearly distinct from sanitation purple
  if (showAnomalies.value && anomalyFamilyIdSet.value.has(house.familyId)) return '#ef4444'

  if (colorMode.value === 'population_density') {
    const members = getTotalMembers(house)
    if (members <= 2) return '#a7f3d0'
    if (members <= 5) return '#34d399'
    return '#047857'
  }

  if (colorMode.value === 'bpl_status') {
    return getBplStatus(house) === 'yes' ? '#ef4444' : '#16a34a'
  }

  if (colorMode.value === 'divyang_presence') {
    return hasDivyangPresence(house) ? '#a855f7' : '#9ca3af'
  }

  if (colorMode.value === 'employment_status' || colorMode.value === 'employment') {
    return hasEmployment(house) ? '#f59e0b' : '#9ca3af'
  }

  if (colorMode.value === 'crops') {
    const k = (house.kharif || '').toLowerCase() === 'yes'
    const r = (house.rabi   || '').toLowerCase() === 'yes'
    if (k && r)  return '#10b981'
    if (k)       return '#f59e0b'
    if (r)       return '#38bdf8'
    return '#64748b'
  }
  if (colorMode.value === 'land') {
    const acres =
      toFiniteNumber(house?.AREA_AGRICULTURE_LAND_ACRES) ??
      toFiniteNumber(house?.area_agriculture_land_acres) ??
      toFiniteNumber(house?.LAND_UNDER_CULTIVATION_ACRES) ??
      toFiniteNumber(house?.land_under_cultivation_acres) ??
      toFiniteNumber(house?.totalLand)
    if (!acres || Number.isNaN(acres)) return '#9ca3af'
    if (acres <= 1)    return '#ef4444'
    if (acres <= 2.5)  return '#f59e0b'
    if (acres <= 5)    return '#22c55e'
    return '#10b981'
  }
  if (colorMode.value === 'irrigation') {
    const irrigation = String(house.SOURCE_WATER_IRRIGATION || house.waterSource || '').trim()
    if (irrigation && irrigation !== 'No') return '#16a34a'
    return '#ef4444'
  }

  // sanitation fallback for other legacy modes
  const lat   = (house.latrine  || '').toLowerCase()
  const light = (house.lighting || '').toLowerCase()
  const hasToilet = lat   && lat   !== 'no latrine' && lat   !== 'none'
  const hasElec   = light && light !== 'kerosene'   && light !== 'none'
  if (!hasToilet && !hasElec) return '#a855f7'
  if (!hasToilet || !hasElec) return '#f59e0b'
  return '#22c55e'
}

const headerLegend = computed(() => {
  let entries
  if (colorMode.value === 'population_density') {
    entries = [
      { color: '#a7f3d0', label: '1-2 members' },
      { color: '#34d399', label: '3-5 members' },
      { color: '#047857', label: '6+ members' },
    ]
  } else if (colorMode.value === 'bpl_status') {
    entries = [
      { color: '#ef4444', label: 'BPL households' },
      { color: '#16a34a', label: 'Non-BPL households' },
    ]
  } else if (colorMode.value === 'divyang_presence') {
    entries = [
      { color: '#a855f7', label: 'Divyang present' },
      { color: '#9ca3af', label: 'No divyang' },
    ]
  } else if (colorMode.value === 'employment_status' || colorMode.value === 'employment') {
    entries = [
      { color: '#f59e0b', label: 'Working households' },
      { color: '#9ca3af', label: 'Non-working households' },
    ]
  } else if (colorMode.value === 'crops') {
    entries = [
      { color: '#10b981', label: 'Both Seasons' },
      { color: '#f59e0b', label: 'Kharif Only' },
      { color: '#38bdf8', label: 'Rabi Only' },
      { color: '#64748b', label: 'No Crops' },
    ]
  } else if (colorMode.value === 'land') {
    entries = [
      { color: '#9ca3af', label: 'Data not available' },
      { color: '#10b981', label: 'Large >5ac' },
      { color: '#22c55e', label: 'Medium 2.5-5ac' },
      { color: '#f59e0b', label: 'Small 1-2.5ac' },
      { color: '#ef4444', label: 'Marginal ≤1ac' },
    ]
  } else if (colorMode.value === 'irrigation') {
    entries = [
      { color: '#16a34a', label: 'Irrigation available' },
      { color: '#ef4444', label: 'No irrigation' },
    ]
  } else {
    entries = [
      { color: '#a855f7', label: 'No Sanitation' },
      { color: '#f59e0b', label: 'Partial' },
      { color: '#22c55e', label: 'Good' },
    ]
  }
  // Append GPS Mismatch entry whenever anomaly detection is active
  if (showAnomalies.value && anomalies.value.length) {
    entries = [...entries, { color: '#ef4444', label: 'GPS Mismatch' }]
  }
  return entries
})

watch(colorMode, () => {
  markerRefs.forEach(({ marker, house }) => {
    marker.setStyle({ fillColor: getMarkerColor(house) })
  })
})

// Recolor all dots whenever anomaly detection is toggled
watch(showAnomalies, () => {
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
    { label: 'No Sanitation', pct: pct(cl.noToilet, cl.count), color: '#a855f7' },
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
  // Ensure map renders properly after layer changes
  await nextTick()
  setTimeout(() => handleMapResize(), 50)
}

// ── Coordinate Anomaly Detection ─────────────────────────────────────────────

/**
 * Haversine great-circle distance between two lat/lng points.
 * Returns distance in kilometres.
 */
function haversineKm(lat1, lng1, lat2, lng2) {
  const R = 6371 // Earth radius in km
  const dLat = (lat2 - lat1) * Math.PI / 180
  const dLng = (lng2 - lng1) * Math.PI / 180
  const a =
    Math.sin(dLat / 2) ** 2 +
    Math.cos(lat1 * Math.PI / 180) *
    Math.cos(lat2 * Math.PI / 180) *
    Math.sin(dLng / 2) ** 2
  return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
}

/**
 * Statistical outlier detection using centroid + z-score distance.
 *
 * Algorithm:
 *  1. Group all households by villageId.
 *  2. Compute each village's centroid (mean lat/lng of all members).
 *  3. Compute every household's Haversine distance to its village centroid.
 *  4. Compute mean + std-dev of those distances per village.
 *  5. Flag a household as an outlier when:
 *       distance > (mean + SIGMA_THRESHOLD * stdDev)
 *       AND distance > ABSOLUTE_MIN_KM  (absolute floor)
 *
 * Returns an enriched array of outlier house objects with extra _* keys.
 */
const SIGMA_THRESHOLD = 2.5   // z-score cut-off
const ABSOLUTE_MIN_KM = 5     // never flag closer than 5 km regardless of stats
const MIN_GROUP_SIZE  = 3     // need at least 3 points to run statistics

const anomalies = computed(() => {
  const rows = houses.value.filter(
    h => h.villageId && typeof h.latitude === 'number' && typeof h.longitude === 'number'
  )
  if (!rows.length) return []

  // 1. Group by village
  const byVillage = new Map()
  for (const h of rows) {
    const key = String(h.villageId)
    if (!byVillage.has(key)) byVillage.set(key, [])
    byVillage.get(key).push(h)
  }

  // Build village centroid lookup: villageId → { name, lat, lng }
  const villageCentroids = []
  for (const [vid, group] of byVillage) {
    if (group.length < MIN_GROUP_SIZE) continue
    const lat = group.reduce((s, h) => s + h.latitude,  0) / group.length
    const lng = group.reduce((s, h) => s + h.longitude, 0) / group.length
    villageCentroids.push({ vid, name: group[0].villageName || `Village ${vid}`, lat, lng })
  }

  const found = []

  for (const [, group] of byVillage) {
    if (group.length < MIN_GROUP_SIZE) continue

    // 2. Centroid
    const centLat = group.reduce((s, h) => s + h.latitude,  0) / group.length
    const centLng = group.reduce((s, h) => s + h.longitude, 0) / group.length

    // 3. Distances
    const distances = group.map(h => haversineKm(h.latitude, h.longitude, centLat, centLng))

    // 4. Mean + std-dev
    const mean = distances.reduce((s, d) => s + d, 0) / distances.length
    const variance = distances.reduce((s, d) => s + (d - mean) ** 2, 0) / distances.length
    const stdDev = Math.sqrt(variance)
    const threshold = Math.max(mean + SIGMA_THRESHOLD * stdDev, ABSOLUTE_MIN_KM)

    // 5. Flag outliers
    group.forEach((h, i) => {
      if (distances[i] > threshold) {
        // Find the nearest *other* village centroid to the actual GPS point
        let nearestName = null
        let nearestDist = Infinity
        for (const vc of villageCentroids) {
          if (vc.vid === String(h.villageId)) continue
          const d = haversineKm(h.latitude, h.longitude, vc.lat, vc.lng)
          if (d < nearestDist) { nearestDist = d; nearestName = vc.name }
        }

        found.push({
          ...h,
          _distanceKm:       +distances[i].toFixed(1),
          _centroidLat:      centLat,
          _centroidLng:      centLng,
          _threshold:        +threshold.toFixed(1),
          _plottedVillage:   nearestName || null,
        })
      }
    })
  }

  return found
})

// O(1) anomaly membership lookup — rebuilt whenever anomalies changes
const anomalyFamilyIdSet = computed(() => new Set(anomalies.value.map(h => h.familyId)))

/**
 * Recolor existing household dots to highlight GPS anomalies.
 * No separate Leaflet layer is needed — we tint the existing circleMarkers.
 * Anomalous dots turn #ef4444 (red); all others restore their normal color.
 */
function buildAnomalyLayer() {
  markerRefs.forEach(({ marker, house }) => {
    marker.setStyle({ fillColor: getMarkerColor(house) })
  })
}

/**
 * Scroll the anomaly list so the row with the given familyId is visible.
 * Uses nextTick so the DOM is up-to-date before querying.
 */
async function scrollAnomalyIntoView(familyId) {
  await nextTick()
  const list = alpListRef.value
  if (!list) return
  const row = list.querySelector(`[data-fid="${familyId}"]`)
  if (row) row.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
}

/** Fly to an anomaly, show its detail panel, and highlight the list row */
function flyToAnomaly(house) {
  if (!map) return
  selectedAnomalyId.value = house.familyId
  selectedHouse.value = house          // enriched with _distanceKm etc.
  scrollAnomalyIntoView(house.familyId)
  map.flyTo([house.latitude, house.longitude], 14, { duration: 1.2 })
}

// Whenever sidebar collapses/expands or opens/closes, resize the map.
// We fire immediately (for v-if removals that have no CSS transition) AND
// after the CSS width transition (240ms) so the map fills the reclaimed space.
function resizeMapAfterTransition() {
  if (map) map.invalidateSize({ animate: false })
  setTimeout(() => { if (map) map.invalidateSize({ animate: false }) }, 60)
  setTimeout(() => { if (map) map.invalidateSize({ animate: false }) }, 260)
}
watch(alpCollapsed, resizeMapAfterTransition)
watch(anomalyDrawerOpen, resizeMapAfterTransition)

/** Restore all household dot colors to their normal (non-anomaly) state */
function clearAnomalyLayer() {
  markerRefs.forEach(({ marker, house }) => {
    marker.setStyle({ fillColor: getMarkerColor(house) })
  })
}

/** Toggle anomaly dot highlighting + side panel */
function toggleAnomalies() {
  showAnomalies.value = !showAnomalies.value
  if (showAnomalies.value) {
    buildAnomalyLayer()
    anomalyDrawerOpen.value = true
    alpCollapsed.value      = false   // open expanded on first activation
  } else {
    clearAnomalyLayer()
    anomalyDrawerOpen.value  = false
    alpCollapsed.value       = false
    selectedAnomalyId.value  = null
  }
}

/** Toggle the panel between collapsed strip ↔ expanded list (like 3D sidebar) */
function toggleAlpCollapse() {
  alpCollapsed.value = !alpCollapsed.value
}

/** Fully close the anomaly sidebar and reset state */
async function closeAnomalySidebar() {
  anomalyDrawerOpen.value  = false
  showAnomalies.value      = false
  alpCollapsed.value       = false
  selectedAnomalyId.value  = null
  clearAnomalyLayer()
  // Wait for Vue to remove the sidebar element from DOM, then resize the map
  await nextTick()
  resizeMapAfterTransition()
}

// Rebuild anomaly layer whenever the underlying houses data changes
watch(houses, () => {
  if (showAnomalies.value) buildAnomalyLayer()
})

// ─────────────────────────────────────────────────────────────────────────────

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
    const res  = await fetch('https://raw.githubusercontent.com/geohacker/india/master/district/india_district.geojson')
    const data = await res.json()

    // Filter to Maharashtra only — geohacker uses ST_NM property
    const mhDistricts = data.features.filter(f => {
      const props = f.properties || {}
      return (
        String(props.ST_NM  || '').toUpperCase().includes('MAHARASHTRA') ||
        String(props.state  || '').toUpperCase().includes('MAHARASHTRA') ||
        String(props.STATE  || '').toUpperCase().includes('MAHARASHTRA') ||
        String(props.NAME_1 || '').toUpperCase().includes('MAHARASHTRA')
      )
    })

    // Note: GeoJSON data is no longer cached since district centroids are now
    // calculated from the database (FAMILY table LATITUDE/LONGITUDE columns)

    mhDistricts.forEach((feature, i) => {
      const color = DISTRICT_PALETTE[i % DISTRICT_PALETTE.length]
      const props = feature.properties || {}
      const districtName = props.DISTRICT || props.dtname || props.NAME_2 || props.district || 'District'
      const layer = L.geoJSON(feature, {
        style: {
          color,
          weight: 1.8,
          opacity: 0.75,
          fillColor: color,
          fillOpacity: 0.10,
          dashArray: null,
        },
      }).addTo(mapInstance).bringToBack()

      layer.on('mouseover', function (e) {
        e.layer.setStyle({ fillOpacity: 0.30, weight: 2.5 })
        e.layer.bindTooltip(
          `<div style="font-weight:700;font-size:0.78rem;color:#1e293b;">${districtName}</div>`,
          { sticky: true, className: 'map-tooltip district-tooltip', direction: 'top' }
        ).openTooltip(e.latlng)
      })
      layer.on('mousemove', function (e) {
        e.layer.getTooltip()?.setLatLng(e.latlng)
      })
      layer.on('mouseout', function (e) {
        e.layer.setStyle({ fillOpacity: 0.10, weight: 1.8 })
        e.layer.closeTooltip()
      })
    })
  } catch (e) {
    console.warn('District boundaries unavailable:', e.message)
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
      // ── Inverted polygon mask ──────────────────────────────────────────────
      // Outer ring: entire world. Inner ring(s): Maharashtra boundary (the "hole").
      // Leaflet renders the world black but leaves Maharashtra transparent.
      const worldRing = [[-90, -180], [90, -180], [90, 180], [-90, 180], [-90, -180]]

      const holeRings = []
      if (mh.geometry.type === 'Polygon') {
        mh.geometry.coordinates.forEach(ring => holeRings.push(ring))
      } else if (mh.geometry.type === 'MultiPolygon') {
        mh.geometry.coordinates.forEach(poly =>
          poly.forEach(ring => holeRings.push(ring))
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

      // Bright amber border tracing Maharashtra's edge
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

      // State label
      const center = L.geoJSON(mh).getBounds().getCenter()
      L.marker(center, {
        icon: L.divIcon({
          className: '',
          html: '<div class="mh-label">Maharashtra</div>',
          iconSize: [120, 24],
          iconAnchor: [60, 12],
        }),
        interactive: false,
      }).addTo(mapInstance)
    }
  } catch (e) {
    console.warn('Maharashtra boundary unavailable:', e.message)
  }
}

function addHouseMarker(house) {
  const color  = getMarkerColor(house)
  const marker = L.circleMarker([house.latitude, house.longitude], {
    radius: 5, fillColor: color, color: '#fff',
    weight: 1.5, opacity: 1, fillOpacity: 0.88,
  }).addTo(map)
  markerRefs.push({ marker, house })

  marker.on('click', (e) => {
    L.DomEvent.stopPropagation(e)

    // If anomaly mode is ON and this is a flagged red dot, enrich the
    // selectedHouse so the detail panel shows the GPS mismatch reason.
    if (showAnomalies.value && anomalyFamilyIdSet.value.has(house.familyId)) {
      const enriched = anomalies.value.find(a => a.familyId === house.familyId)
      selectedHouse.value = enriched || house
      selectedAnomalyId.value = house.familyId
      // Expand sidebar (if user had it collapsed) so the highlighted row is visible
      alpCollapsed.value = false
      scrollAnomalyIntoView(house.familyId)
    } else {
      selectedHouse.value = house
      selectedAnomalyId.value = null
    }
  })

  // Tooltip: include GPS mismatch warning if anomaly detection is active
  marker.bindTooltip(() => {
    const isAnomaly = showAnomalies.value && anomalyFamilyIdSet.value.has(house.familyId)
    if (isAnomaly) {
      const enriched = anomalies.value.find(a => a.familyId === house.familyId)
      const km = enriched?._distanceKm ?? '?'
      return `
        <strong>${house.headName || 'Household'}</strong><br/>
        <span style="color:#ef4444;font-weight:600;">⚠️ Village Mismatch</span><br/>
        <span style="color:#ef4444;font-weight:700;">${km} km</span> <span style="color:#64748b;">from registered village</span><br/>
        <span style="color:#94a3b8;font-size:0.7em;">Click to see details</span>
      `
    }

    if (populationFilters.includes(colorMode.value)) {
      return `
        <strong>${house.headName || getHouseHeadName(house) || 'Household'}</strong><br/>
        House No: ${getHouseNumber(house) || 'N/A'}<br/>
        Members: ${getTotalMembers(house)} · Male: ${getMaleMembers(house)} · Female: ${getFemaleMembers(house)}
      `
    }

    return `
      <strong>${house.headName || 'Household'}</strong><br/>
      Land: ${house.totalLand || '0'} acres · Kharif: ${house.kharif || '—'} · Rabi: ${house.rabi || '—'}
    `
  }, { className: 'map-tooltip', direction: 'top', offset: L.point(0, -6) })
}

function plotMarkers(data, profileRequestToken = null) {
  console.time('plotMarkers')
  clearMarkers()
  const renderToken = markerRenderToken
  let index = 0

  const finishPlotMarkersTiming = (withSummary) => {
    console.timeEnd('plotMarkers')
    if (
      withSummary &&
      profileRequestToken != null &&
      applyClickProfile &&
      applyClickProfile.token === profileRequestToken &&
      applyClickProfile.t7 != null
    ) {
      const t8 = performance.now()
      console.log('Rendering time:', t8 - applyClickProfile.t7, 'ms')
      console.log('Total time:', t8 - applyClickProfile.t0, 'ms')
      logApplyProfileSummary(applyClickProfile, t8)
      applyClickProfile = null
    }
  }

  const renderChunk = () => {
    if (renderToken !== markerRenderToken || !map) {
      finishPlotMarkersTiming(false)
      return
    }
    const end = Math.min(index + MARKER_RENDER_CHUNK_SIZE, data.length)
    for (let i = index; i < end; i += 1) {
      addHouseMarker(data[i])
    }
    index = end
    if (index < data.length) {
      markerRenderFrame = requestAnimationFrame(renderChunk)
    } else {
      markerRenderFrame = null
      finishPlotMarkersTiming(true)
    }
  }

  renderChunk()

  // Auto-zoom to the filtered results when triggered by applyFilters()
  if (fitAfterLoad && data.length > 0 && map) {
    fitAfterLoad = false
    const validPoints = data.filter(
      h => typeof h.latitude === 'number' && typeof h.longitude === 'number'
    )
    if (validPoints.length === 1) {
      // Single point — fly to it at street level
      map.flyTo([validPoints[0].latitude, validPoints[0].longitude], 14, { duration: 1 })
    } else if (validPoints.length > 1) {
      const bounds = L.latLngBounds(validPoints.map(h => [h.latitude, h.longitude]))
      map.flyToBounds(bounds, {
        padding: [60, 60],
        maxZoom: 14,      // never zoom past street level even for a tiny village
        duration: 1.2,
      })
    }
    // Ensure map renders properly after zoom operations
    setTimeout(() => ensureMapReady(), 100)
    setTimeout(() => ensureMapReady(), 200)
  }
}

function clearRetryTimer() {
  if (retryTimer) {
    clearTimeout(retryTimer)
    retryTimer = null
  }
}

async function loadLiveHouseData(requestToken = activeHouseLoadToken) {
  if (requestToken !== activeHouseLoadToken) return

  const profile =
    applyClickProfile &&
    applyClickProfile.token === requestToken

  try {
    const real = await fetchAllHouses(requestToken)
    if (requestToken !== activeHouseLoadToken) return

    console.log('Households count:', real.length)

    if (real.length > 0) {
      clearRetryTimer()
      houses.value = real
      console.log('AFTER LOAD:', houses.value.length)
      if (profile) {
        applyClickProfile.t7 = performance.now()
        console.log('Before rendering markers')
      }
      plotMarkers(real, profile ? requestToken : null)
      if (viewMode.value === 'villages') {
        drawClusters(buildVillageClusters(real))
      }
      loading.value = false
      return
    }

    // Successful fetch with zero households — no plotMarkers run
    console.log('AFTER LOAD:', houses.value.length)
    if (profile && applyClickProfile && applyClickProfile.token === requestToken) {
      applyClickProfile.t7 = performance.now()
      console.log('Before rendering markers')
      const t8 = performance.now()
      console.log('Rendering time:', t8 - applyClickProfile.t7, 'ms')
      console.log('Total time:', t8 - applyClickProfile.t0, 'ms')
      logApplyProfileSummary(applyClickProfile, t8)
      applyClickProfile = null
    }
    loading.value = false
    return
  } catch (e) {
    if (requestToken !== activeHouseLoadToken) return
    console.warn('Houses API not available:', e.message)
    if (applyClickProfile && applyClickProfile.token === requestToken) {
      applyClickProfile = null
    }
  }

  if (requestToken !== activeHouseLoadToken) return
  loading.value = false
}

onMounted(async () => {
  await nextTick()
  isMapVisualReady.value = false

  try {
    await loadDistrictOptionsOnce()
  } catch (e) {
    console.warn('District options unavailable:', e.message)
  } finally {
    isDistrictLoading.value = false
  }

  if (mapContainer.value) {
    // Ensure container has proper dimensions before map init
    const rect = mapContainer.value.getBoundingClientRect()
    if (rect.width < 100 || rect.height < 100) {
      console.warn('Map container too small, waiting for layout...')
      await new Promise(resolve => setTimeout(resolve, 100))
    }

    map = L.map(mapContainer.value, {
      zoomControl: false,
      doubleClickZoom: false,
      center: MAHARASHTRA_CENTER,
      zoom: MAHARASHTRA_INITIAL_ZOOM,
      minZoom: MAHARASHTRA_INITIAL_ZOOM,
      maxBounds: MAHARASHTRA_BOUNDS,
      maxBoundsViscosity: 1.0,
    })
    map.setMaxBounds(MAHARASHTRA_BOUNDS)
    map.setMinZoom(MAHARASHTRA_INITIAL_ZOOM)
    map.options.maxBoundsViscosity = 1.0
    L.control.zoom({ position: 'topleft' }).addTo(map)
    addTiles(map)

    // Ensure first visible frame is the final composed state.
    await Promise.allSettled([
      addDistrictBorders(map),
      addMaharashtraHighlight(map),
      refreshDistrictCentroids(),
    ])

    fitToMaharashtra()
    isMapVisualReady.value = true

    // More aggressive size invalidation for reliable rendering
    setTimeout(() => ensureMapReady(), 50)
    setTimeout(() => ensureMapReady(), 150)
    setTimeout(() => ensureMapReady(), 300)
    window.addEventListener('resize', handleMapResize)
    window.addEventListener('click', closeDropdowns)
    document.addEventListener('fullscreenchange', handleFullscreenChange)
  }

  loading.value = true
  await loadFamilyMembers()
  loading.value = false
})

onUnmounted(() => {
  clearRetryTimer()
  clearAnomalyLayer()
  clearDistrictCentroids()
  clearMarkers()
  window.removeEventListener('resize', handleMapResize)
  window.removeEventListener('click', closeDropdowns)
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
  if (map) { map.remove(); map = null }
})

watch(selectedDistrict, async () => {
  selectedTaluka.value = null
  selectedVillage.value = null
  await loadTalukaOptionsByDistrict(geoFilterParam(selectedDistrict.value))
})

watch(selectedTaluka, async () => {
  selectedVillage.value = null
  await loadVillageOptionsByTaluka(
    geoFilterParam(selectedDistrict.value),
    geoFilterParam(selectedTaluka.value),
  )
})

watch(selectedVillage, () => {
  // Dropdown only — no API (household load is Apply-only)
})

watch(houses, (val) => {
  console.log('HOUSES CHANGED:', val?.length)
})

watch(analyticsPanelOpen, async () => {
  await nextTick()
  handleMapResize()
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

.page-title { font-family: var(--font-display); font-size: 1.45rem; color: var(--text-primary); font-weight: 400; letter-spacing: -0.01em; }
.page-subtitle { color: var(--text-dim); font-size: 0.75rem; margin-top: 0.2rem; display: flex; align-items: center; gap: 0.5rem; }

.map-controls { display: flex; align-items: center; gap: 0.9rem; flex-wrap: wrap; }
.map-control-group { display: flex; align-items: center; gap: 0.45rem; }
.control-label { font-size: 0.63rem; text-transform: uppercase; letter-spacing: 0.07em; color: var(--text-dim); white-space: nowrap; font-weight: 600; }
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

.view-by-btn {
  border-radius: var(--radius-sm);
}

.cs-trigger-placeholder .cs-value {
  color: #9ca3af !important;
  font-style: italic;
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

/* Option group separator label */
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
  font-size: 0.7rem;
  font-weight: 600;
  padding: 0.28rem 0.75rem;
  cursor: pointer;
  font-family: var(--font-body);
  transition: background 0.14s, border-color 0.14s;
  letter-spacing: 0.01em;
}

.apply-btn:hover {
  background: #0d9488;
  border-color: #0d9488;
}

.reset-btn {
  border: 1px solid #e2e8f0;
  background: transparent;
  color: #64748b;
  border-radius: 6px;
  font-size: 0.7rem;
  font-weight: 500;
  padding: 0.28rem 0.75rem;
  cursor: pointer;
  font-family: var(--font-body);
  transition: border-color 0.14s, color 0.14s, background 0.14s;
}

.reset-btn:hover {
  background: #f8fafc;
  border-color: #94a3b8;
  color: #334155;
}

.map-legend { display: flex; gap: 0.6rem; flex-wrap: wrap; align-items: center; }
.legend-item {
  display: flex; align-items: center; gap: 0.3rem;
  font-size: 0.67rem; color: var(--text-muted);
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 999px;
  padding: 0.18rem 0.55rem 0.18rem 0.3rem;
}
.legend-dot {
  width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
  box-shadow: inset 0 0 0 1px rgba(0,0,0,0.08);
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
  pointer-events: auto;
}

.map-container {
  position: absolute;
  inset: 0;
  z-index: 1;
}

.map-container--hidden {
  opacity: 0;
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
  backdrop-filter: blur(4px);
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
  backdrop-filter: blur(4px);
}

.fullscreen-toggle:hover {
  border-color: rgba(20,184,166,0.4);
  color: var(--teal);
  background: rgba(255,255,255,0.98);
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

/* ═══════════════════════════════════════════════
   DETAIL PANEL — household click popup
═══════════════════════════════════════════════ */
.detail-panel {
  position: absolute;
  top: 1rem; right: 1rem;
  width: 296px;
  max-height: calc(100% - 2rem);
  overflow-y: auto;
  z-index: 500;
  background: #ffffff;
  border: 1px solid #e8edf3;
  border-radius: 14px;
  box-shadow: 0 4px 20px rgba(15,23,42,0.10), 0 1px 5px rgba(15,23,42,0.05);
  scrollbar-width: thin;
  scrollbar-color: #e2e8f0 transparent;
}

/* Header */
.detail-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: 0.5rem; padding: 0.85rem 0.95rem 0.75rem;
  border-bottom: 1px solid #f1f5f9;
  background: #f8fafc;
  border-radius: 14px 14px 0 0;
}
.detail-header-info { flex: 1; min-width: 0; }
.detail-badge {
  display: inline-block; padding: 0.14rem 0.48rem;
  border-radius: 20px; border: 1px solid;
  font-size: 0.58rem; font-weight: 700; letter-spacing: 0.05em;
  margin-bottom: 0.32rem; text-transform: uppercase;
}
.detail-name {
  font-size: 0.92rem; font-weight: 700; color: #0f172a;
  line-height: 1.3; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.detail-sub {
  display: flex; align-items: center; flex-wrap: wrap; gap: 0.28rem;
  font-size: 0.63rem; color: #64748b; margin-top: 0.22rem; font-weight: 500;
}
.detail-id-chip {
  background: #334155; color: #f1f5f9;
  border-radius: 4px; padding: 0.07rem 0.34rem;
  font-size: 0.58rem; font-weight: 700; letter-spacing: 0.04em;
}
.detail-close {
  background: transparent; border: 1px solid #e2e8f0;
  border-radius: 50%; color: #94a3b8;
  font-size: 1rem; line-height: 1; cursor: pointer;
  width: 24px; height: 24px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; transition: all 0.15s;
}
.detail-close:hover { background: #fee2e2; border-color: #fecaca; color: #ef4444; }

/* Section labels */
.dp-section-label {
  display: flex; align-items: center; gap: 0.35rem;
  font-size: 0.57rem; text-transform: uppercase; letter-spacing: 0.09em;
  color: #94a3b8; font-weight: 700;
  padding: 0.7rem 0.95rem 0.28rem;
  border-top: 1px solid #f8fafc;
}
.dp-section-icon { font-size: 0.8rem; }

/* Big stat row */
.dp-stat-row {
  display: grid; grid-template-columns: 1fr 1fr; gap: 0.45rem;
  padding: 0 0.95rem;
}
.dp-stat {
  background: #f9fafb; border: 1px solid #eef2f6;
  border-radius: 8px; padding: 0.5rem 0.65rem; text-align: center;
}
.dp-stat-val { font-size: 1.1rem; font-weight: 700; color: #0f172a; line-height: 1.1; }
.dp-stat-val small { font-size: 0.62rem; color: #94a3b8; font-weight: 600; }
.dp-stat-key { font-size: 0.55rem; text-transform: uppercase; letter-spacing: 0.06em; color: #b0bac6; font-weight: 600; margin-top: 0.14rem; }

/* Crop chips */
.dp-chip-row {
  display: grid; grid-template-columns: 1fr 1fr; gap: 0.45rem;
  padding: 0.45rem 0.95rem 0;
}
.dp-chip-block { display: flex; flex-direction: column; gap: 0.18rem; }
.dp-chip-label { font-size: 0.55rem; text-transform: uppercase; letter-spacing: 0.06em; color: #b0bac6; font-weight: 600; }
.dp-chip {
  padding: 0.26rem 0.5rem; border-radius: 6px;
  font-size: 0.7rem; font-weight: 600; text-align: center;
}
.dp-chip-kharif { background: #fffbeb; color: #92400e; border: 1px solid #fde68a; }
.dp-chip-rabi   { background: #eff6ff; color: #1d4ed8; border: 1px solid #bfdbfe; }

/* Field rows */
.dp-field-row {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.45rem 0.95rem;
  border-bottom: 1px solid #f8fafc;
}
.dp-field-icon { font-size: 0.82rem; flex-shrink: 0; width: 1.1rem; text-align: center; }
.dp-field-key  { font-size: 0.66rem; color: #64748b; font-weight: 500; flex: 1; }
.dp-field-val  { font-size: 0.73rem; color: #0f172a; font-weight: 600; text-align: right; max-width: 55%; }

.panel-coords {
  margin-top: 0.6rem;
  padding: 0.5rem 0.95rem;
  border-top: 1px solid #f8fafc;
  font-size: 0.63rem;
  color: #b0bac6;
  font-variant-numeric: tabular-nums;
  text-align: center;
  font-family: 'SF Mono', 'Fira Mono', monospace;
  letter-spacing: 0.02em;
}

/* Village Mismatch card inside detail panel */
.gps-section-label {
  color: #ef4444 !important;
}

.gps-mismatch-card {
  margin: 0 0.65rem 0.65rem;
  border: 1px solid rgba(239,68,68,0.25);
  border-radius: 10px;
  background: #fff8f8;
  padding: 0.65rem 0.75rem 0.6rem;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

/* Mismatch description sentence */
.gps-mismatch-headline {
  font-size: 0.7rem;
  color: #334155;
  font-weight: 500;
  line-height: 1.6;
}
.gps-village-db {
  color: #0f172a;
  font-weight: 700;
}
.gps-village-plotted {
  color: #ef4444;
  font-weight: 700;
}

/* Distance offset — prominent bold red pill */
.gps-offset-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fee2e2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  padding: 0.3rem 0.6rem;
}
.gps-offset-label {
  font-size: 0.6rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: #b91c1c;
  font-weight: 600;
}
.gps-offset-value {
  font-size: 0.82rem;
  font-weight: 800;
  color: #ef4444;
  font-variant-numeric: tabular-nums;
  font-family: 'SF Mono', 'Fira Mono', monospace;
}

/* Village comparison: Database → Plotted At */
.gps-village-compare {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 0.45rem 0.55rem;
}
.gps-vc-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 0;
}
.gps-vc-badge {
  font-size: 0.5rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-weight: 700;
  padding: 0.08rem 0.32rem;
  border-radius: 3px;
  width: fit-content;
}
.gps-vc-db .gps-vc-badge {
  background: #dcfce7;
  color: #15803d;
}
.gps-vc-plotted .gps-vc-badge {
  background: #fee2e2;
  color: #b91c1c;
}
.gps-vc-name {
  font-size: 0.72rem;
  font-weight: 700;
  color: #0f172a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.gps-vc-db .gps-vc-name { color: #15803d; }
.gps-vc-plotted .gps-vc-name { color: #ef4444; }
.gps-vc-sub {
  font-size: 0.58rem;
  color: #94a3b8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.gps-vc-arrow {
  font-size: 1rem;
  color: #cbd5e1;
  font-weight: 700;
  flex-shrink: 0;
}

/* Helpful tip — no jargon */
.gps-mismatch-tip {
  font-size: 0.65rem;
  color: #64748b;
  line-height: 1.55;
  background: #f1f5f9;
  border-radius: 6px;
  padding: 0.38rem 0.5rem;
  border-left: 3px solid #f59e0b;
}
.gps-mismatch-tip strong { color: #92400e; }

/* View mode toggle */
.view-toggle {
  display: flex;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 7px;
  overflow: hidden;
  gap: 1px;
}
.toggle-btn {
  padding: 0.26rem 0.8rem;
  font-size: 0.7rem;
  font-weight: 500;
  font-family: var(--font-body);
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  transition: background 0.13s, color 0.13s;
  white-space: nowrap;
}
.toggle-btn:hover { background: var(--bg-card); color: var(--text-body); }
.toggle-btn.active {
  background: var(--teal);
  color: #fff;
  font-weight: 600;
}

/* Village detail panel extras */
.village-panel { min-width: 270px; padding: 1rem; }
.panel-close {
  float: right;
  width: 22px; height: 22px; border-radius: 50%;
  border: 1px solid #e2e8f0; background: transparent;
  color: #94a3b8; font-size: 0.95rem; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.14s;
}
.panel-close:hover { background: #fee2e2; border-color: #fecaca; color: #ef4444; }
.panel-title {
  font-family: var(--font-display);
  font-size: 1.05rem; font-weight: 500; color: var(--text-primary);
  margin: 0.35rem 0 0.18rem; line-height: 1.3;
}
.panel-id {
  font-size: 0.68rem; color: var(--text-dim); margin-bottom: 0.85rem;
}
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
  .analytics-panel {
    width: 320px;
    max-width: 48vw;
  }

  .chart-layout {
    grid-template-columns: 96px 1fr;
  }

  .map-controls {
    gap: 0.9rem;
  }

  .map-floating-controls {
    top: 16px;
    left: 16px;
    gap: 0.45rem;
  }
}

@media (max-width: 760px) {
  .map-header,
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

  .custom-select {
    min-width: 96px;
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
/* Position zoom control below the "View Analytics / Fullscreen" buttons (top:20px + ~34px height + 8px gap) */
.leaflet-top.leaflet-left {
  top: 62px !important;
  left: 12px !important;
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

.district-survey-marker {
  background: #ff9800;
  color: #ffffff;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  box-shadow: 0 6px 16px rgba(255, 152, 0, 0.35);
  border: 2px solid rgba(255, 255, 255, 0.92);
}

.district-marker {
	background: transparent;
	border: 0;
}

.marker-count {
  background: #ea580c;
  color: #ffffff;
  border-radius: 50%;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  box-shadow: 0 6px 18px rgba(234, 88, 12, 0.42);
  border: 2px solid rgba(255, 255, 255, 0.95);
  font-variant-numeric: tabular-nums;
}

.district-survey-popup {
  font-size: 0.78rem;
  line-height: 1.4;
}

/* ── Anomaly toggle button — purple theme ────────────────────────────────── */
.anomaly-toggle-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.28rem 0.65rem;
  border: 1px solid #fecaca;
  border-radius: 6px;
  background: #fef2f2;
  color: #dc2626;
  font-size: 0.7rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.16s;
  white-space: nowrap;
}
.anomaly-toggle-btn:hover:not(:disabled) { background: #fee2e2; border-color: #ef4444; }
.anomaly-toggle-btn.active {
  background: #ef4444;
  color: #fff;
  border-color: #ef4444;
  box-shadow: 0 0 0 2px rgba(239,68,68,0.18);
}
.anomaly-toggle-btn.no-anomalies {
  border-color: #e2e8f0; background: #f8fafc; color: #b0bac6; cursor: not-allowed;
}
.anomaly-btn-icon { flex-shrink: 0; }
.anomaly-active-dot {
  width: 5px; height: 5px; border-radius: 50%; background: rgba(255,255,255,0.85); margin-left: 1px;
  animation: pulse-dot 1.4s ease-in-out infinite;
}
@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50%       { opacity: 0.35; transform: scale(0.5); }
}

/* ── Anomaly bottom drawer ───────────────────────────────────────────────── */
/* ══════════════════════════════════════════════════
   ANOMALY SIDEBAR  — right-side collapsible panel,
   flex sibling of .map-stage so map auto-resizes
══════════════════════════════════════════════════ */
.anomaly-sidebar {
  position: relative;
  width: 268px;
  min-width: 268px;
  flex-shrink: 0;
  border-left: 1.5px solid #fecaca;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  /* overflow: visible so the toggle button at left:-13px isn't clipped;
     actual content clipping is handled by .asb-body (overflow: hidden) */
  overflow: visible;
  z-index: 10;
  transition: width 0.24s cubic-bezier(.4,0,.2,1),
              min-width 0.24s cubic-bezier(.4,0,.2,1);
}
.anomaly-sidebar.asb-collapsed {
  width: 20px;
  min-width: 20px;
}

/* Toggle button — red theme, on the left edge */
.asb-toggle-btn {
  position: absolute;
  left: -13px;
  top: 50%;
  transform: translateY(-50%);
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: #ffffff;
  border: 1.5px solid #fecaca;
  color: #ef4444;
  font-size: 0.9rem;
  line-height: 1;
  cursor: pointer;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: -2px 0 8px rgba(239,68,68,0.15), 0 2px 6px rgba(0,0,0,0.08);
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}
.asb-toggle-btn:hover {
  border-color: #ef4444;
  color: #dc2626;
  background: #fef2f2;
}

/* Red counter badge — shown only when collapsed */
.asb-badge {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 17px;
  height: 17px;
  border-radius: 50%;
  background: #ef4444;
  color: #fff;
  font-size: 0.52rem;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid #fff;
  box-shadow: 0 1px 4px rgba(239,68,68,0.45);
  pointer-events: none;
}

/* Panel body — fills sidebar width, clipped when collapsed */
.asb-body {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

/* Panel header — red theme */
.asb-head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.7rem 0.8rem 0.6rem;
  border-bottom: 1px solid #fecaca;
  background: #fff5f5;
  flex-shrink: 0;
}
.asb-head-icon {
  width: 24px; height: 24px; border-radius: 6px;
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: #fff;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 2px 6px rgba(239,68,68,0.35);
}
.asb-head-text { flex: 1; min-width: 0; }
.asb-title {
  font-size: 0.75rem; font-weight: 700; color: #1e293b;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.asb-subtitle {
  font-size: 0.62rem; color: #94a3b8; margin-top: 1px;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.asb-close {
  flex-shrink: 0;
  width: 20px; height: 20px; border-radius: 50%;
  border: 1px solid #fecaca; background: transparent;
  color: #dc2626; font-size: 0.9rem; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  line-height: 1; transition: background 0.15s;
}
.asb-close:hover { background: #fee2e2; }

/* Hint strip */
.asb-hint {
  font-size: 0.62rem; color: #94a3b8;
  padding: 0.38rem 0.8rem;
  border-bottom: 1px solid #fef2f2;
  display: flex; align-items: center; gap: 0.35rem;
  background: #fff5f5;
  flex-shrink: 0;
}
.asb-hint-dot {
  display: inline-block;
  width: 6px; height: 6px; border-radius: 50%;
  background: #ef4444; flex-shrink: 0;
}

/* Scrollable list — fills remaining sidebar height (no fixed max-height cap) */
.asb-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
.asb-list::-webkit-scrollbar { width: 3px; }
.asb-list::-webkit-scrollbar-track { background: transparent; }
.asb-list::-webkit-scrollbar-thumb { background: #fca5a5; border-radius: 3px; }

.asb-item {
  width: 100%;
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.45rem 0.8rem;
  background: transparent;
  border: none;
  border-bottom: 1px solid #fef2f2;
  cursor: pointer; text-align: left;
  transition: background 0.1s;
  white-space: nowrap;
}
.asb-item:last-child { border-bottom: none; }
.asb-item:hover { background: #fef2f2; }
.asb-item--active { background: #fee2e2 !important; }

.asb-item-num {
  flex-shrink: 0;
  width: 18px; height: 18px; border-radius: 50%;
  background: #fef2f2; color: #ef4444; border: 1px solid #fecaca;
  font-size: 0.58rem; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
}
.asb-item--active .asb-item-num { background: #fee2e2; border-color: #ef4444; color: #dc2626; }

.asb-item-body { flex: 1; min-width: 0; }
.asb-item-name {
  font-size: 0.7rem; font-weight: 600; color: #1e293b;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.asb-item-meta { display: flex; align-items: center; gap: 0.3rem; margin-top: 1px; }
.asb-item-village {
  font-size: 0.6rem; color: #94a3b8;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 100px;
}
.asb-item-dist {
  flex-shrink: 0;
  font-size: 0.58rem; font-weight: 600;
  background: #fef2f2; color: #dc2626;
  padding: 0.06rem 0.28rem; border-radius: 3px;
}
.asb-item-arrow { flex-shrink: 0; color: #e2e8f0; transition: color 0.1s; }
.asb-item:hover .asb-item-arrow  { color: #fca5a5; }
.asb-item--active .asb-item-arrow { color: #ef4444; }

</style>
