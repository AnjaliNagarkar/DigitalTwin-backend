<template>
  <div class="app-shell" :data-theme="theme">
    <svg class="topo-bg" viewBox="0 0 1440 900" preserveAspectRatio="none">
      <defs>
        <pattern id="topo-pop" x="0" y="0" width="200" height="200" patternUnits="userSpaceOnUse">
          <path d="M0 80 Q50 60 100 80 T200 80" fill="none" stroke="rgba(232,168,56,0.04)" stroke-width="1"/>
          <path d="M0 120 Q50 100 100 120 T200 120" fill="none" stroke="rgba(232,168,56,0.03)" stroke-width="1"/>
          <path d="M0 160 Q50 140 100 160 T200 160" fill="none" stroke="rgba(45,212,191,0.03)" stroke-width="1"/>
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#topo-pop)"/>
    </svg>

    <header class="top-nav">
      <div class="nav-left">
        <router-link to="/population/dashboard" class="brand-link">
          <div class="brand-dot"></div>
          <div class="brand-text">
            <span class="brand-name">PopTwin</span>
            <span class="brand-sub">Digital Twin Platform</span>
          </div>
        </router-link>
      </div>

      <nav class="nav-center" aria-label="Population navigation">
        <router-link to="/population/dashboard" class="nav-link" :class="{ active: $route.path.startsWith('/population/dashboard') }">Dashboard</router-link>
        <router-link to="/population-registry" class="nav-link" :class="{ active: $route.path.startsWith('/population/registry') || $route.path.startsWith('/population-registry') }">Citizens</router-link>
        <router-link to="/population/2d-map" class="nav-link" :class="{ active: $route.path.startsWith('/population/map') || $route.path.startsWith('/population/2d-map') }">2D Map</router-link>
        <router-link to="/population/3d-twin" class="nav-link" :class="{ active: $route.path.startsWith('/population/twin') || $route.path.startsWith('/population/3d-twin') }">3D Twin</router-link>
      </nav>

      <div class="nav-right">
        <button class="theme-toggle" @click="toggleTheme" :title="theme === 'dark' ? 'Switch to Light Mode' : 'Switch to Dark Mode'">
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
          <span class="theme-label">{{ theme === 'dark' ? 'Dark Mode' : 'Light Mode' }}</span>
        </button>

        <div class="api-pill" role="status" aria-live="polite">
          <div class="system-dot" :class="apiStatus"></div>
          <span>API {{ apiStatus === 'online' ? 'Connected' : 'Offline' }}</span>
        </div>

        <button class="logout-btn" @click="handleLogout" title="Sign out">
          <!-- Power-off / logout icon -->
          <svg class="logout-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4"/>
            <polyline points="16 17 21 12 16 7"/>
            <line x1="21" y1="12" x2="9" y2="12"/>
          </svg>
          <span>Logout</span>
        </button>
      </div>
    </header>

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
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { logout } from '../../api/index.js'

const apiStatus = ref('offline')
const theme     = ref(localStorage.getItem('agritwin-theme') || 'dark')

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
}

async function handleLogout() {
  await logout()
}

watch(theme, (val) => {
  localStorage.setItem('agritwin-theme', val)
  document.documentElement.setAttribute('data-theme', val)
})

onMounted(async () => {
  document.documentElement.setAttribute('data-theme', theme.value)
  try {
    const res = await fetch('/api/ping')
    if (res.ok) apiStatus.value = 'online'
  } catch {
    apiStatus.value = 'offline'
  }
})
</script>

<style>
/* ── Dark theme (default) ── */
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

/* ── Light theme ── */
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

/* Shared aliases and typography tokens used across population pages */
:root, [data-theme="dark"], [data-theme="light"] {
  --slate-300: var(--text-body);
  --slate-400: var(--text-muted);
  --slate-500: var(--text-dim);
  --font-display: 'Instrument Serif', Georgia, serif;
  --font-body: 'Outfit', system-ui, sans-serif;
  --top-nav-h: 72px;
}

* { margin: 0; padding: 0; box-sizing: border-box; }

html, body {
  height: 100%;
  background: var(--bg-deep);
  color: var(--text-body);
  font-family: var(--font-body);
  font-size: 14px;
  -webkit-font-smoothing: antialiased;
  transition: background 0.3s ease, color 0.3s ease;
}

#app { height: 100%; }

.app-shell {
  min-height: 100vh;
  overflow-x: hidden;
  position: relative;
}

.topo-bg {
  position: fixed;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 0;
  opacity: 0.9;
}

.top-nav {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: var(--top-nav-h);
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.75rem 1.4rem;
  z-index: 40;
}

.nav-left,
.nav-right {
  display: flex;
  align-items: center;
  flex: 1;
}

.nav-right {
  justify-content: flex-end;
  gap: 0.6rem;
}

.brand-link {
  display: inline-flex;
  align-items: center;
  gap: 0.7rem;
  text-decoration: none;
}

.brand-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--teal);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--teal) 25%, transparent);
}

.brand-text {
  display: flex;
  flex-direction: column;
}

.brand-name {
  font-family: var(--font-display);
  font-size: 1.3rem;
  color: var(--text-primary);
  line-height: 1.1;
}

.brand-sub {
  font-size: 0.62rem;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.12em;
  font-weight: 600;
}

.nav-center {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  flex: 1.3;
}

.nav-link {
  padding: 0.6rem 0.9rem;
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.86rem;
  font-weight: 500;
  border-bottom: 2px solid transparent;
  transition: color 0.2s ease, border-color 0.2s ease;
}

.nav-link:hover {
  color: var(--text-body);
}

.nav-link.active {
  color: var(--text-primary);
  border-bottom-color: var(--amber);
}

.theme-toggle {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.45rem 0.6rem;
  border-radius: 8px;
  background: none;
  border: none;
  cursor: pointer;
  font-family: var(--font-body);
  font-size: 0.82rem;
  color: var(--text-muted);
  white-space: nowrap;
  transition: background 0.2s, color 0.2s;
}

.theme-toggle:hover {
  background: var(--bg-surface);
  color: var(--text-body);
}

.theme-track {
  width: 36px;
  height: 20px;
  border-radius: 10px;
  background: var(--border-light);
  position: relative;
  flex-shrink: 0;
  transition: background 0.3s ease;
}

.theme-track.light {
  background: var(--amber);
}

.theme-thumb {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 1px 4px rgba(0,0,0,0.3);
}

.theme-track.light .theme-thumb {
  transform: translateX(16px);
}

.theme-icon {
  width: 10px;
  height: 10px;
}

.theme-icon.sun {
  color: var(--amber);
}

.theme-icon.moon {
  color: #6366f1;
}

.theme-label {
  font-size: 0.82rem;
}

.api-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.42rem;
  border: 1px solid var(--border);
  background: var(--bg-surface);
  border-radius: 999px;
  padding: 0.3rem 0.65rem;
  font-size: 0.72rem;
  color: var(--text-muted);
}

.system-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--red);
}

.system-dot.online {
  background: var(--green);
}

/* ── Logout button ── */
.logout-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  background: rgba(239, 68, 68, 0.10);
  border: 1px solid rgba(239, 68, 68, 0.30);
  border-radius: 8px;
  padding: 0.42rem 0.85rem;
  color: #fca5a5;
  font-size: 0.82rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.18s ease, border-color 0.18s ease,
              color 0.18s ease, transform 0.1s ease;
}
.logout-btn:hover {
  background: rgba(239, 68, 68, 0.22);
  border-color: rgba(239, 68, 68, 0.55);
  color: #ffffff;
  transform: translateY(-1px);
}
.logout-btn:active {
  transform: translateY(0);
}
.logout-icon {
  width: 15px;
  height: 15px;
  flex-shrink: 0;
}

.main-content {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  padding-top: var(--top-nav-h);
}

.page-enter-active,
.page-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.page-enter-from,
.page-leave-to {
  opacity: 0;
  transform: translateY(4px);
}
</style>
