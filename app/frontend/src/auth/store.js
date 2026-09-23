import { reactive, computed } from 'vue'
import { registerRequest, loginRequest, logoutRequest, refreshRequest, getMe } from '../api/client.js'

const STORAGE_KEY = 'ojozone.auth'

function loadStored() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return { accessToken: '', refreshToken: '' }
    const parsed = JSON.parse(raw)
    return {
      accessToken: typeof parsed.accessToken === 'string' ? parsed.accessToken : '',
      refreshToken: typeof parsed.refreshToken === 'string' ? parsed.refreshToken : '',
    }
  } catch {
    return { accessToken: '', refreshToken: '' }
  }
}

function persist(state) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify({ accessToken: state.accessToken, refreshToken: state.refreshToken }))
}

function clearStored() {
  localStorage.removeItem(STORAGE_KEY)
}

const stored = loadStored()
const state = reactive({
  user: null,
  accessToken: stored.accessToken,
  refreshToken: stored.refreshToken,
  ready: false,
})

export function applyTokens(accessToken, refreshToken) {
  state.accessToken = accessToken
  state.refreshToken = refreshToken
  if (accessToken && refreshToken) {
    persist(state)
  } else {
    clearStored()
  }
}

export function setUser(user) {
  state.user = user
}

export async function authenticate(action) {
  const result = await action()
  applyTokens(result.data.tokens.access_token, result.data.tokens.refresh_token)
  return result.data.user
}

export async function login(email, password) {
  return authenticate(() => loginRequest(email, password))
}

export async function register(email, password, locale) {
  return authenticate(() => registerRequest(email, password, locale))
}

export async function refreshTokens() {
  if (!state.refreshToken) return null
  const result = await refreshRequest(state.refreshToken)
  applyTokens(result.data.tokens.access_token, result.data.tokens.refresh_token)
  return result.data.tokens.access_token
}

export async function logout() {
  if (state.refreshToken) {
    try {
      await logoutRequest(state.refreshToken)
    } catch {
      // ignorer un échec réseau pendant la déconnexion locale
    }
  }
  state.user = null
  applyTokens('', '')
}

export async function loadCurrentUser() {
  if (!state.accessToken) return null
  const result = await getMe(state.accessToken)
  setUser(result.data)
  return result.data
}

export async function initAuth() {
  if (state.ready) return
  if (state.accessToken) {
    try {
      await loadCurrentUser()
    } catch (err) {
      if (err.status === 401) {
        try {
          await refreshTokens()
          await loadCurrentUser()
        } catch {
          state.user = null
          applyTokens('', '')
        }
      }
    }
  }
  state.ready = true
}

export const isAuthenticated = computed(() => state.user != null)
export const currentUser = computed(() => state.user)

export function hasRole(...roles) {
  return state.user != null && roles.includes(state.user.role)
}

export { state }