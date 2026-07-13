import { defineStore } from 'pinia'
import { api } from '../api/client'
import type { Household, User } from '../api/types'

interface MePayload {
  user: User
  household: Household
}

// Phiên đăng nhập: JWT nằm trong cookie HttpOnly — store chỉ giữ hồ sơ,
// bootstrap qua GET /api/me (research D4/D10).
export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    household: null as Household | null,
    ready: false,
  }),
  getters: {
    isAuthenticated: (s) => s.user !== null,
  },
  actions: {
    /** Gọi một lần khi app khởi động (router guard) — xác định phiên hiện có. */
    async bootstrap() {
      if (this.ready) return
      try {
        await this.refresh()
      } catch {
        this.user = null
        this.household = null
      } finally {
        this.ready = true
      }
    },
    async refresh() {
      const me = await api<MePayload>('/api/me')
      this.user = me.user
      this.household = me.household
    },
    /** Ném ApiError (INVALID_CREDENTIALS / NO_HOUSEHOLD) cho view xử lý. */
    async login(email: string, password: string) {
      await api<{ user: User }>('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      })
      await this.refresh()
    },
    async logout() {
      try {
        await api('/api/auth/logout', { method: 'POST' })
      } finally {
        this.user = null
        this.household = null
      }
    },
  },
})
