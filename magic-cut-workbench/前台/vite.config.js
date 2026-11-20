import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 4040,
    proxy: {
      '/api': 'http://127.0.0.1:4041',
      '/health': 'http://127.0.0.1:4041',
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
