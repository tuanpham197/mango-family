<script setup lang="ts">
import { useReportsStore } from '../stores/reports'
import { useInvalidation } from '../composables/useInvalidation'
import TimeRangePicker from '../components/TimeRangePicker.vue'
import CategoryBreakdownChart from '../components/CategoryBreakdownChart.vue'
import TrendChart from '../components/TrendChart.vue'
import CategoryDetailPanel from '../components/CategoryDetailPanel.vue'

// Màn Báo cáo (feature 005) — thay placeholder của 004. Tổng quan theo khoảng + phân bổ
// danh mục + xu hướng + drill-in chi tiết danh mục. Realtime cơ hội (D40).
const reports = useReportsStore()

async function onRange(from: string, to: string) {
  reports.setRange(from, to)
  reports.closeCategory()
  await reports.loadOverview()
}
async function onSelect(id: string) {
  await reports.openCategory(id)
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
