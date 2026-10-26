import { createApp } from 'vue'
import axios from 'axios'
import App from './App.vue'
import router from './router'
import { initTheme } from './composables/useTheme'
import { clearToken, getToken } from './utils/auth'
import './style.css'

// Restarting the interface can take a while (wg-quick runs PostUp hooks);
// polling calls pass their own, shorter timeout.
axios.defaults.timeout = 130000

axios.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

axios.interceptors.response.use(
  (res) => res,
  (err) => {
    // A 401 from the login call itself just means a wrong password.
    if (err.response?.status === 401 && !err.config?.url?.endsWith('/api/login')) {
      clearToken()
      const current = router.currentRoute.value
      if (current.path !== '/login') {
        router.push({ path: '/login', query: { redirect: current.fullPath } })
      }
    }
    return Promise.reject(err)
  }
)

initTheme()
createApp(App).use(router).mount('#app')
