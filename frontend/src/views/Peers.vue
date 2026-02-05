<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold dark:text-white">客户端管理</h1>
      <button @click="showAddModal = true" class="btn-primary">添加客户端</button>
    </div>

    <div class="bg-white rounded-lg shadow overflow-hidden dark:bg-gray-800">
      <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead class="bg-gray-50 dark:bg-gray-700">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase dark:text-gray-300">名称</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase dark:text-gray-300">IP地址</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase dark:text-gray-300">状态</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase dark:text-gray-300">操作</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200 dark:bg-gray-800 dark:divide-gray-700">
          <tr v-for="peer in peers" :key="peer.id">
            <td class="px-6 py-4 whitespace-nowrap dark:text-white">{{ peer.name }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ peer.allowed_ips }}</td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="peer.enabled ? 'badge-green' : 'badge-gray'">
                {{ peer.enabled ? '已启用' : '已禁用' }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm space-x-2">
              <button @click="togglePeer(peer)" class="text-blue-600 hover:text-blue-800">
                {{ peer.enabled ? '禁用' : '启用' }}
              </button>
              <button @click="downloadConfig(peer)" class="text-green-600 hover:text-green-800">下载</button>
              <button @click="showQR(peer)" class="text-purple-600 hover:text-purple-800">二维码</button>
              <button @click="deletePeer(peer)" class="text-red-600 hover:text-red-800">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add Modal -->
    <div v-if="showAddModal" class="modal-overlay" @click.self="showAddModal = false">
      <div class="modal-content">
        <h3 class="text-lg font-semibold mb-4 dark:text-white">添加客户端</h3>
        <input v-model="newPeerName" type="text" placeholder="客户端名称" class="input-field mb-4" />
        <div class="flex justify-end space-x-2">
          <button @click="showAddModal = false" class="btn-secondary">取消</button>
          <button @click="addPeer" class="btn-primary">添加</button>
        </div>
      </div>
    </div>

    <!-- QR Modal -->
    <div v-if="qrPeer" class="modal-overlay" @click.self="qrPeer = null">
      <div class="modal-content text-center">
        <h3 class="text-lg font-semibold mb-4 dark:text-white">{{ qrPeer.name }}</h3>
        <img v-if="qrCodeUrl" :src="qrCodeUrl" class="mx-auto" />
        <div v-else class="py-8 text-gray-500">加载中...</div>
        <button @click="qrPeer = null" class="btn-secondary mt-4">关闭</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const peers = ref([])
const showAddModal = ref(false)
const newPeerName = ref('')
const qrPeer = ref(null)
const qrCodeUrl = ref('')

const loadPeers = async () => {
  try {
    const res = await axios.get('/api/peers')
    peers.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const addPeer = async () => {
  if (!newPeerName.value) return
  await axios.post('/api/peers', { name: newPeerName.value })
  newPeerName.value = ''
  showAddModal.value = false
  loadPeers()
}

const togglePeer = async (peer) => {
  await axios.post(`/api/peers/${peer.id}/toggle`, { enabled: !peer.enabled })
  loadPeers()
}

const deletePeer = async (peer) => {
  if (!confirm(`确定删除 ${peer.name}？`)) return
  await axios.delete(`/api/peers/${peer.id}`)
  loadPeers()
}

const downloadConfig = async (peer) => {
  try {
    const res = await axios.get(`/api/peers/${peer.id}/config`, { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `${peer.name}.conf`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  } catch (e) {
    console.error('下载配置失败', e)
  }
}

const showQR = async (peer) => {
  qrPeer.value = peer
  qrCodeUrl.value = ''
  try {
    const res = await axios.get(`/api/peers/${peer.id}/qrcode`, { responseType: 'blob' })
    qrCodeUrl.value = window.URL.createObjectURL(res.data)
  } catch (e) {
    console.error('获取二维码失败', e)
  }
}

onMounted(loadPeers)
</script>
