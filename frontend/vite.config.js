import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
// import basicSsl from '@vitejs/plugin-basic-ssl'
import Components from 'unplugin-vue-components/vite'
import AutoImport from 'unplugin-auto-import/vite'
import { ArcoResolver } from 'unplugin-vue-components/resolvers'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig(({ mode }) => {
  // Load env file based on `mode` in the current working directory.
  // Set the third parameter to '' to load all env regardless of the `VITE_` prefix.
  const env = loadEnv(mode, process.cwd(), '')

  return {
  base: '/',
  plugins: [
    vue(),
    // Arco 按需自动导入：模板 a-xxx 组件与 icon-xxx 自动注册（含样式），
    // 函数式 API（Message/Modal 等）由 AutoImport 注入——主包不再全量打包 Arco。
    Components({
      resolvers: [ArcoResolver({ sideEffect: true, resolveIcons: true })]
    }),
    AutoImport({
      resolvers: [ArcoResolver({ sideEffect: true, resolveIcons: true })]
    })
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
          echarts: ['echarts', 'vue-echarts'],
          // Vue 全家桶独立 vendor：主包只含业务代码，浏览器长缓存利用最大化
          vendor: ['vue', 'vue-router', 'pinia', 'axios']
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
        target: env.VITE_API_TARGET || 'http://localhost:8080',
        changeOrigin: true
      },
      '/uploads': {
        target: env.VITE_API_TARGET || 'http://localhost:8080',
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

