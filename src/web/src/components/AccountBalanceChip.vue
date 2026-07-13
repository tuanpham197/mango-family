<script setup lang="ts">
import { onMounted } from 'vue'
import { useAccountsStore } from '../stores/accounts'
import { useInvalidation } from '../composables/useInvalidation'

// Hiển thị số dư từng tài khoản (số dư suy ra từ view — D14). Tự làm mới khi
// server phát accounts_changed (sau mỗi thêm/sửa/xóa giao dịch).
const accounts = useAccountsStore()

// Luôn lấy số dư mới mỗi khi sổ hiển thị (không dùng ensure — số dư đổi sau mọi
// mutation); ngoài ra refetch realtime khi server phát accounts_changed.
onMounted(() => accounts.fetch())
useInvalidation('accounts_changed', () => accounts.fetch())

function fmt(n: number) {
  return new Intl.NumberFormat('vi-VN').format(n) + ' ₫'
}
</script>

<template>
  <div class="balances" data-testid="account-balances">
    <div v-for="a in accounts.items" :key="a.id" class="balance-chip" data-testid="account-chip">
      <span class="balance-name">{{ a.name }}</span>
      <span class="balance-amount" :class="{ negative: a.balance < 0 }" :data-testid="`balance-${a.name}`">
        {{ fmt(a.balance) }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.balances {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding-bottom: 4px;
  margin-bottom: 12px;
}
.balance-chip {
  flex-shrink: 0;
  border: 1.5px solid var(--border);
  border-radius: 12px;
  padding: 8px 12px;
  background: var(--card);
  min-width: 120px;
}
.balance-name {
  display: block;
  font-size: 12px;
  color: var(--muted);
}
.balance-amount {
  font-size: 16px;
  color: var(--green);
}
.balance-amount.negative {
  color: var(--red);
}
</style>
