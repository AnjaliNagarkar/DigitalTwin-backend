import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import cesium from 'vite-plugin-cesium'

export default defineConfig({
  plugins: [vue(), cesium()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      },
      '/ping':             { target: 'http://localhost:8081', changeOrigin: true },
      '/houses':           { target: 'http://localhost:8081', changeOrigin: true },
      '/houses/batch-members': { target: 'http://localhost:8081', changeOrigin: true },
      '/house':            { target: 'http://localhost:8081', changeOrigin: true },
      '/districts':        { target: 'http://localhost:8081', changeOrigin: true },
      '/location-options': { target: 'http://localhost:8081', changeOrigin: true },
      '/map':              { target: 'http://localhost:8081', changeOrigin: true },
      '/pdf':              { target: 'http://localhost:8081', changeOrigin: true },
      '/insights':         { target: 'http://localhost:8081', changeOrigin: true },
      '/dashboard':        { target: 'http://localhost:8081', changeOrigin: true },
      '/population':       { target: 'http://localhost:8081', changeOrigin: true },
      '/crops':            { target: 'http://localhost:8081', changeOrigin: true },
      '/land':             { target: 'http://localhost:8081', changeOrigin: true },
      '/irrigation':       { target: 'http://localhost:8081', changeOrigin: true },
      '/citizens':         { target: 'http://localhost:8081', changeOrigin: true },
      '/farmers':          { target: 'http://localhost:8081', changeOrigin: true },
      '/unified-registry': { target: 'http://localhost:8081', changeOrigin: true },
      '/soil':             { target: 'http://localhost:8081', changeOrigin: true },
      '/schemes':          { target: 'http://localhost:8081', changeOrigin: true },
      '/advisory':         { target: 'http://localhost:8081', changeOrigin: true },
      '/market':           { target: 'http://localhost:8081', changeOrigin: true }
    }
  }
})
