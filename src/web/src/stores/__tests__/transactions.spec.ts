import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const apiMock = vi.fn()
const apiPagedMock = vi.fn()
vi.mock('../../api/client', () => ({
  api: (...a: unknown[]) => apiMock(...a),
  apiPaged: (...a: unknown[]) => apiPagedMock(...a),
}))

import { useTransactionsStore } from '../transactions'

const input = {
  amount: 50000,
  type: 'EXPENSE' as const,
  category_id: 'c1',
  account_id: 'a1',
}

describe('transactions store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    apiMock.mockReset()
    apiPagedMock.mockReset()
    apiPagedMock.mockResolvedValue({ data: [], paging: { page: 1, page_size: 50, total: 0 } })
  })

  it('create gọi POST rồi refetch sổ', async () => {
    apiMock.mockResolvedValueOnce({ id: 't1' })
    const store = useTransactionsStore()
    await store.create(input)
    expect(apiMock).toHaveBeenCalledWith('/api/transactions', expect.objectContaining({ method: 'POST' }))
    expect(apiPagedMock).toHaveBeenCalled() // refetch
  })

  it('chống double-submit: gọi create khi đang submitting bị từ chối', async () => {
    const store = useTransactionsStore()
    store.submitting = true
    await expect(store.create(input)).rejects.toThrow()
    expect(apiMock).not.toHaveBeenCalled()
  })

  it('submitting được reset về false sau khi create xong', async () => {
    apiMock.mockResolvedValueOnce({ id: 't1' })
    const store = useTransactionsStore()
    await store.create(input)
    expect(store.submitting).toBe(false)
  })

  it('hasMore đúng theo paging', async () => {
    apiPagedMock.mockResolvedValueOnce({ data: [{ id: 'a' }], paging: { page: 1, page_size: 50, total: 3 } })
    const store = useTransactionsStore()
    await store.fetch()
    expect(store.hasMore).toBe(true)
  })

  it('update gửi PATCH kèm expected_updated_at', async () => {
    apiMock.mockResolvedValueOnce({ id: 't1' })
    const store = useTransactionsStore()
    await store.update('t1', { ...input, expected_updated_at: '2026-01-01T00:00:00Z' })
    expect(apiMock).toHaveBeenCalledWith('/api/transactions/t1', expect.objectContaining({ method: 'PATCH' }))
  })
})
