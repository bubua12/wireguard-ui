<template>
  <div class="flex items-start gap-3">
    <button
      type="button"
      role="switch"
      :aria-checked="modelValue"
      :aria-labelledby="labelId"
      :disabled="disabled"
      class="relative mt-0.5 inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50"
      :class="modelValue ? 'bg-blue-600' : 'bg-gray-300 dark:bg-gray-600'"
      @click="$emit('update:modelValue', !modelValue)"
    >
      <span
        class="inline-block h-5 w-5 transform rounded-full bg-white shadow transition-transform"
        :class="modelValue ? 'translate-x-[1.375rem]' : 'translate-x-0.5'"
      />
    </button>
    <span :id="labelId" class="min-w-0 cursor-pointer select-none" @click="!disabled && $emit('update:modelValue', !modelValue)">
      <span class="block text-sm font-medium text-gray-800 dark:text-gray-200"><slot name="label" /></span>
      <span v-if="$slots.hint" class="mt-0.5 block text-xs text-gray-500 dark:text-gray-400"><slot name="hint" /></span>
    </span>
  </div>
</template>

<script setup>
defineProps({
  modelValue: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false }
})
defineEmits(['update:modelValue'])

const labelId = `switch-${Math.random().toString(36).slice(2, 8)}`
</script>
