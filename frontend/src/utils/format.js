// Byte counters from WireGuard are from the server's side: rx was uploaded by
// the client, tx was downloaded by it. The UI speaks from the client's side,
// so "down" is tx and "up" is rx.
export const downOf = (t) => Number(t?.tx) || 0
export const upOf = (t) => Number(t?.rx) || 0

export const formatBytes = (n) => {
  const v = Math.max(0, Number(n) || 0)
  if (v < 1024) return `${Math.round(v)} B`
  const units = ['KB', 'MB', 'GB', 'TB', 'PB']
  let x = v / 1024
  let i = 0
  while (x >= 1024 && i < units.length - 1) {
    x /= 1024
    i++
  }
  return `${x.toFixed(i >= 2 ? 2 : 1)} ${units[i]}`
}

// Like formatBytes but without trailing zeros ("10 GB", "1.5 KB"); for axis labels.
export const formatBytesCompact = (n) => {
  const v = Math.max(0, Number(n) || 0)
  if (v < 1024) return `${Math.round(v)} B`
  const units = ['KB', 'MB', 'GB', 'TB', 'PB']
  let x = v / 1024
  let i = 0
  while (x >= 1024 && i < units.length - 1) {
    x /= 1024
    i++
  }
  return `${Number(x.toFixed(2))} ${units[i]}`
}

export const formatRate = (bytesPerSec) => `${formatBytes(bytesPerSec)}/s`

// Seconds since an event as a short Chinese phrase; negative means "never".
export const formatAge = (seconds) => {
  const s = Number(seconds)
  if (!Number.isFinite(s) || s < 0) return '从未连接'
  if (s < 5) return '刚刚'
  if (s < 60) return `${Math.floor(s)} 秒前`
  if (s < 3600) return `${Math.floor(s / 60)} 分钟前`
  if (s < 86400) return `${Math.floor(s / 3600)} 小时前`
  return `${Math.floor(s / 86400)} 天前`
}

export const formatDateTime = (unixSeconds) => {
  if (!unixSeconds) return '—'
  return new Date(unixSeconds * 1000).toLocaleString('zh-CN', { hour12: false })
}

const pad = (n) => String(n).padStart(2, '0')
const dayKey = (d) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`

// Expands sparse per-day rows ({day, rx, tx}) into `count` consecutive calendar
// days ending at endDay ('YYYY-MM-DD', the server's today), oldest first, with
// zeros for days that have no data. Days are bucketed in server time, so the
// browser's own clock is only the fallback when endDay is unknown.
export const fillDays = (rows, count, endDay) => {
  const byDay = new Map((rows || []).map((r) => [r.day, r]))
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(endDay || '')
  const end = m ? new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3])) : new Date()
  const out = []
  for (let i = count - 1; i >= 0; i--) {
    const d = new Date(end.getFullYear(), end.getMonth(), end.getDate() - i)
    const key = dayKey(d)
    const row = byDay.get(key)
    out.push({ day: key, rx: Number(row?.rx) || 0, tx: Number(row?.tx) || 0 })
  }
  return out
}

export const ipToNum = (cidr) => {
  const ip = (cidr || '').split('/')[0].split(',')[0].trim()
  const parts = ip.split('.').map((n) => parseInt(n, 10))
  if (parts.length !== 4 || parts.some((n) => Number.isNaN(n))) return 0
  return ((parts[0] << 24) >>> 0) + (parts[1] << 16) + (parts[2] << 8) + parts[3]
}

// navigator.clipboard only exists in secure contexts, and this UI is often
// served over plain http on an IP, so fall back to a hidden textarea.
export const copyText = async (text) => {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
      return true
    }
  } catch (e) {
    // fall through to the legacy path
  }
  const ta = document.createElement('textarea')
  ta.value = text
  ta.setAttribute('readonly', '')
  ta.style.cssText = 'position:fixed;top:-1000px;opacity:0'
  document.body.appendChild(ta)
  ta.select()
  let ok = false
  try {
    ok = document.execCommand('copy')
  } catch (e) {
    ok = false
  }
  document.body.removeChild(ta)
  return ok
}

export const saveBlob = (blob, filename) => {
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}
