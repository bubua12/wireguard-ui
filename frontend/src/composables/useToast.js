import { reactive } from 'vue'

const state = reactive({ items: [] })
let seq = 0

const dismiss = (id) => {
  const i = state.items.findIndex((t) => t.id === id)
  if (i >= 0) state.items.splice(i, 1)
}

const push = (type, message, duration) => {
  const id = ++seq
  state.items.push({ id, type, message })
  if (duration > 0) setTimeout(() => dismiss(id), duration)
  return id
}

export const toastState = state

export const useToast = () => ({
  success: (message, duration = 3500) => push('success', message, duration),
  info: (message, duration = 3500) => push('info', message, duration),
  warning: (message, duration = 6000) => push('warning', message, duration),
  error: (message, duration = 6000) => push('error', message, duration),
  dismiss
})
