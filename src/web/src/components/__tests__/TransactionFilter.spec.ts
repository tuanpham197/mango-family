import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import TransactionFilter from '../TransactionFilter.vue'

function currentMonth(): [string, string] {
  const now = new Date()
  const y = now.getFullYear()
  const m = String(now.getMonth() + 1).padStart(2, '0')
  const last = new Date(y, now.getMonth() + 1, 0).getDate()
  return [`${y}-${m}-01`, `${y}-${m}-${String(last).padStart(2, '0')}`]
}

describe('TransactionFilter', () => {
  it('phát khoảng tháng hiện tại khi mount', () => {
    const w = mount(TransactionFilter)
    const ev = w.emitted('change')
    expect(ev).toBeTruthy()
    const [from, to] = ev![0] as [string, string]
    const [ef, et] = currentMonth()
    expect(from).toBe(ef)
    expect(to).toBe(et)
  })

  it('chọn "Cả năm" (month=0) → khoảng cả năm', async () => {
    const w = mount(TransactionFilter)
    await w.get('[data-testid="filter-month"]').setValue('0')
    const ev = w.emitted('change')!
    const [from, to] = ev[ev.length - 1] as [string, string]
    const y = new Date().getFullYear()
    expect(from).toBe(`${y}-01-01`)
    expect(to).toBe(`${y}-12-31`)
  })

  it('đổi năm phát lại khoảng theo năm mới', async () => {
    const w = mount(TransactionFilter)
    const y = new Date().getFullYear()
    await w.get('[data-testid="filter-year"]').setValue(String(y - 1))
    const ev = w.emitted('change')!
    const [from] = ev[ev.length - 1] as [string, string]
    expect(from.startsWith(String(y - 1))).toBe(true)
  })
})
