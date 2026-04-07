<template>
  <div class="dashboard">
    <header class="page-header">
      <div>
        <h1 class="page-title">Village Command Center</h1>
        <p class="page-subtitle">Simple view of governance, agriculture, and welfare signals</p>
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
      <!-- Key Metrics Row -->
      <div class="metrics-row">
        <div class="metric-card" v-for="m in metrics" :key="m.label">
          <div class="metric-icon" :style="{ background: m.iconBg }">
            <component :is="m.iconSvg" class="metric-svg" />
          </div>
          <div class="metric-body">
            <div class="metric-value" :style="{ color: m.color }">{{ m.value }}</div>
            <div class="metric-label">{{ m.label }}</div>
          </div>
        </div>
      </div>

      <!-- Insights Grid -->
      <div class="insights-grid">
        <!-- Governance Panel -->
        <div class="card insight-panel">
          <div class="panel-header">
            <h2 class="card-title">Governance Insights</h2>
          </div>
          <div class="insight-bars">
            <div class="insight-bar" v-for="item in governanceBars" :key="item.label">
              <div class="bar-header">
                <span class="bar-label">{{ item.label }}</span>
                <span class="bar-value" :style="{ color: item.color }">{{ item.value }}</span>
              </div>
              <div class="bar-track">
                <div class="bar-fill" :style="{ width: item.pct + '%', background: item.color }"></div>
              </div>
            </div>
          </div>
          <div class="distribution-section" v-if="governance.latrineDistribution?.length">
            <h3 class="dist-title">Sanitation Breakdown</h3>
            <div class="dist-items">
              <div class="dist-item" v-for="d in governance.latrineDistribution" :key="d.label">
                <span class="dist-dot" :style="{ background: sanitationColor(d.label) }"></span>
                <span class="dist-label">{{ d.label || 'Not Recorded' }}</span>
                <span class="dist-count">{{ d.count.toLocaleString() }}</span>
              </div>
            </div>
          </div>
          <div v-else class="no-data-note">
            <span>Info:</span> sanitation columns are not available yet; restart backend with updated code.
          </div>
        </div>

        <!-- Agriculture Panel -->
        <div class="card insight-panel">
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

        <!-- Welfare Panel -->
        <div class="card insight-panel welfare-panel">
          <div class="panel-header">
            <h2 class="card-title">Welfare &amp; Scheme Gaps</h2>
          </div>
          <div class="welfare-grid">
            <div class="welfare-card critical">
              <div class="welfare-num">{{ welfare.bplWithoutToilet?.toLocaleString() ?? '—' }}</div>
              <div class="welfare-desc">BPL households without toilet</div>
              <div class="welfare-tag">Swachh Bharat eligible</div>
            </div>
            <div class="welfare-card warning">
              <div class="welfare-num">{{ welfare.bplWithoutElectricity?.toLocaleString() ?? '—' }}</div>
              <div class="welfare-desc">BPL households without electricity</div>
              <div class="welfare-tag">Saubhagya eligible</div>
            </div>
            <div class="welfare-card info">
              <div class="welfare-num">{{ welfare.smallFarmersWithoutIrrigation?.toLocaleString() ?? '—' }}</div>
              <div class="welfare-desc">Small farmers without irrigation</div>
              <div class="welfare-tag">PM-KISAN priority</div>
            </div>
            <div class="welfare-card neutral">
              <div class="welfare-num">{{ welfare.bplHouseholds?.toLocaleString() ?? '—' }}</div>
              <div class="welfare-desc">Total BPL households</div>
              <div class="welfare-tag">Below poverty line</div>
            </div>
          </div>
          <div class="distribution-section" v-if="welfare.rationCardDistribution?.length">
            <h3 class="dist-title">Ration Card Classification</h3>
            <div class="dist-items">
              <div class="dist-item" v-for="d in welfare.rationCardDistribution" :key="d.label">
                <span class="dist-dot" :style="{ background: rationColor(d.label) }"></span>
                <span class="dist-label">{{ d.label || 'Not Recorded' }}</span>
                <span class="dist-count">{{ d.count.toLocaleString() }}</span>
              </div>
            </div>
          </div>
          <div v-else-if="welfare._note" class="no-data-note">
            <span>ℹ️</span> {{ welfare._note }}
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import { getGovernanceInsights, getAgricultureInsights, getWelfareInsights } from '../api/index.js'

const loading = ref(true)
const governance = ref({})
const agriculture = ref({})
const welfare = ref({})

onMounted(async () => {
  const results = await Promise.allSettled([
    getGovernanceInsights(),
    getAgricultureInsights(),
    getWelfareInsights()
  ])
  if (results[0].status === 'fulfilled') governance.value = results[0].value
  if (results[1].status === 'fulfilled') agriculture.value = results[1].value
  if (results[2].status === 'fulfilled') welfare.value = results[2].value
  loading.value = false
})

// Simple inline SVG components
const HouseIcon = { render: () => h('svg', { viewBox: '0 0 20 20', fill: 'currentColor' }, [h('path', { d: 'M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z' })]) }
const GlobeIcon = { render: () => h('svg', { viewBox: '0 0 20 20', fill: 'currentColor' }, [h('path', { 'fill-rule': 'evenodd', d: 'M10 18a8 8 0 100-16 8 8 0 000 16zM4.332 8.027a6.012 6.012 0 011.912-2.706C6.512 5.73 6.974 6 7.5 6A1.5 1.5 0 019 7.5V8a2 2 0 004 0 2 2 0 011.523-1.943A5.977 5.977 0 0116 10c0 .34-.028.675-.083 1H15a2 2 0 00-2 2v2.197A5.973 5.973 0 0110 16v-2a2 2 0 00-2-2 2 2 0 01-2-2 2 2 0 00-1.668-1.973z', 'clip-rule': 'evenodd' })]) }
const WarnIcon  = { render: () => h('svg', { viewBox: '0 0 20 20', fill: 'currentColor' }, [h('path', { 'fill-rule': 'evenodd', d: 'M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z', 'clip-rule': 'evenodd' })]) }
const BoltIcon  = { render: () => h('svg', { viewBox: '0 0 20 20', fill: 'currentColor' }, [h('path', { 'fill-rule': 'evenodd', d: 'M11.3 1.046A1 1 0 0112 2v5h4a1 1 0 01.82 1.573l-7 10A1 1 0 018 18v-5H4a1 1 0 01-.82-1.573l7-10a1 1 0 011.12-.38z', 'clip-rule': 'evenodd' })]) }

const metrics = computed(() => {
  const g = governance.value
  return [
    { label: 'Total Households', value: g.totalHouseholds?.toLocaleString() || '—', color: 'var(--text-primary)', iconBg: 'var(--amber-dim)', iconSvg: HouseIcon },
    { label: 'Geo-Mapped',       value: g.householdsWithGeoData?.toLocaleString() || '—', color: 'var(--teal)', iconBg: 'var(--teal-dim)', iconSvg: GlobeIcon },
    { label: 'No Sanitation',    value: g.householdsWithoutToilet?.toLocaleString() || '—', color: 'var(--red)', iconBg: 'var(--red-dim)', iconSvg: WarnIcon },
    { label: 'No Electricity',   value: g.householdsWithoutElectricity?.toLocaleString() || '—', color: 'var(--amber)', iconBg: 'var(--amber-dim)', iconSvg: BoltIcon },
  ]
})

const governanceBars = computed(() => {
  const g = governance.value
  const total = g.totalHouseholds || 1
  return [
    { label: 'Without Toilet',      value: (g.householdsWithoutToilet || 0).toLocaleString(),      pct: ((g.householdsWithoutToilet || 0) / total * 100).toFixed(1), color: 'var(--red)' },
    { label: 'Without Electricity', value: (g.householdsWithoutElectricity || 0).toLocaleString(), pct: ((g.householdsWithoutElectricity || 0) / total * 100).toFixed(1), color: 'var(--amber)' },
    { label: 'Geo-Mapped',          value: (g.householdsWithGeoData || 0).toLocaleString(),         pct: ((g.householdsWithGeoData || 0) / total * 100).toFixed(1), color: 'var(--teal)' },
  ]
})

function landPct(count) {
  const max = Math.max(...(agriculture.value.landDistribution || []).map(d => d.count), 1)
  return (count / max * 100).toFixed(1)
}

function sanitationColor(label) {
  const l = (label || '').toLowerCase()
  if (!l || l === 'no latrine' || l === 'none' || l === 'unknown') return 'var(--red)'
  if (l.includes('pit') || l.includes('open')) return 'var(--amber)'
  return 'var(--green)'
}

function rationColor(label) {
  if (!label) return 'var(--text-dim)'
  const l = label.toLowerCase()
  if (l.includes('bpl') || l.includes('antyodaya')) return 'var(--red)'
  if (l.includes('apl')) return 'var(--green)'
  return 'var(--amber)'
}
</script>

<style scoped>
.dashboard {
  padding: 1.5rem 2rem;
  max-width: 1220px;
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

/* Insights Grid */
.insights-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.9rem;
}
.insight-panel { padding: 1.1rem; }
.welfare-panel { grid-column: 1 / -1; }

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.9rem;
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
  .agri-stats {
    grid-template-columns: 1fr;
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
