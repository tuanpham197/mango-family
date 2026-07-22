import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import TimeRangePicker from '../TimeRangePicker.vue'

describe('TimeRangePicker', () => {
  it('mặc định phát khoảng "tháng này" khi mount', () => {
    const w = mount(TimeRangePicker)
    const ev = w.emitted('change')
    expect(ev).toBeTruthy()
    const [from, to] = ev![0] as [string, string]
    expect(from).toMatch(/^\d{4}-\d{2}-01$/) // đầu tháng
    expect(to >= from).toBe(true)
  })

  it('chọn "Tuần này" phát khoảng 7 ngày (Thứ Hai → Chủ Nhật)', async () => {
    const w = mount(TimeRangePicker)
    await w.get('[data-testid="range-week"]').trigger('click')
    const ev = w.emitted('change')!
    const [from, to] = ev[ev.length - 1] as [string, string]
    const days = (Date.parse(to) - Date.parse(from)) / 86400000
    expect(days).toBe(6)
  })

  it('tùy chỉnh end < start → hiện lỗi, không phát khoảng sai', async () => {
    const w = mount(TimeRangePicker)
    await w.get('[data-testid="range-custom"]').trigger('click')
    const before = (w.emitted('change') ?? []).length
    await w.get('[data-testid="custom-from"]').setValue('2026-07-31')
    await w.get('[data-testid="custom-to"]').setValue('2026-07-01')
    await w.get('[data-testid="custom-to"]').trigger('change')
    expect(w.get('[data-testid="range-error"]').text()).toBeTruthy()
    expect((w.emitted('change') ?? []).length).toBe(before) // không phát khoảng không hợp lệ
  })
})
