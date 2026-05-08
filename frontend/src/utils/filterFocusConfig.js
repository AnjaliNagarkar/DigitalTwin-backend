/**
 * filterFocusConfig.js
 * ─────────────────────────────────────────────────────────────────────────────
 * Universal Filter Mapper — maps every VIEW BY colorMode to a Focus Section.
 *
 * SOURCE OF TRUTH RULE:
 *   When house.members[] is loaded (after a house click), always compute counts
 *   directly from the member array — never from backend aggregate fields like
 *   membersWithAadhaar / totalFamilyMembers, which can return 0 if the
 *   EXTERNAL_FAMILY_ID JOIN in the aggregate subquery doesn't match.
 *   Fall back to aggregates only when members[] is empty/not yet loaded.
 *
 * DB field → JS property:
 *   FAMILY_BELONG_BPL_CATEGORY  → house.bplCategory
 *   RATION_CARD_TYPE            → house.rationCard
 *   ANNUAL_INCOME               → house.annualIncome
 *   ELECTRICITY_CONNECTION      → house.lighting
 *   SANITATION_TOILET_FACILITY  → house.latrine
 *   SOURCE_WATER_IRRIGATION     → house.waterSource
 *   FAMILY_MEMBER.AADHAAR       → member.aadhaar
 *   FAMILY_MEMBER.CASTE_CERT    → member.casteCertificate
 *   FAMILY_MEMBER.DIVYANG       → member.divyang
 *   FAMILY_MEMBER.DISABILITY_%  → member.disabilityPercentage
 *   FAMILY_MEMBER.OCCUPATION    → member.occupation
 *   FAMILY_MEMBER.DOB           → member.dob
 */

// ── Micro-helpers ─────────────────────────────────────────────────────────────

const n   = (v) => String(v ?? '').toLowerCase().trim()
const pct = (num, den) => (den ? Math.round((num / den) * 100) : 0)
const memberName = (m) =>
  [m.firstName, m.lastName].filter(Boolean).join(' ').trim() || 'Member'

/** Return the loaded members[] array, or null if not yet available. */
const getMembers = (house) => {
  const arr = house?.members
  return Array.isArray(arr) && arr.length > 0 ? arr : null
}

const isBpl = (house) => n(house.bplCategory) === 'yes'

const UNEMPLOYED_VALS = new Set([
  '', 'unemployed', 'not working', 'no work', 'nil', 'none', 'na', 'n/a',
])

// ── Per-filter stat helpers (member-first, aggregate fallback) ────────────────

/** Aadhaar: covered / missing / total — computed from members[] when loaded. */
function aadhaarStats(house) {
  const mems = getMembers(house)
  if (mems) {
    const covered = mems.filter(m => n(m.aadhaar) === 'yes').length
    return { covered, total: mems.length, missing: mems.length - covered }
  }
  // Fallback: backend aggregate (may be 0 if JOIN fails on EXTERNAL_FAMILY_ID)
  const covered = Number(house.membersWithAadhaar ?? 0)
  const total   = Number(house.totalFamilyMembers ?? house.totalMembers ?? 0)
  return { covered, total, missing: Math.max(total - covered, 0) }
}

/** Caste cert: covered / missing / total. */
function casteStats(house) {
  const mems = getMembers(house)
  if (mems) {
    const covered = mems.filter(m => n(m.casteCertificate) === 'yes').length
    return { covered, total: mems.length, missing: mems.length - covered }
  }
  const covered = Number(house.membersWithCasteCertificate ?? 0)
  const total   = Number(house.totalFamilyMembers ?? house.totalMembers ?? 0)
  return { covered, total, missing: Math.max(total - covered, 0) }
}

/** Employment: working / unemployed / total. */
function employmentStats(house) {
  const mems = getMembers(house)
  if (mems) {
    const unemployed = mems.filter(m => UNEMPLOYED_VALS.has(n(m.occupation))).length
    const working    = mems.length - unemployed
    return { working, unemployed, total: mems.length }
  }
  return {
    working:    Number(house.workingMembers    ?? 0),
    unemployed: Number(house.unemployedMembers ?? 0),
    total:      Number(house.totalMembers      ?? 0),
  }
}

/** Divyang: count of divyang members. */
function divyangStats(house) {
  const mems = getMembers(house)
  if (mems) {
    const divyang = mems.filter(m => n(m.divyang) === 'yes')
    return { count: divyang.length, members: divyang }
  }
  return { count: Number(house.divyangMembers ?? 0), members: [] }
}

// ── Main configuration ────────────────────────────────────────────────────────

export const FILTER_FOCUS_CONFIG = {

  // ── Document Gap filters ─────────────────────────────────────────────────

  document_gap: {
    title: 'Document Gap Summary',
    icon: '📋',
    accent: '#dc2626',
    metrics: (_house, gaps = []) => {
      const critical = gaps.filter(g => g.severity === 'critical').length
      const warning  = gaps.filter(g => g.severity === 'warning').length
      const info     = gaps.filter(g => g.severity === 'info').length
      return [
        { label: 'Critical', value: critical, icon: '🔴', status: critical > 0 ? 'critical' : 'ok' },
        { label: 'Warnings', value: warning,  icon: '🟡', status: warning  > 0 ? 'warn'     : 'ok' },
        { label: 'Info',     value: info,     icon: '🔵', status: 'neutral' },
      ]
    },
    status: (_house, gaps = []) => {
      if (!gaps.length)
        return { level: 'ok', message: 'No document gaps detected for this household.' }
      const critical = gaps.filter(g => g.severity === 'critical').length
      if (critical)
        return { level: 'critical', message: `${critical} critical gap(s) require immediate action.` }
      return { level: 'warn', message: `${gaps.length} gap(s) identified — review below.` }
    },
  },

  bpl_ration_status: {
    title: 'BPL & Ration Card',
    icon: '🧾',
    accent: '#d97706',
    metrics: (house) => [
      {
        label:  'BPL Status',
        value:  house.bplCategory || '—',
        icon:   '📋',
        status: isBpl(house) ? 'warn' : 'ok',
      },
      {
        label:  'Ration Card',
        value:  house.rationCard || 'Not Recorded',
        icon:   '🪪',
        status: house.rationCard ? 'ok' : 'critical',
      },
      {
        label:  'Annual Income',
        value:  house.annualIncome || '—',
        icon:   '₹',
        status: 'neutral',
      },
    ],
    status: (house) => {
      if (!isBpl(house))
        return { level: 'ok', message: 'Non-BPL household — no ration card action needed.' }
      const rc = n(house.rationCard || '')
      if (!rc || rc === '—')
        return { level: 'critical', message: 'BPL family has no ration card on record.' }
      if (rc.includes('bpl') || rc.includes('antyodaya'))
        return { level: 'ok', message: 'BPL status matches ration card type.' }
      return { level: 'warn', message: `BPL family but ration card shows "${house.rationCard}" — type mismatch.` }
    },
  },

  aadhaar_coverage: {
    title: 'Aadhaar Coverage',
    icon: '🆔',
    accent: '#2563eb',
    // Always computed from members[] when loaded — aggregate fields can be stale
    metrics: (house) => {
      const { covered, total, missing } = aadhaarStats(house)
      return [
        { label: 'Covered', value: covered, icon: '✅', status: missing === 0 && total > 0 ? 'ok' : 'warn' },
        { label: 'Missing', value: missing, icon: '❌', status: missing > 0 ? 'critical' : 'ok' },
        { label: 'Total',   value: total,   icon: '👥', status: 'neutral' },
      ]
    },
    members: (house) => {
      const mems = getMembers(house)
      if (!mems) return []
      return mems.map(m => ({
        name:   memberName(m),
        value:  n(m.aadhaar) === 'yes' ? 'Aadhaar recorded' : 'Missing Aadhaar',
        icon:   n(m.aadhaar) === 'yes' ? '✅' : '❌',
        status: n(m.aadhaar) === 'yes' ? 'ok' : 'critical',
      }))
    },
    status: (house) => {
      const { covered, total, missing } = aadhaarStats(house)
      if (total === 0)
        return { level: 'warn', message: 'No member records loaded yet.' }
      if (missing === 0)
        return { level: 'ok', message: `All ${total} member(s) have Aadhaar.` }
      if (covered === 0)
        return { level: 'critical', message: `No Aadhaar records for any of the ${total} member(s).` }
      return { level: 'warn', message: `${missing} of ${total} member(s) still need Aadhaar enrollment.` }
    },
  },

  caste_certificate_coverage: {
    title: 'Caste Certificate',
    icon: '📄',
    accent: '#7c3aed',
    metrics: (house) => {
      const { covered, total, missing } = casteStats(house)
      return [
        { label: 'Covered', value: covered, icon: '✅', status: missing === 0 && total > 0 ? 'ok' : 'warn' },
        { label: 'Missing', value: missing, icon: '⚠️', status: missing > 0 ? 'warn' : 'ok' },
        { label: 'Total',   value: total,   icon: '👥', status: 'neutral' },
      ]
    },
    members: (house) => {
      const mems = getMembers(house)
      if (!mems) return []
      return mems.map(m => ({
        name:   memberName(m),
        value:  n(m.casteCertificate) === 'yes' ? 'Certificate present' : 'Not recorded',
        icon:   n(m.casteCertificate) === 'yes' ? '✅' : '⚠️',
        status: n(m.casteCertificate) === 'yes' ? 'ok' : 'warn',
      }))
    },
    status: (house) => {
      const { covered, total, missing } = casteStats(house)
      if (total === 0)
        return { level: 'warn', message: 'No member records loaded yet.' }
      if (missing === 0)
        return { level: 'ok', message: `All ${total} member(s) have caste certificates.` }
      if (covered === 0)
        return { level: 'critical', message: `No caste certificate records for any of the ${total} member(s).` }
      return { level: 'warn', message: `${missing} of ${total} member(s) need caste certificates.` }
    },
  },

  unemployed_gap: {
    title: 'Employment Status',
    icon: '💼',
    accent: '#dc2626',
    metrics: (house) => {
      const { working, unemployed, total } = employmentStats(house)
      const rate = pct(working, total)
      return [
        { label: 'Working',    value: working,       icon: '✅', status: working > 0 ? 'ok' : 'critical' },
        { label: 'Unemployed', value: unemployed,    icon: '⚠️', status: unemployed > 0 ? 'warn' : 'ok' },
        { label: 'Work Rate',  value: `${rate}%`,    icon: '📊', status: rate >= 50 ? 'ok' : 'warn' },
      ]
    },
    members: (house) => {
      const mems = getMembers(house)
      if (!mems) return []
      return mems.map(m => {
        const employed = !UNEMPLOYED_VALS.has(n(m.occupation))
        return {
          name:   memberName(m),
          value:  m.occupation || 'No occupation recorded',
          icon:   employed ? '💼' : '⚠️',
          status: employed ? 'ok' : 'warn',
        }
      })
    },
    status: (house) => {
      const { working, unemployed, total } = employmentStats(house)
      if (total === 0)
        return { level: 'warn', message: 'No member records loaded yet.' }
      if (unemployed === 0 && working > 0)
        return { level: 'ok', message: 'All recorded members have occupations.' }
      if (unemployed > 0)
        return { level: 'warn', message: `${unemployed} of ${total} member(s) are unemployed or have no occupation recorded.` }
      return { level: 'warn', message: 'Occupation data incomplete for this household.' }
    },
  },

  divyang_gap: {
    title: 'Divyang & Disability',
    icon: '♿',
    accent: '#7c3aed',
    metrics: (house) => {
      const { count } = divyangStats(house)
      const total = getMembers(house)?.length ?? Number(house.totalMembers ?? 0)
      return [
        { label: 'Divyang',       value: count,                     icon: '♿', status: count > 0 ? 'warn' : 'ok' },
        { label: 'Total Members', value: total,                     icon: '👥', status: 'neutral' },
        { label: 'Others',        value: Math.max(total - count, 0), icon: '✅', status: 'neutral' },
      ]
    },
    members: (house) => {
      const { members: divMembers } = divyangStats(house)
      if (!divMembers.length) return []
      return divMembers.map(m => {
        const hasCert = m.disabilityPercentage &&
                        m.disabilityPercentage.trim() !== '' &&
                        m.disabilityPercentage !== '0'
        return {
          name:   memberName(m),
          value:  hasCert ? `${m.disabilityPercentage}% disability recorded` : 'Disability % not on record',
          icon:   hasCert ? '✅' : '❌',
          status: hasCert ? 'ok' : 'critical',
        }
      })
    },
    status: (house) => {
      const { count } = divyangStats(house)
      if (count === 0) return { level: 'ok', message: 'No divyang members recorded.' }
      return { level: 'warn', message: `${count} divyang member(s) — verify disability certificate and UDID enrollment.` }
    },
  },

  // ── Population / Social filters ───────────────────────────────────────────

  population_density: {
    title: 'Household Population',
    icon: '👥',
    accent: '#059669',
    metrics: (house) => {
      // Use the household member fields directly - these are the source of truth
      const total  = Number(house.totalMembers  ?? 0)
      const male   = Number(house.maleMembers   ?? 0)
      const female = Number(house.femaleMembers ?? 0)
      return [
        { label: 'Total',   value: total,  icon: '👨‍👩‍👧‍👦', status: 'neutral' },
        { label: 'Male',    value: male,   icon: '👨',       status: 'neutral' },
        { label: 'Female',  value: female, icon: '👩',       status: 'neutral' },
      ]
    },
    status: (house) => {
      const mems = getMembers(house)
      const t = mems ? mems.length : Number(house.totalMembers ?? 0)
      if (t === 0) return { level: 'warn', message: 'No member records found.' }
      if (t <= 2)  return { level: 'ok',   message: 'Small household (1–2 members).' }
      if (t <= 5)  return { level: 'ok',   message: 'Average household (3–5 members).' }
      return               { level: 'warn', message: `Large household (${t} members) — may need additional welfare support.` }
    },
  },

  education_level: {
    title: 'Literacy & Education',
    icon: '📚',
    accent: '#0891b2',
    metrics: (house) => {
      // Use household literacy fields directly - source of truth
      const ill   = Number(house.illiterateMembers ?? 0)
      const total = Number(house.totalMembers ?? 1)
      const rate  = pct(total - ill, total)
      return [
        { label: 'Illiterate',    value: ill,       icon: '📖', status: ill > 0 ? 'warn' : 'ok' },
        { label: 'Literacy Rate', value: `${rate}%`, icon: '📊', status: rate >= 70 ? 'ok' : 'warn' },
        { label: 'Total Members', value: total,     icon: '👥', status: 'neutral' },
      ]
    },
    status: (house) => {
      const ill = Number(house.illiterateMembers ?? 0)
      if (ill === 0) return { level: 'ok', message: 'All recorded members are literate.' }
      return { level: 'warn', message: `${ill} member(s) recorded as illiterate — Saakshar Bharat support recommended.` }
    },
  },

  divyang_presence: {
    title: 'Divyang Presence',
    icon: '♿',
    accent: '#7b1fa2',
    metrics: (house) => {
      // Use household divyang field directly - source of truth
      const count = Number(house.divyangMembers ?? 0)
      const total = Number(house.totalMembers ?? 0)
      return [
        { label: 'Divyang Members', value: count, icon: '♿', status: count > 0 ? 'warn' : 'ok' },
        { label: 'Total Members',   value: total, icon: '👥', status: 'neutral' },
      ]
    },
    status: (house) => {
      const count = Number(house.divyangMembers ?? 0)
      if (count === 0) return { level: 'ok', message: 'No divyang members in this household.' }
      return { level: 'warn', message: `${count} divyang member(s) — ensure pension and certificate enrollment.` }
    },
  },

  occupation: {
    title: 'Occupation Profile',
    icon: '👷',
    accent: '#7c3aed',
    metrics: (house) => {
      const { working, unemployed, total } = employmentStats(house)
      return [
        { label: 'Working',    value: working,    icon: '💼', status: working > 0 ? 'ok' : 'warn' },
        { label: 'Unemployed', value: unemployed, icon: '⚠️', status: unemployed > 0 ? 'warn' : 'ok' },
        { label: 'Primary',    value: house.occupation || '—', icon: '🏭', status: 'neutral' },
      ]
    },
    members: (house) => {
      const mems = getMembers(house)
      if (!mems) return []
      return mems.map(m => ({
        name:   memberName(m),
        value:  m.occupation || 'Not recorded',
        icon:   '💼',
        status: !UNEMPLOYED_VALS.has(n(m.occupation)) ? 'ok' : 'warn',
      }))
    },
    status: (house) => {
      const { working, total } = employmentStats(house)
      if (working === 0) return { level: 'critical', message: 'No working members recorded for this household.' }
      return { level: 'ok', message: `${working} of ${total} member(s) are working.` }
    },
  },

  bpl_status: {
    title: 'BPL Status',
    icon: '📋',
    accent: '#ef4444',
    metrics: (house) => [
      { label: 'BPL Category',  value: house.bplCategory  || '—', icon: '📋', status: isBpl(house) ? 'warn' : 'ok' },
      { label: 'Annual Income', value: house.annualIncome || '—', icon: '₹',  status: 'neutral' },
    ],
    status: (house) => {
      if (isBpl(house)) return { level: 'warn', message: 'BPL household — verify welfare scheme enrollment.' }
      return { level: 'ok', message: 'Non-BPL household.' }
    },
  },

  // ── Agriculture filters ───────────────────────────────────────────────────

  crops: {
    title: 'Crop Profile',
    icon: '🌾',
    accent: '#65a30d',
    metrics: (house) => [
      { label: 'Kharif', value: house.kharif || '—', icon: '🌾', status: house.kharif && n(house.kharif) !== 'no' ? 'ok' : 'warn' },
      { label: 'Rabi',   value: house.rabi   || '—', icon: '🌿', status: house.rabi   && n(house.rabi)   !== 'no' ? 'ok' : 'warn' },
      { label: 'Land',   value: house.totalLand ? `${house.totalLand} ac` : '—', icon: '🗺️', status: 'neutral' },
    ],
    status: (house) => {
      const k = house.kharif && n(house.kharif) !== 'no'
      const r = house.rabi   && n(house.rabi)   !== 'no'
      if (k && r) return { level: 'ok',  message: 'Double crop household — Kharif + Rabi.' }
      if (k || r) return { level: 'ok',  message: 'Single-season crop household.' }
      return        { level: 'warn', message: 'No crop data recorded.' }
    },
  },

  irrigation: {
    title: 'Irrigation & Water',
    icon: '💧',
    accent: '#0284c7',
    metrics: (house) => [
      {
        label: 'Water Source',
        value: house.waterSource || '—',
        icon:  '💧',
        status: house.waterSource && n(house.waterSource) !== 'no' ? 'ok' : 'warn',
      },
      { label: 'Land', value: house.totalLand ? `${house.totalLand} ac` : '—', icon: '🗺️', status: 'neutral' },
    ],
    status: (house) => {
      if (!house.waterSource || n(house.waterSource) === 'no')
        return { level: 'warn', message: 'No irrigation source — rain-fed farming risk.' }
      return { level: 'ok', message: `Irrigated via ${house.waterSource}.` }
    },
  },

  land: {
    title: 'Land Holdings',
    icon: '🗺️',
    accent: '#16a34a',
    metrics: (house) => [
      { label: 'Total',      value: house.totalLand      ? `${house.totalLand} ac`      : '—', icon: '🗺️', status: 'neutral' },
      { label: 'Cultivated', value: house.cultivatedLand ? `${house.cultivatedLand} ac` : '—', icon: '🌱', status: 'neutral' },
      { label: 'Own Land',   value: house.ownLand || '—', icon: '📋', status: 'neutral' },
    ],
    status: (house) => {
      const land = parseFloat(house.totalLand) || 0
      if (land === 0)  return { level: 'warn', message: 'Landless — eligible for land allotment and PM-KISAN support.' }
      if (land <= 1)   return { level: 'warn', message: `Marginal farmer (${land} ac) — small farmer benefit eligible.` }
      if (land <= 2.5) return { level: 'ok',   message: `Small farmer (${land} ac).` }
      return             { level: 'ok',   message: `Medium/large holding (${land} ac).` }
    },
  },

  // ── Infrastructure filters ────────────────────────────────────────────────

  lighting: {
    title: 'Electricity Connection',
    icon: '⚡',
    accent: '#ca8a04',
    metrics: (house) => [
      {
        label: 'Power Source',
        value: house.lighting || 'Not Recorded',
        icon:  '⚡',
        status: house.lighting && !['no', 'none', 'kerosene'].includes(n(house.lighting)) ? 'ok' : 'warn',
      },
    ],
    status: (house) => {
      const l = n(house.lighting || '')
      if (!l || l === 'no' || l === 'none')
        return { level: 'critical', message: 'No electricity connection recorded — Saubhagya scheme eligible.' }
      if (l === 'kerosene')
        return { level: 'warn', message: 'Using kerosene — eligible for grid connection under Saubhagya.' }
      return { level: 'ok', message: `Electricity source: ${house.lighting}.` }
    },
  },

  sanitation: {
    title: 'Sanitation Facility',
    icon: '🚽',
    accent: '#0891b2',
    metrics: (house) => [
      {
        label: 'Latrine Type',
        value: house.latrine || 'Not Recorded',
        icon:  '🚽',
        status: house.latrine && !['no', 'none', 'open defecation'].includes(n(house.latrine)) ? 'ok' : 'warn',
      },
    ],
    status: (house) => {
      const l = n(house.latrine || '')
      if (!l || l === 'no' || l === 'none' || l === 'open defecation')
        return { level: 'critical', message: 'No sanitation facility — Swachh Bharat Mission eligible.' }
      return { level: 'ok', message: `Sanitation type: ${house.latrine}.` }
    },
  },

  ration: {
    title: 'Ration Card',
    icon: '🪪',
    accent: '#16a34a',
    metrics: (house) => [
      { label: 'Card Type',     value: house.rationCard  || '—', icon: '🪪', status: house.rationCard ? 'ok' : 'warn' },
      { label: 'BPL Status',   value: house.bplCategory || '—', icon: '📋', status: isBpl(house) ? 'warn' : 'ok' },
      { label: 'Annual Income', value: house.annualIncome || '—', icon: '₹', status: 'neutral' },
    ],
    status: (house) => {
      if (!house.rationCard) return { level: 'warn', message: 'No ration card recorded for this household.' }
      return { level: 'ok', message: `Ration card: ${house.rationCard}.` }
    },
  },

  infrastructure: {
    title: 'Infrastructure Overview',
    icon: '🏠',
    accent: '#0891b2',
    metrics: (house) => [
      { label: 'Sanitation',  value: house.latrine    || '—', icon: '🚽', status: house.latrine    ? 'ok' : 'warn' },
      { label: 'Electricity', value: house.lighting   || '—', icon: '⚡', status: house.lighting   ? 'ok' : 'warn' },
      { label: 'Ration Card', value: house.rationCard || '—', icon: '🪪', status: house.rationCard ? 'ok' : 'warn' },
    ],
    status: (house) => {
      const gaps = [!house.latrine, !house.lighting, !house.rationCard].filter(Boolean).length
      if (gaps === 0) return { level: 'ok',      message: 'All three infrastructure indicators recorded.' }
      if (gaps === 1) return { level: 'warn',     message: '1 infrastructure record missing.' }
      return               { level: 'critical', message: `${gaps} infrastructure records missing.` }
    },
  },

  income_bracket: {
    title: 'Income Status',
    icon: '₹',
    accent: '#059669',
    metrics: (house) => [
      { label: 'Annual Income', value: house.annualIncome || '—', icon: '₹',  status: 'neutral' },
      { label: 'BPL Status',   value: house.bplCategory  || '—', icon: '📋', status: isBpl(house) ? 'warn' : 'ok' },
      { label: 'Ration Card',  value: house.rationCard   || '—', icon: '🪪', status: 'neutral' },
    ],
    status: (house) => {
      const income = n(house.annualIncome || '')
      if (!income) return { level: 'warn', message: 'Annual income not recorded.' }
      if (income.includes('less than') || income.includes('21000') || isBpl(house))
        return { level: 'warn', message: 'Low income bracket — welfare scheme enrollment recommended.' }
      return { level: 'ok', message: `Income: ${house.annualIncome}.` }
    },
  },
}

// ── Status / metric color metadata ───────────────────────────────────────────

export const STATUS_META = {
  ok:       { color: '#16a34a', bg: '#f0fdf4', border: '#bbf7d0', icon: '✓' },
  warn:     { color: '#d97706', bg: '#fffbeb', border: '#fde68a', icon: '⚠' },
  critical: { color: '#dc2626', bg: '#fef2f2', border: '#fecaca', icon: '!' },
}

export const METRIC_STATUS_COLOR = {
  ok:      '#16a34a',
  warn:    '#d97706',
  critical:'#dc2626',
  neutral: '#6b7280',
}

/**
 * buildFocusData(colorMode, house, gaps)
 * Resolves the config for a given mode and evaluates all functions
 * against the live house data. Returns null if no config exists.
 */
export function buildFocusData(colorMode, house, gaps = []) {
  if (!colorMode || !house) return null
  const cfg = FILTER_FOCUS_CONFIG[colorMode]
  if (!cfg) return null

  return {
    title:   cfg.title,
    icon:    cfg.icon,
    accent:  cfg.accent,
    metrics: cfg.metrics ? cfg.metrics(house, gaps) : [],
    members: cfg.members ? cfg.members(house, gaps) : [],
    status:  cfg.status  ? cfg.status(house, gaps)  : null,
  }
}
