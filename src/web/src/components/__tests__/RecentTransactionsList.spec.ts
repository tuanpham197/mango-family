import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import type { Transaction } from '../../api/types'

const { pushMock } = vi.hoisted(() => ({ pushMock: vi.fn() }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: pushMock }) }))

import RecentTransactionsList from '../RecentTransactionsList.vue'

function txn(o: Partial<Transaction>): Transaction {
  return {
    id: 't', household_id: 'h', created_by: 'u', amount: 0, type: 'EXPENSE', category_id: 'c',
    account_id: 'a', description: null, transaction_date: '', updated_at: '',
    category_name: 'Ăn uống', account_name: 'Tiền mặt', created_by_name: 'Alice', ...o,
  }
}

describe('RecentTransactionsList', () => {
  beforeEach(() => pushMock.mockReset())

  it('hiển thị Chi (−, đỏ) và Thu (+, xanh) với danh mục · tài khoản', () => {
    const w = mount(RecentTransactionsList, {
      props: {
        items: [
          txn({ id: '1', amount: 45000, type: 'EXPENSE', description: 'Bún bò Huế' }),
          txn({ id: '2', amount: 18000000, type: 'INCOME', category_name: 'Lương' }),
        ],
      },
    })
    const rows = w.findAll('[data-testid="recent-row"]')
    expect(rows).toHaveLength(2)
    expect(rows[0].text()).toContain('Bún bò Huế')
    expect(rows[0].text()).toContain('Ăn uống · Tiền mặt')
    expect(rows[0].get('.amount-expense').text().replace(/\s/g, '')).toBe('−45.000')
    expect(rows[1].get('.amount-income').text().replace(/\s/g, '')).toBe('+18.000.000')
  })

  it('rỗng → empty state', () => {
    const w = mount(RecentTransactionsList, { props: { items: [] } })
    expect(w.find('[data-testid="recent-empty"]').exists()).toBe(true)
  })

  it('"Xem tất cả ›" điều hướng tới màn Giao dịch (/ledger)', async () => {
    const w = mount(RecentTransactionsList, { props: { items: [] } })
    await w.get('[data-testid="recent-see-all"]').trigger('click')
    expect(pushMock).toHaveBeenCalledWith('/ledger')
  })
})
