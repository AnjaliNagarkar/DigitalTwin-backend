const BASE = '/api'
const TIMEOUT_DEFAULT = 15000
const MAP_DATA_TIMEOUT = 45000
const RETRY_ATTEMPTS = 2
const RETRY_DELAY_MS = 400

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

export function getPopulationDashboard(params = {}) {
  const query = toQueryString(params)
  return fetchJSON(`/population/dashboard${query ? `?${query}` : ''}`)
}

export function getPopulationDemographics(params = {}) {
  const query = toQueryString(params)
  return fetchJSON(`/population/demographics${query ? `?${query}` : ''}`, 15000)
}

export function getPopulationEducation(params = {}) {
  const query = toQueryString(params)
  return fetchJSON(`/population/education${query ? `?${query}` : ''}`, 15000)
}

export function getPopulationEmployment(params = {}) {
  const query = toQueryString(params)
  return fetchJSON(`/population/employment${query ? `?${query}` : ''}`, 15000)
}

export function getPopulationRegistry(params = {}) {
  const query = toQueryString(params)
  return fetchJSON(`/population/registry${query ? `?${query}` : ''}`, MAP_DATA_TIMEOUT)
}

export function getPopulationMapData(params = {}) {
  const query = toQueryString(params)
  return fetchJSON(`/population/map-data${query ? `?${query}` : ''}`, MAP_DATA_TIMEOUT)
}

export function getDivyangDistrictCounts() {
  return fetchJSON('/divyang/district-count')
}

export function getPopulationMapSummary(params = {}) {
  const query = toQueryString(params)
  return fetchJSON(`/population/map-summary${query ? `?${query}` : ''}`)
}

export function getPopulationMapInsights(params = {}) {
  const query = toQueryString(params)
  return fetchJSON(`/population/map-insights${query ? `?${query}` : ''}`)
}
