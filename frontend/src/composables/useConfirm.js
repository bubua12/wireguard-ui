import { reactive } from 'vue'

const defaults = {
  title: '请确认',
  message: '',
  type: 'danger', // danger | warning
  confirmText: '确定',
  cancelText: '取消'
}

export const confirmState = reactive({ show: false, ...defaults })
let resolver = null

const settle = (value) => {
  confirmState.show = false
  const done = resolver
  resolver = null
  if (done) done(value)
}

export const acceptConfirm = () => settle(true)
export const cancelConfirm = () => settle(false)

// confirm({ title, message, type, confirmText }) resolves true when accepted.
export const confirm = (options) => {
  if (resolver) settle(false)
  Object.assign(confirmState, defaults, options, { show: true })
  return new Promise((resolve) => {
    resolver = resolve
  })
}
