import { createI18n } from 'vue-i18n'
import en from './locales/en.json'
import mr from './locales/mr.json'

const STORAGE_KEY = 'agritwin-locale'

function getStoredLocale() {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'en' || stored === 'mr') return stored
  } catch {}
  return 'en'
}

const i18n = createI18n({
  legacy: false,
  locale: getStoredLocale(),
  fallbackLocale: 'en',
  messages: { en, mr },
})

export default i18n
export { STORAGE_KEY }
