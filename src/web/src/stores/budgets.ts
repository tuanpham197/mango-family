import { defineStore } from 'pinia'
import { api } from '../api/client'
import type { Budget, BudgetType, PeriodType } from '../api/types'

// Ngân sách của hộ + tiến độ/cảnh báo suy ra (GET /api/budgets). Tiến độ (spent/
// percent/alerts) tính ở server khi đọc (D21); FE refetch khi có budgets_changed
// HOẶC transactions_changed (giao dịch đổi → tiến độ đổi — D26, ≤ 5s SC-006).
export interface BudgetInput {
  type: BudgetType
  category_id?: string | null
  limit_amount: number
  period_type: PeriodType
  start_date?: string | null
  end_date?: string | null
}

export const useBudgetsStore = defineStore('budgets', {
  state: () => ({
    items: [] as Budget[],
    loaded: false,
    loading: false,
    submitting: false, // chống double-submit
  }),
  getters: {
    /** Ngân sách tổng đang hoạt động (nếu có) — cho widget Tổng quan. */
    totalBudget: (s) => s.items.find((b) => b.type === 'TOTAL' && b.status === 'ACTIVE') ?? null,
    categoryBudgets: (s) => s.items.filter((b) => b.type === 'CATEGORY'),
  },
  actions: {
    async fetch(status = 'ACTIVE') {
      this.loading = true
      try {
        this.items = await api<Budget[]>(`/api/budgets?status=${status}`)
        this.loaded = true
      } finally {
        this.loading = false
      }
    },
    async ensure() {
      if (!this.loaded) await this.fetch()
    },
    async getById(id: string): Promise<Budget> {
      return api<Budget>(`/api/budgets/${id}`)
    },
    /** Tạo ngân sách. Ném ApiError (LIMIT_INVALID / BUDGET_DUPLICATE + existing_budget_id / …). */
    async create(input: BudgetInput): Promise<Budget> {
      if (this.submitting) throw new Error('đang xử lý')
      this.submitting = true
      try {
        const b = await api<Budget>('/api/budgets', { method: 'POST', body: JSON.stringify(input) })
        await this.fetch()
        return b
      } finally {
        this.submitting = false
      }
    },
    /** Sửa. Ném ApiError CONCURRENCY_CONFLICT (kèm data mới) / RECORD_GONE (D25). */
    async update(id: string, input: BudgetInput & { expected_updated_at: string }): Promise<Budget> {
      if (this.submitting) throw new Error('đang xử lý')
      this.submitting = true
      try {
        const b = await api<Budget>(`/api/budgets/${id}`, { method: 'PATCH', body: JSON.stringify(input) })
        await this.fetch()
        return b
      } finally {
        this.submitting = false
      }
    },
    /** Xóa (UI đã xác nhận trước — UC-BGT-06). Ném ApiError RECORD_GONE nếu đã bị xóa. */
    async remove(id: string, expectedUpdatedAt?: string) {
      await api(`/api/budgets/${id}`, {
        method: 'DELETE',
        body: JSON.stringify({ expected_updated_at: expectedUpdatedAt }),
      })
      await this.fetch()
    },
  },
})
