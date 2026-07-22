import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import type { CategoryTree } from '../../api/types'

const { pushMock } = vi.hoisted(() => ({ pushMock: vi.fn() }))
vi.mock('vue-router', () => ({
  useRoute: () => ({ params: {} }), // chế độ TẠO (không có :id)
  useRouter: () => ({ push: pushMock }),
}))
vi.mock('../../composables/useInvalidation', () => ({ useInvalidation: () => {} }))

const apiMock = vi.fn()
vi.mock('../../api/client', () => ({ api: (...a: unknown[]) => apiMock(...a), ApiError: class ApiError extends Error {} }))

import BudgetFormView from '../BudgetFormView.vue'
import { useCategoriesStore } from '../../stores/categories'

function tree(o: Partial<CategoryTree> & Pick<CategoryTree, 'id' | 'name' | 'type'>): CategoryTree {
  return {
    household_id: 'h', icon: null, parent_id: null, is_default: false, is_hidden: false,
    created_by: null, created_at: '', updated_at: '', children: [], ...o,
  }
}

async function mountForm() {
  const cats = useCategoriesStore()
  cats.trees = [tree({ id: 'eat', name: 'Ăn uống', type: 'EXPENSE' })]
  cats.loaded = true
  const w = mount(BudgetFormView)
  await flushPromises()
  return w
}

describe('BudgetFormView validation', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    apiMock.mockReset()
    pushMock.mockReset()
  })

  it('giới hạn ≤ 0 → chặn, không gọi API', async () => {
    const w = await mountForm()
    await w.get('[data-testid="limit-input"]').setValue(0)
    await w.get('form').trigger('submit')
    expect(w.get('[data-testid="error-limit"]').text()).toBeTruthy()
    expect(apiMock).not.toHaveBeenCalled()
  })

  it('CATEGORY thiếu danh mục → chặn', async () => {
    const w = await mountForm()
    await w.get('[data-testid="limit-input"]').setValue(5000000)
    await w.get('form').trigger('submit')
    expect(w.get('[data-testid="error-category"]').text()).toBeTruthy()
    expect(apiMock).not.toHaveBeenCalled()
  })

  it('kỳ một lần thiếu/sai ngày → chặn', async () => {
    const w = await mountForm()
    // chuyển sang TOTAL để bỏ ràng buộc danh mục, tập trung kiểm tra kỳ
    await w.get('[data-testid="type-total"]').trigger('click')
    await w.get('[data-testid="limit-input"]').setValue(1000000)
    await w.get('[data-testid="period-select"]').setValue('ONE_TIME')
    await w.get('[data-testid="start-date-input"]').setValue('2026-07-10')
    await w.get('[data-testid="end-date-input"]').setValue('2026-07-05') // kết thúc trước bắt đầu
    await w.get('form').trigger('submit')
    expect(w.get('[data-testid="error-period"]').text()).toBeTruthy()
    expect(apiMock).not.toHaveBeenCalled()
  })

  it('TOTAL ẩn ô chọn danh mục (T024)', async () => {
    const w = await mountForm()
    expect(w.find('[data-testid="category-picker"]').exists()).toBe(true)
    await w.get('[data-testid="type-total"]').trigger('click')
    expect(w.find('[data-testid="category-picker"]').exists()).toBe(false)
  })

  it('hợp lệ → gọi store.create (POST)', async () => {
    apiMock.mockResolvedValue({ id: 'b1' }) // POST + refetch
    const w = await mountForm()
    await w.get('[data-testid="limit-input"]').setValue(5000000)
    await w.get('[data-testid="category-option-Ăn uống"]').trigger('click')
    await w.get('form').trigger('submit')
    await flushPromises()
    expect(apiMock).toHaveBeenCalledWith('/api/budgets', expect.objectContaining({ method: 'POST' }))
  })
})
