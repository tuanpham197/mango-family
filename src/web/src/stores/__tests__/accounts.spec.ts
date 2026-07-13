import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import type { Account } from '../../api/types'

const apiMock = vi.fn()
vi.mock('../../api/client', () => ({ api: (...a: unknown[]) => apiMock(...a) }))

import { useAccountsStore } from '../accounts'

const sample: Account[] = [
  { id: 'cash', name: 'Tiền mặt', type: 'CASH', balance: -50000 },
  { id: 'bank', name: 'Vietcombank', type: 'BANK', balance: 200000 },
]

describe('accounts store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    apiMock.mockReset()
  })

  it('total cộng dồn số dư mọi tài khoản', async () => {
    apiMock.mockResolvedValueOnce(sample)
    const store = useAccountsStore()
    await store.fetch()
    expect(store.total).toBe(150000)
  })

  it('defaultId là tài khoản đầu tiên (chọn sẵn khi hộ có ≥1)', async () => {
    apiMock.mockResolvedValueOnce(sample)
    const store = useAccountsStore()
    await store.fetch()
    expect(store.defaultId).toBe('cash')
  })

  it('defaultId null khi chưa có tài khoản', () => {
    const store = useAccountsStore()
    expect(store.defaultId).toBeNull()
  })

  it('ensure chỉ fetch một lần', async () => {
    apiMock.mockResolvedValue(sample)
    const store = useAccountsStore()
    await store.ensure()
    await store.ensure()
    expect(apiMock).toHaveBeenCalledTimes(1)
  })
})
