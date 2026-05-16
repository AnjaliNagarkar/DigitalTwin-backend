/**
 * translateDynamicValue.js
 *
 * Frontend-only translation helpers for dynamic database category values.
 *
 * Rules:
 *  - Never mutates the original API response objects.
 *  - Falls back to the original DB value when no translation is found.
 *  - Never returns undefined or null — always returns a string.
 *  - Works with any locale key ('en', 'mr', etc.).
 *  - Scalable: add new locales by adding a field to each map entry.
 */

import { OCCUPATION_MAP } from '../translations/occupationTranslations.js'
import { CATEGORY_MAP }   from '../translations/categoryTranslations.js'
import { INCOME_MAP }     from '../translations/incomeTranslations.js'

/**
 * Normalize a DB string for map lookup.
 * Converts to UPPERCASE and trims whitespace.
 */
function normalizeKey(value) {
  return String(value ?? '').toUpperCase().trim()
}

/**
 * Translate a single occupation string.
 * Falls back to the original value if no mapping is found.
 *
 * @param {string} value  - Raw occupation string from DB.
 * @param {string} locale - Target locale ('en' | 'mr').
 * @returns {string}
 */
export function translateOccupation(value, locale = 'en') {
  if (value === null || value === undefined) return ''
  const str = String(value).trim()
  if (!str) return str
  const entry = OCCUPATION_MAP[normalizeKey(str)]
  if (!entry) return str
  return entry[locale] ?? entry.en ?? str
}

/**
 * Translate a single category string (house type, ration card, crop, etc.).
 * Falls back to the original value if no mapping is found.
 *
 * @param {string} value  - Raw category string from DB.
 * @param {string} locale - Target locale ('en' | 'mr').
 * @returns {string}
 */
export function translateCategory(value, locale = 'en') {
  if (value === null || value === undefined) return ''
  const str = String(value).trim()
  if (!str) return str
  const entry = CATEGORY_MAP[normalizeKey(str)]
  if (!entry) return str
  return entry[locale] ?? entry.en ?? str
}

/**
 * General-purpose dynamic value translator.
 * Checks occupation map first, then category map.
 * Falls back to the raw value when no translation is found.
 *
 * Use this as a catch-all in templates: td(value)
 *
 * @param {string|null|undefined} value  - Raw DB string value.
 * @param {string}                locale - Target locale ('en' | 'mr').
 * @returns {string}
 */
export function translateDynamicValue(value, locale = 'en') {
  if (value === null || value === undefined) return '—'
  const str = String(value).trim()
  if (!str) return '—'

  const key = normalizeKey(str)

  // Occupation map takes priority (most specific)
  if (OCCUPATION_MAP[key]) return translateOccupation(str, locale)

  // Income range map
  if (INCOME_MAP[key]) return translateIncome(str, locale)

  // Category map (house type, ration card, crop, etc.)
  if (CATEGORY_MAP[key]) return translateCategory(str, locale)

  // No mapping found — return original value as-is
  return str
}

/**
 * Translate an annual income range / category string.
 * Falls back to the original value if no mapping is found.
 * Purely numeric income values are returned as-is.
 *
 * @param {string} value  - Raw income label from DB (e.g. "Less than 21000").
 * @param {string} locale - Target locale ('en' | 'mr').
 * @returns {string}
 */
export function translateIncome(value, locale = 'en') {
  if (value === null || value === undefined) return ''
  const str = String(value).trim()
  if (!str) return str
  const entry = INCOME_MAP[normalizeKey(str)]
  if (!entry) return str
  return entry[locale] ?? entry.en ?? str
}

/**
 * Translate a pipe/comma-separated list of occupations.
 * Each segment is translated independently; the results are rejoined with ', '.
 *
 * Used for getWorkingOccupations() output which can be multi-valued.
 *
 * @param {string} value  - e.g. "Wage Work, Farmer" or "Wage Work|Farmer"
 * @param {string} locale - Target locale ('en' | 'mr').
 * @returns {string}
 */
export function translateOccupationDisplay(value, locale = 'en') {
  if (!value) return value || ''
  const str = String(value).trim()
  if (!str) return str

  // Split on pipe, comma, or semicolon
  return str
    .split(/[|,;]+/)
    .map(v => translateOccupation(v.trim(), locale))
    .filter(Boolean)
    .join(', ')
}

/**
 * Translate a pipe/comma-separated list of general category values.
 *
 * @param {string} value  - Pipe/comma-separated DB values.
 * @param {string} locale - Target locale ('en' | 'mr').
 * @returns {string}
 */
export function translateCategoryList(value, locale = 'en') {
  if (!value) return value || ''
  const str = String(value).trim()
  if (!str) return str

  return str
    .split(/[|,;]+/)
    .map(v => translateDynamicValue(v.trim(), locale))
    .filter(v => v && v !== '—')
    .join(', ')
}
