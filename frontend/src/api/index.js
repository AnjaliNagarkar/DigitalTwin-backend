const BASE = '/api'
const TIMEOUT_DEFAULT = 8000
const TIMEOUT_DATA    = 120000  // large dataset queries can take longer on cold DBs

function shouldRetry(status) {
  return status === 429 || status === 502 || status === 503 || status === 504
}

async function fetchJSON(url, timeoutMs = TIMEOUT_DEFAULT, retries = 1) {
  let attempt = 0
  while (attempt <= retries) {
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), timeoutMs)
    try {
      const res = await fetch(`${BASE}${url}`, { signal: controller.signal })
      if (!res.ok) {
        const err = new Error(`API error: ${res.status}`)
        err.status = res.status
        throw err
      }
      return res.json()
    } catch (error) {
      const status = Number(error?.status || 0)
      const timedOut = error?.name === 'AbortError'
      const retryable = timedOut || shouldRetry(status)
      if (attempt >= retries || !retryable) throw error
      await new Promise(resolve => setTimeout(resolve, 400 * (attempt + 1)))
      attempt += 1
    } finally {
      clearTimeout(timer)
    }
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
  return fetchJSON(qs ? `/houses?${qs}` : '/houses', TIMEOUT_DATA, 1)
}

export function getHousesByViewport(bbox, params = {}) {
  const qs = toQueryString({ ...bbox, ...params })
  return fetchJSON(`/houses?${qs}`, TIMEOUT_DATA, 1)
}

export function getHousesSummary(bbox, grid) {
  const qs = toQueryString({ ...bbox, ...(grid != null ? { grid } : {}) })
  return fetchJSON(`/houses/summary?${qs}`, TIMEOUT_DATA, 1)
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

export function getPopulationDashboard() {
  return fetchJSON('/population/dashboard', TIMEOUT_DATA)
}

export function getFarmers() {
  return fetchJSON('/farmers')
}

export function getCitizens() {
  return fetchJSON('/citizens')
}

export function getUnifiedRegistry() {
  return fetchJSON('/unified-registry', 180000, 1)
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
/**
 * Fetch group advisory for a cluster of households.
 * @param {Array<{key,count,total}>} problems - problem stats from cluster analysis
 * @param {number} total - total households in cluster
 */
export function getClusterAdvisory(problems, total) {
  if (!problems || !problems.length) return Promise.resolve({ actions: [] })
  const problemsParam = problems.map(p => `${p.key}:${p.count}:${p.total}`).join(',')
  return fetchJSON(`/advisory/cluster?problems=${encodeURIComponent(problemsParam)}&total=${total}`)
}

export function getSchemesForProblem(problemKey, profile = {}) {
  const params = { problem: problemKey, ...profile }
  const qs = toQueryString(params)
  return fetchJSON(`/schemes/recommend?${qs}`)
}
