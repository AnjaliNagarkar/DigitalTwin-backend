import { createRouter, createWebHistory } from 'vue-router'
import { verifySSOToken } from '../api/index.js'

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
router.beforeEach(async (to, _from, next) => {

  // ── 1. SSO: handle ?token=<ivdp_token> present in the URL ─────────────────
  // Works on any path — the IVDP portal can deep-link to any page and append
  // its bearer token as a query param.  We verify it once, store our own
  // session, then redirect to a clean URL.
  const ssoToken = typeof to.query.token === 'string' ? to.query.token.trim() : ''

  if (ssoToken) {
    try {
      // Forward to backend → backend validates with IVDP → returns our session
      const data = await verifySSOToken(ssoToken)

      // Store session using the same keys as the normal login flow
      localStorage.setItem('auth_token',    data.token)
      localStorage.setItem('auth_username', data.username)
      localStorage.setItem('auth_expires',  data.expires_at)

      // Build a clean destination: same path the user tried, but without the
      // IVDP token in the query string (keeps the address bar tidy).
      // eslint-disable-next-line no-unused-vars
      const { token: _drop, ...cleanQuery } = to.query
      const destination = {
        path:    to.path === '/' ? '/agriculture/dashboard' : to.path,
        query:   cleanQuery,
        replace: true, // replace the history entry so Back doesn't re-trigger SSO
      }

      return next(destination)

    } catch (err) {
      // IVDP rejected the token (expired, tampered, etc.)
      console.warn('[SSO] token verification failed:', err.message)
      localStorage.removeItem('auth_token')
      localStorage.removeItem('auth_username')
      localStorage.removeItem('auth_expires')

      // Redirect to login and carry a flag so LoginView shows the right message
      return next({ path: '/login', query: { error: 'sso_failed' }, replace: true })
    }
  }

  // ── 2. Normal session check ────────────────────────────────────────────────
  const token     = localStorage.getItem('auth_token')
  const expiresAt = localStorage.getItem('auth_expires')
  const isExpired = expiresAt ? new Date() > new Date(expiresAt) : true
  const isAuthenticated = !!token && !isExpired

  if (to.meta?.public) {
    // Already logged in — skip the login page
    if (isAuthenticated) return next('/agriculture/dashboard')
    return next()
  }

  // Protected route — must have a valid session
  if (!isAuthenticated) {
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_username')
    localStorage.removeItem('auth_expires')
    return next('/login')
  }

  next()
})

export default router
