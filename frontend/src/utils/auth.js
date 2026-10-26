const KEY = 'token'

export const getToken = () => {
  try {
    return localStorage.getItem(KEY)
  } catch (e) {
    return null
  }
}

export const setToken = (token) => {
  try {
    localStorage.setItem(KEY, token)
  } catch (e) {
    // storage unavailable: the session just won't survive a reload
  }
}

export const clearToken = () => {
  try {
    localStorage.removeItem(KEY)
  } catch (e) {
    // ignore
  }
}

// Reads the exp claim to avoid showing pages that are doomed to 401. The server
// still verifies the signature; this is only for navigation.
export const isTokenValid = (token = getToken()) => {
  if (!token) return false
  try {
    const payload = JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')))
    return !payload.exp || payload.exp * 1000 > Date.now()
  } catch (e) {
    return false
  }
}
