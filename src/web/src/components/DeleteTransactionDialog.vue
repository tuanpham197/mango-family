<script setup lang="ts">
import { ref } from 'vue'
import { ApiError } from '../api/client'
import type { Transaction } from '../api/types'
import { useTransactionsStore } from '../stores/transactions'

// Xóa giao dịch LUÔN qua cảnh báo/xác nhận (SC-005). Hủy không đổi gì (FR-010).
// RECORD_GONE → thông báo nhẹ + làm tươi sổ (D17).
const props = defineProps<{ transaction: Transaction }>()
const emit = defineEmits<{ close: []; done: [message?: string] }>()

const transactions = useTransactionsStore()
const busy = ref(false)
const error = ref('')

function fmt(n: number, type: string) {
  const s = new Intl.NumberFormat('vi-VN').format(n)
  return (type === 'INCOME' ? '+' : '−') + s + ' ₫'
}

async function confirm() {
  busy.value = true
  error.value = ''
  try {
    await transactions.remove(props.transaction.id, props.transaction.updated_at)
    emit('done')
  } catch (e) {
    if (e instanceof ApiError && (e.code === 'RECORD_GONE' || e.status === 404)) {
      await transactions.fetch()
      emit('done', 'Giao dịch đã được xóa trước đó.')
    } else if (e instanceof ApiError) {
      error.value = e.message
    } else {
      error.value = 'Không kết nối được máy chủ, vui lòng thử lại.'
    }
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="dialog-backdrop" data-testid="delete-transaction-dialog" @click.self="emit('close')">
    <div class="dialog">
      <h2>Xóa giao dịch?</h2>
      <p class="muted">
        {{ fmt(transaction.amount, transaction.type) }} · {{ transaction.category_name }}
        <template v-if="transaction.description"> · {{ transaction.description }}</template>
      </p>
      <p class="muted">Thao tác này xóa vĩnh viễn, không thể hoàn tác.</p>
      <p v-if="error" class="form-error" data-testid="delete-error">{{ error }}</p>
      <div class="dialog-actions">
        <button type="button" class="btn" data-testid="cancel-delete" @click="emit('close')">Hủy</button>
        <button type="button" class="btn btn-danger" :disabled="busy" data-testid="confirm-delete-transaction" @click="confirm">
          Xóa vĩnh viễn
        </button>
      </div>
    </div>
  </div>
</template>
