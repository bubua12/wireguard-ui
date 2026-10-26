import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from './views/Dashboard.vue'
import Peers from './views/Peers.vue'
import Settings from './views/Settings.vue'
import Login from './views/Login.vue'
import { clearToken, isTokenValid } from './utils/auth'

const routes = [
  { path: '/login', component: Login, meta: { public: true, title: '登录' } },
  { path: '/', component: Dashboard, meta: { title: '仪表盘' } },
  { path: '/peers', component: Peers, meta: { title: '客户端' } },
  { path: '/settings', component: Settings, meta: { title: '设置' } },
  { path: '/:pathMatch(.*)*', redirect: '/' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  const authed = isTokenValid()
  if (!to.meta.public && !authed) {
    clearToken()
    return { path: '/login', query: to.fullPath === '/' ? {} : { redirect: to.fullPath } }
  }
  if (to.path === '/login' && authed) {
    return { path: '/' }
  }
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · WireGuard UI` : 'WireGuard UI'
})

export default router
