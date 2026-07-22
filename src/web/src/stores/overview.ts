import { defineStore } from 'pinia'
import { api } from '../api/client'
import type { OverviewSummary } from '../api/types'

// Tổng hợp giá trị suy ra của màn Tổng quan (GET /api/overview — D27). Tất cả số liệu
// tính ở server khi đọc; FE refetch khi có bất kỳ topic ảnh hưởng (D33) → realtime ≤ 5s.
export const useOverviewStore = defineStore('overview', {
  state: () => ({
    summary: null as OverviewSummary | null,
    loading: false,
  }),
  actions: {
    async fetch() {
      this.loading = true
      try {
        this.summary = await api<OverviewSummary>('/api/overview')
      } finally {
        this.loading = false
      }
    },
  },
})
