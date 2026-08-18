<template>
  <div>
    <h1 class="text-2xl font-bold mb-6 dark:text-white">仪表盘</h1>

    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
      <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800">
        <div class="text-gray-500 text-sm dark:text-gray-400">接口状态</div>
        <div class="text-2xl font-bold" :class="statusClass">{{ statusLabel }}</div>
      </div>
      <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800">
        <div class="text-gray-500 text-sm dark:text-gray-400">客户端总数</div>
        <div class="text-2xl font-bold dark:text-white">{{ status.peer_count || 0 }}</div>
      </div>
      <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800">
        <div class="text-gray-500 text-sm dark:text-gray-400">已启用 / 在线</div>
        <div class="text-2xl font-bold text-blue-500">{{ status.enabled_count || 0 }} / {{ status.online_count || 0 }}</div>
      </div>
      <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800">
        <div class="text-gray-500 text-sm dark:text-gray-400">流量 (↓ / ↑)</div>
        <div class="text-2xl font-bold dark:text-white">{{ formatBytes(status.transfer_rx) }} / {{ formatBytes(status.transfer_tx) }}</div>
      </div>
    </div>

    <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800" v-if="status.configured">
      <h2 class="text-lg font-semibold mb-4 dark:text-white">服务器信息</h2>
      <div class="grid grid-cols-2 gap-4 text-sm dark:text-gray-300">
        <div><span class="text-gray-500 dark:text-gray-400">名称：</span> {{ status.name }}</div>
        <div><span class="text-gray-500 dark:text-gray-400">接口：</span> {{ status.interface }}</div>
        <div><span class="text-gray-500 dark:text-gray-400">公网地址：</span> {{ status.endpoint }}</div>
        <div><span class="text-gray-500 dark:text-gray-400">内网地址：</span> {{ status.address }}</div>
        <div><span class="text-gray-500 dark:text-gray-400">端口：</span> {{ status.listen_port }}</div>
      </div>
    </div>
    <div v-else class="bg-white rounded-lg shadow p-6 dark:bg-gray-800 dark:text-gray-300">
      尚未配置服务器，请先到「设置」页创建或导入。
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import axios from 'axios'

const status = ref({
  configured: false,
  peer_count: 0,
  enabled_count: 0,
  online_count: 0,
  transfer_rx: 0,
  transfer_tx: 0
})
let timer = null

const statusLabel = computed(() => {
  if (!status.value.configured) return '未配置'
  return status.value.interface_up ? '运行中' : '未启动'
})

const statusClass = computed(() => {
  if (!status.value.configured) return 'text-gray-400'
  return status.value.interface_up ? 'text-green-500' : 'text-yellow-500'
})

const formatBytes = (n) => {
  const v = Number(n) || 0
  if (v < 1024) return v + ' B'
  if (v < 1024 * 1024) return (v / 1024).toFixed(1) + ' KB'
  if (v < 1024 * 1024 * 1024) return (v / 1024 / 1024).toFixed(1) + ' MB'
  return (v / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

const load = async () => {
  try {
    const res = await axios.get('/api/status')
    status.value = res.data
  } catch (e) {
    console.error(e)
  }
}

onMounted(() => {
  load()
  timer = setInterval(load, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>
