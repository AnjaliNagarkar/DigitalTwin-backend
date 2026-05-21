<template>
  <router-view />
</template>

<script setup>
import { ref, provide } from 'vue'

function detectEmbedded() {
  if (window.location.search.indexOf('embedded=true') !== -1) return true
  try { if (sessionStorage.getItem('agritwin_embedded') === 'true') return true } catch (_) {}
  try { if (window.self !== window.top) return true } catch (_) { return true }
  return false
}

// Evaluated synchronously during setup — available to all child components before they mount.
const isEmbedded = ref(detectEmbedded())

if (isEmbedded.value) {
  document.documentElement.classList.add('is-embedded')
  try { sessionStorage.setItem('agritwin_embedded', 'true') } catch (_) {}
}

provide('isEmbedded', isEmbedded)
</script>
