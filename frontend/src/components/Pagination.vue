<template>
  <div v-if="total > 0" class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
    <div class="text-sm text-gray-500 dark:text-gray-400">
      共 {{ total }} {{ unit }}，第 {{ page }} / {{ totalPages }} 页
    </div>
    <div class="flex flex-wrap items-center gap-1">
      <select
        v-if="pageSizes.length"
        :value="pageSize"
        class="mr-1 h-8 rounded-md border border-gray-300 bg-white px-2 text-sm dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200"
        aria-label="每页条数"
        @change="$emit('update:pageSize', Number($event.target.value))"
      >
        <option v-for="n in pageSizes" :key="n" :value="n">{{ n }} 条/页</option>
      </select>
      <template v-if="totalPages > 1">
        <button type="button" class="pager-btn" :disabled="page <= 1" aria-label="第一页" @click="go(1)">«</button>
        <button type="button" class="pager-btn" :disabled="page <= 1" aria-label="上一页" @click="go(page - 1)">‹</button>
        <button
          v-for="item in items"
          :key="item.key"
          type="button"
          class="pager-btn"
          :class="{ 'pager-btn-active': item.current }"
          :disabled="item.ellipsis"
          @click="go(item.num)"
        >
          {{ item.label }}
        </button>
        <button type="button" class="pager-btn" :disabled="page >= totalPages" aria-label="下一页" @click="go(page + 1)">›</button>
        <button type="button" class="pager-btn" :disabled="page >= totalPages" aria-label="最后一页" @click="go(totalPages)">»</button>
      </template>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  page: { type: Number, required: true },
  pageSize: { type: Number, required: true },
  total: { type: Number, required: true },
  pageSizes: { type: Array, default: () => [] },
  unit: { type: String, default: '条' }
})
const emit = defineEmits(['update:page', 'update:pageSize'])

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

const go = (n) => {
  if (n == null) return
  emit('update:page', Math.min(Math.max(1, n), totalPages.value))
}

// 1 … 4 5 6 … 20
const items = computed(() => {
  const total = totalPages.value
  const current = props.page
  const nums = new Set([1, total, current - 1, current, current + 1])
  const list = [...nums].filter((n) => n >= 1 && n <= total).sort((a, b) => a - b)
  const out = []
  let prev = 0
  for (const n of list) {
    if (prev && n - prev > 1) out.push({ key: `e-${n}`, ellipsis: true, label: '…' })
    out.push({ key: `p-${n}`, num: n, label: String(n), current: n === current })
    prev = n
  }
  return out
})
</script>
