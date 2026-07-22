<script setup lang="ts">
import { computed } from 'vue'
import type { CategoryReport } from '../api/types'
import LineChart from './charts/LineChart.vue'

// Báo cáo chi tiết theo danh mục (BR-RPT-004/005/006): tổng + xu hướng + danh sách giao dịch.
const props = defineProps<{ report: CategoryReport }>()
const emit = defineEmits<{ close: [] }>()

const labels = computed(() => props.report.trend.map((t) => t.bucket))
const series = computed(() => [{ label: props.report.category_name, data: props.report.trend.map((t) => t.amount), color: '#d98a1f' }])

function money(n: number) {
  return new Intl.NumberFormat('vi-VN').format(Math.round(n))
}
function fmtDate(iso: string) {
  return new Date(iso).toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit', year: 'numeric' })
}
</script>

<template>
  <div class="card" data-testid="category-detail">
    <div class="detail-head">
      <h2>{{ report.category_name }} · {{ money(report.total) }} đ</h2>
      <button type="button" class="link-btn" data-testid="close-category-detail" @click="emit('close')">Đóng ✕</button>
    </div>

    <LineChart v-if="report.trend.length" :labels="labels" :series="series" />

    <p v-if="report.transactions.length === 0" class="muted" data-testid="detail-empty">Không có giao dịch trong khoảng này.</p>
    <div v-for="t in report.transactions" :key="t.id" class="detail-row" data-testid="detail-txn">
      <div class="detail-main">
        <div>{{ t.description || t.category_name }}</div>
        <div class="muted">{{ t.account_name }} · {{ fmtDate(t.transaction_date) }}</div>
      </div>
      <span class="amount-expense">−{{ money(t.amount) }}</span>
    </div>
  </div>
</template>

<style scoped>
.detail-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}
.detail-head h2 {
  font-size: 16px;
  margin: 0;
}
.link-btn {
  border: none;
  background: transparent;
  color: var(--muted);
  font: inherit;
  cursor: pointer;
}
.detail-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
  border-bottom: 1px solid var(--border);
}
.detail-row:last-child {
  border-bottom: none;
}
.detail-main {
  flex: 1;
  min-width: 0;
}
</style>
