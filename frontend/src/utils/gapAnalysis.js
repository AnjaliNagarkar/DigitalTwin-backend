/**
 * gapAnalysis.js
 * ─────────────────────────────────────────────────────────────────────────────
 * Configuration-driven Document Gap Analysis engine.
 *
 * Add or modify rules inside GAP_RULES without touching any UI code.
 * Each rule declares:
 *   id        – unique string key
 *   level     – 'family' | 'member'
 *   severity  – 'critical' | 'warning' | 'info'
 *   icon      – emoji shown in the drawer
 *   label     – short gap title
 *   detail    – dynamic detail function (receives family or member)
 *   scheme    – optional government scheme name
 *   check     – predicate (family, members?) → boolean
 *                 for member-level rules: (member, memberIndex, family) → boolean
 */

// ── helpers ──────────────────────────────────────────────────────────────────

const n = (v) => String(v ?? '').toLowerCase().trim()

/** Parse DD/MM/YYYY, YYYY-MM-DD, or plain YYYY and return age in full years. */
function ageFromDOB(dob) {
  if (!dob) return null
  let d
  if (/^\d{4}-\d{2}-\d{2}/.test(dob)) {
    d = new Date(dob)
  } else if (/^\d{2}\/\d{2}\/\d{4}$/.test(dob)) {
    const [dd, mm, yyyy] = dob.split('/')
    d = new Date(`${yyyy}-${mm}-${dd}`)
  } else if (/^\d{4}$/.test(dob)) {
    d = new Date(`${dob}-01-01`)
  } else {
    d = new Date(dob)
  }
  if (isNaN(d)) return null
  const today = new Date()
  let age = today.getFullYear() - d.getFullYear()
  const m = today.getMonth() - d.getMonth()
  if (m < 0 || (m === 0 && today.getDate() < d.getDate())) age--
  return age >= 0 ? age : null
}

function memberFullName(m) {
  return [m.firstName, m.lastName].filter(Boolean).join(' ').trim() || 'Member'
}

function isBplFamily(family) {
  return n(family.bplCategory) === 'yes'
}

function rationCardType(family) {
  return n(family.rationCard || family.RATION_CARD_TYPE || '')
}

// ── Gap Rule Configuration ────────────────────────────────────────────────────
// Every rule reads ONLY actual DB field values — no hardcoding.
// The Go backend returns '' for any column that is NULL or not present in the DB,
// so every check compares against real persisted values.
//
// To add a new rule: append an object to this array. No other file changes needed.
//
// DB field → JS property mapping:
//   FAMILY.FAMILY_BELONG_BPL_CATEGORY  → family.bplCategory
//   FAMILY.RATION_CARD_TYPE            → family.rationCard
//   FAMILY.ELECTRICITY_CONNECTION      → family.lighting
//   FAMILY.PRADHAN_MANTRI_AWAS         → family.PRADHAN_MANTRI_AWAS
//   FAMILY.TYPE_HOUSE                  → family.TYPE_HOUSE
//   aadhaar_coverage_status (computed) → family.aadhaarCoverageStatus
//   members_with_aadhaar   (computed)  → family.membersWithAadhaar
//   caste_cert_coverage_st (computed)  → family.casteCertificateCoverageStatus
//   members_with_caste_cert(computed)  → family.membersWithCasteCertificate
//   FAMILY_MEMBER.AADHAAR              → member.aadhaar
//   FAMILY_MEMBER.CASTE_CERTIFICATE    → member.casteCertificate
//   FAMILY_MEMBER.DIVYANG              → member.divyang
//   FAMILY_MEMBER.DISABILITY_PERCENTAGE→ member.disabilityPercentage
//   FAMILY_MEMBER.DOB                  → member.dob
//   FAMILY_MEMBER.OCCUPATION           → member.occupation

export const GAP_RULES = [
  // ════════════════════════════════════════════════════════════
  //  FAMILY-LEVEL RULES
  //  DB source: FAMILY table (returned in every /house/:id response)
  // ════════════════════════════════════════════════════════════

  {
    id: 'ration_mismatch',
    level: 'family',
    severity: 'critical',
    icon: '🧾',
    label: 'Ration Card Mismatch',
    // DB: FAMILY_BELONG_BPL_CATEGORY == 'YES' but RATION_CARD_TYPE ∉ {BPL, Antyodaya}
    detail: () => 'Family is classified BPL but ration card type does not reflect BPL/Antyodaya status.',
    scheme: 'PM Garib Kalyan Anna Yojana',
    check: (family) => {
      if (!isBplFamily(family)) return false
      const rc = rationCardType(family)
      return rc !== '' && !rc.includes('bpl') && !rc.includes('antyodaya')
    },
  },

  {
    id: 'bpl_no_ration',
    level: 'family',
    severity: 'critical',
    icon: '🪪',
    label: 'BPL — No Ration Card Recorded',
    // DB: FAMILY_BELONG_BPL_CATEGORY == 'YES' but RATION_CARD_TYPE is empty/null
    detail: () => 'Family is BPL-classified but no ration card is recorded in the database.',
    scheme: 'National Food Security Act (NFSA)',
    check: (family) => {
      if (!isBplFamily(family)) return false
      const rc = rationCardType(family)
      return rc === '' || rc === '—' || rc === 'no'
    },
  },

  {
    id: 'pmay_eligible',
    level: 'family',
    severity: 'info',
    icon: '🏠',
    label: 'Potential PMAY Eligibility',
    // DB: PRADHAN_MANTRI_AWAS is empty/null AND TYPE_HOUSE contains 'kutcha'/'temporary'/'mud'
    detail: (family) => `House type recorded as "${family.TYPE_HOUSE || 'unknown'}" — may be eligible for PMAY housing assistance.`,
    scheme: 'Pradhan Mantri Awas Yojana (PMAY)',
    check: (family) => {
      const pmay = n(family.PRADHAN_MANTRI_AWAS || '')
      const houseType = n(family.TYPE_HOUSE || '')
      return pmay === '' && (houseType.includes('kutcha') || houseType.includes('temporary') || houseType.includes('mud'))
    },
  },

  {
    id: 'no_electricity',
    level: 'family',
    severity: 'warning',
    icon: '⚡',
    label: 'No Electricity Connection',
    // DB: ELECTRICITY_CONNECTION / SOURCE_OF_LIGHTING is empty or negative value
    detail: () => 'No electricity connection recorded — may be eligible for Saubhagya scheme.',
    scheme: 'Pradhan Mantri Sahaj Bijli Har Ghar Yojana (Saubhagya)',
    check: (family) => {
      const light = n(family.lighting || family.SOURCE_OF_LIGHTING || '')
      return light === '' || light === 'no' || light === 'none' || light === 'kerosene'
    },
  },

  {
    id: 'aadhaar_partial',
    level: 'family',
    severity: 'warning',
    icon: '🆔',
    label: 'Incomplete Aadhaar Coverage',
    // DB: aadhaar_coverage_status (computed by backend from FAMILY_MEMBER.AADHAAR per family)
    //     membersWithAadhaar = COUNT(AADHAAR = 'yes'), totalFamilyMembers = COUNT(*)
    detail: (family) => {
      const covered = family.membersWithAadhaar ?? 0
      const total   = family.totalFamilyMembers ?? family.totalMembers ?? '?'
      const missing = Number.isFinite(total - covered) ? total - covered : '?'
      return `${covered} of ${total} members have Aadhaar — ${missing} member(s) still missing.`
    },
    scheme: 'Aadhaar Enrollment Drive (UIDAI)',
    check: (family) => {
      const status = n(family.aadhaarCoverageStatus)
      return status === 'partial' || status === 'missing'
    },
  },

  {
    id: 'caste_cert_partial',
    level: 'family',
    severity: 'warning',
    icon: '📄',
    label: 'Incomplete Caste Certificate Coverage',
    // DB: caste_certificate_coverage_status (computed from FAMILY_MEMBER.CASTE_CERTIFICATE per family)
    //     membersWithCasteCertificate = COUNT(CASTE_CERTIFICATE = 'yes')
    detail: (family) => {
      const covered = family.membersWithCasteCertificate ?? 0
      const total   = family.totalFamilyMembers ?? family.totalMembers ?? '?'
      const missing = Number.isFinite(total - covered) ? total - covered : '?'
      return `${covered} of ${total} members have a caste certificate — ${missing} member(s) missing.`
    },
    scheme: 'Caste Certificate Issuance (Revenue Dept)',
    check: (family) => {
      const status = n(family.casteCertificateCoverageStatus)
      return status === 'partial' || status === 'missing'
    },
  },

  // ════════════════════════════════════════════════════════════
  //  MEMBER-LEVEL RULES
  //  DB source: FAMILY_MEMBER table, fetched per-member in /house/:id
  //  Each member object has fields from the expanded MemberRecord struct.
  // ════════════════════════════════════════════════════════════

  {
    id: 'member_no_aadhaar',
    level: 'member',
    severity: 'critical',
    icon: '🆔',
    label: 'Missing Aadhaar',
    // DB: FAMILY_MEMBER.AADHAAR != 'yes' (null/empty/no)
    //     Age guard uses FAMILY_MEMBER.DOB — skips children under 5
    detail: (member) => `${memberFullName(member)} does not have an Aadhaar card recorded.`,
    scheme: 'Aadhaar Enrollment (UIDAI)',
    check: (member) => {
      const age = ageFromDOB(member.dob)
      if (age !== null && age < 5) return false   // infants don't need Aadhaar
      return n(member.aadhaar) !== 'yes'
    },
  },

  {
    id: 'member_no_caste_cert',
    level: 'member',
    severity: 'warning',
    icon: '📄',
    label: 'Missing Caste Certificate',
    // DB: FAMILY_MEMBER.CASTE_CERTIFICATE != 'yes'
    detail: (member) => `${memberFullName(member)} does not have a caste certificate recorded.`,
    scheme: 'Revenue Department — Certificate Issuance',
    check: (member) => n(member.casteCertificate) !== 'yes',
  },

  {
    id: 'divyang_no_cert',
    level: 'member',
    severity: 'critical',
    icon: '♿',
    label: 'Missing Disability Certificate',
    // DB: FAMILY_MEMBER.DIVYANG = 'yes' AND FAMILY_MEMBER.DISABILITY_PERCENTAGE is empty/null/0
    detail: (member) => `${memberFullName(member)} is registered as Divyang but disability percentage is not on record.`,
    scheme: 'Disability Pension · NHFDC · UDID Card',
    check: (member) => {
      const isDivyang = n(member.divyang) === 'yes'
      const hasPct = member.disabilityPercentage &&
                     member.disabilityPercentage.trim() !== '' &&
                     member.disabilityPercentage !== '0'
      return isDivyang && !hasPct
    },
  },

  {
    id: 'member_unemployed_adult',
    level: 'member',
    severity: 'warning',
    icon: '💼',
    label: 'Unemployed Adult',
    // DB: FAMILY_MEMBER.OCCUPATION is empty/negative AND FAMILY_MEMBER.DOB gives age 18–60
    detail: (member) => {
      const age = ageFromDOB(member.dob)
      return `${memberFullName(member)} (age ~${age ?? '?'}) — no occupation recorded in the database.`
    },
    scheme: 'e-Shram · DDU-GKY · MGNREGA',
    check: (member) => {
      const age = ageFromDOB(member.dob)
      if (age === null || age < 18 || age > 60) return false
      const occ = n(member.occupation)
      return occ === '' || occ === 'unemployed' || occ === 'not working' || occ === 'no work' || occ === 'nil'
    },
  },

]

// ── Severity metadata ─────────────────────────────────────────────────────────

export const SEVERITY_META = {
  critical: { color: '#dc2626', bg: '#fef2f2', border: '#fecaca', badge: 'Critical', order: 0 },
  warning:  { color: '#d97706', bg: '#fffbeb', border: '#fde68a', badge: 'Warning',  order: 1 },
  info:     { color: '#2563eb', bg: '#eff6ff', border: '#bfdbfe', badge: 'Info',     order: 2 },
}

// ── Main analyzeGaps function ─────────────────────────────────────────────────

/**
 * analyzeGaps(householdData)
 *
 * Runs every GAP_RULE against the household and its members.
 * Returns an array of gap objects sorted by severity (critical → warning → info).
 *
 * @param {Object} householdData – the full HouseDetail object from /house/:id
 * @returns {Array<GapResult>}
 */
export function analyzeGaps(householdData) {
  if (!householdData) return []

  const family  = householdData
  const members = Array.isArray(householdData.members) ? householdData.members : []
  const gaps    = []

  for (const rule of GAP_RULES) {
    if (rule.level === 'family') {
      try {
        if (rule.check(family, members)) {
          gaps.push({
            ruleId:   rule.id,
            level:    'family',
            severity: rule.severity,
            icon:     rule.icon,
            label:    rule.label,
            detail:   rule.detail(family, members),
            scheme:   rule.scheme || null,
          })
        }
      } catch (_) { /* skip rule on data error */ }
    } else if (rule.level === 'member') {
      members.forEach((member, idx) => {
        try {
          if (rule.check(member, idx, family)) {
            gaps.push({
              ruleId:      rule.id,
              level:       'member',
              severity:    rule.severity,
              icon:        rule.icon,
              label:       rule.label,
              detail:      rule.detail(member, idx, family),
              scheme:      rule.scheme || null,
              memberName:  memberFullName(member),
            })
          }
        } catch (_) { /* skip rule on data error */ }
      })
    }
  }

  // Sort by severity order then rule order for stable output
  gaps.sort((a, b) => {
    const sa = SEVERITY_META[a.severity]?.order ?? 99
    const sb = SEVERITY_META[b.severity]?.order ?? 99
    return sa - sb
  })

  return gaps
}

/**
 * gapSeverityForHouse(householdData)
 * Returns the worst severity level across all gaps, or null if none.
 * Used to drive the VIEW BY color mode.
 */
export function gapSeverityForHouse(householdData) {
  const gaps = analyzeGaps(householdData)
  if (!gaps.length) return null
  if (gaps.some(g => g.severity === 'critical')) return 'critical'
  if (gaps.some(g => g.severity === 'warning'))  return 'warning'
  return 'info'
}
