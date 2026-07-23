import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
// import basicSsl from '@vitejs/plugin-basic-ssl'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig(({ mode }) => {
  // Load env file based on `mode` in the current working directory.
  // Set the third parameter to '' to load all env regardless of the `VITE_` prefix.
  const env = loadEnv(mode, process.cwd(), '')

  return {
  base: '/',
  plugins: [
    vue(),
    // basicSsl()
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          echarts: ['echarts', 'vue-echarts']
        }
      }
    }
  },
  server: {
    port: 5173,
    host: '0.0.0.0',
    allowedHosts: true, // Allow ngrok/localtunnel hosts
    proxy: {
      '/api': {
        target: env.VITE_API_TARGET || 'http://localhost:3000',
        changeOrigin: true
      },
      '/uploads': {
        target: env.VITE_API_TARGET || 'http://localhost:3000',
        changeOrigin: true
      }
    },
    // 添加安全响应头
    headers: {
      'X-Content-Type-Options': 'nosniff',
      'X-Frame-Options': 'SAMEORIGIN',
      'X-XSS-Protection': '1; mode=block'
    }
  }
}})

