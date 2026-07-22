import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import IncomeExpenseCards from '../IncomeExpenseCards.vue'

describe('IncomeExpenseCards', () => {
  it('hiển thị Thu (+) và Chi (−) đã format', () => {
    const w = mount(IncomeExpenseCards, { props: { month: { income: 18200000, expense: 9400000, net: 8800000 } } })
    expect(w.get('[data-testid="income-amount"]').text().replace(/\s/g, '')).toBe('+18.200.000')
    expect(w.get('[data-testid="expense-amount"]').text().replace(/\s/g, '')).toBe('−9.400.000')
  })

  it('màu: Thu xanh, Chi đỏ', () => {
    const w = mount(IncomeExpenseCards, { props: { month: { income: 1, expense: 2, net: -1 } } })
    expect(w.get('[data-testid="income-amount"]').classes()).toContain('income')
    expect(w.get('[data-testid="expense-amount"]').classes()).toContain('expense')
  })
})
