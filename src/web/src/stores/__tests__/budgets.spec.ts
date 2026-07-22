import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const apiMock = vi.fn()
vi.mock('../../api/client', () => ({
  api: (...a: unknown[]) => apiMock(...a),
  ApiError: class ApiError extends Error {},
}))

import { useBudgetsStore } from '../budgets'
import type { Budget } from '../../api/types'

function budget(o: Partial<Budget>): Budget {
  return {
    id: 'b1', household_id: 'h', type: 'CATEGORY', category_id: 'c', category_name: 'Ăn uống',
    category_hidden: false, limit_amount: 1000, period_type: 'MONTHLY', start_date: null, end_date: null,
    status: 'ACTIVE', period_key: '2026-07', spent: 0, percent: 0, alerts: [], created_by: 'u',
    created_by_name: 'Alice', updated_at: '2026-07-01T00:00:00Z', ...o,
  }
}

const input = { type: 'CATEGORY' as const, category_id: 'c', limit_amount: 5000000, period_type: 'MONTHLY' as const }

describe('budgets store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    apiMock.mockReset()
    apiMock.mockResolvedValue([]) // fetch mặc định
  })

  it('fetch nạp danh sách theo status (mặc định ACTIVE)', async () => {
    apiMock.mockResolvedValueOnce([budget({})])
    const s = useBudgetsStore()
    await s.fetch()
    expect(apiMock).toHaveBeenCalledWith('/api/budgets?status=ACTIVE')
    expect(s.items).toHaveLength(1)
  })

  it('create gọi POST rồi refetch', async () => {
    apiMock.mockResolvedValueOnce(budget({})) // POST
    const s = useBudgetsStore()
    await s.create(input)
    expect(apiMock).toHaveBeenNthCalledWith(1, '/api/budgets', expect.objectContaining({ method: 'POST' }))
    expect(apiMock).toHaveBeenNthCalledWith(2, '/api/budgets?status=ACTIVE') // refetch
  })

  it('chống double-submit khi đang gửi', async () => {
    const s = useBudgetsStore()
    s.submitting = true
    await expect(s.create(input)).rejects.toThrow()
    expect(apiMock).not.toHaveBeenCalled()
  })

  it('update gửi PATCH kèm expected_updated_at; xung đột được ném ra cho view', async () => {
    const s = useBudgetsStore()
    apiMock.mockResolvedValueOnce(budget({})) // PATCH OK
    await s.update('b1', { ...input, expected_updated_at: '2026-07-01T00:00:00Z' })
    expect(apiMock).toHaveBeenNthCalledWith(1, '/api/budgets/b1', expect.objectContaining({ method: 'PATCH' }))

    const conflict = new Error('conflict')
    apiMock.mockReset()
    apiMock.mockRejectedValueOnce(conflict)
    await expect(s.update('b1', { ...input, expected_updated_at: 'old' })).rejects.toBe(conflict)
  })

  it('remove gửi DELETE kèm mốc rồi refetch', async () => {
    apiMock.mockResolvedValueOnce(undefined) // DELETE
    const s = useBudgetsStore()
    await s.remove('b1', '2026-07-01T00:00:00Z')
    expect(apiMock).toHaveBeenNthCalledWith(1, '/api/budgets/b1', expect.objectContaining({ method: 'DELETE' }))
    expect(apiMock).toHaveBeenNthCalledWith(2, '/api/budgets?status=ACTIVE')
  })

  it('getters totalBudget / categoryBudgets', async () => {
    const s = useBudgetsStore()
    s.items = [budget({ id: 't', type: 'TOTAL', category_id: null }), budget({ id: 'c1' })]
    expect(s.totalBudget?.id).toBe('t')
    expect(s.categoryBudgets.map((b) => b.id)).toEqual(['c1'])
  })
})
