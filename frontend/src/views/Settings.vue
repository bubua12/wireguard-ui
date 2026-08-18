<template>
  <div>
    <Transition name="toast">
      <div v-if="toast.show" class="fixed top-4 right-4 z-50 max-w-sm">
        <div class="rounded-lg shadow-lg p-4 flex items-center space-x-3" :class="toastClass">
          <div class="flex-shrink-0">
            <svg v-if="toast.type === 'success'" class="w-6 h-6 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <svg v-else-if="toast.type === 'error'" class="w-6 h-6 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <svg v-else class="w-6 h-6 text-yellow-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
            </svg>
          </div>
          <div class="flex-1">
            <p class="font-medium" :class="toastTextClass">{{ toast.message }}</p>
          </div>
          <button @click="toast.show = false" class="flex-shrink-0 text-gray-400 hover:text-gray-600">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>
      </div>
    </Transition>

    <h1 class="text-2xl font-bold mb-6 dark:text-white">服务器设置</h1>

    <div v-if="!loading && (isNew || showImportPanel)" class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6 dark:bg-yellow-900/20 dark:border-yellow-800">
      <h3 class="text-lg font-semibold text-yellow-800 mb-2 dark:text-yellow-200">导入现有配置</h3>
      <p class="text-sm text-yellow-700 mb-4 dark:text-yellow-300">
        从 <code>/etc/wireguard/*.conf</code> 导入。这会<strong>覆盖</strong>当前 UI 中的服务器和客户端。导入的客户端没有私钥，无法再下载配置或二维码。
      </p>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
        <div>
          <label class="block text-sm font-medium text-yellow-700 mb-1 dark:text-yellow-300">配置文件路径</label>
          <input v-model="importForm.config_path" type="text" placeholder="/etc/wireguard/wg0.conf" class="input-field" />
        </div>
        <div>
          <label class="block text-sm font-medium text-yellow-700 mb-1 dark:text-yellow-300">公网地址 (必填)</label>
          <input v-model="importForm.endpoint" type="text" placeholder="your-server.com:51820" class="input-field" />
        </div>
        <div>
          <label class="block text-sm font-medium text-yellow-700 mb-1 dark:text-yellow-300">DNS</label>
          <input v-model="importForm.dns" type="text" placeholder="8.8.8.8" class="input-field" />
        </div>
      </div>
      <div class="flex space-x-2">
        <button @click="confirmImport" class="bg-yellow-600 text-white px-4 py-2 rounded hover:bg-yellow-700">
          导入并覆盖
        </button>
        <button v-if="!isNew" @click="showImportPanel = false" class="btn-secondary">
          取消
        </button>
      </div>
    </div>

    <div v-if="!loading && !isNew && !showImportPanel" class="mb-6">
      <button @click="showImportPanel = true" class="inline-flex items-center px-4 py-2 bg-yellow-100 text-yellow-700 rounded-md hover:bg-yellow-200 dark:bg-yellow-900/30 dark:text-yellow-400 dark:hover:bg-yellow-900/50 transition-colors duration-200">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"/>
        </svg>
        从系统配置文件导入
      </button>
    </div>

    <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">显示名称</label>
          <input v-model="form.name" type="text" class="input-field" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">接口名</label>
          <input v-model="form.interface" type="text" placeholder="wg0" class="input-field" />
          <p class="text-xs text-gray-500 mt-1">1-15 位，字母开头，对应 /etc/wireguard/&lt;接口名&gt;.conf</p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">公网地址</label>
          <input v-model="form.endpoint" type="text" placeholder="example.com:51820" class="input-field" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">内网地址</label>
          <input v-model="form.address" type="text" placeholder="10.0.0.1/24" class="input-field" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">监听端口</label>
          <input v-model.number="form.listen_port" type="number" class="input-field" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">DNS</label>
          <input v-model="form.dns" type="text" placeholder="8.8.8.8" class="input-field" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">MTU</label>
          <input v-model.number="form.mtu" type="number" class="input-field" />
        </div>
        <div class="flex items-center space-x-6 md:col-span-2 mt-2">
          <label class="inline-flex items-center text-sm text-gray-700 dark:text-gray-300">
            <input v-model="form.full_tunnel" type="checkbox" class="mr-2" />
            全局流量（客户端 AllowedIPs = 0.0.0.0/0）
          </label>
          <label class="inline-flex items-center text-sm text-gray-700 dark:text-gray-300">
            <input v-model="form.enable_nat" type="checkbox" class="mr-2" />
            启用 NAT / 转发（写入 PostUp/PostDown）
          </label>
        </div>
      </div>

      <div class="mt-6 flex space-x-4">
        <button @click="save" class="btn-primary">保存</button>
        <button @click="sync" class="btn-secondary">同步到系统</button>
      </div>
    </div>

    <h2 class="text-xl font-bold mt-8 mb-4 dark:text-white">修改密码</h2>
    <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">原密码</label>
          <input v-model="pwdForm.old_password" type="password" class="input-field" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">新密码</label>
          <input v-model="pwdForm.new_password" type="password" class="input-field" placeholder="至少8位" />
        </div>
      </div>
      <div class="mt-6">
        <button @click="changePassword" class="btn-primary">修改密码</button>
      </div>
    </div>

    <div v-if="confirmModal.show" class="modal-overlay" @click.self="confirmModal.show = false">
      <div class="modal-content">
        <h3 class="text-lg font-semibold mb-4 dark:text-white">{{ confirmModal.title }}</h3>
        <p class="text-gray-600 dark:text-gray-300 mb-6">{{ confirmModal.message }}</p>
        <div class="flex justify-end space-x-2">
          <button @click="confirmModal.show = false" class="btn-secondary">取消</button>
          <button @click="confirmModal.onConfirm" class="btn-danger">确认覆盖</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'

const toast = ref({
  show: false,
  type: 'success',
  message: ''
})

const toastClass = computed(() => ({
  'bg-green-50 border border-green-200 dark:bg-green-900/20 dark:border-green-800': toast.value.type === 'success',
  'bg-red-50 border border-red-200 dark:bg-red-900/20 dark:border-red-800': toast.value.type === 'error',
  'bg-yellow-50 border border-yellow-200 dark:bg-yellow-900/20 dark:border-yellow-800': toast.value.type === 'warning'
}))

const toastTextClass = computed(() => ({
  'text-green-800 dark:text-green-200': toast.value.type === 'success',
  'text-red-800 dark:text-red-200': toast.value.type === 'error',
  'text-yellow-800 dark:text-yellow-200': toast.value.type === 'warning'
}))

const showToast = (type, message, duration = 3000) => {
  toast.value = { show: true, type, message }
  if (duration > 0) {
    setTimeout(() => { toast.value.show = false }, duration)
  }
}

const emptyForm = () => ({
  name: '',
  interface: 'wg0',
  endpoint: '',
  address: '',
  listen_port: 51820,
  dns: '8.8.8.8',
  mtu: 1420,
  full_tunnel: false,
  enable_nat: false
})

const form = ref(emptyForm())
const isNew = ref(true)
const showImportPanel = ref(false)
const loading = ref(true)
const confirmModal = ref({ show: false, title: '', message: '', onConfirm: () => {} })

const importForm = ref({
  config_path: '/etc/wireguard/wg0.conf',
  endpoint: '',
  dns: '8.8.8.8',
  full_tunnel: false,
  enable_nat: false
})

const pwdForm = ref({
  old_password: '',
  new_password: ''
})

const applyServer = (data) => {
  form.value = {
    ...emptyForm(),
    ...data,
    full_tunnel: !!data.full_tunnel,
    enable_nat: !!data.enable_nat
  }
}

onMounted(async () => {
  try {
    const res = await axios.get('/api/server')
    applyServer(res.data)
    isNew.value = false
  } catch (e) {
    // Server not configured yet
  } finally {
    loading.value = false
  }
})

const save = async () => {
  try {
    if (isNew.value) {
      const res = await axios.post('/api/server', form.value)
      applyServer(res.data)
      isNew.value = false
    } else {
      const res = await axios.put('/api/server', form.value)
      applyServer(res.data)
    }
    showToast('success', '已保存到数据库。修改接口/地址/NAT 后请再点「同步到系统」。')
  } catch (e) {
    showToast('error', e.response?.data?.error || '保存失败')
  }
}

const sync = async () => {
  try {
    await axios.post('/api/sync')
    showToast('success', '配置已同步到系统！')
  } catch (e) {
    showToast('error', e.response?.data?.error || '同步失败')
  }
}

const changePassword = async () => {
  if (!pwdForm.value.old_password || !pwdForm.value.new_password) {
    showToast('warning', '请填写完整')
    return
  }
  if (pwdForm.value.new_password.length < 8) {
    showToast('warning', '新密码至少8位')
    return
  }
  try {
    await axios.post('/api/change-password', pwdForm.value)
    showToast('success', '密码修改成功')
    pwdForm.value = { old_password: '', new_password: '' }
  } catch (e) {
    showToast('error', e.response?.data?.error || '修改失败')
  }
}

const confirmImport = () => {
  if (!importForm.value.endpoint) {
    showToast('warning', '请填写公网地址')
    return
  }
  confirmModal.value = {
    show: true,
    title: '覆盖导入',
    message: '导入会删除当前 UI 中的服务器和全部客户端，然后写入配置文件中的内容。确定继续？',
    onConfirm: importConfig
  }
}

const importConfig = async () => {
  confirmModal.value.show = false
  try {
    const res = await axios.post('/api/import', importForm.value)
    showToast('success', res.data.message, 5000)
    const serverRes = await axios.get('/api/server')
    applyServer(serverRes.data)
    isNew.value = false
    showImportPanel.value = false
  } catch (e) {
    showToast('error', e.response?.data?.error || '导入失败')
  }
}
</script>
