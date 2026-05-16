import { createApp } from 'vue'
import App from './App.vue'
import router from './router/index.js'
import VueApexCharts from 'vue3-apexcharts'
import i18n from './i18n.js'

const app = createApp(App)
app.use(router)
app.use(VueApexCharts)
app.use(i18n)
app.mount('#app')
