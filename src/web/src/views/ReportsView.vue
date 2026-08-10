<script setup lang="ts">
import { useReportsStore } from '../stores/reports'
import { useInvalidation } from '../composables/useInvalidation'
import TimeRangePicker from '../components/TimeRangePicker.vue'
import CategoryBreakdownChart from '../components/CategoryBreakdownChart.vue'
import TrendChart from '../components/TrendChart.vue'
import CategoryDetailPanel from '../components/CategoryDetailPanel.vue'
import MemberBreakdown from '../components/MemberBreakdown.vue'
import MemberTransactionsPanel from '../components/MemberTransactionsPanel.vue'

// Màn Báo cáo (feature 005 + 008) — Tổng quan theo khoảng + phân bổ danh mục + xu hướng +
// drill-in danh mục; và báo cáo THEO THÀNH VIÊN + drill-in giao dịch thành viên. Realtime cơ hội (D40).
const reports = useReportsStore()

async function onRange(from: string, to: string) {
  reports.setRange(from, to)
  reports.closeCategory()
  reports.closeMember()
  await Promise.all([reports.loadOverview(), reports.loadMembers()])
}
async function onSelect(id: string) {
  await reports.openCategory(id)
}
async function onSelectMember(id: string) {
  await reports.openMember(id, 1)
}
async function onMemberPage(page: number) {
  if (reports.memberDetail) await reports.openMember(reports.memberDetail.member_id, page)
}

useInvalidation('transactions_changed', () => reports.refresh())
useInvalidation('categories_changed', () => reports.refresh())

function money(n: number) {
  return new Intl.NumberFormat('vi-VN').format(Math.round(n))
}
</script>

<template>
  <div>
    <div class="page-head"><h1>Báo cáo</h1></div>

    <TimeRangePicker @change="onRange" />

    <template v-if="reports.overview">
      <div class="card summary" data-testid="report-summary">
        <div class="sum-cell">
          <span class="muted">Thu nhập</span>
          <div class="amount-income" data-testid="report-income">+{{ money(reports.overview.income) }}</div>
        </div>
        <div class="sum-cell">
          <span class="muted">Chi phí</span>
          <div class="amount-expense" data-testid="report-expense">−{{ money(reports.overview.expense) }}</div>
        </div>
        <div class="sum-cell">
          <span class="muted">Số dư ròng</span>
          <div :class="reports.overview.net >= 0 ? 'amount-income' : 'amount-expense'" data-testid="report-net">
            {{ money(reports.overview.net) }}
          </div>
        </div>
      </div>

      <CategoryBreakdownChart :items="reports.overview.category_breakdown" @select="onSelect" />
      <TrendChart :trend="reports.overview.trend" />
      <CategoryDetailPanel v-if="reports.category" :report="reports.category" @close="reports.closeCategory()" />

      <MemberBreakdown v-if="reports.members" :report="reports.members" @select="onSelectMember" />
      <MemberTransactionsPanel
        v-if="reports.memberDetail"
        :report="reports.memberDetail"
        @close="reports.closeMember()"
        @page="onMemberPage"
      />
    </template>

    <p v-else-if="reports.loading" class="muted">Đang tải…</p>
  </div>
</template>

<style scoped>
.summary {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  text-align: center;
}
.sum-cell {
  flex: 1;
}
.sum-cell .muted {
  font-size: 12px;
}
.sum-cell > div {
  font-size: 18px;
  font-weight: 700;
  margin-top: 2px;
}
</style>
