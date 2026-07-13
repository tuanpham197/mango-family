import { defineStore } from 'pinia'
import { api } from '../api/client'
import type { Category, CategoryTree, CategoryType } from '../api/types'

// Store danh mục: luôn fetch include_hidden=true (một nguồn duy nhất);
// việc lọc theo loại / loại trừ hidden nằm ở getters.
export const useCategoriesStore = defineStore('categories', {
  state: () => ({
    trees: [] as CategoryTree[],
    loaded: false,
  }),
  getters: {
    /** Cây danh mục theo loại (gồm cả hidden — cho màn quản lý). */
    byType: (s) => (type: CategoryType) => s.trees.filter((t) => t.type === type),
    /** Cây để chọn cho giao dịch mới: đúng loại, loại trừ hidden (FR-014/020). */
    pickable: (s) => (type: CategoryType) =>
      s.trees
        .filter((t) => t.type === type && !t.is_hidden)
        .map((t) => ({ ...t, children: t.children.filter((c) => !c.is_hidden) })),
    findById: (s) => (id: string): Category | undefined => {
      for (const t of s.trees) {
        if (t.id === id) return t
        const child = t.children.find((c) => c.id === id)
        if (child) return child
      }
      return undefined
    },
  },
  actions: {
    async fetch() {
      this.trees = await api<CategoryTree[]>('/api/categories?include_hidden=true')
      this.loaded = true
    },
    async ensure() {
      if (!this.loaded) await this.fetch()
    },
    /** Ném ApiError NAME_DUPLICATE_WARNING / NESTING_TOO_DEEP / PARENT_TYPE_MISMATCH. */
    async create(input: {
      name: string
      type?: CategoryType
      icon?: string
      parent_id?: string
      confirm_duplicate?: boolean
    }): Promise<Category> {
      const cat = await api<Category>('/api/categories', {
        method: 'POST',
        body: JSON.stringify(input),
      })
      await this.fetch()
      return cat
    },
    /** Ném ApiError CONCURRENCY_CONFLICT (409) / RECORD_GONE (404) — D6. */
    async update(
      id: string,
      patch: { name?: string; icon?: string; is_hidden?: boolean; expected_updated_at: string },
    ): Promise<Category> {
      const cat = await api<Category>(`/api/categories/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(patch),
      })
      await this.fetch()
      return cat
    },
    /** Ném ApiError CATEGORY_HAS_TRANSACTIONS (kèm counts) / REASSIGN_TYPE_MISMATCH / … */
    async remove(
      id: string,
      body: {
        mode?: 'reassign' | 'delete_transactions'
        target_category_id?: string
        expected_updated_at: string
      },
    ) {
      await api(`/api/categories/${id}`, {
        method: 'DELETE',
        body: JSON.stringify(body),
      })
      await this.fetch()
    },
  },
})
