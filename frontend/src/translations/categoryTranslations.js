/**
 * Category translation map for dynamic DB values (non-occupation fields).
 * Keys are UPPERCASE-normalized strings from the database.
 * Covers: house type, ration card, electricity, toilet, water source,
 *         ownership, BPL, coverage status, crops, PM Awas, wastewater, etc.
 *
 * To extend to Hindi (hi): add an 'hi' field alongside 'en' and 'mr'.
 */
export const CATEGORY_MAP = {

  // ── House / Structure Type ────────────────────────────────────────────────
  'PUCCA':            { en: 'Pucca',        mr: 'पक्के घर' },
  'KUCHA':            { en: 'Kucha',         mr: 'कच्चे घर' },
  'SEMI-PUCCA':       { en: 'Semi-Pucca',   mr: 'अर्ध-पक्के' },
  'SEMI PUCCA':       { en: 'Semi-Pucca',   mr: 'अर्ध-पक्के' },
  'KATCHA':           { en: 'Kucha',         mr: 'कच्चे घर' },

  // ── Ration Card Colors / Types ────────────────────────────────────────────
  'YELLOW':           { en: 'Yellow',        mr: 'पिवळे' },
  'ORANGE':           { en: 'Orange',        mr: 'केशरी' },
  'WHITE':            { en: 'White',         mr: 'पांढरे' },
  'SAFFRON':          { en: 'Saffron',       mr: 'भगवे' },
  'GREEN':            { en: 'Green',         mr: 'हिरवे' },
  'BLUE':             { en: 'Blue',          mr: 'निळे' },
  'PINK':             { en: 'Pink',          mr: 'गुलाबी' },
  'ANTYODAYA':        { en: 'Antyodaya',     mr: 'अंत्योदय' },
  'AAY':              { en: 'AAY',           mr: 'अंत्योदय (AAY)' },
  'NO CARD':          { en: 'No Card',       mr: 'कार्ड नाही' },
  'NO RATION CARD':   { en: 'No Ration Card', mr: 'शिधापत्रिका नाही' },

  // ── Electricity / Lighting ────────────────────────────────────────────────
  'ELECTRIC':         { en: 'Electric',      mr: 'विद्युत' },
  'ELECTRICITY':      { en: 'Electricity',   mr: 'विद्युत' },
  'SOLAR':            { en: 'Solar',         mr: 'सौर ऊर्जा' },
  'KEROSENE':         { en: 'Kerosene',      mr: 'रॉकेल' },
  'GENERATOR':        { en: 'Generator',     mr: 'जनरेटर' },
  'GRID':             { en: 'Grid',          mr: 'विद्युत जाळे' },
  'BATTERY':          { en: 'Battery',       mr: 'बॅटरी' },
  'BIOGAS':           { en: 'Biogas',        mr: 'बायोगॅस' },

  // ── Toilet / Sanitation ───────────────────────────────────────────────────
  'LEACH PIT':        { en: 'Leach Pit',     mr: 'लीच पिट' },
  'TWIN PIT':         { en: 'Twin Pit',      mr: 'ट्विन पिट' },
  'SEPTIC TANK':      { en: 'Septic Tank',   mr: 'सेप्टिक टाकी' },
  'OPEN DEFECATION':  { en: 'Open Defecation', mr: 'उघड्यावर शौच' },
  'PIT LATRINE':      { en: 'Pit Latrine',   mr: 'खड्डा संडास' },
  'FLUSH TOILET':     { en: 'Flush Toilet',  mr: 'फ्लश शौचालय' },
  'NO LATRINE':       { en: 'No Latrine',    mr: 'संडास नाही' },
  'NO TOILET':        { en: 'No Toilet',     mr: 'शौचालय नाही' },

  // ── Ownership ─────────────────────────────────────────────────────────────
  'OWN':              { en: 'Own',           mr: 'स्वतःचे' },
  'OWNED':            { en: 'Owned',         mr: 'स्वतःचे' },
  'RENT':             { en: 'Rented',        mr: 'भाड्याचे' },
  'RENTED':           { en: 'Rented',        mr: 'भाड्याचे' },
  'GOVERNMENT':       { en: 'Government',    mr: 'सरकारी' },
  'GOVT':             { en: 'Government',    mr: 'सरकारी' },
  'LEASE':            { en: 'Lease',         mr: 'भाडेपट्टी' },
  'SHARED':           { en: 'Shared',        mr: 'सामाईक' },

  // ── BPL / Poverty Category ────────────────────────────────────────────────
  'BPL':              { en: 'BPL',           mr: 'दारिद्र्यरेषेखाली' },
  'APL':              { en: 'APL',           mr: 'दारिद्र्यरेषेवरील' },
  'NON-BPL':          { en: 'Non-BPL',       mr: 'दारिद्र्यरेषेवरील' },
  'NON BPL':          { en: 'Non-BPL',       mr: 'दारिद्र्यरेषेवरील' },

  // ── Coverage / Document Status ────────────────────────────────────────────
  'COMPLETE':         { en: 'Complete',      mr: 'संपूर्ण' },
  'PARTIAL':          { en: 'Partial',       mr: 'आंशिक' },
  'MISSING':          { en: 'Missing',       mr: 'अनुपस्थित' },
  'UNKNOWN':          { en: 'Unknown',       mr: 'अज्ञात' },
  'FULL':             { en: 'Full',          mr: 'पूर्ण' },
  'NONE':             { en: 'None',          mr: 'काहीही नाही' },

  // ── General Yes / No ──────────────────────────────────────────────────────
  'YES':              { en: 'Yes',           mr: 'होय' },
  'NO':               { en: 'No',            mr: 'नाही' },
  'AVAILABLE':        { en: 'Available',     mr: 'उपलब्ध' },
  'NOT AVAILABLE':    { en: 'Not Available', mr: 'उपलब्ध नाही' },

  // ── Water Source / Irrigation ─────────────────────────────────────────────
  'RAIN':             { en: 'Rain',          mr: 'पाऊस' },
  'RAINFED':          { en: 'Rainfed',       mr: 'पावसाचे पाणी' },
  'RAIN FED':         { en: 'Rainfed',       mr: 'पावसाचे पाणी' },
  'CANAL':            { en: 'Canal',         mr: 'कालवा' },
  'WELL':             { en: 'Well',          mr: 'विहीर' },
  'BORE WELL':        { en: 'Bore Well',     mr: 'बोअर विहीर' },
  'BOREWELL':         { en: 'Bore Well',     mr: 'बोअर विहीर' },
  'BORE HOLE':        { en: 'Bore Hole',     mr: 'बोअर होल' },
  'RIVER':            { en: 'River',         mr: 'नदी' },
  'POND':             { en: 'Pond',          mr: 'तळे' },
  'TANK':             { en: 'Tank',          mr: 'टाकी' },
  'TAP WATER':        { en: 'Tap Water',     mr: 'नळाचे पाणी' },
  'TAP':              { en: 'Tap',           mr: 'नळ' },
  'HAND PUMP':        { en: 'Hand Pump',     mr: 'हातपंप' },
  'PIPELINE':         { en: 'Pipeline',      mr: 'पाइपलाइन' },
  'GOVT PIPELINE':    { en: 'Govt. Pipeline', mr: 'सरकारी पाइपलाइन' },
  'LAKE':             { en: 'Lake',          mr: 'तलाव' },
  'STREAM':           { en: 'Stream',        mr: 'ओढा' },
  'SPRING':           { en: 'Spring',        mr: 'झरा' },

  // ── Crop Names ────────────────────────────────────────────────────────────
  // Kharif
  'RICE':             { en: 'Rice',          mr: 'भात' },
  'PADDY':            { en: 'Paddy',         mr: 'भात' },
  'JOWAR':            { en: 'Jowar',         mr: 'ज्वारी' },
  'SORGHUM':          { en: 'Sorghum',       mr: 'ज्वारी' },
  'BAJRA':            { en: 'Bajra',         mr: 'बाजरी' },
  'PEARL MILLET':     { en: 'Pearl Millet',  mr: 'बाजरी' },
  'MAIZE':            { en: 'Maize',         mr: 'मका' },
  'CORN':             { en: 'Corn',          mr: 'मका' },
  'TUR':              { en: 'Tur',           mr: 'तूर' },
  'ARHAR':            { en: 'Tur (Arhar)',   mr: 'तूर' },
  'SOYBEAN':          { en: 'Soybean',       mr: 'सोयाबीन' },
  'COTTON':           { en: 'Cotton',        mr: 'कापूस' },
  'GROUNDNUT':        { en: 'Groundnut',     mr: 'शेंगदाणे' },
  'PEANUT':           { en: 'Peanut',        mr: 'शेंगदाणे' },
  'SUGARCANE':        { en: 'Sugarcane',     mr: 'ऊस' },
  'BANANA':           { en: 'Banana',        mr: 'केळी' },
  'URAD':             { en: 'Urad',          mr: 'उडीद' },
  'MOONG':            { en: 'Moong',         mr: 'मूग' },
  'GREEN GRAM':       { en: 'Green Gram',    mr: 'मूग' },
  'BLACK GRAM':       { en: 'Black Gram',    mr: 'उडीद' },
  'SESAME':           { en: 'Sesame',        mr: 'तीळ' },
  'TIL':              { en: 'Sesame',        mr: 'तीळ' },
  'TURMERIC':         { en: 'Turmeric',      mr: 'हळद' },
  // Rabi
  'WHEAT':            { en: 'Wheat',         mr: 'गहू' },
  'GRAM':             { en: 'Gram',          mr: 'हरभरा' },
  'CHICKPEA':         { en: 'Chickpea',      mr: 'हरभरा' },
  'CHANA':            { en: 'Gram',          mr: 'हरभरा' },
  'HARBHARA':         { en: 'Gram',          mr: 'हरभरा' },
  'HARKIS':           { en: 'Gram',          mr: 'हरभरा' },
  'LENTIL':           { en: 'Lentil',        mr: 'मसूर' },
  'MASOOR':           { en: 'Lentil',        mr: 'मसूर' },
  'MUSTARD':          { en: 'Mustard',       mr: 'मोहरी' },
  'RAPESEED':         { en: 'Rapeseed',      mr: 'मोहरी' },
  'VEGETABLES':       { en: 'Vegetables',    mr: 'भाजीपाला' },
  'ONION':            { en: 'Onion',         mr: 'कांदा' },
  'GARLIC':           { en: 'Garlic',        mr: 'लसूण' },
  'SUNFLOWER':        { en: 'Sunflower',     mr: 'सूर्यफूल' },
  'FLAX':             { en: 'Flax',          mr: 'अळशी' },
  'POTATO':           { en: 'Potato',        mr: 'बटाटा' },
  'TOMATO':           { en: 'Tomato',        mr: 'टोमॅटो' },
  'GRAPES':           { en: 'Grapes',        mr: 'द्राक्षे' },
  'MANGO':            { en: 'Mango',         mr: 'आंबा' },
  'ORANGE':           { en: 'Orange (Fruit)', mr: 'संत्री' },
  'POMEGRANATE':      { en: 'Pomegranate',   mr: 'डाळिंब' },

  // ── Wastewater System ─────────────────────────────────────────────────────
  'SOAKPIT':          { en: 'Soakpit',       mr: 'सोकपिट' },
  'SOAK PIT':         { en: 'Soakpit',       mr: 'सोकपिट' },
  'DRAINAGE':         { en: 'Drainage',      mr: 'निचरा' },
  'OPEN DRAIN':       { en: 'Open Drain',    mr: 'उघडी गटार' },
  'CLOSED DRAIN':     { en: 'Closed Drain',  mr: 'बंद गटार' },
  'NO SYSTEM':        { en: 'No System',     mr: 'यंत्रणा नाही' },

  // ── Land Ownership ────────────────────────────────────────────────────────
  'OWNED LAND':       { en: 'Owned Land',    mr: 'स्वतःची जमीन' },
  'LEASED LAND':      { en: 'Leased Land',   mr: 'भाडेपट्टी जमीन' },
  'TENANTED':         { en: 'Tenanted',      mr: 'कुळाने घेतलेली' },

  // ── Other common values ───────────────────────────────────────────────────
  'OTHER':            { en: 'Other',         mr: 'इतर' },
  'N/A':              { en: 'N/A',           mr: 'उपलब्ध नाही' },
  'NA':               { en: 'N/A',           mr: 'उपलब्ध नाही' },
  'NOT APPLICABLE':   { en: 'Not Applicable', mr: 'लागू नाही' },
  'NIL':              { en: 'Nil',           mr: 'शून्य' },
}
