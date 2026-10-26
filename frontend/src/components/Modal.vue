<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="modelValue" class="modal-overlay" :class="{ 'z-[60]': elevated }" @mousedown.self="onBackdrop">
        <div
          ref="panel"
          class="modal-panel"
          :class="sizes[size] || sizes.md"
          role="dialog"
          aria-modal="true"
          :aria-label="title || undefined"
          tabindex="-1"
        >
          <header v-if="title || $slots.header" class="mb-4 flex items-start justify-between gap-4">
            <slot name="header">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ title }}</h3>
            </slot>
            <button
              type="button"
              class="-mr-2 -mt-1 rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-700 dark:hover:text-gray-200"
              aria-label="关闭"
              title="关闭"
              @click="close"
            >
              <Icon name="x" class="h-5 w-5" />
            </button>
          </header>
          <slot />
          <footer v-if="$slots.footer" class="mt-6 flex flex-wrap justify-end gap-2">
            <slot name="footer" />
          </footer>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script>
// Open modals, newest last: only the top one reacts to Esc.
const stack = []
</script>

<script setup>
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import Icon from './Icon.vue'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: '' },
  size: { type: String, default: 'md' },
  // Forms set this so a stray click on the backdrop cannot throw away input.
  persistent: { type: Boolean, default: false },
  // The confirm dialog lives in App.vue, which mounts before any page, so its
  // teleported overlay sits earlier in <body>; it needs the extra z-index to
  // appear above a modal that asks for confirmation.
  elevated: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue', 'close'])

const sizes = { sm: 'max-w-sm', md: 'max-w-md', lg: 'max-w-lg', xl: 'max-w-2xl' }
const panel = ref(null)
const token = {}
let previouslyFocused = null

const close = () => {
  emit('update:modelValue', false)
  emit('close')
}

const onBackdrop = () => {
  if (!props.persistent) close()
}

const onKeydown = (e) => {
  if (e.key === 'Escape' && stack[stack.length - 1] === token) {
    e.stopPropagation()
    close()
  }
}

const focusFirst = () => {
  const el = panel.value
  if (!el) return
  const target =
    el.querySelector('[data-autofocus]') ||
    el.querySelector('input:not([type=hidden]):not([disabled]), textarea:not([disabled]), select:not([disabled])') ||
    el
  target.focus({ preventScroll: true })
}

const open = async () => {
  previouslyFocused = document.activeElement
  stack.push(token)
  window.addEventListener('keydown', onKeydown)
  await nextTick()
  focusFirst()
}

const release = () => {
  const i = stack.indexOf(token)
  if (i < 0) return
  stack.splice(i, 1)
  window.removeEventListener('keydown', onKeydown)
  if (previouslyFocused && document.contains(previouslyFocused)) previouslyFocused.focus({ preventScroll: true })
  previouslyFocused = null
}

watch(
  () => props.modelValue,
  (isOpen) => (isOpen ? open() : release()),
  { immediate: true }
)

onBeforeUnmount(release)
</script>
