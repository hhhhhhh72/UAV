import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

// Vant组件库
import Vant from 'vant'
import 'vant/lib/index.css'

// 全局样式
import './styles/global.css'
import './utils/http'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(Vant)

app.mount('#app')

