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

export function getPopulationDashboard() {
  return fetchJSON('/population/dashboard')
}

export function getPopulationDemographics() {
  return fetchJSON('/population/demographics', 15000)
}

export function getPopulationEducation() {
  return fetchJSON('/population/education', 15000)
}

export function getPopulationEmployment() {
  return fetchJSON('/population/employment', 15000)
}

export function getPopulationRegistry(params = {}) {
  const query = new URLSearchParams(params).toString()
  return fetchJSON(`/population/registry${query ? `?${query}` : ''}`, MAP_DATA_TIMEOUT)
}

export function getPopulationMapData(params = {}) {
  const query = new URLSearchParams(params).toString()
  return fetchJSON(`/population/map-data${query ? `?${query}` : ''}`, MAP_DATA_TIMEOUT)
}

export function getPopulationMapSummary(params = {}) {
  const query = new URLSearchParams(params).toString()
  return fetchJSON(`/population/map-summary${query ? `?${query}` : ''}`)
}

export function getPopulationMapInsights(params = {}) {
  const query = new URLSearchParams(params).toString()
  return fetchJSON(`/population/map-insights${query ? `?${query}` : ''}`)
}
