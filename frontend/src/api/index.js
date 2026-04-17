const BASE = '/api'
const TIMEOUT_DEFAULT = 5000
const TIMEOUT_DATA    = 30000  // large dataset queries can take longer

async function fetchJSON(url, timeoutMs = TIMEOUT_DEFAULT) {
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeoutMs)
  try {
    const res = await fetch(`${BASE}${url}`, { signal: controller.signal })
    if (!res.ok) throw new Error(`API error: ${res.status}`)
    return res.json()
  } finally {
    clearTimeout(timer)
  }
}

function toQueryString(params = {}) {
  const cleaned = Object.entries(params).filter(([, value]) => {
    if (value === undefined || value === null) return false
    if (typeof value === 'string' && value.trim() === '') return false
    return true
  })
  return new URLSearchParams(cleaned).toString()
}

export function getHouses(params = {}) {
  const qs = toQueryString(params)
  return fetchJSON(qs ? `/houses?${qs}` : '/houses', TIMEOUT_DATA)
}

export function getHouseById(id) {
  return fetchJSON(`/house/${id}`)
}

export function getGovernanceInsights() {
  return fetchJSON('/insights/governance')
}

export function getAgricultureInsights() {
  return fetchJSON('/insights/agriculture')
}

export function getWelfareInsights() {
  return fetchJSON('/insights/welfare')
}

export function getFarmers() {
  return fetchJSON('/farmers')
}

export function getCitizens() {
  return fetchJSON('/citizens')
}

export function getUnifiedRegistry() {
  return fetchJSON('/unified-registry', TIMEOUT_DATA)
}

export function getCrops() {
  return fetchJSON('/crops')
}

export function getLand() {
  return fetchJSON('/land')
}

export function getIrrigation() {
  return fetchJSON('/irrigation')
}

export function getGeoClusters() {
  return fetchJSON('/geo-clusters', TIMEOUT_DATA)
}

export function getLocationOptions(params = {}) {
  const qs = toQueryString(params)
  return fetchJSON(qs ? `/location-options?${qs}` : '/location-options', TIMEOUT_DATA)
}

/**
 * Fetch farm advisory for a household's identified problems.
 * @param {string[]} problems  - array of problem keys e.g. ['noIrrigation','singleSeason']
 * @param {object}   profile   - { crop, land_size, bpl, family_id }
 */
export function getAdvisory(problems, profile = {}) {
  if (!problems || !problems.length) return Promise.resolve({ issues: [] })
  const params = { problems: problems.join(','), ...profile }
  const qs = toQueryString(params)
  return fetchJSON(`/advisory?${qs}`)
}

/**
 * Fetch scheme recommendations for a given problem key + optional citizen profile.
 * @param {string} problemKey  - e.g. 'noIrrigation', 'bplFamilies'
 * @param {object} profile     - { land_size, occupation, bpl }
 */
export function getSchemesForProblem(problemKey, profile = {}) {
  const params = { problem: problemKey, ...profile }
  const qs = toQueryString(params)
  return fetchJSON(`/schemes/recommend?${qs}`)
}
