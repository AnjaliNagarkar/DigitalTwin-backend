const BASE = '/api'
const TIMEOUT_DEFAULT = 5000

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

export function getPopulationDashboard() {
  return fetchJSON('/population/dashboard')
}

export function getPopulationDemographics() {
  return fetchJSON('/population/demographics')
}

export function getPopulationEducation() {
  return fetchJSON('/population/education')
}

export function getPopulationEmployment() {
  return fetchJSON('/population/employment')
}

export function getPopulationMapData(params = {}) {
  const query = new URLSearchParams(params).toString()
  return fetchJSON(`/population/map-data${query ? `?${query}` : ''}`)
}

export function getPopulationMapSummary(params = {}) {
  const query = new URLSearchParams(params).toString()
  return fetchJSON(`/population/map-summary${query ? `?${query}` : ''}`)
}

export function getPopulationMapInsights(params = {}) {
  const query = new URLSearchParams(params).toString()
  return fetchJSON(`/population/map-insights${query ? `?${query}` : ''}`)
}
