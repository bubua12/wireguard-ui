<template>
  <div class="space-y-5">
    <div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">客户端管理</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          共 {{ peers.length }} 个客户端，{{ counts.online }} 个在线
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-2 sm:flex-nowrap">
        <div class="relative min-w-0 flex-1 sm:flex-none">
          <Icon name="search" class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
          <input
            v-model="search"
            type="search"
            placeholder="搜索名称、IP 或公钥"
            aria-label="搜索客户端"
            class="input-field input-field-icon w-full sm:w-64"
          />
        </div>
        <button
          type="button"
          class="btn-secondary"
          :title="showAddresses ? '隐藏客户端来源地址' : '显示客户端来源地址'"
          :aria-pressed="showAddresses"
          @click="showAddresses = !showAddresses"
        >
          <Icon :name="showAddresses ? 'eye-off' : 'eye'" class="h-4 w-4" />
          <span class="hidden md:inline">{{ showAddresses ? '隐藏地址' : '显示地址' }}</span>
        </button>
        <button type="button" class="btn-primary" :disabled="noServer" @click="openAdd">
          <Icon name="plus" class="h-4 w-4" />添加客户端
        </button>
      </div>
    </div>

    <div v-if="noServer" class="card p-10 text-center">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">还没有配置服务器</h2>
      <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">需要先创建服务器（或导入现有配置），才能添加客户端。</p>
      <router-link to="/settings" class="btn-primary mt-5">前往设置</router-link>
    </div>

    <template v-else>
      <div class="flex flex-wrap gap-2" role="group" aria-label="状态筛选">
        <button
          v-for="f in filters"
          :key="f.key"
          type="button"
          class="chip"
          :class="{ 'chip-active': filter === f.key }"
          :aria-pressed="filter === f.key"
          @click="filter = f.key"
        >
          {{ f.label }}
          <span class="min-w-[2ch] text-center tabular-nums opacity-75">{{ counts[f.key] }}</span>
        </button>
      </div>

      <div class="card overflow-hidden">
        <!-- 桌面端：表格 -->
        <div class="hidden overflow-x-auto lg:block">
          <!-- 固定列宽、固定行高：握手时间和流量每 5 秒变一次，自动列宽会让整张表左右抖动 -->
          <table class="w-full table-fixed divide-y divide-gray-200 dark:divide-gray-700">
            <colgroup>
              <col />
              <col class="w-[140px]" />
              <col class="w-[88px]" />
              <col class="w-[168px]" />
              <col class="w-[200px]" />
              <col class="w-[216px]" />
            </colgroup>
            <thead class="bg-gray-50 dark:bg-gray-700/60">
              <tr>
                <th class="th">
                  <button type="button" class="sort-btn" @click="toggleSort('name')">
                    名称 <SortCarets :active="sortKey === 'name'" :order="sortOrder" />
                  </button>
                </th>
                <th class="th">
                  <button type="button" class="sort-btn" @click="toggleSort('ip')">
                    IP 地址 <SortCarets :active="sortKey === 'ip'" :order="sortOrder" />
                  </button>
                </th>
                <th class="th">
                  <button type="button" class="sort-btn" @click="toggleSort('status')">
                    状态 <SortCarets :active="sortKey === 'status'" :order="sortOrder" />
                  </button>
                </th>
                <th class="th">
                  <button type="button" class="sort-btn" @click="toggleSort('handshake')">
                    最近握手 <SortCarets :active="sortKey === 'handshake'" :order="sortOrder" />
                  </button>
                </th>
                <th class="th">
                  <button type="button" class="sort-btn" title="下行 = 客户端下载，上行 = 客户端上传" @click="toggleSort('traffic')">
                    流量 ↓ / ↑ <SortCarets :active="sortKey === 'traffic'" :order="sortOrder" />
                  </button>
                </th>
                <th class="th text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
              <template v-if="loading">
                <tr v-for="n in pageSize" :key="`sk-${n}`">
                  <td v-for="c in 6" :key="c" class="px-4 py-3">
                    <div class="flex h-9 items-center"><div class="skeleton h-4 w-full max-w-[7rem]"></div></div>
                  </td>
                </tr>
              </template>
              <tr v-else-if="sortedPeers.length === 0">
                <td colspan="6" class="px-6 py-12 text-center">
                  <p class="text-gray-500 dark:text-gray-400">{{ emptyMessage }}</p>
                  <button v-if="peers.length === 0" type="button" class="btn-primary mt-4" @click="openAdd">添加第一个客户端</button>
                  <button v-else type="button" class="btn-secondary mt-4" @click="clearFilters">清除筛选</button>
                </td>
              </tr>
              <!-- 用 v-else 接在骨架屏后面：数据先到、状态还没到时，骨架行和真实行不能同时出现 -->
              <template v-else>
                <tr v-for="peer in pagedPeers" :key="peer.id" class="hover:bg-gray-50/70 dark:hover:bg-gray-700/30">
                  <td class="px-4 py-3 text-gray-900 dark:text-white">
                    <div class="flex h-9 items-center">
                      <span class="status-dot" :class="dotClass(peer)"></span>
                      <span class="min-w-0 truncate font-medium" :title="peer.name">{{ peer.name }}</span>
                      <span v-if="!peer.has_private_key" class="badge-import ml-2 shrink-0" title="从系统配置导入，没有私钥，无法下载或扫码">导入</span>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                    <div class="flex h-9 items-center"><span class="truncate">{{ peer.allowed_ips }}</span></div>
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex h-9 items-center"><span :class="statusBadge(peer).class">{{ statusBadge(peer).label }}</span></div>
                  </td>
                  <td class="px-4 py-3 text-sm">
                    <div class="flex h-9 flex-col justify-center">
                      <div class="truncate leading-5 text-gray-700 dark:text-gray-200" :title="handshakeTitle(peer)">{{ handshakeText(peer) }}</div>
                      <div
                        class="truncate text-xs leading-4 text-gray-400 dark:text-gray-500"
                        :class="{ 'select-none tracking-wider': endpointOf(peer) && !showAddresses }"
                        :title="showAddresses ? endpointOf(peer) : undefined"
                      >
                        {{ endpointOf(peer) ? (showAddresses ? endpointOf(peer) : '••••••••') : '\u00a0' }}
                      </div>
                    </div>
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex h-9 items-center">
                      <button
                        type="button"
                        class="truncate text-left text-sm tabular-nums text-gray-700 hover:text-blue-600 dark:text-gray-200 dark:hover:text-blue-400"
                        title="累计流量（下行 / 上行），点击查看按天明细"
                        @click="openTraffic(peer)"
                      >
                        {{ formatBytes(downOf(statusOf(peer))) }} / {{ formatBytes(upOf(statusOf(peer))) }}
                      </button>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-sm">
                    <div class="flex h-9 items-center justify-end">
                      <PeerActions
                        :peer="peer"
                        @edit="openEdit(peer)"
                        @toggle="toggleEnabled(peer)"
                        @download="downloadConfig(peer)"
                        @config="openConfig(peer)"
                        @delete="deletePeer(peer)"
                      />
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <!-- 手机端：卡片 -->
        <div class="divide-y divide-gray-100 lg:hidden dark:divide-gray-700">
          <template v-if="loading">
            <!-- 用 padding 而不是 margin：相邻 margin 会塌陷，骨架会比真实卡片矮，还会把整个列表顶下去 -->
            <div v-for="n in pageSize" :key="`skc-${n}`" class="p-4"><div class="skeleton h-[172px]"></div></div>
          </template>
          <div v-else-if="sortedPeers.length === 0" class="px-4 py-10 text-center">
            <p class="text-gray-500 dark:text-gray-400">{{ emptyMessage }}</p>
            <button v-if="peers.length === 0" type="button" class="btn-primary mt-4" @click="openAdd">添加第一个客户端</button>
            <button v-else type="button" class="btn-secondary mt-4" @click="clearFilters">清除筛选</button>
          </div>
          <template v-else>
            <div v-for="peer in pagedPeers" :key="peer.id" class="space-y-3 p-4">
              <div class="flex items-center justify-between gap-3">
                <div class="flex min-w-0 items-center">
                  <span class="status-dot" :class="dotClass(peer)"></span>
                  <span class="truncate font-medium text-gray-900 dark:text-white">{{ peer.name }}</span>
                  <span v-if="!peer.has_private_key" class="badge-import ml-2">导入</span>
                </div>
                <span :class="statusBadge(peer).class">{{ statusBadge(peer).label }}</span>
              </div>
              <dl class="grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
                <dt class="text-gray-500 dark:text-gray-400">IP 地址</dt>
                <dd class="text-right text-gray-800 dark:text-gray-200">{{ peer.allowed_ips }}</dd>
                <dt class="text-gray-500 dark:text-gray-400">最近握手</dt>
                <dd class="text-right text-gray-800 dark:text-gray-200">{{ handshakeText(peer) }}</dd>
                <dt class="text-gray-500 dark:text-gray-400">来源地址</dt>
                <dd class="truncate text-right text-gray-800 dark:text-gray-200" :class="{ 'select-none tracking-wider': endpointOf(peer) && !showAddresses }">
                  {{ endpointOf(peer) ? (showAddresses ? endpointOf(peer) : '••••••••') : '—' }}
                </dd>
                <dt class="text-gray-500 dark:text-gray-400">流量 ↓ / ↑</dt>
                <dd class="text-right">
                  <button type="button" class="tabular-nums text-blue-600 dark:text-blue-400" @click="openTraffic(peer)">
                    {{ formatBytes(downOf(statusOf(peer))) }} / {{ formatBytes(upOf(statusOf(peer))) }}
                  </button>
                </dd>
              </dl>
              <PeerActions
                :peer="peer"
                @edit="openEdit(peer)"
                @toggle="toggleEnabled(peer)"
                @download="downloadConfig(peer)"
                @config="openConfig(peer)"
                @delete="deletePeer(peer)"
              />
            </div>
          </template>
        </div>

        <div v-if="sortedPeers.length > 0" class="border-t border-gray-200 px-4 py-3 dark:border-gray-700">
          <Pagination v-model:page="page" v-model:page-size="pageSize" :total="sortedPeers.length" :page-sizes="[10, 20, 50]" />
        </div>
      </div>
    </template>

    <!-- 添加客户端 -->
    <Modal v-model="addOpen" title="添加客户端" persistent>
      <form id="add-peer-form" class="space-y-4" @submit.prevent="addPeer">
        <div v-if="addError" role="alert" class="rounded-md border border-red-300 bg-red-50 p-3 text-sm text-red-700 dark:border-red-700 dark:bg-red-900/30 dark:text-red-200">
          {{ addError }}
        </div>
        <div>
          <label class="field-label" for="add-name">客户端名称</label>
          <input id="add-name" v-model="addForm.name" type="text" maxlength="64" placeholder="例如：张三的 iPhone" class="input-field" required />
        </div>
        <div>
          <label class="field-label" for="add-ip">IP 地址（选填）</label>
          <input
            id="add-ip"
            v-model="addForm.ip"
            type="text"
            :placeholder="nextIp ? `留空自动分配（下一个：${nextIp}）` : '例如 10.0.0.10/32'"
            class="input-field"
          />
          <p v-if="nextIpReason && !nextIp" class="field-hint text-red-600 dark:text-red-400">{{ nextIpReason }}</p>
          <p v-else class="field-hint">必须是服务器网段内的 x.x.x.x/32，留空则自动分配第一个空闲地址。</p>
        </div>
      </form>
      <template #footer>
        <button type="button" class="btn-secondary" @click="addOpen = false">取消</button>
        <button type="submit" form="add-peer-form" class="btn-primary" :disabled="adding || !addForm.name.trim()">
          <span v-if="adding" class="spinner"></span>添加
        </button>
      </template>
    </Modal>

    <!-- 编辑客户端 -->
    <Modal v-model="editOpen" title="编辑客户端" persistent>
      <form id="edit-peer-form" class="space-y-4" @submit.prevent="savePeer">
        <div v-if="editError" role="alert" class="rounded-md border border-red-300 bg-red-50 p-3 text-sm text-red-700 dark:border-red-700 dark:bg-red-900/30 dark:text-red-200">
          {{ editError }}
        </div>
        <div>
          <label class="field-label" for="edit-name">客户端名称</label>
          <input id="edit-name" v-model="editForm.name" type="text" maxlength="64" class="input-field" required />
        </div>
        <div>
          <label class="field-label" for="edit-ip">IP 地址</label>
          <input id="edit-ip" v-model="editForm.ip" type="text" class="input-field" required />
          <p v-if="ipChanged && editing?.has_private_key" class="field-hint text-amber-600 dark:text-amber-400">
            修改 IP 后，这个客户端必须重新下载配置或重新扫码，旧配置将无法使用。
          </p>
          <p v-else class="field-hint">服务器网段内的 x.x.x.x/32。</p>
        </div>
        <div>
          <label class="field-label" for="edit-keepalive">保活间隔（秒）</label>
          <input id="edit-keepalive" v-model.number="editForm.keepalive" type="number" min="0" max="65535" class="input-field" />
          <p class="field-hint">写入客户端配置的 PersistentKeepalive；0 表示关闭，位于 NAT 后的设备建议 25。修改后需重新下载配置才会生效。</p>
        </div>
      </form>
      <template #footer>
        <button type="button" class="btn-secondary" @click="editOpen = false">取消</button>
        <button type="submit" form="edit-peer-form" class="btn-primary" :disabled="saving">
          <span v-if="saving" class="spinner"></span>保存
        </button>
      </template>
    </Modal>

    <!-- 二维码 / 配置 -->
    <Modal v-model="configOpen" :title="configPeer?.name || ''">
      <div class="text-center">
        <div v-if="configLoading" class="py-16 text-gray-500 dark:text-gray-400"><span class="spinner mr-2"></span>加载中...</div>
        <img v-else-if="qrUrl" :src="qrUrl" alt="客户端配置二维码" class="mx-auto h-60 w-60 rounded-lg bg-white p-2 ring-1 ring-gray-200 dark:ring-gray-600" />
        <p v-if="!configLoading && qrUrl" class="mt-3 text-sm text-gray-500 dark:text-gray-400">用 WireGuard 手机 App 扫描二维码即可导入</p>
      </div>
      <div v-if="configText" class="mt-4">
        <button type="button" class="inline-flex items-center gap-1.5 text-sm text-blue-600 hover:underline dark:text-blue-400" @click="showConfigText = !showConfigText">
          <Icon name="document" class="h-4 w-4" />{{ showConfigText ? '隐藏配置内容' : '查看配置内容（含私钥）' }}
        </button>
        <pre v-if="showConfigText" class="mt-2 max-h-60 overflow-auto rounded-md bg-gray-50 p-3 text-left text-xs leading-relaxed text-gray-800 dark:bg-gray-900/60 dark:text-gray-200">{{ configText }}</pre>
      </div>
      <template #footer>
        <button type="button" class="btn-secondary" :disabled="!configText" @click="copyConfig">
          <Icon name="copy" class="h-4 w-4" />复制配置
        </button>
        <button type="button" class="btn-primary" :disabled="!configText" @click="saveConfigFile">
          <Icon name="download" class="h-4 w-4" />下载 .conf
        </button>
      </template>
    </Modal>

    <!-- 流量明细 -->
    <Modal v-model="trafficOpen" size="xl">
      <template #header>
        <div class="min-w-0">
          <h3 class="truncate text-lg font-semibold text-gray-900 dark:text-white">{{ trafficPeer?.name }} 流量明细</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            累计 ↓ {{ formatBytes(downOf(trafficReport)) }} / ↑ {{ formatBytes(upOf(trafficReport)) }}
          </p>
          <p class="text-xs text-gray-400 dark:text-gray-500">按天统计由后台每分钟采样累计，接口重启后不会丢失；只保留最近 90 天。</p>
        </div>
      </template>
      <div v-if="trafficLoading" class="py-10 text-center text-gray-500"><span class="spinner mr-2"></span>加载中...</div>
      <template v-else>
        <TrafficChart :days="trafficChartDays" :height="160" />
        <div v-if="trafficReport.days.length === 0" class="py-6 text-center text-sm text-gray-500 dark:text-gray-400">暂无按天数据，有流量后会自动记录。</div>
        <template v-else>
          <table class="mt-4 w-full table-fixed text-sm">
            <colgroup>
              <col class="w-[34%]" />
              <col class="w-[22%]" />
              <col class="w-[22%]" />
              <col class="w-[22%]" />
            </colgroup>
            <thead>
              <tr class="text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                <th class="py-2 pr-4">日期</th>
                <th class="py-2 pr-4 text-right">下行</th>
                <th class="py-2 pr-4 text-right">上行</th>
                <th class="py-2 text-right">合计</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
              <tr v-for="row in pagedTrafficDays" :key="row.day" class="dark:text-gray-200">
                <td class="whitespace-nowrap py-2 pr-4">{{ row.day }}</td>
                <td class="py-2 pr-4 text-right tabular-nums">{{ formatBytes(downOf(row)) }}</td>
                <td class="py-2 pr-4 text-right tabular-nums">{{ formatBytes(upOf(row)) }}</td>
                <td class="py-2 text-right tabular-nums">{{ formatBytes(downOf(row) + upOf(row)) }}</td>
              </tr>
            </tbody>
          </table>
          <div v-if="trafficReport.days.length > trafficPageSize" class="mt-3 border-t border-gray-200 pt-3 dark:border-gray-700">
            <Pagination v-model:page="trafficPage" :page-size="trafficPageSize" :total="trafficReport.days.length" unit="天" />
          </div>
        </template>
      </template>
      <template #footer>
        <button type="button" class="btn-secondary" @click="trafficOpen = false">关闭</button>
      </template>
    </Modal>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onDeactivated, ref, watch } from 'vue'
import axios from 'axios'
import Icon from '../components/Icon.vue'
import Modal from '../components/Modal.vue'
import Pagination from '../components/Pagination.vue'
import PeerActions from '../components/PeerActions.vue'
import SortCarets from '../components/SortCarets.vue'
import TrafficChart from '../components/TrafficChart.vue'
import { confirm } from '../composables/useConfirm'
import { usePolling } from '../composables/usePolling'
import { useToast } from '../composables/useToast'
import { errorMessage, errorMessageAsync } from '../utils/api'
import {
  copyText,
  downOf,
  fillDays,
  formatAge,
  formatBytes,
  formatDateTime,
  ipToNum,
  saveBlob,
  upOf
} from '../utils/format'

const toast = useToast()

const peers = ref([])
const peerStatus = ref({})
const loading = ref(true)
const noServer = ref(false)
const search = ref('')
const filter = ref('all')
const sortKey = ref('')
const sortOrder = ref('asc')
const page = ref(1)
const pageSize = ref(10)
const showAddresses = ref(false)

const filters = [
  { key: 'all', label: '全部' },
  { key: 'online', label: '在线' },
  { key: 'offline', label: '离线' },
  { key: 'disabled', label: '已禁用' }
]

const noteWarning = (data) => {
  if (data?.warning) toast.warning('已保存到数据库，但未能应用到接口：' + data.warning, 9000)
}

// ---------- loading ----------

const loadPeers = async () => {
  try {
    const res = await axios.get('/api/peers', { timeout: 10000 })
    peers.value = res.data || []
    noServer.value = false
  } catch (e) {
    if (e.response?.status === 404) noServer.value = true
    else if (loading.value) toast.error(errorMessage(e, '获取客户端列表失败'))
  }
}

const loadStatus = async () => {
  if (noServer.value) return
  try {
    const res = await axios.get('/api/peers/status', { timeout: 10000 })
    const map = {}
    for (const item of res.data || []) {
      map[item.public_key] = {
        online: !!item.online,
        rx: Number(item.rx) || 0,
        tx: Number(item.tx) || 0,
        endpoint: item.endpoint || '',
        age: item.handshake_age == null ? null : Number(item.handshake_age),
        handshake: Number(item.latest_handshake) || 0
      }
    }
    peerStatus.value = map
  } catch (e) {
    // transient; the next tick retries
  }
}

// The skeleton stays until the first status arrives too, so rows never flash
// "offline / 0 B" for clients that are actually connected.
usePolling(async () => {
  await loadPeers()
  await loadStatus()
  loading.value = false
}, 5000)

// ---------- derived state ----------

const statusOf = (peer) => peerStatus.value[peer.public_key]
const isOnline = (peer) => !!statusOf(peer)?.online
const endpointOf = (peer) => statusOf(peer)?.endpoint || ''

const stateOf = (peer) => {
  if (!peer.enabled) return 'disabled'
  return isOnline(peer) ? 'online' : 'offline'
}

const dotClass = (peer) => `status-${stateOf(peer)}`

const statusBadge = (peer) => {
  switch (stateOf(peer)) {
    case 'online':
      return { label: '在线', class: 'badge-green' }
    case 'offline':
      return { label: '离线', class: 'badge-gray' }
    default:
      return { label: '已禁用', class: 'badge-red' }
  }
}

const handshakeText = (peer) => {
  if (!peer.enabled) return '—'
  const age = statusOf(peer)?.age
  return age == null ? '—' : formatAge(age)
}
const handshakeTitle = (peer) => {
  const ts = statusOf(peer)?.handshake
  return ts ? formatDateTime(ts) : '尚未完成握手'
}

const counts = computed(() => {
  const c = { all: peers.value.length, online: 0, offline: 0, disabled: 0 }
  for (const p of peers.value) c[stateOf(p)]++
  return c
})

const emptyMessage = computed(() => (peers.value.length === 0 ? '暂无客户端，点击右上角添加' : '没有匹配的客户端'))

const clearFilters = () => {
  search.value = ''
  filter.value = 'all'
}

const filteredPeers = computed(() => {
  const q = search.value.trim().toLowerCase()
  return peers.value.filter((p) => {
    if (filter.value !== 'all' && stateOf(p) !== filter.value) return false
    if (!q) return true
    return (
      (p.name || '').toLowerCase().includes(q) ||
      (p.allowed_ips || '').toLowerCase().includes(q) ||
      (p.public_key || '').toLowerCase().includes(q)
    )
  })
})

const stateRank = { online: 2, offline: 1, disabled: 0 }

const sortedPeers = computed(() => {
  const list = [...filteredPeers.value]
  if (!sortKey.value) return list
  const dir = sortOrder.value === 'asc' ? 1 : -1
  const total = (p) => downOf(statusOf(p)) + upOf(statusOf(p))
  const age = (p) => {
    const a = statusOf(p)?.age
    return a == null || a < 0 || !p.enabled ? Infinity : a
  }
  return list.sort((a, b) => {
    switch (sortKey.value) {
      case 'name':
        return a.name.localeCompare(b.name, 'zh-CN') * dir
      case 'ip':
        return (ipToNum(a.allowed_ips) - ipToNum(b.allowed_ips)) * dir
      case 'status':
        return (stateRank[stateOf(a)] - stateRank[stateOf(b)]) * dir
      case 'handshake': {
        const x = age(a)
        const y = age(b)
        if (x === y) return 0
        return (x < y ? -1 : 1) * dir
      }
      case 'traffic':
        return (total(a) - total(b)) * dir
      default:
        return 0
    }
  })
})

const toggleSort = (key) => {
  if (sortKey.value !== key) {
    sortKey.value = key
    sortOrder.value = 'asc'
  } else if (sortOrder.value === 'asc') {
    sortOrder.value = 'desc'
  } else {
    sortKey.value = ''
    sortOrder.value = 'asc'
  }
}

const totalPages = computed(() => Math.max(1, Math.ceil(sortedPeers.value.length / pageSize.value)))

const pagedPeers = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return sortedPeers.value.slice(start, start + pageSize.value)
})

watch([sortedPeers, pageSize], () => {
  if (page.value > totalPages.value) page.value = totalPages.value
})
watch([sortKey, sortOrder, search, filter], () => {
  page.value = 1
})

// ---------- add ----------

const addOpen = ref(false)
const adding = ref(false)
const addForm = ref({ name: '', ip: '' })
const addError = ref('')
const nextIp = ref('')
const nextIpReason = ref('')

const openAdd = async () => {
  addForm.value = { name: '', ip: '' }
  addError.value = ''
  nextIp.value = ''
  nextIpReason.value = ''
  addOpen.value = true
  try {
    const res = await axios.get('/api/peers/next-ip', { timeout: 10000 })
    nextIp.value = res.data?.ip || ''
    nextIpReason.value = res.data?.reason || ''
  } catch (e) {
    // the hint is optional
  }
}

const addPeer = async () => {
  const name = addForm.value.name.trim()
  if (!name || adding.value) return
  const body = { name }
  if (addForm.value.ip.trim()) body.allowed_ips = addForm.value.ip.trim()

  adding.value = true
  addError.value = ''
  try {
    const res = await axios.post('/api/peers', body)
    addOpen.value = false
    noteWarning(res.data)
    toast.success(`已添加客户端「${name}」`)
    await loadPeers()
    await loadStatus()
    clearFilters()
    await nextTick()
    const idx = sortedPeers.value.findIndex((p) => p.id === res.data.id)
    if (idx >= 0) page.value = Math.floor(idx / pageSize.value) + 1
  } catch (e) {
    addError.value = errorMessage(e, '添加失败')
  } finally {
    adding.value = false
  }
}

// ---------- edit ----------

const editOpen = ref(false)
const saving = ref(false)
const editing = ref(null)
const editForm = ref({ name: '', ip: '', keepalive: 25 })
const editError = ref('')

const ipChanged = computed(() => !!editing.value && editForm.value.ip.trim() !== editing.value.allowed_ips)

const openEdit = (peer) => {
  editing.value = peer
  editForm.value = { name: peer.name, ip: peer.allowed_ips, keepalive: peer.persistent_keepalive }
  editError.value = ''
  editOpen.value = true
}

const savePeer = async () => {
  const peer = editing.value
  if (!peer || saving.value) return
  const name = editForm.value.name.trim()
  const ip = editForm.value.ip.trim()
  const keepalive = editForm.value.keepalive === '' || editForm.value.keepalive == null ? 0 : Number(editForm.value.keepalive)
  if (!name) {
    editError.value = '名称不能为空'
    return
  }
  if (!Number.isInteger(keepalive) || keepalive < 0 || keepalive > 65535) {
    editError.value = '保活间隔应为 0-65535 的整数'
    return
  }

  const body = {}
  if (name !== peer.name) body.name = name
  if (ip && ip !== peer.allowed_ips) body.allowed_ips = ip
  if (keepalive !== peer.persistent_keepalive) body.persistent_keepalive = keepalive
  if (Object.keys(body).length === 0) {
    editOpen.value = false
    return
  }

  saving.value = true
  editError.value = ''
  try {
    const res = await axios.put(`/api/peers/${peer.id}`, body)
    editOpen.value = false
    noteWarning(res.data)
    toast.success('已保存')
    await loadPeers()
  } catch (e) {
    editError.value = errorMessage(e, '保存失败')
  } finally {
    saving.value = false
  }
}

// ---------- enable / disable / delete ----------

const toggleEnabled = async (peer) => {
  if (peer.enabled) {
    const ok = await confirm({
      title: '禁用客户端',
      message: `禁用「${peer.name}」后会立即把它从服务器移除并断开连接，之后可以随时重新启用。`,
      type: 'warning',
      confirmText: '禁用'
    })
    if (!ok) return
  }
  try {
    const res = await axios.post(`/api/peers/${peer.id}/toggle`, { enabled: !peer.enabled })
    noteWarning(res.data)
    toast.success(peer.enabled ? `已禁用「${peer.name}」` : `已启用「${peer.name}」`)
    await loadPeers()
    await loadStatus()
  } catch (e) {
    toast.error(errorMessage(e, '操作失败'))
  }
}

const deletePeer = async (peer) => {
  const ok = await confirm({
    title: '删除客户端',
    message: `确定要删除「${peer.name}」吗？\n它的密钥和流量记录会一并删除，此操作不可恢复。`,
    type: 'danger',
    confirmText: '删除'
  })
  if (!ok) return
  try {
    const res = await axios.delete(`/api/peers/${peer.id}`)
    noteWarning(res.data)
    toast.success(`已删除「${peer.name}」`)
    await loadPeers()
  } catch (e) {
    toast.error(errorMessage(e, '删除失败'))
  }
}

// ---------- config / QR ----------

const configOpen = ref(false)
const configPeer = ref(null)
const configLoading = ref(false)
const qrUrl = ref('')
const configText = ref('')
const showConfigText = ref(false)

const revokeQr = () => {
  if (qrUrl.value) {
    window.URL.revokeObjectURL(qrUrl.value)
    qrUrl.value = ''
  }
}

// Bumped whenever the modal opens or closes so a slow response for a modal that
// is gone (or was reopened for another client) is dropped.
let configRequest = 0

// Secrets are wiped only after the fade-out so the content does not vanish
// while the modal is still visible.
const clearConfig = () => {
  if (configOpen.value) return
  configText.value = ''
  showConfigText.value = false
  revokeQr()
}

const filenameFrom = (res, fallback) => {
  const m = /filename="?([^";]+)"?/.exec(res.headers?.['content-disposition'] || '')
  return m ? m[1] : fallback
}

const fetchConfig = (peer) =>
  axios.get(`/api/peers/${peer.id}/config`, { responseType: 'text', transformResponse: [(d) => d] })

const openConfig = async (peer) => {
  if (!peer.has_private_key) {
    toast.warning('导入的客户端没有私钥，无法生成二维码')
    return
  }
  const request = ++configRequest
  revokeQr()
  configText.value = ''
  showConfigText.value = false
  configPeer.value = peer
  configLoading.value = true
  configOpen.value = true
  try {
    const [qr, cfg] = await Promise.all([
      axios.get(`/api/peers/${peer.id}/qrcode`, { responseType: 'blob' }),
      fetchConfig(peer)
    ])
    if (request !== configRequest) return
    qrUrl.value = window.URL.createObjectURL(qr.data)
    configText.value = cfg.data
  } catch (e) {
    if (request !== configRequest) return
    toast.error(await errorMessageAsync(e, '获取配置失败'))
    configOpen.value = false
  } finally {
    if (request === configRequest) configLoading.value = false
  }
}

const copyConfig = async () => {
  if (await copyText(configText.value)) toast.success('配置已复制到剪贴板')
  else toast.error('复制失败，请展开配置内容后手动复制')
}

const saveConfigFile = () => {
  if (!configPeer.value || !configText.value) return
  saveBlob(new Blob([configText.value], { type: 'text/plain;charset=utf-8' }), `${configPeer.value.name}.conf`)
}

const downloadConfig = async (peer) => {
  if (!peer.has_private_key) {
    toast.warning('导入的客户端没有私钥，无法生成配置')
    return
  }
  try {
    const res = await fetchConfig(peer)
    saveBlob(new Blob([res.data], { type: 'text/plain;charset=utf-8' }), filenameFrom(res, `${peer.name}.conf`))
  } catch (e) {
    toast.error(await errorMessageAsync(e, '下载配置失败'))
  }
}

// ---------- traffic ----------

const trafficOpen = ref(false)
const trafficPeer = ref(null)
const trafficLoading = ref(false)
const trafficReport = ref({ rx: 0, tx: 0, days: [], today: '' })
const trafficPage = ref(1)
const trafficPageSize = 10

const trafficChartDays = computed(() => fillDays(trafficReport.value.days, 14, trafficReport.value.today))
const trafficTotalPages = computed(() => Math.max(1, Math.ceil(trafficReport.value.days.length / trafficPageSize)))
const pagedTrafficDays = computed(() => {
  const start = (trafficPage.value - 1) * trafficPageSize
  return trafficReport.value.days.slice(start, start + trafficPageSize)
})

watch(trafficTotalPages, (n) => {
  if (trafficPage.value > n) trafficPage.value = n
})

let trafficRequest = 0

const clearTraffic = () => {
  if (trafficOpen.value) return
  trafficReport.value = { rx: 0, tx: 0, days: [], today: '' }
  trafficPage.value = 1
}

const openTraffic = async (peer) => {
  const request = ++trafficRequest
  trafficPeer.value = peer
  trafficLoading.value = true
  trafficReport.value = { rx: 0, tx: 0, days: [], today: '' }
  trafficPage.value = 1
  trafficOpen.value = true
  try {
    const res = await axios.get(`/api/peers/${peer.id}/traffic`)
    if (request !== trafficRequest) return
    trafficReport.value = {
      rx: Number(res.data?.rx) || 0,
      tx: Number(res.data?.tx) || 0,
      days: res.data?.days || [],
      today: res.data?.today || ''
    }
  } catch (e) {
    if (request !== trafficRequest) return
    toast.error(errorMessage(e, '获取流量失败'))
    trafficOpen.value = false
  } finally {
    if (request === trafficRequest) trafficLoading.value = false
  }
}

watch(configOpen, (open) => {
  if (!open) {
    configRequest++
    setTimeout(clearConfig, 300)
  }
})
watch(trafficOpen, (open) => {
  if (!open) {
    trafficRequest++
    setTimeout(clearTraffic, 300)
  }
})

onBeforeUnmount(revokeQr)

// The page stays cached by <KeepAlive> when the user switches tabs; a modal left
// open would keep its Esc handler alive behind the new page.
onDeactivated(() => {
  addOpen.value = false
  editOpen.value = false
  configOpen.value = false
  trafficOpen.value = false
})
</script>
