<template>
  <div class="app-shell" :data-theme="theme">
    <!-- Subtle topographic background -->
    <svg class="topo-bg" viewBox="0 0 1440 900" preserveAspectRatio="none">
      <defs>
        <pattern id="topo" x="0" y="0" width="200" height="200" patternUnits="userSpaceOnUse">
          <path d="M0 80 Q50 60 100 80 T200 80"  fill="none" stroke="rgba(232,168,56,0.04)" stroke-width="1"/>
          <path d="M0 120 Q50 100 100 120 T200 120" fill="none" stroke="rgba(232,168,56,0.03)" stroke-width="1"/>
          <path d="M0 160 Q50 140 100 160 T200 160" fill="none" stroke="rgba(45,212,191,0.03)" stroke-width="1"/>
          <circle cx="150" cy="50" r="1" fill="rgba(232,168,56,0.06)"/>
          <circle cx="30"  cy="140" r="0.8" fill="rgba(45,212,191,0.05)"/>
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#topo)"/>
    </svg>

    <!-- ── Thin top bar: hamburger + logo on left, logout on right ── -->
    <header class="top-bar">
      <div class="top-bar-left">
        <router-link to="/agriculture/dashboard" class="brand-link">
          <img :src="ivdpLogo" alt="IVDP logo" class="brand-logo" />
        </router-link>

        <!-- Hamburger toggle -->
        <button class="hamburger-btn" @click="toggleSidebar"
                :title="isSidebarOpen ? t('header.collapse') : t('header.expand')"
                :aria-expanded="isSidebarOpen" aria-controls="app-sidebar">
          <svg class="hamburger-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="3" y1="6"  x2="21" y2="6"/>
            <line x1="3" y1="12" x2="21" y2="12"/>
            <line x1="3" y1="18" x2="21" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="top-bar-right">
        <!-- Language switcher -->
        <div class="lang-switcher" ref="langSwitcherEl">
          <button class="lang-btn" @click="toggleLangMenu" :title="t('header.language')">
            <svg class="lang-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"/>
              <line x1="2" y1="12" x2="22" y2="12"/>
              <path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/>
            </svg>
            <span class="lang-current">{{ locale === 'mr' ? 'मराठी' : 'English' }}</span>
          </button>
          <div v-if="langMenuOpen" class="lang-menu">
            <button class="lang-option" :class="{ active: locale === 'en' }" @click="setLocale('en')">
              {{ t('lang.en') }}
            </button>
            <button class="lang-option" :class="{ active: locale === 'mr' }" @click="setLocale('mr')">
              {{ t('lang.mr') }}
            </button>
          </div>
        </div>

        <button class="logout-btn" @click="handleLogout" :title="t('header.signOut')">
          <svg class="logout-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4"/>
            <polyline points="16 17 21 12 16 7"/>
            <line x1="21" y1="12" x2="9" y2="12"/>
          </svg>
          <span>{{ t('header.logout') }}</span>
        </button>
      </div>
    </header>

    <!-- ── Body: sidebar + content ── -->
    <div class="body-wrap">

      <!-- ── Left sidebar ── -->
      <nav id="app-sidebar" class="sidebar" :class="{ collapsed: !isSidebarOpen }" aria-label="Primary navigation">
        <div class="sidebar-links">

          <router-link to="/agriculture/dashboard" class="nav-item"
            :class="{ active: $route.path === '/agriculture/dashboard' }"
            :title="t('nav.dashboard')">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="7" height="7" rx="1"/>
              <rect x="14" y="3" width="7" height="7" rx="1"/>
              <rect x="14" y="14" width="7" height="7" rx="1"/>
              <rect x="3"  y="14" width="7" height="7" rx="1"/>
            </svg>
            <span class="nav-label">{{ t('nav.dashboard') }}</span>
          </router-link>

          <router-link to="/agriculture/citizens" class="nav-item"
            :class="{ active: $route.path.startsWith('/agriculture/citizens') || $route.path.startsWith('/agriculture/registry') }"
            :title="t('nav.citizens')">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/>
              <circle cx="9" cy="7" r="4"/>
              <path d="M23 21v-2a4 4 0 00-3-3.87"/>
              <path d="M16 3.13a4 4 0 010 7.75"/>
            </svg>
            <span class="nav-label">{{ t('nav.citizens') }}</span>
          </router-link>

          <router-link to="/agriculture/map" class="nav-item"
            :class="{ active: $route.path.startsWith('/agriculture/map') }"
            :title="t('nav.map2d')">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <polygon points="1 6 1 22 8 18 16 22 23 18 23 2 16 6 8 2 1 6"/>
              <line x1="8"  y1="2"  x2="8"  y2="18"/>
              <line x1="16" y1="6"  x2="16" y2="22"/>
            </svg>
            <span class="nav-label">{{ t('nav.map2d') }}</span>
          </router-link>

          <router-link to="/agriculture/twin" class="nav-item"
            :class="{ active: $route.path.startsWith('/agriculture/twin') }"
            :title="t('nav.twin3d')">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z"/>
              <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
              <line x1="12" y1="22.08" x2="12" y2="12"/>
            </svg>
            <span class="nav-label">{{ t('nav.twin3d') }}</span>
          </router-link>

        </div>

        <!-- Sidebar footer: theme toggle + API status -->
        <div class="sidebar-footer">
          <button class="theme-toggle" @click="toggleTheme"
                  :title="theme === 'dark' ? t('header.lightMode') : t('header.darkMode')">
            <div class="theme-track" :class="{ light: theme === 'light' }">
              <div class="theme-thumb">
                <svg v-if="theme === 'light'" viewBox="0 0 20 20" fill="currentColor" class="theme-icon sun">
                  <path fill-rule="evenodd" d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z" clip-rule="evenodd"/>
                </svg>
                <svg v-else viewBox="0 0 20 20" fill="currentColor" class="theme-icon moon">
                  <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z"/>
                </svg>
              </div>
            </div>
            <span class="theme-label">{{ theme === 'dark' ? t('header.dark') : t('header.light') }}</span>
          </button>

          <div class="api-pill" role="status" aria-live="polite">
            <div class="system-dot" :class="apiStatus"></div>
            <span>{{ apiStatus === 'online' ? t('header.online') : t('header.offline') }}</span>
          </div>
        </div>
      </nav>

      <!-- ── Main content ── -->
      <main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="page" mode="out-in">
            <keep-alive include="UnifiedRegistry">
              <component :is="Component" />
            </keep-alive>
          </transition>
        </router-view>
      </main>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { logout } from '../../api/index.js'
import { STORAGE_KEY } from '../../i18n.js'
const ivdpLogo = 'https://ivdp2.mkcl.org/assets/logo-sm-CXFafxoa.webp'

const { t, locale } = useI18n()

const apiStatus      = ref('offline')
const theme          = ref(localStorage.getItem('agritwin-theme') || 'dark')
const isSidebarOpen  = ref(true)
const langMenuOpen   = ref(false)
const langSwitcherEl = ref(null)

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
}

function toggleSidebar() {
  isSidebarOpen.value = !isSidebarOpen.value
}

async function handleLogout() {
  await logout()
}

function setLocale(lang) {
  locale.value = lang
  localStorage.setItem(STORAGE_KEY, lang)
  langMenuOpen.value = false
}

function toggleLangMenu() {
  langMenuOpen.value = !langMenuOpen.value
}

function closeLangMenu() {
  langMenuOpen.value = false
}

function onDocClick(e) {
  if (langSwitcherEl.value && !langSwitcherEl.value.contains(e.target)) {
    langMenuOpen.value = false
  }
}

watch(theme, (val) => {
  localStorage.setItem('agritwin-theme', val)
  document.documentElement.setAttribute('data-theme', val)
})

onMounted(async () => {
  document.documentElement.setAttribute('data-theme', theme.value)
  document.addEventListener('click', onDocClick)
  try {
    const res = await fetch('/api/ping')
    if (res.ok) apiStatus.value = 'online'
  } catch {
    apiStatus.value = 'offline'
  }
})

onUnmounted(() => {
  document.removeEventListener('click', onDocClick)
})
</script>

<style>
/* ══════════════════════════════════════════════════════════════════════════════
   DESIGN TOKENS — Dark (default) & Light
   ══════════════════════════════════════════════════════════════════════════════ */
:root,
[data-theme="dark"] {
  --bg-deep:       #060b14;
  --bg-primary:    #0a1019;
  --bg-card:       #0f1722;
  --bg-card-hover: #141e2c;
  --bg-surface:    #182030;
  --border:        #1e293b;
  --border-light:  #273548;
  --text-primary:  #ffffff;
  --text-body:     #cbd5e1;
  --text-muted:    #94a3b8;
  --text-dim:      #64748b;
  --shadow:        rgba(0,0,0,0.4);
  --amber:         #e8a838;
  --amber-dim:     rgba(232,168,56,0.15);
  --teal:          #2dd4bf;
  --teal-dim:      rgba(45,212,191,0.15);
  --red:           #f87171;
  --red-dim:       rgba(248,113,113,0.15);
  --green:         #4ade80;
  --green-dim:     rgba(74,222,128,0.15);
}

[data-theme="light"] {
  --bg-deep:       #f0f4f8;
  --bg-primary:    #ffffff;
  --bg-card:       #ffffff;
  --bg-card-hover: #f8fafc;
  --bg-surface:    #f1f5f9;
  --border:        #e2e8f0;
  --border-light:  #cbd5e1;
  --text-primary:  #0f172a;
  --text-body:     #1e293b;
  --text-muted:    #475569;
  --text-dim:      #94a3b8;
  --shadow:        rgba(15,23,42,0.1);
  --amber:         #c47d10;
  --amber-dim:     rgba(196,125,16,0.12);
  --teal:          #0d9488;
  --teal-dim:      rgba(13,148,136,0.12);
  --red:           #dc2626;
  --red-dim:       rgba(220,38,38,0.1);
  --green:         #16a34a;
  --green-dim:     rgba(22,163,74,0.1);
}

:root, [data-theme="dark"], [data-theme="light"] {
  --slate-300: var(--text-body);
  --slate-400: var(--text-muted);
  --slate-500: var(--text-dim);
  --font-display: 'Instrument Serif', Georgia, serif;
  --font-body:    'Outfit', system-ui, sans-serif;
  /* Layout dimensions */
  --topbar-h:   64px;
  --sidebar-w: 220px;
  /* Keep legacy var so child pages that reference it get 0 extra padding */
  --top-nav-h: 0px;
}

/* ── Reset ── */
* { margin: 0; padding: 0; box-sizing: border-box; }

html, body, #app {
  height: 100%;
  overflow: hidden; /* scroll is handled per-region */
  background: var(--bg-deep);
  color: var(--text-body);
  font-family: var(--font-body);
  font-size: 14px;
  -webkit-font-smoothing: antialiased;
  transition: background 0.3s ease, color 0.3s ease;
}

/* ══════════════════════════════════════════════════════════════════════════════
   APP SHELL — full viewport, column layout
   ══════════════════════════════════════════════════════════════════════════════ */
.app-shell {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
}

.topo-bg {
  position: fixed;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 0;
}
[data-theme="light"] .topo-bg { opacity: 0.4; }

/* ══════════════════════════════════════════════════════════════════════════════
   TOP BAR — thin strip: logo left, logout right
   ══════════════════════════════════════════════════════════════════════════════ */
.top-bar {
  height: var(--topbar-h);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.2rem;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border);
  z-index: 50;
  position: relative;
  transition: background 0.3s ease, border-color 0.3s ease;
}

[data-theme="light"] .top-bar {
  box-shadow: 0 1px 8px var(--shadow);
}

/* Thin amber accent line at the bottom of the bar */
.top-bar::after {
  content: '';
  position: absolute;
  bottom: 0; left: 0; right: 0; height: 1px;
  background: linear-gradient(90deg,
    transparent 0%, var(--amber-dim) 30%,
    var(--teal-dim) 70%, transparent 100%);
}

/* Brand logo */
.brand-link {
  display: inline-flex;
  align-items: center;
  text-decoration: none;
}

.brand-logo {
  display: block;
  height: 50px;
  width: auto;
  max-height: calc(var(--topbar-h) - 10px);
  max-width: 100%;
  object-fit: contain;
}

@media (max-width: 768px) {
  .brand-logo {
    height: 42px;
    max-height: calc(var(--topbar-h) - 10px);
  }
}

/* Top-bar left group */
.top-bar-left {
  display: flex;
  align-items: center;
  gap: 0.55rem;
}

/* Hamburger button */
.hamburger-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 7px;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.18s, color 0.18s;
}
.hamburger-btn:hover {
  background: var(--bg-surface);
  color: var(--text-body);
}
.hamburger-icon {
  width: 18px;
  height: 18px;
}

/* Top-bar right group */
.top-bar-right {
  display: flex;
  align-items: center;
  gap: 0.55rem;
}

/* ── Language Switcher ── */
.lang-switcher {
  position: relative;
}

.lang-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: var(--bg-surface);
  border: 1px solid var(--border-light);
  border-radius: 7px;
  padding: 0.34rem 0.6rem;
  color: var(--text-muted);
  font-family: var(--font-body);
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.18s, border-color 0.18s, color 0.18s;
}
.lang-btn:hover {
  background: var(--bg-card-hover);
  border-color: var(--amber);
  color: var(--text-body);
}
.lang-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}
.lang-current {
  font-size: 0.72rem;
  letter-spacing: 0.04em;
}

.lang-menu {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  background: var(--bg-card);
  border: 1px solid var(--border-light);
  border-radius: 8px;
  box-shadow: 0 8px 24px var(--shadow);
  overflow: hidden;
  z-index: 200;
  min-width: 110px;
}

.lang-option {
  display: block;
  width: 100%;
  padding: 0.52rem 0.85rem;
  background: none;
  border: none;
  text-align: left;
  font-family: var(--font-body);
  font-size: 0.83rem;
  color: var(--text-muted);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}
.lang-option:hover {
  background: var(--bg-surface);
  color: var(--text-body);
}
.lang-option.active {
  color: var(--amber);
  font-weight: 600;
  background: var(--amber-dim);
}

/* Logout button (top-right) */
.logout-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.42rem;
  background: rgba(239, 68, 68, 0.10);
  border: 1px solid rgba(239, 68, 68, 0.28);
  border-radius: 7px;
  padding: 0.36rem 0.78rem;
  color: #fca5a5;
  font-family: var(--font-body);
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.18s, border-color 0.18s, color 0.18s, transform 0.1s;
}
.logout-btn:hover {
  background: rgba(239, 68, 68, 0.22);
  border-color: rgba(239, 68, 68, 0.55);
  color: #ffffff;
  transform: translateY(-1px);
}
.logout-btn:active { transform: translateY(0); }
.logout-icon { width: 14px; height: 14px; flex-shrink: 0; }

/* ══════════════════════════════════════════════════════════════════════════════
   BODY WRAP — sidebar + main content side by side
   ══════════════════════════════════════════════════════════════════════════════ */
.body-wrap {
  flex: 1;
  display: flex;
  overflow: hidden; /* prevents double scrollbar */
  position: relative;
  z-index: 1;
}

/* ══════════════════════════════════════════════════════════════════════════════
   SIDEBAR
   ══════════════════════════════════════════════════════════════════════════════ */
.sidebar {
  width: var(--sidebar-w);
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--bg-card);
  border-right: 1px solid var(--border);
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-width: none;
  /* ↓ Smooth open/close transition */
  transition: width 0.3s ease-in-out,
              background 0.3s ease,
              border-color 0.3s ease;
}
.sidebar::-webkit-scrollbar { display: none; }

/* ── Collapsed sidebar: 56px — icons only ── */
.sidebar.collapsed { width: 56px; }

.sidebar-links {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 1rem 0.6rem;
}
.sidebar.collapsed .sidebar-links {
  padding: 1rem 0.35rem;
  align-items: center;
}

/* Each nav item */
.nav-item {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.62rem 0.75rem;
  border-radius: 8px;
  text-decoration: none;
  color: var(--text-muted);
  font-size: 0.88rem;
  font-weight: 500;
  border-left: 3px solid transparent;
  transition: background 0.18s, color 0.18s, border-color 0.18s,
              padding 0.3s ease-in-out;
  position: relative;
  white-space: nowrap;
}

.nav-item:hover {
  background: var(--bg-surface);
  color: var(--text-body);
}

.nav-item.active {
  background: var(--amber-dim);
  color: var(--amber);
  border-left-color: var(--amber);
  font-weight: 600;
}

/* Collapsed nav-item: center icon, no left border visual clutter */
.sidebar.collapsed .nav-item {
  justify-content: center;
  padding: 0.62rem;
  gap: 0;
  width: 40px;
}

.nav-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  opacity: 0.8;
  transition: opacity 0.18s;
}

.nav-item.active .nav-icon,
.nav-item:hover  .nav-icon { opacity: 1; }

/* Nav label — fades & collapses width smoothly */
.nav-label {
  line-height: 1;
  opacity: 1;
  max-width: 160px;
  overflow: hidden;
  transition: opacity 0.2s ease-in-out, max-width 0.3s ease-in-out;
}
.sidebar.collapsed .nav-label {
  opacity: 0;
  max-width: 0;
}

/* Sidebar footer */
.sidebar-footer {
  padding: 0.75rem 0.8rem;
  border-top: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  transition: padding 0.3s ease-in-out;
}
.sidebar.collapsed .sidebar-footer {
  padding: 0.75rem 0.35rem;
  align-items: center;
}

.theme-toggle {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.38rem 0.4rem;
  border-radius: 7px;
  background: none;
  border: none;
  cursor: pointer;
  font-family: var(--font-body);
  font-size: 0.78rem;
  color: var(--text-muted);
  white-space: nowrap;
  width: 100%;
  transition: background 0.2s, color 0.2s;
}
.theme-toggle:hover {
  background: var(--bg-surface);
  color: var(--text-body);
}
/* Collapsed: only the toggle switch, label hidden */
.sidebar.collapsed .theme-toggle {
  justify-content: center;
  width: auto;
  padding: 0.38rem;
  gap: 0;
}

.theme-track {
  width: 32px; height: 18px;
  border-radius: 9px;
  background: var(--border-light);
  position: relative;
  flex-shrink: 0;
  transition: background 0.3s ease;
}
.theme-track.light { background: var(--amber); }

.theme-thumb {
  position: absolute;
  top: 2px; left: 2px;
  width: 14px; height: 14px;
  border-radius: 50%;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.3s cubic-bezier(0.4,0,0.2,1);
  box-shadow: 0 1px 4px rgba(0,0,0,0.3);
}
.theme-track.light .theme-thumb { transform: translateX(14px); }

.theme-icon { width: 9px; height: 9px; }
.theme-icon.sun  { color: var(--amber); }
.theme-icon.moon { color: #6366f1; }
.theme-label {
  font-size: 0.78rem;
  opacity: 1;
  max-width: 120px;
  overflow: hidden;
  transition: opacity 0.2s ease-in-out, max-width 0.3s ease-in-out;
}
.sidebar.collapsed .theme-label {
  opacity: 0;
  max-width: 0;
}

.api-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.42rem;
  border: 1px solid var(--border);
  border-radius: 999px;
  padding: 0.28rem 0.6rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  background: var(--bg-surface);
  width: fit-content;
  transition: padding 0.3s ease-in-out, border-radius 0.3s ease-in-out;
}
/* Collapsed: pill becomes a small dot indicator */
.sidebar.collapsed .api-pill {
  padding: 0.28rem;
  border-radius: 50%;
  gap: 0;
}
.api-pill span {
  opacity: 1;
  max-width: 60px;
  overflow: hidden;
  white-space: nowrap;
  transition: opacity 0.2s ease-in-out, max-width 0.3s ease-in-out;
}
.sidebar.collapsed .api-pill span {
  opacity: 0;
  max-width: 0;
}

.system-dot {
  width: 6px; height: 6px;
  border-radius: 50%;
  background: var(--red);
  flex-shrink: 0;
}
.system-dot.online {
  background: var(--green);
  box-shadow: 0 0 5px var(--green);
}

/* ══════════════════════════════════════════════════════════════════════════════
   MAIN CONTENT — fills remaining width, scrolls independently
   ══════════════════════════════════════════════════════════════════════════════ */
.main-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  background: var(--bg-deep);
  transition: background 0.3s ease;
  position: relative;
  z-index: 1;
}

/* ── Page transitions ── */
.page-enter-active { transition: all 0.22s ease-out; }
.page-leave-active { transition: all 0.14s ease-in; }
.page-enter-from   { opacity: 0; transform: translateY(6px); }
.page-leave-to     { opacity: 0; transform: translateY(-3px); }

/* ── Scrollbar ── */
::-webkit-scrollbar { width: 5px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: var(--border-light); border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: var(--text-dim); }

/* ── Shared card component ── */
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.25rem;
  transition: border-color 0.2s, box-shadow 0.2s, background 0.3s;
}
.card:hover {
  border-color: var(--border-light);
  box-shadow: 0 4px 24px var(--shadow);
}
.card-title {
  font-family: var(--font-display);
  font-size: 1.1rem;
  color: var(--text-primary);
  margin-bottom: 1rem;
}

/* ── Spinner ── */
.spinner {
  width: 32px; height: 32px;
  border: 3px solid var(--border);
  border-top-color: var(--amber);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 3rem;
  color: var(--text-dim);
  font-size: 0.85rem;
}

/* ── Mobile: sidebar goes to bottom ── */
@media (max-width: 768px) {
  :root, [data-theme="dark"], [data-theme="light"] {
    --sidebar-w: 0px;
  }

  .body-wrap { flex-direction: column; }

  .sidebar {
    width: 100%;
    height: auto;
    flex-direction: row;
    border-right: none;
    border-top: 1px solid var(--border);
    order: 2;
    overflow-x: auto;
    overflow-y: hidden;
  }

  .sidebar-links {
    flex-direction: row;
    flex: 1;
    padding: 0.4rem 0.5rem;
    gap: 0;
  }

  .nav-item {
    flex-direction: column;
    gap: 0.25rem;
    padding: 0.5rem 0.7rem;
    border-left: none;
    border-bottom: 3px solid transparent;
    border-radius: 6px;
    font-size: 0.72rem;
  }

  .nav-item.active {
    border-bottom-color: var(--amber);
    border-left-color: transparent;
  }

  .sidebar-footer { display: none; }

  .main-content { order: 1; }
}
</style>
