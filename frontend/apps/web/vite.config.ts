import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    dedupe: ['react', 'react-dom']
  },
  server: {
    proxy: {
      '/api': {
        target: process.env.API_TARGET || 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/health': {
        target: process.env.API_TARGET || 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      }
    }
  }
})
