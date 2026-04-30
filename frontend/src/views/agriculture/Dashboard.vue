<template>
  <div class="dashboard">
    <header class="page-header">
      <div>
        <h1 class="page-title">Village Command Center</h1>
        <p class="page-subtitle">Unified population and agriculture intelligence</p>
      </div>
    </header>

    <section class="card dashboard-filter">
      <div class="dashboard-filter-head">Location Filter</div>
      <div class="dashboard-filter-grid">
        <select v-model="selectedDistrict" class="dashboard-filter-select" @change="onDistrictChange">
          <option value="">Select District</option>
          <option v-for="district in districtOptions" :key="district.id" :value="String(district.id)">
            {{ district.name }}
          </option>
        </select>

        <select
          v-model="selectedTaluka"
          class="dashboard-filter-select"
          :disabled="!selectedDistrict"
          @change="onTalukaChange"
        >
          <option value="">Select Taluka</option>
          <option v-for="taluka in talukaOptions" :key="taluka.id" :value="String(taluka.id)">
            {{ taluka.name }}
          </option>
        </select>

        <select
          v-model="selectedVillage"
          class="dashboard-filter-select"
          :disabled="!selectedTaluka"
        >
          <option value="">Select Village</option>
          <option v-for="village in villageOptions" :key="village.id" :value="String(village.id)">
            {{ village.name }}
          </option>
        </select>

        <div class="dashboard-filter-actions">
          <button type="button" class="dashboard-apply-btn" @click="applyLocationFilters">Apply</button>
          <button type="button" class="dashboard-reset-btn" @click="resetLocationFilters">Reset</button>
        </div>
      </div>
    </section>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <span>Loading intelligence data...</span>
    </div>

    <template v-else>
      <!-- Population Summary Cards -->
      <div class="metrics-row">
        <div class="metric-card" v-for="m in populationMetrics" :key="m.label">
          <div class="metric-icon" :style="{ background: m.iconBg }">
            <component :is="m.iconSvg" class="metric-svg" />
          </div>
          <div class="metric-body">
            <div class="metric-value" :style="{ color: m.color }">{{ m.value }}</div>
            <div class="metric-label">{{ m.label }}</div>
          </div>
        </div>
      </div>

      <section class="demographic-section">
        <div class="section-head">
          <h2 class="card-title">Demographic Insights</h2>
        </div>

        <div class="insights-grid demographic-grid">
          <article class="card insight-panel gender-panel">
            <div class="panel-header">
              <h3 class="chart-title">Gender Distribution</h3>
              <span class="total-note">Total: {{ genderTotal.toLocaleString() }}</span>
            </div>

            <div v-if="genderTotal === 0" class="empty-state">
              No demographic records available.
            </div>
            <div v-else class="chart-layout gender-chart-layout">
              <div class="donut" :style="genderPieStyle">
                <div class="donut-hole">
                  <div class="donut-value">{{ genderTotal.toLocaleString() }}</div>
                  <div class="donut-label">Total Gender</div>
                </div>
              </div>
              <div class="dist-items">
                <div class="dist-item" v-for="item in genderSegments" :key="item.label">
                  <span class="dist-dot" :style="{ background: item.color }"></span>
                  <span class="dist-label">{{ item.label }}</span>
                  <span class="dist-count">{{ item.value.toLocaleString() }}</span>
                </div>
              </div>
            </div>
          </article>

          <article class="card insight-panel divyang-panel">
            <div class="panel-header">
              <h3 class="chart-title">Divyang Distribution</h3>
              <span class="total-note">Total: {{ divyangTotal.toLocaleString() }}</span>
            </div>

            <div v-if="divyangTotal === 0" class="empty-state">
              No divyang records available.
            </div>
            <div v-else class="chart-layout gender-chart-layout">
              <div class="donut" :style="divyangPieStyle">
                <div class="donut-hole">
                  <div class="donut-value">{{ divyangTotal.toLocaleString() }}</div>
                  <div class="donut-label">Total Divyang</div>
                </div>
              </div>
              <div class="dist-items">
                <div class="dist-item" v-for="item in divyangSegments" :key="item.label">
                  <span class="dist-dot" :style="{ background: item.color }"></span>
                  <span class="dist-label">{{ item.label }}</span>
                  <span class="dist-count">{{ item.value.toLocaleString() }}</span>
                </div>
              </div>
            </div>
          </article>

          <article class="card insight-panel bpl-panel">
            <div class="panel-header">
              <h3 class="chart-title">BPL Status</h3>
              <span class="total-note">Total: {{ bplTotal.toLocaleString() }}</span>
            </div>

            <div v-if="bplTotal === 0" class="empty-state">
              No BPL records available.
            </div>
            <div v-else class="chart-layout gender-chart-layout">
              <div class="donut" :style="bplPieStyle">
                <div class="donut-hole">
                  <div class="donut-value">{{ bplTotal.toLocaleString() }}</div>
                  <div class="donut-label">Total Families</div>
                </div>
              </div>
              <div class="dist-items">
                <div class="dist-item" v-for="item in bplSegments" :key="item.label">
                  <span class="dist-dot" :style="{ background: item.color }"></span>
                  <span class="dist-label">{{ item.label }}</span>
                  <span class="dist-count">{{ item.value.toLocaleString() }}</span>
                </div>
              </div>
            </div>
          </article>

          <article class="card insight-panel">
            <div class="panel-header">
              <h3 class="chart-title">Age-wise Family Income Distribution</h3>
            </div>

            <div class="age-mixed-chart">
              <div class="age-income-gender-canvas-wrap chart-container">
                <canvas ref="ageIncomeGenderCanvas" class="age-income-gender-canvas"></canvas>
              </div>
            </div>
          </article>
        </div>
      </section>

      <section class="intelligence-layout">
        <div class="insights-grid intelligence-row-1">
          <article class="card insight-panel">
            <div class="panel-header">
              <h2 class="card-title">Education Intelligence</h2>
            </div>

            <div class="mini-stats mini-stats-5">
              <div class="mini-stat" v-for="m in educationMetrics" :key="m.label">
                <div class="mini-stat-num" :style="{ color: m.color }">{{ m.value }}</div>
                <div class="mini-stat-label">{{ m.label }}</div>
              </div>
            </div>

            <div class="distribution-section">
              <h3 class="dist-title">Qualification Distribution</h3>
              <div v-if="qualificationTotal === 0" class="empty-state">No qualification records available.</div>
              <div v-else class="distribution-bars">
                <div class="distribution-row" v-for="item in qualificationSegments" :key="item.label">
                  <div class="distribution-label">{{ item.label }}</div>
                  <div class="distribution-track">
                    <div class="distribution-fill" :style="{ width: qualificationBarWidth(item.value), background: item.color }"></div>
                  </div>
                  <div class="distribution-value">{{ item.value.toLocaleString() }}</div>
                </div>
              </div>
            </div>

            <div class="compact-note">
              <div><strong>Literacy Rate:</strong> {{ literacyRateLabel }}</div>
              <div>{{ Number(education.literate_population || 0).toLocaleString() }} literate out of {{ Number(populationStats.total_population || 0).toLocaleString() }} population</div>
            </div>
          </article>

          <article class="card insight-panel">
            <div class="panel-header">
              <h2 class="card-title">Employment Insights</h2>
            </div>

            <div class="mini-stats mini-stats-4">
              <div class="mini-stat" v-for="m in employmentMetrics" :key="m.label">
                <div class="mini-stat-num" :style="{ color: m.color }">{{ m.value }}</div>
                <div class="mini-stat-label">{{ m.label }}</div>
              </div>
            </div>

            <div class="distribution-section">
              <h3 class="dist-title">Occupation Distribution</h3>
              <div v-if="occupationTotal === 0" class="empty-state">No occupation records available.</div>
              <div v-else class="distribution-bars">
                <div class="distribution-row" v-for="item in occupationSegments" :key="item.label">
                  <div class="distribution-label">{{ item.label }}</div>
                  <div class="distribution-track">
                    <div class="distribution-fill" :style="{ width: occupationBarWidth(item.value), background: item.color }"></div>
                  </div>
                  <div class="distribution-value">{{ item.value.toLocaleString() }}</div>
                </div>
              </div>
            </div>
          </article>
        </div>

        <div class="insights-grid intelligence-row-2">
          <div class="card insight-panel agriculture-panel">
            <div class="panel-header">
              <h2 class="card-title">Agriculture Intelligence</h2>
            </div>

            <div class="agri-stats">
              <div class="agri-stat">
                <div class="agri-stat-num" style="color: var(--amber)">{{ agriculture.totalFarmers?.toLocaleString() || '—' }}</div>
                <div class="agri-stat-label">Total Farmers</div>
              </div>
              <div class="agri-stat">
                <div class="agri-stat-num" style="color: var(--red)">{{ agriculture.farmersWithoutIrrigation?.toLocaleString() || '—' }}</div>
                <div class="agri-stat-label">No Irrigation</div>
              </div>
              <div class="agri-stat">
                <div class="agri-stat-num" style="color: var(--teal)">{{ agriculture.kharifFarmers?.toLocaleString() || '—' }}</div>
                <div class="agri-stat-label">Kharif Active</div>
              </div>
              <div class="agri-stat">
                <div class="agri-stat-num" style="color: var(--green)">{{ agriculture.rabiFarmers?.toLocaleString() || '—' }}</div>
                <div class="agri-stat-label">Rabi Active</div>
              </div>
            </div>

            <div class="distribution-section">
              <div class="agri-chart-grid">
                <article class="agri-chart-card">
                  <h3 class="dist-title">Land Holdings Distribution</h3>
                  <div v-if="agriculture.landDistribution?.length" class="agri-land-bars">
                    <div class="land-bar-item" v-for="d in agriculture.landDistribution" :key="d.label">
                      <div class="land-bar-label">{{ d.label }}</div>
                      <div class="land-bar-track">
                        <div class="land-bar-fill" :style="{ width: landPct(d.count) + '%' }"></div>
                      </div>
                      <div class="land-bar-count">{{ d.count.toLocaleString() }}</div>
                    </div>
                  </div>
                  <div v-else class="empty-state">No land holdings records available.</div>
                </article>

                <article class="agri-chart-card">
                  <h3 class="dist-title">Land Utilization</h3>
                  <div v-if="landUtilizationHasData" class="agri-apex-wrap">
                    <apexchart
                      height="220"
                      type="donut"
                      :options="landUtilizationOptions"
                      :series="landUtilizationSeries"
                    />
                    <div class="land-utilization-footnote">
                      <div>Based on {{ landUtilizationRows.validRecords.toLocaleString() }} valid survey records</div>
                      <div><span class="land-utilization-footnote-warn">{{ landUtilizationRows.invalidRecords.toLocaleString() }}</span> records excluded due to invalid values</div>
                    </div>
                  </div>
                  <div v-else class="empty-state">No land utilization records available.</div>
                </article>

                <article class="agri-chart-card season-crops-card">
                  <h3 class="dist-title chart-header">Season-wise Crops</h3>
                  <div v-if="seasonCropHasData" class="agri-apex-wrap agri-chart chart-container">
                    <apexchart
                      height="280"
                      type="bar"
                      :options="seasonCropOptions"
                      :series="seasonCropSeries"
                    />
                  </div>
                  <div v-else class="empty-state">No season-wise crop records available.</div>
                </article>
              </div>
            </div>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, h, watch, nextTick } from 'vue'
import { Chart, registerables } from 'chart.js'
import ChartDataLabels from 'chartjs-plugin-datalabels'
import { getDashboardSummary, getLocationOptions } from '../../api/index.js'

Chart.register(...registerables, ChartDataLabels)
Chart.defaults.font.family = 'inherit'

const loading = ref(true)
const agriculture = ref({})
const populationStats = ref({
  total_population: 0,
  total_households: 0,
  working_population: 0,
  dependent_population: 0,
})
const demographics = ref({
  gender_distribution: { male: 0, female: 0, other: 0 },
  age_distribution: { age_0_5: 0, age_6_18: 0, age_19_35: 0, age_36_60: 0, age_60_plus: 0 },
  age_income_gender_distribution: [],
})
const ageIncomeGenderSegments = ref([])
const education = ref({
  literate_population: 0,
  illiterate_population: 0,
  students_count: 0,
  dropout_count: 0,
  graduate_population: 0,
  literacy_rate: 0,
  qualification_distribution: { below_10th: 0, tenth: 0, twelfth: 0, graduate_above: 0 },
})
const employment = ref({
  employed_members: 0,
  unemployed_members: 0,
  daily_wage_workers: 0,
  skilled_workers: 0,
  occupation_distribution: {
    farm_based: 0,
    agri_allied: 0,
    non_farm: 0,
    salaried: 0,
    wage_workers: 0,
    housewife: 0,
    students: 0,
    unemployed: 0,
    other: 0,
  },
})
const bplDistribution = ref({ bpl: 0, non_bpl: 0, total_households: 0 })
const selectedDistrict = ref('')
const selectedTaluka = ref('')
const selectedVillage = ref('')
const districtOptions = ref([])
const talukaOptions = ref([])
const villageOptions = ref([])
const ageIncomeGenderCanvas = ref(null)
let isActive = true
let ageIncomeGenderChart = null
const AGE_GROUP_ORDER = ['18-30', '31-45', '46-60', '60+']

onUnmounted(() => {
  isActive = false
  destroyAgeIncomeGenderChart()
})

function applyDemographicsData(data) {
  demographics.value = data
  const distribution = demographics.value?.age_income_gender_distribution || []
  const mapped = Array.isArray(distribution)
    ? distribution.map(d => ({
        age_group: d.age_group,
        families: Number(d.families ?? 0),
        avg_income: Number(d.avg_income ?? d.total_income ?? d.income) || 0,
      }))
    : []

  const grouped = {}

  AGE_GROUP_ORDER.forEach(ageGroup => {
    grouped[ageGroup] = { age_group: ageGroup, families: 0, avg_income: 0 }
  })

  mapped.forEach(item => {
    const ageGroup = item?.age_group || ''
    if (!grouped[ageGroup]) return

    grouped[ageGroup].families += Number(item?.families ?? 0)
    grouped[ageGroup].avg_income = Number(item?.avg_income ?? 0)
  })

  ageIncomeGenderSegments.value = AGE_GROUP_ORDER.map(ageGroup => grouped[ageGroup])
}

function formatIncome(value) {
  const amount = Math.round(Number(value || 0))
  return `₹${amount.toLocaleString('en-IN')}`
}

function destroyAgeIncomeGenderChart() {
  if (ageIncomeGenderChart) {
    ageIncomeGenderChart.destroy()
    ageIncomeGenderChart = null
  }
}

function syncAgeIncomeGenderChart() {
  const canvas = ageIncomeGenderCanvas.value
  if (!canvas) return

  const data = ageIncomeGenderSegments.value || []
  const avgIncomeValues = data.map(item => Number(item.avg_income) || 0)
  const chartData = {
    datasets: [
      {
        data: data.map(item => ({
          x: item.age_group,
          y: Number(item.families) || 0,
          families: Number(item.families) || 0,
          avg_income: Number(item.avg_income) || 0,
        })),
        backgroundColor: '#3b82f6',
        borderRadius: 4,
        barPercentage: 0.6,
        categoryPercentage: 0.7,
      },
    ],
  }

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    animation: { duration: 250 },
    layout: {
      padding: {
        top: 30,
        right: 10,
        left: 10,
        bottom: 10,
      },
    },
    interaction: {
      mode: 'index',
      intersect: false,
    },
    plugins: {
      legend: {
        display: false,
        position: 'top',
        labels: {
          usePointStyle: true,
          boxWidth: 8,
          padding: 15,
          color: '#6b7280',
        },
      },
      tooltip: {
        enabled: true,
        backgroundColor: '#ffffff',
        titleColor: '#111827',
        bodyColor: '#374151',
        borderColor: '#e5e7eb',
        borderWidth: 1,
        cornerRadius: 8,
        padding: 10,
        displayColors: true,
        callbacks: {
          title(items) {
            return `Age group: ${items?.[0]?.raw?.x || items?.[0]?.label || ''}`
          },
          label(context) {
            const raw = context.raw || {}
            const families = Number(raw.families || 0)
            return `Families: ${families.toLocaleString()}`
          },
        },
      },
      datalabels: {
        display: true,
        anchor: 'end',
        align: 'top',
        offset: 4,
        clamp: true,
        clip: false,
        color: '#374151',
        formatter: (_, context) => formatIncome(avgIncomeValues[context.dataIndex] || 0),
        font: {
          weight: '600',
          size: 11,
        },
      },
    },
    scales: {
      x: {
        border: {
          color: '#e5e7eb',
        },
        grid: {
          color: '#f3f4f6',
          drawBorder: false,
        },
        ticks: {
          color: '#6b7280',
          autoSkip: false,
          maxRotation: 0,
        },
        title: {
          display: true,
          text: 'Age group',
        },
      },
      y: {
        beginAtZero: true,
        grace: '20%',
        border: {
          color: '#e5e7eb',
        },
        grid: {
          color: '#f3f4f6',
          drawBorder: false,
        },
        title: {
          display: true,
          text: 'Families count',
        },
        ticks: {
          color: '#6b7280',
          precision: 0,
        },
      },
    },
    datasets: {
      bar: {
        barPercentage: 0.6,
        categoryPercentage: 0.7,
      },
    },
    layout: {
      padding: {
        top: 30,
        right: 10,
        left: 10,
        bottom: 10,
      },
    },
  }

  if (!ageIncomeGenderChart) {
    ageIncomeGenderChart = new Chart(canvas, {
      type: 'bar',
      data: chartData,
      options: chartOptions,
    })
    return
  }

  ageIncomeGenderChart.data = chartData
  ageIncomeGenderChart.options = chartOptions
  ageIncomeGenderChart.update()
}

function isValidFilterValue(value) {
  if (value === null || value === undefined) return false
  const normalized = String(value).trim().toLowerCase()
  return normalized !== '' && normalized !== '0' && normalized !== 'null' && normalized !== 'undefined'
}

function buildLocationParams() {
  const params = {}

  if (isValidFilterValue(selectedDistrict.value)) {
    params.district_id = String(selectedDistrict.value).trim()
  }
  if (isValidFilterValue(selectedTaluka.value)) {
    params.taluka_id = String(selectedTaluka.value).trim()
  }
  if (isValidFilterValue(selectedVillage.value)) {
    params.village_id = String(selectedVillage.value).trim()
  }

  return params
}

async function loadLocationOptions() {
  try {
    const optionParams = {}
    if (isValidFilterValue(selectedDistrict.value)) {
      optionParams.district_id = String(selectedDistrict.value).trim()
    }
    if (isValidFilterValue(selectedTaluka.value)) {
      optionParams.taluka_id = String(selectedTaluka.value).trim()
    }

    const options = await getLocationOptions(optionParams)

    if (!isActive) return
    districtOptions.value = options?.districts || []
    talukaOptions.value = options?.talukas || []
    villageOptions.value = options?.villages || []
  } catch (error) {
    console.warn('Dashboard location options unavailable:', error?.message || error)
  }
}

async function fetchDashboardData(params = {}) {
  loading.value = true

  const coreResults = await Promise.allSettled([
    getDashboardSummary(params),
  ])

  if (!isActive) return

  if (coreResults[0].status === 'fulfilled') {
    const payload = coreResults[0].value || {}
    agriculture.value = payload.agriculture || {}
    populationStats.value = payload.population || populationStats.value
    applyDemographicsData(payload.demographics || {})
    education.value = payload.education || education.value
    employment.value = payload.employment || employment.value
    bplDistribution.value = payload.demographics?.bpl_distribution || payload.population?.bpl_distribution || { bpl: 0, non_bpl: 0, total_households: Number((payload.population || {}).total_households || 0) }
  }

  loading.value = false
}

async function onDistrictChange() {
  selectedTaluka.value = ''
  selectedVillage.value = ''
  await loadLocationOptions()
}

async function onTalukaChange() {
  selectedVillage.value = ''
  await loadLocationOptions()
}

async function applyLocationFilters() {
  await fetchDashboardData(buildLocationParams())
}

async function resetLocationFilters() {
  selectedDistrict.value = ''
  selectedTaluka.value = ''
  selectedVillage.value = ''
  const dashboardPromise = fetchDashboardData()
  const optionsPromise = loadLocationOptions()
  await Promise.allSettled([dashboardPromise, optionsPromise])
}

onMounted(async () => {
  const dashboardPromise = fetchDashboardData()
  const optionsPromise = loadLocationOptions()
  await Promise.allSettled([dashboardPromise, optionsPromise])
})

watch(
  [loading, ageIncomeGenderSegments],
  async ([isLoading]) => {
    if (isLoading) {
      destroyAgeIncomeGenderChart()
      return
    }

    await nextTick()
    syncAgeIncomeGenderChart()
  },
  { deep: true, immediate: true }
)

const GroupIcon = {
  render: () => h('svg', { viewBox: '0 0 24 24', fill: 'currentColor' }, [
    h('path', { d: 'M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5s-3 1.34-3 3 1.34 3 3 3zM8 11c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5 5 6.34 5 8s1.34 3 3 3zM8 13c-2.33 0-7 1.17-7 3.5V19h14v-2.5C15 14.17 10.33 13 8 13zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5C23 14.17 18.33 13 16 13z' }),
  ]),
}

const HomeIcon = {
  render: () => h('svg', { viewBox: '0 0 24 24', fill: 'currentColor' }, [
    h('path', { d: 'M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z' }),
  ]),
}

const BriefcaseIcon = {
  render: () => h('svg', { viewBox: '0 0 24 24', fill: 'currentColor' }, [
    h('path', { d: 'M10 4h4v2h-4V4zm10 3h-4V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2H4a2 2 0 0 0-2 2v3h20V9a2 2 0 0 0-2-2zm-8 6c-1.1 0-2-.9-2-2H2v8a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-8h-8c0 1.1-.9 2-2 2z' }),
  ]),
}

const ChildIcon = {
  render: () => h('svg', { viewBox: '0 0 24 24', fill: 'currentColor' }, [
    h('path', { d: 'M13 7a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm4 11h-2v-4h-2v4H7v-6a3 3 0 0 1 3-3h4a3 3 0 0 1 3 3v6z' }),
  ]),
}


const populationMetrics = computed(() => {
  const s = populationStats.value
  return [
    { label: 'Total Households', value: Number(s.total_households || 0).toLocaleString(), color: 'var(--teal)', iconBg: 'var(--teal-dim)', iconSvg: HomeIcon },
    { label: 'Total Population', value: Number(s.total_population || 0).toLocaleString(), color: 'var(--text-primary)', iconBg: 'var(--amber-dim)', iconSvg: GroupIcon },
  ]
})

const genderSegments = computed(() => {
  const g = demographics.value.gender_distribution || {}
  return [
    { label: 'Male', value: Number(g.male || 0), color: 'var(--teal)' },
    { label: 'Female', value: Number(g.female || 0), color: 'var(--amber)' },
    { label: 'Other', value: Number(g.other || 0), color: 'var(--text-dim)' },
  ]
})
const genderTotal = computed(() => genderSegments.value.reduce((sum, item) => sum + item.value, 0))

const genderPieStyle = computed(() => {
  const total = genderTotal.value
  if (!total) {
    return { background: 'conic-gradient(var(--bg-surface) 0 100%)' }
  }
  const malePct = (genderSegments.value[0].value / total) * 100
  const femalePct = (genderSegments.value[1].value / total) * 100
  const otherPct = Math.max(0, 100 - malePct - femalePct)
  return {
    background: `conic-gradient(var(--teal) 0 ${malePct}%, var(--amber) ${malePct}% ${malePct + femalePct}%, var(--text-dim) ${malePct + femalePct}% ${malePct + femalePct + otherPct}%)`,
  }
})

const bplSegments = computed(() => {
  const bpl = Number(bplDistribution.value?.bpl || 0)
  const nonBpl = Number(bplDistribution.value?.non_bpl || 0)

  return [
    { label: 'BPL', value: bpl, color: '#ef4444' },
    { label: 'Non-BPL', value: nonBpl, color: '#10b981' },
  ]
})

const bplTotal = computed(() => bplSegments.value.reduce((sum, item) => sum + item.value, 0))

const bplPieStyle = computed(() => {
  const total = bplTotal.value
  if (!total) {
    return { background: 'conic-gradient(var(--bg-surface) 0 100%)' }
  }

  const bplPct = (bplSegments.value[0].value / total) * 100
  const nonBplPct = Math.max(0, 100 - bplPct)
  return {
    background: `conic-gradient(#ef4444 0 ${bplPct}%, #10b981 ${bplPct}% ${bplPct + nonBplPct}%)`,
  }
})

const DIVYANG_COLORS = ['#f59e0b', '#06b6d4', '#8b5cf6', '#ef4444', '#22c55e', '#3b82f6', '#14b8a6', '#9ca3af']

function mapDisabilityGroup(name) {
  const n = String(name || '').toLowerCase()

  if (n.includes('blind') || n.includes('low-vision') || n.includes('low vision')) return 'Visual Disability'

  if (
    n.includes('locomotor') ||
    n.includes('cerebral') ||
    n.includes('muscular') ||
    n.includes('dwarf')
  ) return 'Locomotor Disability'

  if (
    n.includes('mental') ||
    n.includes('autism') ||
    n.includes('intellectual') ||
    n.includes('learning')
  ) return 'Intellectual Disability'

  if (n.includes('hearing')) return 'Hearing Disability'
  if (n.includes('speech')) return 'Speech Disability'
  if (n.includes('multiple')) return 'Multiple Disabilities'

  if (
    n.includes('parkinson') ||
    n.includes('sclerosis') ||
    n.includes('sickle') ||
    n.includes('thalassemia') ||
    n.includes('neurological')
  ) return 'Chronic Conditions'

  if (
    n.includes('acid attack') ||
    n.includes('leprosy')
  ) return 'Other'

  return name ? String(name).trim() : 'Other'
}

const divyangSegments = computed(() => {
  const source = Array.isArray(demographics.value?.disability_distribution)
    ? demographics.value.disability_distribution
    : []

  const grouped = new Map()
  for (const item of source) {
    const group = mapDisabilityGroup(item?.name)
    const next = Number(item?.value || 0)
    grouped.set(group, Number(grouped.get(group) || 0) + next)
  }

  const rows = [...grouped.entries()]
    .map(([label, value], index) => ({
      label,
      value,
      color: DIVYANG_COLORS[index % DIVYANG_COLORS.length],
    }))
    .filter(item => item.value > 0)
    .sort((a, b) => b.value - a.value)

  return rows
})

const divyangTotal = computed(() => {
  const apiTotal = Number(demographics.value?.total_divyang || 0)
  if (apiTotal > 0) return apiTotal
  return divyangSegments.value.reduce((sum, item) => sum + item.value, 0)
})

const divyangPieStyle = computed(() => {
  const total = divyangTotal.value
  if (!total || !divyangSegments.value.length) {
    return { background: 'conic-gradient(var(--bg-surface) 0 100%)' }
  }

  let current = 0
  const stops = divyangSegments.value.map(segment => {
    const pct = (segment.value / total) * 100
    const start = current
    const end = current + pct
    current = end
    return `${segment.color} ${start}% ${end}%`
  })

  return {
    background: `conic-gradient(${stops.join(', ')})`,
  }
})

const educationMetrics = computed(() => {
  const e = education.value
  return [
    { label: 'Literate Population', value: Number(e.literate_population || 0).toLocaleString(), color: 'var(--teal)' },
    { label: 'Illiterate Population', value: Number(e.illiterate_population || 0).toLocaleString(), color: 'var(--red)' },
    { label: 'Students', value: Number(e.students_count || 0).toLocaleString(), color: 'var(--amber)' },
    { label: 'School Dropouts', value: Number(e.dropout_count || 0).toLocaleString(), color: 'var(--text-primary)' },
    { label: 'Graduates', value: Number(e.graduate_population || 0).toLocaleString(), color: 'var(--green)' },
  ]
})

const qualificationSegments = computed(() => {
  const q = education.value.qualification_distribution || {}
  return [
    { label: 'Below 10th', value: Number(q.below_10th || 0), color: 'var(--text-dim)' },
    { label: '10th', value: Number(q.tenth || 0), color: 'var(--teal)' },
    { label: '12th', value: Number(q.twelfth || 0), color: 'var(--amber)' },
    { label: 'Graduation & Above', value: Number(q.graduate_above || 0), color: 'var(--green)' },
  ].sort((a, b) => b.value - a.value)
})
const qualificationTotal = computed(() => qualificationSegments.value.reduce((sum, item) => sum + item.value, 0))
const qualificationMax = computed(() => Math.max(...qualificationSegments.value.map((item) => item.value), 1))
const qualificationBarWidth = (value) => `${(Number(value || 0) / qualificationMax.value) * 100}%`
const literacyRateLabel = computed(() => `${Math.round(Number(education.value.literacy_rate || 0))}%`)

const employmentMetrics = computed(() => {
  const e = employment.value
  return [
    { label: 'Employed Members', value: Number(e.employed_members || 0).toLocaleString(), color: 'var(--green)' },
    { label: 'Unemployed Members', value: Number(e.unemployed_members || 0).toLocaleString(), color: 'var(--red)' },
    { label: 'Daily Wage Workers', value: Number(e.daily_wage_workers || 0).toLocaleString(), color: 'var(--amber)' },
    { label: 'Skilled Workers', value: Number(e.skilled_workers || 0).toLocaleString(), color: 'var(--teal)' },
  ]
})

const occupationSegments = computed(() => {
  const o = employment.value.occupation_distribution || {}
  return [
    { label: 'Farm Based', value: Number(o.farm_based || 0), color: 'var(--green)' },
    { label: 'Agri Allied', value: Number(o.agri_allied || 0), color: 'var(--teal)' },
    { label: 'Non Farm Self Employed', value: Number(o.non_farm || 0), color: 'var(--amber)' },
    { label: 'Salaried', value: Number(o.salaried || 0), color: 'var(--teal)' },
    { label: 'Wage Workers', value: Number(o.wage_workers || 0), color: 'var(--red)' },
    { label: 'Housewife', value: Number(o.housewife || 0), color: 'var(--text-muted)' },
    { label: 'Students', value: Number(o.students || 0), color: 'var(--amber)' },
    { label: 'Unemployed', value: Number(o.unemployed || 0), color: 'var(--red)' },
    { label: 'Other', value: Number(o.other || 0), color: 'var(--text-dim)' },
  ].sort((a, b) => b.value - a.value)
})
const occupationTotal = computed(() => occupationSegments.value.reduce((sum, item) => sum + item.value, 0))
const occupationMax = computed(() => Math.max(...occupationSegments.value.map((item) => item.value), 1))
const occupationBarWidth = (value) => `${(Number(value || 0) / occupationMax.value) * 100}%`

const landUtilizationRows = computed(() => {
  const summary = agriculture.value?.land_utilization_summary || agriculture.value?.landUtilizationSummary
  if (summary) {
    const total = Number(summary.total_land ?? summary.totalLand ?? 0)
    const cultivated = Number(summary.cultivated_land ?? summary.cultivatedLand ?? 0)
    const unused = Number(summary.unused_land ?? summary.unusedLand ?? Math.max(total - cultivated, 0))
    const validRecords = Number(summary.valid_records ?? summary.validRecords ?? 0)
    const invalidRecords = Number(summary.invalid_records ?? summary.invalidRecords ?? 0)
    const cultivatedPercent = Number(summary.cultivated_percent ?? summary.cultivatedPercent ?? (total > 0 ? (cultivated * 100) / total : 0))
    const unusedPercent = Number(summary.unused_percent ?? summary.unusedPercent ?? (total > 0 ? (Math.max(unused, 0) * 100) / total : 0))

    return {
      total: Number(total.toFixed(2)),
      cultivated: Number(cultivated.toFixed(2)),
      unused: Number(Math.max(unused, 0).toFixed(2)),
      validRecords,
      invalidRecords,
      cultivatedPercent: Number(cultivatedPercent.toFixed(2)),
      unusedPercent: Number(unusedPercent.toFixed(2)),
    }
  }

  return {
    total: 0,
    cultivated: 0,
    unused: 0,
    validRecords: 0,
    invalidRecords: 0,
    cultivatedPercent: 0,
    unusedPercent: 0,
  }
})

const landUtilizationSeries = computed(() => [
  landUtilizationRows.value.cultivated,
  landUtilizationRows.value.unused,
])

const landUtilizationHasData = computed(() =>
  landUtilizationRows.value.total > 0
)

const landUtilizationOptions = computed(() => {
  const total = landUtilizationRows.value.total
  const totalFormatted = `${Number(total || 0).toLocaleString()} acres`
  const cultivatedPercent = landUtilizationRows.value.cultivatedPercent
  const unusedPercent = landUtilizationRows.value.unusedPercent
  
  return {
    chart: {
      type: 'donut',
      toolbar: { show: false },
    },
    labels: ['Cultivated Land', 'Unused Land'],
    plotOptions: {
      pie: {
        donut: {
          size: '60%',
          labels: {
            show: true,
            name: {
              show: true,
              fontSize: '12px',
              fontFamily: 'var(--font-display)',
              color: 'var(--text-muted)',
            },
            value: {
              show: true,
              fontSize: '18px',
              fontFamily: 'var(--font-display)',
              color: 'var(--text-primary)',
              formatter: () => totalFormatted,
            },
            total: {
              show: true,
              label: 'Total Land',
              fontSize: '11px',
              fontFamily: 'var(--font-body)',
              color: 'var(--text-dim)',
            },
          },
        },
      },
    },
    colors: ['#22c55e', '#ef4444'],
    stroke: {
      width: 2,
      colors: ['var(--bg-surface)'],
    },
    legend: {
      position: 'bottom',
      fontSize: '12px',
    },
    tooltip: {
      y: {
        formatter: (value, { seriesIndex }) => {
          const pct = seriesIndex === 0 ? cultivatedPercent : unusedPercent
          return `${Number(value || 0).toLocaleString()} acres (${Number(pct || 0).toFixed(2)}%)`
        },
      },
    },
    dataLabels: {
      enabled: false,
    },
  }
})

function toCropList(value) {
  if (!value) return []
  if (Array.isArray(value)) return value.map(item => String(item).trim()).filter(Boolean)

  return String(value)
    .split(/[,;/|&]+|\band\b/gi)
    .map(item => item.trim())
    .filter(Boolean)
}

const seasonCropCounts = computed(() => {
  const source = Array.isArray(agriculture.value?.season_crop_rows)
    ? agriculture.value.season_crop_rows
    : Array.isArray(agriculture.value?.seasonCropRows)
      ? agriculture.value.seasonCropRows
      : []

  const kharif = new Map()
  const rabi = new Map()

  for (const row of source) {
    const season = String(row?.season || '').trim().toLowerCase()
    const cropName = String(row?.crop || '').trim()
    const aggregatedCount = Number(row?.count ?? row?.cnt ?? 0)

    if (season && cropName && aggregatedCount > 0) {
      const key = cropName.toUpperCase()
      if (season === 'kharif') {
        kharif.set(key, (kharif.get(key) || 0) + aggregatedCount)
      } else if (season === 'rabi') {
        rabi.set(key, (rabi.get(key) || 0) + aggregatedCount)
      }
      continue
    }

    const kharifCrops = toCropList(row?.CULTIVATING_DURING_KHARIF_SEASON ?? row?.cultivating_during_kharif_season)
    const rabiCrops = toCropList(row?.CULTIVATING_DURING_RABI_SEASON ?? row?.cultivating_during_rabi_season)

    for (const crop of kharifCrops) {
      const key = crop.toUpperCase()
      kharif.set(key, (kharif.get(key) || 0) + 1)
    }

    for (const crop of rabiCrops) {
      const key = crop.toUpperCase()
      rabi.set(key, (rabi.get(key) || 0) + 1)
    }
  }

  const cropSet = new Set([...kharif.keys(), ...rabi.keys()])
  const ranked = [...cropSet]
    .map(crop => ({
      crop,
      kharif: kharif.get(crop) || 0,
      rabi: rabi.get(crop) || 0,
      total: (kharif.get(crop) || 0) + (rabi.get(crop) || 0),
    }))
    .filter(item => item.total > 0)
    .sort((a, b) => b.total - a.total)
    .slice(0, 6)

  return ranked
})

const seasonCropSeries = computed(() => ([
  {
    name: 'Kharif',
    data: seasonCropCounts.value.map(item => item.kharif),
  },
  {
    name: 'Rabi',
    data: seasonCropCounts.value.map(item => item.rabi),
  },
]))

const seasonCropHasData = computed(() => seasonCropCounts.value.length > 0)

const seasonCropOptions = computed(() => ({
  chart: {
    type: 'bar',
    height: 280,
    stacked: false,
    toolbar: { show: false },
  },
  plotOptions: {
    bar: {
      columnWidth: '50%',
      borderRadius: 6,
    },
  },
  colors: ['#0ea5e9', '#14b8a6'],
  grid: {
    borderColor: '#e5e7eb',
    strokeDashArray: 3,
    row: {
      colors: ['transparent'],
      opacity: 0.5,
    },
    padding: {
      left: 20,
      right: 10,
      top: 10,
      bottom: 10,
    },
  },
  xaxis: {
    categories: seasonCropCounts.value.map(item => item.crop),
    labels: {
      rotate: -30,
      offsetY: 5,
      formatter: (value) => String(value || '')
        .toLowerCase()
        .replace(/\b\w/g, (char) => char.toUpperCase()),
      style: {
        fontSize: '12px',
        colors: '#475569',
      },
    },
    title: {
      text: 'Crop name',
      offsetY: -5,
      style: {
        fontSize: '13px',
        fontWeight: 500,
        color: '#475569',
      },
    },
  },
  yaxis: {
    tickAmount: 4,
    min: 0,
    forceNiceScale: true,
    labels: {
      offsetX: -8,
      formatter: (val) => Math.round(val),
      style: {
        fontSize: '12px',
        colors: '#475569',
      },
    },
    title: {
      text: 'Count',
      offsetX: -10,
      style: {
        color: '#475569',
      },
    },
  },
  legend: {
    position: 'top',
    horizontalAlign: 'center',
    offsetY: 5,
    labels: {
      colors: '#475569',
    },
  },
  dataLabels: {
    enabled: false,
  },
}))

function landPct(count) {
  const max = Math.max(...(agriculture.value.landDistribution || []).map(d => d.count), 1)
  return (count / max * 100).toFixed(1)
}
</script>

<style scoped>
.dashboard {
  padding: 1.5rem 2rem;
  max-width: none;
  width: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
}

.page-title {
  font-family: var(--font-display);
  font-size: 1.8rem;
  color: var(--text-primary);
  line-height: 1.1;
  font-weight: 400;
}

.page-subtitle {
  color: var(--text-dim);
  font-size: 0.86rem;
  margin-top: 0.35rem;
}

.header-badge {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 999px;
  padding: 0.38rem 0.85rem;
  font-size: 0.72rem;
  color: var(--text-muted);
}

.badge-dot {
  width: 6px; height: 6px;
  border-radius: 50%;
  background: var(--green);
  box-shadow: 0 0 8px var(--green);
  animation: pulse 2s infinite;
}

@keyframes pulse { 0%,100% { opacity:1; } 50% { opacity:0.4; } }

/* Metrics */
.metrics-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.85rem;
  margin-bottom: 1rem;
}

.metric-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 0.95rem 1rem;
  display: flex;
  align-items: center;
  gap: 0.9rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.metric-card:hover {
  border-color: var(--border-light);
  box-shadow: 0 2px 10px var(--shadow);
}

.metric-icon {
  width: 40px; height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.metric-svg { width: 20px; height: 20px; color: var(--text-muted); }

.metric-value {
  font-family: var(--font-display);
  font-size: 1.4rem;
  line-height: 1.1;
}

.metric-label {
  font-size: 0.74rem;
  color: var(--text-dim);
  letter-spacing: 0.02em;
  margin-top: 0.2rem;
}

.demographic-section {
  margin-top: 0.2rem;
}

.intelligence-section {
  margin-top: 1.15rem;
}

.intelligence-layout {
  margin-top: 1.15rem;
}

.section-head {
  margin-bottom: 0.9rem;
}

/* Insights Grid */
.insights-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.9rem;
}

.intelligence-row-1 {
  grid-template-columns: repeat(2, 1fr);
}

.intelligence-row-2 {
  grid-template-columns: 1fr;
  margin-top: 0.9rem;
}

.demographic-grid {
  grid-template-columns: repeat(3, 1fr);
}

.insight-panel { padding: 1.1rem; }
.welfare-panel { grid-column: 1 / -1; }
.agriculture-panel { grid-column: 1 / -1; }
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.9rem;
}

.chart-title {
  font-size: 0.9rem;
  color: var(--text-primary);
  font-weight: 500;
}

.total-note {
  font-size: 0.72rem;
  color: var(--text-dim);
}

.chart-layout {
  display: grid;
  grid-template-columns: 150px 1fr;
  gap: 1rem;
  align-items: center;
}

.gender-chart-layout {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 24px;
}

.donut {
  width: 130px;
  height: 130px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.donut-hole {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: var(--bg-card);
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
  color: var(--text-primary);
  font-size: 1.05rem;
  line-height: 1.1;
}

.mini-stats {
  display: grid;
  gap: 0.65rem;
  margin-bottom: 1rem;
}

.mini-stats-5 { grid-template-columns: repeat(5, 1fr); }
.mini-stats-4 { grid-template-columns: repeat(4, 1fr); }

.mini-stat {
  text-align: center;
  padding: 0.65rem;
  background: var(--bg-surface);
  border-radius: 8px;
}

.mini-stat-num {
  font-family: var(--font-display);
  font-size: 1.35rem;
  line-height: 1;
}

.mini-stat-label {
  font-size: 0.65rem;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-top: 0.2rem;
}

.distribution-bars {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.distribution-row {
  display: grid;
  grid-template-columns: 138px 1fr 62px;
  align-items: center;
  gap: 0.7rem;
}

.distribution-label {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-align: right;
}

.distribution-track {
  height: 8px;
  background: var(--bg-deep);
  border-radius: 4px;
  overflow: hidden;
}

.age-distribution-bars .distribution-row {
  grid-template-columns: 90px 1fr 40px;
  gap: 0.55rem;
}

.age-distribution-bars .distribution-label {
  text-align: right;
}

.age-distribution-bars .distribution-track {
  background: #f1f5f9;
}

.distribution-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.8s ease;
}

.distribution-value {
  font-size: 0.75rem;
  color: var(--text-body);
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.compact-note {
  margin-top: 0.85rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--border);
  color: var(--text-dim);
  font-size: 0.76rem;
  line-height: 1.4;
}

.empty-state {
  color: var(--text-dim);
  font-size: 0.78rem;
  padding: 0.7rem;
  border-radius: 8px;
  background: var(--bg-surface);
}

.age-mixed-chart {
  margin-top: 0.25rem;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 3rem;
  color: var(--text-dim);
  font-size: 0.85rem;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border);
  border-top-color: var(--amber);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.dashboard-filter {
  margin-bottom: 1rem;
  padding: 0.9rem 1rem;
}

.dashboard-filter-head {
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-dim);
  margin-bottom: 0.6rem;
  font-weight: 600;
}

.dashboard-filter-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr auto;
  gap: 0.65rem;
  align-items: center;
}

.dashboard-filter-select {
  width: 100%;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg-surface);
  color: var(--text-body);
  padding: 0.55rem 0.7rem;
  font-size: 0.78rem;
}

.dashboard-filter-select:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.dashboard-filter-actions {
  display: flex;
  gap: 0.45rem;
  justify-content: flex-end;
}

.dashboard-apply-btn,
.dashboard-reset-btn {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0.53rem 0.9rem;
  font-size: 0.76rem;
  font-weight: 600;
  cursor: pointer;
}

.dashboard-apply-btn {
  background: var(--teal);
  color: #ffffff;
  border-color: transparent;
}

.dashboard-reset-btn {
  background: var(--bg-surface);
  color: var(--text-body);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Bars */
.insight-bar { margin-bottom: 0.9rem; }
.bar-header { display: flex; justify-content: space-between; margin-bottom: 0.3rem; }
.bar-label { font-size: 0.78rem; color: var(--text-muted); }
.bar-value { font-size: 0.78rem; font-weight: 600; }
.bar-track { height: 6px; background: var(--bg-surface); border-radius: 3px; overflow: hidden; }
.bar-fill { height: 100%; border-radius: 3px; transition: width 1s ease; }

/* Agriculture Stats */
.agri-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.65rem;
  margin-bottom: 1rem;
}

.agri-stat {
  text-align: center;
  padding: 0.65rem;
  background: var(--bg-surface);
  border-radius: 8px;
}

.agri-stat-num { font-family: var(--font-display); font-size: 1.35rem; }
.agri-stat-label { font-size: 0.65rem; color: var(--text-dim); text-transform: uppercase; letter-spacing: 0.06em; margin-top: 0.2rem; }

.agri-chart-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.9rem;
}

.agri-chart-card {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 12px 16px;
  min-height: 300px;
  height: 300px;
  display: flex;
  flex-direction: column;
}

.season-crops-card {
  padding: 16px 20px;
}

.chart-header {
  margin-bottom: 12px;
}

.agri-apex-wrap {
  margin-top: 0.15rem;
  margin-bottom: 0;
  flex: 1;
  min-height: 220px;
  display: flex;
  flex-direction: column;
}

.agri-chart {
  height: 280px;
  width: 100%;
}

.season-crops-card .chart-container {
  margin-bottom: 0;
  overflow: hidden;
}

.labels {
  color: #475569;
  text-transform: none;
}

.season-crops-card :deep(.apexcharts-xaxis-label) {
  text-transform: none;
}

.season-crops-card :deep(.apexcharts-legend) {
  margin-bottom: 10px !important;
}

.agri-land-bars {
  margin-top: 0.15rem;
  max-height: 220px;
  overflow-y: auto;
  padding-right: 0.2rem;
}

.agri-chart-card .empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.land-utilization-footnote {
  margin-top: 0.35rem;
  font-size: 12px;
  color: #64748b;
  line-height: 1.35;
}

.land-utilization-footnote-warn {
  color: #f59e0b;
  font-weight: 600;
}

/* Land Bars */
.land-bar-item {
  display: grid;
  grid-template-columns: 150px 1fr 60px;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}
.land-bar-label { font-size: 0.75rem; color: var(--text-muted); text-align: right; }
.land-bar-track { height: 8px; background: var(--bg-deep); border-radius: 4px; overflow: hidden; }
.land-bar-fill { height: 100%; background: linear-gradient(90deg, var(--amber), var(--teal)); border-radius: 4px; transition: width 1s ease; }
.land-bar-count { font-size: 0.75rem; color: var(--text-body); font-weight: 500; }

/* Age-Income-Gender Chart */
.age-income-gender-chart {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.age-income-gender-canvas-wrap {
  position: relative;
  width: 100%;
  height: 260px;
  overflow: visible;
}

.chart-container {
  overflow: visible;
}

.age-income-gender-canvas {
  width: 100% !important;
  height: 100% !important;
}

.chart-legend-row {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.75rem;
  color: var(--text-muted);
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.aig-grid {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.aig-row {
  display: grid;
  grid-template-columns: 60px 1fr 120px;
  align-items: center;
  gap: 0.85rem;
}

.aig-age-label {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-align: right;
  font-weight: 500;
}

.aig-income-bar {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.aig-bar-track {
  flex: 1;
  height: 10px;
  background: var(--bg-deep);
  border-radius: 5px;
  overflow: hidden;
  min-width: 100px;
}

.aig-bar-fill {
  height: 100%;
  border-radius: 5px;
  transition: width 0.8s ease;
}

.aig-bar-value {
  font-size: 0.7rem;
  color: var(--text-body);
  font-weight: 600;
  min-width: 50px;
  text-align: right;
}

.aig-counts {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
}

.aig-count {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.7rem;
  color: var(--text-body);
  font-weight: 500;
}

.aig-count-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}

/* Welfare Grid */
.welfare-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.welfare-card {
  background: var(--bg-surface);
  border-radius: 8px;
  padding: 0.9rem;
  border-left: 3px solid transparent;
}
.welfare-card.critical { border-left-color: var(--red); }
.welfare-card.warning  { border-left-color: var(--amber); }
.welfare-card.info     { border-left-color: var(--teal); }
.welfare-card.neutral  { border-left-color: var(--text-dim); }

.welfare-num { font-family: var(--font-display); font-size: 1.55rem; color: var(--text-primary); line-height: 1; }
.welfare-desc { font-size: 0.75rem; color: var(--text-muted); margin-top: 0.4rem; line-height: 1.3; }
.welfare-tag {
  margin-top: 0.6rem;
  font-size: 0.6rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-dim);
  background: var(--bg-card);
  display: inline-block;
  padding: 0.2rem 0.5rem;
  border-radius: 3px;
}

/* Distribution */
.distribution-section { margin-top: 1.25rem; padding-top: 1rem; border-top: 1px solid var(--border); }
.dist-title { font-size: 0.7rem; text-transform: uppercase; letter-spacing: 0.1em; color: var(--text-dim); margin-bottom: 0.75rem; font-weight: 600; }
.dist-items { display: flex; flex-direction: column; gap: 0.4rem; }
.dist-item { display: flex; align-items: center; gap: 0.5rem; font-size: 0.78rem; }
.dist-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.dist-label { color: var(--text-muted); flex: 1; }
.dist-count { color: var(--text-body); font-weight: 500; font-variant-numeric: tabular-nums; }

.gender-panel .donut {
  width: 122px;
  height: 122px;
}

.gender-panel .dist-items {
  min-width: 132px;
}

.bpl-panel .donut {
  width: 122px;
  height: 122px;
}

.bpl-panel .dist-items {
  min-width: 132px;
}

.divyang-panel .donut {
  width: 122px;
  height: 122px;
}

.divyang-panel .dist-items {
  min-width: 132px;
}

.no-data-note {
  margin-top: 1rem;
  padding: 0.75rem 1rem;
  background: var(--bg-surface);
  border-radius: 8px;
  font-size: 0.75rem;
  color: var(--text-dim);
  border-left: 3px solid var(--amber);
}

@media (max-width: 1100px) {
  .dashboard-filter-grid {
    grid-template-columns: 1fr 1fr;
  }

  .dashboard-filter-actions {
    grid-column: 1 / -1;
    justify-content: flex-start;
  }

  .metrics-row {
    grid-template-columns: repeat(2, 1fr);
  }

  .insights-grid {
    grid-template-columns: 1fr;
  }

  .welfare-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .agri-chart-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .dashboard {
    padding: 1rem;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .dashboard-filter-grid {
    grid-template-columns: 1fr;
  }

  .metrics-row,
  .welfare-grid,
  .agri-stats,
  .mini-stats-5,
  .mini-stats-4 {
    grid-template-columns: 1fr;
  }

  .chart-layout {
    grid-template-columns: 1fr;
    justify-items: center;
  }

  .gender-chart-layout {
    flex-direction: column;
    gap: 0.8rem;
  }

  .distribution-row {
    grid-template-columns: 95px 1fr 50px;
    gap: 0.5rem;
  }

  .distribution-label {
    text-align: left;
  }

  .land-bar-item {
    grid-template-columns: 1fr;
    gap: 0.25rem;
  }

  .land-bar-label {
    text-align: left;
  }
}
</style>
