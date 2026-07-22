import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const apiMock = vi.fn()
vi.mock('../../api/client', () => ({ api: (...a: unknown[]) => apiMock(...a) }))

import { useReportsStore } from '../reports'

describe('reports store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    apiMock.mockReset()
  })

  it('loadOverview gọi API với from/to đã chọn', async () => {
    apiMock.mockResolvedValueOnce({ income: 10, expense: 4, net: 6 })
    const s = useReportsStore()
    s.setRange('2026-07-01', '2026-07-31')
    await s.loadOverview()
    expect(apiMock).toHaveBeenCalledWith('/api/reports/overview?from=2026-07-01&to=2026-07-31')
    expect(s.overview?.net).toBe(6)
  })

  it('loadOverview bỏ qua khi chưa có khoảng', async () => {
    const s = useReportsStore()
    await s.loadOverview()
    expect(apiMock).not.toHaveBeenCalled()
  })

  it('openCategory/closeCategory quản lý drill-in', async () => {
    apiMock.mockResolvedValueOnce({ category_id: 'c1', total: 500 })
    const s = useReportsStore()
    s.setRange('2026-07-01', '2026-07-31')
    await s.openCategory('c1')
    expect(apiMock).toHaveBeenCalledWith('/api/reports/category/c1?from=2026-07-01&to=2026-07-31')
    expect(s.category?.total).toBe(500)
    s.closeCategory()
    expect(s.category).toBeNull()
  })
})
