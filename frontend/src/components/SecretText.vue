<template>
  <span class="inline-flex min-w-0 max-w-full items-center gap-1">
    <span class="truncate" :class="{ 'select-none tracking-wider': !visible }">{{ display }}</span>
    <button
      type="button"
      tabindex="-1"
      class="shrink-0 p-0.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
      :title="visible ? '隐藏' : '显示'"
      :aria-label="visible ? '隐藏' : '显示'"
      @click="visible = !visible"
    >
      <Icon :name="visible ? 'eye-off' : 'eye'" class="h-4 w-4" />
    </button>
  </span>
</template>

<script setup>
import { computed, ref } from 'vue'
import Icon from './Icon.vue'

const props = defineProps({
  value: { type: String, default: '' }
})

const visible = ref(false)

const display = computed(() => {
  const v = props.value || '—'
  if (visible.value) return v
  return '•'.repeat(Math.max(10, Math.min(v.length, 22)))
})
</script>
