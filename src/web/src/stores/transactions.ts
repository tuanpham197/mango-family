import { defineStore } from 'pinia'
import { api, apiPaged, type Paging } from '../api/client'
import type { CategoryType, Transaction } from '../api/types'

export interface CreateTransactionInput {
  amount: number
  type: CategoryType
  category_id: string
  description?: string
  transaction_date?: string
}

export const useTransactionsStore = defineStore('transactions', {
  state: () => ({
    items: [] as Transaction[],
    paging: null as Paging | null,
    loading: false,
  }),
  actions: {
    async fetch(page = 1) {
      this.loading = true
      try {
        const res = await apiPaged<Transaction[]>(`/api/transactions?page=${page}&page_size=50`)
        this.items = res.data
        this.paging = res.paging ?? null
      } finally {
        this.loading = false
      }
    },
    /** Ném ApiError (CATEGORY_REQUIRED / CATEGORY_TYPE_MISMATCH / …) cho view. */
    async create(input: CreateTransactionInput): Promise<Transaction> {
      const t = await api<Transaction>('/api/transactions', {
        method: 'POST',
        body: JSON.stringify(input),
      })
      await this.fetch()
      return t
    },
  },
})
