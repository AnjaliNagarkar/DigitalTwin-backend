<template>
  <div class="farmers-page">
    <header class="page-header">
      <div>
        <h1 class="page-title">Farmer Registry</h1>
        <p class="page-subtitle">Household-level agricultural records from survey data</p>
      </div>
      <div class="header-controls">
        <div class="search-box">
          <svg viewBox="0 0 20 20" fill="currentColor" class="search-icon">
            <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd"/>
          </svg>
          <input v-model="search" placeholder="Search by name..." class="search-input" />
        </div>
        <div class="filter-chips">
          <button v-for="f in filters" :key="f.value" class="chip"
            :class="{ active: activeFilter === f.value }"
            @click="activeFilter = activeFilter === f.value ? '' : f.value">
            {{ f.label }}
          </button>
        </div>
      </div>
    </header>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <span>Fetching farmer records...</span>
    </div>

    <div v-else class="table-container">
      <div class="table-info">
        Showing <strong>{{ filteredFarmers.length }}</strong> of {{ farmers.length }} records
      </div>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th class="th-index">#</th>
              <th @click="toggleSort('firstName')" class="sortable">
                First Name
                <span v-if="sortKey === 'firstName'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="toggleSort('lastName')" class="sortable">
                Last Name
                <span v-if="sortKey === 'lastName'" class="sort-arrow">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th>Owns Agri Land</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(farmer, i) in paginatedFarmers" :key="i" class="table-row">
              <td class="td-index">{{ (currentPage - 1) * pageSize + i + 1 }}</td>
              <td class="td-name">{{ farmer.firstName || '—' }}</td>
              <td class="td-name">{{ farmer.lastName || '—' }}</td>
              <td>
                <span class="land-badge" :class="landClass(farmer.ownAgricultureLand)">
                  {{ farmer.ownAgricultureLand || 'Unknown' }}
                </span>
              </td>
              <td>
                <span class="status-dot" :class="farmer.ownAgricultureLand === 'Yes' ? 'active' : 'inactive'"></span>
              </td>
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
import { getFarmers } from '../../api/index.js'

const loading = ref(true)
const farmers = ref([])
const search = ref('')
const activeFilter = ref('')
const sortKey = ref('')
const sortDir = ref('asc')
const currentPage = ref(1)
const pageSize = 50

const filters = [
  { label: 'Land Owners', value: 'Yes' },
  { label: 'No Land',     value: 'No'  },
]

onMounted(async () => {
  try { farmers.value = await getFarmers() }
  catch (e) { console.error(e) }
  finally { loading.value = false }
})

function toggleSort(key) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
}

const filteredFarmers = computed(() => {
  let list = [...farmers.value]
  if (search.value) {
    const s = search.value.toLowerCase()
    list = list.filter(f => (f.firstName||'').toLowerCase().includes(s) || (f.lastName||'').toLowerCase().includes(s))
  }
  if (activeFilter.value) list = list.filter(f => f.ownAgricultureLand === activeFilter.value)
  if (sortKey.value) {
    list.sort((a, b) => {
      const av = (a[sortKey.value]||'').toLowerCase()
      const bv = (b[sortKey.value]||'').toLowerCase()
      return sortDir.value === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av)
    })
  }
  return list
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredFarmers.value.length / pageSize)))
const paginatedFarmers = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredFarmers.value.slice(start, start + pageSize)
})

watch([search, activeFilter], () => { currentPage.value = 1 })

function landClass(val) {
  if (val === 'Yes') return 'land-yes'
  if (val === 'No')  return 'land-no'
  return 'land-unknown'
}
</script>

<style scoped>
.farmers-page { padding: 2rem 2.5rem; max-width: 1200px; }

.page-header { margin-bottom: 1.5rem; }

.page-title {
  font-family: var(--font-display);
  font-size: 2rem;
  color: var(--text-primary);
  font-weight: 400;
}
.page-subtitle { color: var(--text-dim); font-size: 0.8rem; margin-top: 0.35rem; }

.header-controls { display: flex; align-items: center; gap: 1rem; margin-top: 1.25rem; }

.search-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0.5rem 0.75rem;
  flex: 1;
  max-width: 320px;
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

.filter-chips { display: flex; gap: 0.5rem; }
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

.status-dot { display: inline-block; width: 7px; height: 7px; border-radius: 50%; }
.status-dot.active   { background: var(--green); box-shadow: 0 0 6px var(--green-dim); }
.status-dot.inactive { background: var(--text-dim); }

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
