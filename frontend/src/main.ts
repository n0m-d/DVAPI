import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { toast } from 'vue-sonner'
import App from './App.vue'
import router from './router'
import './styles/globals.css'

const app = createApp(App)
const pinia = createPinia()

app.config.globalProperties.$toast = toast
app.provide('toast', toast)

app.use(pinia)
app.use(router)
app.mount('#app')

declare module 'vue' {
  interface ComponentCustomProperties {
    $toast: typeof toast
  }
}
