import path from 'node:path';
import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(import.meta.dirname, './src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      // Locally stored profile media is served by the API from /uploads
      // (see e.Static in cmd/api/main.go) — proxy it like /api so
      // relative avatar/banner URLs resolve in dev.
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('/react-dom/')) {
              return 'react-dom';
            }
            if (id.includes('/@base-ui/')) {
              return 'base-ui';
            }
            if (id.includes('/@reduxjs/') || id.includes('react-redux')) {
              return 'redux';
            }
          }
        },
      },
    },
  },
});
