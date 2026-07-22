<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError } from '../api/client'
import type { Budget, BudgetType, PeriodType } from '../api/types'
import CategoryPicker from '../components/CategoryPicker.vue'
import MoneyInput from '../components/MoneyInput.vue'
import { useBudgetsStore } from '../stores/budgets'
import { useCategoriesStore } from '../stores/categories'
import { useInvalidation } from '../composables/useInvalidation'

// Form tạo/sửa ngân sách (UC-BGT-01/02/05). Chế độ SỬA khi có :id (loại bất biến).
const route = useRoute()
const router = useRouter()
const budgets = useBudgetsStore()
const categories = useCategoriesStore()

const editId = computed(() => (route.params.id as string | undefined) ?? null)
const isEdit = computed(() => editId.value !== null)

const type = ref<BudgetType>('CATEGORY')
const categoryId = ref<string | null>(null)
const limit = ref<number | null>(null)
const periodType = ref<PeriodType>('MONTHLY')
const startDate = ref('')
const endDate = ref('')
const expectedUpdatedAt = ref('')

const fieldErrors = ref<Record<string, string>>({})
const formError = ref('')
const duplicateId = ref<string | null>(null)
const busy = ref(false)

function applyBudget(b: Budget) {
  type.value = b.type
  categoryId.value = b.category_id
  limit.value = b.limit_amount
  periodType.value = b.period_type
  startDate.value = b.start_date ? b.start_date.slice(0, 10) : ''
  endDate.value = b.end_date ? b.end_date.slice(0, 10) : ''
  expectedUpdatedAt.value = b.updated_at
}

onMounted(async () => {
  await categories.ensure()
  if (editId.value) {
    try {
      applyBudget(await budgets.getById(editId.value))
    } catch (e) {
      if (e instanceof ApiError && (e.status === 404 || e.code === 'RECORD_GONE')) {
        formError.value = 'Ngân sách không còn tồn tại.'
      } else {
        formError.value = 'Không tải được ngân sách.'
      }
    }
  }
})
// Danh mục thành viên khác tạo hiện ra trong picker (đồng bộ ≤ 5s).
useInvalidation('categories_changed', () => categories.fetch())

// Ngày một lần gửi dạng noon-UTC để không lệch ngày qua múi giờ (như 002).
function toApiDate(d: string): string {
  return d + 'T12:00:00Z'
}

async function submit() {
  fieldErrors.value = {}
  formError.value = ''
  duplicateId.value = null

  if (!limit.value || limit.value <= 0) {
    fieldErrors.value.limit_amount = 'Giới hạn phải lớn hơn 0'
    return
  }
  if (type.value === 'CATEGORY' && !categoryId.value) {
    fieldErrors.value.category_id = 'Vui lòng chọn danh mục Chi'
    return
  }
  if (periodType.value === 'ONE_TIME') {
    if (!startDate.value || !endDate.value) {
      fieldErrors.value.end_date = 'Kỳ một lần cần ngày bắt đầu và kết thúc'
      return
    }
    if (endDate.value < startDate.value) {
      fieldErrors.value.end_date = 'Ngày kết thúc không được trước ngày bắt đầu'
      return
    }
  }

  const payload = {
    type: type.value,
    category_id: type.value === 'CATEGORY' ? categoryId.value : null,
    limit_amount: limit.value,
    period_type: periodType.value,
    start_date: periodType.value === 'ONE_TIME' ? toApiDate(startDate.value) : null,
    end_date: periodType.value === 'ONE_TIME' ? toApiDate(endDate.value) : null,
  }

  busy.value = true
  try {
    if (isEdit.value) {
      await budgets.update(editId.value!, { ...payload, expected_updated_at: expectedUpdatedAt.value })
    } else {
      await budgets.create(payload)
    }
    router.push('/budgets')
  } catch (e) {
    handleError(e)
  } finally {
    busy.value = false
  }
}

function handleError(e: unknown) {
  if (!(e instanceof ApiError)) {
    formError.value = 'Không kết nối được máy chủ, vui lòng thử lại.'
    return
  }
  if (e.code === 'BUDGET_DUPLICATE') {
    duplicateId.value = (e.body.existing_budget_id as string) ?? null
    formError.value = 'Đã có ngân sách đang hoạt động cho lựa chọn này.'
  } else if (e.code === 'CONCURRENCY_CONFLICT') {
    const fresh = e.body.data as Budget | undefined
    if (fresh) applyBudget(fresh)
    formError.value = 'Ngân sách vừa được thành viên khác sửa — đã tải bản mới nhất, kiểm tra rồi lưu lại.'
  } else if (e.code === 'RECORD_GONE' || e.status === 404) {
    formError.value = 'Ngân sách đã bị xóa bởi thành viên khác.'
    setTimeout(() => router.push('/budgets'), 1200)
  } else if (e.field) {
    fieldErrors.value[e.field] = e.message
  } else {
    formError.value = e.message
  }
}

function goExisting() {
  if (duplicateId.value) router.push(`/budgets/${duplicateId.value}/edit`)
}
</script>

<template>
  <div>
    <div class="page-head">
      <h1>{{ isEdit ? 'Sửa ngân sách' : 'Ngân sách mới' }}</h1>
    </div>

    <form @submit.prevent="submit">
      <div v-if="formError" class="form-error" data-testid="form-error">
        {{ formError }}
        <button v-if="duplicateId" type="button" class="link-btn" data-testid="goto-existing-budget" @click="goExisting">
          Tới ngân sách hiện có ›
        </button>
      </div>

      <!-- Loại ngân sách: bất biến khi sửa (UC-BGT-05) -->
      <div class="segments" data-testid="budget-type-toggle">
        <button
          type="button"
          :class="{ 'active-neutral': type === 'CATEGORY' }"
          :disabled="isEdit"
          data-testid="type-category"
          @click="type = 'CATEGORY'"
        >
          Theo danh mục
        </button>
        <button
          type="button"
          :class="{ 'active-neutral': type === 'TOTAL' }"
          :disabled="isEdit"
          data-testid="type-total"
          @click="type = 'TOTAL'; categoryId = null"
        >
          Tổng chi tiêu
        </button>
      </div>

      <div v-if="type === 'CATEGORY'" class="field">
        <label>Danh mục Chi</label>
        <CategoryPicker v-model="categoryId" type="EXPENSE" />
        <p v-if="fieldErrors.category_id" class="field-error" data-testid="error-category">{{ fieldErrors.category_id }}</p>
      </div>

      <div class="field">
        <label for="limit">Giới hạn (₫)</label>
        <MoneyInput id="limit" v-model="limit" data-testid="limit-input" />
        <p v-if="fieldErrors.limit_amount" class="field-error" data-testid="error-limit">{{ fieldErrors.limit_amount }}</p>
      </div>

      <div class="field">
        <label for="period">Kỳ áp dụng</label>
        <select id="period" v-model="periodType" data-testid="period-select">
          <option value="MONTHLY">Hàng tháng</option>
          <option value="WEEKLY">Hàng tuần</option>
          <option value="ONE_TIME">Một lần</option>
        </select>
      </div>

      <template v-if="periodType === 'ONE_TIME'">
        <div class="field">
          <label for="start">Từ ngày</label>
          <input id="start" v-model="startDate" type="date" data-testid="start-date-input" />
        </div>
        <div class="field">
          <label for="end">Đến ngày</label>
          <input id="end" v-model="endDate" type="date" data-testid="end-date-input" />
          <p v-if="fieldErrors.end_date" class="field-error" data-testid="error-period">{{ fieldErrors.end_date }}</p>
        </div>
      </template>

      <button class="btn btn-primary btn-block" type="submit" :disabled="busy || budgets.submitting" data-testid="save-budget">
        {{ busy ? 'Đang lưu…' : 'Lưu ngân sách' }}
      </button>
    </form>
  </div>
</template>

<style scoped>
.link-btn {
  display: inline-block;
  margin-left: 8px;
  border: none;
  background: transparent;
  color: var(--green);
  font: inherit;
  font-weight: 600;
  cursor: pointer;
  padding: 0;
}
</style>
