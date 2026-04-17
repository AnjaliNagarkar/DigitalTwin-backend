<template>
  <div class="citizens-page">
    <header class="page-header">
      <div>
        <h1 class="page-title">Citizen Registry</h1>
        <p class="page-subtitle">Household-level records from survey data</p>
      </div>
      <div class="header-controls">
        <div class="search-box">
          <svg viewBox="0 0 20 20" fill="currentColor" class="search-icon">
            <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd"/>
          </svg>
          <input v-model="search" placeholder="Search by name..." class="search-input" />
        </div>
        <div class="filter-controls">
          <button v-for="f in filters" :key="f.value" class="chip"
            :class="{ active: activeFilter === f.value }"
            @click="activeFilter = activeFilter === f.value ? '' : f.value">
            {{ f.label }}
          </button>
          <div class="filter-divider"></div>
          <button v-for="w in waterFilters" :key="w.value" class="chip"
            :class="{ active: irrigationFilter === w.value }"
            @click="irrigationFilter = irrigationFilter === w.value ? '' : w.value">
            {{ w.label }}
          </button>
          <select v-model="occupationFilter" class="filter-select" aria-label="Filter by occupation">
            <option value="">All Occupations</option>
            <option v-for="o in occupationOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
          <select v-model="incomeFilter" class="filter-select" aria-label="Filter by income range">
            <option value="">All Income Ranges</option>
            <option v-for="r in incomeOptions" :key="r.value" :value="r.value">{{ r.label }}</option>
          </select>
        </div>
        <button class="reset-btn" @click="resetFilters">Reset Filters</button>
      </div>
    </header>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <span>Fetching citizen records...</span>
    </div>

    <div v-else class="table-container">
      <div class="table-info">
        Showing <strong>{{ filteredCitizens.length }}</strong> of {{ citizens.length }} records
      </div>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th class="th-index">#</th>
              <th @click="toggleSort('fullName')" class="sortable th-name">
                Full Name
                <span v-if="sortKey === 'fullName'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="toggleSort('totalLand')" class="sortable">
                Land (Acre)
                <span v-if="sortKey === 'totalLand'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="toggleSort('waterSource')" class="sortable">
                Water Source
                <span v-if="sortKey === 'waterSource'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="toggleSort('irrigationType')" class="sortable">
                Irrigation
                <span v-if="sortKey === 'irrigationType'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="toggleSort('workDetails')" class="sortable">
                What They Do
                <span v-if="sortKey === 'workDetails'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="toggleSort('annualIncome')" class="sortable">
                Income
                <span v-if="sortKey === 'annualIncome'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="toggleSort('childrenCount')" class="sortable">
                Children
                <span v-if="sortKey === 'childrenCount'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(citizen, i) in paginatedCitizens" :key="i" class="table-row">
              <td class="td-index">{{ (currentPage - 1) * pageSize + i + 1 }}</td>
              <td class="td-name">{{ formatFullName(citizen.firstName, citizen.lastName) }}</td>
              <td class="td-num">{{ formatLand(citizen.totalLand) }}</td>
              <td class="td-water">
                <span class="water-badge" :class="waterClass(citizen.waterSource)">
                  {{ citizen.waterSource || '—' }}
                </span>
              </td>
              <td>
                <span class="ration-badge" :class="waterClass(citizen.waterSource)">
                  {{ getIrrigationType(citizen.waterSource) }}
                </span>
              </td>
              <td class="td-name">{{ formatWorkDetails(citizen.workDetails, citizen.occupation) }}</td>
              <td class="td-num">{{ formatIncome(citizen.annualIncome) }}</td>
              <td class="td-num">{{ citizen.childrenCount ?? 0 }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="pagination">
        <button class="pg-btn" :disabled="currentPage <= 1" @click="currentPage--">← Prev</button>
        <span class="pg-info">Page {{ currentPage }} of {{ totalPages }}</span>
        <button class="pg-btn" :disabled="currentPage >= totalPages" @click="currentPage++">Next →</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { getCitizens } from '../../api/index.js'

const loading = ref(true)
const citizens = ref([])
const search = ref('')
const activeFilter = ref('')
const sortKey = ref('')
const sortDir = ref('asc')
const currentPage = ref(1)
const pageSize = 50

const filters = [
  { label: 'Land Owners', value: 'with_land' },
  { label: 'No Land',     value: 'no_land'   },
]

const waterFilters = [
  { label: 'Irrigated',  value: 'Irrigated' },
  { label: 'Rain-fed',   value: 'Rain-fed'  },
]

const occupationOptions = [
  { label: 'Farm Based', value: 'farm_based' },
  { label: 'Non-farm Based', value: 'non_farm_based' },
  { label: 'Wage Work', value: 'wage_work' },
  { label: 'Studying', value: 'studying' },
  { label: 'Housewife', value: 'housewife' },
]

const irrigationFilter = ref('')
const occupationFilter = ref('')
const incomeFilter = ref('')

const incomeOptions = [
  { label: 'Less than 21,000', value: 'lt_21000' },
  { label: '21,001 to 50,000', value: '21001_50000' },
  { label: '50,001+', value: 'gt_50000' },
]

onMounted(async () => {
  try { citizens.value = await getCitizens() }
  catch (e) { console.error(e) }
  finally { loading.value = false }
})

function toggleSort(key) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
}

function capitalizeWords(text) {
  return String(text || '')
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

function formatFullName(firstName, lastName) {
  const full = `${firstName || ''} ${lastName || ''}`.trim()
  return full ? capitalizeWords(full) : '—'
}

function formatLand(value) {
  if (value === null || value === undefined || value === '') return '—'
  return String(value)
}

function getIrrigationType(waterSource) {
  const v = String(waterSource || '').toLowerCase()
  if (!v) return 'Rain-fed'
  if (v.includes('rain') || v === 'none') return 'Rain-fed'
  return 'Irrigated'
}

function formatIncome(value) {
  if (value === null || value === undefined || String(value).trim() === '') return '—'
  return String(value)
}

function formatWorkDetails(workDetails, occupation) {
  const value = String(workDetails || occupation || '').trim()
  if (!value) return 'Not Working'
  return capitalizeWords(value)
}

function parseLandValue(value) {
  const num = Number(String(value ?? '').replace(/,/g, '').trim())
  return Number.isFinite(num) ? num : 0
}

function classifyOccupation(citizen) {
  const value = String(citizen.workDetails || citizen.occupation || '').toLowerCase().trim()
  if (!value) return ''

  if (value.includes('housewife')) return 'housewife'
  if (value.includes('stud')) return 'studying'
  if (value.includes('wage')) return 'wage_work'
  if (value.includes('non-farm') || value.includes('non farm')) return 'non_farm_based'
  if (value.includes('farm based') || value.includes('agri') || value.includes('cultivator') || value.includes('farmer')) return 'farm_based'

  return ''
}

function parseIncomeNumber(value) {
  const raw = String(value ?? '').trim()
  if (!raw) return Number.NaN
  const cleaned = raw.replace(/[^0-9]/g, '')
  if (!cleaned) return Number.NaN
  return Number(cleaned)
}

function classifyIncome(value) {
  const text = String(value ?? '').toLowerCase().trim()
  if (!text) return ''

  if (text.includes('less than') && text.includes('21')) return 'lt_21000'
  if ((text.includes('21') || text.includes('21001')) && text.includes('50')) return '21001_50000'
  if (text.includes('50,001') || text.includes('50001') || text.includes('50,000+') || text.includes('50000+') || text.includes('above')) return 'gt_50000'

  const income = parseIncomeNumber(value)
  if (!Number.isFinite(income)) return ''
  if (income <= 21000) return 'lt_21000'
  if (income <= 50000) return '21001_50000'
  return 'gt_50000'
}

function resetFilters() {
  search.value = ''
  activeFilter.value = ''
  irrigationFilter.value = ''
  occupationFilter.value = ''
  incomeFilter.value = ''
}

const filteredCitizens = computed(() => {
  let list = [...citizens.value]
  if (search.value) {
    const s = search.value.toLowerCase()
    list = list.filter(f => (f.firstName||'').toLowerCase().includes(s) || (f.lastName||'').toLowerCase().includes(s))
  }

  if (activeFilter.value) {
    if (activeFilter.value === 'with_land') {
      list = list.filter(f => parseLandValue(f.totalLand) > 0)
    } else if (activeFilter.value === 'no_land') {
      list = list.filter(f => parseLandValue(f.totalLand) === 0)
    }
  }

  if (irrigationFilter.value) {
    list = list.filter(f => getIrrigationType(f.waterSource) === irrigationFilter.value)
  }

  if (occupationFilter.value) {
    list = list.filter(f => classifyOccupation(f) === occupationFilter.value)
  }

  if (incomeFilter.value) {
    list = list.filter(f => classifyIncome(f.annualIncome) === incomeFilter.value)
  }

  if (sortKey.value) {
    list.sort((a, b) => {
      if (sortKey.value === 'totalLand') {
        const av = parseFloat(a.totalLand) || 0
        const bv = parseFloat(b.totalLand) || 0
        return sortDir.value === 'asc' ? av - bv : bv - av
      }

      if (sortKey.value === 'childrenCount') {
        const av = Number(a.childrenCount) || 0
        const bv = Number(b.childrenCount) || 0
        return sortDir.value === 'asc' ? av - bv : bv - av
      }

      if (sortKey.value === 'fullName') {
        const av = formatFullName(a.firstName, a.lastName).toLowerCase()
        const bv = formatFullName(b.firstName, b.lastName).toLowerCase()
        return sortDir.value === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av)
      }

      if (sortKey.value === 'irrigationType') {
        const av = getIrrigationType(a.waterSource).toLowerCase()
        const bv = getIrrigationType(b.waterSource).toLowerCase()
        return sortDir.value === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av)
      }

      if (sortKey.value === 'annualIncome') {
        const av = parseIncomeNumber(a.annualIncome)
        const bv = parseIncomeNumber(b.annualIncome)
        if (Number.isFinite(av) && Number.isFinite(bv)) {
          return sortDir.value === 'asc' ? av - bv : bv - av
        }
      }

      const av = String(a[sortKey.value] || '').toLowerCase()
      const bv = String(b[sortKey.value] || '').toLowerCase()
      return sortDir.value === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av)
    })
  }
  return list
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredCitizens.value.length / pageSize)))
const paginatedCitizens = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredCitizens.value.slice(start, start + pageSize)
})

watch([search, activeFilter, irrigationFilter, occupationFilter, incomeFilter], () => { currentPage.value = 1 })

function waterClass(val) {
  if (!val) return 'water-unknown'
  const v = val.toLowerCase()
  if (v.includes('rain') || v === 'none') return 'water-rainfed'
  return 'water-irrigated'
}
</script>

<style scoped>
.citizens-page { padding: 2rem 2.5rem; max-width: none; width: 100%; }

.page-header { margin-bottom: 1.5rem; }

.page-title {
  font-family: var(--font-display);
  font-size: 2rem;
  color: var(--text-primary);
  font-weight: 400;
}
.page-subtitle { color: var(--text-dim); font-size: 0.8rem; margin-top: 0.35rem; }

.header-controls {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  margin-top: 1.25rem;
  flex-wrap: wrap;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0.5rem 0.75rem;
  flex: 0 1 320px;
  transition: border-color 0.2s;
}
.search-box:focus-within { border-color: var(--amber); }
.search-icon { width: 16px; height: 16px; color: var(--text-dim); flex-shrink: 0; }
.search-input {
  background: none; border: none; outline: none;
  color: var(--text-body);
  font-family: var(--font-body); font-size: 0.85rem; width: 100%;
}
.search-input::placeholder { color: var(--text-dim); }

.filter-controls {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
  flex: 1 1 420px;
}
.filter-divider { width: 1px; height: 20px; background: var(--border); }
.chip {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 20px;
  padding: 0.35rem 0.85rem;
  font-size: 0.75rem;
  color: var(--text-muted);
  cursor: pointer;
  font-family: var(--font-body);
  transition: all 0.2s;
}
.chip:hover { border-color: var(--border-light); color: var(--text-body); }
.chip.active { background: var(--amber-dim); border-color: var(--amber); color: var(--amber); }

.filter-select {
  height: 32px;
  min-width: 170px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text-body);
  font-family: var(--font-body);
  font-size: 0.78rem;
  padding: 0 0.65rem;
}
.filter-select:focus {
  outline: none;
  border-color: var(--amber);
}

.reset-btn {
  margin-left: auto;
  height: 32px;
  padding: 0 0.9rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--bg-surface);
  color: var(--text-muted);
  font-family: var(--font-body);
  font-size: 0.76rem;
  cursor: pointer;
  transition: all 0.2s;
}
.reset-btn:hover {
  border-color: var(--amber);
  color: var(--amber);
}

/* Table */
.table-container {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
  overflow: hidden;
}

.table-info {
  padding: 0.75rem 1.25rem;
  border-bottom: 1px solid var(--border);
  font-size: 0.75rem;
  color: var(--text-dim);
}
.table-info strong { color: var(--text-body); }
.table-wrap { overflow-x: auto; }

.data-table { width: 100%; border-collapse: collapse; }

.data-table th {
  text-align: left;
  padding: 0.65rem 1.25rem;
  font-size: 0.65rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-dim);
  font-weight: 600;
  border-bottom: 1px solid var(--border);
  white-space: nowrap;
  user-select: none;
  background: var(--bg-surface);
}
.sortable { cursor: pointer; }
.sortable:hover { color: var(--text-body); }
.sort-arrow { color: var(--amber); margin-left: 0.25rem; }
.th-index { width: 50px; }
.th-name { min-width: 200px; }

.table-row { transition: background 0.15s; }
.table-row:hover { background: var(--bg-surface); }
.table-row td {
  padding: 0.6rem 1.25rem;
  border-bottom: 1px solid var(--border);
  font-size: 0.82rem;
}

.td-index { color: var(--text-dim); font-variant-numeric: tabular-nums; font-size: 0.75rem; }
.td-name  { color: var(--text-body); }

.land-badge {
  display: inline-block;
  padding: 0.15rem 0.55rem;
  border-radius: 4px;
  font-size: 0.72rem;
  font-weight: 500;
}
.land-yes     { background: var(--green-dim); color: var(--green); }
.land-no      { background: var(--red-dim);   color: var(--red); }
.land-unknown { background: var(--bg-surface); color: var(--text-dim); }

.td-num   { font-variant-numeric: tabular-nums; color: var(--text-body); font-size: 0.8rem; }
.td-water { max-width: 160px; }

.water-badge {
  display: inline-block;
  padding: 0.15rem 0.55rem;
  border-radius: 4px;
  font-size: 0.72rem;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 140px;
}
.water-irrigated { background: #dcfce7; color: #15803d; }
.water-rainfed   { background: #fee2e2; color: #b91c1c; }
.water-unknown   { background: var(--bg-surface); color: var(--text-dim); }

.ration-badge {
  display: inline-block;
  padding: 0.15rem 0.55rem;
  border-radius: 4px;
  font-size: 0.72rem;
  font-weight: 500;
  white-space: nowrap;
}

/* Pagination */
.pagination {
  display: flex; align-items: center; justify-content: center;
  gap: 1rem; padding: 0.85rem 1.25rem;
  border-top: 1px solid var(--border);
}
.pg-btn {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 0.35rem 0.85rem;
  font-size: 0.75rem; color: var(--text-muted);
  cursor: pointer; font-family: var(--font-body);
  transition: all 0.15s;
}
.pg-btn:hover:not(:disabled) { border-color: var(--amber); color: var(--amber); }
.pg-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.pg-info { font-size: 0.75rem; color: var(--text-dim); font-variant-numeric: tabular-nums; }
</style>
