<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-100 px-4 py-10 dark:bg-gray-900">
    <div class="card w-full max-w-md p-8">
      <div class="mb-4 flex justify-center">
        <LogoMark class="h-20 w-20" />
      </div>
      <h1 class="mb-1 text-center text-2xl font-bold text-gray-900 dark:text-white">WireGuard UI</h1>
      <p class="mb-6 text-center text-sm text-gray-500 dark:text-gray-400">
        {{ initialized ? '登录后管理你的 WireGuard 服务器和客户端' : '创建第一个管理员账户' }}
      </p>

      <div
        v-if="!initialized"
        class="mb-4 rounded-md bg-blue-50 p-3 text-sm text-blue-700 dark:bg-blue-900/30 dark:text-blue-200"
      >
        首次使用，请创建管理员账户（密码至少 8 位，不要使用 admin）
      </div>

      <div
        v-if="error"
        role="alert"
        class="mb-4 flex items-start gap-2 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-200"
      >
        <Icon name="x-circle" class="mt-0.5 h-4 w-4" />
        <span>{{ error }}</span>
      </div>

      <form @submit.prevent="submit">
        <div class="mb-4">
          <label class="field-label" for="username">用户名</label>
          <input
            id="username"
            ref="usernameInput"
            v-model.trim="form.username"
            type="text"
            name="username"
            autocomplete="username"
            class="input-field"
            required
          />
        </div>
        <div class="mb-6">
          <label class="field-label" for="password">密码</label>
          <SecretField
            id="password"
            v-model="form.password"
            as-password
            name="password"
            :autocomplete="initialized ? 'current-password' : 'new-password'"
            :minlength="initialized ? undefined : 8"
            required
          />
        </div>
        <button type="submit" class="btn-primary w-full" :disabled="loading">
          <span v-if="loading" class="spinner"></span>
          {{ initialized ? '登录' : '创建账户' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import Icon from '../components/Icon.vue'
import LogoMark from '../components/LogoMark.vue'
import SecretField from '../components/SecretField.vue'
import { useToast } from '../composables/useToast'
import { errorMessage } from '../utils/api'
import { setToken } from '../utils/auth'

const route = useRoute()
const router = useRouter()
const toast = useToast()

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
