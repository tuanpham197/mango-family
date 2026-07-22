<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { Budget } from '../api/types'
import BudgetProgressBar from '../components/BudgetProgressBar.vue'
import BudgetAlertBadge from '../components/BudgetAlertBadge.vue'
import DeleteBudgetDialog from '../components/DeleteBudgetDialog.vue'
import { useBudgetsStore } from '../stores/budgets'
import { useInvalidation } from '../composables/useInvalidation'

// Danh sách ngân sách của hộ (UC-BGT-03): tiến độ + cảnh báo, realtime ≤ 5s.
const router = useRouter()
const budgets = useBudgetsStore()
const deleting = ref<Budget | null>(null)
const notice = ref('')

onMounted(() => budgets.fetch())
// Realtime (SC-006): giao dịch đổi → tiến độ đổi; CRUD ngân sách của thành viên khác.
useInvalidation('budgets_changed', () => budgets.fetch())
useInvalidation('transactions_changed', () => budgets.fetch())

function labelOf(b: Budget) {
  return b.type === 'TOTAL' ? 'Tổng chi tiêu' : (b.category_name ?? 'Danh mục')
}
function periodLabel(b: Budget) {
  return { MONTHLY: 'Hàng tháng', WEEKLY: 'Hàng tuần', ONE_TIME: 'Một lần' }[b.period_type]
}
function openEdit(b: Budget) {
  router.push(`/budgets/${b.id}/edit`)
}
function onDeleted(message?: string) {
  deleting.value = null
  if (message) {
    notice.value = message
    setTimeout(() => (notice.value = ''), 3000)
  }
}
</script>

<template>
  <div>
    <div class="page-head">
      <h1>Ngân sách</h1>
      <button class="btn btn-primary" data-testid="new-budget" @click="router.push('/budgets/new')">＋ Tạo</button>
    </div>

    <div v-if="notice" class="form-error" data-testid="notice">{{ notice }}</div>

    <div class="card" data-testid="budget-list">
      <p v-if="budgets.items.length === 0 && !budgets.loading" class="muted" data-testid="empty-budgets">
        Chưa có ngân sách nào — bấm “＋ Tạo” để đặt giới hạn chi tiêu đầu tiên.
      </p>

      <div
        v-for="b in budgets.items"
        :key="b.id"
        class="budget-row"
        data-testid="budget-row"
        @click="openEdit(b)"
      >
        <div class="budget-head">
          <div class="budget-title">
            <span data-testid="budget-label">{{ labelOf(b) }}</span>
            <span class="muted period">· {{ periodLabel(b) }}</span>
            <span v-if="b.category_hidden" class="muted hidden-tag" data-testid="category-hidden-tag">(danh mục đã ẩn)</span>
            <span v-if="b.status === 'ENDED'" class="muted hidden-tag" data-testid="budget-ended-tag">(đã kết thúc)</span>
          </div>
          <div class="budget-actions">
            <BudgetAlertBadge :alerts="b.alerts" />
            <button
              type="button"
              class="row-del"
              :data-testid="`delete-budget-${b.id}`"
              @click.stop="deleting = b"
            >
              ✕
            </button>
          </div>
        </div>
        <BudgetProgressBar :budget="b" />
      </div>
    </div>

    <DeleteBudgetDialog v-if="deleting" :budget="deleting" @close="deleting = null" @done="onDeleted" />
  </div>
</template>

<style scoped>
.budget-row {
  padding: 14px 4px;
  border-bottom: 1px solid var(--border);
  cursor: pointer;
}
.budget-row:last-child {
  border-bottom: none;
}
.budget-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.budget-title {
  font-weight: 600;
}
.period,
.hidden-tag {
  font-weight: 400;
  font-size: 13px;
  margin-left: 4px;
}
.budget-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.row-del {
  border: none;
  background: transparent;
  color: var(--muted);
  font-size: 16px;
  cursor: pointer;
  padding: 4px 6px;
}
.row-del:hover {
  color: var(--red);
}
</style>
