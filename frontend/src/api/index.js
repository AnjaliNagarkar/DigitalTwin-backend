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
