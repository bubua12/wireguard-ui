<template>
  <div>
    <h1 class="text-2xl font-bold mb-6 dark:text-white">客户端管理</h1>

    <div class="flex justify-end mb-4">
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
            <td class="px-6 py-4 whitespace-nowrap dark:text-white">
              <div class="flex items-center">
                <span class="status-dot" :class="isOnline(peer.public_key) ? 'status-online' : 'status-offline'"></span>
                {{ peer.name }}
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ peer.allowed_ips }}</td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="peer.enabled ? 'badge-green' : 'badge-gray'">
                {{ peer.enabled ? '已启用' : '已禁用' }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
              <div class="flex justify-end space-x-1">
                <button @click="editPeer(peer)" class="btn-icon btn-icon-yellow" title="编辑">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                  </svg>
                </button>
                <button @click="confirmToggle(peer)" class="btn-icon" :class="peer.enabled ? 'btn-icon-blue' : 'btn-icon-green'" :title="peer.enabled ? '禁用' : '启用'">
                  <svg v-if="peer.enabled" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636"/>
                  </svg>
                  <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                  </svg>
                </button>
                <button @click="downloadConfig(peer)" class="btn-icon btn-icon-green" title="下载配置">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/>
                  </svg>
                </button>
                <button @click="showQR(peer)" class="btn-icon btn-icon-purple" title="二维码">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z"/>
                  </svg>
                </button>
                <button @click="confirmDelete(peer)" class="btn-icon btn-icon-red" title="删除">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add Modal -->
    <div v-if="showAddModal" class="modal-overlay">
      <div class="modal-content">
        <h3 class="text-lg font-semibold mb-4 dark:text-white">添加客户端</h3>
        <input v-model="newPeerName" type="text" placeholder="客户端名称" class="input-field mb-4" />
        <input v-model="newPeerIP" type="text" placeholder="IP地址（选填，如 10.0.8.100/32）" class="input-field mb-4" />
        <div class="flex justify-end space-x-2">
          <button @click="closeAddModal" class="btn-secondary">取消</button>
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

    <!-- Edit Modal -->
    <div v-if="editingPeer" class="modal-overlay">
      <div class="modal-content">
        <h3 class="text-lg font-semibold mb-4 dark:text-white">编辑客户端</h3>
        <input v-model="editPeerName" type="text" placeholder="客户端名称" class="input-field mb-4" />
        <div class="flex justify-end space-x-2">
          <button @click="editingPeer = null" class="btn-secondary">取消</button>
          <button @click="savePeer" class="btn-primary">保存</button>
        </div>
      </div>
    </div>

    <!-- Confirm Modal -->
    <div v-if="confirmModal.show" class="modal-overlay" @click.self="confirmModal.show = false">
      <div class="modal-content">
        <div class="flex items-center mb-4">
          <div class="w-10 h-10 rounded-full flex items-center justify-center mr-3" :class="confirmModal.type === 'danger' ? 'bg-red-100 dark:bg-red-900' : 'bg-yellow-100 dark:bg-yellow-900'">
            <svg class="w-6 h-6" :class="confirmModal.type === 'danger' ? 'text-red-600 dark:text-red-400' : 'text-yellow-600 dark:text-yellow-400'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
            </svg>
          </div>
          <h3 class="text-lg font-semibold dark:text-white">{{ confirmModal.title }}</h3>
        </div>
        <p class="text-gray-600 dark:text-gray-300 mb-6 ml-13">{{ confirmModal.message }}</p>
        <div class="flex justify-end space-x-2">
          <button @click="confirmModal.show = false" class="btn-secondary">取消</button>
          <button @click="confirmModal.onConfirm" :class="confirmModal.type === 'danger' ? 'btn-danger' : 'btn-warning'">
            {{ confirmModal.confirmText }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import axios from 'axios'

const peers = ref([])
const peerStatus = ref({})
const statusTimer = ref(null)
const showAddModal = ref(false)
const newPeerName = ref('')
const newPeerIP = ref('')
const qrPeer = ref(null)
const qrCodeUrl = ref('')
const editingPeer = ref(null)
const editPeerName = ref('')
const confirmModal = ref({
  show: false,
  title: '',
  message: '',
  type: 'danger',
  confirmText: '确定',
  onConfirm: () => {}
})

const loadPeers = async () => {
  try {
    const res = await axios.get('/api/peers')
    peers.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const loadStatus = async () => {
  try {
    const res = await axios.get('/api/peers/status')
    const statusMap = {}
    for (const item of res.data || []) {
      statusMap[item.public_key] = item.online
    }
    peerStatus.value = statusMap
  } catch (e) {
    console.error(e)
  }
}

const isOnline = (publicKey) => {
  return peerStatus.value[publicKey] || false
}

const addPeer = async () => {
  if (!newPeerName.value) return
  const data = { name: newPeerName.value }
  if (newPeerIP.value) {
    data.allowed_ips = newPeerIP.value
  }
  await axios.post('/api/peers', data)
  closeAddModal()
  loadPeers()
}

const closeAddModal = () => {
  showAddModal.value = false
  newPeerName.value = ''
  newPeerIP.value = ''
}

const editPeer = (peer) => {
  editingPeer.value = peer
  editPeerName.value = peer.name
}

const savePeer = async () => {
  if (!editPeerName.value) return
  await axios.put(`/api/peers/${editingPeer.value.id}`, { name: editPeerName.value })
  editingPeer.value = null
  loadPeers()
}

const confirmToggle = (peer) => {
  confirmModal.value = {
    show: true,
    title: peer.enabled ? '禁用客户端' : '启用客户端',
    message: `确定要${peer.enabled ? '禁用' : '启用'} "${peer.name}" 吗？`,
    type: 'warning',
    confirmText: peer.enabled ? '禁用' : '启用',
    onConfirm: () => doToggle(peer)
  }
}

const doToggle = async (peer) => {
  await axios.post(`/api/peers/${peer.id}/toggle`, { enabled: !peer.enabled })
  confirmModal.value.show = false
  loadPeers()
}

const confirmDelete = (peer) => {
  confirmModal.value = {
    show: true,
    title: '删除客户端',
    message: `确定要删除 "${peer.name}" 吗？此操作不可恢复。`,
    type: 'danger',
    confirmText: '删除',
    onConfirm: () => doDelete(peer)
  }
}

const doDelete = async (peer) => {
  await axios.delete(`/api/peers/${peer.id}`)
  confirmModal.value.show = false
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

onMounted(() => {
  loadPeers()
  loadStatus()
  statusTimer.value = setInterval(loadStatus, 5000)
})

onUnmounted(() => {
  if (statusTimer.value) {
    clearInterval(statusTimer.value)
  }
})
</script>
