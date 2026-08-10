<script setup lang="ts">
import { computed } from 'vue'
import type { MemberReport } from '../api/types'

// Drill-in giao dịch của một thành viên (008 — FR-006/007): danh sách phân trang, trạng
// thái trống, tổng Thu/Chi/ròng; quay lại KHÔNG đổi khoảng (do store giữ from/to). Đổi
// trang phát 'page' để view gọi lại openMember(id, page).
const props = defineProps<{ report: MemberReport }>()
const emit = defineEmits<{ close: []; page: [page: number] }>()

const pageCount = computed(() => Math.max(1, Math.ceil(props.report.total / props.report.page_size)))

function money(n: number) {
  return new Intl.NumberFormat('vi-VN').format(Math.round(n))
}
function fmtDate(iso: string) {
  return new Date(iso).toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit', year: 'numeric' })
}
function go(page: number) {
  if (page >= 1 && page <= pageCount.value && page !== props.report.page) emit('page', page)
}
</script>

<template>
  <div class="card" data-testid="member-detail">
    <div class="detail-head">
      <h2>{{ report.display_name }}</h2>
      <button type="button" class="link-btn" data-testid="close-member-detail" @click="emit('close')">Đóng ✕</button>
    </div>

    <div class="mini-summary">
      <span class="amount-income">+{{ money(report.income) }}</span>
      <span class="amount-expense">−{{ money(report.expense) }}</span>
      <span :class="report.net >= 0 ? 'amount-income' : 'amount-expense'">Ròng {{ money(report.net) }}</span>
    </div>

    <p v-if="report.transactions.length === 0" class="muted" data-testid="member-detail-empty">
      Không có giao dịch trong khoảng này.
    </p>
    <div v-for="t in report.transactions" :key="t.id" class="detail-row" data-testid="member-txn">
      <div class="detail-main">
        <div>{{ t.description || t.category_name }}</div>
        <div class="muted">{{ t.category_name }} · {{ t.account_name }} · {{ fmtDate(t.transaction_date) }}</div>
      </div>
      <span :class="t.type === 'INCOME' ? 'amount-income' : 'amount-expense'">
        {{ t.type === 'INCOME' ? '+' : '−' }}{{ money(t.amount) }}
      </span>
    </div>

    <div v-if="pageCount > 1" class="pager" data-testid="member-pager">
      <button type="button" :disabled="report.page <= 1" data-testid="member-prev" @click="go(report.page - 1)">‹ Trước</button>
      <span class="muted">Trang {{ report.page }}/{{ pageCount }}</span>
      <button type="button" :disabled="report.page >= pageCount" data-testid="member-next" @click="go(report.page + 1)">Sau ›</button>
    </div>
  </div>
</template>

<style scoped>
.detail-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
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
.mini-summary {
  display: flex;
  gap: 14px;
  font-weight: 600;
  margin-bottom: 10px;
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
.pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 12px;
}
.pager button {
  padding: 4px 10px;
}
</style>
