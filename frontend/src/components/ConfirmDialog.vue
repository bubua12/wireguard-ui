<template>
  <Modal :model-value="confirmState.show" size="sm" elevated @update:model-value="onToggle">
    <div class="flex items-start gap-3">
      <div
        class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full"
        :class="danger ? 'bg-red-100 dark:bg-red-900/50' : 'bg-yellow-100 dark:bg-yellow-900/50'"
      >
        <Icon
          name="warning"
          class="h-6 w-6"
          :class="danger ? 'text-red-600 dark:text-red-400' : 'text-yellow-600 dark:text-yellow-400'"
        />
      </div>
      <div class="min-w-0 flex-1">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ confirmState.title }}</h3>
        <p class="mt-2 whitespace-pre-line text-sm text-gray-600 dark:text-gray-300">{{ confirmState.message }}</p>
      </div>
    </div>
    <template #footer>
      <button type="button" class="btn-secondary" @click="cancelConfirm">{{ confirmState.cancelText }}</button>
      <button type="button" :class="danger ? 'btn-danger' : 'btn-warning'" data-autofocus @click="acceptConfirm">
        {{ confirmState.confirmText }}
      </button>
    </template>
  </Modal>
</template>

<script setup>
import { computed } from 'vue'
import Icon from './Icon.vue'
import Modal from './Modal.vue'
import { acceptConfirm, cancelConfirm, confirmState } from '../composables/useConfirm'

const danger = computed(() => confirmState.type === 'danger')

// Esc or the close button count as "cancel".
const onToggle = (open) => {
  if (!open) cancelConfirm()
}
</script>
