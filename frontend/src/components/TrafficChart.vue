<template>
  <div ref="root" class="relative w-full select-none">
    <svg
      :width="width"
      :height="height"
      :viewBox="`0 0 ${width} ${height}`"
      role="img"
      :aria-label="`最近 ${days.length} 天的流量`"
      class="block touch-pan-y"
      @pointermove="onMove"
      @pointerdown="onMove"
      @pointerleave="hover = -1"
    >
      <g v-for="t in ticks" :key="t.value">
        <line
          :x1="pad.left"
          :x2="width - pad.right"
          :y1="t.y"
          :y2="t.y"
          stroke-dasharray="3 3"
          class="stroke-gray-200 dark:stroke-gray-700"
        />
        <text :x="pad.left - 8" :y="t.y + 3.5" text-anchor="end" class="fill-gray-500 text-[10px] dark:fill-gray-400">
          {{ t.label }}
        </text>
      </g>

      <rect
        v-if="hover >= 0"
        :x="pad.left + hover * groupW"
        :y="pad.top"
        :width="groupW"
        :height="innerH"
        class="fill-gray-500/10"
      />

      <g v-for="(bar, i) in bars" :key="bar.day">
        <rect v-if="bar.downH > 0" :x="bar.downX" :y="bar.downY" :width="barW" :height="bar.downH" rx="2" class="fill-blue-500" />
        <rect v-if="bar.upH > 0" :x="bar.upX" :y="bar.upY" :width="barW" :height="bar.upH" rx="2" class="fill-emerald-500" />
        <text
          v-if="bar.label"
          :x="pad.left + i * groupW + groupW / 2"
          :y="height - 6"
          text-anchor="middle"
          class="fill-gray-500 text-[10px] dark:fill-gray-400"
        >
          {{ bar.label }}
        </text>
      </g>
    </svg>

    <div
      v-if="hover >= 0 && days[hover]"
      class="pointer-events-none absolute top-0 z-10 w-max -translate-x-1/2 rounded-md bg-gray-900/95 px-2.5 py-1.5 text-xs text-white shadow-lg"
      :style="{ left: tooltipLeft + 'px' }"
    >
      <div class="font-medium">{{ days[hover].day }}</div>
      <div class="mt-0.5 flex items-center gap-1.5 tabular-nums">
        <span class="h-2 w-2 rounded-full bg-blue-500"></span>下行 {{ formatBytes(downOf(days[hover])) }}
      </div>
      <div class="flex items-center gap-1.5 tabular-nums">
        <span class="h-2 w-2 rounded-full bg-emerald-500"></span>上行 {{ formatBytes(upOf(days[hover])) }}
      </div>
    </div>

    <div v-if="empty" class="pointer-events-none absolute inset-0 flex items-center justify-center text-sm text-gray-400 dark:text-gray-500">
      暂无流量数据，有流量后会自动记录
    </div>

    <div class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
      <span class="inline-flex items-center gap-1.5"><span class="h-2.5 w-2.5 rounded-sm bg-blue-500"></span>下行（客户端下载）</span>
      <span class="inline-flex items-center gap-1.5"><span class="h-2.5 w-2.5 rounded-sm bg-emerald-500"></span>上行（客户端上传）</span>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { downOf, formatBytes, formatBytesCompact, upOf } from '../utils/format'

const props = defineProps({
  // Oldest first, one row per calendar day: { day: 'YYYY-MM-DD', rx, tx }.
  days: { type: Array, default: () => [] },
  height: { type: Number, default: 200 }
})

const pad = { top: 10, right: 8, bottom: 24, left: 54 }
const root = ref(null)
const width = ref(600)
const hover = ref(-1)
let observer = null

onMounted(() => {
  const measure = () => {
    if (root.value) width.value = Math.max(240, Math.floor(root.value.clientWidth))
  }
  measure()
  if (typeof ResizeObserver !== 'undefined') {
    observer = new ResizeObserver(measure)
    observer.observe(root.value)
  }
})
onBeforeUnmount(() => observer?.disconnect())

const innerW = computed(() => width.value - pad.left - pad.right)
const innerH = computed(() => props.height - pad.top - pad.bottom)
const groupW = computed(() => innerW.value / Math.max(1, props.days.length))
const barW = computed(() => Math.max(2, Math.min(14, groupW.value * 0.34)))
const empty = computed(() => props.days.every((d) => downOf(d) === 0 && upOf(d) === 0))

// Round the top of the axis up to 1/2/5 × 10^n of the largest unit in play so
// the labels read as clean numbers (e.g. 200 MB instead of 187.4 MB).
const niceMax = computed(() => {
  const max = Math.max(0, ...props.days.map((d) => Math.max(downOf(d), upOf(d))))
  if (max <= 0) return 1024
  let unit = 1
  while (max / unit >= 1024) unit *= 1024
  const scaled = max / unit
  const step = [1, 2, 5, 10, 20, 50, 100, 200, 500, 1000].find((s) => s >= scaled) ?? 1024
  return step * unit
})

const yFor = (bytes) => pad.top + innerH.value - (bytes / niceMax.value) * innerH.value

const ticks = computed(() =>
  [0, 0.5, 1].map((f) => {
    const value = niceMax.value * f
    return { value, y: yFor(value), label: value === 0 ? '0' : formatBytesCompact(value) }
  })
)

const labelEvery = computed(() => Math.max(1, Math.ceil(46 / groupW.value)))

const bars = computed(() =>
  props.days.map((d, i) => {
    const down = downOf(d)
    const up = upOf(d)
    const mid = pad.left + i * groupW.value + groupW.value / 2
    const h = (v) => (v > 0 ? Math.max(1.5, (v / niceMax.value) * innerH.value) : 0)
    const downH = h(down)
    const upH = h(up)
    const last = i === props.days.length - 1
    const show = (props.days.length - 1 - i) % labelEvery.value === 0
    return {
      day: d.day,
      downH,
      upH,
      downX: mid - barW.value - 0.5,
      upX: mid + 0.5,
      downY: pad.top + innerH.value - downH,
      upY: pad.top + innerH.value - upH,
      label: show || last ? d.day.slice(5) : ''
    }
  })
)

const onMove = (e) => {
  const rect = e.currentTarget.getBoundingClientRect()
  const i = Math.floor((e.clientX - rect.left - pad.left) / groupW.value)
  hover.value = i >= 0 && i < props.days.length ? i : -1
}

const tooltipLeft = computed(() => {
  const center = pad.left + hover.value * groupW.value + groupW.value / 2
  return Math.min(Math.max(center, 70), width.value - 70)
})
</script>
