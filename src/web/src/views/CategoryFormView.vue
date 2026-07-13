<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError } from '../api/client'
import type { Category, CategoryType } from '../api/types'
import { useCategoriesStore } from '../stores/categories'

// Tạo/sửa danh mục (UC-CAT-02/04): loại bắt buộc khi tạo, BẤT BIẾN khi sửa
// (FR-004/005); chọn cha → khóa loại theo cha (FR-010/011); cảnh báo trùng tên
// cho xác nhận tiếp tục (FR-017); conflict → thông báo tải lại (D6).
const route = useRoute()
const router = useRouter()
const categories = useCategoriesStore()

const editId = computed(() => (route.params.id as string | undefined) ?? null)
const isEdit = computed(() => editId.value !== null)

const name = ref('')
const icon = ref('')
const type = ref<CategoryType | ''>('')
const parentId = ref('')
const expectedUpdatedAt = ref('')
const fieldErrors = ref<Record<string, string>>({})
const formError = ref('')
const duplicatePending = ref(false)
const busy = ref(false)

const original = ref<Category | null>(null)

onMounted(async () => {
  await categories.ensure()
  // Tạo nhanh từ luồng nhập giao dịch: ?type=EXPENSE giữ đúng loại (UC-CAT-07).
  const queryType = route.query.type
  if (queryType === 'INCOME' || queryType === 'EXPENSE') type.value = queryType
  if (editId.value) {
    const cat = categories.findById(editId.value)
    if (!cat) {
      formError.value = 'Không tìm thấy danh mục.'
      return
    }
    original.value = cat
    name.value = cat.name
    icon.value = cat.icon ?? ''
    type.value = cat.type
    parentId.value = cat.parent_id ?? ''
    expectedUpdatedAt.value = cat.updated_at
  }
})

// Cha chỉ có thể là danh mục gốc (FR-010); tạo con → khóa loại theo cha.
const parentOptions = computed(() =>
  type.value === '' ? [] : categories.byType(type.value as CategoryType).filter((t) => !t.parent_id),
)

const effectiveType = computed<CategoryType | ''>(() => {
  if (parentId.value) {
    const parent = categories.findById(parentId.value)
    if (parent) return parent.type
  }
  return type.value
})

async function submit() {
  fieldErrors.value = {}
  formError.value = ''
  if (!name.value.trim()) {
    fieldErrors.value.name = 'Tên danh mục là bắt buộc'
    return
  }
  if (!isEdit.value && !effectiveType.value) {
    // Loại là bắt buộc (FR-004, quickstart #2).
    fieldErrors.value.type = 'Vui lòng chọn loại Thu hoặc Chi'
    return
  }
  busy.value = true
  try {
    if (isEdit.value) {
      await categories.update(editId.value!, {
        name: name.value.trim(),
        icon: icon.value.trim(),
        expected_updated_at: expectedUpdatedAt.value,
      })
    } else {
      await categories.create({
        name: name.value.trim(),
        type: effectiveType.value as CategoryType,
        icon: icon.value.trim() || undefined,
        parent_id: parentId.value || undefined,
        confirm_duplicate: duplicatePending.value,
      })
    }
    router.push('/categories')
  } catch (e) {
    if (e instanceof ApiError && e.code === 'NAME_DUPLICATE_WARNING') {
      duplicatePending.value = true
      formError.value = e.message
    } else if (e instanceof ApiError && (e.code === 'CONCURRENCY_CONFLICT' || e.code === 'RECORD_GONE')) {
      formError.value = e.message
      await categories.fetch()
      const fresh = editId.value ? categories.findById(editId.value) : null
      if (fresh) {
        original.value = fresh
        expectedUpdatedAt.value = fresh.updated_at
      }
    } else if (e instanceof ApiError && e.field) {
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
      <h1>{{ isEdit ? 'Sửa danh mục' : 'Danh mục mới' }}</h1>
    </div>

    <form @submit.prevent="submit">
      <div v-if="formError" class="form-error" data-testid="form-error">
        {{ formError }}
        <template v-if="duplicatePending"> — bấm Lưu lần nữa để vẫn tạo.</template>
      </div>

      <div class="field">
        <label>Loại</label>
        <div class="segments" data-testid="type-toggle">
          <button
            type="button"
            :disabled="isEdit || !!parentId"
            :class="{ 'active-expense': effectiveType === 'EXPENSE' }"
            data-testid="type-expense"
            @click="type = 'EXPENSE'"
          >
            Chi phí
          </button>
          <button
            type="button"
            :disabled="isEdit || !!parentId"
            :class="{ 'active-income': effectiveType === 'INCOME' }"
            data-testid="type-income"
            @click="type = 'INCOME'"
          >
            Thu nhập
          </button>
        </div>
        <p v-if="isEdit" class="muted">Loại không thể thay đổi sau khi tạo (FR-005).</p>
        <p v-if="fieldErrors.type" class="field-error" data-testid="error-type">{{ fieldErrors.type }}</p>
      </div>

      <div class="field">
        <label for="name">Tên danh mục</label>
        <input id="name" v-model="name" type="text" data-testid="category-name" />
        <p v-if="fieldErrors.name" class="field-error" data-testid="error-name">{{ fieldErrors.name }}</p>
      </div>

      <div class="field">
        <label for="icon">Biểu tượng (emoji)</label>
        <input id="icon" v-model="icon" type="text" maxlength="4" data-testid="category-icon" />
      </div>

      <div v-if="!isEdit" class="field">
        <label for="parent">Danh mục cha (tùy chọn — con kế thừa loại cha)</label>
        <select id="parent" v-model="parentId" data-testid="category-parent" :disabled="!type && !parentId">
          <option value="">— Không (danh mục gốc) —</option>
          <option v-for="p in parentOptions" :key="p.id" :value="p.id">{{ p.icon ?? '🏷️' }} {{ p.name }}</option>
        </select>
        <p v-if="fieldErrors.parent_id" class="field-error">{{ fieldErrors.parent_id }}</p>
      </div>

      <button class="btn btn-primary btn-block" type="submit" :disabled="busy" data-testid="save-category">
        {{ busy ? 'Đang lưu…' : 'Lưu' }}
      </button>
    </form>
  </div>
</template>
