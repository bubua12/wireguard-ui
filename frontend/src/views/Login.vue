<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900">
    <div class="bg-white dark:bg-gray-800 p-8 rounded-lg shadow-md w-full max-w-md">
      <h1 class="text-2xl font-bold text-center mb-6 dark:text-white">WireGuard UI</h1>

      <div v-if="!initialized" class="mb-4 p-3 bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-200 rounded">
        首次使用，请创建管理员账户
      </div>

      <form @submit.prevent="submit">
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">用户名</label>
          <input v-model="form.username" type="text" class="input-field" required />
        </div>
        <div class="mb-6">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">密码</label>
          <input v-model="form.password" type="password" class="input-field" required />
        </div>
        <div v-if="error" class="mb-4 p-3 bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-200 rounded">
          {{ error }}
        </div>
        <button type="submit" class="w-full btn-primary">
          {{ initialized ? '登录' : '创建账户' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

const router = useRouter()
const initialized = ref(true)
const error = ref('')
const form = ref({ username: '', password: '' })

onMounted(async () => {
  const saved = localStorage.getItem('theme')
  if (saved === 'dark') {
    document.documentElement.classList.add('dark')
  }

  try {
    const res = await axios.get('/api/init')
    initialized.value = res.data.initialized
  } catch (e) {
    console.error(e)
  }
})

const submit = async () => {
  error.value = ''
  try {
    const url = initialized.value ? '/api/login' : '/api/register'
    const res = await axios.post(url, form.value)

    if (!initialized.value) {
      initialized.value = true
      form.value.password = ''
      error.value = ''
      return
    }

    localStorage.setItem('token', res.data.token)
    router.push('/')
  } catch (e) {
    error.value = e.response?.data?.error || '操作失败'
  }
}
</script>
