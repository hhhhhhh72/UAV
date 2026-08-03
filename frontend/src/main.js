import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

// Element Plus (桌面端管理后台)
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

// 全局样式
import './styles/global.css'
import './utils/http'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus)

app.mount('#app')

