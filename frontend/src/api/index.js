const BASE = '/api'
const TIMEOUT_DEFAULT = 15000
const TIMEOUT_DATA    = 30000  // large dataset queries can take longer
const RETRY_ATTEMPTS  = 2
const RETRY_DELAY_MS  = 400

function delay(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

function isRetryableError(err) {
  return err?.name === 'AbortError' || err instanceof TypeError
}

async function fetchJSON(url, timeoutMs = TIMEOUT_DEFAULT) {
  for (let attempt = 0; attempt <= RETRY_ATTEMPTS; attempt += 1) {
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), timeoutMs)

    try {
      const res = await fetch(`${BASE}${url}`, {
        signal: controller.signal,
        cache: 'no-store',
        headers: { Accept: 'application/json' },
      })

      if (!res.ok) {
        const isServerError = res.status >= 500
        if (isServerError && attempt < RETRY_ATTEMPTS) {
          await delay(RETRY_DELAY_MS * (attempt + 1))
          continue
        }
        throw new Error(`API error: ${res.status}`)
      }

      return res.json()
    } catch (err) {
      if (attempt < RETRY_ATTEMPTS && isRetryableError(err)) {
        await delay(RETRY_DELAY_MS * (attempt + 1))
        continue
      }
      throw err
    } finally {
      clearTimeout(timer)
    }
  }

  throw new Error('Request failed after retries')
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

export function getHousesMapPoints(params = {}) {
  const qs = toQueryString(params)
  return fetchJSON(qs ? `/houses/map-points?${qs}` : '/houses/map-points', TIMEOUT_DATA, 1)
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

export function getAgricultureInsights(params = {}) {
  const qs = toQueryString(params)
  return fetchJSON(qs ? `/insights/agriculture?${qs}` : '/insights/agriculture')
}

export function getDashboardSummary(params = {}) {
  const qs = toQueryString(params)
  return fetchJSON(qs ? `/dashboard/summary?${qs}` : '/dashboard/summary', TIMEOUT_DATA)
}

export function getWelfareInsights() {
  return fetchJSON('/insights/welfare')
}

export function getPopulationDashboard(params = {}) {
  const qs = toQueryString(params)
  return fetchJSON(qs ? `/population/dashboard?${qs}` : '/population/dashboard', TIMEOUT_DATA)
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
 * Fetch talukas for multiple selected districts
 * Accepts comma-separated district IDs (e.g., "1,2,3")
 * Backend should support: ?district_ids=1,2,3 or similar
 */
export function getTalukasByDistricts(districtIds) {
  if (!districtIds || districtIds.length === 0) {
    return Promise.resolve({ talukas: [] })
  }
  
  const ids = Array.isArray(districtIds) ? districtIds.join(',') : String(districtIds)
  return fetchJSON(`/location-options?district_ids=${encodeURIComponent(ids)}`, TIMEOUT_DATA)
}

/**
 * Fetch villages for multiple selected talukas
 * Accepts comma-separated taluka IDs (e.g., "1,2,3")
 * Backend should support: ?taluka_ids=1,2,3 or similar
 */
export function getVillagesByTalukas(talukaIds) {
  if (!talukaIds || talukaIds.length === 0) {
    return Promise.resolve({ villages: [] })
  }
  
  const ids = Array.isArray(talukaIds) ? talukaIds.join(',') : String(talukaIds)
  return fetchJSON(`/location-options?taluka_ids=${encodeURIComponent(ids)}`, TIMEOUT_DATA)
}

export function getDistricts() {
  return fetchJSON('/districts', TIMEOUT_DATA)
}

export function getDistrictSurveyCounts(params = {}) {
  const qs = toQueryString(params)
  return fetchJSON(qs ? `/map/district-survey-counts?${qs}` : '/map/district-survey-counts', TIMEOUT_DATA)
}

export function getDistrictCentroids(params = {}) {
  const qs = toQueryString(params)
  return fetchJSON(qs ? `/map/district-centroids?${qs}` : '/map/district-centroids', TIMEOUT_DATA)
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
  return fetchJSON(qs ? `/advisory?${qs}` : '/advisory')
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
