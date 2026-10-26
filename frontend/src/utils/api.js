// Best-effort message for a failed axios call.
export const errorMessage = (e, fallback = '操作失败') => {
  const msg = e?.response?.data?.error
  if (typeof msg === 'string' && msg) return msg
  if (!e?.response) return '无法连接到服务器，请检查网络或服务是否在运行'
  return fallback
}

// Requests made with responseType 'blob' deliver error bodies as a Blob too.
export const errorMessageAsync = async (e, fallback = '操作失败') => {
  const data = e?.response?.data
  if (data instanceof Blob) {
    try {
      const parsed = JSON.parse(await data.text())
      if (parsed?.error) return String(parsed.error)
    } catch (err) {
      // not JSON; use the fallback
    }
    return fallback
  }
  return errorMessage(e, fallback)
}
