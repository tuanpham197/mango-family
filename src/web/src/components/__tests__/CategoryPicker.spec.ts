import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import type { CategoryTree } from '../../api/types'

const apiMock = vi.fn()
vi.mock('../../api/client', () => ({
  api: (...args: unknown[]) => apiMock(...args),
  ApiError: class ApiError extends Error {},
}))

import CategoryPicker from '../CategoryPicker.vue'
import { useCategoriesStore } from '../../stores/categories'

function tree(o: Partial<CategoryTree> & Pick<CategoryTree, 'id' | 'name' | 'type'>): CategoryTree {
  return {
    household_id: 'h1', icon: null, parent_id: null, is_default: false, is_hidden: false,
    created_by: null, created_at: '', updated_at: '', children: [], ...o,
  }
}

describe('CategoryPicker', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    apiMock.mockReset()
  })

  it('chỉ hiển thị danh mục đúng loại, loại trừ ẩn, nhóm cha → con', async () => {
    const store = useCategoriesStore()
    store.trees = [
      tree({
        id: 'eat', name: 'Ăn uống', type: 'EXPENSE',
        children: [{ ...tree({ id: 'out', name: 'Ăn ngoài', type: 'EXPENSE' }), parent_id: 'eat' }] as never,
      }),
      tree({ id: 'hid', name: 'Ẩn', type: 'EXPENSE', is_hidden: true }),
      tree({ id: 'salary', name: 'Lương', type: 'INCOME' }),
    ]
    store.loaded = true

    const wrapper = mount(CategoryPicker, { props: { type: 'EXPENSE', modelValue: null } })
    const text = wrapper.text()
    expect(text).toContain('Ăn uống')
    expect(text).toContain('Ăn ngoài')
    expect(text).not.toContain('Ẩn')
    expect(text).not.toContain('Lương') // khác loại
  })

  it('click danh mục phát update:modelValue với id đúng', async () => {
    const store = useCategoriesStore()
    store.trees = [tree({ id: 'eat', name: 'Ăn uống', type: 'EXPENSE' })]
    store.loaded = true

    const wrapper = mount(CategoryPicker, { props: { type: 'EXPENSE', modelValue: null } })
    await wrapper.get('[data-testid="category-option-Ăn uống"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['eat'])
  })
})
