<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import BudgetProgressBar from './BudgetProgressBar.vue'
import BudgetAlertBadge from './BudgetAlertBadge.vue'
import { useBudgetsStore } from '../stores/budgets'
import { useInvalidation } from '../composables/useInvalidation'

// Tóm tắt ngân sách trên màn Tổng quan (FR-014, wireframe màn 1): tiến độ ngân sách
// tổng (nếu có) + vài danh mục nổi bật, cùng mã màu ngưỡng; realtime + "Xem tất cả".
const router = useRouter()
const budgets = useBudgetsStore()

onMounted(() => budgets.fetch())
useInvalidation('budgets_changed', () => budgets.fetch())
useInvalidation('transactions_changed', () => budgets.fetch())

// Danh mục nổi bật = tiến độ cao nhất (dễ chạm ngưỡng) — tối đa 3.
const highlights = computed(() =>
  [...budgets.categoryBudgets]
    .filter((b) => b.status === 'ACTIVE')
    .sort((a, b) => b.percent - a.percent)
    .slice(0, 3),
)
const hasAny = computed(() => budgets.totalBudget !== null || highlights.value.length > 0)
</script>

<template>
  <div class="card" data-testid="overview-budget-widget">
    <div class="widget-head">
      <h2>Ngân sách tháng này</h2>
      <button type="button" class="link-btn" data-testid="budget-see-all" @click="router.push('/budgets')">
        Xem tất cả ›
      </button>
    </div>

    <p v-if="!hasAny" class="muted" data-testid="overview-no-budget">
      Chưa có ngân sách — <a href="#" @click.prevent="router.push('/budgets/new')">tạo ngân sách</a> để theo dõi chi tiêu.
    </p>

    <div v-if="budgets.totalBudget" class="widget-item" data-testid="overview-total">
      <div class="widget-item-head">
        <span class="widget-label">Tổng chi tiêu</span>
        <BudgetAlertBadge :alerts="budgets.totalBudget.alerts" />
      </div>
      <BudgetProgressBar :budget="budgets.totalBudget" />
    </div>

    <div v-for="b in highlights" :key="b.id" class="widget-item" data-testid="overview-category">
      <div class="widget-item-head">
        <span class="widget-label">{{ b.category_name }}</span>
        <BudgetAlertBadge :alerts="b.alerts" />
      </div>
      <BudgetProgressBar :budget="b" />
    </div>
  </div>
</template>

<style scoped>
.widget-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}
.widget-head h2 {
  font-size: 16px;
  margin: 0;
}
.link-btn {
  border: none;
  background: transparent;
  color: var(--green);
  font: inherit;
  font-weight: 600;
  cursor: pointer;
  padding: 0;
}
.widget-item {
  margin-bottom: 14px;
}
.widget-item:last-child {
  margin-bottom: 0;
}
.widget-item-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}
.widget-label {
  font-weight: 600;
  font-size: 14px;
}
</style>
