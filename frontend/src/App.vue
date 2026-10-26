<template>
  <div class="min-h-screen">
    <button
      v-if="isLoginPage"
      type="button"
      class="fixed right-4 top-4 z-30 rounded-md bg-white p-2 shadow hover:bg-gray-100 dark:bg-gray-800 dark:hover:bg-gray-700"
      :title="isDark ? '切换到亮色' : '切换到暗色'"
      aria-label="切换主题"
      @click="toggle"
    >
      <Icon :name="isDark ? 'sun' : 'moon'" class="h-5 w-5" :class="isDark ? 'text-yellow-400' : 'text-gray-600'" />
    </button>

    <header
      v-if="!isLoginPage"
      class="fixed inset-x-0 top-0 z-40 border-b border-gray-200 bg-white/90 backdrop-blur dark:border-gray-700 dark:bg-gray-800/90"
    >
      <div class="mx-auto flex h-16 max-w-7xl items-center justify-between gap-2 px-4">
        <router-link to="/" class="flex shrink-0 items-center gap-2" aria-label="WireGuard UI">
          <LogoMark class="h-8 w-8" />
          <span class="hidden text-xl font-bold text-gray-800 dark:text-white sm:inline">WireGuard UI</span>
        </router-link>

        <nav class="flex items-center gap-1" aria-label="主导航">
          <router-link
            v-for="item in nav"
            :key="item.to"
            :to="item.to"
            class="nav-link"
            active-class="nav-link-active"
            :aria-label="item.label"
          >
            <Icon :name="item.icon" class="h-5 w-5 sm:h-4 sm:w-4" />
            <span class="hidden sm:inline">{{ item.label }}</span>
          </router-link>
        </nav>

        <div class="flex items-center gap-1">
          <button
            type="button"
            class="rounded-md p-2 hover:bg-gray-100 dark:hover:bg-gray-700"
            :title="isDark ? '切换到亮色' : '切换到暗色'"
            aria-label="切换主题"
            @click="toggle"
          >
            <Icon :name="isDark ? 'sun' : 'moon'" class="h-5 w-5" :class="isDark ? 'text-yellow-400' : 'text-gray-600'" />
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-md px-2 py-2 text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
            title="退出登录"
            @click="logout"
          >
            <Icon name="logout" class="h-5 w-5 sm:h-4 sm:w-4" />
            <span class="hidden sm:inline">退出</span>
          </button>
        </div>
      </div>
    </header>

    <main :class="isLoginPage ? '' : 'mx-auto max-w-7xl px-4 pb-10 pt-24'">
      <router-view />
    </main>

    <ToastHost />
    <ConfirmDialog />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ConfirmDialog from './components/ConfirmDialog.vue'
import Icon from './components/Icon.vue'
import LogoMark from './components/LogoMark.vue'
import ToastHost from './components/ToastHost.vue'
import { useTheme } from './composables/useTheme'
import { clearToken } from './utils/auth'

const route = useRoute()
const router = useRouter()
const { isDark, toggle } = useTheme()

const nav = [
  { to: '/', label: '仪表盘', icon: 'home' },
  { to: '/peers', label: '客户端', icon: 'users' },
  { to: '/settings', label: '设置', icon: 'cog' }
]

const isLoginPage = computed(() => route.path === '/login')

const logout = () => {
  clearToken()
  router.push('/login')
}
</script>
