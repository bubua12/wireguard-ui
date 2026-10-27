<template>
  <div
    class="relative isolate flex min-h-screen flex-col bg-gradient-to-br from-blue-600 via-blue-800 to-indigo-950 text-white dark:from-blue-900 dark:via-indigo-950 dark:to-gray-950 lg:grid lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]"
  >
    <!-- 整页只有这一张渐变背景：品牌区直接放在上面，表单是浮在上面的一张卡片，中间没有分块的边框 -->
    <div class="pointer-events-none absolute inset-0 -z-10 overflow-hidden" aria-hidden="true">
      <div class="login-dots absolute inset-0"></div>
      <div class="absolute -left-24 -top-24 h-96 w-96 rounded-full bg-blue-400/30 blur-3xl"></div>
      <div class="absolute -bottom-48 right-0 h-[30rem] w-[30rem] rounded-full bg-indigo-500/30 blur-3xl"></div>
    </div>

    <!-- 品牌区：手机和平板上只是顶部的标题，桌面端占左边一半 -->
    <aside class="flex flex-col px-6 pb-8 pt-10 lg:sticky lg:top-0 lg:h-screen lg:p-12">
      <div class="flex flex-col items-center gap-3 text-center lg:flex-row lg:text-left">
        <!-- Logo 是深红色的，放在白底方块里，深浅主题下都清楚 -->
        <span
          class="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl bg-white shadow-lg shadow-blue-950/30 lg:h-10 lg:w-10 lg:rounded-xl"
        >
          <LogoMark class="h-9 w-9 lg:h-7 lg:w-7" />
        </span>
        <div>
          <div class="text-xl font-semibold tracking-tight">WireGuard UI</div>
          <p class="mt-1 text-sm text-blue-100 lg:hidden">简洁的 WireGuard VPN 管理界面</p>
        </div>
      </div>

      <div class="hidden min-h-0 flex-1 items-center justify-center py-6 lg:flex">
        <NetworkArt class="h-full w-full max-w-xl" />
      </div>

      <div class="hidden lg:block">
        <p class="text-3xl font-bold leading-tight tracking-tight xl:text-4xl">
          简洁的 WireGuard<br />VPN 管理界面
        </p>
        <p class="mt-3 text-blue-100">管理服务器参数、客户端密钥和配置下发。</p>
        <ul class="mt-8 space-y-4 [@media(max-height:760px)]:hidden">
          <li v-for="f in features" :key="f.title" class="flex items-start gap-3">
            <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-white/10 ring-1 ring-white/15">
              <Icon :name="f.icon" class="h-5 w-5" />
            </span>
            <div>
              <div class="text-sm font-medium">{{ f.title }}</div>
              <div class="text-sm text-blue-100">{{ f.desc }}</div>
            </div>
          </li>
        </ul>
      </div>
    </aside>

    <!-- 手机上表单是从底部升起的一块白色面板，平板和桌面上是浮在背景上的一张卡片。
         从上往下排而不是垂直居中：出现错误提示时标题和输入框不会跟着晃动，
         上边距让卡片落在视觉中心附近（平板上要多减去顶部标题的高度） -->
    <main
      class="flex flex-1 flex-col md:items-center md:px-6 md:pb-12 md:pt-[max(1.5rem,calc(50vh-19rem))] lg:min-h-screen lg:px-12 lg:pt-[max(2.5rem,calc(50vh-13rem))]"
    >
      <div
        class="w-full flex-1 rounded-t-3xl bg-white px-6 pb-10 pt-9 text-gray-900 dark:bg-gray-900 dark:text-gray-100 md:max-w-md md:flex-none md:rounded-2xl md:p-10 md:shadow-2xl md:shadow-blue-950/40 md:ring-1 md:ring-white/20 dark:md:shadow-black/50 dark:md:ring-white/10"
      >
        <div class="mx-auto w-full max-w-sm md:max-w-none">
          <h1 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white lg:text-3xl">
            {{ initialized ? '欢迎回来' : '创建管理员账户' }}
          </h1>
          <p class="mt-1.5 text-sm text-gray-500 dark:text-gray-400">
            {{ initialized ? '登录后管理你的 WireGuard 服务器和客户端' : '首次使用，设置好账户后就可以开始管理了' }}
          </p>

          <div
            v-if="!initialized"
            class="mt-6 flex items-start gap-2 rounded-lg bg-blue-50 p-3 text-sm text-blue-700 ring-1 ring-inset ring-blue-100 dark:bg-blue-900/30 dark:text-blue-200 dark:ring-blue-800/50"
          >
            <Icon name="info" class="mt-0.5 h-4 w-4" />
            <span>密码至少 8 位，不要使用 admin</span>
          </div>

          <div
            v-if="error"
            role="alert"
            class="mt-6 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-200"
          >
            <Icon name="x-circle" class="mt-0.5 h-4 w-4" />
            <span>{{ error }}</span>
          </div>

          <form class="mt-8 [&_.input-field]:py-2.5" @submit.prevent="submit">
            <div class="mb-5">
              <label class="field-label" for="username">用户名</label>
              <div class="relative">
                <input
                  id="username"
                  ref="usernameInput"
                  v-model.trim="form.username"
                  type="text"
                  name="username"
                  autocomplete="username"
                  class="peer input-field input-field-icon"
                  required
                />
                <Icon
                  name="user"
                  class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400 transition-colors peer-focus:text-blue-500"
                />
              </div>
            </div>
            <div class="mb-7">
              <label class="field-label" for="password">密码</label>
              <SecretField
                id="password"
                v-model="form.password"
                as-password
                icon="lock"
                name="password"
                :autocomplete="initialized ? 'current-password' : 'new-password'"
                :minlength="initialized ? undefined : 8"
                required
              />
            </div>
            <button type="submit" class="btn-primary group w-full py-2.5 shadow-lg shadow-blue-600/25" :disabled="loading">
              <span v-if="loading" class="spinner"></span>
              {{ initialized ? '登录' : '创建账户' }}
              <Icon v-if="!loading" name="arrow-right" class="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
            </button>
          </form>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import Icon from '../components/Icon.vue'
import LogoMark from '../components/LogoMark.vue'
import NetworkArt from '../components/NetworkArt.vue'
import SecretField from '../components/SecretField.vue'
import { useToast } from '../composables/useToast'
import { errorMessage } from '../utils/api'
import { setToken } from '../utils/auth'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const features = [
  { icon: 'users', title: '客户端管理', desc: '添加、启用或禁用客户端，下载配置或扫码连接' },
  { icon: 'chart', title: '流量统计', desc: '实时速率、累计流量和每日趋势一目了然' },
  { icon: 'power', title: '接口控制', desc: '在页面上启动、停止和重启 WireGuard 接口' }
]

const initialized = ref(true)
const form = ref({ username: '', password: '' })
const error = ref('')
const loading = ref(false)
const usernameInput = ref(null)

// Only follow redirects to pages inside this app.
const redirectTarget = () => {
  const r = route.query.redirect
  return typeof r === 'string' && r.startsWith('/') && !r.startsWith('//') ? r : '/'
}

onMounted(async () => {
  try {
    const res = await axios.get('/api/init', { timeout: 10000 })
    initialized.value = res.data.initialized
  } catch (e) {
    error.value = errorMessage(e, '无法获取初始化状态')
  }
  usernameInput.value?.focus()
})

const submit = async () => {
  error.value = ''
  loading.value = true
  try {
    const res = await axios.post(initialized.value ? '/api/login' : '/api/register', form.value)

    if (!initialized.value) {
      initialized.value = true
      form.value.password = ''
      toast.success('管理员账户已创建，请登录')
      return
    }

    setToken(res.data.token)
    router.push(redirectTarget())
  } catch (e) {
    const data = e.response?.data
    let msg = errorMessage(e, '操作失败')
    if (typeof data?.remaining_attempts === 'number' && data.remaining_attempts <= 5) {
      msg += `（还可尝试 ${data.remaining_attempts} 次）`
    }
    error.value = msg
  } finally {
    loading.value = false
  }
}
</script>
