import { i18n } from '@/locales'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import pinia from './stores/pinia'
import './styles/index.scss'

const app = createApp(App)

app.use(pinia)
app.use(i18n)
app.use(router)
app.use(ElementPlus)

app.mount('#app')
