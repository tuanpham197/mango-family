<script setup lang="ts">
import { ref } from 'vue'
import { ApiError } from '../api/client'
import type { Budget } from '../api/types'
import { useBudgetsStore } from '../stores/budgets'

// Xóa ngân sách LUÔN qua xác nhận (UC-BGT-06 AC-1/2). Hủy không đổi gì.
// RECORD_GONE → thông báo nhẹ + làm tươi danh sách (D25).
const props = defineProps<{ budget: Budget }>()
const emit = defineEmits<{ close: []; done: [message?: string] }>()

const budgets = useBudgetsStore()
const busy = ref(false)
const error = ref('')

const label = props.budget.type === 'TOTAL' ? 'Tổng chi tiêu' : (props.budget.category_name ?? 'Ngân sách')

async function confirm() {
  busy.value = true
  error.value = ''
  try {
    await budgets.remove(props.budget.id, props.budget.updated_at)
    emit('done')
  } catch (e) {
    if (e instanceof ApiError && (e.code === 'RECORD_GONE' || e.status === 404)) {
      await budgets.fetch()
      emit('done', 'Ngân sách đã được xóa trước đó.')
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
  <div class="dialog-backdrop" data-testid="delete-budget-dialog" @click.self="emit('close')">
    <div class="dialog">
      <h2>Xóa ngân sách?</h2>
      <p class="muted">{{ label }} · {{ new Intl.NumberFormat('vi-VN').format(budget.limit_amount) }} ₫</p>
      <p class="muted">Thao tác này xóa vĩnh viễn ngân sách và cảnh báo liên quan.</p>
      <p v-if="error" class="form-error" data-testid="delete-error">{{ error }}</p>
      <div class="dialog-actions">
        <button type="button" class="btn" data-testid="cancel-delete-budget" @click="emit('close')">Hủy</button>
        <button type="button" class="btn btn-danger" :disabled="busy" data-testid="confirm-delete-budget" @click="confirm">
          Xóa vĩnh viễn
        </button>
      </div>
    </div>
  </div>
</template>
