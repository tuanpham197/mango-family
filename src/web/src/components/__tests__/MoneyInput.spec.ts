import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import MoneyInput from '../MoneyInput.vue'

function val(w: ReturnType<typeof mount>) {
  return (w.get('input').element as HTMLInputElement).value
}
function lastEmit(w: ReturnType<typeof mount>) {
  const e = w.emitted('update:modelValue')!
  return e[e.length - 1]
}

describe('MoneyInput', () => {
  it('gõ số → format dấu ngăn cách + emit giá trị số', async () => {
    const w = mount(MoneyInput, { props: { modelValue: null } })
    await w.get('input').setValue('3500000')
    expect(val(w)).toBe('3.500.000')
    expect(lastEmit(w)).toEqual([3500000])
  })

  it('loại ký tự không phải số khi gõ', async () => {
    const w = mount(MoneyInput, { props: { modelValue: null } })
    await w.get('input').setValue('1a2b3')
    expect(val(w)).toBe('123')
    expect(lastEmit(w)).toEqual([123])
  })

  it('rỗng → emit null', async () => {
    const w = mount(MoneyInput, { props: { modelValue: 5000 } })
    await w.get('input').setValue('')
    expect(lastEmit(w)).toEqual([null])
  })

  it('prefill từ prop (form sửa) → hiển thị đã format', () => {
    const w = mount(MoneyInput, { props: { modelValue: 80000 } })
    expect(val(w)).toBe('80.000')
  })

  it('prop đổi từ ngoài → cập nhật hiển thị', async () => {
    const w = mount(MoneyInput, { props: { modelValue: null } })
    await w.setProps({ modelValue: 1234567 })
    expect(val(w)).toBe('1.234.567')
  })
})
