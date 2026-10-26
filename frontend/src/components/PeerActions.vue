<template>
  <div class="flex items-center justify-end gap-1">
    <button type="button" class="btn-icon btn-icon-yellow" title="编辑" aria-label="编辑" @click="$emit('edit')">
      <Icon name="pencil" class="h-4 w-4" />
    </button>
    <button
      type="button"
      class="btn-icon"
      :class="peer.enabled ? 'btn-icon-blue' : 'btn-icon-green'"
      :title="peer.enabled ? '禁用' : '启用'"
      :aria-label="peer.enabled ? '禁用' : '启用'"
      @click="$emit('toggle')"
    >
      <Icon :name="peer.enabled ? 'ban' : 'check-circle'" class="h-4 w-4" />
    </button>
    <button
      type="button"
      class="btn-icon btn-icon-green"
      :title="peer.has_private_key ? '下载配置' : '导入的客户端没有私钥，无法下载'"
      aria-label="下载配置"
      :disabled="!peer.has_private_key"
      @click="$emit('download')"
    >
      <Icon name="download" class="h-4 w-4" />
    </button>
    <button
      type="button"
      class="btn-icon btn-icon-purple"
      :title="peer.has_private_key ? '二维码 / 查看配置' : '导入的客户端没有私钥，无法生成二维码'"
      aria-label="二维码和配置"
      :disabled="!peer.has_private_key"
      @click="$emit('config')"
    >
      <Icon name="qrcode" class="h-4 w-4" />
    </button>
    <button type="button" class="btn-icon btn-icon-red" title="删除" aria-label="删除" @click="$emit('delete')">
      <Icon name="trash" class="h-4 w-4" />
    </button>
  </div>
</template>

<script setup>
import Icon from './Icon.vue'

defineProps({
  peer: { type: Object, required: true }
})
defineEmits(['edit', 'toggle', 'download', 'config', 'delete'])
</script>
