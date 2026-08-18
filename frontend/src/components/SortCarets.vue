<template>
  <span class="inline-flex flex-col justify-center leading-none" aria-hidden="true">
    <svg class="w-2.5 h-2.5" viewBox="0 0 10 6" fill="currentColor">
      <path d="M5 0.8L9.3 5.2H0.7L5 0.8Z" :class="caretClass('asc')" />
    </svg>
    <svg class="w-2.5 h-2.5 -mt-0.5" viewBox="0 0 10 6" fill="currentColor">
      <path d="M5 5.2L0.7 0.8H9.3L5 5.2Z" :class="caretClass('desc')" />
    </svg>
  </span>
</template>

<script setup>
const props = defineProps({
  active: { type: Boolean, default: false },
  order: { type: String, default: 'asc' }
})

const caretClass = (dir) => {
  if (!props.active) return 'text-gray-400/70 dark:text-gray-500'
  return props.order === dir
    ? 'text-blue-500 dark:text-blue-400'
    : 'text-gray-300 dark:text-gray-600'
}
</script>
