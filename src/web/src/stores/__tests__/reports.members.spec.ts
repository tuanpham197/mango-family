import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const apiMock = vi.fn()
vi.mock('../../api/client', () => ({ api: (...a: unknown[]) => apiMock(...a) }))

import { useReportsStore } from '../reports'

describe('reports store — theo thành viên (008)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    apiMock.mockReset()
  })

  it('loadMembers gọi API với from/to đã chọn', async () => {
    apiMock.mockResolvedValueOnce({ members: [{ member_id: 'a', display_name: 'Alice' }], totals: { income: 1, expense: 0, net: 1 } })
    const s = useReportsStore()
    s.setRange('2026-08-01', '2026-08-31')
    await s.loadMembers()
    expect(apiMock).toHaveBeenCalledWith('/api/reports/members?from=2026-08-01&to=2026-08-31')
    expect(s.members?.members[0].display_name).toBe('Alice')
  })

  it('loadMembers bỏ qua khi chưa chọn khoảng', async () => {
    const s = useReportsStore()
    await s.loadMembers()
    expect(apiMock).not.toHaveBeenCalled()
  })

  it('openMember truyền id + page; closeMember xoá drill-in', async () => {
    apiMock.mockResolvedValueOnce({ member_id: 'a', display_name: 'Alice', page: 2, total: 30, transactions: [] })
    const s = useReportsStore()
    s.setRange('2026-08-01', '2026-08-31')
    await s.openMember('a', 2)
    expect(apiMock).toHaveBeenCalledWith('/api/reports/member/a?from=2026-08-01&to=2026-08-31&page=2')
    expect(s.memberDetail?.page).toBe(2)
    s.closeMember()
    expect(s.memberDetail).toBeNull()
  })

  it('openMember hỗ trợ sentinel former', async () => {
    apiMock.mockResolvedValueOnce({ member_id: 'former', display_name: 'Thành viên cũ', is_former: true, page: 1, total: 1, transactions: [] })
    const s = useReportsStore()
    s.setRange('2026-08-01', '2026-08-31')
    await s.openMember('former')
    expect(apiMock).toHaveBeenCalledWith('/api/reports/member/former?from=2026-08-01&to=2026-08-31&page=1')
    expect(s.memberDetail?.is_former).toBe(true)
  })

  it('refresh giữ khoảng + trang drill-in hiện tại', async () => {
    const s = useReportsStore()
    s.setRange('2026-08-01', '2026-08-31')
    // overview + members + memberDetail đang mở
    s.overview = { income: 0, expense: 0, net: 0 } as never
    s.members = { members: [], totals: { income: 0, expense: 0, net: 0 } } as never
    s.memberDetail = { member_id: 'a', page: 3, total: 90, transactions: [] } as never
    apiMock.mockResolvedValue({})
    await s.refresh()
    // memberDetail refetch với đúng page 3 và khoảng cũ
    expect(apiMock).toHaveBeenCalledWith('/api/reports/member/a?from=2026-08-01&to=2026-08-31&page=3')
  })
})
