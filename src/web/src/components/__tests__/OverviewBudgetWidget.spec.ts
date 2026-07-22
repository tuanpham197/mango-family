import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import type { Budget } from '../../api/types'

const { pushMock } = vi.hoisted(() => ({ pushMock: vi.fn() }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: pushMock }) }))
vi.mock('../../composables/useInvalidation', () => ({ useInvalidation: () => {} }))

const apiMock = vi.fn()
vi.mock('../../api/client', () => ({ api: (...a: unknown[]) => apiMock(...a), ApiError: class extends Error {} }))

import OverviewBudgetWidget from '../OverviewBudgetWidget.vue'

function budget(o: Partial<Budget>): Budget {
  return {
    id: 'b', household_id: 'h', type: 'CATEGORY', category_id: 'c', category_name: 'Cat',
    category_hidden: false, limit_amount: 1000, period_type: 'MONTHLY', start_date: null, end_date: null,
    status: 'ACTIVE', period_key: '2026-07', spent: 0, percent: 0, alerts: [], created_by: 'u',
    created_by_name: 'Alice', updated_at: '', ...o,
  }
}

describe('OverviewBudgetWidget', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    apiMock.mockReset()
    pushMock.mockReset()
  })

  it('hiển thị ngân sách tổng + tối đa 3 danh mục nổi bật (tiến độ cao nhất)', async () => {
    apiMock.mockResolvedValue([
      budget({ id: 't', type: 'TOTAL', category_id: null, category_name: null, percent: 30 }),
      budget({ id: 'a', category_name: 'A', percent: 90 }),
      budget({ id: 'b', category_name: 'B', percent: 50 }),
      budget({ id: 'c', category_name: 'C', percent: 85 }),
      budget({ id: 'd', category_name: 'D', percent: 10 }),
    ])
    const w = mount(OverviewBudgetWidget)
    await flushPromises()
    expect(w.find('[data-testid="overview-total"]').exists()).toBe(true)
    const cats = w.findAll('[data-testid="overview-category"]')
    expect(cats).toHaveLength(3) // nổi bật = top 3 theo percent
    const text = w.text()
    expect(text).toContain('A')
    expect(text).toContain('C')
    expect(text).not.toContain('D') // percent thấp nhất bị loại
  })

  it('"Xem tất cả" điều hướng tới /budgets (FR-014)', async () => {
    apiMock.mockResolvedValue([budget({ id: 'a', percent: 40 })])
    const w = mount(OverviewBudgetWidget)
    await flushPromises()
    await w.get('[data-testid="budget-see-all"]').trigger('click')
    expect(pushMock).toHaveBeenCalledWith('/budgets')
  })

  it('không có ngân sách → thông báo trống', async () => {
    apiMock.mockResolvedValue([])
    const w = mount(OverviewBudgetWidget)
    await flushPromises()
    expect(w.find('[data-testid="overview-no-budget"]').exists()).toBe(true)
  })
})
