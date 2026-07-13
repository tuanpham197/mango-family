<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useTransactionsStore } from '../stores/transactions'
import { useInvalidation } from '../composables/useInvalidation'

const auth = useAuthStore()
const router = useRouter()
const transactions = useTransactionsStore()

onMounted(() => transactions.fetch())
// Sổ chung của hộ: giao dịch thành viên khác nhập hiện ra ≤ 5s (FR-022, D8).
useInvalidation('transactions_changed', () => transactions.fetch())

function formatAmount(amount: number, type: string) {
  const n = new Intl.NumberFormat('vi-VN').format(amount)
  return type === 'INCOME' ? `+${n}` : `−${n}`
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString('vi-VN', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function logout() {
  await auth.logout()
  router.push('/login')
}
</script>

<template>
  <div>
    <div class="page-head">
      <div>
        <h1 data-testid="household-name">{{ auth.household?.name }}</h1>
        <p class="muted" data-testid="current-user">{{ auth.user?.display_name }}</p>
      </div>
      <button class="btn" data-testid="logout" @click="logout">Đăng xuất</button>
    </div>

    <div class="card" data-testid="ledger">
      <p v-if="transactions.items.length === 0" class="muted">
        Chưa có giao dịch nào — bấm ＋ để nhập giao dịch đầu tiên.
      </p>
      <div v-for="t in transactions.items" :key="t.id" class="list-row" data-testid="transaction-row">
        <span class="row-icon">{{ t.type === 'INCOME' ? '💰' : '💸' }}</span>
        <div class="row-main">
          <div class="row-title">{{ t.description || t.category_name }}</div>
          <div class="row-sub">
            {{ t.category_name }} · do {{ t.created_by_name }} nhập · {{ formatDate(t.transaction_date) }}
          </div>
        </div>
        <span :class="t.type === 'INCOME' ? 'amount-income' : 'amount-expense'">
          {{ formatAmount(t.amount, t.type) }}
        </span>
      </div>
    </div>
  </div>
</template>
