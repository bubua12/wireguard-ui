<template>
  <div v-if="total > 0" class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
    <div class="text-sm text-gray-500 dark:text-gray-400">
      共 {{ total }} {{ unit }}，第 {{ page }} / {{ totalPages }} 页
    </div>
    <div class="flex flex-wrap items-center gap-1">
      <div v-if="pageSizes.length" class="mr-1">
        <button
          ref="sizeBtn"
          type="button"
          class="inline-flex h-8 items-center gap-1.5 rounded-md border bg-white px-2.5 text-sm tabular-nums text-gray-700 transition-colors hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700"
          :class="sizeOpen ? 'border-blue-500 ring-2 ring-blue-500/30' : 'border-gray-300 dark:border-gray-600'"
          aria-haspopup="listbox"
          :aria-expanded="sizeOpen"
          aria-label="每页条数"
          @click="toggleSize"
          @keydown="onSizeKey"
        >
          {{ pageSize }} 条/页
          <svg
            viewBox="0 0 20 20"
            fill="currentColor"
            aria-hidden="true"
            class="h-4 w-4 text-gray-400 transition-transform"
            :class="{ 'rotate-180': sizeOpen }"
          >
            <path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.17l3.71-3.94a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd" />
          </svg>
        </button>
        <Teleport to="body">
          <div
            v-if="sizeOpen"
            ref="sizeMenu"
            class="fixed z-40 rounded-lg border border-gray-200 bg-white p-1 shadow-lg ring-1 ring-black/5 dark:border-gray-600 dark:bg-gray-800 dark:ring-white/10"
            :style="sizeStyle"
            role="listbox"
            aria-label="每页条数"
          >
            <button
              v-for="(n, i) in pageSizes"
              :key="n"
              type="button"
              role="option"
              class="flex w-full items-center justify-between gap-3 rounded-md px-2.5 py-1.5 text-left text-sm tabular-nums"
              :class="optionClass(n, i)"
              :aria-selected="n === pageSize"
              @mouseenter="activeIndex = i"
              @click="chooseSize(n)"
            >
              {{ n }} 条/页
              <svg v-if="n === pageSize" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true" class="h-4 w-4">
                <path fill-rule="evenodd" d="M16.7 5.3a1 1 0 010 1.4l-7.2 7.2a1 1 0 01-1.4 0L3.3 9.1a1 1 0 011.4-1.4l4.1 4.1 6.5-6.5a1 1 0 011.4 0z" clip-rule="evenodd" />
              </svg>
            </button>
          </div>
        </Teleport>
      </div>
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
import { computed, nextTick, onBeforeUnmount, onDeactivated, ref, watch } from 'vue'

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

// The list sits at the bottom of an overflow-hidden card, so a native select
// menu is clipped or uses the browser's own popup. This one is portaled and
// flips upward when the button is near the bottom of the window.
const sizeOpen = ref(false)
const sizeBtn = ref(null)
const sizeMenu = ref(null)
const sizeStyle = ref({ visibility: 'hidden' })
const activeIndex = ref(0)

const optionClass = (n, i) => {
  if (n === props.pageSize) return 'bg-blue-600 text-white'
  if (i === activeIndex.value) return 'bg-gray-100 text-gray-900 dark:bg-gray-700 dark:text-white'
  return 'text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-gray-700'
}

const placeSizeMenu = () => {
  const button = sizeBtn.value
  const menu = sizeMenu.value
  if (!button || !menu) return
  const rect = button.getBoundingClientRect()
  const height = menu.offsetHeight
  const width = Math.max(rect.width, menu.offsetWidth)
  const up = rect.bottom + height + 8 > window.innerHeight && rect.top > height + 8
  let left = rect.left
  if (left + width > window.innerWidth - 8) left = window.innerWidth - 8 - width
  sizeStyle.value = {
    left: `${Math.max(8, left)}px`,
    minWidth: `${Math.ceil(rect.width)}px`,
    top: `${Math.round(up ? rect.top - height - 6 : rect.bottom + 6)}px`,
    visibility: 'visible'
  }
}

const onDocPointer = (event) => {
  if (sizeBtn.value?.contains(event.target) || sizeMenu.value?.contains(event.target)) return
  sizeOpen.value = false
  // The menu floats over the table. Closing it on pointerdown reveals the row
  // underneath, and the click that follows would otherwise hit that button.
  const swallow = (click) => {
    click.preventDefault()
    click.stopPropagation()
    document.removeEventListener('click', swallow, true)
  }
  document.addEventListener('click', swallow, true)
  setTimeout(() => document.removeEventListener('click', swallow, true), 0)
}

const bindSizeListeners = (on) => {
  const method = on ? 'addEventListener' : 'removeEventListener'
  document[method]('pointerdown', onDocPointer, true)
  window[method]('resize', placeSizeMenu)
  window[method]('scroll', placeSizeMenu, true)
}

watch(sizeOpen, async (open) => {
  bindSizeListeners(false)
  if (!open) return
  activeIndex.value = Math.max(0, props.pageSizes.indexOf(props.pageSize))
  sizeStyle.value = { visibility: 'hidden' }
  await nextTick()
  if (!sizeOpen.value) return
  placeSizeMenu()
  bindSizeListeners(true)
})

const toggleSize = () => {
  sizeOpen.value = !sizeOpen.value
}

const chooseSize = (n) => {
  if (n !== props.pageSize) emit('update:pageSize', Number(n))
  sizeOpen.value = false
  sizeBtn.value?.focus()
}

const onSizeKey = (event) => {
  if (event.key === 'Escape') {
    if (!sizeOpen.value) return
    event.preventDefault()
    sizeOpen.value = false
    return
  }
  if (!sizeOpen.value && (event.key === 'Enter' || event.key === ' ' || event.key === 'ArrowDown')) {
    event.preventDefault()
    sizeOpen.value = true
    return
  }
  if (!sizeOpen.value) return
  const last = props.pageSizes.length - 1
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    activeIndex.value = Math.min(last, activeIndex.value + 1)
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    activeIndex.value = Math.max(0, activeIndex.value - 1)
  } else if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    chooseSize(props.pageSizes[activeIndex.value])
  }
}

onDeactivated(() => {
  sizeOpen.value = false
})
onBeforeUnmount(() => {
  sizeOpen.value = false
  bindSizeListeners(false)
})
</script>
