/**
 * Occupation translation map.
 * Keys are UPPERCASE-normalized DB values from FAMILY_MEMBER.OCCUPATION.
 * Add new entries here as new occupation types appear in the database.
 * Supports: en (English), mr (Marathi). Extend with 'hi' for Hindi support.
 */
export const OCCUPATION_MAP = {
  // Core farm-based occupations
  'FARMER':                        { en: 'Farmer',                mr: 'शेतकरी' },
  'SELF EMPLOYED - FARM BASED':    { en: 'Self Employed (Farm)',  mr: 'स्वयंरोजगार (शेती)' },
  'SELF EMPLOYED-FARM BASED':      { en: 'Self Employed (Farm)',  mr: 'स्वयंरोजगार (शेती)' },
  'SELF EMPLOYED FARM BASED':      { en: 'Self Employed (Farm)',  mr: 'स्वयंरोजगार (शेती)' },
  'AGRICULTURE':                   { en: 'Agriculture',           mr: 'शेती' },
  'AGRICULTURAL LABOUR':           { en: 'Agricultural Labour',   mr: 'शेतमजूर' },
  'FARM LABOUR':                   { en: 'Farm Labour',           mr: 'शेतमजूर' },

  // Wage and daily labour
  'WAGE WORK':                     { en: 'Wage Work',             mr: 'मजुरी काम' },
  'DAILY WAGE LABOUR':             { en: 'Daily Wage Labour',     mr: 'दैनंदिन मजुरी' },
  'DAILY WAGE':                    { en: 'Daily Wage',            mr: 'दैनंदिन मजुरी' },
  'LABOUR':                        { en: 'Labour',                mr: 'मजुरी' },
  'LABOURER':                      { en: 'Labourer',              mr: 'मजूर' },
  'UNSKILLED LABOUR':              { en: 'Unskilled Labour',      mr: 'अकुशल मजूर' },
  'SKILLED LABOUR':                { en: 'Skilled Labour',        mr: 'कुशल मजूर' },
  'CONSTRUCTION':                  { en: 'Construction',          mr: 'बांधकाम' },
  'CONSTRUCTION WORKER':           { en: 'Construction Worker',   mr: 'बांधकाम कामगार' },

  // Service and employment
  'SERVICE':                       { en: 'Service',               mr: 'नोकरी' },
  'GOVERNMENT SERVICE':            { en: 'Government Service',    mr: 'सरकारी नोकरी' },
  'GOVT SERVICE':                  { en: 'Government Service',    mr: 'सरकारी नोकरी' },
  'PRIVATE SERVICE':               { en: 'Private Service',       mr: 'खाजगी नोकरी' },
  'SALARIED':                      { en: 'Salaried',              mr: 'वेतनधारी' },
  'TEACHER':                       { en: 'Teacher',               mr: 'शिक्षक' },
  'DOCTOR':                        { en: 'Doctor',                mr: 'डॉक्टर' },
  'NURSE':                         { en: 'Nurse',                 mr: 'परिचारिका' },
  'DRIVER':                        { en: 'Driver',                mr: 'चालक' },
  'POLICE':                        { en: 'Police',                mr: 'पोलीस' },
  'ENGINEER':                      { en: 'Engineer',              mr: 'अभियंता' },

  // Self-employment and business
  'BUSINESS':                      { en: 'Business',              mr: 'व्यवसाय' },
  'SELF EMPLOYED':                 { en: 'Self Employed',         mr: 'स्वयंरोजगार' },
  'SELF-EMPLOYED':                 { en: 'Self Employed',         mr: 'स्वयंरोजगार' },
  'TRADER':                        { en: 'Trader',                mr: 'व्यापारी' },
  'SHOPKEEPER':                    { en: 'Shopkeeper',            mr: 'दुकानदार' },
  'ARTISAN':                       { en: 'Artisan',               mr: 'कारागीर' },
  'CRAFTSMAN':                     { en: 'Craftsman',             mr: 'कारागीर' },

  // Non-working (may appear in data; kept for completeness)
  'HOUSEWIFE':                     { en: 'Housewife',             mr: 'गृहिणी' },
  'HOMEMAKER':                     { en: 'Homemaker',             mr: 'गृहिणी' },
  'STUDENT':                       { en: 'Student',               mr: 'विद्यार्थी' },
  'STUDYING':                      { en: 'Studying',              mr: 'शिक्षण घेत आहे' },
  'UNEMPLOYED':                    { en: 'Unemployed',            mr: 'बेरोजगार' },
  'NOT WORKING':                   { en: 'Not Working',           mr: 'काम नाही' },
  'RETIRED':                       { en: 'Retired',               mr: 'निवृत्त' },
  'CHILD':                         { en: 'Child',                 mr: 'मुल' },
  'INFANT':                        { en: 'Infant',                mr: 'बालक' },
  'DISABLED':                      { en: 'Disabled',              mr: 'दिव्यांग' },
}
