import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import BudgetAlertBadge from '../BudgetAlertBadge.vue'
import type { BudgetAlert } from '../../api/types'

const t80: BudgetAlert = { level: 'THRESHOLD_80', over_amount: null, fired_at: '' }
const over: BudgetAlert = { level: 'OVER_100', over_amount: 50000, fired_at: '' }

describe('BudgetAlertBadge', () => {
  it('không có cảnh báo → không hiển thị gì', () => {
    const w = mount(BudgetAlertBadge, { props: { alerts: [] } })
    expect(w.find('[data-testid="alert-over"]').exists()).toBe(false)
    expect(w.find('[data-testid="alert-threshold"]').exists()).toBe(false)
  })

  it('chỉ 80% → badge amber "Đạt 80%"', () => {
    const w = mount(BudgetAlertBadge, { props: { alerts: [t80] } })
    expect(w.get('[data-testid="alert-threshold"]').text()).toContain('80%')
  })

  it('vượt → badge đỏ kèm đúng số tiền vượt (SC-005), ưu tiên hơn 80%', () => {
    const w = mount(BudgetAlertBadge, { props: { alerts: [t80, over] } })
    expect(w.find('[data-testid="alert-threshold"]').exists()).toBe(false)
    expect(w.get('[data-testid="alert-over"]').text().replace(/\s/g, '')).toContain('50.000')
  })
})
