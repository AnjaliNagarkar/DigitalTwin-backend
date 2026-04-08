<template>
  <div class="main-shell">
    <aside class="sidebar">
      <div class="brand">
        <h1>Digital Twin</h1>
        <p>Module Gateway</p>
      </div>

      <nav class="nav">
        <router-link to="/" class="nav-item" :class="{ active: $route.path === '/' }">
          Dashboard
        </router-link>

        <div class="group">
          <div class="group-title">Digital Twin</div>
          <router-link
            to="/agriculture/dashboard"
            class="nav-item sub-item"
            :class="{ active: $route.path.startsWith('/agriculture') }"
          >
            Agriculture
          </router-link>
          <router-link
            to="/population/dashboard"
            class="nav-item sub-item"
            :class="{ active: $route.path.startsWith('/population') }"
          >
            Population
          </router-link>
          <router-link
            to="/infrastructure/dashboard"
            class="nav-item sub-item"
            :class="{ active: $route.path.startsWith('/infrastructure') }"
          >
            Infrastructure
          </router-link>
        </div>
      </nav>
    </aside>

    <main class="content">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.main-shell {
  min-height: 100vh;
  display: flex;
  background: #f5f7f4;
  color: #1a2a1f;
  font-family: 'Outfit', sans-serif;
}

.sidebar {
  width: 260px;
  min-width: 260px;
  background: linear-gradient(180deg, #0f2a22 0%, #15362c 100%);
  color: #d5e5dd;
  border-right: 1px solid rgba(255, 255, 255, 0.08);
  padding: 1.5rem 0.9rem;
}

.brand {
  padding: 0.4rem 0.7rem 1.2rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.12);
  margin-bottom: 1rem;
}

.brand h1 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 600;
  color: #ffffff;
}

.brand p {
  margin: 0.25rem 0 0;
  font-size: 0.72rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: #8bb3a1;
}

.nav {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.group {
  margin-top: 0.5rem;
}

.group-title {
  font-size: 0.72rem;
  color: #8bb3a1;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  padding: 0.45rem 0.7rem;
}

.nav-item {
  display: block;
  text-decoration: none;
  color: #d5e5dd;
  border-radius: 8px;
  padding: 0.58rem 0.7rem;
  font-size: 0.9rem;
  transition: background 0.2s ease, color 0.2s ease;
}

.sub-item {
  padding-left: 1.2rem;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.08);
}

.nav-item.active {
  background: #d09b2a;
  color: #1c1302;
  font-weight: 600;
}

.content {
  flex: 1;
  min-width: 0;
}

@media (max-width: 900px) {
  .main-shell {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
    min-width: 100%;
  }
}
</style>
