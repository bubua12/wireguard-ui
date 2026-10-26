<template>
  <section class="card p-5">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex min-w-0 items-center gap-3">
        <div
          class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg"
          :class="up ? 'bg-green-100 text-green-600 dark:bg-green-900/40 dark:text-green-300' : 'bg-amber-100 text-amber-600 dark:bg-amber-900/40 dark:text-amber-300'"
        >
          <Icon name="power" class="h-6 w-6" />
        </div>
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h2 class="truncate text-lg font-semibold text-gray-900 dark:text-white">{{ status.name || status.interface }}</h2>
            <span :class="up ? 'badge-green' : 'badge-amber'">{{ up ? '运行中' : '未启动' }}</span>
          </div>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            接口 <code class="rounded bg-gray-100 px-1 py-0.5 text-xs dark:bg-gray-700">{{ status.interface }}</code>
            · UDP {{ status.listen_port }}
          </p>
        </div>
      </div>

      <div class="flex flex-wrap gap-2">
        <template v-if="up">
          <button type="button" class="btn-secondary btn-sm" :disabled="!!busy" title="写入配置并热更新客户端列表，不会断开现有连接" @click="sync">
            <span v-if="busy === 'sync'" class="spinner"></span><Icon v-else name="refresh" class="h-4 w-4" />同步配置
          </button>
          <button type="button" class="btn-outline-warning btn-sm" :disabled="!!busy" title="停止后重新启动，应用地址 / MTU / NAT 的修改" @click="restart">
            <span v-if="busy === 'restart'" class="spinner"></span><Icon v-else name="power" class="h-4 w-4" />重启接口
          </button>
          <button type="button" class="btn-outline-danger btn-sm" :disabled="!!busy" @click="stop">
            <span v-if="busy === 'down'" class="spinner"></span><Icon v-else name="stop" class="h-4 w-4" />停止
          </button>
        </template>
        <button v-else type="button" class="btn-primary btn-sm" :disabled="!!busy" @click="start">
          <span v-if="busy === 'up'" class="spinner"></span><Icon v-else name="play" class="h-4 w-4" />启动接口
        </button>
      </div>
    </div>

    <div
      v-if="up && needsRestart"
      class="mt-4 flex flex-col gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800 sm:flex-row sm:items-center sm:justify-between dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-200"
    >
      <span class="flex items-start gap-2">
        <Icon name="warning" class="mt-0.5 h-4 w-4" />
        地址、MTU、NAT 的修改无法热更新，需要重启接口后才会生效。
      </span>
      <button type="button" class="btn-warning btn-sm shrink-0" :disabled="!!busy" @click="restart">立即重启</button>
    </div>
    <div
      v-else-if="!up"
      class="mt-4 flex items-start gap-2 rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-600 dark:border-gray-700 dark:bg-gray-900/30 dark:text-gray-300"
    >
      <Icon name="info" class="mt-0.5 h-4 w-4" />
      接口当前没有运行，客户端无法连接。点击「启动接口」会先写入最新配置，再执行 wg-quick up。
    </div>
  </section>
</template>

<script setup>
import { computed, ref } from 'vue'
import axios from 'axios'
import Icon from './Icon.vue'
import { confirm } from '../composables/useConfirm'
import { useToast } from '../composables/useToast'
import { errorMessage } from '../utils/api'

const props = defineProps({
  status: { type: Object, required: true },
  // Set by the settings page right after a save that needs a restart.
  pendingRestart: { type: Boolean, default: false }
})
// changed carries which action ran: up | sync | restart | down.
const emit = defineEmits(['changed'])

const toast = useToast()
const busy = ref('')

const up = computed(() => !!props.status.interface_up)
const needsRestart = computed(() => props.pendingRestart || !!props.status.restart_required)
const name = computed(() => props.status.interface || '')

const run = async (key, request, onSuccess) => {
  busy.value = key
  try {
    const res = await request()
    onSuccess(res.data || {})
    if (res.data?.warning) {
      toast.warning(`已完成，但 PostUp / PostDown 脚本报错：${res.data.warning}`, 12000)
    }
    emit('changed', key)
  } catch (e) {
    toast.error(errorMessage(e))
  } finally {
    busy.value = ''
  }
}

const start = () =>
  run('up', () => axios.post('/api/interface/up'), (d) => toast.success(d.message || '接口已启动'))

const sync = () =>
  run('sync', () => axios.post('/api/sync'), (d) => {
    if (d.restart_required) {
      toast.warning('配置已同步，但地址 / MTU 的修改需要「重启接口」才会生效')
    } else {
      toast.success('配置已同步到系统')
    }
  })

const restart = async () => {
  const ok = await confirm({
    title: '重启接口',
    message: `将先停止再启动 ${name.value}，所有客户端会断线几秒钟后自动重连。\n如果你正通过这条 VPN 访问本页面，页面会短暂失联。`,
    type: 'warning',
    confirmText: '重启'
  })
  if (ok) {
    await run('restart', () => axios.post('/api/interface/restart'), (d) => toast.success(d.message || '接口已重启'))
  }
}

const stop = async () => {
  const ok = await confirm({
    title: '停止接口',
    message: `停止 ${name.value} 后所有客户端会立即断线。\n如果你正通过这条 VPN 访问本页面，之后将无法再连到这里，需要从服务器上手动启动。`,
    type: 'danger',
    confirmText: '停止'
  })
  if (ok) {
    await run('down', () => axios.post('/api/interface/down'), (d) => toast.success(d.message || '接口已停止'))
  }
}
</script>
