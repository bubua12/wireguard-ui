<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">服务器设置</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">配置 WireGuard 接口，以及下发给客户端的默认参数</p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <button
          v-if="!loading"
          type="button"
          class="btn-sm inline-flex items-center rounded-md bg-yellow-100 px-3 py-2 text-sm font-medium text-yellow-700 transition-colors hover:bg-yellow-200 dark:bg-yellow-900/30 dark:text-yellow-400 dark:hover:bg-yellow-900/50"
          @click="openImportModal"
        >
          <Icon name="upload" class="mr-1.5 h-4 w-4" />从系统配置文件导入
        </button>
        <button type="button" class="btn-secondary btn-sm" @click="openPwdModal">
          <Icon name="key" class="h-4 w-4" />修改密码
        </button>
      </div>
    </div>

    <div v-if="loading" class="space-y-4" aria-busy="true">
      <div class="skeleton h-40 w-full"></div>
      <div class="skeleton h-40 w-full"></div>
    </div>

    <template v-else>
      <InterfaceControls v-if="!isNew && status.configured" :status="status" :pending-restart="pendingRestart" @changed="onControlChanged" />

      <form class="space-y-6" @submit.prevent="save">
        <section class="card p-6">
          <h2 class="card-title">基本配置</h2>
          <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label class="field-label" for="s-name">显示名称</label>
              <input id="s-name" v-model="form.name" type="text" maxlength="64" placeholder="My VPN" class="input-field" />
              <p class="field-hint">只在界面里显示，留空则使用接口名</p>
            </div>
            <div>
              <label class="field-label" for="s-iface">接口名</label>
              <input id="s-iface" v-model.trim="form.interface" type="text" placeholder="wg0" class="input-field" required />
              <p class="field-hint">系统网卡名，一般是 wg0，不要填显示名称或域名</p>
            </div>
            <div>
              <label class="field-label" for="s-endpoint">公网地址</label>
              <SecretField id="s-endpoint" v-model="form.endpoint" placeholder="example.com:51820" required />
              <p class="field-hint">客户端连接用的 host:port，写入客户端配置</p>
            </div>
            <div>
              <label class="field-label" for="s-address">内网地址</label>
              <SecretField id="s-address" v-model="form.address" placeholder="10.0.0.1/24" required />
              <p class="field-hint">服务器在 VPN 网段里的地址（CIDR），客户端从这个网段分配</p>
            </div>
            <div>
              <label class="field-label" for="s-port">监听端口</label>
              <input id="s-port" v-model.number="form.listen_port" type="number" min="1" max="65535" class="input-field" required />
              <p class="field-hint">UDP 端口，记得在防火墙 / 安全组里放行</p>
            </div>
          </div>
        </section>

        <section class="card p-6">
          <h2 class="card-title">客户端默认配置</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">这些参数会写进下载或扫码得到的客户端配置，修改后客户端需要重新导入配置。</p>
          <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label class="field-label" for="s-dns">DNS</label>
              <input id="s-dns" v-model="form.dns" type="text" placeholder="8.8.8.8, 1.1.1.1" class="input-field" />
              <p class="field-hint">多个地址用逗号分隔；留空则不下发，客户端沿用自己的 DNS</p>
            </div>
            <div>
              <label class="field-label" for="s-mtu">MTU</label>
              <input id="s-mtu" v-model.number="form.mtu" type="number" min="0" max="9000" placeholder="1420" class="input-field" />
              <p class="field-hint">576–9000；填 0 由 wg-quick 自动决定</p>
            </div>
          </div>
          <div class="mt-6 grid grid-cols-1 gap-5 md:grid-cols-2">
            <ToggleSwitch v-model="form.full_tunnel">
              <template #label>全局流量</template>
              <template #hint>客户端所有流量都走隧道（AllowedIPs = 0.0.0.0/0, ::/0）；关闭则只有内网网段走隧道</template>
            </ToggleSwitch>
            <ToggleSwitch v-model="form.enable_nat">
              <template #label>启用 NAT / 转发</template>
              <template #hint>写入 iptables PostUp / PostDown，让客户端能经本机出网。系统还需要开启 net.ipv4.ip_forward</template>
            </ToggleSwitch>
          </div>
        </section>

        <section v-if="!isNew && form.public_key" class="card p-6">
          <h2 class="card-title">服务器公钥</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">手动配置客户端时，把它填进 [Peer] 的 PublicKey。</p>
          <button
            type="button"
            class="group mt-3 flex w-full items-center justify-between gap-3 rounded-md bg-gray-50 px-3 py-2 text-left font-mono text-sm text-gray-700 hover:bg-gray-100 dark:bg-gray-900/40 dark:text-gray-300 dark:hover:bg-gray-900/70"
            title="点击复制"
            @click="copyPublicKey"
          >
            <span class="break-all">{{ form.public_key }}</span>
            <Icon name="copy" class="h-4 w-4 shrink-0 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-200" />
          </button>
        </section>

        <div class="flex flex-wrap items-center gap-3">
          <button type="submit" class="btn-primary" :disabled="saving || (!isNew && !dirty)">
            <span v-if="saving" class="spinner"></span>{{ isNew ? '创建服务器' : '保存' }}
          </button>
          <button v-if="dirty && !isNew" type="button" class="btn-secondary" :disabled="saving" @click="resetForm">还原</button>
          <span v-if="dirty && !isNew" class="inline-flex items-center gap-1.5 text-sm text-amber-600 dark:text-amber-400">
            <Icon name="warning" class="h-4 w-4" />有未保存的修改
          </span>
          <p v-else-if="!isNew" class="text-sm text-gray-500 dark:text-gray-400">
            保存只写入数据库。要让修改生效，请用上方的「同步配置」；地址 / MTU / NAT 需要「重启接口」。
          </p>
        </div>
      </form>
    </template>

    <!-- 导入现有配置 -->
    <Modal v-model="showImportModal" title="导入现有配置" size="lg" persistent>
      <div class="mb-4 rounded-md border border-yellow-200 bg-yellow-50 px-3 py-2 text-sm text-yellow-800 dark:border-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-200">
        从 <code class="text-xs">/etc/wireguard/*.conf</code> 导入。会<strong>覆盖</strong>当前 UI 中的服务器和客户端。导入的客户端没有私钥，无法再下载配置或二维码。
      </div>
      <form id="import-form" class="space-y-4" @submit.prevent="confirmImport">
        <div>
          <label class="field-label" for="i-path">配置文件路径</label>
          <input id="i-path" v-model="importForm.config_path" type="text" placeholder="/etc/wireguard/wg0.conf" class="input-field" required />
        </div>
        <div>
          <label class="field-label" for="i-endpoint">公网地址（必填）</label>
          <SecretField id="i-endpoint" v-model="importForm.endpoint" placeholder="your-server.com:51820" required />
        </div>
        <div>
          <label class="field-label" for="i-dns">DNS</label>
          <input id="i-dns" v-model="importForm.dns" type="text" placeholder="8.8.8.8" class="input-field" />
        </div>
      </form>
      <template #footer>
        <button type="button" class="btn-secondary" @click="showImportModal = false">取消</button>
        <button type="submit" form="import-form" class="btn-warning" :disabled="importing">
          <span v-if="importing" class="spinner"></span>导入并覆盖
        </button>
      </template>
    </Modal>

    <!-- 修改密码 -->
    <Modal v-model="showPwdModal" title="修改密码" persistent>
      <form id="pwd-form" class="space-y-4" @submit.prevent="changePassword">
        <div>
          <label class="field-label" for="p-old">原密码</label>
          <SecretField id="p-old" v-model="pwdForm.old_password" as-password autocomplete="current-password" required />
        </div>
        <div>
          <label class="field-label" for="p-new">新密码</label>
          <SecretField id="p-new" v-model="pwdForm.new_password" as-password placeholder="至少 8 位" :minlength="8" required />
        </div>
        <div>
          <label class="field-label" for="p-confirm">确认新密码</label>
          <SecretField id="p-confirm" v-model="pwdForm.confirm_password" as-password placeholder="再输入一次" required />
        </div>
        <p class="field-hint">修改成功后，其他已登录的设备需要重新登录。</p>
      </form>
      <template #footer>
        <button type="button" class="btn-secondary" @click="showPwdModal = false">取消</button>
        <button type="submit" form="pwd-form" class="btn-primary" :disabled="changingPwd">
          <span v-if="changingPwd" class="spinner"></span>确认修改
        </button>
      </template>
    </Modal>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { onBeforeRouteLeave } from 'vue-router'
import axios from 'axios'
import Icon from '../components/Icon.vue'
import InterfaceControls from '../components/InterfaceControls.vue'
import Modal from '../components/Modal.vue'
import SecretField from '../components/SecretField.vue'
import ToggleSwitch from '../components/ToggleSwitch.vue'
import { confirm } from '../composables/useConfirm'
import { usePolling } from '../composables/usePolling'
import { useToast } from '../composables/useToast'
import { errorMessage } from '../utils/api'
import { setToken } from '../utils/auth'
import { copyText } from '../utils/format'

const toast = useToast()

const emptyForm = () => ({
  name: '',
  interface: 'wg0',
  endpoint: '',
  address: '',
  listen_port: 51820,
  dns: '8.8.8.8',
  mtu: 1420,
  full_tunnel: false,
  enable_nat: false,
  public_key: ''
})

const emptyPwd = () => ({ old_password: '', new_password: '', confirm_password: '' })
const emptyImport = () => ({
  config_path: '/etc/wireguard/wg0.conf',
  endpoint: '',
  dns: '8.8.8.8',
  full_tunnel: false,
  enable_nat: false
})

const form = ref(emptyForm())
const snapshot = ref('')
const isNew = ref(true)
const loading = ref(true)
const saving = ref(false)
const status = ref({ configured: false })
const pendingRestart = ref(false)

const showImportModal = ref(false)
const importForm = ref(emptyImport())
const importing = ref(false)
const showPwdModal = ref(false)
const pwdForm = ref(emptyPwd())
const changingPwd = ref(false)

// Only the editable fields take part in change detection.
const editable = (f) =>
  JSON.stringify([f.name, f.interface, f.endpoint, f.address, f.listen_port, f.dns, f.mtu, f.full_tunnel, f.enable_nat])
const dirty = computed(() => snapshot.value !== '' && editable(form.value) !== snapshot.value)

const applyServer = (data) => {
  form.value = {
    ...emptyForm(),
    name: data.name ?? '',
    interface: data.interface ?? 'wg0',
    endpoint: data.endpoint ?? '',
    address: data.address ?? '',
    listen_port: data.listen_port ?? 51820,
    dns: data.dns ?? '',
    mtu: data.mtu ?? 0,
    full_tunnel: !!data.full_tunnel,
    enable_nat: !!data.enable_nat,
    public_key: data.public_key ?? ''
  }
  snapshot.value = editable(form.value)
}

const resetForm = () => {
  const saved = JSON.parse(snapshot.value)
  Object.assign(form.value, {
    name: saved[0],
    interface: saved[1],
    endpoint: saved[2],
    address: saved[3],
    listen_port: saved[4],
    dns: saved[5],
    mtu: saved[6],
    full_tunnel: saved[7],
    enable_nat: saved[8]
  })
}

const loadStatus = async () => {
  try {
    const res = await axios.get('/api/status', { timeout: 10000 })
    status.value = res.data
  } catch (e) {
    // the interface card just keeps its last state
  }
}

const { refresh: refreshStatus } = usePolling(loadStatus, 10000)

onMounted(async () => {
  try {
    const res = await axios.get('/api/server')
    applyServer(res.data)
    isNew.value = false
  } catch (e) {
    if (e.response?.status !== 404) toast.error(errorMessage(e, '获取服务器配置失败'))
  } finally {
    loading.value = false
  }
})

onBeforeRouteLeave(() => {
  if (!dirty.value) return true
  return confirm({
    title: '放弃未保存的修改？',
    message: '服务器设置里有还没保存的修改，离开后会丢失。',
    type: 'warning',
    confirmText: '离开'
  })
})

const payload = () => ({
  name: form.value.name.trim(),
  interface: form.value.interface.trim(),
  endpoint: form.value.endpoint.trim(),
  address: form.value.address.trim(),
  listen_port: Number(form.value.listen_port) || 0,
  dns: form.value.dns.trim(),
  mtu: Number(form.value.mtu) || 0,
  full_tunnel: !!form.value.full_tunnel,
  enable_nat: !!form.value.enable_nat
})

const save = async () => {
  if (saving.value) return
  saving.value = true
  try {
    if (isNew.value) {
      const res = await axios.post('/api/server', payload())
      applyServer(res.data)
      isNew.value = false
      toast.success('服务器已创建。接下来点击「启动接口」让它生效。')
    } else {
      const res = await axios.put('/api/server', payload())
      applyServer(res.data)
      if (res.data.restart_required) {
        pendingRestart.value = true
        toast.warning('已保存。地址 / MTU / NAT 的修改需要「重启接口」后才会生效。', 8000)
      } else {
        toast.success('已保存。点击上方的「同步配置」让修改生效。')
      }
    }
    await refreshStatus()
  } catch (e) {
    toast.error(errorMessage(e, '保存失败'))
  } finally {
    saving.value = false
  }
}

const onControlChanged = async (action) => {
  if (action === 'restart' || action === 'up' || action === 'down') pendingRestart.value = false
  await refreshStatus()
}

const copyPublicKey = async () => {
  if (await copyText(form.value.public_key)) toast.success('服务器公钥已复制')
  else toast.error('复制失败，请手动选择复制')
}

// ---------- import ----------

const openImportModal = () => {
  importForm.value = emptyImport()
  showImportModal.value = true
}

const confirmImport = async () => {
  if (!importForm.value.endpoint.trim()) {
    toast.warning('请填写公网地址')
    return
  }
  const ok = await confirm({
    title: '覆盖导入',
    message: '导入会删除当前 UI 中的服务器和全部客户端，然后写入配置文件中的内容。确定继续？',
    type: 'danger',
    confirmText: '确认覆盖'
  })
  if (!ok) return

  importing.value = true
  try {
    const res = await axios.post('/api/import', importForm.value)
    toast.success(res.data.message, 8000)
    const serverRes = await axios.get('/api/server')
    applyServer(serverRes.data)
    isNew.value = false
    showImportModal.value = false
    await refreshStatus()
  } catch (e) {
    toast.error(errorMessage(e, '导入失败'))
  } finally {
    importing.value = false
  }
}

// ---------- password ----------

const openPwdModal = () => {
  pwdForm.value = emptyPwd()
  showPwdModal.value = true
}

const changePassword = async () => {
  const f = pwdForm.value
  if (!f.old_password || !f.new_password || !f.confirm_password) {
    toast.warning('请填写完整')
    return
  }
  if (f.new_password.length < 8) {
    toast.warning('新密码至少 8 位')
    return
  }
  if (f.new_password !== f.confirm_password) {
    toast.warning('两次输入的新密码不一致')
    return
  }
  changingPwd.value = true
  try {
    const res = await axios.post('/api/change-password', {
      old_password: f.old_password,
      new_password: f.new_password
    })
    // The old token was just invalidated; keep this session signed in.
    if (res.data?.token) setToken(res.data.token)
    toast.success('密码修改成功')
    showPwdModal.value = false
    pwdForm.value = emptyPwd()
  } catch (e) {
    toast.error(errorMessage(e, '修改失败'))
  } finally {
    changingPwd.value = false
  }
}
</script>
