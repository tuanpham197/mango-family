import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import CategorySpendingList from '../CategorySpendingList.vue'
import type { CategorySpending } from '../../api/types'

function cat(o: Partial<CategorySpending>): CategorySpending {
  return { category_id: 'c', category_name: 'X', category_hidden: false, amount: 0, percent: 0, ...o }
}

describe('CategorySpendingList', () => {
  it('liệt kê từng danh mục kèm số tiền + %', () => {
    const w = mount(CategorySpendingList, {
      props: {
        items: [
          cat({ category_id: 'a', category_name: 'Ăn uống', amount: 3500000, percent: 37 }),
          cat({ category_id: 'b', category_name: 'Di chuyển', amount: 900000, percent: 10 }),
        ],
      },
    })
    const rows = w.findAll('[data-testid="category-spending-row"]')
    expect(rows).toHaveLength(2)
    expect(rows[0].text()).toContain('Ăn uống')
    expect(rows[0].text().replace(/\s/g, '')).toContain('3.500.000đ')
    expect(rows[0].text()).toContain('37%')
  })

  it('danh mục ẩn → nhãn "đã ẩn"', () => {
    const w = mount(CategorySpendingList, { props: { items: [cat({ category_hidden: true, amount: 1, percent: 1 })] } })
    expect(w.find('[data-testid="cs-hidden"]').exists()).toBe(true)
  })

  it('rỗng → empty state', () => {
    const w = mount(CategorySpendingList, { props: { items: [] } })
    expect(w.find('[data-testid="category-spending-empty"]').exists()).toBe(true)
    expect(w.findAll('[data-testid="category-spending-row"]')).toHaveLength(0)
  })
})
