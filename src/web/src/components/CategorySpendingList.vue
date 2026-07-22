<script setup lang="ts">
import type { CategorySpending } from '../api/types'

// Chi tiêu theo danh mục tháng này (MỚI — trước Ngân sách, D29): số tiền + tỷ trọng %,
// sắp giảm dần (đã sắp ở server). Danh mục ẩn vẫn hiển thị kèm nhãn.
defineProps<{ items: CategorySpending[] }>()

function money(n: number) {
  return new Intl.NumberFormat('vi-VN').format(Math.round(n))
}
</script>

<template>
  <div class="card" data-testid="category-spending">
    <h2>Chi tiêu theo danh mục</h2>
    <p v-if="items.length === 0" class="muted" data-testid="category-spending-empty">
      Chưa có chi tiêu trong tháng này.
    </p>
    <div v-for="c in items" :key="c.category_id" class="cs-row" data-testid="category-spending-row">
      <span class="cs-name">
        {{ c.category_name }}
        <span v-if="c.category_hidden" class="muted" data-testid="cs-hidden"> (đã ẩn)</span>
      </span>
      <span class="cs-value">
        {{ money(c.amount) }} đ <span class="muted">({{ c.percent }}%)</span>
      </span>
    </div>
  </div>
</template>

<style scoped>
.card h2 {
  font-size: 16px;
  margin: 0 0 10px;
}
.cs-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid var(--border);
  font-size: 15px;
}
.cs-row:last-child {
  border-bottom: none;
}
.cs-value {
  font-weight: 600;
  white-space: nowrap;
}
</style>
