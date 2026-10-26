<template>
  <div class="card flex items-start gap-4 p-5">
    <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg" :class="tones[tone]">
      <Icon :name="icon" class="h-6 w-6" />
    </div>
    <div class="min-w-0 flex-1">
      <div class="text-sm text-gray-500 dark:text-gray-400">{{ label }}</div>
      <div class="mt-0.5 truncate text-2xl font-bold tabular-nums text-gray-900 dark:text-white">
        <slot />
      </div>
      <div v-if="$slots.hint" class="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">
        <slot name="hint" />
      </div>
    </div>
  </div>
</template>

<script setup>
import Icon from './Icon.vue'

defineProps({
  label: { type: String, required: true },
  icon: { type: String, required: true },
  tone: { type: String, default: 'blue' }
})

const tones = {
  blue: 'bg-blue-100 text-blue-600 dark:bg-blue-900/40 dark:text-blue-300',
  green: 'bg-green-100 text-green-600 dark:bg-green-900/40 dark:text-green-300',
  purple: 'bg-purple-100 text-purple-600 dark:bg-purple-900/40 dark:text-purple-300',
  amber: 'bg-amber-100 text-amber-600 dark:bg-amber-900/40 dark:text-amber-300',
  gray: 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-300'
}
</script>
