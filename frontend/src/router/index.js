import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  // ── Public ─────────────────────────────────────────────────────────────────
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/LoginView.vue'),
    meta: { public: true },
  },

  // ── Protected ──────────────────────────────────────────────────────────────
  {
    path: '/',
    redirect: '/agriculture/dashboard',
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
    redirect: '/agriculture/dashboard',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// ── Navigation guard ──────────────────────────────────────────────────────────
// Every route that is not marked { meta: { public: true } } requires a valid
// session token in localStorage.  If the token is missing or the session has
// expired the user is redirected to /login.
router.beforeEach((to, _from, next) => {
  const token     = localStorage.getItem('auth_token')
  const expiresAt = localStorage.getItem('auth_expires')
  const isExpired = expiresAt ? new Date() > new Date(expiresAt) : true
  const isAuthenticated = !!token && !isExpired

  if (to.meta?.public) {
    // Already logged in — skip the login page
    if (isAuthenticated) return next('/agriculture/dashboard')
    return next()
  }

  // Protected route
  if (!isAuthenticated) {
    // Clear stale tokens
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_username')
    localStorage.removeItem('auth_expires')
    return next('/login')
  }

  next()
})

export default router
