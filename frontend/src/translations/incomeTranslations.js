/**
 * Annual income range / category translation map.
 * Keys are UPPERCASE-normalized DB strings for income range values.
 *
 * DB stores values like "Less than 21000", "21001 to 50000", "50001 and above".
 * Numeric income fields (raw integers) are NOT translated here — only text labels.
 *
 * To extend to Hindi (hi): add an 'hi' field alongside 'en' and 'mr'.
 */
export const INCOME_MAP = {
  // ── Common income range text patterns ────────────────────────────────────
  'LESS THAN 21000':          { en: 'Less than ₹21,000',      mr: '₹21,000 पेक्षा कमी' },
  'LESS THAN 21,000':         { en: 'Less than ₹21,000',      mr: '₹21,000 पेक्षा कमी' },
  'BELOW 21000':              { en: 'Below ₹21,000',          mr: '₹21,000 पेक्षा कमी' },
  'BELOW 21,000':             { en: 'Below ₹21,000',          mr: '₹21,000 पेक्षा कमी' },
  'UPTO 21000':               { en: 'Up to ₹21,000',          mr: '₹21,000 पर्यंत' },
  'UP TO 21000':              { en: 'Up to ₹21,000',          mr: '₹21,000 पर्यंत' },

  '21001 TO 50000':           { en: '₹21,001 – ₹50,000',     mr: '₹21,001 – ₹50,000' },
  '21001 TO 50,000':          { en: '₹21,001 – ₹50,000',     mr: '₹21,001 – ₹50,000' },
  '21,001 TO 50,000':         { en: '₹21,001 – ₹50,000',     mr: '₹21,001 – ₹50,000' },
  '21001-50000':              { en: '₹21,001 – ₹50,000',     mr: '₹21,001 – ₹50,000' },
  '21001 - 50000':            { en: '₹21,001 – ₹50,000',     mr: '₹21,001 – ₹50,000' },

  '50001 AND ABOVE':          { en: '₹50,001 and above',      mr: '₹50,001 आणि त्यापेक्षा जास्त' },
  'ABOVE 50000':              { en: 'Above ₹50,000',          mr: '₹50,000 पेक्षा जास्त' },
  'ABOVE 50,000':             { en: 'Above ₹50,000',          mr: '₹50,000 पेक्षा जास्त' },
  'MORE THAN 50000':          { en: 'More than ₹50,000',      mr: '₹50,000 पेक्षा जास्त' },
  'MORE THAN 50,000':         { en: 'More than ₹50,000',      mr: '₹50,000 पेक्षा जास्त' },
  '50001+':                   { en: '₹50,001+',               mr: '₹50,001+' },
  '50,001+':                  { en: '₹50,001+',               mr: '₹50,001+' },

  // ── Broader income buckets ────────────────────────────────────────────────
  'LOW INCOME':               { en: 'Low Income',             mr: 'कमी उत्पन्न' },
  'MIDDLE INCOME':            { en: 'Middle Income',          mr: 'मध्यम उत्पन्न' },
  'HIGH INCOME':              { en: 'High Income',            mr: 'जास्त उत्पन्न' },
  'NO INCOME':                { en: 'No Income',              mr: 'उत्पन्न नाही' },
  'ZERO INCOME':              { en: 'No Income',              mr: 'उत्पन्न नाही' },
  'BPL INCOME':               { en: 'BPL Income',            mr: 'BPL उत्पन्न' },
}
