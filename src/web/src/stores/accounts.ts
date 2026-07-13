import { defineStore } from 'pinia'
import { api } from '../api/client'
import type { Account } from '../api/types'

// Tài khoản + số dư suy ra (GET /api/accounts). 002 chỉ đọc; refetch khi có
// event accounts_changed (số dư đổi sau mỗi mutation giao dịch — D14).
export const useAccountsStore = defineStore('accounts', {
  state: () => ({
    items: [] as Account[],
    loaded: false,
  }),
  getters: {
    total: (s) => s.items.reduce((sum, a) => sum + a.balance, 0),
    /** Chọn sẵn khi hộ chỉ có 1 tài khoản (FR-006). */
    defaultId: (s) => (s.items.length >= 1 ? s.items[0].id : null),
  },
  actions: {
    async fetch() {
      this.items = await api<Account[]>('/api/accounts')
      this.loaded = true
    },
    async ensure() {
      if (!this.loaded) await this.fetch()
    },
  },
})
