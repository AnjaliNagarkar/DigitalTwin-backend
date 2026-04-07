import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/', name: 'Dashboard', component: () => import('../views/Dashboard.vue') },
  { path: '/farmers', name: 'Farmers', component: () => import('../views/Farmers.vue') },
  { path: '/map', name: 'Map', component: () => import('../views/MapView.vue') },
  { path: '/twin', name: 'DigitalTwin', component: () => import('../views/DigitalTwin.vue') },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
