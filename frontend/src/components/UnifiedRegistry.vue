<template>
  <div class="registry-wrap" :class="{ 'registry-embedded': embedded }">

    <!-- ── Page header (full-page mode only) ─────────────────────────── -->
    <header v-if="!embedded" class="page-header">
      <div class="header-title-row">
        <div>
          <h1 class="page-title">Citizen Registry</h1>
          <p class="page-subtitle">{{ activeCategoryConfig.subtitle }}</p>
        </div>
        <!-- Category dropdown — the master controller -->
        <div class="category-select-wrap">
          <label class="category-label">Category</label>
          <div class="category-select-box">
            <select v-model="category" class="category-select" @change="onCategoryChange">
              <option v-for="c in CATEGORIES" :key="c.value" :value="c.value">
                {{ c.label }}
              </option>
            </select>
            <span class="category-icon">{{ activeCategoryConfig.icon }}</span>
          </div>
        </div>
      </div>

      <!-- Search + sub-filters row -->
      <div class="header-controls">
        <div class="search-box">
          <svg viewBox="0 0 20 20" fill="currentColor" class="search-icon">
            <path fill-rule="evenodd"
              d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
              clip-rule="evenodd"/>
          </svg>
          <input v-model="search" placeholder="Search by name…" class="search-input"/>
        </div>

        <!-- Sub-filters — rendered dynamically per category -->
        <div v-if="activeSubFilters.length" class="subfilter-bar">
          <template v-for="sf in activeSubFilters" :key="sf.key">
            <span class="subfilter-label">{{ sf.label }}:</span>
            <button v-for="opt in sf.options" :key="opt.value"
              class="chip"
              :class="{ active: subFilters[sf.key] === opt.value }"
              @click="toggleSubFilter(sf.key, opt.value)">
              {{ opt.label }}
            </button>
            <div class="filter-divider"></div>
          </template>
        </div>

        <button class="reset-btn" @click="resetFilters">Reset</button>
      </div>
    </header>

    <!-- ── Embedded compact toolbar ──────────────────────────────────── -->
    <div v-if="embedded" class="embedded-toolbar">
      <select v-model="category" class="category-select-inline" @change="onCategoryChange">
        <option v-for="c in CATEGORIES" :key="c.value" :value="c.value">{{ c.label }}</option>
      </select>
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
          :class="{ active: subFilters[sf.key] === opt.value }"
          @click="toggleSubFilter(sf.key, opt.value)">
          {{ opt.label }}
        </button>
      </template>
      <button class="reset-btn" @click="resetFilters">Reset</button>
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
import { ref, computed, onMounted, watch } from 'vue'
import { getUnifiedRegistry } from '../api/index.js'

// ── Props ──────────────────────────────────────────────────────────────────
const props = defineProps({
  embedded: { type: Boolean, default: false },
  maxRows:  { type: Number,  default: 0 },
})

// ── Category definitions ───────────────────────────────────────────────────
// Each category declares: icon, label, subtitle, color, which columns to show,
// which sub-filters to expose, and how to pre-filter the rows.

const CATEGORIES = [
  { value: '',         label: 'All Citizens'   },
  { value: 'farmer',  label: 'Farmer'          },
  { value: 'student', label: 'Student'         },
  { value: 'disabled',label: 'Disabled'        },
  { value: 'housewife',label: 'Housewife'      },
  { value: 'senior',  label: 'Senior Citizen'  },
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
    subtitle: 'Agricultural land-owners with crop & irrigation data',
    icon: '🌾',
    color: '#16a34a',
    rowFilter: r => r.isFarmer,
    columns: [
      { key: 'fullName',       label: 'Full Name',     minWidth: true, tdClass: 'td-name' },
      { key: 'totalLand',      label: 'Land (Acre)',    tdClass: 'td-num' },
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
    label: 'Disabled',
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
      { key: 'annualIncome',      label: 'Annual Income',     tdClass: 'td-num' },
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
      { key: 'sourceOfIncome', label: 'Source of Income' },
      { key: 'annualIncome',   label: 'Annual Income',    tdClass: 'td-num' },
      { key: 'childrenCount',  label: 'Children',         tdClass: 'td-num' },
    ],
    subFilters: [
      {
        key: 'sourceOfIncome', label: 'Source of Income',
        options: [
          { label: 'None',                   value: 'None'                   },
          { label: 'Small Business/SHG',     value: 'Small Business/SHG'     },
          { label: 'Tailoring',              value: 'Tailoring'              },
          { label: 'Poultry',                value: 'Poultry'                },
          { label: 'Remittance from family', value: 'Remittance from family' },
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
const category    = ref('')
const search      = ref('')
const subFilters  = ref({})   // { [filterKey]: selectedValue }
const sortKey     = ref('')
const sortDir     = ref('asc')
const currentPage = ref(1)
const pageSize    = computed(() => props.maxRows > 0 ? props.maxRows : 50)

// ── Derived config ─────────────────────────────────────────────────────────
const activeCategoryConfig = computed(() => CATEGORY_CONFIG[category.value] || CATEGORY_CONFIG[''])
const activeSubFilters     = computed(() => activeCategoryConfig.value.subFilters || [])

// ── Data loading ───────────────────────────────────────────────────────────
onMounted(async () => {
  try { records.value = await getUnifiedRegistry() }
  catch (e) { console.error('Citizen registry load failed:', e) }
  finally { loading.value = false }
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

function toggleSort(key) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
}

function toggleSubFilter(key, value) {
  subFilters.value = {
    ...subFilters.value,
    [key]: subFilters.value[key] === value ? '' : value,
  }
}

function onCategoryChange() {
  // Clear all sub-filters when switching categories to avoid data conflicts
  subFilters.value = {}
  sortKey.value = ''
  sortDir.value = 'asc'
  currentPage.value = 1
}

function resetFilters() {
  search.value = ''
  subFilters.value = {}
  sortKey.value = ''
  sortDir.value = 'asc'
  currentPage.value = 1
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

  // Gender sub-filter
  if (sf.gender) {
    list = list.filter(r => String(r.gender || '').toLowerCase() === sf.gender.toLowerCase())
  }

  // Land size sub-filter (farmer category)
  if (sf.landSize) {
    list = list.filter(r => {
      const ac = parseLand(r.totalLand)
      if (sf.landSize === 'large')    return ac > 2
      if (sf.landSize === 'medium')   return ac >= 1 && ac <= 2
      if (sf.landSize === 'marginal') return ac > 0 && ac < 1
      return true
    })
  }

  // Irrigation type sub-filter (farmer category)
  if (sf.irrigationType) {
    list = list.filter(r => r.irrigationType === sf.irrigationType)
  }

  // Crop type sub-filter (farmer category)
  if (sf.cropType) {
    list = list.filter(r => {
      const hasKharif = r.kharifCrop && r.kharifCrop !== 'N/A'
      const hasRabi   = r.rabiCrop   && r.rabiCrop   !== 'N/A'
      if (sf.cropType === 'kharif') return hasKharif
      if (sf.cropType === 'rabi')   return hasRabi
      if (sf.cropType === 'both')   return hasKharif && hasRabi
      return true
    })
  }

  // Education level sub-filter (student category)
  if (sf.educationLevel) {
    list = list.filter(r => r.educationLevel === sf.educationLevel)
  }

  // Scholarship sub-filter (student category)
  if (sf.scholarship) {
    list = list.filter(r => r.scholarship === sf.scholarship)
  }

  // Source-of-income sub-filter (housewife category)
  if (sf.sourceOfIncome) {
    list = list.filter(r => r.sourceOfIncome === sf.sourceOfIncome)
  }

  // Income range — global, works across ALL categories
  if (sf.incomeRange) {
    list = list.filter(r => classifyIncome(r.annualIncome) === sf.incomeRange)
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

watch([search, subFilters, category], () => { currentPage.value = 1 }, { deep: true })

// ── Cell renderer ──────────────────────────────────────────────────────────
// Returns an HTML string so the template can use v-html for badge rendering.
// All name values are already capitalised by the backend.

function renderCell(r, col) {
  const v = r[col.key]

  switch (col.key) {
    case 'fullName': {
      const gClass = genderClass(r.gender)
      const tag = r.gender
        ? `<span class="gender-tag ${gClass}">${esc(r.gender)}</span>`
        : ''
      return `<span class="name-text">${esc(r.fullName)}</span>${tag}`
    }

    case 'totalLand': {
      const ac = parseLand(v)
      if (ac === 0) return `<span class="badge badge-muted">No Land</span>`
      return `<span class="badge badge-green">${esc(v)} ac</span>`
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
.page-header { margin-bottom: 1.5rem; }

.header-title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1.5rem;
  flex-wrap: wrap;
}

.page-title {
  font-family: var(--font-display);
  font-size: 2rem;
  color: var(--text-primary);
  font-weight: 400;
}
.page-subtitle { color: var(--text-dim); font-size: 0.8rem; margin-top: 0.35rem; }

/* ── Category selector ───────────────────────────────────────────────────── */
.category-select-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  min-width: 200px;
}
.category-label {
  font-size: 0.65rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-dim);
  font-weight: 600;
}
.category-select-box {
  position: relative;
  display: flex;
  align-items: center;
}
.category-select {
  width: 100%;
  height: 38px;
  background: var(--bg-card);
  border: 1.5px solid var(--amber);
  border-radius: 8px;
  color: var(--text-body);
  font-family: var(--font-body);
  font-size: 0.88rem;
  font-weight: 500;
  padding: 0 2rem 0 0.75rem;
  appearance: none;
  cursor: pointer;
  transition: box-shadow 0.2s;
}
.category-select:focus { outline: none; box-shadow: 0 0 0 3px var(--amber-dim); }
.category-icon {
  position: absolute;
  right: 0.6rem;
  font-size: 1rem;
  pointer-events: none;
}

/* Inline category selector (embedded mode) */
.category-select-inline {
  height: 30px;
  background: var(--bg-card);
  border: 1px solid var(--amber);
  border-radius: 6px;
  color: var(--text-body);
  font-family: var(--font-body);
  font-size: 0.8rem;
  padding: 0 0.6rem;
  cursor: pointer;
}

/* ── Controls row ────────────────────────────────────────────────────────── */
.header-controls {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-top: 1rem;
  flex-wrap: wrap;
}

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
  gap: 0.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0.5rem 0.75rem;
  flex: 0 1 260px;
  transition: border-color 0.2s;
}
.search-box-sm { flex: 0 1 200px; }
.search-box:focus-within { border-color: var(--amber); }
.search-icon { width: 16px; height: 16px; color: var(--text-dim); flex-shrink: 0; }
.search-input {
  background: none; border: none; outline: none;
  color: var(--text-body); font-family: var(--font-body); font-size: 0.85rem; width: 100%;
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
.filter-divider { width: 1px; height: 20px; background: var(--border); margin: 0 0.1rem; }

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
  height: 30px;
  padding: 0 0.9rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--bg-surface);
  color: var(--text-muted);
  font-family: var(--font-body);
  font-size: 0.76rem;
  cursor: pointer;
  transition: all 0.18s;
  white-space: nowrap;
}
.reset-btn:hover { border-color: var(--amber); color: var(--amber); }

/* ── Table container ─────────────────────────────────────────────────────── */
.table-container {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
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
