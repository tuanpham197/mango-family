<script setup lang="ts">
import type { MonthCashflow } from '../api/types'

// Hai thẻ Thu nhập / Chi phí tháng này (dashboard.png — dưới tài sản ròng).
defineProps<{ month: MonthCashflow }>()

function money(n: number) {
  return new Intl.NumberFormat('vi-VN').format(Math.round(n))
}
</script>

<template>
  <div class="ie-row">
    <div class="card ie" data-testid="income-card">
      <div class="ie-label">Thu nhập</div>
      <div class="ie-amount income" data-testid="income-amount">+{{ money(month.income) }}</div>
    </div>
    <div class="card ie" data-testid="expense-card">
      <div class="ie-label">Chi phí</div>
      <div class="ie-amount expense" data-testid="expense-amount">−{{ money(month.expense) }}</div>
    </div>
  </div>
</template>

<style scoped>
.ie-row {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
}
.ie {
  flex: 1;
  margin-bottom: 0;
}
.ie-label {
  font-size: 13px;
  color: var(--muted);
}
.ie-amount {
  font-size: 20px;
  font-weight: 700;
  margin-top: 4px;
}
.ie-amount.income {
  color: var(--green);
}
.ie-amount.expense {
  color: var(--red);
}
</style>
