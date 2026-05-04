<template>
  <div class="registry-wrap" :class="{ 'registry-embedded': embedded }">

    <!-- ── Page header (full-page mode only) ─────────────────────────── -->
    <header v-if="!embedded" class="page-header" @click="categoryDropdownOpen = false">

      <!-- Row 1: Title + CATEGORY custom dropdown -->
      <div class="header-row header-row-1">
        <div class="title-block">
          <h1 class="page-title">Citizen Registry</h1>
          <p class="page-subtitle">{{ activeCategoryConfig.subtitle }}</p>
        </div>

        <div class="reg-filter-group" @click.stop>
          <span class="reg-filter-label">CATEGORY</span>
          <div class="reg-custom-select" :class="{ open: categoryDropdownOpen }"
               @click="categoryDropdownOpen = !categoryDropdownOpen">
            <button class="reg-cs-trigger" type="button">
              <span class="reg-cs-value">
                {{ activeCategoryConfig.icon }} {{ activeCategoryConfig.label }}
              </span>
              <span class="reg-cs-arrow">▾</span>
            </button>
            <div class="reg-cs-dropdown" v-show="categoryDropdownOpen" @click.stop>
              <div v-for="c in CATEGORIES" :key="c.value"
                   class="reg-cs-option"
                   :class="{ selected: category === c.value }"
                   @click="selectCategory(c.value)">
                <span class="reg-cs-opt-icon">{{ c.icon }}</span>{{ c.fullLabel }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Row 2: Search + Sub-filters -->
      <div class="header-row header-row-2">
        <div class="search-box">
          <svg viewBox="0 0 20 20" fill="currentColor" class="search-icon">
            <path fill-rule="evenodd"
              d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
              clip-rule="evenodd"/>
          </svg>
          <input v-model="search" placeholder="Search by name…" class="search-input"/>
        </div>

        <template v-if="dynamicSubFilters.length">
          <div class="subfilter-bar">
            <template v-for="sf in dynamicSubFilters" :key="sf.key">
              <span class="subfilter-label">{{ sf.label }}:</span>
              <button v-for="opt in sf.options" :key="opt.value"
                class="chip"
                :class="{ active: isSubFilterActive(sf.key, opt.value) }"
                @click="toggleSubFilter(sf.key, opt.value)">
                {{ opt.label }}
              </button>
              <div class="filter-divider"></div>
            </template>
          </div>
        </template>

        <button class="reset-btn" @click="resetFilters">↺ Reset</button>
      </div>

    </header>

    <!-- ── Embedded compact toolbar ──────────────────────────────────── -->
    <div v-if="embedded" class="embedded-toolbar" @click="categoryDropdownOpen = false">
      <div class="reg-custom-select reg-custom-select--sm"
           :class="{ open: categoryDropdownOpen }"
           @click.stop="categoryDropdownOpen = !categoryDropdownOpen">
        <button class="reg-cs-trigger reg-cs-trigger--sm" type="button">
          <span class="reg-cs-value">{{ activeCategoryConfig.icon }} {{ activeCategoryConfig.label }}</span>
          <span class="reg-cs-arrow">▾</span>
        </button>
        <div class="reg-cs-dropdown" v-show="categoryDropdownOpen" @click.stop>
          <div v-for="c in CATEGORIES" :key="c.value"
               class="reg-cs-option"
               :class="{ selected: category === c.value }"
               @click="selectCategory(c.value)">
            {{ c.icon }} {{ c.fullLabel }}
          </div>
        </div>
      </div>
      <div class="search-box search-box-sm">
        <svg viewBox="0 0 20 20" fill="currentColor" class="search-icon">
          <path fill-rule="evenodd"
            d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
            clip-rule="evenodd"/>
        </svg>
        <input v-model="search" placeholder="Search…" class="search-input"/>
      </div>
      <template v-for="sf in dynamicSubFilters" :key="sf.key">
        <button v-for="opt in sf.options" :key="opt.value"
          class="chip"
          :class="{ active: isSubFilterActive(sf.key, opt.value) }"
          @click="toggleSubFilter(sf.key, opt.value)">
          {{ opt.label }}
        </button>
      </template>
      <button class="reset-btn" @click="resetFilters">↺ Reset</button>
    </div>

    <!-- ── Loading ────────────────────────────────────────────────────── -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <span>Loading citizen registry…</span>
    </div>

    <!-- ── Table ──────────────────────────────────────────────────────── -->
    <div v-else class="table-container">
      <div class="table-info">
        <div class="table-info-main">
          <span>
            Showing <strong>{{ filteredRecords.length }}</strong> of {{ categoryRecords.length }}
            <em>{{ activeCategoryConfig.label.toLowerCase() }}</em> citizens
          </span>
          <div class="gender-legend" style="display: flex; gap: 1.2rem; align-items: center;">
            <span class="legend-item"><span class="gender-dot male"></span> Male</span>
            <span class="legend-item"><span class="gender-dot female"></span> Female</span>
            <span class="legend-item"><span class="gender-dot other"></span> Other</span>
          </div>
          <div class="income-legend" style="display:flex;gap:1rem;align-items:center;font-size:0.75rem;color:var(--text-dim);border-left:1px solid var(--border);padding-left:1rem">
            <span class="legend-item"><span style="width:8px;height:8px;border-radius:50%;display:inline-block;background:#20c997"></span> High</span>
            <span class="legend-item"><span style="width:8px;height:8px;border-radius:50%;display:inline-block;background:#ffc107"></span> Mid</span>
            <span class="legend-item"><span style="width:8px;height:8px;border-radius:50%;display:inline-block;background:#ff4d4f"></span> Low</span>
          </div>
        </div>
        <!-- Active category pill -->
        <span v-if="category" class="active-cat-pill" :style="{ background: activeCategoryConfig.color + '22', color: activeCategoryConfig.color, borderColor: activeCategoryConfig.color + '55' }">
          {{ activeCategoryConfig.icon }} {{ activeCategoryConfig.label }}
        </span>
      </div>

      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th class="th-index">#</th>
              <!-- Dynamic columns based on category -->
              <template v-for="col in activeCategoryConfig.columns" :key="col.key">
                <th
                  :class="['sortable', col.minWidth ? 'th-name' : '']"
                  @click="toggleSort(col.key)">
                  {{ col.label }}
                  <span v-if="sortKey === col.key" class="sort-arrow">
                    {{ sortDir === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
              </template>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(r, i) in paginatedRecords" :key="i" class="table-row">
              <td class="td-index">{{ (currentPage - 1) * pageSize + i + 1 }}</td>
              <template v-for="col in activeCategoryConfig.columns" :key="col.key">
                <td :class="col.tdClass || ''">
                  <component :is="'span'" v-html="renderCell(r, col)"></component>
                </td>
              </template>
            </tr>
            <tr v-if="paginatedRecords.length === 0">
              <td :colspan="activeCategoryConfig.columns.length + 1" class="td-empty">
                No records match the selected filters.
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
defineOptions({ name: 'UnifiedRegistry' })

import { ref, computed, onMounted, onActivated, onBeforeUnmount, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getUnifiedRegistry, getCitizens } from '../api/index.js'
import { getRegistryState } from '../state/registryStateCache.js'

// ── Props ──────────────────────────────────────────────────────────────────
const props = defineProps({
  embedded: { type: Boolean, default: false },
  maxRows:  { type: Number,  default: 0 },
})

const route = useRoute()
const registryScope = computed(() => route.path.startsWith('/population') ? 'population' : 'agriculture')
const registryState = computed(() => getRegistryState(registryScope.value))

// ── Category definitions ───────────────────────────────────────────────────
// Each category declares: icon, label, subtitle, color, which columns to show,
// which sub-filters to expose, and how to pre-filter the rows.

const CATEGORIES = [
  { value: '',          label: 'All',             fullLabel: 'All Citizens',    icon: '👥' },
  { value: 'farmer',   label: 'Farmers',          fullLabel: 'Farmers',         icon: '🌾' },
  { value: 'student',  label: 'Students',         fullLabel: 'Students',        icon: '🎓' },
  { value: 'disabled', label: 'Divyang',          fullLabel: 'Divyang',         icon: '♿' },
  { value: 'housewife',label: 'Homemakers',       fullLabel: 'Homemakers',      icon: '🏠' },
  { value: 'senior',   label: 'Senior Citizens',  fullLabel: 'Senior Citizens', icon: '👴' },
]

const CATEGORY_CONFIG = {
  '': {
    label: 'All Citizens',
    subtitle: 'Complete citizen records from population & agri survey',
    icon: '👥',
    color: '#6b7280',
    rowFilter: () => true,
    columns: [
      { key: 'fullName',     label: 'Full Name',     minWidth: true, tdClass: 'td-name' },
      { key: 'age',          label: 'Age',            tdClass: 'td-num' }, // Re-add age column
      { key: 'occupation',   label: 'Occupation' },
      { key: 'maritalStatus', label: 'Marital Status' },
      { key: 'annualIncome', label: 'Annual Income',  tdClass: 'td-num' },
    ],
    subFilters: [
      { key: 'incomeRange', label: 'Income Range' },
      { key: 'gender', label: 'Gender' },
      { key: 'occupationType', label: 'Occupation' },
      { key: 'isDivyang', label: 'Divyang' },
    ],
  },

  farmer: {
    label: 'Farmer',
    subtitle: 'Citizens bucketed as Farmer by occupation',
    icon: '🌾',
    color: '#16a34a',
    rowFilter: r => getOccupationBucket(r.occupation) === 'Farmer',
    columns: [
      { key: 'fullName',       label: 'Full Name',     minWidth: true, tdClass: 'td-name' },
      { key: 'totalLand',      label: 'Land (Acre)',    tdClass: 'td-num' },
      { key: 'crops',          label: 'Crops' },
      { key: 'irrigationType', label: 'Irrigation' },
      { key: 'waterSource',    label: 'Water Source' },
      { key: 'annualIncome',   label: 'Annual Income',  tdClass: 'td-num' },
      { key: 'occupation',     label: 'Occupation' },
    ],
    subFilters: [
      { key: 'landSize', label: 'Land Size' },
      { key: 'irrigationType', label: 'Irrigation' },
      { key: 'cropType', label: 'Crop' },
      { key: 'gender', label: 'Gender' },
      { key: 'incomeRange', label: 'Income Range' },
      { key: 'maritalStatus', label: 'Marital Status' },
    ],
  },

  student: {
    label: 'Student',
    subtitle: 'Citizens currently pursuing education',
    icon: '📚',
    color: '#2563eb',
    rowFilter: r => r.isStudent,
    columns: [
      { key: 'fullName',       label: 'Full Name',        minWidth: true, tdClass: 'td-name' },
      { key: 'age',            label: 'Age',              tdClass: 'td-num' },
      { key: 'educationLevel', label: 'Education Level' },
      { key: 'schoolName',     label: 'School / College' },
      { key: 'scholarship',    label: 'Scholarship' },
    ],
    subFilters: [
      { key: 'educationLevel', label: 'Level' },
      { key: 'scholarship', label: 'Scholarship' },
      { key: 'ageGroup', label: 'Age Group' },
      { key: 'gender', label: 'Gender' },
    ],
  },

  disabled: {
    label: 'Divyang',
    subtitle: 'Citizens with reported disability (Divyang)',
    icon: '♿',
    color: '#d97706',
    rowFilter: r => r.isDivyang,
    columns: [
      { key: 'fullName',          label: 'Full Name',         minWidth: true, tdClass: 'td-name' },
      { key: 'disabilityType',    label: 'Disability Type' },
      { key: 'disabilityPercent', label: 'Disability %',      tdClass: 'td-num' },
      { key: 'annualIncome',      label: 'Annual Income',     tdClass: 'td-num' },
      { key: 'pensionStatus',     label: 'Pension Status' },
      { key: 'govtPensionAmount', label: 'Govt. Pension Amt', tdClass: 'td-num' },
      { key: 'caretakerName',     label: 'Caretaker' },
      { key: 'divyangCertificate', label: 'Divyang Certificate' },
    ],
    subFilters: [
      { key: 'pensionStatus', label: 'Pension' },
      { key: 'disabilitySeverity', label: 'Disability %' },
      { key: 'divyangCertificate', label: 'Certificate' },
      { key: 'gender', label: 'Gender' },
      { key: 'incomeRange', label: 'Income Range' },
      { key: 'maritalStatus', label: 'Marital Status' },
    ],
  },

  housewife: {
    label: 'Homemaker',
    subtitle: 'Citizens with occupation recorded as homemaker',
    icon: '🏠',
    color: '#db2777',
    rowFilter: r => r.isHousewife,
    columns: [
      { key: 'fullName',       label: 'Full Name',       minWidth: true, tdClass: 'td-name' },
      { key: 'age',            label: 'Age',              tdClass: 'td-num' }, // Re-add age column
      { key: 'annualIncome',   label: 'Annual Income',    tdClass: 'td-num' },
      { key: 'childrenCount',  label: 'Children',         tdClass: 'td-num' },
      { key: 'maritalStatus',  label: 'Marital Status' },
    ],
    subFilters: [
      { key: 'ageGroup', label: 'Age Group' },
      { key: 'childrenGroup', label: 'Children' },
      { key: 'gender', label: 'Gender' },
      { key: 'incomeRange', label: 'Income Range' },
      { key: 'maritalStatus', label: 'Marital Status' },
    ],
  },

  senior: {
    label: 'Senior Citizen',
    subtitle: 'Citizens aged 60 and above',
    icon: '👴',
    color: '#7c3aed',
    rowFilter: r => r.isSenior,
    columns: [
      { key: 'fullName',      label: 'Full Name',      minWidth: true, tdClass: 'td-name' },
      { key: 'age',           label: 'Age',             tdClass: 'td-num' }, // Re-add age column
      { key: 'pensionStatus', label: 'Pension Status' },
      { key: 'annualIncome',  label: 'Annual Income',   tdClass: 'td-num' },
    ],
    subFilters: [
      { key: 'incomeRange', label: 'Income Range' },
      { key: 'maritalStatus', label: 'Marital Status' },
    ],
  },
}

// ── State ──────────────────────────────────────────────────────────────────
const loading     = ref(true)
const records     = ref([])
const category             = ref('')
const categoryDropdownOpen = ref(false)
const search               = ref('')
const subFilters           = ref({})   // { [filterKey]: string[] }

function selectCategory(val) {
  category.value = val
  categoryDropdownOpen.value = false
  onCategoryChange()
  persistToCache()
}
const sortKey     = ref('')
const sortDir     = ref('asc')
const currentPage = ref(1)
const pageSize    = computed(() => props.maxRows > 0 ? props.maxRows : 50)

function hydrateFromCache() {
  const cached = registryState.value
  if (!cached) return false

  if (Array.isArray(cached.records) && cached.records.length > 0) {
    records.value = cached.records
  }
  category.value = cached.category || ''
  search.value = cached.search || ''
  subFilters.value = cached.subFilters || {}
  sortKey.value = cached.sortKey || ''
  sortDir.value = cached.sortDir || 'asc'
  currentPage.value = cached.currentPage || 1
  return records.value.length > 0
}

function persistToCache() {
  const cached = registryState.value
  if (!cached) return
  cached.records = records.value
  cached.loadedAt = Date.now()
  cached.category = category.value
  cached.search = search.value
  cached.subFilters = JSON.parse(JSON.stringify(subFilters.value || {}))
  cached.sortKey = sortKey.value
  cached.sortDir = sortDir.value
  cached.currentPage = currentPage.value
}

function mapLegacyCitizenToUnified(row) {
  const first = String(row?.firstName || '').trim()
  const last = String(row?.lastName || '').trim()
  const fullName = [first, last].filter(Boolean).join(' ').trim()
  const occupation = String(row?.workDetails || row?.occupation || '').trim() || 'Not Working'
  const annualIncome = String(row?.annualIncome || '').trim()
  const totalLand = String(row?.totalLand || '').trim()
  const irrigationType = String(row?.waterSource || '').toLowerCase().includes('rain') ? 'Rain-fed' : 'Irrigated'
  return {
    fullName,
    firstName: first,
    lastName: last,
    age: 0,
    gender: '',
    education: 'Not Available',
    familyId: 0,
    totalLand,
    irrigationType,
    waterSource: String(row?.waterSource || '').trim(),
    kharifCrop: '',
    rabiCrop: '',
    annualIncome,
    occupation,
    childrenCount: Number(row?.childrenCount || 0),
    sanitationStatus: 'Not Available',
    isFarmer: totalLand !== '' && totalLand !== '0',
    isStudent: false,
    isDivyang: false,
    isHousewife: occupation.toLowerCase().includes('housewife') || occupation.toLowerCase().includes('homemaker'),
    isSenior: false,
    schoolName: '',
    gradeStandard: '',
    educationLevel: 'Not Available',
    scholarship: 'No',
    disabilityType: '',
    disabilityPercent: '',
    caretakerName: 'Not Available',
    maritalStatus: '', // New field
    sourceOfIncome: 'None',
  }
}

async function loadRegistryData() {
  try {
    const data = await getUnifiedRegistry()
    records.value = Array.isArray(data) ? data : []
    persistToCache()
    return
  } catch (e) {
    const isAbort = e?.name === 'AbortError'
    console.warn('Unified registry endpoint slow/unavailable, trying fallback:', isAbort ? 'timeout' : (e?.message || e))
  }

  try {
    const legacy = await getCitizens()
    records.value = (Array.isArray(legacy) ? legacy : []).map(mapLegacyCitizenToUnified)
    persistToCache()
  } catch (e2) {
    console.error('Citizen registry load failed:', e2)
    records.value = []
  }
}

// ── Derived config ─────────────────────────────────────────────────────────
const hasPensionData = computed(() => records.value.some(r => {
  if (!r || typeof r !== 'object') return false
  return Object.prototype.hasOwnProperty.call(r, 'pensionStatus') ||
    Object.prototype.hasOwnProperty.call(r, 'govtPensionAmount')
}))

const activeCategoryConfig = computed(() => {
  const base = CATEGORY_CONFIG[category.value] || CATEGORY_CONFIG['']
  if (category.value !== 'disabled' || hasPensionData.value) return base

  return {
    ...base,
    columns: base.columns.filter(col => col.key !== 'pensionStatus' && col.key !== 'govtPensionAmount'),
    subFilters: (base.subFilters || []).filter(filter => filter.key !== 'pensionStatus'),
  }
})

// ── Data loading ───────────────────────────────────────────────────────────
async function ensureRegistryData() {
  if (hydrateFromCache()) {
    loading.value = false
    return
  }

  loading.value = true
  try {
    await loadRegistryData()
  } finally {
    loading.value = false
    persistToCache()
  }
}

onMounted(async () => {
  await ensureRegistryData()
})

onActivated(async () => {
  if (!records.value.length) {
    await ensureRegistryData()
  }
})

onBeforeUnmount(() => {
  persistToCache()
})

// ── Helpers ────────────────────────────────────────────────────────────────
function parseLand(v) {
  const n = parseFloat(String(v ?? '').replace(/,/g, ''))
  return Number.isFinite(n) ? n : 0
}

function parseIncome(v) {
  const cleaned = String(v ?? '').replace(/[^0-9]/g, '')
  return cleaned ? Number(cleaned) : Number.NaN
}

// classifyIncome: maps an income string to 'low' | 'medium' | 'high' | ''
function classifyIncome(inc) {
  const v = String(inc || '').toLowerCase();
  if (!v) return '';
  // Strictly define what is High
  if (v.includes('100001') || v.includes('250000') || v.includes('above') || v.includes('high')) return 'High';
  // Strictly define what is Mid
  if (v.includes('50001') || v.includes('100000') || v.includes('to 50000') || v.includes('mid')) return 'Mid';
  // Everything else is Low
  return 'Low';
}

function classifyDisabilitySeverity(v) {
  const n = parseFloat(String(v ?? '').replace(/[^0-9.]/g, ''))
  if (!Number.isFinite(n)) return 'unknown'
  if (n < 40) return 'low'
  if (n <= 70) return 'moderate'
  return 'high'
}

function hasDivyangCertificate(r) {
  return (r.isDivyang === true) ||
    (String(r.disabilityType || '').trim() !== '' && String(r.disabilityType || '').trim().toLowerCase() !== 'not recorded') ||
    (String(r.disabilityPercent || '').trim() !== '' && String(r.disabilityPercent || '').trim() !== '0')
}

function classifyAgeGroup(v) {
  const n = Number(v)
  if (!Number.isFinite(n) || n <= 0) return ''
  if (n <= 35) return 'young'
  if (n <= 55) return 'middle'
  return 'senior'
}

function classifyChildrenGroup(v) {
  const n = Number(v)
  if (!Number.isFinite(n) || n <= 0) return 'none'
  if (n <= 2) return 'oneToTwo'
  return 'threePlus'
}

function getGenderDotClass(gender) {
  const g = (gender || '').toLowerCase().trim();
  if (g === 'male' || g === 'm') return 'male';
  if (g === 'female' || g === 'f') return 'female';
  return 'other';
}

function getOccupationBucket(rawOccupation) {
  const v = String(rawOccupation || '').toLowerCase().trim();
  if (!v || v === 'not working' || v === 'none' || v === 'n/a') return 'Not Working';
  if (v.includes('farm') || v.includes('agri') || v.includes('sheti')) return 'Farmer';
  if (v.includes('student') || v.includes('study') || v.includes('education')) return 'Student';
  if (v.includes('housewife') || v.includes('homemaker')) return 'Housewife';
  if (v.includes('wage') || v.includes('labor') || v.includes('labour') || v.includes('majoor') || v.includes('majuri') || v.includes('worker')) return 'Wage Work';
  if (v.includes('salary') || v.includes('salaried') || v.includes('service') || v.includes('teacher') || v.includes('asha') || v.includes('police') || v.includes('job') || v.includes('anganwadi')) return 'Salaried Job';
  if (v.includes('business') || v.includes('shop') || v.includes('shg') || v.includes('tailor') || v.includes('vendor') || v.includes('vyapar')) return 'Business/SHG';
  if (v.includes('unemployed')) return 'Unemployed';
  return 'Other'; // Catch-all for anything else
}

function normalizeOccupationText(record) {
  return String(record?.occupation || '').toLowerCase().trim()
}

function matchesOccupationType(record, value) {
  const occupation = normalizeOccupationText(record)
  const selected = String(value || '').toLowerCase()

  if (selected === 'farmer') {
    return Boolean(record?.isFarmer)
  }

  if (selected === 'student') {
    return Boolean(record?.isStudent || occupation.includes('student') || occupation.includes('studying'))
  }

  if (selected === 'housewife') {
    return Boolean(record?.isHousewife || occupation.includes('housewife') || occupation.includes('homemaker'))
  }

  if (selected === 'salaried job') {
    return occupation.includes('salaried') || occupation.includes('salary') || occupation.includes('service')
  }

  if (selected === 'wage work') {
    return occupation.includes('wage') || occupation.includes('labour') || occupation.includes('labor')
  }

  if (selected === 'unemployed') {
    return occupation.includes('unemployed')
  }

  if (selected === 'not working') {
    return occupation === 'not working' || occupation === ''
  }

  return occupation === selected
}

function toggleSort(key) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
}

function getSelectedValues(key) {
  const values = subFilters.value[key]
  return Array.isArray(values) ? values : []
}

function isSubFilterActive(key, value) {
  return getSelectedValues(key).includes(value)
}

function toggleSubFilter(key, value) {
  const current = getSelectedValues(key)
  const next = current.includes(value)
    ? current.filter((v) => v !== value)
    : [...current, value]

  subFilters.value = {
    ...subFilters.value,
    [key]: next,
  }
  persistToCache()
}

function onCategoryChange() {
  // Clear all sub-filters when switching categories to avoid data conflicts
  subFilters.value = {}
  sortKey.value = ''
  sortDir.value = 'asc'
  currentPage.value = 1
  persistToCache()
}

function resetFilters() {
  search.value = ''
  subFilters.value = {}
  sortKey.value = ''
  sortDir.value = 'asc'
  currentPage.value = 1
  persistToCache()
}

// ── Category-pre-filtered list ─────────────────────────────────────────────
const categoryRecords = computed(() =>
  records.value.filter(activeCategoryConfig.value.rowFilter)
)

function isBooleanFilterKey(key) {
  return key === 'isDivyang' || key === 'isFarmer' || key === 'isStudent' || key === 'isHousewife' || key === 'isSenior'
}

function formatDynamicLabel(value) {
  const str = String(value || '').trim()
  if (!str) return ''
  if (str === 'true') return 'Yes'
  if (str === 'false') return 'No'

  return str
    .replace(/[_-]+/g, ' ')
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/\b\w/g, ch => ch.toUpperCase())
}

function dynamicValueForFilter(key, record) {
  if (key === 'incomeRange') return classifyIncome(record.annualIncome)
  if (key === 'ageGroup') return classifyAgeGroup(record.age)
  if (key === 'childrenGroup') return classifyChildrenGroup(record.childrenCount)
  if (key === 'disabilitySeverity') return classifyDisabilitySeverity(record.disabilityPercent)
  if (key === 'divyangCertificate') return hasDivyangCertificate(record) ? 'yes' : 'no'

  if (key === 'landSize') {
    const ac = parseLand(record.totalLand)
    if (ac > 2) return 'large'
    if (ac >= 1 && ac <= 2) return 'medium'
    if (ac > 0 && ac < 1) return 'marginal'
    return ''
  }

  if (key === 'occupationType') {
    return getOccupationBucket(record.occupation)
  }

  if (key === 'cropType') {
    const hasKharif = record.kharifCrop && record.kharifCrop !== 'N/A'
    const hasRabi = record.rabiCrop && record.rabiCrop !== 'N/A'
    if (hasKharif && hasRabi) return 'both'
    if (hasKharif) return 'kharif'
    if (hasRabi) return 'rabi'
    return ''
  }

  if (isBooleanFilterKey(key)) {
    return record[key] ? 'true' : 'false'
  }

  return String(record[key] ?? '').trim()
}

const dynamicSubFilters = computed(() => {
  const base = activeCategoryConfig.value.subFilters || []
  return base
    .map((sf) => {
      const unique = new Set()
      for (const record of categoryRecords.value) {
        const value = sf.key === 'occupationType'
          ? getOccupationBucket(record.occupation)
          : dynamicValueForFilter(sf.key, record)
        if (value !== '' && value !== 'N/A' && value != null) {
          unique.add(String(value))
        }
      }

      const sortedValues = Array.from(unique).sort((a, b) => a.localeCompare(b, undefined, { sensitivity: 'base' }))
      const options = sortedValues.map((value) => ({
        label: sf.key === 'occupationType' ? (value === 'Housewife' ? 'Homemaker' : value) : formatDynamicLabel(value),
        value,
      }))

      return { ...sf, options }
    })
    .filter((sf) => sf.options.length > 0)
})

// ── Full filtered + sorted list ────────────────────────────────────────────
const filteredRecords = computed(() => {
  let list = [...categoryRecords.value]

  // Search always by name (works across all categories)
  if (search.value) {
    const s = search.value.toLowerCase()
    list = list.filter(r => (r.fullName || '').toLowerCase().includes(s))
  }

  // Apply active sub-filters
  const sf = subFilters.value
  const selected = (key) => Array.isArray(sf[key]) ? sf[key] : []
  const hasAny = (key) => selected(key).length > 0

  if (hasAny('occupationType')) {
    const selectedBuckets = new Set(selected('occupationType'));
    list = list.filter(r => selectedBuckets.has(getOccupationBucket(r.occupation)));
  }

  for (const sfKey of Object.keys(sf)) {
    if (sfKey === 'occupationType') continue
    if (!hasAny(sfKey)) continue
    const selectedSet = new Set(selected(sfKey).map(v => String(v)))
    list = list.filter(r => selectedSet.has(String(dynamicValueForFilter(sfKey, r))))
  }

  // Sort
  if (sortKey.value) {
    list.sort((a, b) => {
      const dir = sortDir.value === 'asc' ? 1 : -1
      const key = sortKey.value

      if (key === 'totalLand')      return dir * (parseLand(a.totalLand)   - parseLand(b.totalLand))
      if (key === 'childrenCount')  return dir * ((a.childrenCount || 0)   - (b.childrenCount || 0))
      if (key === 'age')            return dir * ((a.age || 0)             - (b.age || 0))
      if (key === 'annualIncome') {
        const av = parseIncome(a.annualIncome), bv = parseIncome(b.annualIncome)
        if (Number.isFinite(av) && Number.isFinite(bv)) return dir * (av - bv)
      }

      const av = String(a[key] || '').toLowerCase()
      const bv = String(b[key] || '').toLowerCase()
      return dir * av.localeCompare(bv)
    })
  }

  return list
})

const totalPages = computed(() =>
  Math.max(1, Math.ceil(filteredRecords.value.length / pageSize.value))
)
const paginatedRecords = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredRecords.value.slice(start, start + pageSize.value)
})

watch([search, subFilters, category], () => {
  currentPage.value = 1
  persistToCache()
}, { deep: true })

watch([records, sortKey, sortDir, currentPage], () => {
  persistToCache()
}, { deep: true })

// ── Cell renderer ──────────────────────────────────────────────────────────
// Returns an HTML string so the template can use v-html for badge rendering.
// All name values are already capitalised by the backend.

function renderCell(r, col) {
  const v = r[col.key]

  switch (col.key) {
    case 'fullName': {
      const gCls = getGenderDotClass(r.gender);
      const nameCls = gCls === 'male' ? 'name-male' : (gCls === 'female' ? 'name-female' : 'name-other');
      return `<div class="name-cell"><span class="gender-dot ${gCls}" title="${r.gender || 'Other'}"></span> <span class="name-text ${nameCls}">${esc(r.fullName)}</span></div>`;
    }

    case 'totalLand': {
      const ac = parseLand(v)
      if (ac === 0) return `<span class="badge badge-muted">No Land</span>`
      return `<span class="badge badge-green">${esc(v)} ac</span>`
    }

    case 'crops': {
      const kharif = r.kharifCrop && r.kharifCrop !== 'N/A' ? r.kharifCrop : null
      const rabi   = r.rabiCrop   && r.rabiCrop   !== 'N/A' ? r.rabiCrop   : null
      if (!kharif && !rabi) return `<span class="text-dim-sm">—</span>`
      const parts = []
      if (kharif) parts.push(`<span class="badge badge-kharif" title="Kharif">☀ ${esc(kharif)}</span>`)
      if (rabi)   parts.push(`<span class="badge badge-rabi"   title="Rabi">❄ ${esc(rabi)}</span>`)
      return parts.join(' ')
    }

    case 'irrigationType': {
      const cls = v === 'Irrigated' ? 'badge-irrigated' : v === 'Rain-fed' ? 'badge-rainfed' : 'badge-muted'
      return `<span class="badge ${cls}">${esc(v || '—')}</span>`
    }

    case 'waterSource':
      return `<span class="text-dim-sm">${esc(v || '—')}</span>`

    case 'educationLevel': {
      const map = { Graduate: 'badge-blue', Undergraduate: 'badge-teal', 'Anganwadi/Primary': 'badge-orange', 'Not Available': 'badge-muted' }
      const cls = map[v] || 'badge-muted'
      return `<span class="badge ${cls}">${esc(v || 'Not Available')}</span>`
    }

    case 'scholarship': {
      const cls = v === 'Yes' ? 'badge-green' : 'badge-muted'
      return `<span class="badge ${cls}">${esc(v || '—')}</span>`
    }

    case 'pensionStatus': {
      const cls = v === 'Eligible' ? 'badge-green' : 'badge-muted'
      return `<span class="badge ${cls}">${esc(v || '—')}</span>`
    }

    case 'disabilityType':
      return v ? `<span class="text-body-sm">${esc(v)}</span>` : `<span class="text-dim-sm">Not Recorded</span>`

    case 'disabilityPercent':
      return v && v !== '0' ? `<span class="badge badge-orange">${esc(v)}%</span>` : `<span class="text-dim-sm">—</span>`

    case 'divyangCertificate': {
      const hasDisabilityMarker = hasDivyangCertificate(r)
      const cls = hasDisabilityMarker ? 'badge-green' : 'badge-muted'
      const label = hasDisabilityMarker ? 'Available' : 'Not Available'
      return `<span class="badge ${cls}">${label}</span>`
    }

    case 'govtPensionAmount': {
      if (!v || v === 'N/A' || v === '0') return `<span class="text-dim-sm">N/A</span>`
      return `<span class="badge badge-teal">₹ ${esc(v)}</span>`
    }

    case 'sourceOfIncome': {
      const colorMap = {
        'Tailoring':              'badge-blue',
        'Poultry':                'badge-green',
        'Small Business/SHG':    'badge-orange',
        'Remittance from family': 'badge-teal',
        'None':                   'badge-muted',
      }
      const cls = colorMap[v] || 'badge-muted'
      return `<span class="badge ${cls}">${esc(v || 'None')}</span>`
    }

    case 'caretakerName':
      return `<span class="text-dim-sm">${esc(v || 'Not Available')}</span>`

    case 'schoolName':
      return v ? `<span class="text-body-sm">${esc(v)}</span>` : `<span class="text-dim-sm">Not Recorded</span>`

       case 'occupation':
      if (String(v || '').trim() === 'Housewife') return esc('Homemaker')
      return esc(v || 'Not Working')

    case 'age':
      return v && v > 0 ? String(v) : '—'

    case 'annualIncome': {
      const inc = r.annualIncome;
      if (!inc) return '<span class="text-muted">Not Recorded</span>';
      
      const b = classifyIncome(inc);
      let color = '#ff4d4f'; // Default to Low
      if (b === 'High') color = '#20c997';
      else if (b === 'Mid') color = '#ffc107';
      
      return `<span style="color: ${color}; font-weight: 500;">${esc(inc)}</span>`;
    }
    
    case 'maritalStatus': {
      return v ? `<span class="badge badge-blue">${esc(v)}</span>` : `<span class="text-dim-sm">—</span>`
    }

    case 'childrenCount':
      return String(r.childrenCount ?? 0)

    default:
      return esc(String(v ?? '—'))
  }
}

// Minimal HTML escaping for v-html cells
function esc(s) {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}
</script>

<style scoped>
/* ── Layout ──────────────────────────────────────────────────────────────── */
.registry-wrap    { padding: 2rem 2.5rem; width: 100%; }
.registry-embedded { padding: 0.75rem 1rem; }

/* ── Header ──────────────────────────────────────────────────────────────── */
.page-header {
  margin-bottom: 0;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px 12px 0 0;
  overflow: visible;   /* allow dropdown to escape */
}

/* Row 1 — Title + CATEGORY dropdown */
.header-row {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 1rem 1.5rem;
  flex-wrap: wrap;
}
.header-row-1 {
  border-bottom: 1px solid var(--border);
  background: var(--bg-surface);
  gap: 1.25rem;
}
/* Row 2 — search + sub-filters */
.header-row-2 {
  gap: 0.55rem;
  padding: 0.65rem 1.5rem;
}

.title-block { flex-shrink: 0; }
.page-title {
  font-family: var(--font-display);
  font-size: 1.6rem;
  color: var(--text-primary);
  font-weight: 400;
  line-height: 1.2;
}
.page-subtitle { color: var(--text-dim); font-size: 0.75rem; margin-top: 0.2rem; }

/* ── CATEGORY custom dropdown (mirrors 3D Twin "View By") ─────────────────── */
.reg-filter-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-left: auto;  /* push to right side of Row 1 */
}
.reg-filter-label {
  font-size: 0.63rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: var(--text-dim);
  user-select: none;
}

.reg-custom-select {
  position: relative;
  min-width: 190px;
}

.reg-cs-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.4rem;
  width: 100%;
  background: #ffffff;
  border: 1.5px solid #9ca3af;
  border-radius: 7px;
  color: #111827;
  font-size: 0.82rem;
  font-weight: 500;
  padding: 0.38rem 0.65rem;
  cursor: pointer;
  outline: none;
  box-shadow: 0 1px 3px rgba(0,0,0,0.07);
  transition: border-color 0.15s, box-shadow 0.15s;
  text-align: left;
  white-space: nowrap;
}
.reg-cs-trigger:hover          { border-color: #6b7280; }
.reg-custom-select.open .reg-cs-trigger {
  border-color: #16a34a;
  box-shadow: 0 0 0 3px rgba(22,163,74,0.15);
}

.reg-cs-trigger--sm { padding: 0.28rem 0.55rem; font-size: 0.76rem; }

.reg-cs-value { flex: 1; overflow: hidden; text-overflow: ellipsis; }
.reg-cs-arrow {
  font-size: 0.6rem;
  color: #6b7280;
  flex-shrink: 0;
  transition: transform 0.15s;
}
.reg-custom-select.open .reg-cs-arrow { transform: rotate(180deg); }

/* Dropdown panel — anchored right so it doesn't overlap Row 2 content */
.reg-cs-dropdown {
  position: absolute;
  top: calc(100% + 5px);
  right: 0;
  left: auto;
  min-width: 100%;
  background: #ffffff;
  border: 1.5px solid #d1d5db;
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.13), 0 3px 8px rgba(0,0,0,0.08);
  z-index: 9999;   /* float over row 2 and table */
  overflow: hidden;
}

.reg-cs-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.46rem 0.8rem;
  font-size: 0.8rem;
  color: #111827;
  cursor: pointer;
  transition: background 0.1s, color 0.1s;
  white-space: nowrap;
}
.reg-cs-option:hover    { background: #f0fdf4; color: #15803d; }
.reg-cs-option.selected { background: #dcfce7; color: #15803d; font-weight: 600; }
.reg-cs-opt-icon        { font-size: 0.9rem; line-height: 1; width: 1.1rem; flex-shrink: 0; }

/* Compact variant for embedded mode */
.reg-custom-select--sm { min-width: 150px; }

/* ── Embedded toolbar ────────────────────────────────────────────────────── */
.embedded-toolbar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 0.75rem;
}

/* ── Search box ──────────────────────────────────────────────────────────── */
.search-box {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  background: var(--bg-card);
  border: 1.5px solid var(--border-light);
  border-radius: 8px;
  padding: 0 0.75rem;
  height: 36px;
  flex: 0 1 260px;
  transition: border-color 0.18s, box-shadow 0.18s;
}
.search-box-sm { flex: 0 1 200px; height: 32px; }
.search-box:focus-within { border-color: var(--amber); box-shadow: 0 0 0 3px var(--amber-dim); }
.search-icon { width: 15px; height: 15px; color: var(--text-dim); flex-shrink: 0; }
.search-input {
  background: none; border: none; outline: none;
  color: var(--text-body); font-family: var(--font-body); font-size: 0.84rem; width: 100%;
}
.search-input::placeholder { color: var(--text-dim); }

/* ── Sub-filter bar ──────────────────────────────────────────────────────── */
.subfilter-bar {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  flex-wrap: wrap;
}
.subfilter-label {
  font-size: 0.68rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--text-dim);
  font-weight: 600;
  white-space: nowrap;
}
.filter-divider { width: 1px; height: 20px; background: var(--border); margin: 0 0.1rem; flex-shrink: 0; }

.chip {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 20px;
  padding: 0.3rem 0.75rem;
  font-size: 0.73rem;
  color: var(--text-muted);
  cursor: pointer;
  font-family: var(--font-body);
  transition: all 0.18s;
  white-space: nowrap;
}
.chip:hover  { border-color: var(--border-light); color: var(--text-body); }
.chip.active { background: var(--amber-dim); border-color: var(--amber); color: var(--amber); }

.reset-btn {
  height: 36px;
  padding: 0 0.9rem;
  border-radius: 8px;
  border: 1.5px solid var(--border-light);
  background: var(--bg-surface);
  color: var(--text-muted);
  font-family: var(--font-body);
  font-size: 0.76rem;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}
.reset-btn:hover { border-color: var(--amber); color: var(--amber); background: var(--amber-dim); }

/* ── Table container ─────────────────────────────────────────────────────── */
.table-container {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: none;
  border-radius: 0 0 12px 12px;
  overflow: hidden;
}

.table-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.75rem 1.25rem;
  border-bottom: 1px solid var(--border);
  font-size: 0.75rem;
  color: var(--text-dim);
}
.table-info-main {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}
.table-info strong { color: var(--text-body); }
.table-info em { font-style: normal; }

.active-cat-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.2rem 0.65rem;
  border-radius: 20px;
  border: 1px solid;
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.table-wrap { overflow-x: auto; }

/* ── Table ───────────────────────────────────────────────────────────────── */
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
.th-name  { min-width: 200px; }
.sortable { cursor: pointer; }
.sortable:hover { color: var(--text-body); }
.sort-arrow { color: var(--amber); margin-left: 0.25rem; }

.table-row { transition: background 0.15s; }
.table-row:hover { background: var(--bg-surface); }
.table-row td {
  padding: 0.6rem 1.25rem;
  border-bottom: 1px solid var(--border);
  font-size: 0.82rem;
  vertical-align: middle;
}

.td-index { color: var(--text-dim); font-variant-numeric: tabular-nums; font-size: 0.75rem; }
.td-name  { color: var(--text-body); }
.td-num   { font-variant-numeric: tabular-nums; color: var(--text-body); font-size: 0.8rem; }
.td-empty { text-align: center; padding: 2rem; color: var(--text-dim); font-size: 0.82rem; }

/* ── Cell micro-components (rendered via v-html) ─────────────────────────── */
/* These are NOT scoped — they live inside v-html, so we need :deep or global.  */
/* We use unique class names prefixed with `badge` / `text-*` to avoid leaks. */
:deep(.name-text) {
  color: var(--text-body);
  transition: color 0.15s ease-out;
}
:deep(.name-male) { color: #1e40af; }
:deep(.name-female) { color: #be185d; }
:deep(.name-other) { color: var(--text-body); }

/* ── Gender Legend ───────────────────────────────────────────────────────── */
.gender-legend { display: flex; gap: 1.2rem; align-items: center; font-size: 0.75rem; color: var(--text-dim); }
.legend-item { display: flex; align-items: center; gap: 0.4rem; }
.gender-dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; flex-shrink: 0; }
.gender-dot.male { background-color: #1e40af; }
.gender-dot.female { background-color: #be185d; }
.gender-dot.other { background-color: var(--text-dim, #6c757d); }
:deep(.name-cell) { display: flex; align-items: center; gap: 0.6rem; }

:deep(.gender-tag),
:deep(.badge) {
  display: inline-block;
  padding: 0.15rem 0.55rem;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 500;
  white-space: nowrap;
  line-height: 1.4;
}

:deep(.badge-green)    { background: #dcfce7; color: #15803d; }
:deep(.badge-blue)     { background: #dbeafe; color: #1e40af; }
:deep(.badge-teal)     { background: #ccfbf1; color: #0f766e; }
:deep(.badge-orange)   { background: #ffedd5; color: #c2410c; }
:deep(.badge-irrigated){ background: #dcfce7; color: #15803d; }
:deep(.badge-rainfed)  { background: #fee2e2; color: #b91c1c; }
:deep(.badge-kharif)   { background: #fef9c3; color: #854d0e; }
:deep(.badge-rabi)     { background: #e0f2fe; color: #075985; }
:deep(.badge-muted)    { background: var(--bg-surface); color: var(--text-dim); }

:deep(.text-body-sm)   { font-size: 0.82rem; color: var(--text-body); }
:deep(.text-dim-sm)    { font-size: 0.82rem; color: var(--text-dim); }

/* ── Pagination ──────────────────────────────────────────────────────────── */
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

/* ── Loading ─────────────────────────────────────────────────────────────── */
.loading-state {
  display: flex; align-items: center; justify-content: center;
  gap: 0.75rem; padding: 3rem 0; color: var(--text-dim);
}
.spinner {
  width: 18px; height: 18px;
  border: 2px solid var(--border-light);
  border-top-color: var(--amber);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
