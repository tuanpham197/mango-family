<script setup lang="ts">
import { computed } from 'vue'
import type { Budget } from '../api/types'

// Thanh tiến độ "đã chi/giới hạn (%)" + mã màu ngưỡng (BR-BGT-006/007):
// bình thường (green) · amber ≥ 80% · đỏ khi vượt 100%.
const props = defineProps<{ budget: Budget }>()

const pct = computed(() => (props.budget.limit_amount > 0 ? (props.budget.spent / props.budget.limit_amount) * 100 : 0))
const level = computed<'over' | 'warn' | 'ok'>(() => {
  if (props.budget.spent > props.budget.limit_amount) return 'over'
  if (pct.value >= 80) return 'warn'
  return 'ok'
})
const barWidth = computed(() => Math.min(pct.value, 100))

function money(n: number) {
  return new Intl.NumberFormat('vi-VN').format(Math.round(n))
}
function pctLabel(n: number) {
  // Bỏ đuôi ,0 (70%) nhưng giữ phần lẻ có nghĩa (87,5%).
  return new Intl.NumberFormat('vi-VN', { maximumFractionDigits: 1 }).format(n)
}
</script>

<template>
  <div class="progress" :data-testid="`budget-progress-${budget.id}`" :data-level="level">
    <div class="progress-text">
      <span data-testid="progress-amount">{{ money(budget.spent) }}/{{ money(budget.limit_amount) }} ₫</span>
      <span class="progress-pct" :class="level" data-testid="progress-percent">({{ pctLabel(pct) }}%)</span>
    </div>
    <div class="progress-track">
      <div class="progress-fill" :class="level" :style="{ width: barWidth + '%' }"></div>
    </div>
  </div>
</template>

<style scoped>
.progress {
  width: 100%;
}
.progress-text {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  margin-bottom: 4px;
}
.progress-pct {
  font-weight: 600;
}
.progress-pct.ok {
  color: var(--green);
}
.progress-pct.warn {
  color: var(--amber);
}
.progress-pct.over {
  color: var(--red);
}
.progress-track {
  height: 8px;
  border-radius: 6px;
  background: var(--border);
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  border-radius: 6px;
  transition: width 0.3s ease;
}
.progress-fill.ok {
  background: var(--green);
}
.progress-fill.warn {
  background: var(--amber);
}
.progress-fill.over {
  background: var(--red);
}
</style>
