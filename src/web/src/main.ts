import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createHead } from '@vueuse/head'
import './style.css'
import App from './App.vue'
import router from './router'
import { setUnauthorizedHandler } from './api/client'
import { useAuthStore } from './stores/auth'

const pinia = createPinia()
const app = createApp(App).use(pinia).use(createHead()).use(router)

// Phiên hết hạn / cookie trỏ tới user không còn tồn tại (vd sau khi reset DB):
// mọi 401 → xoá phiên + về /login, tránh 401 lặp vô hạn (/ws, refetch...).
setUnauthorizedHandler(() => {
  const auth = useAuthStore(pinia)
  auth.$patch({ user: null, household: null, ready: false })
  if (router.currentRoute.value.name !== 'login') {
    router.replace({ name: 'login' })
  }
})

app.mount('#app')
