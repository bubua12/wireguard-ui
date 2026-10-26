import { onActivated, onBeforeUnmount, onDeactivated, onMounted } from 'vue'

// Runs fn now and then every `interval` ms while the page is on screen. Ticks are
// skipped while the tab is hidden or the previous call is still running, and
// the tab refreshes as soon as it becomes visible again.
//
// Pages are cached by <KeepAlive>, so "on screen" means activated, not mounted:
// polling stops when the user switches away and restarts (with an immediate
// refresh) when they come back.
export const usePolling = (fn, interval = 5000) => {
  let timer = null
  let running = false
  let active = false

  const tick = async () => {
    if (running || document.hidden) return
    running = true
    try {
      await fn()
    } finally {
      running = false
    }
  }
  const onVisible = () => {
    if (!document.hidden) tick()
  }

  const start = () => {
    if (active) return
    active = true
    tick()
    timer = setInterval(tick, interval)
    document.addEventListener('visibilitychange', onVisible)
  }
  const stop = () => {
    active = false
    clearInterval(timer)
    timer = null
    document.removeEventListener('visibilitychange', onVisible)
  }

  // Inside <KeepAlive> activated also fires after the first mount; the guard in
  // start() makes the pair harmless, and mounted covers pages that are not cached.
  onMounted(start)
  onActivated(start)
  onDeactivated(stop)
  onBeforeUnmount(stop)

  // refresh ignores the in-flight guard so an explicit action always reloads.
  return { refresh: () => fn() }
}
