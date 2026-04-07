<template>
  <div class="app-shell" :data-theme="theme">
    <!-- Topographic background pattern -->
    <svg class="topo-bg" viewBox="0 0 1440 900" preserveAspectRatio="none">
      <defs>
        <pattern id="topo" x="0" y="0" width="200" height="200" patternUnits="userSpaceOnUse">
          <path d="M0 80 Q50 60 100 80 T200 80" fill="none" stroke="rgba(232,168,56,0.04)" stroke-width="1"/>
          <path d="M0 120 Q50 100 100 120 T200 120" fill="none" stroke="rgba(232,168,56,0.03)" stroke-width="1"/>
          <path d="M0 160 Q50 140 100 160 T200 160" fill="none" stroke="rgba(45,212,191,0.03)" stroke-width="1"/>
          <circle cx="150" cy="50" r="1" fill="rgba(232,168,56,0.06)"/>
          <circle cx="30" cy="140" r="0.8" fill="rgba(45,212,191,0.05)"/>
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#topo)"/>
    </svg>

    <nav class="sidebar">
      <div class="sidebar-brand">
        <div class="brand-icon">
          <svg viewBox="0 0 32 32" fill="none">
            <rect x="2" y="14" width="6" height="16" rx="1" fill="var(--amber)"/>
            <rect x="10" y="8" width="6" height="22" rx="1" fill="var(--teal)"/>
            <rect x="18" y="4" width="6" height="26" rx="1" fill="var(--amber)" opacity="0.7"/>
            <rect x="26" y="10" width="4" height="20" rx="1" fill="var(--teal)" opacity="0.6"/>
            <line x1="0" y1="30" x2="32" y2="30" stroke="var(--slate-400)" stroke-width="1"/>
          </svg>
        </div>
        <div class="brand-text">
          <span class="brand-name">AgriTwin</span>
          <span class="brand-sub">Digital Twin Platform</span>
        </div>
      </div>

      <div class="nav-section-label">Navigation</div>
      <router-link to="/" class="nav-item" :class="{ active: $route.path === '/' }">
        <svg viewBox="0 0 20 20" fill="currentColor"><path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 6a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zm10 0a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z"/></svg>
        <span>Dashboard</span>
      </router-link>
      <router-link to="/farmers" class="nav-item" :class="{ active: $route.path === '/farmers' }">
        <svg viewBox="0 0 20 20" fill="currentColor"><path d="M9 6a3 3 0 11-6 0 3 3 0 016 0zm8 0a3 3 0 11-6 0 3 3 0 016 0zm-4.07 11c.046-.327.07-.66.07-1a6.97 6.97 0 00-1.5-4.33A5 5 0 0119 16v1h-6.07zM6 11a5 5 0 015 5v1H1v-1a5 5 0 015-5z"/></svg>
        <span>Farmers</span>
      </router-link>
      <router-link to="/map" class="nav-item" :class="{ active: $route.path === '/map' }">
        <svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M12 1.586l-4 4v12.828l4-4V1.586zM3.707 3.293A1 1 0 002 4v10a1 1 0 00.293.707L6 18.414V5.586L3.707 3.293zM14 5.586v12.828l2.293-2.293A1 1 0 0017 16V6a1 1 0 00-.293-.707L14 2.586v3z" clip-rule="evenodd"/></svg>
        <span>2D Map</span>
      </router-link>
      <router-link to="/twin" class="nav-item" :class="{ active: $route.path === '/twin' }">
        <svg viewBox="0 0 20 20" fill="currentColor"><path d="M10.394 2.08a1 1 0 00-.788 0l-7 3a1 1 0 000 1.84L5.25 8.051a.999.999 0 01.356-.257l4-1.714a1 1 0 11.788 1.838l-3.14 1.346 2.352 1.005a1 1 0 00.788 0l7-3a1 1 0 000-1.838l-7-3.001z"/><path d="M3.31 9.397L5 10.12v4.102a8.969 8.969 0 00-1.05-.174 1 1 0 01-.89-.89 11.115 11.115 0 01.25-3.762zM9.3 16.573A9.026 9.026 0 007 14.935v-3.957l1.818.78a3 3 0 002.364 0l5.508-2.361a11.026 11.026 0 01.25 3.762 1 1 0 01-.89.89 8.968 8.968 0 00-5.35 2.524 1 1 0 01-1.4 0z"/></svg>
        <span>3D Twin</span>
      </router-link>

      <div class="nav-spacer"></div>

      <div class="nav-section-label">System</div>

      <!-- Theme Toggle -->
      <button class="theme-toggle" @click="toggleTheme" :title="theme === 'dark' ? 'Switch to Light Mode' : 'Switch to Dark Mode'">
        <div class="theme-track" :class="{ light: theme === 'light' }">
          <div class="theme-thumb">
            <!-- Sun icon (light) -->
            <svg v-if="theme === 'light'" viewBox="0 0 20 20" fill="currentColor" class="theme-icon sun">
              <path fill-rule="evenodd" d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z" clip-rule="evenodd"/>
            </svg>
            <!-- Moon icon (dark) -->
            <svg v-else viewBox="0 0 20 20" fill="currentColor" class="theme-icon moon">
              <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z"/>
            </svg>
          </div>
        </div>
        <span class="theme-label">{{ theme === 'dark' ? 'Dark Mode' : 'Light Mode' }}</span>
      </button>

      <div class="nav-item system-info">
        <div class="system-dot" :class="apiStatus"></div>
        <span>API: {{ apiStatus === 'online' ? 'Connected' : 'Offline' }}</span>
      </div>
    </nav>

    <main class="main-content">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'

const apiStatus = ref('offline')
const theme = ref(localStorage.getItem('agritwin-theme') || 'dark')

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
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

/* Alias old var names to new semantic names */
:root, [data-theme="dark"], [data-theme="light"] {
  --slate-300: var(--text-body);
  --slate-400: var(--text-muted);
  --slate-500: var(--text-dim);
  --font-display: 'Instrument Serif', Georgia, serif;
  --font-body: 'Outfit', system-ui, sans-serif;
  --sidebar-w: 240px;
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
  display: flex;
  height: 100vh;
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
  opacity: 1;
}

[data-theme="light"] .topo-bg { opacity: 0.4; }

/* ── Sidebar ── */
.sidebar {
  width: var(--sidebar-w);
  min-width: var(--sidebar-w);
  background: var(--bg-primary);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 1.25rem 0.75rem;
  z-index: 10;
  position: relative;
  transition: background 0.3s ease, border-color 0.3s ease;
}

[data-theme="light"] .sidebar {
  box-shadow: 2px 0 12px var(--shadow);
}

.sidebar::after {
  content: '';
  position: absolute;
  right: 0; top: 0; bottom: 0;
  width: 1px;
  background: linear-gradient(180deg, var(--amber-dim), transparent 30%, transparent 70%, var(--teal-dim));
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0 0.5rem;
  margin-bottom: 2rem;
}

.brand-icon { width: 36px; height: 36px; flex-shrink: 0; }
.brand-icon svg { width: 100%; height: 100%; }

.brand-text { display: flex; flex-direction: column; }

.brand-name {
  font-family: var(--font-display);
  font-size: 1.35rem;
  color: var(--text-primary);
  line-height: 1.1;
}

.brand-sub {
  font-size: 0.65rem;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.12em;
  font-weight: 500;
}

.nav-section-label {
  font-size: 0.6rem;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  color: var(--text-dim);
  padding: 0 0.75rem;
  margin-bottom: 0.5rem;
  font-weight: 600;
}

.nav-spacer { flex: 1; }

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.6rem 0.75rem;
  border-radius: 8px;
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 400;
  transition: all 0.2s ease;
  margin-bottom: 2px;
  cursor: pointer;
  position: relative;
}

.nav-item svg { width: 18px; height: 18px; flex-shrink: 0; opacity: 0.6; transition: opacity 0.2s; }
.nav-item:hover { background: var(--bg-surface); color: var(--text-body); }
.nav-item:hover svg { opacity: 0.9; }

.nav-item.active {
  background: var(--amber-dim);
  color: var(--amber);
  font-weight: 500;
}
.nav-item.active svg { opacity: 1; color: var(--amber); }
.nav-item.active::before {
  content: '';
  position: absolute;
  left: 0; top: 50%;
  transform: translateY(-50%);
  width: 3px; height: 60%;
  background: var(--amber);
  border-radius: 0 3px 3px 0;
}

/* ── Theme Toggle ── */
.theme-toggle {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.6rem 0.75rem;
  border-radius: 8px;
  background: none;
  border: none;
  cursor: pointer;
  font-family: var(--font-body);
  font-size: 0.82rem;
  color: var(--text-muted);
  width: 100%;
  text-align: left;
  margin-bottom: 4px;
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

.theme-icon { width: 10px; height: 10px; }
.theme-icon.sun { color: var(--amber); }
.theme-icon.moon { color: #6366f1; }

.theme-label { font-size: 0.82rem; }

/* ── System info ── */
.system-info { cursor: default; font-size: 0.75rem; }

.system-dot {
  width: 7px; height: 7px;
  border-radius: 50%;
  background: var(--red);
  flex-shrink: 0;
}
.system-dot.online {
  background: var(--green);
  box-shadow: 0 0 6px var(--green);
}

/* ── Main Content ── */
.main-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  z-index: 1;
  position: relative;
  background: var(--bg-deep);
  transition: background 0.3s ease;
}

/* ── Page transitions ── */
.page-enter-active { transition: all 0.25s ease-out; }
.page-leave-active { transition: all 0.15s ease-in; }
.page-enter-from { opacity: 0; transform: translateY(8px); }
.page-leave-to { opacity: 0; transform: translateY(-4px); }

/* ── Scrollbar ── */
::-webkit-scrollbar { width: 6px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: var(--border-light); border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: var(--text-dim); }

/* ── Shared cards ── */
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
</style>
