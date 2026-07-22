<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError } from '../api/client'
import type { CategoryType, Transaction } from '../api/types'
import CategoryPicker from '../components/CategoryPicker.vue'
import MoneyInput from '../components/MoneyInput.vue'
import SuggestionChip from '../components/SuggestionChip.vue'
import { useAccountsStore } from '../stores/accounts'
import { useCategoriesStore } from '../stores/categories'
import { useTransactionsStore } from '../stores/transactions'
import { useInvalidation } from '../composables/useInvalidation'

// Form nhập/sửa giao dịch đầy đủ (UC-TRK-02/04). Chế độ SỬA khi có :id.
const route = useRoute()
const router = useRouter()
const categories = useCategoriesStore()
const accounts = useAccountsStore()
const transactions = useTransactionsStore()

const editId = computed(() => (route.params.id as string | undefined) ?? null)
const isEdit = computed(() => editId.value !== null)

const type = ref<CategoryType>('EXPENSE')
const amount = ref<number | null>(null) // giá trị số (MoneyInput tự format hiển thị)
const description = ref('')
const categoryId = ref<string | null>(null)
const accountId = ref<string | null>(null)
const date = ref('') // YYYY-MM-DD (rỗng = hôm nay)
const expectedUpdatedAt = ref('')
const fieldErrors = ref<Record<string, string>>({})
const formError = ref('')
const busy = ref(false)

const today = new Date().toISOString().slice(0, 10)

function toDateInput(iso: string) {
  return new Date(iso).toISOString().slice(0, 10)
}

function applyTransaction(t: Transaction) {
  type.value = t.type
  amount.value = t.amount
  description.value = t.description ?? ''
  categoryId.value = t.category_id
  accountId.value = t.account_id
  date.value = toDateInput(t.transaction_date)
  expectedUpdatedAt.value = t.updated_at
}

onMounted(async () => {
  await Promise.all([categories.ensure(), accounts.ensure()])
  // Chọn sẵn tài khoản khi hộ chỉ có 1 (FR-006).
  if (!accountId.value) accountId.value = accounts.defaultId
  if (editId.value) {
    try {
      applyTransaction(await transactions.getById(editId.value))
    } catch (e) {
      if (e instanceof ApiError && (e.status === 404 || e.code === 'RECORD_GONE')) {
        formError.value = 'Giao dịch không còn tồn tại.'
      } else {
        formError.value = 'Không tải được giao dịch.'
      }
    }
  } else {
    // Tạo mới: mặc định ngày = hôm nay (người dùng vẫn đổi được).
    date.value = today
  }
})
// Danh mục thành viên khác tạo hiện ra trong picker (đồng bộ ≤ 5s).
useInvalidation('categories_changed', () => categories.fetch())

// Đổi loại → danh mục cũ khác loại coi như "chưa chọn" (D18).
watch(type, () => {
  if (categoryId.value && categories.findById(categoryId.value)?.type !== type.value) {
    categoryId.value = null
  }
})

// Mốc ngày gửi lên API (xem chú thích trong payload). Tạo mới + ngày hôm nay/rỗng
// → undefined (server now()); ngày quá khứ hoặc chế độ sửa → noon của ngày chọn.
function transactionDatePayload(): string | undefined {
  if (!isEdit.value && (!date.value || date.value === today)) return undefined
  return date.value ? new Date(date.value + 'T12:00:00').toISOString() : undefined
}

async function submit() {
  fieldErrors.value = {}
  formError.value = ''
  if (!amount.value || amount.value <= 0) {
    fieldErrors.value.amount = 'Số tiền phải lớn hơn 0'
    return
  }
  if (!categoryId.value) {
    fieldErrors.value.category_id = 'Vui lòng chọn danh mục'
    return
  }
  if (!accountId.value) {
    fieldErrors.value.account_id = 'Vui lòng chọn tài khoản'
    return
  }
  const payload = {
    amount: amount.value,
    type: type.value,
    category_id: categoryId.value,
    account_id: accountId.value,
    description: description.value.trim() || undefined,
    // Tạo mới với ngày = hôm nay (mặc định) → KHÔNG gửi transaction_date, để server
    // đóng dấu now() (thời gian thực). Nếu gửi mốc noon cố định, mọi giao dịch nhập
    // trong ngày sẽ trùng transaction_date → sai thứ tự "mới nhất trước". Chỉ ngày
    // quá khứ người dùng chọn (hoặc chế độ sửa) mới gửi mốc cụ thể.
    transaction_date: transactionDatePayload(),
  }
  busy.value = true
  try {
    if (isEdit.value) {
      await transactions.update(editId.value!, { ...payload, expected_updated_at: expectedUpdatedAt.value })
    } else {
      await transactions.create(payload)
    }
    router.push('/ledger')
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
  if (e.code === 'CONCURRENCY_CONFLICT') {
    // Hiện dữ liệu mới nhất, cho sửa tiếp (D17).
    const fresh = e.body.data as Transaction | undefined
    if (fresh) applyTransaction(fresh)
    formError.value = 'Giao dịch vừa được thành viên khác sửa — đã tải bản mới nhất, kiểm tra rồi lưu lại.'
  } else if (e.code === 'RECORD_GONE' || e.status === 404) {
    formError.value = 'Giao dịch đã bị xóa bởi thành viên khác.'
    setTimeout(() => router.push('/ledger'), 1200)
  } else if (e.field) {
    fieldErrors.value[e.field] = e.message
  } else {
    formError.value = e.message
  }
}
</script>

<template>
  <div>
    <div class="page-head">
      <h1>{{ isEdit ? 'Sửa giao dịch' : 'Giao dịch mới' }}</h1>
    </div>

    <form @submit.prevent="submit">
      <div v-if="formError" class="form-error" data-testid="form-error">{{ formError }}</div>

      <div class="segments" data-testid="type-toggle">
        <button type="button" :class="{ 'active-expense': type === 'EXPENSE' }" data-testid="type-expense" @click="type = 'EXPENSE'">
          Chi phí
        </button>
        <button type="button" :class="{ 'active-income': type === 'INCOME' }" data-testid="type-income" @click="type = 'INCOME'">
          Thu nhập
        </button>
      </div>

      <div class="field">
        <label for="amount">Số tiền (₫)</label>
        <MoneyInput id="amount" v-model="amount" data-testid="amount-input" />
        <p v-if="fieldErrors.amount" class="field-error" data-testid="error-amount">{{ fieldErrors.amount }}</p>
      </div>

      <div class="field">
        <label>Danh mục</label>
        <CategoryPicker v-model="categoryId" :type="type" />
        <p v-if="fieldErrors.category_id" class="field-error" data-testid="error-category">{{ fieldErrors.category_id }}</p>
      </div>

      <div class="field">
        <label for="account">Tài khoản</label>
        <select id="account" v-model="accountId" data-testid="account-select">
          <option :value="null" disabled>— Chọn tài khoản —</option>
          <option v-for="a in accounts.items" :key="a.id" :value="a.id">{{ a.name }}</option>
        </select>
        <p v-if="fieldErrors.account_id" class="field-error" data-testid="error-account">{{ fieldErrors.account_id }}</p>
      </div>

      <div class="field">
        <label for="date">Ngày</label>
        <input id="date" v-model="date" type="date" :max="today" data-testid="date-input" />
      </div>

      <div class="field">
        <label for="description">Ghi chú</label>
        <input id="description" v-model="description" type="text" maxlength="255" data-testid="description-input" />
        <p v-if="fieldErrors.description" class="field-error">{{ fieldErrors.description }}</p>
        <SuggestionChip :type="type" :description="description" :selected-id="categoryId" @apply="categoryId = $event" />
      </div>

      <button class="btn btn-primary btn-block" type="submit" :disabled="busy || transactions.submitting" data-testid="save-transaction">
        {{ busy ? 'Đang lưu…' : 'Lưu giao dịch' }}
      </button>
    </form>
  </div>
</template>
