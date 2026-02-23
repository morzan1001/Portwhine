import { create } from 'zustand'
import type { Timestamp } from '@bufbuild/protobuf/wkt'

interface AuthState {
  accessToken: string | null
  refreshToken: string | null
  expiresAt: Date | null
  userId: string | null
  username: string | null
  email: string | null
  role: string | null

  setTokens: (tokens: {
    accessToken: string
    refreshToken: string
    expiresAt?: Timestamp
    userId?: string
    username?: string
    email?: string
    role?: string
  }) => void
  clearAuth: () => void
  isAuthenticated: () => boolean
  hydrate: () => void
}

const STORAGE_KEY = 'portwhine-auth'

const getFromStorage = (key: string): unknown => {
  try {
    const item = localStorage.getItem(key)
    return item ? JSON.parse(item) : null
  } catch {
    return null
  }
}

const setToStorage = (key: string, value: unknown) => {
  try {
    localStorage.setItem(key, JSON.stringify(value))
  } catch {
    // localStorage may be full or unavailable
  }
}

const removeFromStorage = (key: string) => {
  try {
    localStorage.removeItem(key)
  } catch {
    // Ignore
  }
}

export const useAuthStore = create<AuthState>((set, get) => ({
  accessToken: null,
  refreshToken: null,
  expiresAt: null,
  userId: null,
  username: null,
  email: null,
  role: null,

  setTokens: (tokens) => {
    // Convert protobuf Timestamp to Date
    let expiresDate: Date | null = null
    if (tokens.expiresAt) {
      const seconds = Number(tokens.expiresAt.seconds)
      const nanos = tokens.expiresAt.nanos || 0
      expiresDate = new Date(seconds * 1000 + nanos / 1000000)
    }

    const newState = {
      accessToken: tokens.accessToken,
      refreshToken: tokens.refreshToken,
      expiresAt: expiresDate,
      userId: tokens.userId || null,
      username: tokens.username || null,
      email: tokens.email || null,
      role: tokens.role || null,
    }

    // Save to localStorage
    setToStorage(STORAGE_KEY, newState)

    set(newState)
  },

  clearAuth: () => {
    removeFromStorage(STORAGE_KEY)
    set({
      accessToken: null,
      refreshToken: null,
      expiresAt: null,
      userId: null,
      username: null,
      email: null,
      role: null,
    })
  },

  isAuthenticated: () => {
    const { accessToken, expiresAt } = get()
    if (!accessToken) return false
    if (!expiresAt) return true
    return new Date() < expiresAt
  },

  hydrate: () => {
    // Load from localStorage on client mount
    const stored = getFromStorage(STORAGE_KEY) as Record<string, unknown> | null
    if (stored) {
      set({
        ...stored,
        expiresAt: stored.expiresAt ? new Date(stored.expiresAt as string) : null,
      })
    }
  },
}))
