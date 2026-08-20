<template>
  <div>
    <Transition name="toast">
      <div v-if="toast.show" class="fixed top-4 right-4 z-50 max-w-sm">
        <div class="rounded-lg shadow-lg p-4 bg-yellow-50 border border-yellow-200 dark:bg-yellow-900/20 dark:border-yellow-800">
          <p class="text-yellow-800 dark:text-yellow-200">{{ toast.message }}</p>
        </div>
      </div>
    </Transition>

    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-6">
      <h1 class="text-2xl font-bold dark:text-white">客户端管理</h1>
      <div class="flex items-center gap-3 w-full sm:w-auto">
        <div class="relative flex-1 sm:flex-none">
          <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-4.35-4.35M17 10a7 7 0 11-14 0 7 7 0 0114 0z"/>
          </svg>
          <input
            v-model="search"
            type="search"
            placeholder="搜索名称或 IP"
            class="input-field input-field-icon w-full sm:w-56"
          />
        </div>
        <button @click="showAddModal = true" class="btn-primary whitespace-nowrap">添加客户端</button>
      </div>
    </div>

    <div class="bg-white rounded-lg shadow overflow-hidden dark:bg-gray-800">
      <div class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead class="bg-gray-50 dark:bg-gray-700">
          <tr>
            <th class="px-6 py-3 text-left">
              <button type="button" class="sort-btn" @click="toggleSort('name')">
                名称
                <SortCarets :active="sortKey === 'name'" :order="sortOrder" />
              </button>
            </th>
            <th class="px-6 py-3 text-left">
              <button type="button" class="sort-btn" @click="toggleSort('ip')">
                IP地址
                <SortCarets :active="sortKey === 'ip'" :order="sortOrder" />
              </button>
            </th>
            <th class="px-6 py-3 text-left">
              <button type="button" class="sort-btn" @click="toggleSort('status')">
                状态
                <SortCarets :active="sortKey === 'status'" :order="sortOrder" />
              </button>
            </th>
            <th class="px-6 py-3 text-left">
              <button type="button" class="sort-btn" @click="toggleSort('traffic')">
                流量 ↓ / ↑
                <SortCarets :active="sortKey === 'traffic'" :order="sortOrder" />
              </button>
            </th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase dark:text-gray-300">操作</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200 dark:bg-gray-800 dark:divide-gray-700">
          <tr v-if="sortedPeers.length === 0">
            <td colspan="5" class="px-6 py-10 text-center text-gray-500 dark:text-gray-400">
              {{ emptyMessage }}
            </td>
          </tr>
          <tr v-for="peer in pagedPeers" :key="peer.id">
            <td class="px-6 py-4 whitespace-nowrap dark:text-white">
              <div class="flex items-center">
                <span class="status-dot" :class="isOnline(peer.public_key) ? 'status-online' : 'status-offline'"></span>
                {{ peer.name }}
                <span
                  v-if="!peer.has_private_key"
                  class="badge-import ml-2"
                  title="从系统配置导入，没有私钥，无法下载或扫码"
                >导入</span>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ peer.allowed_ips }}</td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="peer.enabled ? 'badge-green' : 'badge-gray'">
                {{ peer.enabled ? '已启用' : '已禁用' }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <button
                type="button"
                class="text-left text-sm tabular-nums text-gray-700 dark:text-gray-200 hover:text-blue-600 dark:hover:text-blue-400"
                title="累计流量，点击查看按天明细"
                @click="openTraffic(peer)"
              >
                {{ formatBytes(peerRx(peer)) }} / {{ formatBytes(peerTx(peer)) }}
              </button>
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
                <button @click="downloadConfig(peer)" class="btn-icon btn-icon-green disabled:opacity-40" title="下载配置" :disabled="!peer.has_private_key">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/>
                  </svg>
                </button>
                <button @click="showQR(peer)" class="btn-icon btn-icon-purple disabled:opacity-40" title="二维码" :disabled="!peer.has_private_key">
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

      <div v-if="sortedPeers.length > 0" class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 px-4 py-3 border-t border-gray-200 dark:border-gray-700">
        <div class="text-sm text-gray-500 dark:text-gray-400">
          共 {{ sortedPeers.length }} 条，第 {{ page }} / {{ totalPages }} 页
        </div>
        <div class="flex items-center gap-2 flex-wrap">
          <select v-model.number="pageSize" class="h-8 px-2 text-sm rounded-md border border-gray-300 bg-white dark:bg-gray-800 dark:border-gray-600 dark:text-gray-200">
            <option :value="10">10 条/页</option>
            <option :value="20">20 条/页</option>
            <option :value="50">50 条/页</option>
          </select>
          <button type="button" class="pager-btn" :disabled="page <= 1" @click="page = 1">«</button>
          <button type="button" class="pager-btn" :disabled="page <= 1" @click="page -= 1">‹</button>
          <button
            v-for="item in pageItems"
            :key="item.key"
            type="button"
            class="pager-btn"
            :class="{ 'pager-btn-active': item.current }"
            :disabled="item.ellipsis"
            @click="!item.ellipsis && (page = item.num)"
          >{{ item.label }}</button>
          <button type="button" class="pager-btn" :disabled="page >= totalPages" @click="page += 1">›</button>
          <button type="button" class="pager-btn" :disabled="page >= totalPages" @click="page = totalPages">»</button>
        </div>
      </div>
    </div>

    <div v-if="showAddModal" class="modal-overlay">
      <div class="modal-content">
        <h3 class="text-lg font-semibold mb-4 dark:text-white">添加客户端</h3>
        <div v-if="addError" class="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded dark:bg-red-900 dark:border-red-700 dark:text-red-200">
          {{ addError }}
        </div>
        <input v-model="newPeerName" type="text" placeholder="客户端名称" class="input-field mb-4" />
        <input v-model="newPeerIP" type="text" placeholder="IP地址（选填，如 10.0.0.10/32）" class="input-field mb-4" />
        <div class="flex justify-end space-x-2">
          <button @click="closeAddModal" class="btn-secondary">取消</button>
          <button @click="addPeer" class="btn-primary">添加</button>
        </div>
      </div>
    </div>

    <div v-if="qrPeer" class="modal-overlay" @click.self="closeQR">
      <div class="modal-content text-center">
        <h3 class="text-lg font-semibold mb-4 dark:text-white">{{ qrPeer.name }}</h3>
        <img v-if="qrCodeUrl" :src="qrCodeUrl" class="mx-auto" />
        <div v-else class="py-8 text-gray-500">加载中...</div>
        <button @click="closeQR" class="btn-secondary mt-4">关闭</button>
      </div>
    </div>

    <div v-if="trafficPeer" class="modal-overlay" @click.self="closeTraffic">
      <div class="modal-content-lg max-h-[80vh] overflow-y-auto">
        <h3 class="text-lg font-semibold mb-1 dark:text-white">{{ trafficPeer.name }} 流量明细</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-1">
          累计 ↓ {{ formatBytes(trafficReport.rx) }} / ↑ {{ formatBytes(trafficReport.tx) }}
        </p>
        <p class="text-xs text-gray-400 dark:text-gray-500 mb-4">按天统计由后台采样累计，接口重启后不会丢失。</p>
        <div v-if="trafficLoading" class="py-8 text-center text-gray-500">加载中...</div>
        <div v-else-if="trafficReport.days.length === 0" class="py-8 text-center text-gray-500 dark:text-gray-400">
          暂无按天数据，有流量后会自动记录。
        </div>
        <table v-else class="min-w-full text-sm">
          <thead>
            <tr class="text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
              <th class="py-2 pr-4">日期</th>
              <th class="py-2 pr-4 text-right">下行</th>
              <th class="py-2 pr-4 text-right">上行</th>
              <th class="py-2 text-right">合计</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
            <tr v-for="row in trafficReport.days" :key="row.day" class="dark:text-gray-200">
              <td class="py-2 pr-4 whitespace-nowrap">{{ row.day }}</td>
              <td class="py-2 pr-4 text-right tabular-nums">{{ formatBytes(row.rx) }}</td>
              <td class="py-2 pr-4 text-right tabular-nums">{{ formatBytes(row.tx) }}</td>
              <td class="py-2 text-right tabular-nums">{{ formatBytes((Number(row.rx) || 0) + (Number(row.tx) || 0)) }}</td>
            </tr>
          </tbody>
        </table>
        <div class="flex justify-end mt-4">
          <button @click="closeTraffic" class="btn-secondary">关闭</button>
        </div>
      </div>
    </div>

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
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import axios from 'axios'
import SortCarets from '../components/SortCarets.vue'

const peers = ref([])
const peerStatus = ref({})
const statusTimer = ref(null)
const search = ref('')
const sortKey = ref('')
const sortOrder = ref('asc')
const page = ref(1)
const pageSize = ref(10)
const trafficPeer = ref(null)
const trafficLoading = ref(false)
const trafficReport = ref({ rx: 0, tx: 0, days: [] })
const showAddModal = ref(false)
const newPeerName = ref('')
const newPeerIP = ref('')
const addError = ref('')
const qrPeer = ref(null)
const qrCodeUrl = ref('')
const editingPeer = ref(null)
const editPeerName = ref('')
const toast = ref({ show: false, message: '' })
const confirmModal = ref({
  show: false,
  title: '',
  message: '',
  type: 'danger',
  confirmText: '确定',
  onConfirm: () => {}
})

const showToast = (message) => {
  toast.value = { show: true, message }
  setTimeout(() => { toast.value.show = false }, 4000)
}

const noteWarning = (data) => {
  if (data && data.warning) {
    showToast('已保存到数据库，但未能应用到接口：' + data.warning)
  }
}

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
      statusMap[item.public_key] = {
        online: !!item.online,
        rx: Number(item.rx) || 0,
        tx: Number(item.tx) || 0
      }
    }
    peerStatus.value = statusMap
  } catch (e) {
    console.error(e)
  }
}

const isOnline = (publicKey) => {
  return peerStatus.value[publicKey]?.online || false
}

const peerRx = (peer) => peerStatus.value[peer.public_key]?.rx || 0
const peerTx = (peer) => peerStatus.value[peer.public_key]?.tx || 0

const formatBytes = (n) => {
  const v = Number(n) || 0
  if (v < 1024) return v + ' B'
  if (v < 1024 * 1024) return (v / 1024).toFixed(1) + ' KB'
  if (v < 1024 * 1024 * 1024) return (v / 1024 / 1024).toFixed(1) + ' MB'
  return (v / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

const toggleSort = (key) => {
  if (sortKey.value !== key) {
    sortKey.value = key
    sortOrder.value = 'asc'
    return
  }
  if (sortOrder.value === 'asc') {
    sortOrder.value = 'desc'
    return
  }
  sortKey.value = ''
  sortOrder.value = 'asc'
}

const ipToNum = (cidr) => {
  const ip = (cidr || '').split('/')[0].split(',')[0].trim()
  const parts = ip.split('.').map(n => parseInt(n, 10))
  if (parts.length !== 4 || parts.some(n => Number.isNaN(n))) return 0
  return ((parts[0] << 24) >>> 0) + (parts[1] << 16) + (parts[2] << 8) + parts[3]
}

const filteredPeers = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return peers.value
  return peers.value.filter((p) => {
    const name = (p.name || '').toLowerCase()
    const ip = (p.allowed_ips || '').toLowerCase()
    return name.includes(q) || ip.includes(q)
  })
})

const emptyMessage = computed(() => {
  if (peers.value.length === 0) return '暂无客户端，点击右上角添加'
  return '没有匹配的客户端'
})

const sortedPeers = computed(() => {
  const list = [...filteredPeers.value]
  if (!sortKey.value) return list
  const dir = sortOrder.value === 'asc' ? 1 : -1
  return list.sort((a, b) => {
    if (sortKey.value === 'name') {
      return a.name.localeCompare(b.name) * dir
    }
    if (sortKey.value === 'ip') {
      return (ipToNum(a.allowed_ips) - ipToNum(b.allowed_ips)) * dir
    }
    if (sortKey.value === 'status') {
      const va = a.enabled ? 1 : 0
      const vb = b.enabled ? 1 : 0
      return (va - vb) * dir
    }
    if (sortKey.value === 'traffic') {
      return ((peerRx(a) + peerTx(a)) - (peerTx(b) + peerRx(b))) * dir
    }
    return 0
  })
})

const totalPages = computed(() => Math.max(1, Math.ceil(sortedPeers.value.length / pageSize.value)))

const pagedPeers = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return sortedPeers.value.slice(start, start + pageSize.value)
})

const pageItems = computed(() => {
  const total = totalPages.value
  const current = page.value
  const nums = new Set([1, total, current - 1, current, current + 1])
  const list = [...nums].filter(n => n >= 1 && n <= total).sort((a, b) => a - b)
  const items = []
  let prev = 0
  for (const n of list) {
    if (prev && n - prev > 1) {
      items.push({ key: `e-${n}`, ellipsis: true, label: '…' })
    }
    items.push({ key: `p-${n}`, num: n, label: String(n), current: n === current })
    prev = n
  }
  return items
})

watch([sortedPeers, pageSize], () => {
  if (page.value > totalPages.value) {
    page.value = totalPages.value
  }
})

watch([sortKey, sortOrder, search], () => {
  page.value = 1
})

const addPeer = async () => {
  if (!newPeerName.value) return
  addError.value = ''
  const data = { name: newPeerName.value }
  if (newPeerIP.value) {
    data.allowed_ips = newPeerIP.value
  }
  try {
    const res = await axios.post('/api/peers', data)
    noteWarning(res.data)
    closeAddModal()
    await loadPeers()
    if (!sortKey.value) {
      page.value = Math.max(1, Math.ceil(peers.value.length / pageSize.value))
    } else {
      page.value = 1
    }
  } catch (e) {
    addError.value = e.response?.data?.error || '添加失败'
  }
}

const closeAddModal = () => {
  showAddModal.value = false
  newPeerName.value = ''
  newPeerIP.value = ''
  addError.value = ''
}

const editPeer = (peer) => {
  editingPeer.value = peer
  editPeerName.value = peer.name
}

const savePeer = async () => {
  if (!editPeerName.value) return
  try {
    await axios.put(`/api/peers/${editingPeer.value.id}`, { name: editPeerName.value })
    editingPeer.value = null
    loadPeers()
  } catch (e) {
    showToast(e.response?.data?.error || '保存失败')
  }
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
  try {
    const res = await axios.post(`/api/peers/${peer.id}/toggle`, { enabled: !peer.enabled })
    noteWarning(res.data)
    confirmModal.value.show = false
    loadPeers()
  } catch (e) {
    showToast(e.response?.data?.error || '操作失败')
  }
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
  try {
    const res = await axios.delete(`/api/peers/${peer.id}`)
    noteWarning(res.data)
    confirmModal.value.show = false
    loadPeers()
  } catch (e) {
    showToast(e.response?.data?.error || '删除失败')
  }
}

const downloadConfig = async (peer) => {
  if (!peer.has_private_key) {
    showToast('导入的客户端没有私钥，无法生成配置')
    return
  }
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
    showToast(e.response?.data?.error || '下载配置失败')
  }
}

const closeTraffic = () => {
  trafficPeer.value = null
  trafficReport.value = { rx: 0, tx: 0, days: [] }
  trafficLoading.value = false
}

const openTraffic = async (peer) => {
  trafficPeer.value = peer
  trafficLoading.value = true
  trafficReport.value = { rx: 0, tx: 0, days: [] }
  try {
    const res = await axios.get(`/api/peers/${peer.id}/traffic`)
    if (trafficPeer.value?.id !== peer.id) return
    trafficReport.value = {
      rx: Number(res.data?.rx) || 0,
      tx: Number(res.data?.tx) || 0,
      days: res.data?.days || []
    }
  } catch (e) {
    if (trafficPeer.value?.id !== peer.id) return
    showToast(e.response?.data?.error || '获取流量失败')
    closeTraffic()
  } finally {
    if (trafficPeer.value?.id === peer.id) {
      trafficLoading.value = false
    }
  }
}

const closeQR = () => {
  qrPeer.value = null
  if (qrCodeUrl.value) {
    window.URL.revokeObjectURL(qrCodeUrl.value)
    qrCodeUrl.value = ''
  }
}

const showQR = async (peer) => {
  if (!peer.has_private_key) {
    showToast('导入的客户端没有私钥，无法生成二维码')
    return
  }
  closeQR()
  qrPeer.value = peer
  try {
    const res = await axios.get(`/api/peers/${peer.id}/qrcode`, { responseType: 'blob' })
    qrCodeUrl.value = window.URL.createObjectURL(res.data)
  } catch (e) {
    showToast(e.response?.data?.error || '获取二维码失败')
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
  closeQR()
  closeTraffic()
})
</script>
