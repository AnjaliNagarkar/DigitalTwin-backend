<template>
  <div class="dashboard">
    <header class="page-header">
      <div>
        <h1 class="page-title">Village Command Center</h1>
        <p class="page-subtitle">Unified population and agriculture intelligence</p>
      </div>
      <div class="header-badge">
        <span class="badge-dot"></span>
        <span>Live API</span>
      </div>
    </header>

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

          <article class="card insight-panel age-panel">
            <div class="panel-header">
              <h3 class="chart-title">Age Distribution</h3>
              <span class="total-note">Total: {{ ageTotal.toLocaleString() }}</span>
            </div>

            <div v-if="ageTotal === 0" class="empty-state">
              No age records available.
            </div>
            <div v-else class="distribution-bars age-distribution-bars">
              <div class="distribution-row" v-for="item in ageSegments" :key="item.label">
                <div class="distribution-label">{{ item.label }}</div>
                <div class="distribution-track">
                  <div class="distribution-fill" :style="{ width: ageBarWidth(item.value), background: item.color }"></div>
                </div>
                <div class="distribution-value">{{ item.value.toLocaleString() }}</div>
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
        </div>
      </section>

      <section class="intelligence-layout">
        <div class="insights-grid intelligence-row-1">
          <!-- Agriculture Panel -->
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
            <div class="distribution-section" v-if="agriculture.landDistribution?.length">
              <h3 class="dist-title">Land Holdings Distribution</h3>
              <div class="land-bars">
                <div class="land-bar-item" v-for="d in agriculture.landDistribution" :key="d.label">
                  <div class="land-bar-label">{{ d.label }}</div>
                  <div class="land-bar-track">
                    <div class="land-bar-fill" :style="{ width: landPct(d.count) + '%' }"></div>
                  </div>
                  <div class="land-bar-count">{{ d.count.toLocaleString() }}</div>
                </div>
              </div>
            </div>
          </div>

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
        </div>

        <div class="insights-grid intelligence-row-2">
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
      </section>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import { getAgricultureInsights } from '../../api/index.js'
import { getPopulationDashboard, getPopulationDemographics, getPopulationEducation, getPopulationEmployment, getPopulationMapData } from '../population/api.js'

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
})
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
const families = ref([])

onMounted(async () => {
  const agricultureResults = await Promise.allSettled([
    getAgricultureInsights(),
  ])

  const populationResults = await Promise.allSettled([
    getPopulationDashboard(),
    getPopulationDemographics(),
    getPopulationEducation(),
    getPopulationEmployment(),
    getPopulationMapData(),
  ])

  if (agricultureResults[0].status === 'fulfilled') agriculture.value = agricultureResults[0].value

  if (populationResults[0].status === 'fulfilled') populationStats.value = populationResults[0].value
  if (populationResults[1].status === 'fulfilled') demographics.value = populationResults[1].value
  if (populationResults[2].status === 'fulfilled') education.value = populationResults[2].value
  if (populationResults[3].status === 'fulfilled') employment.value = populationResults[3].value
  if (populationResults[4].status === 'fulfilled') {
    const rows = Array.isArray(populationResults[4].value) ? populationResults[4].value : []
    families.value = rows
  }

  loading.value = false
})

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
    { label: 'Total Population', value: Number(s.total_population || 0).toLocaleString(), color: 'var(--text-primary)', iconBg: 'var(--amber-dim)', iconSvg: GroupIcon },
    { label: 'Total Households', value: Number(s.total_households || 0).toLocaleString(), color: 'var(--teal)', iconBg: 'var(--teal-dim)', iconSvg: HomeIcon },
    { label: 'Working Population', value: Number(s.working_population || 0).toLocaleString(), color: 'var(--green)', iconBg: 'var(--green-dim)', iconSvg: BriefcaseIcon },
    { label: 'Dependent Population', value: Number(s.dependent_population || 0).toLocaleString(), color: 'var(--red)', iconBg: 'var(--red-dim)', iconSvg: ChildIcon },
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

const ageSegments = computed(() => {
  const a = demographics.value.age_distribution || {}
  return [
    { label: '0-5', value: Number(a.age_0_5 || 0), color: 'var(--teal)' },
    { label: '6-18', value: Number(a.age_6_18 || 0), color: 'var(--amber)' },
    { label: '19-35', value: Number(a.age_19_35 || 0), color: 'var(--green)' },
    { label: '36-60', value: Number(a.age_36_60 || 0), color: 'var(--red)' },
    { label: '60+', value: Number(a.age_60_plus || 0), color: 'var(--text-dim)' },
  ]
})
const ageTotal = computed(() => ageSegments.value.reduce((sum, item) => sum + item.value, 0))
const ageMax = computed(() => Math.max(...ageSegments.value.map((item) => item.value), 1))
const ageBarWidth = (value) => `${(Number(value || 0) / ageMax.value) * 100}%`

const bplSegments = computed(() => {
  const rows = Array.isArray(families.value) ? families.value : []
  let bpl = 0
  let nonBpl = 0

  for (const family of rows) {
    const category = String(family?.FAMILY_BELONG_BPL_CATEGORY || '').trim().toUpperCase()
    if (category === 'YES') bpl += 1
    if (category === 'NO') nonBpl += 1
  }

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

.intelligence-row-1 .agriculture-panel {
  grid-column: auto;
}

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
  .metrics-row {
    grid-template-columns: repeat(2, 1fr);
  }

  .insights-grid {
    grid-template-columns: 1fr;
  }

  .welfare-grid {
    grid-template-columns: repeat(2, 1fr);
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
