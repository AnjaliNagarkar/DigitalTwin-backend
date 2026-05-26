import { defineComponent, provide, h, getCurrentInstance } from 'vue'
import { I18nInjectionKey } from 'vue-i18n'
import i18n from '../i18n.js'

export function withI18n(component) {
  return defineComponent({
    name: `Federated_${component.__name ?? 'Component'}`,
    setup(_, { attrs, slots }) {
      // vue-i18n 9.x useI18n() does NOT inject via the exported I18nInjectionKey.
      // It reads instance.appContext.app.__VUE_I18N_SYMBOL__ and injects via that
      // private symbol (set by app.use(i18n)). Inside IVDP that symbol points to
      // IVDP's legacy-mode i18n → "Not available in legacy mode" crash.
      //
      // Fix: read the real symbol from the host app and re-provide AgriTwin's
      // composition-mode i18n under that same key so inject() in Dashboard (and
      // every child) resolves to AgriTwin's instance, not IVDP's.
      const instance = getCurrentInstance()
      const hostSymbol = instance?.appContext?.app?.__VUE_I18N_SYMBOL__
      if (hostSymbol) {
        provide(hostSymbol, i18n)
      }
      // Fallback: custom elements use I18nInjectionKey directly.
      provide(I18nInjectionKey, i18n)
      return () => h(component, attrs, slots)
    },
  })
}
