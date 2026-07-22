<script setup lang="ts">
import { computed } from 'vue'
import type { BudgetAlert } from '../api/types'

// Cảnh báo in-app: đỏ "vượt {số tiền}" (ưu tiên) hoặc amber "đạt 80%" (FR-007/008,
// SC-005). Hiển thị cho MỌI thành viên (realtime từ budgets_changed — SC-006).
const props = defineProps<{ alerts: BudgetAlert[] }>()

const over = computed(() => props.alerts.find((a) => a.level === 'OVER_100'))
const threshold = computed(() => props.alerts.find((a) => a.level === 'THRESHOLD_80'))

function money(n: number) {
  return new Intl.NumberFormat('vi-VN').format(Math.round(n))
}
</script>

<template>
  <span v-if="over" class="badge over" data-testid="alert-over">
    ⚠ Vượt {{ money(over.over_amount ?? 0) }} ₫
  </span>
  <span v-else-if="threshold" class="badge warn" data-testid="alert-threshold">
    ⚠ Đạt 80%
  </span>
</template>

<style scoped>
.badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 999px;
  white-space: nowrap;
}
.badge.warn {
  color: var(--amber);
  background: #fbf1df;
  border: 1px solid var(--amber);
}
.badge.over {
  color: var(--red);
  background: #fbe6e4;
  border: 1px solid var(--red);
}
</style>
