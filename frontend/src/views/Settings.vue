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

    <h1 class="text-2xl font-bold mb-4 dark:text-white">服务器设置</h1>
    <div class="flex items-center justify-end gap-2 mb-4">
      <button
        v-if="!loading"
        @click="openImportModal"
        class="inline-flex items-center px-3 py-2 text-sm bg-yellow-100 text-yellow-700 rounded-md hover:bg-yellow-200 dark:bg-yellow-900/30 dark:text-yellow-400 dark:hover:bg-yellow-900/50 transition-colors duration-200"
      >
        <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"/>
        </svg>
        从系统配置文件导入
      </button>
      <button @click="openPwdModal" class="btn-secondary text-sm">修改密码</button>
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
          <p class="text-xs text-gray-500 mt-1">系统网卡名，一般是 wg0，不要填显示名称或域名</p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">公网地址</label>
          <SecretField v-model="form.endpoint" placeholder="example.com:51820" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">内网地址</label>
          <SecretField v-model="form.address" placeholder="10.0.0.1/24" />
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
        <div class="md:col-span-2 grid grid-cols-1 md:grid-cols-2 gap-4 mt-2">
          <label class="flex items-start gap-2 text-sm text-gray-700 dark:text-gray-300 cursor-pointer">
            <input v-model="form.full_tunnel" type="checkbox" class="mt-1" />
            <span>
              <span class="block font-medium">全局流量</span>
              <span class="block text-xs text-gray-500 dark:text-gray-400 mt-0.5">客户端全部上网走隧道（AllowedIPs = 0.0.0.0/0）</span>
            </span>
          </label>
          <label class="flex items-start gap-2 text-sm text-gray-700 dark:text-gray-300 cursor-pointer">
            <input v-model="form.enable_nat" type="checkbox" class="mt-1" />
            <span>
              <span class="block font-medium">启用 NAT / 转发</span>
              <span class="block text-xs text-gray-500 dark:text-gray-400 mt-0.5">写入 iptables PostUp/PostDown，让客户端能出网</span>
            </span>
          </label>
        </div>
      </div>

      <div class="mt-6 flex space-x-4">
        <button @click="save" class="btn-primary">保存</button>
        <button @click="sync" class="btn-secondary">同步到系统</button>
      </div>
    </div>

    <div v-if="showImportModal" class="modal-overlay" @click.self="closeImportModal">
      <div class="modal-content-lg">
        <h3 class="text-lg font-semibold mb-3 dark:text-white">导入现有配置</h3>
        <div class="mb-4 rounded-md border border-yellow-200 bg-yellow-50 px-3 py-2 text-sm text-yellow-800 dark:border-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-200">
          从 <code class="text-xs">/etc/wireguard/*.conf</code> 导入。会<strong>覆盖</strong>当前 UI 中的服务器和客户端。导入的客户端没有私钥，无法再下载配置或二维码。
        </div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">配置文件路径</label>
            <input v-model="importForm.config_path" type="text" placeholder="/etc/wireguard/wg0.conf" class="input-field" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">公网地址（必填）</label>
            <SecretField v-model="importForm.endpoint" placeholder="your-server.com:51820" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">DNS</label>
            <input v-model="importForm.dns" type="text" placeholder="8.8.8.8" class="input-field" />
          </div>
        </div>
        <div class="flex justify-end space-x-2 mt-6">
          <button @click="closeImportModal" class="btn-secondary">取消</button>
          <button @click="confirmImport" class="bg-yellow-600 text-white px-4 py-2 rounded hover:bg-yellow-700">导入并覆盖</button>
        </div>
      </div>
    </div>

    <div v-if="showPwdModal" class="modal-overlay" @click.self="closePwdModal">
      <div class="modal-content">
        <h3 class="text-lg font-semibold mb-4 dark:text-white">修改密码</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">原密码</label>
            <SecretField v-model="pwdForm.old_password" as-password />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">新密码</label>
            <SecretField v-model="pwdForm.new_password" as-password placeholder="至少8位" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">确认新密码</label>
            <SecretField v-model="pwdForm.confirm_password" as-password placeholder="再输入一次" />
          </div>
        </div>
        <div class="flex justify-end space-x-2 mt-6">
          <button @click="closePwdModal" class="btn-secondary">取消</button>
          <button @click="changePassword" class="btn-primary">确认修改</button>
        </div>
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
import SecretField from '../components/SecretField.vue'

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

const emptyPwd = () => ({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const emptyImport = () => ({
  config_path: '/etc/wireguard/wg0.conf',
  endpoint: '',
  dns: '8.8.8.8',
  full_tunnel: false,
  enable_nat: false
})

const form = ref(emptyForm())
const isNew = ref(true)
const showImportModal = ref(false)
const loading = ref(true)
const showPwdModal = ref(false)
const confirmModal = ref({ show: false, title: '', message: '', onConfirm: () => {} })

const importForm = ref(emptyImport())

const pwdForm = ref(emptyPwd())

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

const openImportModal = () => {
  importForm.value = emptyImport()
  showImportModal.value = true
}

const closeImportModal = () => {
  showImportModal.value = false
}

const openPwdModal = () => {
  pwdForm.value = emptyPwd()
  showPwdModal.value = true
}

const closePwdModal = () => {
  showPwdModal.value = false
  pwdForm.value = emptyPwd()
}

const changePassword = async () => {
  if (!pwdForm.value.old_password || !pwdForm.value.new_password || !pwdForm.value.confirm_password) {
    showToast('warning', '请填写完整')
    return
  }
  if (pwdForm.value.new_password.length < 8) {
    showToast('warning', '新密码至少8位')
    return
  }
  if (pwdForm.value.new_password !== pwdForm.value.confirm_password) {
    showToast('warning', '两次输入的新密码不一致')
    return
  }
  try {
    await axios.post('/api/change-password', {
      old_password: pwdForm.value.old_password,
      new_password: pwdForm.value.new_password
    })
    showToast('success', '密码修改成功')
    closePwdModal()
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
    closeImportModal()
  } catch (e) {
    showToast('error', e.response?.data?.error || '导入失败')
  }
}
</script>
