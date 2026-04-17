import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    children: [
      {
        path: '',
        name: 'MainDashboard',
        component: () => import('../views/agriculture/Dashboard.vue'),
      },
      {
        path: 'infrastructure/dashboard',
        name: 'InfrastructureDashboard',
        component: () => import('../views/agriculture/Dashboard.vue'),
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
        component: () => import('../views/agriculture/Dashboard.vue'),
      },
      {
        path: 'citizens',
        name: 'AgricultureCitizens',
        component: () => import('../components/UnifiedRegistry.vue'),
      },
      {
        path: 'registry',
        name: 'AgricultureRegistry',
        component: () => import('../components/UnifiedRegistry.vue'),
      },
      {
        path: 'farmers',
        redirect: '/agriculture/citizens',
      },
      {
        path: 'map',
        name: 'AgricultureMap',
        component: () => import('../views/agriculture/MapView.vue'),
      },
      {
        path: 'twin',
        name: 'AgricultureTwin',
        component: () => import('../views/agriculture/DigitalTwin.vue'),
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
        alias: ['/population/dashboard'],
        component: () => import('../views/population/PopulationDashboard.vue'),
      },
      {
        path: 'registry',
        name: 'PopulationRegistry',
        component: () => import('../components/UnifiedRegistry.vue'),
      },
      {
        path: 'map',
        name: 'PopulationMap',
        alias: ['/population/2d-map'],
        component: () => import('../views/population/PopulationMap.vue'),
      },
      {
        path: 'twin',
        name: 'PopulationTwin',
        alias: ['/population/3d-twin'],
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
