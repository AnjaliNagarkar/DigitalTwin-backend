import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    children: [
      {
        path: '',
        name: 'MainDashboard',
        component: () => import('../views/Dashboard.vue'),
      },
      {
        path: 'infrastructure/dashboard',
        name: 'InfrastructureDashboard',
        component: () => import('../views/Dashboard.vue'),
      },
    ],
  },
  {
    path: '/agriculture',
    component: () => import('../views/agriculture/AgricultureLayout.vue'),
    children: [
      {
        path: '',
        redirect: '/agriculture/dashboard',
      },
      {
        path: 'dashboard',
        name: 'AgricultureDashboard',
        component: () => import('../views/Dashboard.vue'),
      },
      {
        path: 'farmers',
        name: 'AgricultureFarmers',
        component: () => import('../views/Farmers.vue'),
      },
      {
        path: 'map',
        name: 'AgricultureMap',
        component: () => import('../views/MapView.vue'),
      },
      {
        path: 'twin',
        name: 'AgricultureTwin',
        component: () => import('../views/DigitalTwin.vue'),
      },
    ],
  },
  {
    path: '/population',
    component: () => import('../views/population/PopulationLayout.vue'),
    children: [
      {
        path: '',
        redirect: '/population/dashboard',
      },
      {
        path: 'dashboard',
        name: 'PopulationDashboard',
        component: () => import('../views/population/PopulationDashboard.vue'),
      },
      {
        path: 'registry',
        name: 'PopulationRegistry',
        component: () => import('../views/population/PopulationRegistry.vue'),
      },
      {
        path: 'map',
        name: 'PopulationMap',
        component: () => import('../views/population/PopulationMap.vue'),
      },
      {
        path: 'twin',
        name: 'PopulationTwin',
        component: () => import('../views/population/PopulationTwin.vue'),
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
