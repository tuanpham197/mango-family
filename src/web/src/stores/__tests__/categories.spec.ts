import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import type { CategoryTree } from '../../api/types'

// Mock lớp client để store không gọi fetch thật.
const apiMock = vi.fn()
vi.mock('../../api/client', () => ({
  api: (...args: unknown[]) => apiMock(...args),
}))

import { useCategoriesStore } from '../categories'

function tree(overrides: Partial<CategoryTree> & Pick<CategoryTree, 'id' | 'name' | 'type'>): CategoryTree {
  return {
    household_id: 'h1',
    icon: null,
    parent_id: null,
    is_default: false,
    is_hidden: false,
    created_by: null,
    created_at: '',
    updated_at: '',
    children: [],
    ...overrides,
  }
}

const sample: CategoryTree[] = [
  tree({
    id: 'eat',
    name: 'Ăn uống',
    type: 'EXPENSE',
    children: [
      { ...tree({ id: 'eat-out', name: 'Ăn ngoài', type: 'EXPENSE' }), parent_id: 'eat' },
      { ...tree({ id: 'eat-hidden', name: 'Ăn ẩn', type: 'EXPENSE', is_hidden: true }), parent_id: 'eat' },
    ] as never,
  }),
  tree({ id: 'ent', name: 'Giải trí', type: 'EXPENSE', is_hidden: true }),
  tree({ id: 'salary', name: 'Lương', type: 'INCOME' }),
]

describe('categories store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    apiMock.mockReset()
  })

  it('byType lọc đúng loại, giữ cả danh mục ẩn (cho màn quản lý)', async () => {
    apiMock.mockResolvedValueOnce(sample)
    const store = useCategoriesStore()
    await store.fetch()

    const expense = store.byType('EXPENSE')
    expect(expense.map((c) => c.id)).toEqual(['eat', 'ent']) // gồm cả 'ent' đang ẩn
    expect(store.byType('INCOME').map((c) => c.id)).toEqual(['salary'])
  })

  it('pickable loại trừ danh mục ẩn (gốc lẫn con) — FR-020', async () => {
    apiMock.mockResolvedValueOnce(sample)
    const store = useCategoriesStore()
    await store.fetch()

    const pickable = store.pickable('EXPENSE')
    expect(pickable.map((c) => c.id)).toEqual(['eat']) // 'ent' ẩn bị loại
    expect(pickable[0].children.map((c) => c.id)).toEqual(['eat-out']) // con ẩn bị loại
  })

  it('findById tìm được cả danh mục gốc lẫn con', async () => {
    apiMock.mockResolvedValueOnce(sample)
    const store = useCategoriesStore()
    await store.fetch()

    expect(store.findById('eat')?.name).toBe('Ăn uống')
    expect(store.findById('eat-out')?.name).toBe('Ăn ngoài')
    expect(store.findById('missing')).toBeUndefined()
  })

  it('ensure chỉ fetch một lần', async () => {
    apiMock.mockResolvedValue(sample)
    const store = useCategoriesStore()
    await store.ensure()
    await store.ensure()
    expect(apiMock).toHaveBeenCalledTimes(1)
  })
})
