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

        <template v-if="activeSubFilters.length">
          <div class="subfilter-bar">
            <template v-for="sf in activeSubFilters" :key="sf.key">
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
      <template v-for="sf in activeSubFilters" :key="sf.key">
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
        <span>
          Showing <strong>{{ filteredRecords.length }}</strong> of {{ categoryRecords.length }}
          <em>{{ activeCategoryConfig.label.toLowerCase() }}</em> citizens
        </span>
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
  { value: 'housewife',label: 'Housewives',       fullLabel: 'Housewives',      icon: '🏠' },
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
      { key: 'age',          label: 'Age',            tdClass: 'td-num' },
      { key: 'gender',       label: 'Gender' },
      { key: 'occupation',   label: 'Occupation' },
      { key: 'annualIncome', label: 'Annual Income',  tdClass: 'td-num' },
    ],
    subFilters: [
      {
        key: 'gender', label: 'Gender',
        options: [
          { label: 'Male',   value: 'Male'   },
          { label: 'Female', value: 'Female' },
        ],
      },
      {
        key: 'incomeRange', label: 'Income Range',
        options: [
          { label: 'Low  (<₹21k)',      value: 'low'    },
          { label: 'Mid  (₹21k–50k)',   value: 'medium' },
          { label: 'High (>₹50k)',      value: 'high'   },
        ],
      },
    ],
  },

  farmer: {
    label: 'Farmer',
    subtitle: 'Citizen-level records in land-owning households (count differs from 3D household total)',
    icon: '🌾',
    color: '#16a34a',
    rowFilter: r => r.isFarmer,
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
      {
        key: 'landSize', label: 'Land Size',
        options: [
          { label: '>2 Acres',   value: 'large'    },
          { label: '1–2 Acres',  value: 'medium'   },
          { label: '<1 Acre',    value: 'marginal' },
        ],
      },
      {
        key: 'irrigationType', label: 'Irrigation',
        options: [
          { label: 'Irrigated', value: 'Irrigated' },
          { label: 'Rain-fed',  value: 'Rain-fed'  },
        ],
      },
      {
        key: 'cropType', label: 'Crop',
        options: [
          { label: 'Kharif', value: 'kharif' },
          { label: 'Rabi',   value: 'rabi'   },
          { label: 'Both',   value: 'both'   },
        ],
      },
      {
        key: 'incomeRange', label: 'Income Range',
        options: [
          { label: 'Low  (<₹21k)',    value: 'low'    },
          { label: 'Mid  (₹21k–50k)', value: 'medium' },
          { label: 'High (>₹50k)',    value: 'high'   },
        ],
      },
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
      { key: 'age',            label: 'Age',               tdClass: 'td-num' },
      { key: 'gender',         label: 'Gender' },
      { key: 'educationLevel', label: 'Education Level' },
      { key: 'schoolName',     label: 'School / College' },
      { key: 'scholarship',    label: 'Scholarship' },
      { key: 'annualIncome',   label: 'Annual Income',     tdClass: 'td-num' },
    ],
    subFilters: [
      {
        key: 'educationLevel', label: 'Level',
        options: [
          { label: 'Graduate',          value: 'Graduate'          },
          { label: 'Undergraduate',     value: 'Undergraduate'     },
          { label: 'Anganwadi/Primary', value: 'Anganwadi/Primary' },
        ],
      },
      {
        key: 'gender', label: 'Gender',
        options: [
          { label: 'Male',   value: 'Male'   },
          { label: 'Female', value: 'Female' },
        ],
      },
      {
        key: 'scholarship', label: 'Scholarship',
        options: [
          { label: 'Yes', value: 'Yes' },
          { label: 'No',  value: 'No'  },
        ],
      },
      {
        key: 'incomeRange', label: 'Income Range',
        options: [
          { label: 'Low  (<₹21k)',    value: 'low'    },
          { label: 'Mid  (₹21k–50k)', value: 'medium' },
          { label: 'High (>₹50k)',    value: 'high'   },
        ],
      },
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
      { key: 'pensionStatus',     label: 'Pension Status' },
      { key: 'govtPensionAmount', label: 'Govt. Pension Amt', tdClass: 'td-num' },
      { key: 'caretakerName',     label: 'Caretaker' },
      { key: 'divyangCertificate', label: 'Divyang Certificate' },
    ],
    subFilters: [
      {
        key: 'gender', label: 'Gender',
        options: [
          { label: 'Male',   value: 'Male'   },
          { label: 'Female', value: 'Female' },
        ],
      },
      {
        key: 'pensionStatus', label: 'Pension',
        options: [
          { label: 'Eligible',     value: 'Eligible'     },
          { label: 'Not Eligible', value: 'Not Eligible' },
        ],
      },
      {
        key: 'disabilitySeverity', label: 'Disability %',
        options: [
          { label: 'Low (<40%)',      value: 'low'      },
          { label: 'Moderate (40-70%)', value: 'moderate' },
          { label: 'High (>70%)',     value: 'high'     },
          { label: 'Not Recorded',    value: 'unknown'  },
        ],
      },
      {
        key: 'divyangCertificate', label: 'Certificate',
        options: [
          { label: 'Available',     value: 'yes' },
          { label: 'Not Available', value: 'no'  },
        ],
      },
      {
        key: 'incomeRange', label: 'Income Range',
        options: [
          { label: 'Low  (<₹21k)',    value: 'low'    },
          { label: 'Mid  (₹21k–50k)', value: 'medium' },
          { label: 'High (>₹50k)',    value: 'high'   },
        ],
      },
    ],
  },

  housewife: {
    label: 'Housewife',
    subtitle: 'Citizens with occupation recorded as housewife / homemaker',
    icon: '🏠',
    color: '#db2777',
    rowFilter: r => r.isHousewife,
    columns: [
      { key: 'fullName',       label: 'Full Name',       minWidth: true, tdClass: 'td-name' },
      { key: 'age',            label: 'Age',              tdClass: 'td-num' },
      { key: 'gender',         label: 'Gender' },
      { key: 'annualIncome',   label: 'Annual Income',    tdClass: 'td-num' },
      { key: 'childrenCount',  label: 'Children',         tdClass: 'td-num' },
    ],
    subFilters: [
      {
        key: 'gender', label: 'Gender',
        options: [
          { label: 'Male',   value: 'Male'   },
          { label: 'Female', value: 'Female' },
        ],
      },
      {
        key: 'ageGroup', label: 'Age Group',
        options: [
          { label: '18-35', value: 'young'  },
          { label: '36-55', value: 'middle' },
          { label: '56+',   value: 'senior' },
        ],
      },
      {
        key: 'childrenGroup', label: 'Children',
        options: [
          { label: 'No Children', value: 'none'     },
          { label: '1-2',         value: 'oneToTwo' },
          { label: '3+',          value: 'threePlus' },
        ],
      },
      {
        key: 'sourceOfIncome', label: 'Income Source',
        options: [
          { label: 'Tailoring',              value: 'Tailoring' },
          { label: 'Poultry',                value: 'Poultry' },
          { label: 'Small Business/SHG',     value: 'Small Business/SHG' },
          { label: 'Remittance from family', value: 'Remittance from family' },
          { label: 'None',                   value: 'None' },
        ],
      },
      {
        key: 'incomeRange', label: 'Income Range',
        options: [
          { label: 'Low  (<₹21k)',    value: 'low'    },
          { label: 'Mid  (₹21k–50k)', value: 'medium' },
          { label: 'High (>₹50k)',    value: 'high'   },
        ],
      },
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
      { key: 'age',           label: 'Age',             tdClass: 'td-num' },
      { key: 'gender',        label: 'Gender' },
      { key: 'occupation',    label: 'Occupation' },
      { key: 'pensionStatus', label: 'Pension Status' },
      { key: 'annualIncome',  label: 'Annual Income',   tdClass: 'td-num' },
    ],
    subFilters: [
      {
        key: 'incomeRange', label: 'Income Range',
        options: [
          { label: 'Low  (<₹21k)',    value: 'low'    },
          { label: 'Mid  (₹21k–50k)', value: 'medium' },
          { label: 'High (>₹50k)',    value: 'high'   },
        ],
      },
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
    pensionStatus: 'Not Eligible',
    caretakerName: 'Not Available',
    govtPensionAmount: 'N/A',
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
const activeCategoryConfig = computed(() => CATEGORY_CONFIG[category.value] || CATEGORY_CONFIG[''])
const activeSubFilters     = computed(() => activeCategoryConfig.value.subFilters || [])

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
function classifyIncome(v) {
  const raw = String(v ?? '').toLowerCase().trim()
  if (!raw) return ''

  // Handle text labels stored in the DB (e.g. "Less than 21,000", "21001 to 50000")
  if (raw.includes('less than') && raw.includes('21')) return 'low'
  if ((raw.includes('21') || raw.includes('21001')) && raw.includes('50')) return 'medium'
  if (raw.includes('50,001') || raw.includes('50001') || raw.includes('above') || raw.includes('50,000+')) return 'high'

  const n = parseIncome(v)
  if (!Number.isFinite(n) || n === 0) return ''
  if (n <= 21000)  return 'low'
  if (n <= 50000)  return 'medium'
  return 'high'
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

  // Gender sub-filter
  if (hasAny('gender')) {
    const genderSet = new Set(selected('gender').map(v => String(v).toLowerCase()))
    list = list.filter(r => genderSet.has(String(r.gender || '').toLowerCase()))
  }

  // Land size sub-filter (farmer category)
  if (hasAny('landSize')) {
    const sizeSet = new Set(selected('landSize'))
    list = list.filter(r => {
      const ac = parseLand(r.totalLand)
      const matchesLarge = sizeSet.has('large') && ac > 2
      const matchesMedium = sizeSet.has('medium') && ac >= 1 && ac <= 2
      const matchesMarginal = sizeSet.has('marginal') && ac > 0 && ac < 1
      if (matchesLarge || matchesMedium || matchesMarginal) return true
      return false
    })
  }

  // Irrigation type sub-filter (farmer category)
  if (hasAny('irrigationType')) {
    const irrigSet = new Set(selected('irrigationType'))
    list = list.filter(r => irrigSet.has(r.irrigationType))
  }

  // Crop type sub-filter (farmer category)
  if (hasAny('cropType')) {
    const cropSet = new Set(selected('cropType'))
    list = list.filter(r => {
      const hasKharif = r.kharifCrop && r.kharifCrop !== 'N/A'
      const hasRabi   = r.rabiCrop   && r.rabiCrop   !== 'N/A'
      const matchesKharif = cropSet.has('kharif') && hasKharif
      const matchesRabi = cropSet.has('rabi') && hasRabi
      const matchesBoth = cropSet.has('both') && hasKharif && hasRabi
      if (matchesKharif || matchesRabi || matchesBoth) return true
      return false
    })
  }

  // Education level sub-filter (student category)
  if (hasAny('educationLevel')) {
    const eduSet = new Set(selected('educationLevel'))
    list = list.filter(r => eduSet.has(r.educationLevel))
  }

  // Scholarship sub-filter (student category)
  if (hasAny('scholarship')) {
    const scholSet = new Set(selected('scholarship'))
    list = list.filter(r => scholSet.has(r.scholarship))
  }

  // Pension status sub-filter (divyang/senior categories)
  if (hasAny('pensionStatus')) {
    const pensionSet = new Set(selected('pensionStatus').map(v => String(v).toLowerCase()))
    list = list.filter(r => pensionSet.has(String(r.pensionStatus || '').toLowerCase()))
  }

  // Disability severity sub-filter (divyang category)
  if (hasAny('disabilitySeverity')) {
    const severitySet = new Set(selected('disabilitySeverity'))
    list = list.filter(r => severitySet.has(classifyDisabilitySeverity(r.disabilityPercent)))
  }

  // Divyang certificate sub-filter (divyang category)
  if (hasAny('divyangCertificate')) {
    const certSet = new Set(selected('divyangCertificate'))
    list = list.filter(r => {
      const hasCert = hasDivyangCertificate(r)
      return (certSet.has('yes') && hasCert) || (certSet.has('no') && !hasCert)
    })
  }

  // Age-group sub-filter (housewife category)
  if (hasAny('ageGroup')) {
    const ageSet = new Set(selected('ageGroup'))
    list = list.filter(r => ageSet.has(classifyAgeGroup(r.age)))
  }

  // Children-group sub-filter (housewife category)
  if (hasAny('childrenGroup')) {
    const childrenSet = new Set(selected('childrenGroup'))
    list = list.filter(r => childrenSet.has(classifyChildrenGroup(r.childrenCount)))
  }

  // Source-of-income sub-filter (housewife category)
  if (hasAny('sourceOfIncome')) {
    const sourceSet = new Set(selected('sourceOfIncome'))
    list = list.filter(r => sourceSet.has(r.sourceOfIncome))
  }

  // Income range — global, works across ALL categories
  if (hasAny('incomeRange')) {
    const incomeSet = new Set(selected('incomeRange'))
    list = list.filter(r => incomeSet.has(classifyIncome(r.annualIncome)))
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
    case 'fullName':
      return `<span class="name-text">${esc(r.fullName)}</span>`

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

    case 'gender': {
      const cls = genderClass(v)
      return v ? `<span class="badge ${cls}">${esc(v)}</span>` : '—'
    }

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
      return esc(v || 'Not Working')

    case 'age':
      return v && v > 0 ? String(v) : '—'

    case 'annualIncome': {
      if (!v || v === '0') return `<span class="text-dim-sm">—</span>`
      const rangeMap = { low: 'badge-rainfed', medium: 'badge-orange', high: 'badge-green' }
      const rangeCls = rangeMap[classifyIncome(v)] || ''
      const rangeLabel = { low: 'Low', medium: 'Mid', high: 'High' }[classifyIncome(v)] || ''
      const rangePill = rangeCls ? `<span class="badge ${rangeCls}" style="margin-left:0.35rem;font-size:0.62rem">${rangeLabel}</span>` : ''
      return `<span>${esc(v)}</span>${rangePill}`
    }

    case 'childrenCount':
      return String(r.childrenCount ?? 0)

    default:
      return esc(String(v ?? '—'))
  }
}

function genderClass(g) {
  const v = String(g || '').toLowerCase()
  if (v === 'male')   return 'gender-male'
  if (v === 'female') return 'gender-female'
  return 'gender-other'
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
  padding: 0.75rem 1.25rem;
  border-bottom: 1px solid var(--border);
  font-size: 0.75rem;
  color: var(--text-dim);
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
:deep(.name-text) { color: var(--text-body); }

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

:deep(.gender-male)   { background: #dbeafe; color: #1d4ed8; }
:deep(.gender-female) { background: #fce7f3; color: #be185d; }
:deep(.gender-other)  { background: var(--bg-surface); color: var(--text-dim); }

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
