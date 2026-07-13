/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Dev: proxy /api + /ws → Go API :8080 (cookie same-origin — research D10).
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/ws': { target: 'ws://localhost:8080', ws: true },
    },
  },
  test: {
    environment: 'jsdom',
    include: ['src/**/__tests__/**/*.spec.ts'],
  },
})
