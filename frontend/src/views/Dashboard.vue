<template>
  <div>
    <h1 class="text-2xl font-bold mb-6 dark:text-white">仪表盘</h1>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
      <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800">
        <div class="text-gray-500 text-sm dark:text-gray-400">服务器状态</div>
        <div class="text-2xl font-bold text-green-500">运行中</div>
      </div>
      <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800">
        <div class="text-gray-500 text-sm dark:text-gray-400">客户端总数</div>
        <div class="text-2xl font-bold dark:text-white">{{ peers.length }}</div>
      </div>
      <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800">
        <div class="text-gray-500 text-sm dark:text-gray-400">已启用</div>
        <div class="text-2xl font-bold text-blue-500">{{ activePeers }}</div>
      </div>
    </div>

    <div class="bg-white rounded-lg shadow p-6 dark:bg-gray-800" v-if="server">
      <h2 class="text-lg font-semibold mb-4 dark:text-white">服务器信息</h2>
      <div class="grid grid-cols-2 gap-4 text-sm dark:text-gray-300">
        <div><span class="text-gray-500 dark:text-gray-400">名称：</span> {{ server.name }}</div>
        <div><span class="text-gray-500 dark:text-gray-400">公网地址：</span> {{ server.endpoint }}</div>
        <div><span class="text-gray-500 dark:text-gray-400">内网地址：</span> {{ server.address }}</div>
        <div><span class="text-gray-500 dark:text-gray-400">端口：</span> {{ server.listen_port }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'

const server = ref(null)
const peers = ref([])

const activePeers = computed(() => peers.value.filter(p => p.enabled).length)

onMounted(async () => {
  try {
    const [serverRes, peersRes] = await Promise.all([
      axios.get('/api/server'),
      axios.get('/api/peers')
    ])
    server.value = serverRes.data
    peers.value = peersRes.data || []
  } catch (e) {
    console.error(e)
  }
})
</script>
