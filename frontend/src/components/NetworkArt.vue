<template>
  <svg viewBox="0 0 520 380" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true" focusable="false">
    <defs>
      <radialGradient id="net-hub-glow">
        <stop offset="0" stop-color="#fff" stop-opacity="0.35" />
        <stop offset="1" stop-color="#fff" stop-opacity="0" />
      </radialGradient>
    </defs>

    <g stroke="#fff">
      <circle :cx="hub.x" :cy="hub.y" r="92" stroke-opacity="0.12" />
      <circle :cx="hub.x" :cy="hub.y" r="150" stroke-opacity="0.12" stroke-dasharray="2 8" />
      <circle :cx="hub.x" :cy="hub.y" r="188" stroke-opacity="0.07" />
    </g>

    <!-- 每条连线：一条淡的底线，加一条流动的小点（数据包） -->
    <g stroke="#fff">
      <template v-for="(p, i) in peers" :key="i">
        <path :d="p.path" stroke-opacity="0.18" />
        <path
          :d="p.path"
          class="net-flow"
          :class="{ 'net-flow-rev': i % 2 === 1 }"
          :style="{ animationDuration: `${1 + (i % 3) * 0.35}s` }"
          stroke-opacity="0.9"
          stroke-width="2.5"
        />
      </template>
    </g>

    <!-- 服务器 -->
    <circle :cx="hub.x" :cy="hub.y" r="100" fill="url(#net-hub-glow)" />
    <circle class="net-pulse" :cx="hub.x" :cy="hub.y" r="40" stroke="#fff" stroke-opacity="0.5" />
    <circle class="net-pulse" style="animation-delay: -1.6s" :cx="hub.x" :cy="hub.y" r="40" stroke="#fff" stroke-opacity="0.5" />
    <circle :cx="hub.x" :cy="hub.y" r="40" fill="#fff" fill-opacity="0.12" stroke="#fff" stroke-opacity="0.3" />
    <circle :cx="hub.x" :cy="hub.y" r="29" fill="#fff" />
    <path
      :d="serverIcon"
      :transform="`translate(${hub.x - 13.2} ${hub.y - 13.2}) scale(1.1)`"
      stroke="#1d4ed8"
      stroke-width="1.8"
      stroke-linecap="round"
      stroke-linejoin="round"
    />

    <!-- 客户端 -->
    <g v-for="(p, i) in peers" :key="`node-${i}`">
      <circle :cx="p.x" :cy="p.y" :r="p.r + 8" fill="#fff" fill-opacity="0.07" stroke="#fff" stroke-opacity="0.2" />
      <circle
        v-if="p.online"
        class="net-pulse net-pulse-sm"
        :style="{ animationDelay: `-${i * 0.7}s` }"
        :cx="p.x"
        :cy="p.y"
        :r="p.r + 8"
        stroke="#4ade80"
        stroke-opacity="0.7"
      />
      <circle :cx="p.x" :cy="p.y" :r="p.r" fill="#fff" />
      <circle
        v-if="p.online"
        :cx="p.x + p.r * 0.8"
        :cy="p.y - p.r * 0.8"
        r="3.5"
        fill="#4ade80"
        stroke="#1e3a8a"
        stroke-width="1.5"
      />
      <text
        v-if="p.label"
        :x="p.x"
        :y="p.y + p.ly"
        text-anchor="middle"
        fill="#fff"
        fill-opacity="0.6"
        font-size="10.5"
        font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
      >
        {{ p.label }}
      </text>
    </g>
  </svg>
</template>

<script setup>
// Decorative hub-and-spoke picture for the login page: one server, several clients.
const hub = { x: 260, y: 190 }

// Same path as the "server" icon in Icon.vue (Heroicons, MIT).
const serverIcon =
  'M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01'

// ly: where the address label sits relative to the node; bend: how far the link curves.
const nodes = [
  { x: 82, y: 125, r: 8, online: true, label: '10.8.0.2', ly: -22, bend: 0.14 },
  { x: 221, y: 45, r: 6, bend: -0.12 },
  { x: 391, y: 59, r: 8, online: true, label: '10.8.0.3', ly: -22, bend: 0.12 },
  { x: 449, y: 174, r: 6, bend: -0.14 },
  { x: 375, y: 286, r: 8, online: true, label: '10.8.0.5', ly: 28, bend: 0.12 },
  { x: 234, y: 338, r: 6, bend: -0.12 },
  { x: 91, y: 287, r: 7, label: '10.8.0.4', ly: 28, bend: 0.14 }
]

// Quadratic curve from the client to the server; the control point is pushed sideways.
const peers = nodes.map((n) => {
  const cx = (n.x + hub.x) / 2 - (hub.y - n.y) * n.bend
  const cy = (n.y + hub.y) / 2 + (hub.x - n.x) * n.bend
  return { ...n, path: `M${n.x} ${n.y}Q${cx.toFixed(1)} ${cy.toFixed(1)} ${hub.x} ${hub.y}` }
})
</script>
