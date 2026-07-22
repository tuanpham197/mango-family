import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import BudgetProgressBar from '../BudgetProgressBar.vue'
import type { Budget } from '../../api/types'

function budget(o: Partial<Budget>): Budget {
  return {
    id: 'b1', household_id: 'h', type: 'CATEGORY', category_id: 'c', category_name: 'Ăn uống',
    category_hidden: false, limit_amount: 5000000, period_type: 'MONTHLY', start_date: null,
    end_date: null, status: 'ACTIVE', period_key: '2026-07', spent: 0, percent: 0, alerts: [],
    created_by: 'u', created_by_name: 'Alice', updated_at: '', ...o,
  }
}

describe('BudgetProgressBar', () => {
  it('hiển thị "đã chi/giới hạn (%)" đúng định dạng (70%)', () => {
    const w = mount(BudgetProgressBar, { props: { budget: budget({ spent: 3500000, limit_amount: 5000000 }) } })
    const amount = w.get('[data-testid="progress-amount"]').text()
    expect(amount.replace(/\s/g, '')).toContain('3.500.000/5.000.000₫')
    expect(w.get('[data-testid="progress-percent"]').text()).toContain('70')
    expect(w.get('[data-testid="budget-progress-b1"]').attributes('data-level')).toBe('ok')
  })

  it('≥ 80% → mức cảnh báo amber (warn), giữ phần lẻ 87,5%', () => {
    const w = mount(BudgetProgressBar, { props: { budget: budget({ spent: 3500000, limit_amount: 4000000 }) } })
    expect(w.get('[data-testid="progress-percent"]').text()).toMatch(/87[.,]5/)
    expect(w.get('[data-testid="budget-progress-b1"]').attributes('data-level')).toBe('warn')
  })

  it('vượt 100% → mức đỏ (over), bề rộng thanh giới hạn 100%', () => {
    const w = mount(BudgetProgressBar, { props: { budget: budget({ spent: 6000000, limit_amount: 5000000 }) } })
    expect(w.get('[data-testid="budget-progress-b1"]').attributes('data-level')).toBe('over')
    const fill = w.get('.progress-fill')
    expect(fill.attributes('style')).toContain('width: 100%')
  })
})
