import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import NetWorthCard from '../NetWorthCard.vue'

describe('NetWorthCard', () => {
  it('hiển thị số tài sản ròng đã format', () => {
    const w = mount(NetWorthCard, { props: { netWorth: 24560000, changePercent: 5.2 } })
    expect(w.get('[data-testid="networth-amount"]').text().replace(/\s/g, '')).toContain('24.560.000đ')
  })

  it('% tăng → mũi tên lên', () => {
    const w = mount(NetWorthCard, { props: { netWorth: 100, changePercent: 5.2 } })
    const c = w.get('[data-testid="networth-change"]')
    expect(c.text()).toContain('▲')
    expect(c.text()).toContain('5.2%')
  })

  it('% giảm → mũi tên xuống, hiển thị trị tuyệt đối', () => {
    const w = mount(NetWorthCard, { props: { netWorth: 100, changePercent: -3 } })
    const c = w.get('[data-testid="networth-change"]')
    expect(c.text()).toContain('▼')
    expect(c.text()).toContain('3%')
  })

  it('changePercent null → ẩn dòng thay đổi (D28)', () => {
    const w = mount(NetWorthCard, { props: { netWorth: 100, changePercent: null } })
    expect(w.find('[data-testid="networth-change"]').exists()).toBe(false)
  })

  it('tài sản ròng âm → thẻ đỏ (class negative) + số có dấu trừ', () => {
    const w = mount(NetWorthCard, { props: { netWorth: -5000000, changePercent: -12 } })
    const card = w.get('[data-testid="networth-card"]')
    expect(card.classes()).toContain('negative')
    expect(w.get('[data-testid="networth-amount"]').text().replace(/\s/g, '')).toContain('-5.000.000đ')
    // % giảm so tháng trước vẫn hiển thị với mũi tên xuống
    expect(w.get('[data-testid="networth-change"]').text()).toContain('▼')
  })

  it('tài sản ròng dương → KHÔNG có class negative (thẻ xanh)', () => {
    const w = mount(NetWorthCard, { props: { netWorth: 100, changePercent: 5.2 } })
    expect(w.get('[data-testid="networth-card"]').classes()).not.toContain('negative')
  })
})
