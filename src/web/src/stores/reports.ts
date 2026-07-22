import { defineStore } from 'pinia'
import { api } from '../api/client'
import type { CategoryReport, ReportOverview } from '../api/types'

// Báo cáo (feature 005) — chỉ ĐỌC, tổng hợp ở server theo [from,to] (D34). Giữ khoảng
// đang chọn + báo cáo tổng quan + báo cáo chi tiết một danh mục (drill-in).
export const useReportsStore = defineStore('reports', {
  state: () => ({
    from: '' as string, // YYYY-MM-DD
    to: '' as string,
    overview: null as ReportOverview | null,
    category: null as CategoryReport | null, // đang drill-in một danh mục (null = không)
    loading: false,
  }),
  actions: {
    setRange(from: string, to: string) {
      this.from = from
      this.to = to
    },
    async loadOverview() {
      if (!this.from || !this.to) return
      this.loading = true
      try {
        this.overview = await api<ReportOverview>(`/api/reports/overview?from=${this.from}&to=${this.to}`)
      } finally {
        this.loading = false
      }
    },
    async openCategory(id: string) {
      this.category = await api<CategoryReport>(`/api/reports/category/${id}?from=${this.from}&to=${this.to}`)
    },
    closeCategory() {
      this.category = null
    },
    /** Làm mới cả tổng quan + chi tiết đang mở (dùng khi đổi khoảng / dữ liệu nền đổi — D40). */
    async refresh() {
      await this.loadOverview()
      if (this.category) await this.openCategory(this.category.category_id)
    },
  },
})
