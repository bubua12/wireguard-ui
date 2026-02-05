<template>
  <div>
    <h1 class="text-2xl font-bold mb-6 dark:text-white">服务器设置</h1>

    <!-- 导入现有配置 -->
    <div v-if="!loading && (isNew || showImportPanel)" class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6 dark:bg-yellow-900/20 dark:border-yellow-800">
      <h3 class="text-lg font-semibold text-yellow-800 mb-2 dark:text-yellow-200">导入现有配置</h3>
      <p class="text-sm text-yellow-700 mb-4 dark:text-yellow-300">
        从服务器现有的 WireGuard 配置文件导入。<strong>注意：这会覆盖当前 UI 中的配置！</strong>
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
        <button @click="importConfig" class="bg-yellow-600 text-white px-4 py-2 rounded hover:bg-yellow-700">
          导入配置
        </button>
        <button v-if="!isNew" @click="showImportPanel = false" class="btn-secondary">
          取消
        </button>
      </div>
    </div>

    <!-- 显示导入按钮（当已有配置且面板隐藏时） -->
    <div v-if="!loading && !isNew && !showImportPanel" class="mb-6">
      <button @click="showImportPanel = true" class="text-yellow-600 hover:text-yellow-800 text-sm">
        从系统配置文件导入...
      </button>
    </div>

    <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">名称</label>
          <input v-model="form.name" type="text" class="input-field" />
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
      </div>

      <div class="mt-6 flex space-x-4">
        <button @click="save" class="btn-primary">保存</button>
        <button @click="sync" class="btn-secondary">同步到系统</button>
      </div>
    </div>

    <!-- 修改密码 -->
    <h2 class="text-xl font-bold mt-8 mb-4 dark:text-white">修改密码</h2>
    <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">原密码</label>
          <input v-model="pwdForm.old_password" type="password" class="input-field" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1 dark:text-gray-300">新密码</label>
          <input v-model="pwdForm.new_password" type="password" class="input-field" placeholder="至少6位" />
        </div>
      </div>
      <div class="mt-6">
        <button @click="changePassword" class="btn-primary">修改密码</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const form = ref({
  name: '',
  endpoint: '',
  address: '',
  listen_port: 51820,
  dns: '8.8.8.8',
  mtu: 1420
})

const isNew = ref(true)
const showImportPanel = ref(false)
const loading = ref(true)

const importForm = ref({
  config_path: '/etc/wireguard/wg0.conf',
  endpoint: '',
  dns: '8.8.8.8'
})

const pwdForm = ref({
  old_password: '',
  new_password: ''
})

onMounted(async () => {
  try {
    const res = await axios.get('/api/server')
    form.value = res.data
    isNew.value = false
  } catch (e) {
    // Server not configured yet
  } finally {
    loading.value = false
  }
})

const save = async () => {
  if (isNew.value) {
    await axios.post('/api/server', form.value)
  } else {
    await axios.put('/api/server', form.value)
  }
  alert('保存成功！')
}

const sync = async () => {
  await axios.post('/api/sync')
  alert('配置已同步到系统！')
}

const changePassword = async () => {
  if (!pwdForm.value.old_password || !pwdForm.value.new_password) {
    alert('请填写完整')
    return
  }
  if (pwdForm.value.new_password.length < 6) {
    alert('新密码至少6位')
    return
  }
  try {
    await axios.post('/api/change-password', pwdForm.value)
    alert('密码修改成功')
    pwdForm.value = { old_password: '', new_password: '' }
  } catch (e) {
    alert(e.response?.data?.error || '修改失败')
  }
}

const importConfig = async () => {
  if (!importForm.value.endpoint) {
    alert('请填写公网地址')
    return
  }
  try {
    const res = await axios.post('/api/import', importForm.value)
    alert(res.data.message)
    // 重新加载服务器配置
    const serverRes = await axios.get('/api/server')
    form.value = serverRes.data
    isNew.value = false
  } catch (e) {
    alert(e.response?.data?.error || '导入失败')
  }
}
</script>
