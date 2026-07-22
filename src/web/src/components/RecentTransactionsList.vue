<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { Transaction } from '../api/types'

// Giao dịch gần đây trên màn Tổng quan (dashboard.png — cuối màn): xem nhanh, read-only.
// "Xem tất cả ›" mở màn Giao dịch (sổ đầy đủ) tại /ledger.
defineProps<{ items: Transaction[] }>()

const router = useRouter()

function fmt(amount: number, type: string) {
  const n = new Intl.NumberFormat('vi-VN').format(amount)
  return type === 'INCOME' ? `+${n}` : `−${n}`
}
</script>

<template>
  <div class="card" data-testid="recent-transactions">
    <div class="rt-head">
      <h2>Giao dịch gần đây</h2>
      <button type="button" class="link-btn" data-testid="recent-see-all" @click="router.push('/ledger')">
        Xem tất cả ›
      </button>
    </div>
    <p v-if="items.length === 0" class="muted" data-testid="recent-empty">Chưa có giao dịch nào.</p>
    <div v-for="t in items" :key="t.id" class="rt-row" data-testid="recent-row">
      <span class="row-icon">{{ t.type === 'INCOME' ? '💰' : '💸' }}</span>
      <div class="rt-main">
        <div class="rt-title">{{ t.description || t.category_name }}</div>
        <div class="rt-sub muted">{{ t.category_name }} · {{ t.account_name }}</div>
      </div>
      <span :class="t.type === 'INCOME' ? 'amount-income' : 'amount-expense'">{{ fmt(t.amount, t.type) }}</span>
    </div>
  </div>
</template>

<style scoped>
.rt-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}
.rt-head h2 {
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
.rt-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
  border-bottom: 1px solid var(--border);
}
.rt-row:last-child {
  border-bottom: none;
}
.rt-main {
  flex: 1;
  min-width: 0;
}
.rt-title {
  font-size: 15px;
}
.rt-sub {
  font-size: 13px;
}
</style>
