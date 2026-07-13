import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError, api, apiPaged } from '../client'

describe('api client', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('trả về phần data khi thành công', async () => {
    ;(fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ data: { id: '1' } }),
    })
    const result = await api<{ id: string }>('/api/x')
    expect(result).toEqual({ id: '1' })
  })

  it('204 trả về data undefined không parse body', async () => {
    ;(fetch as ReturnType<typeof vi.fn>).mockResolvedValue({ ok: true, status: 204 })
    const result = await api('/api/x', { method: 'DELETE' })
    expect(result).toBeUndefined()
  })

  it('lỗi ném ApiError giữ code + field + extra', async () => {
    ;(fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: false,
      status: 422,
      json: async () => ({
        error: { code: 'CATEGORY_HAS_TRANSACTIONS', message: 'còn giao dịch', transaction_count: 3 },
      }),
    })
    await expect(api('/api/x')).rejects.toMatchObject({
      status: 422,
      code: 'CATEGORY_HAS_TRANSACTIONS',
    })
    try {
      await api('/api/x')
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError)
      expect((e as ApiError).body.transaction_count).toBe(3)
    }
  })

  it('apiPaged trả về data + paging', async () => {
    ;(fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ data: [1, 2], paging: { page: 1, page_size: 50, total: 2 } }),
    })
    const res = await apiPaged<number[]>('/api/x')
    expect(res.data).toEqual([1, 2])
    expect(res.paging?.total).toBe(2)
  })

  it('gửi Content-Type khi có body', async () => {
    const spy = fetch as ReturnType<typeof vi.fn>
    spy.mockResolvedValue({ ok: true, status: 200, json: async () => ({ data: null }) })
    await api('/api/x', { method: 'POST', body: JSON.stringify({ a: 1 }) })
    expect(spy).toHaveBeenCalledWith(
      '/api/x',
      expect.objectContaining({
        credentials: 'same-origin',
        headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
      }),
    )
  })
})
