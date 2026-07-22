<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { Transaction } from '../api/types'
import AccountBalanceChip from '../components/AccountBalanceChip.vue'
import DeleteTransactionDialog from '../components/DeleteTransactionDialog.vue'
import TransactionFilter from '../components/TransactionFilter.vue'
import { useAuthStore } from '../stores/auth'
import { useTransactionsStore } from '../stores/transactions'
import { useInvalidation } from '../composables/useInvalidation'

const auth = useAuthStore()
const router = useRouter()
const transactions = useTransactionsStore()

const deleting = ref<Transaction | null>(null)
const notice = ref('')
const sentinel = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | null = null

onMounted(() => {
  // Lần tải đầu do TransactionFilter phát khoảng mặc định (tháng này) → setRange → fetch.
  // Infinite scroll: nạp trang kế khi chạm sentinel cuối danh sách (D16).
  observer = new IntersectionObserver((entries) => {
    if (entries[0].isIntersecting) transactions.loadMore()
  })
  if (sentinel.value) observer.observe(sentinel.value)
})
onUnmounted(() => observer?.disconnect())

// Sổ chung: giao dịch thành viên khác nhập/sửa/xóa hiện ra ≤ 5s (SC-006, D8) — giữ filter.
useInvalidation('transactions_changed', () => transactions.fetch())

// Đổi filter tháng/năm → tải lại sổ theo khoảng đã chọn (infinite scroll cũng theo khoảng).
function onFilter(from: string, to: string) {
  transactions.setRange(from, to)
}

function fmt(amount: number, type: string) {
  const n = new Intl.NumberFormat('vi-VN').format(amount)
  return type === 'INCOME' ? `+${n}` : `−${n}`
}
function fmtDate(iso: string) {
  return new Date(iso).toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit', year: 'numeric' })
}
function openEdit(t: Transaction) {
  router.push(`/transactions/${t.id}/edit`)
}
function onDeleted(message?: string) {
  deleting.value = null
  if (message) {
    notice.value = message
    setTimeout(() => (notice.value = ''), 3000)
  }
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

    <AccountBalanceChip />

    <TransactionFilter @change="onFilter" />

    <div v-if="notice" class="form-error" data-testid="notice">{{ notice }}</div>

    <div class="card" data-testid="ledger">
      <p v-if="transactions.items.length === 0 && !transactions.loading" class="muted" data-testid="empty-ledger">
        Không có giao dịch nào trong khoảng đã chọn — thử đổi tháng/năm hoặc bấm ＋ để nhập.
      </p>
      <div
        v-for="t in transactions.items"
        :key="t.id"
        class="list-row txn-row"
        data-testid="transaction-row"
        @click="openEdit(t)"
      >
        <span class="row-icon">{{ t.type === 'INCOME' ? '💰' : '💸' }}</span>
        <div class="row-main">
          <div class="row-title">{{ t.description || t.category_name }}</div>
          <div class="row-sub">
            {{ t.category_name }} · {{ t.account_name }} · do {{ t.created_by_name }} nhập · {{ fmtDate(t.transaction_date) }}
          </div>
        </div>
        <span :class="t.type === 'INCOME' ? 'amount-income' : 'amount-expense'">{{ fmt(t.amount, t.type) }}</span>
        <button type="button" class="row-del" :data-testid="`delete-txn-${t.id}`" @click.stop="deleting = t">✕</button>
      </div>
      <div ref="sentinel" class="sentinel"></div>
      <p v-if="transactions.loading" class="muted">Đang tải…</p>
    </div>

    <DeleteTransactionDialog v-if="deleting" :transaction="deleting" @close="deleting = null" @done="onDeleted" />
  </div>
</template>

<style scoped>
.txn-row {
  cursor: pointer;
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
.sentinel {
  height: 1px;
}
</style>
