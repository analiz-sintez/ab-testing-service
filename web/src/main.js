import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import routes from './router'
import { createPinia } from 'pinia'
import './assets/index.css'

const pinia = createPinia()
const app = createApp(App)

app.use(pinia)
app.use(routes)
app.use(ElementPlus)
app.mount('#app')
