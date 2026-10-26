<template>
  <div
    class="pointer-events-none fixed inset-x-4 top-4 z-[100] flex flex-col items-end gap-2 sm:left-auto sm:right-4 sm:w-96"
    aria-live="polite"
  >
    <TransitionGroup name="toast" tag="div" class="relative flex w-full flex-col gap-2">
      <div
        v-for="t in toastState.items"
        :key="t.id"
        class="pointer-events-auto flex items-start gap-3 rounded-lg border p-3 shadow-lg backdrop-blur"
        :class="styles[t.type].box"
        role="status"
      >
        <Icon :name="styles[t.type].icon" class="mt-0.5 h-5 w-5" :class="styles[t.type].icon_color" />
        <p class="min-w-0 flex-1 break-words text-sm font-medium" :class="styles[t.type].text">{{ t.message }}</p>
        <button
          type="button"
          class="shrink-0 rounded p-0.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
          aria-label="关闭提示"
          @click="dismiss(t.id)"
        >
          <Icon name="x" class="h-4 w-4" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup>
import Icon from './Icon.vue'
import { toastState, useToast } from '../composables/useToast'

const { dismiss } = useToast()

const styles = {
  success: {
    icon: 'check-circle',
    box: 'border-green-200 bg-green-50/95 dark:border-green-800 dark:bg-green-900/60',
    text: 'text-green-800 dark:text-green-100',
    icon_color: 'text-green-500'
  },
  error: {
    icon: 'x-circle',
    box: 'border-red-200 bg-red-50/95 dark:border-red-800 dark:bg-red-900/60',
    text: 'text-red-800 dark:text-red-100',
    icon_color: 'text-red-500'
  },
  warning: {
    icon: 'warning',
    box: 'border-yellow-200 bg-yellow-50/95 dark:border-yellow-800 dark:bg-yellow-900/60',
    text: 'text-yellow-800 dark:text-yellow-100',
    icon_color: 'text-yellow-500'
  },
  info: {
    icon: 'info',
    box: 'border-blue-200 bg-blue-50/95 dark:border-blue-800 dark:bg-blue-900/60',
    text: 'text-blue-800 dark:text-blue-100',
    icon_color: 'text-blue-500'
  }
}
</script>
