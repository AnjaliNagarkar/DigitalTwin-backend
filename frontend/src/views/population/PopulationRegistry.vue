<template>
  <div class="farmers-page">
    <header class="page-header">
      <div>
        <h1 class="page-title">{{ t('registry.title') }}</h1>
        <p class="page-subtitle">{{ t('registry.subtitle') }}</p>
      </div>
      <div class="header-controls">
        <div class="search-box">
          <svg viewBox="0 0 20 20" fill="currentColor" class="search-icon">
            <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd"/>
          </svg>
          <input v-model="search" :placeholder="t('common.searchByName')" class="search-input" />
        </div>
        <div class="filter-chips">
          <button v-for="filter in filters" :key="filter.value" class="chip"
            :class="{ active: activeFilter === filter.value }"
            @click="activeFilter = activeFilter === filter.value ? '' : filter.value">
            {{ filter.label }}
          </button>
        </div>
      </div>
    </header>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <span>{{ t('registry.loading') }}</span>
    </div>

    <div v-else class="table-container">
      <div class="table-info">
        {{ t('common.showing') }} <strong>{{ filteredRecords.length }}</strong> {{ t('common.of') }} {{ records.length }} {{ t('common.records') }}
      </div>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th class="th-index">#</th>
              <th>{{ t('registry.colName') }}</th>
              <th>{{ t('registry.colGender') }}</th>
              <th>{{ t('registry.colAge') }}</th>
              <th>{{ t('registry.colEducation') }}</th>
              <th>{{ t('registry.colOccupation') }}</th>
              <th>{{ t('registry.colDisability') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(record, index) in paginatedRecords" :key="`${record.name}-${index}`" class="table-row">
              <td class="td-index">{{ (currentPage - 1) * pageSize + index + 1 }}</td>
              <td class="td-name">{{ record.name || '—' }}</td>
              <td>{{ record.gender || '—' }}</td>
              <td>{{ record.age || '—' }}</td>
              <td>{{ record.education || t('common.notAvailable') }}</td>
              <td>{{ record.occupation || t('common.notWorking') }}</td>
              <td>{{ record.disabilityType || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="pagination">
        <button class="pg-btn" :disabled="currentPage <= 1" @click="currentPage--">{{ t('common.prev') }}</button>
        <span class="pg-info">{{ t('common.pageOf', { current: currentPage, total: totalPages }) }}</span>
        <button class="pg-btn" :disabled="currentPage >= totalPages" @click="currentPage++">{{ t('common.next') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getPopulationRegistry } from './api.js'

const { t } = useI18n()

const loading = ref(true)
const records = ref([])
const search = ref('')
const activeFilter = ref('')
const currentPage = ref(1)
const pageSize = 50

const filters = computed(() => [
  { label: t('registry.filterAll'),     value: '' },
  { label: t('registry.filterBpl'),     value: 'bpl' },
  { label: t('registry.filterStudent'), value: 'student' },
  { label: t('registry.filterDivyang'), value: 'divyang' },
])

onMounted(async () => {
  await loadRecords()
})

const filteredRecords = computed(() => {
  let list = [...records.value]
  if (search.value) {
    const term = search.value.toLowerCase()
    list = list.filter((record) => (record.name || '').toLowerCase().includes(term))
  }
  return list
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredRecords.value.length / pageSize)))
const paginatedRecords = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredRecords.value.slice(start, start + pageSize)
})

watch(search, () => {
  currentPage.value = 1
})

watch(activeFilter, async () => {
  currentPage.value = 1
  await loadRecords()
})

async function loadRecords() {
  loading.value = true
  try {
    const params = {}
    if (activeFilter.value) params.filter = activeFilter.value
    records.value = await getPopulationRegistry(params)
  } catch (error) {
    console.error(error)
    records.value = []
  } finally {
    loading.value = false
  }
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

.header-controls { display: flex; align-items: center; gap: 1rem; margin-top: 1.25rem; flex-wrap: wrap; }

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

.filter-chips { display: flex; gap: 0.5rem; flex-wrap: wrap; }
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

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 3rem 0;
  color: var(--text-dim);
}
.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid var(--border-light);
  border-top-color: var(--amber);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>