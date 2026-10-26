<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">仪表盘</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">状态每 5 秒自动刷新</p>
      </div>
      <span v-if="stale" class="badge-red" title="最近一次刷新失败，显示的是上次获取到的数据">
        <Icon name="warning" class="h-3.5 w-3.5" />无法获取最新状态
      </span>
    </div>

    <div v-if="loading" class="space-y-6" aria-busy="true">
      <div class="skeleton h-28 w-full"></div>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <div v-for="n in 4" :key="n" class="skeleton h-24"></div>
      </div>
      <div class="skeleton h-64 w-full"></div>
    </div>

    <div v-else-if="loadError" class="card p-8 text-center">
      <p class="text-gray-700 dark:text-gray-300">{{ loadError }}</p>
      <button type="button" class="btn-primary mt-4" @click="refreshAll">重试</button>
    </div>

    <div v-else-if="!status.configured" class="card p-10 text-center">
      <div class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-blue-100 text-blue-600 dark:bg-blue-900/40 dark:text-blue-300">
        <Icon name="server" class="h-7 w-7" />
      </div>
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">还没有配置服务器</h2>
      <p class="mx-auto mt-2 max-w-md text-sm text-gray-500 dark:text-gray-400">
        先到「设置」页创建一个 WireGuard 服务器，或者从 <code>/etc/wireguard</code> 里已有的配置文件导入。
      </p>
      <router-link to="/settings" class="btn-primary mt-5">前往设置</router-link>
    </div>

    <template v-else>
      <InterfaceControls :status="status" @changed="refreshAll" />

      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="客户端" icon="users" tone="blue">
          {{ status.peer_count }}
          <template #hint>已启用 {{ status.enabled_count }} 个</template>
        </StatCard>
        <StatCard label="当前在线" icon="signal" tone="green">
          {{ status.online_count }}
          <template #hint>3 分钟内有过握手</template>
        </StatCard>
        <StatCard label="实时速率" icon="chart" tone="purple">
          <template v-if="rate.ready">
            <span class="text-base font-semibold">↓ {{ formatRate(rate.down) }}</span>
            <span class="ml-2 text-base font-semibold">↑ {{ formatRate(rate.up) }}</span>
          </template>
          <span v-else class="text-gray-400">—</span>
          <template #hint>{{ status.interface_up ? '按两次刷新之间的流量估算' : '接口未运行' }}</template>
        </StatCard>
        <StatCard label="累计流量" icon="arrow-down" tone="amber">
          <span class="text-base font-semibold">↓ {{ formatBytes(downOf({ tx: status.transfer_tx })) }}</span>
          <span class="ml-2 text-base font-semibold">↑ {{ formatBytes(upOf({ rx: status.transfer_rx })) }}</span>
          <template #hint>接口重启后不会清零</template>
        </StatCard>
      </div>

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div class="space-y-6 lg:col-span-2">
          <section class="card p-5">
            <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 class="card-title">流量趋势</h2>
                <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
                  最近 {{ trafficDays.length }} 天合计 ↓ {{ formatBytes(trafficTotals.down) }} · ↑ {{ formatBytes(trafficTotals.up) }}
                </p>
              </div>
              <div class="inline-flex rounded-md bg-gray-100 p-0.5 dark:bg-gray-700" role="group" aria-label="时间范围">
                <button
                  v-for="n in ranges"
                  :key="n"
                  type="button"
                  class="rounded px-2.5 py-1 text-xs font-medium transition-colors"
                  :class="range === n ? 'bg-white text-blue-600 shadow-sm dark:bg-gray-600 dark:text-white' : 'text-gray-600 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white'"
                  @click="setRange(n)"
                >
                  {{ n }} 天
                </button>
              </div>
            </div>
            <TrafficChart :days="trafficDays" />
          </section>

          <section class="card">
            <div class="flex items-center justify-between gap-3 border-b border-gray-100 px-5 py-4 dark:border-gray-700">
              <div>
                <h2 class="card-title">在线客户端</h2>
                <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">按最近握手时间排序</p>
              </div>
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-xs text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700"
                @click="showAddresses = !showAddresses"
              >
                <Icon :name="showAddresses ? 'eye-off' : 'eye'" class="h-4 w-4" />
                {{ showAddresses ? '隐藏来源地址' : '显示来源地址' }}
              </button>
            </div>

            <p v-if="onlinePeers.length === 0" class="px-5 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
              {{ status.interface_up ? '当前没有在线的客户端' : '接口未运行，没有在线的客户端' }}
            </p>
            <ul v-else class="divide-y divide-gray-100 dark:divide-gray-700">
              <li v-for="p in onlinePeers" :key="p.id" class="flex flex-wrap items-center gap-x-4 gap-y-1 px-5 py-3">
                <span class="status-dot status-online"></span>
                <div class="min-w-0 flex-1 basis-40">
                  <div class="truncate font-medium text-gray-900 dark:text-white">{{ p.name }}</div>
                  <div class="truncate text-xs text-gray-500 dark:text-gray-400">
                    {{ p.allowed_ips }}
                    <template v-if="peerStatus[p.public_key]?.endpoint">
                      · <span :class="{ 'select-none tracking-wider': !showAddresses }">{{ showAddresses ? peerStatus[p.public_key].endpoint : '••••••••' }}</span>
                    </template>
                  </div>
                </div>
                <div class="text-right text-xs text-gray-500 tabular-nums dark:text-gray-400">
                  <div>握手 {{ formatAge(peerStatus[p.public_key]?.age) }}</div>
                  <div>↓ {{ formatBytes(downOf(peerStatus[p.public_key])) }} · ↑ {{ formatBytes(upOf(peerStatus[p.public_key])) }}</div>
                </div>
              </li>
            </ul>
            <div v-if="onlineTotal > onlinePeers.length" class="border-t border-gray-100 px-5 py-3 text-center text-sm dark:border-gray-700">
              <router-link to="/peers" class="text-blue-600 hover:underline dark:text-blue-400">
                还有 {{ onlineTotal - onlinePeers.length }} 个在线，查看全部客户端 →
              </router-link>
            </div>
          </section>
        </div>

        <section class="card h-fit p-5">
          <h2 class="card-title mb-4">服务器信息</h2>
          <dl class="space-y-3 text-sm">
            <div class="flex items-center justify-between gap-4">
              <dt class="shrink-0 text-gray-500 dark:text-gray-400">名称</dt>
              <dd class="truncate font-medium text-gray-900 dark:text-white">{{ status.name }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="shrink-0 text-gray-500 dark:text-gray-400">接口</dt>
              <dd class="font-medium text-gray-900 dark:text-white">{{ status.interface }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="shrink-0 text-gray-500 dark:text-gray-400">公网地址</dt>
              <dd class="min-w-0 text-gray-900 dark:text-white"><SecretText :value="status.endpoint" /></dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="shrink-0 text-gray-500 dark:text-gray-400">内网地址</dt>
              <dd class="min-w-0 text-gray-900 dark:text-white"><SecretText :value="status.address" /></dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="shrink-0 text-gray-500 dark:text-gray-400">监听端口</dt>
              <dd class="font-medium text-gray-900 dark:text-white">UDP {{ status.listen_port }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="shrink-0 text-gray-500 dark:text-gray-400">DNS</dt>
              <dd class="truncate text-gray-900 dark:text-white">{{ status.dns || '不下发' }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="shrink-0 text-gray-500 dark:text-gray-400">MTU</dt>
              <dd class="text-gray-900 dark:text-white">{{ status.mtu || '自动' }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="shrink-0 text-gray-500 dark:text-gray-400">客户端流量</dt>
              <dd><span :class="status.full_tunnel ? 'badge-amber' : 'badge-gray'">{{ status.full_tunnel ? '全局（0.0.0.0/0）' : '仅内网' }}</span></dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="shrink-0 text-gray-500 dark:text-gray-400">NAT 转发</dt>
              <dd><span :class="status.enable_nat ? 'badge-green' : 'badge-gray'">{{ status.enable_nat ? '已启用' : '未启用' }}</span></dd>
            </div>
            <div class="border-t border-gray-100 pt-3 dark:border-gray-700">
              <dt class="mb-1 text-gray-500 dark:text-gray-400">服务器公钥</dt>
              <dd>
                <button
                  type="button"
                  class="group flex w-full items-center justify-between gap-2 rounded-md bg-gray-50 px-2.5 py-1.5 text-left font-mono text-xs text-gray-700 hover:bg-gray-100 dark:bg-gray-900/40 dark:text-gray-300 dark:hover:bg-gray-900/70"
                  title="点击复制"
                  @click="copyKey"
                >
                  <span class="truncate">{{ status.public_key }}</span>
                  <Icon name="copy" class="h-4 w-4 shrink-0 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-200" />
                </button>
              </dd>
            </div>
          </dl>
        </section>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import axios from 'axios'
import Icon from '../components/Icon.vue'
import InterfaceControls from '../components/InterfaceControls.vue'
import SecretText from '../components/SecretText.vue'
import StatCard from '../components/StatCard.vue'
import TrafficChart from '../components/TrafficChart.vue'
import { usePolling } from '../composables/usePolling'
import { useToast } from '../composables/useToast'
import { errorMessage } from '../utils/api'
import { copyText, downOf, formatAge, formatBytes, formatRate, upOf } from '../utils/format'

const toast = useToast()

const ranges = [7, 14, 30]
const loadRange = () => {
  try {
    const n = Number(localStorage.getItem('dashboard.range'))
    return ranges.includes(n) ? n : 14
  } catch (e) {
    return 14
  }
}

const status = ref({ configured: false })
const peers = ref([])
const peerStatus = ref({})
const traffic = ref([])
const range = ref(loadRange())
const loading = ref(true)
const loadError = ref('')
const stale = ref(false)
const showAddresses = ref(false)
const rate = ref({ down: 0, up: 0, ready: false })

// ---------- status & live rate ----------

let lastSample = null
const smooth = (prev, next) => (prev > 0 ? prev * 0.4 + next * 0.6 : next)

const updateRate = (s) => {
  const now = performance.now()
  const down = downOf({ tx: s.transfer_tx })
  const up = upOf({ rx: s.transfer_rx })
  if (!s.interface_up) {
    rate.value = { down: 0, up: 0, ready: false }
    lastSample = null
    return
  }
  const dt = lastSample ? (now - lastSample.t) / 1000 : 0
  // Background tabs are throttled to about one poll a minute, which still gives
  // a usable average; after a longer gap it says nothing, so skip that sample.
  if (dt > 0.5 && dt < 120) {
    rate.value = {
      down: smooth(rate.value.down, Math.max(0, down - lastSample.down) / dt),
      up: smooth(rate.value.up, Math.max(0, up - lastSample.up) / dt),
      ready: true
    }
  }
  lastSample = { t: now, down, up }
}

const loadStatus = async () => {
  try {
    const res = await axios.get('/api/status', { timeout: 10000 })
    updateRate(res.data)
    status.value = res.data
    loadError.value = ''
    stale.value = false
  } catch (e) {
    if (loading.value) loadError.value = errorMessage(e, '获取状态失败')
    else stale.value = true
  } finally {
    loading.value = false
  }
}

const loadPeerStatus = async () => {
  if (!status.value.configured) return
  try {
    const res = await axios.get('/api/peers/status', { timeout: 10000 })
    const map = {}
    for (const item of res.data || []) {
      map[item.public_key] = {
        online: !!item.online,
        rx: Number(item.rx) || 0,
        tx: Number(item.tx) || 0,
        endpoint: item.endpoint || '',
        age: item.handshake_age == null ? null : Number(item.handshake_age)
      }
    }
    peerStatus.value = map
  } catch (e) {
    // the status call above already reports connectivity problems
  }
}

const loadLive = async () => {
  await loadStatus()
  await loadPeerStatus()
}

const loadPeers = async () => {
  try {
    const res = await axios.get('/api/peers', { timeout: 10000 })
    peers.value = res.data || []
  } catch (e) {
    // keep the previous list
  }
}

const loadTraffic = async () => {
  const requested = range.value
  try {
    const res = await axios.get('/api/traffic', { params: { days: requested }, timeout: 10000 })
    // A slower answer for a window the user already left must not replace the newer one.
    if (requested === range.value) traffic.value = res.data?.days || []
  } catch (e) {
    // keep the previous chart
  }
}

const { refresh: refreshLive } = usePolling(loadLive, 5000)
const { refresh: refreshPeers } = usePolling(loadPeers, 30000)
const { refresh: refreshTraffic } = usePolling(loadTraffic, 30000)

const refreshAll = async () => {
  await refreshLive()
  await Promise.all([refreshPeers(), refreshTraffic()])
}

const setRange = (n) => {
  range.value = n
  try {
    localStorage.setItem('dashboard.range', String(n))
  } catch (e) {
    // preference just won't persist
  }
}
watch(range, () => refreshTraffic())

// ---------- derived ----------

// The API already returns one row per day of the window, in server time.
const trafficDays = computed(() => traffic.value)
const trafficTotals = computed(() =>
  trafficDays.value.reduce(
    (acc, d) => ({ down: acc.down + downOf(d), up: acc.up + upOf(d) }),
    { down: 0, up: 0 }
  )
)

const onlineAll = computed(() =>
  peers.value
    .filter((p) => peerStatus.value[p.public_key]?.online)
    .sort((a, b) => (peerStatus.value[a.public_key].age ?? Infinity) - (peerStatus.value[b.public_key].age ?? Infinity))
)
const onlineTotal = computed(() => onlineAll.value.length)
const onlinePeers = computed(() => onlineAll.value.slice(0, 8))

const copyKey = async () => {
  if (await copyText(status.value.public_key)) toast.success('服务器公钥已复制')
  else toast.error('复制失败，请手动选择复制')
}
</script>
