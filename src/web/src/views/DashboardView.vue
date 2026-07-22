<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useOverviewStore } from '../stores/overview'
import { useInvalidation } from '../composables/useInvalidation'
import { initials } from '../utils/avatar'
import NetWorthCard from '../components/NetWorthCard.vue'
import IncomeExpenseCards from '../components/IncomeExpenseCards.vue'
import CategorySpendingList from '../components/CategorySpendingList.vue'
import OverviewBudgetWidget from '../components/OverviewBudgetWidget.vue'
import RecentTransactionsList from '../components/RecentTransactionsList.vue'

// Màn Tổng quan — trang chủ (`/`), bố cục theo dashboard.png (feature 004):
// lời chào → Tổng tài sản ròng → Thu/Chi → Chi tiêu theo danh mục → Ngân sách → Giao dịch gần đây.
const auth = useAuthStore()
const router = useRouter()
const overview = useOverviewStore()

onMounted(() => overview.fetch())
// Realtime (D33): mọi phần suy ra từ giao dịch/tài khoản/danh mục/ngân sách.
useInvalidation('transactions_changed', () => overview.fetch())
useInvalidation('accounts_changed', () => overview.fetch())
useInvalidation('budgets_changed', () => overview.fetch())
useInvalidation('categories_changed', () => overview.fetch())

function greeting() {
  const h = new Date().getHours()
  if (h < 11) return 'Chào buổi sáng'
  if (h < 18) return 'Chào buổi chiều'
  return 'Chào buổi tối'
}
async function logout() {
  await auth.logout()
  router.replace('/login')
}
</script>

<template>
  <div>
    <div class="dash-head">
      <div class="dash-greet">
        <p class="muted" data-testid="greeting">{{ greeting() }},</p>
        <h1 class="user-name" data-testid="current-user">{{ auth.user?.display_name }}</h1>
        <p class="muted" data-testid="household-name">{{ auth.household?.name }}</p>
      </div>
      <div class="dash-actions">
        <button class="btn btn-ghost" data-testid="manage-categories" @click="router.push('/categories')">Danh mục</button>
        <button class="btn btn-ghost" data-testid="logout" @click="logout">Đăng xuất</button>
        <button
          class="avatar"
          type="button"
          data-testid="avatar"
          :title="auth.user?.display_name"
          aria-label="Trang cá nhân"
          @click="router.push('/profile')"
        >
          {{ initials(auth.user?.display_name) }}
        </button>
      </div>
    </div>

    <template v-if="overview.summary">
      <NetWorthCard
        :net-worth="overview.summary.net_worth"
        :change-percent="overview.summary.net_worth_change_percent"
      />
      <IncomeExpenseCards :month="overview.summary.month" />
      <CategorySpendingList :items="overview.summary.category_spending" />
      <OverviewBudgetWidget />
      <RecentTransactionsList :items="overview.summary.recent_transactions" />
    </template>
    <p v-else-if="overview.loading" class="muted">Đang tải…</p>
  </div>
</template>

<style scoped>
.dash-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 14px;
}
.dash-greet .user-name {
  margin: 2px 0;
  font-size: 22px;
}
.dash-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.btn-ghost {
  background: transparent;
  border: 1px solid var(--border);
  padding: 6px 10px;
  font-size: 13px;
}
.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--green);
  color: #fff;
  font-weight: 700;
  font-size: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: none;
  padding: 0;
  cursor: pointer;
}
.avatar:hover {
  opacity: 0.9;
}
</style>
