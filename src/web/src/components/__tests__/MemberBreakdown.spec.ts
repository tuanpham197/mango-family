import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import MemberBreakdown from '../MemberBreakdown.vue'
import type { MembersReport } from '../../api/types'

const report: MembersReport = {
  from: '2026-08-01',
  to: '2026-08-31',
  members: [
    { member_id: 'a', display_name: 'Alice', is_former: false, income: 12000000, expense: 4500000, net: 7500000 },
    { member_id: 'c', display_name: 'Carol', is_former: false, income: 0, expense: 0, net: 0 },
    { member_id: 'former', display_name: 'Thành viên cũ / Đã rời hộ', is_former: true, income: 0, expense: 800000, net: -800000 },
  ],
  totals: { income: 12000000, expense: 5300000, net: 6700000 },
}

describe('MemberBreakdown (008)', () => {
  it('hiển thị mọi thành viên (kể cả 0/0/0) + dòng Thành viên cũ + tổng hộ', () => {
    const w = mount(MemberBreakdown, { props: { report } })
    const rows = w.findAll('[data-testid="member-row"]')
    expect(rows).toHaveLength(3)
    expect(w.findAll('[data-testid="member-name"]')[1].text()).toBe('Carol') // 0/0/0 vẫn hiện
    expect(w.get('[data-testid="member-totals"]').text().replace(/\s/g, '')).toContain('6.700.000')
  })

  it('bấm một dòng phát select với member_id (drill-in)', async () => {
    const w = mount(MemberBreakdown, { props: { report } })
    await w.findAll('[data-testid="member-row"]')[0].trigger('click')
    expect(w.emitted('select')?.[0]).toEqual(['a'])
  })
})
