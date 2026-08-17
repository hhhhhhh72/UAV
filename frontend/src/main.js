import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

// Arco Design Vue（B 端管理后台组件库）
// 按需自动导入：unplugin-vue-components 处理模板 a-xxx/icon-xxx 组件，
// unplugin-auto-import 注入 Message/Modal 等函数式 API（vite.config.js 配置）。
// 不再 app.use(ArcoVue) 全量注册——主包体积大幅下降（首屏优化）。
import '@arco-design/web-vue/dist/arco.css'

// 消除 ResizeObserver loop 良性警告（el-table 自适应宽度在容器尺寸变化时触发，
// Chrome 已知行为，不影响功能；用 requestAnimationFrame 节流后不再报 console 噪音）
if (typeof window !== 'undefined' && window.ResizeObserver) {
  const NativeResizeObserver = window.ResizeObserver
  window.ResizeObserver = class extends NativeResizeObserver {
    constructor(callback) {
      super((entries, observer) => {
        window.requestAnimationFrame(() => callback(entries, observer))
      })
    }
  }
}

// Arco Design Vue（B 端管理后台组件库，参照 Arco Design Pro 效果）

// 全局样式
import './styles/global.css'
import './utils/http'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')

