<template>
  <div class="population-shell">
    <aside class="sidebar">
      <div class="sidebar-brand">
        <div class="brand-icon">P</div>
        <div class="brand-text">
          <span class="brand-name">Population Twin</span>
          <span class="brand-sub">Demography Module</span>
        </div>
      </div>

      <div class="nav-section-label">Navigation</div>
      <router-link to="/population/dashboard" class="nav-item" :class="{ active: $route.path === '/population/dashboard' }">
        Dashboard
      </router-link>
      <router-link to="/population/registry" class="nav-item" :class="{ active: $route.path === '/population/registry' }">
        Population Registry
      </router-link>
      <router-link to="/population/map" class="nav-item" :class="{ active: $route.path === '/population/map' }">
        2D Map
      </router-link>
      <router-link to="/population/twin" class="nav-item" :class="{ active: $route.path === '/population/twin' }">
        3D Twin
      </router-link>
    </aside>

    <main class="main-content">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<style scoped>
.population-shell {
  min-height: 100vh;
  display: flex;
  background: #eef6f0;
  color: #15261f;
  font-family: 'Outfit', sans-serif;
}

.sidebar {
  width: 240px;
  min-width: 240px;
  background: linear-gradient(180deg, #11281f 0%, #19382d 100%);
  border-right: 1px solid rgba(255, 255, 255, 0.08);
  padding: 1.2rem 0.75rem;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  margin-bottom: 1.35rem;
  padding: 0 0.45rem;
}

.brand-icon {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  display: grid;
  place-items: center;
  background: #e3b042;
  color: #1f1504;
  font-weight: 700;
}

.brand-text {
  display: flex;
  flex-direction: column;
}

.brand-name {
  font-size: 1rem;
  color: #ffffff;
  font-weight: 600;
}

.brand-sub {
  font-size: 0.65rem;
  color: #90b7a7;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.nav-section-label {
  font-size: 0.62rem;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  color: #90b7a7;
  margin-bottom: 0.45rem;
  padding: 0 0.6rem;
}

.nav-item {
  display: block;
  margin-bottom: 0.2rem;
  border-radius: 8px;
  text-decoration: none;
  color: #d6e8df;
  padding: 0.6rem 0.7rem;
  font-size: 0.86rem;
  transition: background 0.2s ease, color 0.2s ease;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.nav-item.active {
  background: #e3b042;
  color: #1c1402;
  font-weight: 600;
}

.main-content {
  flex: 1;
  min-width: 0;
  overflow: auto;
}

.page-enter-active,
.page-leave-active {
  transition: opacity 0.16s ease;
}

.page-enter-from,
.page-leave-to {
  opacity: 0;
}

@media (max-width: 900px) {
  .population-shell {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
    min-width: 100%;
  }
}
</style>
