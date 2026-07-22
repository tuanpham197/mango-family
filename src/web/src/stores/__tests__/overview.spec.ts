import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const apiMock = vi.fn()
vi.mock('../../api/client', () => ({ api: (...a: unknown[]) => apiMock(...a) }))

import { useOverviewStore } from '../overview'

describe('overview store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    apiMock.mockReset()
  })

  it('fetch gọi GET /api/overview và gán summary', async () => {
    const summary = {
      net_worth: 100, net_worth_change_percent: 5.2,
      month: { income: 10, expense: 4, net: 6 }, category_spending: [], recent_transactions: [],
    }
    apiMock.mockResolvedValueOnce(summary)
    const s = useOverviewStore()
    await s.fetch()
    expect(apiMock).toHaveBeenCalledWith('/api/overview')
    expect(s.summary).toEqual(summary)
    expect(s.loading).toBe(false)
  })
})
