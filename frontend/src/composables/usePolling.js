import { onMounted, onUnmounted } from 'vue'

// Runs fn now and then every `interval` ms while the page is mounted. Ticks are
// skipped while the tab is hidden or the previous call is still running, and
// the tab refreshes as soon as it becomes visible again.
export const usePolling = (fn, interval = 5000) => {
  let timer = null
  let running = false

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

  onMounted(() => {
    tick()
    timer = setInterval(tick, interval)
    document.addEventListener('visibilitychange', onVisible)
  })
  onUnmounted(() => {
    clearInterval(timer)
    document.removeEventListener('visibilitychange', onVisible)
  })

  // refresh ignores the in-flight guard so an explicit action always reloads.
  return { refresh: () => fn() }
}
