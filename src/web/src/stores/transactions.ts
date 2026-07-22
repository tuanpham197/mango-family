import { defineStore } from 'pinia'
import { api, apiPaged, type Paging } from '../api/client'
import type { CategoryType, Transaction } from '../api/types'

export interface TransactionInput {
  amount: number
  type: CategoryType
  category_id: string
  account_id: string
  description?: string
  transaction_date?: string
}

const PAGE_SIZE = 50

// URL sổ chung + phân trang + filter khoảng (from/to) tùy chọn — filter tháng/năm.
function listUrl(page: number, from: string, to: string): string {
  const range = from && to ? `&from=${from}&to=${to}` : ''
  return `/api/transactions?page=${page}&page_size=${PAGE_SIZE}${range}`
}

export const useTransactionsStore = defineStore('transactions', {
  state: () => ({
    items: [] as Transaction[],
    paging: null as Paging | null,
    loading: false,
    submitting: false, // chống double-submit (quickstart #23)
    from: '' as string, // filter khoảng (YYYY-MM-DD); rỗng = không lọc
    to: '' as string,
  }),
  getters: {
    hasMore: (s) => (s.paging ? s.items.length < s.paging.total : false),
  },
  actions: {
    // Đặt filter khoảng tháng/năm rồi tải lại trang 1 (from/to rỗng = xem tất cả).
    async setRange(from: string, to: string) {
      this.from = from
      this.to = to
      await this.fetch()
    },
    // Tải trang 1 (làm mới sổ) — dùng cho realtime refetch; giữ nguyên filter hiện tại.
    async fetch() {
      this.loading = true
      try {
        const res = await apiPaged<Transaction[]>(listUrl(1, this.from, this.to))
        this.items = res.data
        this.paging = res.paging ?? null
      } finally {
        this.loading = false
      }
    },
    // Nạp thêm trang kế (infinite scroll — D16); giữ nguyên filter khoảng.
    async loadMore() {
      if (!this.paging || this.loading || !this.hasMore) return
      this.loading = true
      try {
        const next = this.paging.page + 1
        const res = await apiPaged<Transaction[]>(listUrl(next, this.from, this.to))
        this.items.push(...res.data)
        this.paging = res.paging ?? this.paging
      } finally {
        this.loading = false
      }
    },
    async getById(id: string): Promise<Transaction> {
      return api<Transaction>(`/api/transactions/${id}`)
    },
    /**
     * Tạo giao dịch. Chống double-submit: nếu đang gửi thì bỏ qua lần gọi sau
     * (thử lại sau lỗi mạng không tạo trùng — quickstart #23). Ném ApiError cho view.
     */
    async create(input: TransactionInput): Promise<Transaction> {
      if (this.submitting) throw new Error('đang xử lý')
      this.submitting = true
      try {
        const t = await api<Transaction>('/api/transactions', { method: 'POST', body: JSON.stringify(input) })
        await this.fetch()
        return t
      } finally {
        this.submitting = false
      }
    },
    /** Sửa. Ném ApiError CONCURRENCY_CONFLICT (kèm data mới) / RECORD_GONE (D17). */
    async update(id: string, input: TransactionInput & { expected_updated_at: string }): Promise<Transaction> {
      if (this.submitting) throw new Error('đang xử lý')
      this.submitting = true
      try {
        const t = await api<Transaction>(`/api/transactions/${id}`, { method: 'PATCH', body: JSON.stringify(input) })
        await this.fetch()
        return t
      } finally {
        this.submitting = false
      }
    },
    /** Xóa (UI đã xác nhận trước). Ném ApiError RECORD_GONE nếu đã bị xóa. */
    async remove(id: string, expectedUpdatedAt?: string) {
      await api(`/api/transactions/${id}`, {
        method: 'DELETE',
        body: JSON.stringify({ expected_updated_at: expectedUpdatedAt }),
      })
      await this.fetch()
    },
  },
})
