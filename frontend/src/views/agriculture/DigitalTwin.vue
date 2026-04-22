<template>
  <div class="twin-page">
    <div class="cesium-wrap" ref="cesiumContainer">
      <!-- MAP-ONLY FULLSCREEN TOGGLE -->
      <button
        class="map-fs-btn"
        :class="{ shifted: selectedHouse || selectedCluster }"
        @click="toggleTwinFullscreen"
        :title="isTwinFullscreen ? 'Exit fullscreen' : 'Fullscreen'"
        aria-label="Toggle fullscreen"
      >
        {{ isTwinFullscreen ? '⤡' : '⤢' }}
      </button>

      <!-- FULLSCREEN-ONLY LEGEND (house coloring) -->
      <div v-if="isTwinFullscreen && colorMode" class="fs-legend">
        <div class="card-title">{{ legendTitle }}</div>
        <div class="legend-item" v-for="leg in currentLegend" :key="leg.label">
          <span class="mini-house" :style="{ '--mh-roof': leg.color }">
            <span class="mh-roof"></span>
            <span class="mh-wall"></span>
          </span>
          <span class="legend-text">{{ leg.label }}</span>
        </div>
        <div class="legend-note">Roof color = {{ legendTitle.toLowerCase() }} status</div>
      </div>

      <!-- DETAIL PANEL (must live inside fullscreen container) -->
      <transition name="slide">
        <div
          v-if="selectedHouse"
          class="detail-panel"
          :class="{ 'detail-panel-fs': isTwinFullscreen }"
        >

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
            <button class="detail-close" @click.stop="selectedHouse = null" title="Close">×</button>
          </div>

          <button class="focus-btn" @click="flyToHouse(selectedHouse)">📍 Zoom to Location</button>

          <!-- ── Section A: Population ── -->
          <div class="dp-section-label">
            <span class="dp-section-icon">👥</span> Family Details
          </div>

          <div class="dp-stat-row">
            <div class="dp-stat">
              <div class="dp-stat-val">{{ selectedHouse.totalMembers || 0 }}</div>
              <div class="dp-stat-key">Members</div>
            </div>
            <div class="dp-stat">
              <div class="dp-stat-val">{{ selectedHouse.maleMembers || 0 }}</div>
              <div class="dp-stat-key">Male</div>
            </div>
            <div class="dp-stat">
              <div class="dp-stat-val">{{ selectedHouse.femaleMembers || 0 }}</div>
              <div class="dp-stat-key">Female</div>
            </div>
          </div>

          <div class="dp-field-row">
            <span class="dp-field-icon">💼</span>
            <span class="dp-field-key">Working Members</span>
            <span class="dp-field-val" :class="selectedHouse.workingMembers > 0 ? 'dp-ok' : 'dp-warn'">
              {{ selectedHouse.workingMembers || 0 }}
            </span>
          </div>
          <div class="dp-field-row" v-if="selectedHouse.divyangMembers > 0">
            <span class="dp-field-icon">♿</span>
            <span class="dp-field-key">Divyang Members</span>
            <span class="dp-field-val dp-warn">{{ selectedHouse.divyangMembers }}</span>
          </div>
          <div class="dp-field-row">
            <span class="dp-field-icon">📚</span>
            <span class="dp-field-key">Illiterate Members</span>
            <span class="dp-field-val" :class="selectedHouse.illiterateMembers > 0 ? 'dp-warn' : 'dp-ok'">
              {{ selectedHouse.illiterateMembers || 0 }}
            </span>
          </div>
          <div class="dp-field-row" v-if="selectedHouse.bplCategory">
            <span class="dp-field-icon">🧾</span>
            <span class="dp-field-key">BPL Category</span>
            <span class="dp-field-val" :class="isBPL(selectedHouse) ? 'dp-warn' : 'dp-ok'">
              {{ selectedHouse.bplCategory }}
            </span>
          </div>
          <div class="dp-field-row" v-if="selectedHouse.annualIncome">
            <span class="dp-field-icon">₹</span>
            <span class="dp-field-key">Annual Income</span>
            <span class="dp-field-val">{{ selectedHouse.annualIncome }}</span>
          </div>

          <!-- ── Section B: Agriculture ── -->
          <div class="dp-section-label">
            <span class="dp-section-icon">🌾</span> Farming Details
          </div>

          <div class="dp-stat-row">
            <div class="dp-stat">
              <div class="dp-stat-val">{{ displayLandValue(selectedHouse.totalLand) }} <small v-if="displayLandValue(selectedHouse.totalLand) !== '—'">ac</small></div>
              <div class="dp-stat-key">Total Land</div>
            </div>
            <div class="dp-stat">
              <div class="dp-stat-val">{{ displayLandValue(selectedHouse.cultivatedLand) }} <small v-if="displayLandValue(selectedHouse.cultivatedLand) !== '—'">ac</small></div>
              <div class="dp-stat-key">Cultivated</div>
            </div>
          </div>

          <div class="dp-chip-row">
            <div class="dp-chip-block">
              <div class="dp-chip-label">Kharif Crop</div>
              <div class="dp-chip dp-chip-kharif">{{ displayCropValue(selectedHouse.kharif) }}</div>
            </div>
            <div class="dp-chip-block">
              <div class="dp-chip-label">Rabi Crop</div>
              <div class="dp-chip dp-chip-rabi">{{ displayCropValue(selectedHouse.rabi) }}</div>
            </div>
          </div>

          <div v-if="selectedHouseFarmingNote" class="dp-empty-note">{{ selectedHouseFarmingNote }}</div>

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

          <!-- ── Farm Advisory (DB-driven) ── -->
          <div class="dp-section-label">
            <span class="dp-section-icon">⚠️</span> Farm Advisory
          </div>

          <!-- Loading -->
          <div v-if="advisoryCache[selectedHouse.familyId]?.loading" class="advisory-loading">
            <span class="advisory-spinner"></span> Loading advisory…
          </div>

          <!-- Issues -->
          <template v-else-if="advisoryCache[selectedHouse.familyId]?.issues?.length">
            <div
              class="advisory-card"
              v-for="iss in advisoryCache[selectedHouse.familyId].issues"
              :key="iss.problemKey"
              :style="{ borderLeftColor: iss.color }"
            >
              <!-- Title row -->
              <div class="advisory-title-row">
                <span class="advisory-title" :style="{ color: iss.color }">{{ iss.problemLabel }}</span>
                <span v-if="iss.cropContext" class="advisory-crop-tag">🌾 {{ iss.cropContext }}</span>
              </div>

              <!-- Cause -->
              <div class="advisory-row">
                <span class="advisory-tag cause">Cause</span>
                <span class="advisory-text">{{ iss.cause }}</span>
              </div>

              <!-- Solution -->
              <div class="advisory-row">
                <span class="advisory-tag solution">Solution</span>
                <span class="advisory-text">{{ iss.solution }}</span>
              </div>

              <!-- Scheme / Source footer -->
              <div class="advisory-footer">
                <span class="advisory-scheme-pill"
                      :class="iss.schemeType === 'government_scheme' ? 'pill-gov' : 'pill-tech'"
                      :style="{ borderColor: iss.color + '55' }">
                  <span class="pill-icon">{{ iss.schemeType === 'government_scheme' ? '🏛' : '🔬' }}</span>
                  {{ iss.schemeName }}
                </span>
                <span class="advisory-source-tag"
                      :class="iss.source === 'advisory_master' ? 'src-db' : iss.source === 'scheme_criteria' ? 'src-scheme' : 'src-curated'">
                  {{ iss.source === 'advisory_master' ? '● Agriculture Dept DB' :
                     iss.source === 'scheme_criteria' ? '● Scheme Database' :
                     '● Agriculture Dept' }}
                </span>
              </div>
            </div>
          </template>

          <!-- No issues -->
          <div v-else-if="advisoryCache[selectedHouse.familyId] && !advisoryCache[selectedHouse.familyId].issues?.length" class="all-good">
            <span>✓</span> This household looks well-resourced
          </div>

        </div>
      </transition>
    </div>

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
          <label class="filter-label">View by</label>
          <div class="custom-select cs-align-right" :class="{ open: openDropdown === 'colorMode' }"
               @click.stop="toggleDropdown('colorMode')">
            <button class="cs-trigger" type="button" :class="{ 'cs-trigger-placeholder': !colorMode }">
              <span class="cs-value">{{ selectedColorModeLabel }}</span>
              <span class="cs-arrow">▾</span>
            </button>
            <div class="cs-dropdown cs-dropdown-right" v-show="openDropdown === 'colorMode'" @click.stop>
              <div class="cs-option-group-label">— Financial —</div>
              <div class="cs-option" :class="{ selected: colorMode === 'income_bracket' }"     @click="selectColorMode('income_bracket')">Family Income Status</div>
              <div class="cs-option" :class="{ selected: colorMode === 'bpl_status' }"         @click="selectColorMode('bpl_status')">BPL Status</div>
              <div class="cs-option-group-label">— Population —</div>
              <div class="cs-option" :class="{ selected: colorMode === 'population_density' }" @click="selectColorMode('population_density')">Population Density</div>
              <div class="cs-option" :class="{ selected: colorMode === 'education_level' }"    @click="selectColorMode('education_level')">Education Level</div>
              <div class="cs-option" :class="{ selected: colorMode === 'divyang_presence' }"   @click="selectColorMode('divyang_presence')">Divyang Presence</div>
              <div class="cs-option" :class="{ selected: colorMode === 'occupation' }"         @click="selectColorMode('occupation')">Occupation</div>
              <div class="cs-option-group-label">— Agriculture —</div>
              <div class="cs-option" :class="{ selected: colorMode === 'crops' }"              @click="selectColorMode('crops')">Crop Type</div>
              <div class="cs-option" :class="{ selected: colorMode === 'irrigation' }"         @click="selectColorMode('irrigation')">Irrigation</div>
              <div class="cs-option" :class="{ selected: colorMode === 'land' }"               @click="selectColorMode('land')">Land Holdings</div>
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

    <!-- VIEWPORT LOADING — shown while a viewport fetch is in-flight -->
    <div class="loading-overlay map-bg-loading-overlay" v-if="!loadingLiveData && viewportLoading">
      <div class="loading-spinner"></div>
      <div class="loading-text">Loading map data…</div>
    </div>

    <!-- CENTERING MAP — shown while the initial fit-bounds fly animation runs -->
    <div class="loading-overlay centering-overlay" v-if="centeringMap">
      <div class="loading-spinner"></div>
      <div class="loading-text">Centering map…</div>
    </div>

    <!-- EMPTY VIEWPORT HINT (non-blocking; preserves Cesium interactions) -->
    <div class="map-empty-toast" v-if="showEmptyViewportHint && !loadingLiveData && !viewportLoading">
      No households in this view. Pan or zoom to a populated area.
    </div>

    <!-- STATS BAR -->
    <div class="stats-bar" v-if="!loadingLiveData">
      <span class="stat-item">
        <span class="stat-dot" style="background:#16a34a"></span>
        <strong>{{
          isLocationFiltered
            ? householdsOnMapCount.toLocaleString()
            : (agricultureInsights?.totalHouseholds || houses.length).toLocaleString()
        }}</strong> households
        <span v-if="isLocationFiltered" class="stat-filter-note">
          ({{ householdsOnMapCount.toLocaleString() }} on map)
        </span>
      </span>
      <span class="stat-sep">·</span>
      <span class="stat-item"><strong>{{ totalPopulation.toLocaleString() }}</strong> population</span>
      <span class="stat-sep">·</span>
      <span class="stat-item"><strong>{{ farmersOwnLandCount.toLocaleString() }}</strong> farmer households</span>
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

        <!-- LEGEND — always visible; default state when no mode is selected -->
        <div class="panel-card">
          <div class="card-title">{{ colorMode ? legendTitle : 'Map Legend' }}</div>
          <template v-if="colorMode">
            <div class="legend-item" v-for="leg in currentLegend" :key="leg.label">
              <span class="mini-house" :style="{ '--mh-roof': leg.color }">
                <span class="mh-roof"></span>
                <span class="mh-wall"></span>
              </span>
              <span class="legend-text">{{ leg.label }}</span>
            </div>
            <div class="legend-note">Roof color = {{ legendTitle.toLowerCase() }} status</div>
          </template>
          <template v-else>
            <div class="legend-item">
              <span class="mini-house" :style="{ '--mh-roof': '#ef4444' }">
                <span class="mh-roof"></span>
                <span class="mh-wall"></span>
              </span>
              <span class="legend-text">All Households
                <span class="legend-count-pill">{{ householdsOnMapCount.toLocaleString() }}</span>
              </span>
            </div>
            <div class="legend-note">Select a category from <strong>View By</strong> to colour the map</div>
          </template>
        </div>

        <!-- VILLAGE SUMMARY — shown only when no View By is selected -->
        <div class="panel-card" v-if="!colorMode">
          <div class="card-title">Village Summary</div>
          <div class="issue-row">
            <span class="issue-pip" style="background:#16a34a"></span>
            <div class="issue-body">
              <div class="issue-top">
                <span class="issue-name">Total Households</span>
                <span class="issue-count">{{
                  isLocationFiltered
                    ? householdsOnMapCount.toLocaleString()
                    : (agricultureInsights?.totalHouseholds || houses.length).toLocaleString()
                }}</span>
              </div>
              <div class="issue-track"><div class="issue-fill" style="width:100%;background:#16a34a"></div></div>
            </div>
          </div>
          <div class="issue-row">
            <span class="issue-pip" style="background:#2563eb"></span>
            <div class="issue-body">
              <div class="issue-top">
                <span class="issue-name">Total Population</span>
                <span class="issue-count">{{ totalPopulation.toLocaleString() }}</span>
              </div>
              <div class="issue-track"><div class="issue-fill" style="width:100%;background:#2563eb"></div></div>
            </div>
          </div>
          <div class="issue-row">
            <span class="issue-pip" style="background:#2563eb"></span>
            <div class="issue-body">
              <div class="issue-top">
                <span class="issue-name">Male {{ malePct }}%</span>
                <span class="issue-count">{{ maleTotal.toLocaleString() }}</span>
              </div>
              <div class="issue-track"><div class="issue-fill" :style="{ width: malePct + '%', background: '#2563eb' }"></div></div>
            </div>
          </div>
          <div class="issue-row">
            <span class="issue-pip" style="background:#ec4899"></span>
            <div class="issue-body">
              <div class="issue-top">
                <span class="issue-name">Female {{ femalePct }}%</span>
                <span class="issue-count">{{ femaleTotal.toLocaleString() }}</span>
              </div>
              <div class="issue-track"><div class="issue-fill" :style="{ width: femalePct + '%', background: '#ec4899' }"></div></div>
            </div>
          </div>
          <div class="issue-row">
            <span class="issue-pip" style="background:#16a34a"></span>
            <div class="issue-body">
              <div class="issue-top">
                <span class="issue-name">Farmers (own land)</span>
                <span class="issue-count">{{ farmersOwnLandCount.toLocaleString() }}</span>
              </div>
              <div class="issue-track"><div class="issue-fill" :style="{ width: farmersOwnLandPct + '%', background: '#16a34a' }"></div></div>
            </div>
          </div>
          <div class="pf-hint" style="margin-top:10px">Select a category from <strong>View By</strong> to explore detailed analytics.</div>
        </div>

        <!-- PROBLEM FILTER — only shown when a mode is active -->
        <div class="panel-card" v-if="colorMode && availableProblemFilters.length">
          <div class="card-title">Problem Filter
            <span class="card-title-sub">highlight on map</span>
          </div>
          <div class="pf-context-label">Filters for {{ selectedColorModeLabel }}</div>
          <transition-group name="pf-fade" tag="div" class="pf-list">
            <label class="pf-item" v-for="pf in availableProblemFilters" :key="pf.key">
              <input class="pf-check" type="checkbox" :value="pf.key" v-model="activeProblemFilters" />
              <span class="mini-house mini-house-sm" :style="{ '--mh-roof': pf.color }">
                <span class="mh-roof"></span>
                <span class="mh-wall"></span>
              </span>
              <span class="pf-label">{{ pf.label }}</span>
              <span class="pf-count">{{ formatProblemCount(problemFilterStats[pf.key]) }}</span>
            </label>
          </transition-group>
          <div class="pf-hint" v-if="!hasDetailedHouseData">Loading detailed household counts for current view…</div>
          <div class="pf-summary" v-if="activeProblemFilters.length">
            <span><strong>{{ problemMatchCount }}</strong> flagged</span>
            <button class="pf-clear-btn" @click="activeProblemFilters = []">✕ Clear</button>
          </div>
          <div class="pf-hint" v-else>
            Select filters to highlight at-risk households on the map
          </div>
        </div>

        <!-- FIELD ISSUES — only shown when a mode is active -->
        <div class="panel-card" v-if="colorMode && issueList.length">
          <div class="card-title">{{ selectedColorModeLabel }} Analysis
            <span class="card-title-sub">tap to expand schemes</span>
          </div>
          <transition-group name="fi-fade" tag="div" class="fi-list">
          <div v-for="issue in issueList" :key="issue.key">
            <div class="issue-row" :class="{ active: colorMode === issue.mode }"
                 @click="colorMode = issue.mode; toggleSchemeDrawer(issue.key)">
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
              <span class="issue-chevron" :class="{ open: schemeDrawer === issue.key }">›</span>
            </div>
            <transition name="drawer">
              <div v-if="schemeDrawer === issue.key" class="issue-drawer scheme-drawer" :style="{ borderLeftColor: issue.color }">
                <!-- Cause — from API (DB-verified or fallback) -->
                <div v-if="schemeCache[issue.key]">
                  <p class="drawer-cause">
                    <strong>Cause:</strong>
                    {{ schemeCache[issue.key].loading ? 'Analysing…' : schemeCache[issue.key].cause }}
                  </p>

                  <!-- Scheme list -->
                  <template v-if="!schemeCache[issue.key].loading">
                    <div class="scheme-header">
                      <span class="scheme-header-icon">🏛</span>
                      <span class="scheme-header-text">Recommended Schemes based on eligibility</span>
                      <span class="scheme-source-tag" :class="schemeCache[issue.key].source === 'db' ? 'tag-db' : 'tag-fb'">
                        {{ schemeCache[issue.key].source === 'db' ? '● Live DB' : '● Curated' }}
                      </span>
                    </div>
                    <div v-if="schemeCache[issue.key].schemes.length === 0" class="scheme-empty">
                      No matching schemes found. Contact the local Gram Panchayat for assistance.
                    </div>
                    <div v-for="s in schemeCache[issue.key].schemes" :key="s.name" class="scheme-card" :style="{ borderLeftColor: issue.color }">
                      <div class="scheme-card-name">{{ s.name }}</div>
                      <div class="scheme-card-desc">{{ s.description }}</div>
                      <div class="scheme-card-row">
                        <span class="scheme-tag scheme-tag-benefit">💰 {{ s.benefit }}</span>
                      </div>
                      <div class="scheme-card-row">
                        <span class="scheme-tag scheme-tag-eligibility">✅ {{ s.eligibility }}</span>
                      </div>
                      <div class="scheme-card-reason" :style="{ color: issue.color }">
                        <span class="scheme-reason-icon">🎯</span> {{ s.matchReason }}
                      </div>
                    </div>
                  </template>
                  <div v-else class="scheme-loading">
                    <span class="scheme-spinner"></span> Loading schemes…
                  </div>
                </div>
              </div>
            </transition>
          </div>
          </transition-group>
        </div>

        <!-- OVERVIEW CHARTS — context-aware: population or agriculture based on View By -->
        <div class="panel-card" v-if="colorMode && availablePieCharts.length">
          <div class="card-title card-title-toggle" @click="agriOverviewOpen = !agriOverviewOpen">
            <span>{{ isPopulationMode ? 'Population' : 'Agriculture' }} Overview</span>
            <span class="card-toggle-icon" :class="{ open: agriOverviewOpen }">▾</span>
          </div>
          <transition name="agri-collapse">
            <div v-show="agriOverviewOpen" class="agri-collapsible-body">
              <transition-group name="agri-fade" tag="div" class="agri-list">
              <div class="agri-chart" v-for="chart in availablePieCharts" :key="chart.title">
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
              </transition-group>
            </div>
          </transition>
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

    <!-- CLUSTER SOLUTION PANEL -->
    <transition name="slide">
      <div v-if="selectedCluster" class="cluster-panel">
        <button class="detail-close cluster-close" @click="selectedCluster = null; clusterAdvisory = null; highlightClusterBoundary(null)">×</button>

        <!-- ── Header ── -->
        <div class="cluster-header">
          <div class="cluster-priority-badge"
               :class="clusterAdvisory && !clusterAdvisory.loading && clusterAdvisory.priorityLabel?.includes('High') ? 'badge-high' : 'badge-moderate'">
            {{ clusterAdvisory && !clusterAdvisory.loading ? clusterAdvisory.priorityLabel : '⚠ Analysing Cluster…' }}
          </div>
          <div class="cluster-count">
            <strong>{{ selectedCluster.count }}</strong> households in this zone
          </div>
          <div class="cluster-location-pill" v-if="selectedCluster.lat">
            📍 {{ selectedCluster.lat.toFixed(4) }}°, {{ selectedCluster.lng.toFixed(4) }}°
          </div>
        </div>

        <!-- ── Loading state ── -->
        <div v-if="!clusterAdvisory || clusterAdvisory.loading" class="cluster-loading">
          <span class="advisory-spinner"></span> Loading group advisory…
        </div>

        <template v-else>
          <!-- ── Group Action Cards ── -->
          <div class="cluster-section-title" v-if="clusterAdvisory.actions.length">
            🔍 Group Issues &amp; Community Actions
          </div>

          <div class="cp-group-card"
               v-for="action in clusterAdvisory.actions"
               :key="action.problemKey"
               :class="{ 'cp-mass': action.isMassIssue }">

            <!-- Mass Issue heading -->
            <div v-if="action.isMassIssue" class="cp-mass-heading">
              🚨 {{ action.massHeading }}
            </div>

            <!-- Problem bar -->
            <div class="cp-top">
              <span class="cp-emoji">{{ selectedCluster.problems.find(p=>p.key===action.problemKey)?.emoji || '⚠' }}</span>
              <span class="cp-label">{{ action.problemLabel }}</span>
              <span class="cp-stat">{{ action.count }} of {{ action.total }} families ({{ action.affectedPct }}%)</span>
            </div>
            <div class="cp-bar-track">
              <div class="cp-bar-fill" :class="action.isMassIssue ? 'fill-red' : 'fill-amber'"
                   :style="{ width: action.affectedPct + '%' }"></div>
            </div>

            <!-- Cause -->
            <div class="cp-cause-row">
              <span class="cp-tag cp-tag-cause">Cause</span>
              <span class="cp-cause-text">{{ action.cause }}</span>
            </div>

            <!-- Group Action (Recommended) -->
            <div class="cp-action-row">
              <span class="cp-tag cp-tag-action">Recommended Action</span>
              <span class="cp-action-text">{{ action.groupAction }}</span>
            </div>

            <!-- Scheme footer -->
            <div class="cp-scheme-footer">
              <span class="cp-scheme-pill"
                    :class="action.schemeType === 'community_scheme' ? 'pill-community' : 'pill-gov'">
                {{ action.schemeType === 'community_scheme' ? '🤝' : '🏛' }}
                {{ action.schemeName }}
              </span>
              <span v-if="action.schemeBenefit" class="cp-benefit-pill">💰 {{ action.schemeBenefit }}</span>
              <span class="cp-source-tag"
                    :class="action.source === 'scheme_criteria' ? 'src-db' : 'src-curated'">
                {{ action.source === 'scheme_criteria' ? '● Scheme Database' : '● Agriculture Dept' }}
              </span>
            </div>
          </div>

          <div v-if="!clusterAdvisory.actions.length" class="cluster-ok">
            ✅ No major issues detected in this cluster based on current filters.
          </div>

          <!-- ── Drill-down button ── -->
          <div class="cp-drill-row" v-if="selectedCluster.count > 0">
            <button class="cp-drill-btn" @click="drillIntoCluster(selectedCluster)">
              🔎 View Individual Households
            </button>
            <span class="cp-drill-hint">Zooms in to show individual household dots</span>
          </div>
        </template>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { getHouses, getHousesByViewport, getHousesMapPoints, getHouseById, getLocationOptions, getHousesSummary, getAgricultureInsights, getPopulationDashboard, getSchemesForProblem, getAdvisory, getClusterAdvisory } from '../../api/index.js'
import * as Cesium from 'cesium'
import Supercluster from 'supercluster'
import 'cesium/Build/Cesium/Widgets/widgets.css'

Cesium.Ion.defaultAccessToken = ''

// ── Core state ────────────────────────────────────────────────────────────────
const houses              = ref([])
const mapPoints           = ref([])
const selectedHouse       = ref(null)
const hoveredHouse        = ref(null)
const mouseX              = ref(0)
const mouseY              = ref(0)
const colorMode           = ref(null)
const activeIssue         = ref(null)
const agriOverviewOpen    = ref(false)

// ── Scheme recommendations ────────────────────────────────────────────────────
// Cache: problemKey → { loading, cause, source, schemes[] }
const schemeCache  = reactive({})
const schemeDrawer = ref(null) // currently open issue key in scheme drawer

// ── Farm Advisory (per-household, DB-driven) ──────────────────────────────────
// Cache: familyId → { loading, issues[] }
const advisoryCache = reactive({})

async function loadAdvisoryForHouse(house) {
  if (!house) return
  const cacheKey = house.familyId
  if (advisoryCache[cacheKey]) return // already loaded or loading

  // Detect problem keys from the house's field data
  const problems = []
  const totalLand = parseFloat(house.totalLand) || 0
  const cultivated = parseFloat(house.cultivatedLand) || 0
  const ownLand = (house.ownLand || '').toLowerCase()

  if (ownLand !== 'yes' || totalLand <= 0) {
    problems.push('noLand')
  } else if (totalLand <= 1) {
    problems.push('marginalHolding')
  }
  if (cultivated <= 0 && totalLand > 0) problems.push('uncultivated')
  if (isRainFed(house)) problems.push('noIrrigation')

  const normCrop = (v) => {
    const s = String(v || '').trim().toLowerCase()
    return s && s !== 'no' && s !== 'none' && s !== 'na' ? s : ''
  }
  const k = normCrop(house.kharif)
  const r = normCrop(house.rabi)
  if (!k && !r) problems.push('noCropRecord')
  else if (!k || !r) problems.push('singleSeason')

  if (!problems.length) {
    advisoryCache[cacheKey] = { loading: false, issues: [] }
    return
  }

  advisoryCache[cacheKey] = { loading: true, issues: [] }

  // Build profile for contextual filtering
  const profile = { family_id: house.familyId }
  const primaryCrop = k || r
  if (primaryCrop) profile.crop = primaryCrop
  if (totalLand <= 1) profile.land_size = 'marginal'
  else if (totalLand <= 2.5) profile.land_size = 'small'
  else profile.land_size = 'large'
  if (isBPL(house)) profile.bpl = 'yes'

  try {
    const data = await getAdvisory(problems, profile)
    advisoryCache[cacheKey] = { loading: false, issues: data.issues || [] }
  } catch {
    // Fallback: build advisory locally from getIssues() if API fails
    advisoryCache[cacheKey] = { loading: false, issues: getIssues(house).map(i => ({
      problemKey: i.label,
      problemLabel: i.label,
      color: i.color,
      cause: i.cause,
      solution: i.solution,
      schemeName: i.scheme,
      schemeType: 'curated',
      source: 'curated',
      cropContext: '',
    }))}
  }
}
const demoOverviewOpen    = ref(true)
const tileStyle           = ref('satellite')   // default satellite view
const cesiumContainer     = ref(null)
const agricultureInsights = ref(null)
const populationDashboard = ref(null)
const loadingLiveData     = ref(true)
const viewportLoading     = ref(false)   // true while a viewport fetch is in-flight
const showEmptyViewportHint = ref(false) // subtle non-blocking notice when viewport returns zero rows
const centeringMap        = ref(false)   // true while the initial fit-bounds fly is in progress
const sidebarCollapsed    = ref(false)
const cameraHeight        = ref(120000)
// Persisted pitch (radians) — updated on every camera move, used by all fly functions
// so filter changes and house clicks never reset the user's tilt.
let currentMapPitch = Cesium.Math.toRadians(-48)  // default oblique; updated live
const isTwinFullscreen    = ref(false)

// Location filters — applied state (drives filteredHouses + map)
const filterDistrict = ref('')
const filterTaluka   = ref('')
const filterVillage  = ref('')

// Pending state — bound to UI dropdowns, only committed on Apply
const pendingDistrict = ref('')
const pendingTaluka   = ref('')
const pendingVillage  = ref('')

const districtOptions = ref([])
const talukaOptions = ref([])
const villageOptions = ref([])

let viewer        = null
let ptCollection  = null          // PointPrimitiveCollection for all 40k household dots
let clusterBillboardCollection = null // BillboardCollection for clustered markers
let clusterIndex = null            // Supercluster index
let clusterRenderTimer = null
const clusterImageCache = new Map()
let buildSeq      = 0             // incremented each buildEntities() call; stale async runs check this
let prevShowBuildings = false     // tracks last known building-zoom state for lazy entity creation
let buildingPanTimer = null       // debounce handle for viewport-pan rebuilds within building zoom
const jitterCache = new Map()     // familyId → {lat, lng}  populated during chunked build
const entityMap  = new Map()   // entityId → house  (building boxes + cluster entities)
const ptPrimMap  = new Map()   // familyId → PointPrimitive  (fast primitive lookup)
const buildingIds = new Set()  // 3D box entity IDs
const clusterIds   = new Set()  // High-Need cluster entity IDs (problem filter rings)
const clusterMap   = new Map()  // clusterEntityId → { count, lat, lng, problems[] }
const macroClusIds       = new Set()  // Grid cluster markers at district/state zoom level
const miniClusIds        = new Set()  // Grid cluster markers at taluka zoom level
const zoomClusterDataMap = new Map()  // entityId → { lat, lng, houses[] } for drill-down
const houseToPointId     = new Map()  // familyId → point entity ID for spiderfy (kept for compat)
const spiderfyHouseEntityIds = []
let spiderfyCenter = null
let retryTimer         = null
let twinLoadSeq        = 0
let viewportSeq        = 0
let viewportDebounce   = null
let lastLoadedBbox     = null
let isInitialLoadDone  = false   // true after first page of data has been fetched and rendered
let vpInFlight         = 0       // count of loadViewportData calls currently awaiting a fetch
const viewportTileCache = new Map()   // cacheKey → { ts, data }
const VIEWPORT_CACHE_TTL = 5 * 60 * 1000   // 5 minutes
const VIEWPORT_CACHE_MAX = 30

function toMapPointHouse(point) {
  const id = Number(point?.id)
  const latitude = Number(point?.lat)
  const longitude = Number(point?.lng)
  return {
    familyId: id,
    latitude,
    longitude,
    headName: `Household ${id}`,
    villageName: '',
    talukaName: '',
    districtName: '',
    totalMembers: 0,
    maleMembers: 0,
    femaleMembers: 0,
    workingMembers: 0,
    illiterateMembers: 0,
    divyangMembers: 0,
    unemployedMembers: 0,
    totalLand: '',
    cultivatedLand: '',
    ownLand: '',
    waterSource: '',
    kharif: '',
    rabi: '',
    latrine: '',
    lighting: '',
    rationCard: '',
    occupation: '',
    bplCategory: '',
    annualIncome: '',
  }
}

function handleTwinFullscreenChange() {
  isTwinFullscreen.value = !!document.fullscreenElement
  handleResize()
}

// Zoom thresholds (meters)
const THRESHOLD_BUILDINGS = 3500    // below: show 3D boxes
const THRESHOLD_DOTS      = 15000   // below: show individual point beacons (village level)
const THRESHOLD_MACRO     = 80000   // above: show macro grid clusters (district/state)
// Cluster glow rings are shown from village level ALL THE WAY DOWN (incl. building view)
// so the cluster boundary stays visible even when the user zooms into individual houses.
const THRESHOLD_CLUSTER_HIDE = THRESHOLD_DOTS  // rings hidden only at taluka+ zoom

async function loadLocationOptions() {
  try {
    const res = await getLocationOptions({
      district_id: pendingDistrict.value || undefined,
      taluka_id: pendingTaluka.value || undefined,
    })
    districtOptions.value = Array.isArray(res?.districts) ? res.districts : []
    talukaOptions.value = Array.isArray(res?.talukas) ? res.talukas : []
    villageOptions.value = Array.isArray(res?.villages) ? res.villages : []
  } catch (error) {
    console.warn('[location-options] failed:', error?.message || error)
    districtOptions.value = []
    talukaOptions.value = []
    villageOptions.value = []
  }
}

// Active filtered subset displayed on map
const filteredHouses = computed(() => {
  let result = houses.value
  if (filterDistrict.value) result = result.filter(h => String(h.districtId) === String(filterDistrict.value))
  if (filterTaluka.value)   result = result.filter(h => String(h.talukaId)   === String(filterTaluka.value))
  if (filterVillage.value)  result = result.filter(h => String(h.villageId)  === String(filterVillage.value))
  return result
})

const detailedHouseById = computed(() => {
  const byId = new Map()
  for (const house of filteredHouses.value) {
    const id = Number(house?.familyId)
    if (Number.isFinite(id)) byId.set(id, house)
  }
  return byId
})

const householdsOnMapCount = computed(() => {
  // mapPoints represents the full filtered household set from DB (not viewport-limited)
  // while filteredHouses may be capped by viewport fetch limit.
  if (mapPoints.value.length > 0) return mapPoints.value.length
  return filteredHouses.value.length
})

// ── Filter handlers ───────────────────────────────────────────────────────────
const openDropdown = ref(null)

function toggleDropdown(name) {
  openDropdown.value = openDropdown.value === name ? null : name
  if (name === 'colorMode' && openDropdown.value === 'colorMode') {
    selectedHouse.value = null
  }
}

function closeDropdowns() {
  openDropdown.value = null
}

// Reset child pending selections when a parent changes
function onDistrictChange() {
  pendingTaluka.value  = ''
  pendingVillage.value = ''
}
function onTalukaChange() {
  pendingVillage.value = ''
}

async function selectDistrict(id) {
  pendingDistrict.value = id
  onDistrictChange()
  await loadLocationOptions()
}

async function selectTaluka(id) {
  pendingTaluka.value = id
  onTalukaChange()
  await loadLocationOptions()
}

function selectVillage(id) {
  pendingVillage.value = id
  closeDropdowns()
}

// Apply: copy pending → applied, reload filtered map points, then focus camera.
async function applyFilters() {
  _forceNextFly = true
  filterDistrict.value = pendingDistrict.value
  filterTaluka.value   = pendingTaluka.value
  filterVillage.value  = pendingVillage.value
  await loadInitialDataWithCleanup()

  // Fit camera to filtered map points so location-based Apply is always visible.
  if (viewer && mapPoints.value.length > 0) {
    setTimeout(() => flyToPoints(mapPoints.value.map(toMapPointHouse)), 120)
  }
}

async function resetFilters() {
  pendingDistrict.value = ''
  pendingTaluka.value = ''
  pendingVillage.value = ''
  filterDistrict.value = ''
  filterTaluka.value = ''
  filterVillage.value = ''
  _forceNextFly = true
  await loadLocationOptions()
  await loadInitialDataWithCleanup()
}

const filtersDirty = computed(() =>
  pendingDistrict.value !== filterDistrict.value ||
  pendingTaluka.value !== filterTaluka.value ||
  pendingVillage.value !== filterVillage.value
)

const COLOR_MODE_LABELS = {
  irrigation:          'Irrigation',
  occupation:          'Occupation',
  crops:               'Crop Type',
  land:                'Land Holdings',
  ration:              'Ration Card',
  population_density:  'Population Density',
  bpl_status:          'BPL Status',
  education_level:     'Education Level',
  divyang_presence:    'Divyang Presence',
  income_bracket:      'Family Income Status',
}

// Category group for each color mode — used to reset problem filters on category switch
const COLOR_MODE_CATEGORY = {
  income_bracket:     'financial',
  bpl_status:         'financial',
  ration:             'financial',
  population_density: 'population',
  education_level:    'population',
  divyang_presence:   'population',
  occupation:         'population',
  crops:              'agriculture',
  irrigation:         'agriculture',
  land:               'agriculture',
}

const DISABLED_COLOR_MODES = new Set(['sanitation', 'lighting'])

function isColorModeEnabled(mode) {
  if (!mode) return false
  return !DISABLED_COLOR_MODES.has(String(mode)) && String(mode) !== 'ration'
}

function selectColorMode(mode) {
  if (!isColorModeEnabled(mode)) {
    closeDropdowns()
    return
  }
  const prevCategory = COLOR_MODE_CATEGORY[colorMode.value]
  const nextCategory = COLOR_MODE_CATEGORY[mode]
  // Clear problem filters when switching between categories
  if (prevCategory !== nextCategory) {
    activeProblemFilters.value = []
  }
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
const selectedColorModeLabel = computed(() => colorMode.value ? (COLOR_MODE_LABELS[colorMode.value] || colorMode.value) : 'Select a view…')

// ── Problem Filter state ──────────────────────────────────────────────────────
// Array of active problem keys; v-model on checkboxes drives buildEntities()
const activeProblemFilters = ref([])

// Static metadata — one entry per problem type
const PROBLEM_FILTER_META = [
  // Agri problems
  { key: 'noSanitation',      label: 'No Sanitation',      color: '#ef4444' },
  { key: 'noRationCard',      label: 'No Ration Card',     color: '#f97316' },
  { key: 'noIrrigation',      label: 'No Irrigation',      color: '#a78bfa' },
  { key: 'noLand',            label: 'No Own Land',        color: '#ef4444' },
  { key: 'unemployed',        label: 'Unemployed',         color: '#ef4444' },
  { key: 'laborers',          label: 'Laborers',           color: '#f59e0b' },
  // Population problems
  { key: 'bplFamilies',       label: 'BPL Families',       color: '#60a5fa' },
  { key: 'illiterateMembers', label: 'Illiterate Members', color: '#f59e0b' },
  { key: 'unemployedMembers', label: 'Unemployed Members', color: '#ef4444' },
  { key: 'divyangMembers',    label: 'Divyang Members',    color: '#7b1fa2' },
]

const PROBLEM_FILTERS_BY_MODE = {
  irrigation:         ['noIrrigation', 'noLand'],
  crops:              ['noIrrigation', 'noLand'],
  occupation:         ['unemployed', 'laborers'],
  population_density: ['bplFamilies', 'illiterateMembers', 'unemployedMembers', 'divyangMembers'],
  bpl_status:         ['bplFamilies', 'noRationCard'],
  education_level:    ['illiterateMembers'],
  divyang_presence:   ['divyangMembers'],
  income_bracket:     ['bplFamilies', 'unemployedMembers', 'unemployed'],
}

const availableProblemFilterKeys = computed(() => {
  if (!colorMode.value) return []
  return PROBLEM_FILTERS_BY_MODE[colorMode.value] || PROBLEM_FILTER_META.map(p => p.key)
})

const availableProblemFilters = computed(() => {
  const allowed = new Set(availableProblemFilterKeys.value)
  return PROBLEM_FILTER_META.filter(p => allowed.has(p.key))
})

function getOccupationText(house) {
  return String(house.occupation || '').toLowerCase().trim()
}

function isUnemployedHouse(house) {
  const occ = getOccupationText(house)
  return occ.includes('unemploy') || occ.includes('not working') || occ.includes('no work') || occ.includes('jobless')
}

function isLaborHouse(house) {
  const occ = getOccupationText(house)
  return occ.includes('labor') || occ.includes('labour') || occ.includes('wage') || occ.includes('worker')
}

// Returns true if the house matches the given problem key
function matchesProblemFilter(house, key) {
  if (key === 'noSanitation') {
    const s = (house.latrine || '').toLowerCase().trim()
    return !s || s === 'no latrine' || s === 'none'
  }
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
  if (key === 'unemployed') return isUnemployedHouse(house)
  if (key === 'laborers')   return isLaborHouse(house)
  // Population problem filters
  if (key === 'bplFamilies')       return isBPL(house)
  if (key === 'illiterateMembers') return (house.illiterateMembers || 0) > 0
  if (key === 'unemployedMembers') return (house.unemployedMembers || 0) > 0
  if (key === 'divyangMembers')    return (house.divyangMembers    || 0) > 0
  return false
}

// Returns true if house satisfies ALL active problem filters (AND logic)
function matchesAllProblems(house) {
  return activeProblemFilters.value.every(k => matchesProblemFilter(house, k))
}

/**
 * getSolutionsByCriteria — fetch scheme recommendations for a problem key.
 * Builds a citizen profile from the currently selected house (if any) or
 * falls back to aggregated data. Results are cached by problemKey.
 */
async function getSolutionsByCriteria(problemKey) {
  if (schemeCache[problemKey]) return // already loaded or loading
  schemeCache[problemKey] = { loading: true, cause: '', source: '', schemes: [] }

  // Build a lightweight citizen profile from the selected house or filtered aggregate
  const house = selectedHouse.value
  const profile = {}
  if (house) {
    const land = parseFloat(house.totalLand) || 0
    if (land <= 1)   profile.land_size = 'marginal'
    else if (land <= 2.5) profile.land_size = 'small'
    else              profile.land_size = 'large'
    const occ = (house.occupation || '').toLowerCase().trim()
    if (occ && occ !== 'unemployed') profile.occupation = occ.split(' ')[0]
    if (isBPL(house)) profile.bpl = 'yes'
  }

  try {
    const data = await getSchemesForProblem(problemKey, profile)
    schemeCache[problemKey] = { loading: false, ...data }
  } catch {
    schemeCache[problemKey] = {
      loading: false,
      cause: `Data analysis suggests a gap in this area. Please review with the relevant department.`,
      source: 'error',
      schemes: [],
    }
  }
}

// Toggle scheme drawer — load on first open
function toggleSchemeDrawer(issueKey) {
  if (schemeDrawer.value === issueKey) {
    schemeDrawer.value = null
    return
  }
  schemeDrawer.value = issueKey
  getSolutionsByCriteria(issueKey)
}

// Per-key counts for the sidebar display
const problemFilterStats = computed(() => {
  const list = filteredHouses.value
  const counts = {}
  for (const pf of PROBLEM_FILTER_META) {
    counts[pf.key] = list.filter(h => matchesProblemFilter(h, pf.key)).length
  }
  return counts
})

const hasDetailedHouseData = computed(() => filteredHouses.value.length > 0)

function formatProblemCount(value) {
  if (!hasDetailedHouseData.value) return '—'
  return Number.isFinite(Number(value)) ? Number(value) : 0
}

// Total households matching ALL active problem filters simultaneously
const problemMatchCount = computed(() => {
  if (!activeProblemFilters.value.length) return 0
  return filteredHouses.value.filter(matchesAllProblems).length
})

// ── Cluster solution panel state ──────────────────────────────────────────────
const selectedCluster = ref(null)  // { count, lat, lng, problems[], houses[] }
// Group advisory state for the selected cluster
const clusterAdvisory = ref(null)  // null | { loading, priorityLabel, actions[] }
// ID of the boundary-highlight entity for the active cluster
let clusterBoundaryId = null

async function loadClusterAdvisory(cluster) {
  if (!cluster || !cluster.problems) return
  clusterAdvisory.value = { loading: true, priorityLabel: '', actions: [] }
  try {
    const stats = cluster.problems.map(p => ({ key: p.key, count: p.count, total: cluster.count }))
    const data  = await getClusterAdvisory(stats, cluster.count)
    clusterAdvisory.value = { loading: false, ...data }
  } catch {
    clusterAdvisory.value = { loading: false, priorityLabel: 'Cluster Advisory Unavailable', actions: [] }
  }
}

// Draw/remove a boundary ring around the active cluster on the map
function highlightClusterBoundary(cluster) {
  if (!viewer) return
  // Remove previous boundary
  if (clusterBoundaryId) {
    viewer.entities.removeById(clusterBoundaryId)
    clusterBoundaryId = null
  }
  if (!cluster) return
  const pos = Cesium.Cartesian3.fromDegrees(cluster.lng, cluster.lat, 0)
  const r   = Math.min(150 + cluster.count * 4, 700)
  const ent = viewer.entities.add({
    position: pos,
    ellipse: {
      semiMajorAxis:   r,
      semiMinorAxis:   r,
      material:        Cesium.Color.fromCssColorString('#ef4444').withAlpha(0.08),
      outline:         true,
      outlineColor:    Cesium.Color.fromCssColorString('#ef4444').withAlpha(0.7),
      outlineWidth:    2,
      heightReference: Cesium.HeightReference.CLAMP_TO_GROUND,
    },
  })
  clusterBoundaryId = ent.id
}

// Zoom in on the cluster and switch to individual-household view
function drillIntoCluster(cluster) {
  if (!viewer || !cluster) return
  selectedCluster.value = null
  clusterAdvisory.value = null
  // Remove boundary highlight
  if (clusterBoundaryId) {
    viewer.entities.removeById(clusterBoundaryId)
    clusterBoundaryId = null
  }
  // Fly to the cluster location at village-level altitude, preserving pitch
  viewer.camera.flyTo({
    destination: Cesium.Cartesian3.fromDegrees(cluster.lng, cluster.lat, 1800),
    orientation: {
      heading: viewer.camera.heading,
      pitch: Math.max(Math.min(currentMapPitch, MAX_PITCH_RAD), MIN_PITCH_RAD),
      roll: 0,
    },
    duration: 1.5,
    easingFunction: Cesium.EasingFunction.QUADRATIC_IN_OUT,
  })
}


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
    key:      'noSanitation',
    label:    'No Sanitation',
    emoji:    '🚽',
    action:   'Run sanitation infrastructure drive for this cluster',
    solution: 'Eligible households can apply for toilet subsidy and local implementation support through Gram Panchayat sanitation plans.',
    scheme:   'Swachh Bharat Mission (Gramin)',
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
  {
    key:      'unemployed',
    label:    'Unemployed',
    emoji:    '🧭',
    action:   'Arrange local livelihood placement and skill linkage camp',
    solution: 'Connect eligible members to skill centers, self-employment support, and local job fairs coordinated by district livelihood missions.',
    scheme:   'DDU-GKY / NRLM Livelihood Support',
  },
  {
    key:      'laborers',
    label:    'Laborers',
    emoji:    '🛠',
    action:   'Prioritize wage-worker support and social security enrollment',
    solution: 'Register wage workers for social security, job cards, and safety schemes through block-level facilitation camps.',
    scheme:   'MGNREGA + e-Shram Registration',
  },
  {
    key:      'bplFamilies',
    label:    'BPL Families',
    emoji:    '🧾',
    action:   'Ensure food and health entitlement access across this zone',
    solution: 'Ensure access to food security and health support benefits through Gram Panchayat facilitation camps.',
    scheme:   'NFSA Ration Card eligibility · Ayushman Bharat',
  },
  {
    key:      'illiterateMembers',
    label:    'Illiterate Members',
    emoji:    '📚',
    action:   'Run school and Anganwadi enrollment outreach for this cluster',
    solution: 'Encourage enrollment in school or Anganwadi education services and adult literacy programs.',
    scheme:   'Anganwadi · Saakshar Bharat Mission',
  },
  {
    key:      'unemployedMembers',
    label:    'Unemployed Members',
    emoji:    '🧰',
    action:   'Mobilize e-Shram and SHG registration camp in this area',
    solution: 'Encourage registration on e-Shram portal and participation in Self Help Groups.',
    scheme:   'e-Shram · DDU-GKY',
  },
  {
    key:      'divyangMembers',
    label:    'Divyang Members',
    emoji:    '♿',
    action:   'Verify disability certification and pension enrollment status',
    solution: 'Ensure disability certificate and enrollment in disability pension and welfare schemes.',
    scheme:   'Disability Pension Support · NHFDC',
  },
]

// Analyse the houses inside a cluster → all problems sorted by count, each with pct
function analyzeCluster(houseList) {
  const total = houseList.length
  return CLUSTER_PROBLEM_META
    .map(meta => ({
      ...meta,
      count: houseList.filter(h => matchesProblemFilter(h, meta.key)).length,
    }))
    .filter(p => p.count > 0)
    .sort((a, b) => b.count - a.count)
    .map(p => ({ ...p, pct: Math.round((p.count / total) * 100) }))
}

// When location filter changes (district/taluka/village applied), rebuild
// entities and fly to the filtered set — always fly when the user pressed Apply
// (forcefly flag), otherwise only fly if houses are not already in view.
let _forceNextFly = false   // set by applyFilters, consumed once by the watcher

watch(filteredHouses, (newHouses) => {
  if (!viewer) return
  // Do not clear/rebuild while viewport data is still resolving.
  // Otherwise the scene blanks out temporarily and looks like houses vanish.
  if (loadingLiveData.value || viewportLoading.value) {
    viewer.scene.requestRender()
    return
  }

  buildEntities()
  if (!newHouses.length) return

  // Keep the current zoom while data is still resolving so the user can
  // see the empty-state hint in the same camera context.
  if (showEmptyViewportHint.value) return

  const force = _forceNextFly
  _forceNextFly = false

  // Do not auto-fly on initial load. Fly only when explicitly forced
  // (Apply filter) or when refreshed data is outside the current viewport.
  if (force || !housesInView(newHouses)) {
    setTimeout(() => flyToPoints(newHouses), 150)
  }
}, { flush: 'post' })

// ── Zoom label ────────────────────────────────────────────────────────────────
const zoomLabel = computed(() => {
  const h = cameraHeight.value
  if (h < THRESHOLD_BUILDINGS) return '3D buildings visible'
  if (h < THRESHOLD_DOTS)      return 'Village view — individual households'
  if (h < THRESHOLD_MACRO)     return 'Taluka view — cluster density'
  return 'District / State view — zoom in to explore'
})

// ── Population aggregates ─────────────────────────────────────────────────────
// When no location filter is active, use the DB-wide totals from agricultureInsights
// (covers ALL families, not just the ~2 000 that have GPS coordinates).
// When a filter is active (district / taluka / village), fall back to summing from
// the visible filtered houses so the numbers match what the map is showing.
const isLocationFiltered = computed(() => !!(filterDistrict.value || filterTaluka.value || filterVillage.value))

const totalPopulation = computed(() => {
  if (!isLocationFiltered.value && populationDashboard.value?.total_population)
    return populationDashboard.value.total_population
  if (!isLocationFiltered.value && agricultureInsights.value?.totalPopulation)
    return agricultureInsights.value.totalPopulation
  return filteredHouses.value.reduce((s, h) => s + (h.totalMembers || 0), 0)
})
const maleTotal = computed(() => {
  if (!isLocationFiltered.value && agricultureInsights.value?.totalMale)
    return agricultureInsights.value.totalMale
  return filteredHouses.value.reduce((s, h) => s + (h.maleMembers || 0), 0)
})
const femaleTotal = computed(() => {
  if (!isLocationFiltered.value && agricultureInsights.value?.totalFemale)
    return agricultureInsights.value.totalFemale
  return filteredHouses.value.reduce((s, h) => s + (h.femaleMembers || 0), 0)
})
const malePct   = computed(() => totalPopulation.value ? Math.round(maleTotal.value / totalPopulation.value * 100) : 0)
const femalePct = computed(() => totalPopulation.value ? Math.round(femaleTotal.value / totalPopulation.value * 100) : 0)
const workingHouseholds = computed(() => filteredHouses.value.filter(h => (h.workingMembers || 0) >= 1).length)
const divyangHouseholds = computed(() => filteredHouses.value.filter(h => (h.divyangMembers || 0) >= 1).length)
const literacyRate      = computed(() => {
  const total = totalPopulation.value
  const illiterate = filteredHouses.value.reduce((s, h) => s + (h.illiterateMembers || 0), 0)
  if (!total) return 100
  return Math.round(((total - illiterate) / total) * 100)
})

function isBPL(house) {
  const v = String(house.bplCategory || house.rationCard || '').toLowerCase()
  return v.includes('bpl') || v.includes('antyodaya') || v === 'yes'
}

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

const farmersOwnLandCount = computed(() => {
  if (!isLocationFiltered.value && agricultureInsights.value?.totalFarmers != null) {
    return Number(agricultureInsights.value.totalFarmers) || 0
  }
  if (stats.value?.farmers != null) return stats.value.farmers
  return 0
})

const farmersOwnLandPct = computed(() => {
  const denom = isLocationFiltered.value
    ? householdsOnMapCount.value
    : Number(agricultureInsights.value?.totalHouseholds || householdsOnMapCount.value || 0)
  if (!denom) return 0
  return Math.max(0, Math.min(100, Math.round((farmersOwnLandCount.value / denom) * 100)))
})

function displayLandValue(value) {
  const raw = String(value ?? '').trim()
  if (!raw) return '—'
  const n = Number(raw)
  if (Number.isFinite(n) && n <= 0) return '—'
  return raw
}

function displayCropValue(value) {
  const raw = String(value ?? '').trim()
  if (!raw) return '—'
  const normalized = raw.toLowerCase()
  if (['no', 'none', 'n/a', 'na', '-', '--'].includes(normalized)) return '—'
  return raw
}

function hasRecordedCrop(value) {
  return displayCropValue(value) !== '—'
}

const selectedHouseFarmingNote = computed(() => {
  const house = selectedHouse.value
  if (!house) return ''

  const hasLandData = displayLandValue(house.totalLand) !== '—' || displayLandValue(house.cultivatedLand) !== '—'
  const hasCropData = displayCropValue(house.kharif) !== '—' || displayCropValue(house.rabi) !== '—'
  const ownLand = String(house.ownLand || '').toLowerCase()
  if (ownLand && ownLand !== 'yes' && !hasLandData && !hasCropData) {
    return 'Household has no own agricultural land.'
  }

  // Some records have cultivation values even when OWN_AGRICULTURE_LAND is "No"
  // (leased/shared land). Prefer showing available data over a hard "no land" note.
  if (ownLand && ownLand !== 'yes' && (hasLandData || hasCropData)) {
    return 'Household may be cultivating non-owned (leased/shared) land.'
  }

  if (!hasLandData && !hasCropData) {
    return 'Farming fields are not available in survey data for this household.'
  }
  return ''
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
  noRationCard: {
    cause: 'Household may be excluded from subsidised food benefits due to missing ration card status',
    solution: 'Run local enrollment and document verification camps to onboard eligible families into PDS/NFSA.',
    scheme: 'National Food Security Act (NFSA)',
  },
  land: {
    cause: 'Landless or near-zero owned land limits agriculture-based income opportunities',
    solution: 'Link households to lease-farming groups, livelihood missions, and producer collectives.',
    scheme: 'NRLM / FPO Linkages',
  },
  occupation: {
    cause: 'Low access to stable work opportunities increases livelihood vulnerability',
    solution: 'Connect members to skill programs and district employment facilitation drives.',
    scheme: 'DDU-GKY / Skill India',
  },
}

const issueListAll = computed(() => {
  if (!stats.value) return []
  const { total, noToilet, noElec, noIrrig, bpl } = stats.value
  const list = filteredHouses.value
  const noLand           = list.filter(h => matchesProblemFilter(h, 'noLand')).length
  const noRationCard     = list.filter(h => matchesProblemFilter(h, 'noRationCard')).length
  const unemployed       = list.filter(h => isUnemployedHouse(h)).length
  const laborers         = list.filter(h => isLaborHouse(h)).length
  // Population
  const bplFamilies      = list.filter(h => matchesProblemFilter(h, 'bplFamilies')).length
  const illiterateCnt    = list.filter(h => matchesProblemFilter(h, 'illiterateMembers')).length
  const unemployedMems   = list.filter(h => matchesProblemFilter(h, 'unemployedMembers')).length
  const divyangCnt       = list.filter(h => matchesProblemFilter(h, 'divyangMembers')).length

  const pct = (n) => Math.round(n / total * 100)
  return [
    { key: 'noSanitation',      label: 'No Sanitation',      count: noToilet,     pct: pct(noToilet),     color: '#ef4444', mode: 'sanitation',         ...ISSUE_META.sanitation    },
    { key: 'noRationCard',      label: 'No Ration Card',     count: noRationCard,  pct: pct(noRationCard), color: '#f97316', mode: 'sanitation',         ...ISSUE_META.noRationCard  },
    { key: 'noIrrigation',      label: 'No Irrigation',      count: noIrrig,       pct: pct(noIrrig),      color: '#a78bfa', mode: 'irrigation',         ...ISSUE_META.irrigation    },
    { key: 'noLand',            label: 'No Own Land',        count: noLand,        pct: pct(noLand),       color: '#ef4444', mode: 'land',               ...ISSUE_META.land          },
    { key: 'unemployed',        label: 'Unemployed',         count: unemployed,    pct: pct(unemployed),   color: '#ef4444', mode: 'occupation',         ...ISSUE_META.occupation    },
    { key: 'laborers',          label: 'Laborers',           count: laborers,      pct: pct(laborers),     color: '#f59e0b', mode: 'occupation',         ...ISSUE_META.occupation    },
    { key: 'noElectricity',     label: 'No Electricity',     count: noElec,        pct: pct(noElec),       color: '#f59e0b', mode: 'lighting',           ...ISSUE_META.lighting      },
    { key: 'bplHouseholds',     label: 'BPL Households',     count: bpl,           pct: pct(bpl),          color: '#60a5fa', mode: 'ration',             ...ISSUE_META.ration        },
    { key: 'bplFamilies',       label: 'BPL Families',       count: bplFamilies,   pct: pct(bplFamilies),  color: '#60a5fa', mode: 'bpl_status',         cause: 'Household classified as BPL — economically vulnerable.',           solution: 'Ensure NFSA ration card, Ayushman Bharat, and PM-KISAN enrollment.', scheme: 'NFSA · Ayushman Bharat' },
    { key: 'illiterateMembers', label: 'Illiterate Members', count: illiterateCnt, pct: pct(illiterateCnt),color: '#f59e0b', mode: 'education_level',    cause: 'Households with illiterate members — limits income opportunities.', solution: 'Enroll in adult literacy programs and Anganwadi services.', scheme: 'Saakshar Bharat Mission' },
    { key: 'unemployedMembers', label: 'Unemployed Members', count: unemployedMems,pct: pct(unemployedMems),color:'#ef4444', mode: 'income_bracket',     cause: 'Household members recorded as unemployed.',                        solution: 'Connect to skill programs, SHGs, and e-Shram registration.', scheme: 'e-Shram · DDU-GKY'      },
    { key: 'divyangMembers',    label: 'Divyang Members',    count: divyangCnt,    pct: pct(divyangCnt),   color: '#7b1fa2', mode: 'divyang_presence',   cause: 'Households with divyang (disabled) members.',                      solution: 'Ensure disability certificate and pension scheme enrollment.', scheme: 'Disability Pension · NHFDC' },
  ]
})

const FIELD_ISSUES_BY_MODE = {
  irrigation:         ['noIrrigation', 'noLand'],
  crops:              ['noIrrigation', 'noLand'],
  occupation:         ['unemployed', 'laborers'],
  population_density: ['bplFamilies', 'illiterateMembers', 'unemployedMembers', 'divyangMembers'],
  bpl_status:         ['bplFamilies', 'noRationCard'],
  education_level:    ['illiterateMembers'],
  divyang_presence:   ['divyangMembers'],
  income_bracket:     ['bplFamilies', 'unemployedMembers'],
}

const availableFieldIssueKeys = computed(() => {
  if (!colorMode.value) return []
  return FIELD_ISSUES_BY_MODE[colorMode.value] || issueListAll.value.map(i => i.key)
})

const issueList = computed(() => {
  const allowed = new Set(availableFieldIssueKeys.value)
  return issueListAll.value.filter(i => allowed.has(i.key))
})

// ── Legend ────────────────────────────────────────────────────────────────────
const legendTitle = computed(() => ({
  irrigation: 'Irrigation',
  occupation: 'Occupation',
  ration: 'Ration Card',
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
  if (colorMode.value === 'occupation') return [
    { color: '#16a34a', label: 'Other working categories' },
    { color: '#f59e0b', label: 'Laborers / wage work' },
    { color: '#ef4444', label: 'Unemployed / not working' },
    { color: '#94a3b8', label: 'No occupation data' },
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
    { color: '#94a3b8', label: 'No land / data unavailable' },
  ]
  if (colorMode.value === 'population_density') return [
    { color: '#22c55e', label: '1–2 members (Small)' },
    { color: '#f59e0b', label: '3–5 members (Average)' },
    { color: '#ef4444', label: '6+ members (Large)' },
  ]
  if (colorMode.value === 'bpl_status') return [
    { color: '#ef4444', label: 'BPL household' },
    { color: '#16a34a', label: 'Non-BPL household' },
  ]
  if (colorMode.value === 'education_level') return [
    { color: '#16a34a', label: 'Literacy dominant (>60%)' },
    { color: '#f59e0b', label: 'Needs literacy support' },
  ]
  if (colorMode.value === 'divyang_presence') return [
    { color: '#7b1fa2', label: 'Divyang member present' },
    { color: '#16a34a', label: 'No disability recorded' },
  ]
  if (colorMode.value === 'income_bracket') return [
    { color: '#ef4444', label: 'Low income  (≤ ₹21k)' },
    { color: '#f59e0b', label: 'Mid income  (₹21k–50k)' },
    { color: '#16a34a', label: 'High income (> ₹50k)' },
    { color: '#94a3b8', label: 'No income data' },
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
  const noSanitation = list.filter(h => matchesProblemFilter(h, 'noSanitation')).length
  const unemployed = list.filter(h => isUnemployedHouse(h)).length
  const laborers = list.filter(h => isLaborHouse(h)).length
  const occupiedOther = Math.max(total - unemployed - laborers, 0)

  list.forEach(h => {
    const k = hasRecordedCrop(h.kharif)
    const r = hasRecordedCrop(h.rabi)
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
    {
      title: 'Sanitation Coverage',
      segments: [
        { label: 'Has Sanitation', pct: pct(total - noSanitation, total), color: '#16a34a' },
        { label: 'No Sanitation',  pct: pct(noSanitation, total), color: '#ef4444' },
      ],
    },
    {
      title: 'Occupation Profile',
      segments: [
        { label: 'Unemployed', pct: pct(unemployed, total), color: '#ef4444' },
        { label: 'Laborers', pct: pct(laborers, total), color: '#f59e0b' },
        { label: 'Other / Working', pct: pct(occupiedOther, total), color: '#16a34a' },
      ],
    },
  ]
})

const AGRI_OVERVIEW_BY_MODE = {
  sanitation: ['Sanitation Coverage'],
  irrigation: ['Irrigation Coverage', 'Land Holdings'],
  crops:      ['Crop Seasons', 'Irrigation Coverage'],
  occupation: ['Occupation Profile'],
  land:       ['Land Holdings'],
  lighting:   [],
  ration:     [],
}

// Population pie charts — computed from member aggregates
const popPieCharts = computed(() => {
  const list  = filteredHouses.value
  const total = list.length
  if (!total) return []
  const totalPop   = totalPopulation.value
  const illiterate = list.reduce((s, h) => s + (h.illiterateMembers || 0), 0)
  const literate   = Math.max(totalPop - illiterate, 0)
  const working    = list.reduce((s, h) => s + (h.workingMembers || 0), 0)
  const divyang    = list.reduce((s, h) => s + (h.divyangMembers  || 0), 0)
  const bplCount   = list.filter(h => isBPL(h)).length
  return [
    {
      title: 'Education Status',
      segments: [
        { label: 'Literate',   pct: pct(literate,   totalPop || 1), color: '#16a34a' },
        { label: 'Illiterate', pct: pct(illiterate, totalPop || 1), color: '#ef4444' },
      ],
    },
    {
      title: 'Gender Ratio',
      segments: [
        { label: 'Male',   pct: malePct.value,   color: '#2563eb' },
        { label: 'Female', pct: femalePct.value, color: '#ec4899' },
      ],
    },
    {
      title: 'Work Status (households)',
      segments: [
        { label: 'Working members present', pct: pct(workingHouseholds.value, total), color: '#16a34a' },
        { label: 'No working member',        pct: pct(total - workingHouseholds.value, total), color: '#ef4444' },
      ],
    },
    {
      title: 'BPL Status',
      segments: [
        { label: 'BPL household',     pct: pct(bplCount, total), color: '#ef4444' },
        { label: 'Non-BPL household', pct: pct(total - bplCount, total), color: '#16a34a' },
      ],
    },
    {
      title: 'Divyang Presence',
      segments: [
        { label: 'Divyang present', pct: pct(divyangHouseholds.value, total), color: '#7b1fa2' },
        { label: 'None recorded',   pct: pct(total - divyangHouseholds.value, total), color: '#16a34a' },
      ],
    },
  ]
})

const POP_OVERVIEW_BY_MODE = {
  population_density: ['Gender Ratio', 'Work Status (households)'],
  education_level:    ['Education Status'],
  divyang_presence:   ['Divyang Presence'],
  income_bracket:     ['BPL Status', 'Work Status (households)'],
  bpl_status:         ['BPL Status'],
  occupation:         ['Work Status (households)', 'Gender Ratio'],
}

// Returns true when the active mode belongs to the population category
const isPopulationMode = computed(() => COLOR_MODE_CATEGORY[colorMode.value] === 'population' || COLOR_MODE_CATEGORY[colorMode.value] === 'financial')

const availablePieCharts = computed(() => {
  if (!colorMode.value) return []
  // Population modes → population charts
  if (POP_OVERVIEW_BY_MODE[colorMode.value]) {
    const allowed = new Set(POP_OVERVIEW_BY_MODE[colorMode.value])
    return popPieCharts.value.filter(c => allowed.has(c.title))
  }
  // Agriculture modes → agri charts
  const allowedTitles = AGRI_OVERVIEW_BY_MODE[colorMode.value]
  if (allowedTitles) {
    const allowed = new Set(allowedTitles)
    return pieCharts.value.filter(chart => allowed.has(chart.title))
  }
  return pieCharts.value
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
  if (!colorMode.value) return '#ef4444'
  if (colorMode.value === 'sanitation') {
    const l = (house.latrine || '').toLowerCase()
    if (!l || l === 'no latrine' || l === 'none') return '#ef4444'
    if (l.includes('pit') || l.includes('open'))  return '#f59e0b'
    return '#16a34a'
  }
  if (colorMode.value === 'irrigation') {
    return isRainFed(house) ? '#ef4444' : '#16a34a'
  }
  if (colorMode.value === 'occupation') {
    const occ = getOccupationText(house)
    if (!occ) return '#94a3b8'
    if (isUnemployedHouse(house)) return '#ef4444'
    if (isLaborHouse(house)) return '#f59e0b'
    return '#16a34a'
  }
  if (colorMode.value === 'lighting') {
    const l = (house.lighting || '').toLowerCase()
    return l === 'electricity' ? '#16a34a' : l === 'kerosene' ? '#f59e0b' : '#ef4444'
  }
  if (colorMode.value === 'crops') {
    const k = hasRecordedCrop(house.kharif)
    const r = hasRecordedCrop(house.rabi)
    if (k && r) return '#16a34a'
    if (k)      return '#f59e0b'
    if (r)      return '#38bdf8'
    return '#94a3b8'
  }
  if (colorMode.value === 'land') {
    const rawLand = String(house.totalLand ?? '').trim()
    const a = parseFloat(rawLand) || 0
    const own = String(house.ownLand || '').toLowerCase().trim()
    if (!rawLand && !own) return '#94a3b8'
    if (a === 0)  return '#94a3b8'
    if (a <= 1)   return '#ef4444'
    if (a <= 2.5) return '#f59e0b'
    if (a <= 5)   return '#4ade80'
    return '#16a34a'
  }
  // Population modes
  if (colorMode.value === 'population_density') {
    const m = house.totalMembers || 0
    if (m <= 2) return '#22c55e'
    if (m <= 5) return '#f59e0b'
    return '#ef4444'
  }
  if (colorMode.value === 'bpl_status') {
    return isBPL(house) ? '#ef4444' : '#16a34a'
  }
  if (colorMode.value === 'education_level') {
    const ill = house.illiterateMembers || 0
    const total = house.totalMembers || 1
    return (ill / total) > 0.4 ? '#f59e0b' : '#16a34a'
  }
  if (colorMode.value === 'divyang_presence') {
    return (house.divyangMembers || 0) > 0 ? '#7b1fa2' : '#16a34a'
  }
  // Combined
  if (colorMode.value === 'income_bracket') {
    const v = String(house.annualIncome || house.rationCard || '').toLowerCase()
    if (v.includes('less than') || v.includes('bpl') || v.includes('antyodaya')) return '#ef4444'
    if (v.includes('apl') || v.includes('50') || v.includes('21')) return '#f59e0b'
    if (v.includes('above') || v.includes('high')) return '#16a34a'
    const n = parseFloat(String(house.annualIncome || '').replace(/[^0-9.]/g, ''))
    if (Number.isFinite(n) && n > 0) {
      if (n <= 21000) return '#ef4444'
      if (n <= 50000) return '#f59e0b'
      return '#16a34a'
    }
    return '#94a3b8'
  }
  const r = (house.rationCard || '').toLowerCase()
  if (r.includes('bpl') || r.includes('antyodaya')) return '#ef4444'
  if (r.includes('apl')) return '#16a34a'
  return '#94a3b8'
}

function getConditionLabel(house) {
  const color = getConditionColor(house)
  if (colorMode.value === 'occupation') {
    if (color === '#ef4444') return 'Unemployed'
    if (color === '#f59e0b') return 'Laborer / Wage Work'
    if (color === '#16a34a') return 'Working'
    return 'Occupation Not Available'
  }
  if (colorMode.value === 'crops') {
    if (color === '#16a34a') return 'Double Crop'
    if (color === '#f59e0b') return 'Kharif Only'
    if (color === '#38bdf8') return 'Rabi Only'
    return 'No Crop Data'
  }
  if (colorMode.value === 'land') {
    const rawLand = String(house.totalLand ?? '').trim()
    const own = String(house.ownLand || '').toLowerCase().trim()
    if (!rawLand && !own) return 'Land Data Unavailable'
    const a = parseFloat(rawLand) || 0
    if (a === 0)  return 'Landless'
    if (a <= 1)   return 'Marginal Farmer'
    if (a <= 2.5) return 'Small Farmer'
    if (a <= 5)   return 'Medium Holding'
    return 'Large Holding'
  }
  if (!colorMode.value) return 'Household'
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
    return new Cesium.OpenStreetMapImageryProvider({
      url: 'https://tile.openstreetmap.org/',
      credit: '© OpenStreetMap contributors',
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

// ── Camera ────────────────────────────────────────────────────────────────────
function flyToMaharashtra() {
  if (!viewer) return
  viewer.camera.flyTo({
    destination: Cesium.Cartesian3.fromDegrees(76.0, 19.5, 150000),
    orientation: { heading: viewer.camera.heading, pitch: Cesium.Math.toRadians(-48), roll: 0 },
    duration: 1.8,
    easingFunction: Cesium.EasingFunction.QUADRATIC_IN_OUT,
  })
}

// Returns true if the majority of houses in `list` are already within the
// camera's current view frustum, meaning no auto-fly is needed.
function housesInView(list) {
  if (!viewer) return false
  const valid = list.filter(h => Number.isFinite(h.longitude) && Number.isFinite(h.latitude))
  if (!valid.length) return false
  const canvas    = viewer.scene.canvas
  const w = canvas.clientWidth, h = canvas.clientHeight
  let inCount = 0
  for (const house of valid) {
    const cart  = Cesium.Cartesian3.fromDegrees(house.longitude, house.latitude, 0)
    const win   = Cesium.SceneTransforms.worldToWindowCoordinates(viewer.scene, cart)
    if (win && win.x >= 0 && win.x <= w && win.y >= 0 && win.y <= h) inCount++
  }
  // Fly only if fewer than 60% of houses are currently on-screen
  return inCount / valid.length >= 0.60
}

function flyToPoints(list) {
  if (!viewer || !list.length) return
  const valid = list.filter(h => Number.isFinite(h.longitude) && Number.isFinite(h.latitude))
  if (!valid.length) return

  const lats    = valid.map(h => h.latitude)
  const lngs    = valid.map(h => h.longitude)
  const cLat    = (Math.min(...lats) + Math.max(...lats)) / 2
  const cLng    = (Math.min(...lngs) + Math.max(...lngs)) / 2
  const spanLat = Math.max(...lats) - Math.min(...lats)
  const spanLng = Math.max(...lngs) - Math.min(...lngs)
  const spanDeg = Math.max(spanLat, spanLng, 0.002)

  // Convert span to meters and use as bounding sphere radius.
  // Add generous padding so houses don't crowd the edges.
  const radiusM = Math.max(spanDeg * 111000 * 0.65, 800)

  // Clamp the range (eye-to-sphere-centre distance) so we land between
  // THRESHOLD_BUILDINGS (3 500 m) and taluka level (40 000 m).
  // This ensures individual dot markers are always visible on arrival.
  const range = Math.max(Math.min(radiusM * 3.5, 38000), 5000)

  const centre = Cesium.Cartesian3.fromDegrees(cLng, cLat, 0)
  const sphere = new Cesium.BoundingSphere(centre, radiusM)

  // flyToBoundingSphere with an explicit HeadingPitchRange is the most
  // reliable Cesium API for landing at a specific pitch — it does NOT
  // re-derive orientation from the terrain, unlike flyTo({destination, orientation}).
  const pitch   = Math.max(Math.min(currentMapPitch, MAX_PITCH_RAD), MIN_PITCH_RAD)
  viewer.camera.flyToBoundingSphere(sphere, {
    duration: 2.0,
    offset: new Cesium.HeadingPitchRange(
      viewer.camera.heading,   // preserve current bearing
      pitch,                   // persisted user tilt (never resets to 0)
      range,
    ),
  })
}

function flyToVillage() { flyToPoints(filteredHouses.value) }

function flyToHouse(house) {
  if (!viewer || !house) return
  const pitch = Math.max(Math.min(currentMapPitch, Cesium.Math.toRadians(-30)), Cesium.Math.toRadians(-70))
  viewer.camera.flyTo({
    destination: Cesium.Cartesian3.fromDegrees(house.longitude, house.latitude, 200),
    orientation: { heading: viewer.camera.heading, pitch, roll: 0 },
    duration: 1.5,
    easingFunction: Cesium.EasingFunction.QUADRATIC_IN_OUT,
  })
}

// ── Zoom-based entity visibility ──────────────────────────────────────────────
// Four zoom levels — something meaningful is always shown at every level:
//   < 3500 m  (THRESHOLD_BUILDINGS) : 3D box buildings
//   3500–15000 m (THRESHOLD_DOTS)   : individual point beacons + high-need glow rings
//   15000–80000 m (THRESHOLD_MACRO) : mini grid cluster circles (taluka density)
//   > 80000 m  (THRESHOLD_MACRO)    : macro grid cluster circles (district/state density)
// Minimum pitch: user can't tilt the camera flatter than this (in radians).
// Cesium pitch is negative (−90° = straight down, 0° = horizon).
const MIN_PITCH_RAD = Cesium.Math.toRadians(-80)   // 80° down — keeps buildings visible
const MAX_PITCH_RAD = Cesium.Math.toRadians(-15)   // 15° down — can't go near-horizontal

function getClusterZoomFromHeight(height) {
  if (height > 3000000) return 4
  if (height > 1500000) return 5
  if (height > 900000) return 7
  if (height > 500000) return 9
  if (height > 250000) return 11
  if (height > 120000) return 12
  if (height > 60000) return 13
  if (height > 30000) return 14
  if (height > 15000) return 15
  if (height > 7000) return 16
  if (height > 3500) return 17
  return 18
}

function getCurrentSuperclusterBBox() {
  if (!viewer || viewer.isDestroyed()) return null
  const rect = viewer.camera.computeViewRectangle()
  if (!rect) return null

  const west = Cesium.Math.toDegrees(rect.west)
  const south = Cesium.Math.toDegrees(rect.south)
  const east = Cesium.Math.toDegrees(rect.east)
  const north = Cesium.Math.toDegrees(rect.north)

  // Dateline wrap handling: use world-longitude bounds for stable clustering.
  if (west > east) return [-180, south, 180, north]
  return [west, south, east, north]
}

function ensureClusterCollections() {
  if (!viewer || viewer.isDestroyed()) return
  if (!ptCollection || ptCollection.isDestroyed()) {
    ptCollection = viewer.scene.primitives.add(new Cesium.PointPrimitiveCollection())
  }
  if (!clusterBillboardCollection || clusterBillboardCollection.isDestroyed()) {
    clusterBillboardCollection = viewer.scene.primitives.add(new Cesium.BillboardCollection())
  }
}

function buildSuperclusterIndexFromHouses() {
  const source = mapPoints.value
  const features = source
    .filter((h) => Number.isFinite(Number(h.lng)) && Number.isFinite(Number(h.lat)))
    .map((h) => ({
      type: 'Feature',
      geometry: {
        type: 'Point',
        coordinates: [Number(h.lng), Number(h.lat)],
      },
      properties: { id: Number(h.id) },
    }))

  clusterIndex = new Supercluster({
    // Merge nearby points earlier so clusters do not overlap visually.
    radius: 100,
    // Any group smaller than 20 is rendered as individual houses (not clustered).
    minPoints: 20,
    maxZoom: 18,
    minZoom: 0,
    nodeSize: 64,
  })
  clusterIndex.load(features)
}

function getHeightFromClusterZoom(zoom) {
  const z = Math.max(1, Math.min(18, Number(zoom || 10)))
  const h = 16000000 / Math.pow(2, z)
  return Math.max(600, Math.min(300000, h))
}

function getAltitudeFromMeters(meters) {
  return Math.max(700, Math.min(120000, Number(meters || 1000)))
}

function getSpiderfyOffsets(count, radiusMeters = 24) {
  const offsets = []
  if (count <= 0) return offsets

  const radius = Math.max(20, Math.min(30, radiusMeters))
  const step = (2 * Math.PI) / count
  for (let i = 0; i < count; i += 1) {
    const angle = step * i
    offsets.push({
      east: Math.cos(angle) * radius,
      north: Math.sin(angle) * radius,
    })
  }
  return offsets
}

function offsetCartesianDegrees(lat, lng, eastMeters, northMeters) {
  const metersPerDegLat = 111320
  const metersPerDegLng = Math.max(Math.cos(Cesium.Math.toRadians(lat)) * 111320, 1)
  return {
    lat: lat + (northMeters / metersPerDegLat),
    lng: lng + (eastMeters / metersPerDegLng),
  }
}

function addSpiderLeg(centerCartesian, offsetCartesian) {
  if (!viewer || viewer.isDestroyed()) return null
  const ent = viewer.entities.add({
    polyline: {
      positions: [centerCartesian, offsetCartesian],
      width: 1.5,
      material: Cesium.Color.fromCssColorString('#f8fafc').withAlpha(0.95),
    },
  })
  return ent?.id ?? null
}

function getClusterIconSize(count) {
  const c = Number(count || 0)
  return Math.max(30, Math.min(45, 30 + (Math.log10(c + 1) * 6)))
}

function generateClusterBillboardImage(count, expansionZoom) {
  const size = getClusterIconSize(count)
  const spiderfyCap = Number(expansionZoom || 0) >= 19
  const variant = spiderfyCap ? 'spiderfy' : 'zoom'
  const key = `${Math.round(size)}:${count > 999 ? 'k' : 'n'}:${variant}`
  const cached = clusterImageCache.get(key)
  if (cached) return cached

  const canvasSize = 96
  const canvas = document.createElement('canvas')
  canvas.width = canvasSize
  canvas.height = canvasSize
  const ctx = canvas.getContext('2d')
  if (!ctx) return ''

  const cx = canvasSize / 2
  const cy = canvasSize / 2
  const r = (size / 2) * 1.85

  // Use a distinct color for spiderfy-cap clusters (no further zoom expansion).
  ctx.beginPath()
  ctx.arc(cx, cy, r, 0, Math.PI * 2)
  ctx.fillStyle = spiderfyCap ? '#F9A825' : '#E65100'
  ctx.fill()

  // Crisp white border.
  ctx.lineWidth = 3
  ctx.strokeStyle = '#FFFFFF'
  ctx.stroke()

  const text = count > 999 ? `${Math.round(count / 100) / 10}k` : String(count)
  ctx.fillStyle = '#FFFFFF'
  ctx.font = text.length > 3 ? '700 24px Inter, Segoe UI, sans-serif' : '700 28px Inter, Segoe UI, sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(text, cx, cy)

  const image = canvas.toDataURL('image/png')
  clusterImageCache.set(key, image)
  return image
}

function addHouseModelEntity(house, lng, lat) {
  const selectedId       = selectedHouse.value?.familyId
  const hasProblemFilter = activeProblemFilters.value.length > 0
  const isSelected       = house.familyId === selectedId
  const isProblem        = hasProblemFilter && matchesAllProblems(house)
  const isBackground     = hasProblemFilter && !isProblem && !isSelected

  const conditionColor = cesiumColor(house)
  const roofAlpha      = isSelected ? 1.0 : isBackground ? 0.35 : 1.0
  const roofColor      = isSelected
    ? Cesium.Color.fromCssColorString('#facc15').withAlpha(1.0)
    : conditionColor.withAlpha(roofAlpha)

  const wallColor = isSelected
    ? Cesium.Color.fromCssColorString('#fef3c7').withAlpha(1.0)
    : isProblem
      ? Cesium.Color.fromCssColorString('#f4b8b8').withAlpha(0.95)
      : Cesium.Color.fromCssColorString('#f87171').withAlpha(isBackground ? 0.3 : 1.0)

  const wallOutline = isSelected
    ? Cesium.Color.fromCssColorString('#f59e0b').withAlpha(1.0)
    : isProblem
      ? Cesium.Color.fromCssColorString('#dc2626').withAlpha(1.0)
      : Cesium.Color.fromCssColorString('#b91c1c').withAlpha(isBackground ? 0.2 : 1.0)

  const footprint = 10
  const baseH     = 7
  const roofH     = Math.max(2.5, Math.min(landHeight(house) * 0.22, 5))

  const baseEnt = viewer.entities.add({
    position: Cesium.Cartesian3.fromDegrees(lng, lat, baseH / 2),
    show: true,
    box: {
      dimensions:   new Cesium.Cartesian3(footprint, footprint, baseH),
      material:     wallColor,
      outline:      true,
      outlineColor: wallOutline,
      outlineWidth: isSelected ? 2 : isProblem ? 2 : 1.5,
    },
  })

  const roofEnt = viewer.entities.add({
    position: Cesium.Cartesian3.fromDegrees(lng, lat, baseH + roofH / 2),
    show: true,
    box: {
      dimensions:   new Cesium.Cartesian3(footprint * 0.88, footprint * 0.88, roofH),
      material:     roofColor,
      outline:      true,
      outlineColor: isSelected
        ? Cesium.Color.WHITE
        : isProblem
          ? Cesium.Color.fromCssColorString('#fff5f5').withAlpha(0.95)
          : roofColor.darken(0.25, new Cesium.Color()),
      outlineWidth: isSelected ? 2.5 : isProblem ? 2.5 : isBackground ? 0.5 : 1.5,
    },
  })

  buildingIds.add(baseEnt.id)
  buildingIds.add(roofEnt.id)
  entityMap.set(baseEnt.id, house)
  entityMap.set(roofEnt.id, house)
  return [baseEnt.id, roofEnt.id]
}

function resolveHouseForRendering(pointId, lng, lat) {
  const id = Number(pointId)
  const detailed = detailedHouseById.value.get(id)
  if (detailed) {
    const latNum = Number(detailed.latitude)
    const lngNum = Number(detailed.longitude)
    return {
      ...detailed,
      latitude: Number.isFinite(latNum) ? latNum : Number(lat),
      longitude: Number.isFinite(lngNum) ? lngNum : Number(lng),
      familyId: Number.isFinite(Number(detailed.familyId)) ? Number(detailed.familyId) : id,
    }
  }
  return toMapPointHouse({ id, lng, lat })
}

function renderClustersForCurrentView() {
  if (!viewer || viewer.isDestroyed()) return
  ensureClusterCollections()

  viewer.entities.removeAll()
  entityMap.clear()
  buildingIds.clear()
  ptCollection.removeAll()
  clusterBillboardCollection.removeAll()
  ptPrimMap.clear()

  if (!clusterIndex) {
    viewer.scene.requestRender()
    return
  }

  const bbox = getCurrentSuperclusterBBox() || [-180, -85, 180, 85]
  const height = viewer.camera.positionCartographic?.height ?? cameraHeight.value
  const zoom = getClusterZoomFromHeight(height)
  const nodes = clusterIndex.getClusters(bbox, zoom)

  const selectedId = selectedHouse.value?.familyId

  nodes.forEach((node) => {
    const [lng, lat] = node.geometry.coordinates
    if (!Number.isFinite(lng) || !Number.isFinite(lat)) return

    if (node.properties?.cluster) {
      const count = Number(node.properties.point_count || 0)
      const expansionZoom = clusterIndex.getClusterExpansionZoom(Number(node.id))
      const billboard = clusterBillboardCollection.add({
        position: Cesium.Cartesian3.fromDegrees(lng, lat, 1),
        image: generateClusterBillboardImage(count, expansionZoom),
        width: getClusterIconSize(count),
        height: getClusterIconSize(count),
        verticalOrigin: Cesium.VerticalOrigin.CENTER,
        horizontalOrigin: Cesium.HorizontalOrigin.CENTER,
        disableDepthTestDistance: Number.POSITIVE_INFINITY,
        id: {
          kind: 'cluster',
          clusterId: node.id,
          expansionZoom,
          count,
          lng,
          lat,
        },
      })
      billboard.show = true
      return
    }

    const house = resolveHouseForRendering(node.properties?.id, lng, lat)
    if (!Number.isFinite(house.familyId)) return
    addHouseModelEntity(house, lng, lat)
  })

  viewer.scene.requestRender()
}

function queueClusterRender(delay = 120) {
  clearTimeout(clusterRenderTimer)
  clusterRenderTimer = setTimeout(async () => {
    clusterRenderTimer = null
    // Load viewport data (detailed household records) BEFORE rendering
    // This ensures detailedHouseById has the totalLand field for color rendering
    await loadViewportData()
    renderClustersForCurrentView()
  }, delay)
}

function updateZoomVisibility() {
  if (!viewer || viewer.isDestroyed()) return
  const pos = viewer.camera.positionCartographic
  if (!pos) return
  const h = pos.height
  cameraHeight.value = Math.round(h)

  // ── Pitch guard: capture current pitch, enforce 3D floor ─────────────────
  const pitch = viewer.camera.pitch   // radians, negative = tilted down
  // Store every user-driven pitch for reuse by fly functions
  if (pitch >= MIN_PITCH_RAD && pitch <= MAX_PITCH_RAD) {
    currentMapPitch = pitch
  }
  if (pitch > MAX_PITCH_RAD) {
    currentMapPitch = MAX_PITCH_RAD
  }

  if (spiderfyHouseEntityIds.length && h > THRESHOLD_BUILDINGS) {
    clearSpiderfy()
  }

  // Legacy entity-based cluster/building overlays are disabled in supercluster mode.
  if (buildingIds.size) {
    buildingIds.forEach((id) => {
      const ent = viewer.entities.getById(id)
      if (ent) viewer.entities.remove(ent)
    })
    buildingIds.clear()
  }
}

function setupZoomListener() {
  if (!viewer) return
  viewer.camera.percentageChanged = 0.03
  viewer.camera.changed.addEventListener(() => {
    updateZoomVisibility()
    queueClusterRender(120)
  })
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

// ── Spiderfy: spread stacked dots so each can be clicked individually ─────────
let spiderfyEntityIds   = []   // IDs of temporary spiderfy entities (legs + houses)
let spiderfyOriginals   = []   // reserved for future positional restore flows
let spiderfyFamilyIds   = new Set()   // family IDs currently spread

function clearSpiderfy() {
  if (viewer && !viewer.isDestroyed()) {
    spiderfyEntityIds.forEach((id) => {
      const ent = viewer.entities.getById(id)
      if (ent) viewer.entities.remove(ent)
      buildingIds.delete(id)
      entityMap.delete(id)
    })
  }
  spiderfyEntityIds = []
  spiderfyHouseEntityIds.length = 0
  spiderfyFamilyIds = new Set()
  spiderfyCenter = null
}

// Returns the group of houses that share the same GPS coordinate as `house`.
// Uses a 0.0001° tolerance (~11 m) to catch floating-point duplicates.
function getSamePositionGroup(house) {
  return mapPoints.value.filter(h =>
    Math.abs(Number(h.lat) - Number(house.latitude))  < 0.0001 &&
    Math.abs(Number(h.lng) - Number(house.longitude)) < 0.0001
  )
}

function hasOverlappingCoordinates(houseGroup, precision = 6) {
  const seen = new Set()
  for (const house of houseGroup || []) {
    const lat = Number(house?.latitude)
    const lng = Number(house?.longitude)
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) continue
    const key = `${lat.toFixed(precision)},${lng.toFixed(precision)}`
    if (seen.has(key)) return true
    seen.add(key)
  }
  return false
}

async function selectHouseDetailsById(id, fallbackHouse = null, options = {}) {
  const preserveSpiderfy = !!options.preserveSpiderfy
  const numericId = Number(id)
  if (!Number.isFinite(numericId)) return

  selectedCluster.value = null
  if (!preserveSpiderfy) {
    clearSpiderfy()
  }

  try {
    const detail = await getHouseById(numericId)
    if (detail) {
      selectedHouse.value = detail
      return
    }
  } catch (error) {
    console.warn('[house-detail] fetch failed:', error?.message || error)
  }

  if (fallbackHouse) {
    selectedHouse.value = fallbackHouse
  }
}

function spiderfyCluster(clusterOrPoints, centerCartesian) {
  if (!viewer || viewer.isDestroyed()) return false

  const houses = Array.isArray(clusterOrPoints)
    ? clusterOrPoints.map((item) => {
        const pointId = Number(item?.properties?.id ?? item?.id ?? item?.properties?.house?.familyId)
        const lng = Number(item?.geometry?.coordinates?.[0] ?? item?.lng ?? item?.longitude)
        const lat = Number(item?.geometry?.coordinates?.[1] ?? item?.lat ?? item?.latitude)
        if (!Number.isFinite(pointId) || !Number.isFinite(lng) || !Number.isFinite(lat)) return null
        return resolveHouseForRendering(pointId, lng, lat)
      }).filter(Boolean)
    : []
  if (houses.length < 2) return false

  let center = centerCartesian
  if (!center) {
    const ref = houses.find((h) => Number.isFinite(Number(h?.longitude)) && Number.isFinite(Number(h?.latitude)))
    if (!ref) return false
    center = Cesium.Cartesian3.fromDegrees(Number(ref.longitude), Number(ref.latitude), 0)
  }

  const carto = Cesium.Cartographic.fromCartesian(center)
  if (!carto) return false
  const centerLat = Cesium.Math.toDegrees(carto.latitude)
  const centerLng = Cesium.Math.toDegrees(carto.longitude)

  clearSpiderfy()
  spiderfyCenter = { lat: centerLat, lng: centerLng }

  const radiusMeters = Math.max(15, Math.min(25, 15 + houses.length * 0.9))
  const offsets = getSpiderfyOffsets(houses.length, radiusMeters)

  houses.forEach((house, index) => {
    const offset = offsets[index]
    const pos = offsetCartesianDegrees(centerLat, centerLng, offset.east, offset.north)
    const offsetCartesian = Cesium.Cartesian3.fromDegrees(pos.lng, pos.lat, 0)

    const legId = addSpiderLeg(center, offsetCartesian)
    if (legId) spiderfyEntityIds.push(legId)

    const createdIds = addHouseModelEntity(house, pos.lng, pos.lat)
    createdIds.forEach((id) => {
      spiderfyEntityIds.push(id)
      spiderfyHouseEntityIds.push(id)
    })

    spiderfyFamilyIds.add(house.familyId)
  })

  viewer.scene.requestRender()
  return true
}

function spiderfyHouseGroup(houseGroup, centerLat, centerLng) {
  const center = Cesium.Cartesian3.fromDegrees(Number(centerLng), Number(centerLat), 0)
  return spiderfyCluster(houseGroup, center)
}

function spiderfyClusterLeaves(clusterId, centerLat, centerLng) {
  if (!clusterIndex) return false
  const leaves = clusterIndex.getLeaves(Number(clusterId), Infinity) || []
  const center = Cesium.Cartesian3.fromDegrees(Number(centerLng), Number(centerLat), 0)
  return spiderfyCluster(leaves, center)
}

// If `house` is one of ≥2 stacked dots, spread them in a circle and draw
// connector lines. Returns true if spiderfy was applied.
function applySpiderfy(house) {
  const group = getSamePositionGroup(house)
  if (group.length < 2) return false
  return spiderfyHouseGroup(group, Number(house.latitude), Number(house.longitude))
}

// ── Panel nudge: pan camera so selected house stays visible beside the panel ──
function nudgeCameraForPanel(house) {
  if (!viewer || sidebarCollapsed.value) return
  if (!Number.isFinite(house.longitude) || !Number.isFinite(house.latitude)) return
  // Give Cesium one render tick to settle before reading screen coordinates
  requestAnimationFrame(() => {
    if (!viewer || viewer.isDestroyed()) return
    const screenPos = viewer.scene.cartesianToCanvasCoordinates(
      Cesium.Cartesian3.fromDegrees(house.longitude, house.latitude, 0)
    )
    if (!screenPos) return
    const panelRight = 300   // sidebar width + margin (px)
    const margin     = 30
    if (screenPos.x < panelRight + margin) {
      // How many pixels the house needs to move right
      const shiftPx   = (panelRight + margin) - screenPos.x + 80
      const canvasW   = viewer.canvas.clientWidth || 800
      const camH      = viewer.camera.positionCartographic?.height ?? 5000
      // Convert pixel offset to world longitude offset (approximate)
      const hFovRad   = viewer.camera.frustum.fov || 1.0
      const degPerPx  = (hFovRad * (180 / Math.PI)) / canvasW
      viewer.camera.moveRight(-(shiftPx * degPerPx * (Math.PI / 180)) * camH)
    }
  })
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

// Group problem houses using a lat/lng grid to guarantee a bounded number of
// clusters regardless of how many houses match the filter.
// Uses a 0.04° cell (~4.4 km) so adjacent cluster labels never overlap on screen.
// Returns the top MAX_CLUSTERS largest clusters (by house count).
function computeProblemClusters(houseList) {
  const CELL_DEG    = 0.04   // ~4.4 km per cell — wide enough to prevent overlap
  const MIN_SIZE    = 6      // ignore tiny isolated groups
  const MAX_CLUSTERS = 18    // hard cap so labels never flood the screen

  const grid = new Map()
  houseList.forEach(h => {
    if (!Number.isFinite(h.latitude) || !Number.isFinite(h.longitude)) return
    const row = Math.floor(h.latitude  / CELL_DEG)
    const col = Math.floor(h.longitude / CELL_DEG)
    const key = `${row},${col}`
    if (!grid.has(key)) grid.set(key, [])
    grid.get(key).push(h)
  })

  const clusters = []
  grid.forEach(members => {
    if (members.length < MIN_SIZE) return
    const lat = members.reduce((s, h) => s + h.latitude,  0) / members.length
    const lng = members.reduce((s, h) => s + h.longitude, 0) / members.length
    clusters.push({ lat, lng, count: members.length, houses: members })
  })

  // Sort largest first, cap so labels never flood the map
  return clusters.sort((a, b) => b.count - a.count).slice(0, MAX_CLUSTERS)
}

// Group houses into a lat/lng grid. cellDeg is the grid cell size in degrees.
// Returns clusters with ≥ 2 members (centroid, count, houses).
function computeGridClusters(houseList, cellDeg) {
  const grid = new Map()
  houseList.forEach(h => {
    if (!Number.isFinite(h.latitude) || !Number.isFinite(h.longitude)) return
    const row = Math.floor(h.latitude  / cellDeg)
    const col = Math.floor(h.longitude / cellDeg)
    const key = `${row},${col}`
    if (!grid.has(key)) grid.set(key, [])
    grid.get(key).push(h)
  })
  const result = []
  grid.forEach(members => {
    if (members.length < 2) return
    const lat = members.reduce((s, h) => s + h.latitude,  0) / members.length
    const lng = members.reduce((s, h) => s + h.longitude, 0) / members.length
    result.push({ lat, lng, count: members.length, houses: members })
  })
  return result
}

// Render grid-based cluster markers (circle + count badge).
// idSet receives all created entity IDs so updateZoomVisibility can toggle them.
function addGridClusterMarkers(idSet, cellDeg, radiusM, initShow) {
  const houseList = filteredHouses.value
  const clusters  = computeGridClusters(houseList, cellDeg)
  const hasPF     = activeProblemFilters.value.length > 0

  clusters.forEach(({ lat, lng, count, houses }) => {
    const matchCnt = hasPF ? houses.filter(matchesAllProblems).length : 0
    const pct      = hasPF && count > 0 ? matchCnt / count : 0

    // Color: problem-density-weighted (red → orange → green) or colorMode-derived neutral
    const neutralColor = colorMode.value
      ? (() => {
          // Use dominant color from the cluster's houses under the current colorMode
          const colorCounts = {}
          houses.forEach(h => {
            const c = getConditionColor(h)
            colorCounts[c] = (colorCounts[c] || 0) + 1
          })
          const dominant = Object.entries(colorCounts).sort((a, b) => b[1] - a[1])[0]
          return dominant ? dominant[0] : '#3b82f6'
        })()
      : '#3b82f6'
    const hue = hasPF
      ? (pct > 0.6 ? '#ef4444' : pct > 0.3 ? '#f97316' : '#16a34a')
      : neutralColor

    const circEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, 0),
      show: initShow,
      ellipse: {
        semiMajorAxis:   radiusM,
        semiMinorAxis:   radiusM,
        material:        Cesium.Color.fromCssColorString(hue).withAlpha(0.18),
        outline:         false,
        heightReference: Cesium.HeightReference.CLAMP_TO_GROUND,
      },
    })

    const labelEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, 5),
      show: initShow,
      label: {
        text:             hasPF && matchCnt > 0 ? `${matchCnt}+` : String(count),
        font:             'bold 13px system-ui, sans-serif',
        fillColor:        Cesium.Color.WHITE,
        outlineColor:     Cesium.Color.fromCssColorString('#111827'),
        outlineWidth:     2,
        style:            Cesium.LabelStyle.FILL_AND_OUTLINE,
        verticalOrigin:   Cesium.VerticalOrigin.CENTER,
        horizontalOrigin: Cesium.HorizontalOrigin.CENTER,
        disableDepthTestDistance: Number.POSITIVE_INFINITY,
        showBackground:   true,
        backgroundColor:  Cesium.Color.fromCssColorString(hue).withAlpha(0.88),
        backgroundPadding: new Cesium.Cartesian2(9, 6),
        scaleByDistance:  new Cesium.NearFarScalar(1e3, 1.2, 5e5, 0.8),
      },
    })

    // Register for drill-down: clicking either entity flies to that cluster's houses
    zoomClusterDataMap.set(circEnt.id,  { lat, lng, houses })
    zoomClusterDataMap.set(labelEnt.id, { lat, lng, houses })

    idSet.add(circEnt.id)
    idSet.add(labelEnt.id)
  })
}

// Draw soft glow rings + a compact badge label for each problem cluster.
// Ring radius scales with cluster size. Labels use distanceDisplayCondition so
// only clusters near the camera show their text — preventing screen flooding.
// Entities sit at height=0 (BELOW house dots) for correct priority layering.
function addClusterEntities(problemHouses) {
  clusterMap.clear()
  const clusters = computeProblemClusters(problemHouses)
  const camH     = viewer.camera.positionCartographic?.height ?? cameraHeight.value
  // Rings visible whenever camera is below THRESHOLD_CLUSTER_HIDE (taluka zoom and closer)
  const initShow = camH < THRESHOLD_CLUSTER_HIDE

  clusters.forEach(({ lat, lng, count, houses }) => {
    const pos = Cesium.Cartesian3.fromDegrees(lng, lat, 0)

    const problems    = analyzeCluster(houses)
    const clusterData = { count, lat, lng, problems, houses }

    // Ring radius: 120 m base + 3 m per house, capped at 600 m.
    const baseR = Math.min(120 + count * 3, 600)

    // Soft radial glow — 3 concentric rings, no hard outline
    const glowRings = [
      { scale: 1.6,  alpha: 0.04 },
      { scale: 1.15, alpha: 0.10 },
      { scale: 0.75, alpha: 0.22 },
    ]
    glowRings.forEach(({ scale, alpha }) => {
      const r = baseR * scale
      const glowEnt = viewer.entities.add({
        position: pos,
        show: initShow,
        ellipse: {
          semiMajorAxis:   r,
          semiMinorAxis:   r,
          material:        Cesium.Color.fromCssColorString('#ef4444').withAlpha(alpha),
          outline:         false,
          heightReference: Cesium.HeightReference.CLAMP_TO_GROUND,
        },
      })
      clusterIds.add(glowEnt.id)
      clusterMap.set(glowEnt.id, clusterData)
    })

    // Badge label: visible up to 12 km (persists at building-level zoom too)
    const LABEL_MAX_DIST = 12000
    const labelEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, 5),
      show: initShow,
      label: {
        text:             `⚠ ${count} HH`,
        font:             'bold 12px system-ui, sans-serif',
        fillColor:        Cesium.Color.WHITE,
        outlineColor:     Cesium.Color.fromCssColorString('#7f1d1d'),
        outlineWidth:     2,
        style:            Cesium.LabelStyle.FILL_AND_OUTLINE,
        verticalOrigin:   Cesium.VerticalOrigin.BOTTOM,
        horizontalOrigin: Cesium.HorizontalOrigin.CENTER,
        pixelOffset:      new Cesium.Cartesian2(0, -4),
        disableDepthTestDistance: Number.POSITIVE_INFINITY,
        showBackground:   true,
        backgroundColor:  Cesium.Color.fromCssColorString('#ef4444').withAlpha(0.88),
        backgroundPadding: new Cesium.Cartesian2(6, 4),
        distanceDisplayCondition: new Cesium.DistanceDisplayCondition(0, LABEL_MAX_DIST),
        scaleByDistance: new Cesium.NearFarScalar(300, 1.15, LABEL_MAX_DIST, 0.7),
      },
    })

    clusterIds.add(labelEnt.id)
    clusterMap.set(labelEnt.id, clusterData)
  })
}

// ── Build Cesium entities ─────────────────────────────────────────────────────
// ── buildEntities: async, chunked — never blocks the main thread ─────────────
// Dot primitives (PointPrimitiveCollection) are added 2 000 at a time with a
// browser-yield between chunks.  3D building entities are NOT created here —
// they are created lazily by buildBuildingEntitiesForViewport() only when the
// user zooms in below THRESHOLD_BUILDINGS, and only for the visible subset.
async function buildEntities() {
  if (!viewer) return

  const seq = ++buildSeq

  spiderfyEntityIds = []
  spiderfyOriginals = []
  spiderfyFamilyIds = new Set()
  viewer.entities.removeAll()
  entityMap.clear()
  ptPrimMap.clear()
  jitterCache.clear()
  buildingIds.clear()
  clusterIds.clear()
  clusterMap.clear()
  macroClusIds.clear()
  miniClusIds.clear()
  zoomClusterDataMap.clear()
  houseToPointId.clear()
  selectedCluster.value = null

  ensureClusterCollections()
  buildSuperclusterIndexFromHouses()
  if (seq !== buildSeq) return
  renderClustersForCurrentView()
}

// ── Lazy 3D buildings — viewport-only, created on zoom-in ────────────────────
// Creates box entities only for houses near the camera.  At building zoom
// (< 3 500 m) the viewport covers ~50–300 houses, not 40 000, so Entity count
// stays manageable.  Called by updateZoomVisibility() on zoom-in transition.
function buildBuildingEntitiesForViewport() {
  if (!viewer || jitterCache.size === 0) return

  // Clear any previously built building entities before rebuilding
  buildingIds.forEach(id => {
    const ent = viewer.entities.getById(id)
    if (ent) viewer.entities.remove(ent)
  })
  buildingIds.clear()
  // Clear building-entity entries from entityMap (keep cluster entries)
  for (const [id, h] of entityMap) {
    if (!clusterIds.has(id) && !macroClusIds.has(id) && !miniClusIds.has(id)) {
      entityMap.delete(id)
    }
  }

  const camH = viewer.camera.positionCartographic?.height ?? cameraHeight.value
  if (camH >= THRESHOLD_BUILDINGS) return

  const pos  = viewer.camera.positionCartographic
  const cLat = Cesium.Math.toDegrees(pos.latitude)
  const cLng = Cesium.Math.toDegrees(pos.longitude)
  // Viewport half-span in degrees: generous 3× margin around camera centre.
  const span = Math.max((camH / 111_000) * 3, 0.005)

  const selectedId       = selectedHouse.value?.familyId
  const hasProblemFilter = activeProblemFilters.value.length > 0
  const houseList        = filteredHouses.value

  for (const house of houseList) {
    const coords = jitterCache.get(house.familyId)
    if (!coords) continue
    const { lat, lng } = coords
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) continue
    if (Math.abs(lat - cLat) > span || Math.abs(lng - cLng) > span) continue

    const isSelected   = house.familyId === selectedId
    const isProblem    = hasProblemFilter && matchesAllProblems(house)
    const isBackground = hasProblemFilter && !isProblem && !isSelected

    const conditionColor = cesiumColor(house)
    const roofAlpha      = isSelected ? 1.0 : isBackground ? 0.35 : 1.0
    const roofColor      = isSelected
      ? Cesium.Color.fromCssColorString('#facc15').withAlpha(1.0)
      : conditionColor.withAlpha(roofAlpha)

    const wallColor = isSelected
      ? Cesium.Color.fromCssColorString('#fef3c7').withAlpha(1.0)
      : isProblem
        ? Cesium.Color.fromCssColorString('#f4b8b8').withAlpha(0.95)
        : Cesium.Color.fromCssColorString('#f87171').withAlpha(isBackground ? 0.3 : 1.0)

    const wallOutline = isSelected
      ? Cesium.Color.fromCssColorString('#f59e0b').withAlpha(1.0)
      : isProblem
        ? Cesium.Color.fromCssColorString('#dc2626').withAlpha(1.0)
        : Cesium.Color.fromCssColorString('#b91c1c').withAlpha(isBackground ? 0.2 : 1.0)

    const footprint = 10
    const baseH     = 7
    const roofH     = Math.max(2.5, Math.min(landHeight(house) * 0.22, 5))

    const baseEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, baseH / 2),
      show: true,
      box: {
        dimensions:   new Cesium.Cartesian3(footprint, footprint, baseH),
        material:     wallColor,
        outline:      true,
        outlineColor: wallOutline,
        outlineWidth: isSelected ? 2 : isProblem ? 2 : 1.5,
      },
    })

    const roofEnt = viewer.entities.add({
      position: Cesium.Cartesian3.fromDegrees(lng, lat, baseH + roofH / 2),
      show: true,
      box: {
        dimensions:   new Cesium.Cartesian3(footprint * 0.88, footprint * 0.88, roofH),
        material:     roofColor,
        outline:      true,
        outlineColor: isSelected
          ? Cesium.Color.WHITE
          : isProblem
            ? Cesium.Color.fromCssColorString('#fff5f5').withAlpha(0.95)
            : roofColor.darken(0.25, new Cesium.Color()),
        outlineWidth: isSelected ? 2.5 : isProblem ? 2.5 : isBackground ? 0.5 : 1.5,
      },
    })

    buildingIds.add(baseEnt.id)
    buildingIds.add(roofEnt.id)
    entityMap.set(baseEnt.id, house)
    entityMap.set(roofEnt.id, house)
  }
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
    const charts = availablePieCharts.value
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
watch(colorMode, (mode) => {
  if (mode && !isColorModeEnabled(mode)) {
    colorMode.value = 'irrigation'
    return
  }

  const allowed = new Set(availableProblemFilterKeys.value)
  const next = activeProblemFilters.value.filter(key => allowed.has(key))
  if (next.length !== activeProblemFilters.value.length) {
    activeProblemFilters.value = next
  }

  const issueAllowed = new Set(availableFieldIssueKeys.value)
  if (activeIssue.value && !issueAllowed.has(activeIssue.value)) {
    activeIssue.value = null
  }

  if (viewer) buildEntities()
})

// Bridge reactive data updates to Cesium explicitly.
// Every new houses payload (initial load or viewport refresh) redraws primitives.
watch(houses, (newValue) => {
  if (!viewer) return
  if (loadingLiveData.value || viewportLoading.value) {
    viewer.scene.requestRender()
    return
  }
  // Keep currently rendered primitives when viewport payload is empty.
  if (!Array.isArray(newValue) || newValue.length === 0) {
    viewer.scene.requestRender()
    return
  }
  buildEntities()
  viewer.scene.requestRender()
}, { deep: false, flush: 'post' })

watch(selectedHouse, (house) => {
  if (viewer) buildEntities()
  if (house) loadAdvisoryForHouse(house)
})
watch(activeProblemFilters, () => { if (viewer) buildEntities() }, { deep: true })

// ── Viewport-based data loading ───────────────────────────────────────────────
// Fetches only the households visible in the current camera viewport.
// Results are cached in-memory by snapped bbox (5-min TTL, 30-tile LRU).

function clearRetryTimer() {
  if (retryTimer) { clearTimeout(retryTimer); retryTimer = null }
}

// Returns the current camera viewport as {min_lat, max_lat, min_lng, max_lng}.
// Falls back to an estimate from camera position+height if Cesium hasn't
// computed the rectangle yet (common on the very first frame).
function getCurrentViewportBbox(padDeg = 0.05) {
  if (!viewer) return null
  try {
    const rect = viewer.camera.computeViewRectangle()
    if (rect) {
      return {
        min_lat: Math.max(Cesium.Math.toDegrees(rect.south) - padDeg, -90),
        max_lat: Math.min(Cesium.Math.toDegrees(rect.north) + padDeg,  90),
        min_lng: Math.max(Cesium.Math.toDegrees(rect.west)  - padDeg, -180),
        max_lng: Math.min(Cesium.Math.toDegrees(rect.east)  + padDeg,  180),
      }
    }
    // Fallback: approximate from camera position + altitude
    const pos = viewer.camera.positionCartographic
    if (!pos) return null
    const lat     = Cesium.Math.toDegrees(pos.latitude)
    const lng     = Cesium.Math.toDegrees(pos.longitude)
    const spanDeg = Math.max((pos.height / 111_000) * 1.5, 0.5)
    return {
      min_lat: Math.max(lat - spanDeg - padDeg, -90),
      max_lat: Math.min(lat + spanDeg + padDeg,  90),
      min_lng: Math.max(lng - spanDeg - padDeg, -180),
      max_lng: Math.min(lng + spanDeg + padDeg,  180),
    }
  } catch {
    return null
  }
}

// Snap bbox to a grid so nearby viewports share a cache key.
function snapBbox(bbox, grid = 0.1) {
  return {
    min_lat: Math.floor(bbox.min_lat / grid) * grid,
    max_lat: Math.ceil(bbox.max_lat  / grid) * grid,
    min_lng: Math.floor(bbox.min_lng / grid) * grid,
    max_lng: Math.ceil(bbox.max_lng  / grid) * grid,
  }
}

function evictOldTiles() {
  const now = Date.now()
  for (const [k, v] of viewportTileCache) {
    if (now - v.ts > VIEWPORT_CACHE_TTL) viewportTileCache.delete(k)
  }
  if (viewportTileCache.size > VIEWPORT_CACHE_MAX) {
    const sorted = [...viewportTileCache.entries()].sort((a, b) => a[1].ts - b[1].ts)
    sorted.slice(0, viewportTileCache.size - VIEWPORT_CACHE_MAX).forEach(([k]) => viewportTileCache.delete(k))
  }
}

// ── Initial load — runs once on mount ────────────────────────────────────────
// Fetches the first page of households with NO bbox constraint so we always
// get data on the first render, regardless of camera state.  After this
// completes, isInitialLoadDone = true and the debounce-driven loadViewportData
// takes over for all subsequent camera pan/zoom refreshes.
async function loadInitialData() {
  console.log('[initial] start')
  loadingLiveData.value = true
  viewportLoading.value = false
  showEmptyViewportHint.value = false

  let attempt = 0
  while (attempt <= 5) {
    try {
      const res = await getHousesMapPoints({
        district_id: filterDistrict.value || undefined,
        taluka_id: filterTaluka.value || undefined,
        village_id: filterVillage.value || undefined,
      })

      const points = Array.isArray(res) ? res : []
      mapPoints.value = points
        .map((point) => ({
          id: Number(point?.id),
          lat: Number(point?.lat),
          lng: Number(point?.lng),
        }))
        .filter((point) => Number.isFinite(point.id) && Number.isFinite(point.lat) && Number.isFinite(point.lng))

      console.log('[initial] map points loaded — records:', mapPoints.value.length)

      // Snapshot the current bbox so camera-change debounce starts from current view.
      lastLoadedBbox    = getCurrentViewportBbox()
      isInitialLoadDone = true

      buildSuperclusterIndexFromHouses()
      renderClustersForCurrentView()
      return

    } catch (err) {
      attempt++
      console.error('[initial] map points fetch failed attempt', attempt, err?.message || err)
      if (attempt > 5) break
      await new Promise(r => setTimeout(r, 1200 * attempt))
    }
  }

  console.error('[initial] all retries failed')
  isInitialLoadDone = true
}

// Wrapped version so loading is always cleared via finally
async function loadInitialDataWithCleanup() {
  try {
    await loadInitialData()
    // Prime detailed household rows for counters/problem filters on first view.
    await loadViewportData()
  } finally {
    loadingLiveData.value = false
    viewportLoading.value = false
    console.log('[initial] loading cleared — mapPoints:', mapPoints.value.length)
  }
}

// ── Viewport load — runs on camera pan/zoom after initial load ────────────────
// Fetches households for the current camera bbox only.  Results are cached
// in-memory by snapped bbox tile key (5-min TTL, 30-tile LRU).
async function loadViewportData() {
  if (!isInitialLoadDone) {
    console.warn('[viewport] called before initial load — ignored')
    return
  }

  const seq = ++viewportSeq
  vpInFlight++
  viewportLoading.value = true
  showEmptyViewportHint.value = false
  console.log('[viewport] start seq=', seq, 'inFlight=', vpInFlight)

  try {
    const bbox = getCurrentViewportBbox()
    if (!bbox) {
      console.warn('[viewport] no bbox — skipping')
      return
    }

    const camH    = viewer?.camera?.positionCartographic?.height ?? 0
    const snapped = snapBbox(bbox, camH > THRESHOLD_DOTS ? 0.2 : 0.05)
    const cacheKey = [
      snapped.min_lat, snapped.max_lat, snapped.min_lng, snapped.max_lng,
      filterDistrict.value, filterTaluka.value, filterVillage.value,
    ].join('|')

    // Cache hit
    const cached = viewportTileCache.get(cacheKey)
    if (cached && Date.now() - cached.ts < VIEWPORT_CACHE_TTL) {
      console.log('[viewport] cache hit:', cached.data.length, 'records')
      if (seq === viewportSeq) {
        lastLoadedBbox = bbox
        showEmptyViewportHint.value = cached.data.length === 0
        if (cached.data.length > 0) {
          houses.value = cached.data
        }
      }
      return
    }

    // Live fetch
    const params = { ...snapped, limit: 2000 }
    if (filterDistrict.value) params.district_id = filterDistrict.value
    if (filterTaluka.value)   params.taluka_id   = filterTaluka.value
    if (filterVillage.value)  params.village_id  = filterVillage.value

    console.log('[viewport] fetching bbox:', snapped)
    let res
    try {
      res = await getHousesByViewport(snapped, params)
    } catch (fetchErr) {
      console.error('[viewport] fetch error:', fetchErr?.message || fetchErr)
      return
    }
    console.log('[viewport] fetch done — records:', res?.data?.length)

    if (seq !== viewportSeq) {
      console.log('[viewport] stale — discarded')
      return
    }

    const data = Array.isArray(res?.data) ? res.data : []
    viewportTileCache.set(cacheKey, { ts: Date.now(), data })
    evictOldTiles()
    lastLoadedBbox = bbox
    showEmptyViewportHint.value = data.length === 0
    if (data.length > 0) {
      houses.value = data
    }
    console.log('[viewport] applied:', data.length, 'records')

  } finally {
    vpInFlight--
    viewportLoading.value = vpInFlight > 0
    console.log('[viewport] finally seq=', seq, 'inFlight=', vpInFlight, 'loading=', viewportLoading.value)
  }
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
      imageryProvider:      false,
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

    viewer.imageryLayers.removeAll()
    viewer.imageryLayers.addImageryProvider(buildImageryProvider('satellite'))

    viewer.scene.backgroundColor              = Cesium.Color.fromCssColorString('#0c1a2e')
    viewer.scene.globe.baseColor              = Cesium.Color.fromCssColorString('#4a7c59')
    viewer.scene.globe.enableLighting         = false
    viewer.scene.globe.showGroundAtmosphere   = false
    viewer.scene.fog.enabled                  = false
    viewer.scene.globe.depthTestAgainstTerrain = false

    // Allow closer inspection without camera bounce.
    viewer.scene.screenSpaceCameraController.minimumZoomDistance = 60

    // ── Pitch constraints — keep 3D perspective locked ────────────────────────
    // Cesium uses enableCollisionDetection-style tilt limits via minimumZoomDistance
    // but pitch min/max must be enforced through the camera-changed guard above.
    // Set the controller's tilt range so mouse drag can't go fully overhead.
    const ctrl = viewer.scene.screenSpaceCameraController
    ctrl.minimumZoomDistance = 60
    ctrl.enableCollisionDetection = false
    // Allow generous tilt via mouse without collision-driven camera pushback

    // Start at a neutral holding position — loadViewportData() will fly
    // to the actual household bounding box as soon as data arrives.
    viewer.camera.setView({
      destination: Cesium.Cartesian3.fromDegrees(76.0, 19.5, 220000),
      orientation: { heading: 0, pitch: Cesium.Math.toRadians(-48), roll: 0 },
    })

    // ── Resolve a scene.pick() result to a house object ─────────────────────────
    // PointPrimitiveCollection: picked.primitive is a PointPrimitive with .id = house
    // Entity API (buildings, clusters):  picked.id is the Entity object, .id.id is UUID
    function resolvePickedHouse(picked) {
      if (!Cesium.defined(picked)) return null
      if (picked.primitive instanceof Cesium.Billboard) {
        const payload = picked.id
        if (payload && typeof payload === 'object' && payload.kind === 'cluster') return null
      }
      if (picked.primitive instanceof Cesium.PointPrimitive) {
        const payload = picked.id
        if (payload && typeof payload === 'object' && payload.kind === 'cluster') return null
        return payload ?? null
      }
      if (picked.id) return entityMap.get(picked.id.id) ?? null
      return null
    }

    // Click → select house | drill-down cluster | open problem panel | clear
    viewer.screenSpaceEventHandler.setInputAction(async (e) => {
      const picked = viewer.scene.pick(e.position)

      // ── Cluster billboard click → expansion zoom fly-to ───────────────────
      if (Cesium.defined(picked) && picked.primitive instanceof Cesium.Billboard) {
        const payload = picked.id
        if (payload && typeof payload === 'object' && payload.kind === 'cluster') {
          const expansionZoom = Number(payload.expansionZoom ?? clusterIndex?.getClusterExpansionZoom?.(Number(payload.clusterId)) ?? 12)
          const maxClusterZoom = 19

          if (expansionZoom >= maxClusterZoom) {
            const leaves = clusterIndex?.getLeaves?.(Number(payload.clusterId), Infinity) || []
            if (leaves.length >= 2) {
              const centerCartesian = Cesium.Cartesian3.fromDegrees(Number(payload.lng), Number(payload.lat), 0)
              spiderfyCluster(leaves, centerCartesian)
              return
            }
          }

          const targetH = getHeightFromClusterZoom(expansionZoom)
          viewer.camera.flyTo({
            destination: Cesium.Cartesian3.fromDegrees(Number(payload.lng), Number(payload.lat), targetH),
            orientation: {
              heading: viewer.camera.heading,
              pitch: Math.max(Math.min(currentMapPitch, MAX_PITCH_RAD), MIN_PITCH_RAD),
              roll: 0,
            },
            duration: 0.9,
            easingFunction: Cesium.EasingFunction.QUADRATIC_IN_OUT,
          })
          queueClusterRender(140)
          return
        }
      }

      // ── PointPrimitive (household dot) ───────────────────────────────────────
      if (Cesium.defined(picked) && picked.primitive instanceof Cesium.PointPrimitive) {
        const payload = picked.id
        if (payload && typeof payload === 'object' && payload.kind === 'cluster') {
          const nextH = Math.max((viewer.camera.positionCartographic?.height ?? 120000) * 0.45, 700)
          viewer.camera.flyTo({
            destination: Cesium.Cartesian3.fromDegrees(Number(payload.lng), Number(payload.lat), nextH),
            orientation: {
              heading: viewer.camera.heading,
              pitch: Math.max(Math.min(currentMapPitch, MAX_PITCH_RAD), MIN_PITCH_RAD),
              roll: 0,
            },
            duration: 0.7,
            easingFunction: Cesium.EasingFunction.QUADRATIC_OUT,
          })
          queueClusterRender(120)
          return
        }

        const house = payload
        if (house) {
          const isAlreadySpiderfied = spiderfyFamilyIds.has(house.familyId)
          if (!isAlreadySpiderfied) {
            const wasSpiderfied = applySpiderfy(house)
            if (wasSpiderfied && getSamePositionGroup(house).length >= 2) return
          }
          await selectHouseDetailsById(house.familyId, house, { preserveSpiderfy: isAlreadySpiderfied })
          selectedCluster.value = null
          nudgeCameraForPanel(selectedHouse.value || house)
        }
        return
      }

      if (Cesium.defined(picked) && picked.id) {
        const entityId = picked.id.id

        // ── Zoom-cluster drill-down: fly into that cluster's houses ──────────
        const zCluster = zoomClusterDataMap.get(entityId)
        if (zCluster) {
          clearSpiderfy()
          flyToPoints(zCluster.houses)
          return
        }

        // ── House building entity (close zoom 3D box) ────────────────────────
        const house = entityMap.get(entityId)
        if (house) {
          const isAlreadySpiderfied = spiderfyFamilyIds.has(house.familyId)
          if (!isAlreadySpiderfied) {
            const wasSpiderfied = applySpiderfy(house)
            if (wasSpiderfied && getSamePositionGroup(house).length >= 2) return
          }
          await selectHouseDetailsById(house.familyId, house, { preserveSpiderfy: isAlreadySpiderfied })
          selectedCluster.value = null
          nudgeCameraForPanel(selectedHouse.value || house)
          return
        }

        // ── Problem-cluster entity → open Group Action Card ──────────────────
        const cluster = clusterMap.get(entityId)
        if (cluster) {
          clearSpiderfy()
          let resolvedCluster = cluster
          if (activeProblemFilters.value.length > 0 && (!cluster.problems || cluster.problems.length === 0)) {
            const syntheticProblems = activeProblemFilters.value.map(key => {
              const meta = CLUSTER_PROBLEM_META.find(m => m.key === key) || { key, label: key }
              const count = cluster.houses.filter(h => matchesProblemFilter(h, key)).length
              const pct = cluster.count > 0 ? Math.round((count / cluster.count) * 100) : 0
              return { ...meta, count, total: cluster.count, pct }
            }).filter(p => p.count > 0)
            resolvedCluster = { ...cluster, problems: syntheticProblems }
          }
          selectedCluster.value = resolvedCluster
          selectedHouse.value   = null
          clusterAdvisory.value = null
          highlightClusterBoundary(resolvedCluster)
          loadClusterAdvisory(resolvedCluster)
          if (resolvedCluster.houses && resolvedCluster.houses.length > 0) {
            const lats = resolvedCluster.houses.map(h => h.latitude).filter(Number.isFinite)
            const lngs = resolvedCluster.houses.map(h => h.longitude).filter(Number.isFinite)
            if (lats.length && lngs.length) {
              const south = Math.min(...lats) - 0.0005
              const north = Math.max(...lats) + 0.0005
              const west  = Math.min(...lngs) - 0.0005
              const east  = Math.max(...lngs) + 0.0005
              viewer.camera.flyTo({
                destination: Cesium.Rectangle.fromDegrees(west, south, east, north),
                orientation: {
                  heading: viewer.camera.heading,
                  pitch: Math.max(Math.min(currentMapPitch, MAX_PITCH_RAD), MIN_PITCH_RAD),
                  roll: 0,
                },
                duration: 1.2,
                easingFunction: Cesium.EasingFunction.QUADRATIC_IN_OUT,
              })
            }
          }
          return
        }
      }

      // Clicked empty space → clear everything
      clearSpiderfy()
      selectedHouse.value   = null
      selectedCluster.value = null
    }, Cesium.ScreenSpaceEventType.LEFT_CLICK)

    // Double-click → zoom to house
    viewer.screenSpaceEventHandler.setInputAction((e) => {
      const picked = viewer.scene.pick(e.position)
      const house  = resolvePickedHouse(picked)
      if (house) flyToHouse(house)
    }, Cesium.ScreenSpaceEventType.LEFT_DOUBLE_CLICK)

    // Hover → tooltip
    viewer.screenSpaceEventHandler.setInputAction((e) => {
      mouseX.value = e.endPosition.x + 16
      mouseY.value = e.endPosition.y + 12
      const picked = viewer.scene.pick(e.endPosition)
      hoveredHouse.value = resolvePickedHouse(picked)
    }, Cesium.ScreenSpaceEventType.MOUSE_MOVE)

    setupZoomListener()

  } catch (err) {
    console.warn('Cesium init failed:', err)
  }

  // Load insights in parallel (small, fast); initial house data fetched separately
  getAgricultureInsights().then(v => { agricultureInsights.value = v }).catch(() => {})
  getPopulationDashboard().then(v => { populationDashboard.value  = v }).catch(() => {})
  await loadLocationOptions()
  loadInitialDataWithCleanup()

  // Safety net: force-clear loading if still stuck after 30 s
  setTimeout(() => {
    if (loadingLiveData.value || viewportLoading.value) {
      console.error('[initial] loading timeout — forcing clear')
      loadingLiveData.value = false
      viewportLoading.value = false
      isInitialLoadDone     = true
    }
  }, 30_000)

  setTimeout(handleResize, 60)
  setTimeout(handleResize, 300)
  window.addEventListener('resize', handleResize)
  // Close any open custom dropdown when clicking outside
  window.addEventListener('click', closeDropdowns)
  document.addEventListener('fullscreenchange', handleTwinFullscreenChange)
})

onUnmounted(() => {
  clearRetryTimer()
  if (clusterRenderTimer) clearTimeout(clusterRenderTimer)
  if (viewportDebounce) clearTimeout(viewportDebounce)
  if (buildingPanTimer) clearTimeout(buildingPanTimer)
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('click', closeDropdowns)
  document.removeEventListener('fullscreenchange', handleTwinFullscreenChange)
  if (viewer && !viewer.isDestroyed()) {
    viewer.camera.changed.removeEventListener(updateZoomVisibility)
    viewer.destroy()
  }
  viewer = null
  ptCollection = null
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

.fs-legend {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 9998;
  width: 235px;
  max-height: calc(100vh - 24px);
  overflow-y: auto;
  background: rgba(255,255,255,0.82);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border: 1.5px solid rgba(209,213,219,0.6);
  border-radius: 10px;
  padding: 0.55rem 0.65rem;
  box-shadow: 0 8px 22px rgba(0,0,0,0.14), 0 3px 8px rgba(0,0,0,0.08);
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

.cs-trigger-placeholder .cs-value {
  color: #9ca3af;
  font-style: italic;
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
  z-index: 10000;         /* above detail panel (9990) and fs-legend (9998) */
  scrollbar-width: thin;
  scrollbar-color: #d1d5db transparent;
}
/* Right-aligned variant for "Color by" which sits near the right edge */
.cs-dropdown-right {
  left: auto;
  right: 0;
}

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
/* Map background-loading overlay: bottom-centre pill, non-blocking */
.loading-overlay.map-bg-loading-overlay {
  inset: auto;
  bottom: 48px;
  left: 50%;
  transform: translateX(-50%);
  width: auto;
  padding: 10px 24px;
  border-radius: 20px;
  background: rgba(12, 26, 46, 0.88);
  backdrop-filter: blur(6px);
  border: 1px solid rgba(255,255,255,0.12);
  flex-direction: column;
  gap: 4px;
  z-index: 510;
}
.loading-overlay.map-bg-loading-overlay .loading-spinner {
  width: 16px; height: 16px;
  border-width: 2px;
  border-color: rgba(255,255,255,0.25);
  border-top-color: #4ade80;
}
.loading-overlay.map-bg-loading-overlay .loading-text {
  color: #e2e8f0;
  font-size: 0.8rem;
  white-space: nowrap;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}
.map-bg-progress {
  color: #86efac;
  font-size: 0.72rem;
  font-variant-numeric: tabular-nums;
}
/* Centering overlay: anchored to bottom-centre, non-blocking — map is visible behind it */
.loading-overlay.centering-overlay {
  inset: auto;
  bottom: 48px;
  left: 50%;
  transform: translateX(-50%);
  width: auto;
  padding: 8px 20px;
  border-radius: 20px;
  background: rgba(12, 26, 46, 0.82);
  backdrop-filter: blur(6px);
  border: 1px solid rgba(255,255,255,0.12);
  flex-direction: row;
  gap: 10px;
}
.loading-overlay.centering-overlay .loading-spinner {
  width: 16px; height: 16px;
  border-width: 2px;
  border-color: rgba(255,255,255,0.25);
  border-top-color: #4ade80;
}
.loading-overlay.centering-overlay .loading-text {
  color: #e2e8f0;
  font-size: 0.78rem;
  white-space: nowrap;
}
.loading-spinner {
  width: 36px; height: 36px; border-radius: 50%;
  border: 3px solid #e2e8f0;
  border-top-color: #16a34a;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.loading-text { font-size: 0.84rem; color: #374151; font-weight: 500; }

.map-empty-toast {
  position: absolute;
  left: 50%;
  bottom: 14px;
  transform: translateX(-50%);
  z-index: 520;
  pointer-events: none;
  user-select: none;
  white-space: nowrap;
  background: rgba(12, 26, 46, 0.84);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: #e2e8f0;
  border-radius: 16px;
  padding: 7px 14px;
  font-size: 0.76rem;
  font-weight: 500;
  backdrop-filter: blur(5px);
  box-shadow: 0 4px 10px rgba(0,0,0,0.2);
}

/* ═══════════════════════════════════════════════
   LEFT SIDEBAR
═══════════════════════════════════════════════ */
.sidebar {
  position: absolute;
  left: 0.75rem;
  top: 70px;
  z-index: 100;
  width: 240px;
  overflow: visible;
  scrollbar-width: thin;
  scrollbar-color: var(--c-border) transparent;
  transition: width 0.22s ease;
}
.sidebar.collapsed { width: 20px; overflow: visible; }

.sidebar-toggle {
  position: absolute;
  right: -13px; top: 2px;
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
  display: flex; flex-direction: column; gap: 0.65rem;
  max-height: calc(100vh - 92px);
  overflow-y: auto;
  overflow-x: visible;
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

.card-title-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  margin-bottom: 0;
}

.card-toggle-icon {
  font-size: 0.7rem;
  color: #6b7280;
  transition: transform 0.16s ease;
}

.card-toggle-icon.open {
  transform: rotate(180deg);
}

.agri-collapsible-body {
  margin-top: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid #e5e7eb;
}

.agri-collapse-enter-active,
.agri-collapse-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.agri-collapse-enter-from,
.agri-collapse-leave-to {
  opacity: 0;
  transform: translateY(-2px);
}
.empty-note { font-size: 0.73rem; color: #6b7280; line-height: 1.5; padding: 0.2rem 0; }

/* ── Legend ── */
.legend-item   { display: flex; align-items: center; gap: 0.55rem; margin-bottom: 0.38rem; }
.legend-text   { font-size: 0.7rem; color: #111827; font-weight: 500; }
.legend-note   { font-size: 0.62rem; color: #9ca3af; margin-top: 0.45rem; font-style: italic; }
.legend-count-pill {
  display: inline-block;
  margin-left: 0.4rem;
  padding: 0.05rem 0.45rem;
  border-radius: 999px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  font-size: 0.68rem;
  font-weight: 600;
  color: #475569;
  vertical-align: middle;
}

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
/* ── Scheme drawer ── */
.scheme-drawer { padding: 0.6rem 0.7rem; }

.scheme-header {
  display: flex; align-items: center; gap: 0.35rem;
  margin-bottom: 0.6rem; flex-wrap: wrap;
}
.scheme-header-icon  { font-size: 0.85rem; }
.scheme-header-text  { font-size: 0.68rem; font-weight: 700; color: #1e40af; flex: 1; }
.scheme-source-tag   { font-size: 0.6rem; font-weight: 700; padding: 0.1rem 0.4rem;
                       border-radius: 999px; letter-spacing: 0.03em; }
.tag-db  { background: #dcfce7; color: #15803d; }
.tag-fb  { background: #fef9c3; color: #92400e; }

.scheme-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-left: 3px solid;
  border-radius: 0 6px 6px 0;
  padding: 0.55rem 0.6rem;
  margin-bottom: 0.5rem;
  box-shadow: 0 1px 4px rgba(0,0,0,0.05);
}
.scheme-card-name {
  font-size: 0.72rem; font-weight: 700; color: #111827;
  margin-bottom: 0.25rem; line-height: 1.35;
}
.scheme-card-desc {
  font-size: 0.67rem; color: #4b5563; line-height: 1.45; margin-bottom: 0.35rem;
}
.scheme-card-row { margin-bottom: 0.2rem; }
.scheme-tag {
  display: inline-block; font-size: 0.63rem; font-weight: 600;
  padding: 0.12rem 0.45rem; border-radius: 4px; line-height: 1.4;
}
.scheme-tag-benefit     { background: #f0fdf4; color: #15803d; border: 1px solid #bbf7d0; }
.scheme-tag-eligibility { background: #eff6ff; color: #1d4ed8; border: 1px solid #bfdbfe; }
.scheme-card-reason {
  font-size: 0.62rem; font-weight: 600; margin-top: 0.3rem;
  display: flex; align-items: flex-start; gap: 0.25rem; line-height: 1.35;
}
.scheme-reason-icon { flex-shrink: 0; }
.scheme-empty {
  font-size: 0.68rem; color: #6b7280; font-style: italic; padding: 0.3rem 0;
}
.scheme-loading {
  display: flex; align-items: center; gap: 0.4rem;
  font-size: 0.68rem; color: #6b7280; padding: 0.3rem 0;
}
.scheme-spinner {
  width: 12px; height: 12px; border: 2px solid #d1d5db;
  border-top-color: #2563eb; border-radius: 50%;
  animation: spin 0.7s linear infinite; flex-shrink: 0;
}
@keyframes spin { to { transform: rotate(360deg); } }

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
  z-index: 9990;
  background: rgba(255,255,255,0.92);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1.5px solid rgba(226,232,240,0.75);
  border-radius: 12px;
  box-shadow: 0 12px 40px rgba(0,0,0,0.14), 0 4px 12px rgba(0,0,0,0.07);
  scrollbar-width: thin;
  scrollbar-color: #cbd5e1 transparent;
}

.detail-panel-fs {
  top: 12px;
  right: 12px;
  max-height: calc(100vh - 24px);
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
.dp-empty-note {
  margin: 0.45rem 1rem 0;
  padding: 0.5rem 0.6rem;
  border-radius: 8px;
  border: 1px dashed #d1d5db;
  background: #f8fafc;
  color: #64748b;
  font-size: 0.69rem;
  font-weight: 600;
  line-height: 1.45;
}

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

/* ── Advisory cards (DB-driven) ── */
.advisory-loading {
  display: flex; align-items: center; gap: 0.5rem;
  font-size: 0.7rem; color: #6b7280; padding: 0.5rem 1rem 0.8rem;
}
.advisory-spinner {
  width: 13px; height: 13px; border: 2px solid #e5e7eb;
  border-top-color: #16a34a; border-radius: 50%;
  animation: spin 0.7s linear infinite; flex-shrink: 0;
}

.advisory-card {
  border-left: 3px solid; border-radius: 0 8px 8px 0;
  background: #f8fafc;
  border-top: 1px solid #e2e8f0;
  border-right: 1px solid #e2e8f0;
  border-bottom: 1px solid #e2e8f0;
  padding: 0.65rem 0.75rem;
  margin-bottom: 0.55rem;
}
.advisory-title-row {
  display: flex; align-items: center; gap: 0.4rem;
  flex-wrap: wrap; margin-bottom: 0.42rem;
}
.advisory-title { font-size: 0.8rem; font-weight: 700; flex: 1; }
.advisory-crop-tag {
  font-size: 0.6rem; font-weight: 700;
  padding: 0.1rem 0.4rem; border-radius: 4px;
  background: #fef9c3; color: #92400e; border: 1px solid #fde68a;
  text-transform: capitalize;
}
.advisory-row   { display: flex; gap: 0.48rem; margin-bottom: 0.32rem; align-items: flex-start; }
.advisory-tag {
  font-size: 0.59rem; text-transform: uppercase; letter-spacing: 0.05em;
  padding: 0.14rem 0.4rem; border-radius: 3px;
  font-weight: 800; flex-shrink: 0; margin-top: 0.04rem;
}
.advisory-tag.cause    { background: #fef2f2; color: #dc2626; border: 1px solid #fca5a5; }
.advisory-tag.solution { background: #f0fdf4; color: #15803d; border: 1px solid #86efac; }
.advisory-text  { font-size: 0.73rem; color: #374151; line-height: 1.55; }

/* Scheme + source footer */
.advisory-footer {
  display: flex; align-items: center; gap: 0.4rem;
  flex-wrap: wrap; margin-top: 0.42rem;
}
.advisory-scheme-pill {
  display: inline-flex; align-items: center; gap: 0.25rem;
  font-size: 0.63rem; font-weight: 700;
  padding: 0.14rem 0.5rem; border-radius: 999px;
  border: 1.5px solid; background: #ffffff;
  line-height: 1.4; flex: 1; min-width: 0;
}
.pill-gov  { color: #1e40af; }
.pill-tech { color: #6b21a8; }
.pill-icon { font-size: 0.7rem; }
.advisory-source-tag {
  font-size: 0.58rem; font-weight: 700; letter-spacing: 0.03em;
  padding: 0.1rem 0.4rem; border-radius: 999px; flex-shrink: 0;
}
.src-db     { background: #dcfce7; color: #15803d; }
.src-scheme { background: #dbeafe; color: #1d4ed8; }
.src-curated{ background: #fef9c3; color: #92400e; }

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
  padding: 1rem 1rem 0.8rem;
  background: linear-gradient(135deg, #fef2f2 0%, #fff7f7 100%);
  border-bottom: 1.5px solid #fecaca;
  border-radius: var(--radius) var(--radius) 0 0;
}

/* Priority badge — replaces old .cluster-badge */
.cluster-priority-badge {
  display: inline-block;
  font-size: 0.67rem; font-weight: 700;
  padding: 0.22rem 0.65rem; border-radius: 20px;
  letter-spacing: 0.04em; margin-bottom: 0.42rem;
}
.badge-high     { background: #ef4444; color: #fff; }
.badge-moderate { background: #f59e0b; color: #fff; }

.cluster-count {
  font-size: 0.82rem; color: #374151; line-height: 1.3;
}
.cluster-count strong { color: #ef4444; font-size: 1.1rem; }

.cluster-location-pill {
  font-size: 0.62rem; color: #6b7280; margin-top: 0.3rem;
}

.cluster-loading {
  display: flex; align-items: center; gap: 0.5rem;
  font-size: 0.7rem; color: #6b7280; padding: 0.8rem 1rem;
}

.cluster-section-title {
  font-size: 0.62rem; font-weight: 800;
  text-transform: uppercase; letter-spacing: 0.1em;
  color: #374151; padding: 0.75rem 1rem 0.3rem;
}

/* Group problem card */
.cp-group-card {
  margin: 0 0.75rem 0.65rem;
  padding: 0.65rem 0.75rem;
  background: #fafafa;
  border: 1px solid #e5e7eb;
  border-left: 3px solid #ef4444;
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
}
.cp-group-card.cp-mass {
  background: #fff5f5;
  border-color: #fca5a5;
  border-left-color: #dc2626;
  border-left-width: 4px;
}

.cp-mass-heading {
  font-size: 0.72rem; font-weight: 800; color: #dc2626;
  margin-bottom: 0.48rem; letter-spacing: 0.01em;
}

.cp-top {
  display: flex; align-items: flex-start; gap: 0.5rem; margin-bottom: 0.42rem;
}
.cp-emoji { font-size: 1.2rem; flex-shrink: 0; line-height: 1.2; }
.cp-info  { flex: 1; min-width: 0; }
.cp-label { display: block; font-size: 0.78rem; font-weight: 700; color: #111827; }
.cp-stat  { font-size: 0.64rem; color: #6b7280; margin-top: 0.08rem; display: block; }
.cp-pct-badge {
  font-size: 0.68rem; font-weight: 800;
  padding: 0.08rem 0.38rem; border-radius: 4px; flex-shrink: 0;
}
.pct-red   { background: #fef2f2; color: #dc2626; }
.pct-amber { background: #fffbeb; color: #d97706; }

.cp-bar-track {
  height: 5px; background: #f3f4f6; border-radius: 3px;
  overflow: hidden; margin-bottom: 0.52rem;
}
.cp-bar-fill { height: 100%; border-radius: 3px; transition: width 0.5s ease; }
.fill-red   { background: linear-gradient(90deg, #ef4444, #f87171); }
.fill-amber { background: linear-gradient(90deg, #f59e0b, #fbbf24); }

/* Cause / Action rows */
.cp-cause-row, .cp-action-row {
  display: flex; gap: 0.4rem; align-items: flex-start; margin-bottom: 0.38rem;
}
.cp-tag {
  font-size: 0.56rem; font-weight: 800; text-transform: uppercase;
  letter-spacing: 0.06em; padding: 0.12rem 0.35rem;
  border-radius: 3px; flex-shrink: 0; margin-top: 0.04rem;
}
.cp-tag-cause  { background: #fef2f2; color: #dc2626; border: 1px solid #fca5a5; }
.cp-tag-action { background: #f0fdf4; color: #15803d; border: 1px solid #86efac; }
.cp-cause-text, .cp-action-text {
  font-size: 0.7rem; color: #374151; line-height: 1.55;
}

/* Scheme footer */
.cp-scheme-footer {
  display: flex; flex-wrap: wrap; gap: 0.3rem; margin-top: 0.42rem;
  align-items: center;
}
.cp-scheme-pill {
  display: inline-flex; align-items: center; gap: 0.2rem;
  font-size: 0.62rem; font-weight: 700;
  padding: 0.14rem 0.45rem; border-radius: 999px;
  border: 1.5px solid; flex: 1; min-width: 0;
}
.pill-community { color: #0f766e; border-color: #99f6e4; background: #f0fdfa; }
.pill-gov       { color: #1e40af; border-color: #bfdbfe; background: #eff6ff; }
.cp-benefit-pill {
  font-size: 0.6rem; font-weight: 600;
  padding: 0.1rem 0.4rem; border-radius: 999px;
  background: #f0fdf4; color: #15803d; border: 1px solid #bbf7d0;
}
.cp-source-tag {
  font-size: 0.58rem; font-weight: 700; letter-spacing: 0.03em;
  padding: 0.1rem 0.4rem; border-radius: 999px;
}

/* Drill-down button */
.cp-drill-row {
  margin: 0.5rem 0.75rem 0.85rem;
  display: flex; flex-direction: column; gap: 0.3rem;
}
.cp-drill-btn {
  background: #1e40af; color: #fff;
  border: none; border-radius: var(--radius-sm);
  font-size: 0.73rem; font-weight: 700;
  padding: 0.5rem 0.9rem; cursor: pointer;
  width: 100%; text-align: center;
  transition: background 0.15s;
}
.cp-drill-btn:hover { background: #1d4ed8; }
.cp-drill-hint {
  font-size: 0.6rem; color: #9ca3af; text-align: center; font-style: italic;
}

.cluster-ok {
  font-size: 0.73rem; color: #15803d;
  padding: 0.75rem 1rem 1rem; font-weight: 500;
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

.pf-list {
  min-height: 88px;
}

.fi-list,
.agri-list {
  min-height: 70px;
}

.section-context-label {
  font-size: 0.64rem;
  color: #6b7280;
  margin: 0.1rem 0 0.4rem;
}

.pf-context-label {
  font-size: 0.64rem;
  color: #6b7280;
  margin: 0.1rem 0 0.4rem;
}

.pf-fade-enter-active,
.pf-fade-leave-active {
  transition: opacity 0.16s ease, transform 0.16s ease;
}
.pf-fade-enter-from,
.pf-fade-leave-to {
  opacity: 0;
  transform: translateY(2px);
}

.fi-fade-enter-active,
.fi-fade-leave-active,
.agri-fade-enter-active,
.agri-fade-leave-active {
  transition: opacity 0.16s ease, transform 0.16s ease;
}

.fi-fade-enter-from,
.fi-fade-leave-to,
.agri-fade-enter-from,
.agri-fade-leave-to {
  opacity: 0;
  transform: translateY(2px);
}

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
  /* On small screens stack the detail panel below the topbar so it never overlaps the View By dropdown */
  .detail-panel  { width: calc(100vw - 1.5rem); right: 0.75rem; top: 56px; max-height: 55vh; }
  .topbar        { gap: 0.5rem; }
  .brand-sub     { display: none; }
  .map-fs-btn.shifted { right: 12px; }
}

/* On medium screens, shrink detail panel width so it doesn't extend under the topbar right controls */
@media (min-width: 701px) and (max-width: 1100px) {
  .detail-panel { width: 280px; }
}
</style>
