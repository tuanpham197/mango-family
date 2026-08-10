import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import MemberTransactionsPanel from '../MemberTransactionsPanel.vue'
import type { MemberReport } from '../../api/types'

function report(over: Partial<MemberReport> = {}): MemberReport {
  return {
    member_id: 'a',
    display_name: 'Alice',
    is_former: false,
    from: '2026-08-01',
    to: '2026-08-31',
    income: 12000000,
    expense: 4500000,
    net: 7500000,
    transactions: [
      { id: 't1', type: 'EXPENSE', amount: 250000, category_name: 'Ăn uống', account_name: 'Tiền mặt', description: 'Chợ', transaction_date: '2026-08-12T09:30:00+07:00' } as never,
    ],
    page: 1,
    page_size: 20,
    total: 45,
    ...over,
  }
}

describe('MemberTransactionsPanel (008)', () => {
  it('hiển thị giao dịch + phân trang khi total > page_size', () => {
    const w = mount(MemberTransactionsPanel, { props: { report: report() } })
    expect(w.findAll('[data-testid="member-txn"]')).toHaveLength(1)
    expect(w.get('[data-testid="member-pager"]').text()).toContain('1/3') // ceil(45/20)
  })

  it('bấm Sau phát page kế tiếp; đóng phát close', async () => {
    const w = mount(MemberTransactionsPanel, { props: { report: report() } })
    await w.get('[data-testid="member-next"]').trigger('click')
    expect(w.emitted('page')?.[0]).toEqual([2])
    await w.get('[data-testid="close-member-detail"]').trigger('click')
    expect(w.emitted('close')).toBeTruthy()
  })

  it('trạng thái trống rõ ràng, không phân trang', () => {
    const w = mount(MemberTransactionsPanel, { props: { report: report({ transactions: [], total: 0 }) } })
    expect(w.find('[data-testid="member-detail-empty"]').exists()).toBe(true)
    expect(w.find('[data-testid="member-pager"]').exists()).toBe(false)
  })
})
