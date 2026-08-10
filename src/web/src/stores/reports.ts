import { defineStore } from 'pinia'
import { api } from '../api/client'
import type { CategoryReport, MemberReport, MembersReport, ReportOverview } from '../api/types'

// Báo cáo (feature 005 + 008) — chỉ ĐỌC, tổng hợp ở server theo [from,to] (D34). Giữ khoảng
// đang chọn + báo cáo tổng quan + drill-in một danh mục (005) + báo cáo theo thành viên +
// drill-in giao dịch một thành viên (008). Quay lại drill-in KHÔNG đổi khoảng (FR-007).
export const useReportsStore = defineStore('reports', {
  state: () => ({
    from: '' as string, // YYYY-MM-DD
    to: '' as string,
    overview: null as ReportOverview | null,
    category: null as CategoryReport | null, // đang drill-in một danh mục (null = không)
    members: null as MembersReport | null, // báo cáo theo thành viên (008)
    memberDetail: null as MemberReport | null, // đang drill-in một thành viên (null = không)
    loading: false,
  }),
  actions: {
    setRange(from: string, to: string) {
      this.from = from
      this.to = to
    },
    async loadOverview() {
      if (!this.from || !this.to) return
      this.loading = true
      try {
        this.overview = await api<ReportOverview>(`/api/reports/overview?from=${this.from}&to=${this.to}`)
      } finally {
        this.loading = false
      }
    },
    async openCategory(id: string) {
      this.category = await api<CategoryReport>(`/api/reports/category/${id}?from=${this.from}&to=${this.to}`)
    },
    closeCategory() {
      this.category = null
    },
    /** Báo cáo theo thành viên cho khoảng hiện tại (008 — FR-001/002). */
    async loadMembers() {
      if (!this.from || !this.to) return
      this.members = await api<MembersReport>(`/api/reports/members?from=${this.from}&to=${this.to}`)
    },
    /** Drill-in giao dịch của một thành viên (id UUID hoặc 'former'), phân trang (008 — FR-006/007). */
    async openMember(id: string, page = 1) {
      this.memberDetail = await api<MemberReport>(
        `/api/reports/member/${id}?from=${this.from}&to=${this.to}&page=${page}`,
      )
    },
    closeMember() {
      this.memberDetail = null
    },
    /** Làm mới mọi phần đang mở (đổi khoảng / dữ liệu nền đổi — D40). Giữ trang drill-in hiện tại. */
    async refresh() {
      await this.loadOverview()
      if (this.category) await this.openCategory(this.category.category_id)
      if (this.members) await this.loadMembers()
      if (this.memberDetail) await this.openMember(this.memberDetail.member_id, this.memberDetail.page)
    },
  },
})
