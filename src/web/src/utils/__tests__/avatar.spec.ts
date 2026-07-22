import { describe, expect, it } from 'vitest'
import { initials } from '../avatar'

describe('initials', () => {
  it('tên một từ → 1 chữ', () => {
    expect(initials('Alice')).toBe('A')
  })
  it('tên nhiều từ → chữ đầu của từ đầu + từ cuối', () => {
    expect(initials('Nguyễn Văn A')).toBe('NA')
  })
  it('viết hoa + bỏ khoảng trắng thừa', () => {
    expect(initials('  bob   smith ')).toBe('BS')
  })
  it('rỗng/undefined → "?"', () => {
    expect(initials('')).toBe('?')
    expect(initials(undefined)).toBe('?')
    expect(initials(null)).toBe('?')
  })
})
