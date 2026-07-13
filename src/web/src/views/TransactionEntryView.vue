<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ApiError } from '../api/client'
import type { CategoryType } from '../api/types'
import CategoryPicker from '../components/CategoryPicker.vue'
import SuggestionChip from '../components/SuggestionChip.vue'
import { useCategoriesStore } from '../stores/categories'
import { useTransactionsStore } from '../stores/transactions'
import { useInvalidation } from '../composables/useInvalidation'

const router = useRouter()
const categories = useCategoriesStore()
const transactions = useTransactionsStore()

const type = ref<CategoryType>('EXPENSE')
const amount = ref<number | null>(null)
const description = ref('')
const categoryId = ref<string | null>(null)
const fieldErrors = ref<Record<string, string>>({})
const formError = ref('')
const busy = ref(false)

onMounted(() => categories.ensure())
// Danh mục do thành viên khác tạo hiện ra ≤ 5s ngay trong picker (quickstart #13).
useInvalidation('categories_changed', () => categories.fetch())

// Đổi loại → picker lọc lại; bỏ chọn nếu danh mục cũ khác loại (FR-014).
watch(type, () => {
  if (categoryId.value && categories.findById(categoryId.value)?.type !== type.value) {
    categoryId.value = null
  }
})

async function submit() {
  fieldErrors.value = {}
  formError.value = ''
  if (!amount.value || amount.value <= 0) {
    fieldErrors.value.amount = 'Số tiền phải lớn hơn 0'
    return
  }
  if (!categoryId.value) {
    // Chặn lưu khi thiếu danh mục — FR-013 (server cũng chặn: CATEGORY_REQUIRED).
    fieldErrors.value.category_id = 'Vui lòng chọn danh mục'
    return
  }
  busy.value = true
  try {
    await transactions.create({
      amount: amount.value,
      type: type.value,
      category_id: categoryId.value,
      description: description.value.trim() || undefined,
    })
    router.push('/')
  } catch (e) {
    if (e instanceof ApiError && e.field) {
      fieldErrors.value[e.field] = e.message
    } else if (e instanceof ApiError) {
      formError.value = e.message
    } else {
      formError.value = 'Không kết nối được máy chủ, vui lòng thử lại.'
    }
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div>
    <div class="page-head">
      <h1>Giao dịch mới</h1>
    </div>

    <form @submit.prevent="submit">
      <div v-if="formError" class="form-error" data-testid="form-error">{{ formError }}</div>

      <div class="segments" data-testid="type-toggle">
        <button
          type="button"
          :class="{ 'active-expense': type === 'EXPENSE' }"
          data-testid="type-expense"
          @click="type = 'EXPENSE'"
        >
          Chi phí
        </button>
        <button
          type="button"
          :class="{ 'active-income': type === 'INCOME' }"
          data-testid="type-income"
          @click="type = 'INCOME'"
        >
          Thu nhập
        </button>
      </div>

      <div class="field">
        <label for="amount">Số tiền (₫)</label>
        <input
          id="amount"
          v-model.number="amount"
          type="number"
          inputmode="numeric"
          min="0"
          step="any"
          data-testid="amount-input"
        />
        <p v-if="fieldErrors.amount" class="field-error" data-testid="error-amount">{{ fieldErrors.amount }}</p>
      </div>

      <div class="field">
        <label>Danh mục</label>
        <CategoryPicker v-model="categoryId" :type="type" />
        <p v-if="fieldErrors.category_id" class="field-error" data-testid="error-category">
          {{ fieldErrors.category_id }}
        </p>
      </div>

      <div class="field">
        <label for="description">Ghi chú</label>
        <input id="description" v-model="description" type="text" maxlength="255" data-testid="description-input" />
        <p v-if="fieldErrors.description" class="field-error">{{ fieldErrors.description }}</p>
        <SuggestionChip
          :type="type"
          :description="description"
          :selected-id="categoryId"
          @apply="categoryId = $event"
        />
      </div>

      <button class="btn btn-primary btn-block" type="submit" :disabled="busy" data-testid="save-transaction">
        {{ busy ? 'Đang lưu…' : 'Lưu giao dịch' }}
      </button>
    </form>
  </div>
</template>
