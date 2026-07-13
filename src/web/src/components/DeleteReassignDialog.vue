<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ApiError } from '../api/client'
import type { Category } from '../api/types'
import { useCategoriesStore } from '../stores/categories'

// Xóa an toàn (FR-007/008/009/012, D12): thử xóa ngay; nếu còn giao dịch →
// server trả CATEGORY_HAS_TRANSACTIONS kèm counts → hiện lựa chọn gán lại
// (đích cùng loại) hoặc xóa giao dịch. RECORD_GONE → báo đã bị xóa trước đó.
const props = defineProps<{ category: Category }>()
const emit = defineEmits<{ close: []; done: [message?: string] }>()

const categories = useCategoriesStore()

const step = ref<'confirm' | 'choose'>('confirm')
const txnCount = ref(0)
const childrenCount = ref(0)
const mode = ref<'reassign' | 'delete_transactions'>('reassign')
const targetId = ref('')
const error = ref('')
const busy = ref(false)

// Đích gán lại: cùng loại, loại trừ chính nó + con của nó (FR-009).
const targets = computed(() => {
  const excluded = new Set([props.category.id])
  for (const t of categories.trees) {
    if (t.id === props.category.id) t.children.forEach((c) => excluded.add(c.id))
  }
  const result: Category[] = []
  for (const t of categories.byType(props.category.type)) {
    if (!excluded.has(t.id)) result.push(t)
    for (const c of t.children) if (!excluded.has(c.id)) result.push(c)
  }
  return result
})

async function attempt(body: {
  mode?: 'reassign' | 'delete_transactions'
  target_category_id?: string
  expected_updated_at: string
}) {
  busy.value = true
  error.value = ''
  try {
    await categories.remove(props.category.id, body)
    emit('done')
  } catch (e) {
    if (e instanceof ApiError && e.code === 'CATEGORY_HAS_TRANSACTIONS') {
      txnCount.value = Number(e.body.transaction_count ?? 0)
      childrenCount.value = Number(e.body.children_count ?? 0)
      step.value = 'choose'
    } else if (e instanceof ApiError && e.code === 'RECORD_GONE') {
      await categories.fetch()
      emit('done', 'Danh mục đã được thành viên khác xóa trước đó.')
    } else if (e instanceof ApiError && e.code === 'CONCURRENCY_CONFLICT') {
      await categories.fetch()
      emit('done', e.message)
    } else if (e instanceof ApiError) {
      error.value = e.message
    } else {
      error.value = 'Không kết nối được máy chủ, vui lòng thử lại.'
    }
  } finally {
    busy.value = false
  }
}

onMounted(() => categories.ensure())

function confirmDelete() {
  attempt({ expected_updated_at: props.category.updated_at })
}

function confirmChoice() {
  if (mode.value === 'reassign' && !targetId.value) {
    error.value = 'Vui lòng chọn danh mục đích để gán lại.'
    return
  }
  attempt({
    mode: mode.value,
    target_category_id: mode.value === 'reassign' ? targetId.value : undefined,
    expected_updated_at: props.category.updated_at,
  })
}
</script>

<template>
  <div class="dialog-backdrop" data-testid="delete-dialog" @click.self="emit('close')">
    <div class="dialog">
      <template v-if="step === 'confirm'">
        <h2>Xóa danh mục "{{ category.name }}"?</h2>
        <p class="muted">Nếu danh mục còn giao dịch, bạn sẽ được chọn cách xử lý ở bước tiếp theo.</p>
        <p v-if="error" class="form-error">{{ error }}</p>
        <div class="dialog-actions">
          <button type="button" class="btn" @click="emit('close')">Hủy</button>
          <button type="button" class="btn btn-danger" :disabled="busy" data-testid="confirm-delete" @click="confirmDelete">
            Xóa
          </button>
        </div>
      </template>

      <template v-else>
        <h2>Danh mục còn {{ txnCount }} giao dịch</h2>
        <p class="muted">
          "{{ category.name }}"<template v-if="childrenCount > 0"> (và {{ childrenCount }} danh mục con)</template>
          đang có {{ txnCount }} giao dịch. Chọn cách xử lý — không giao dịch nào bị bỏ mồ côi.
        </p>
        <label class="option">
          <input v-model="mode" type="radio" value="reassign" data-testid="mode-reassign" />
          Gán lại giao dịch sang danh mục khác (cùng loại)
        </label>
        <div v-if="mode === 'reassign'" class="field">
          <select v-model="targetId" data-testid="reassign-target">
            <option value="" disabled>— Chọn danh mục đích —</option>
            <option v-for="t in targets" :key="t.id" :value="t.id">
              {{ t.icon ?? '🏷️' }} {{ t.name }}
            </option>
          </select>
        </div>
        <label class="option">
          <input v-model="mode" type="radio" value="delete_transactions" data-testid="mode-delete-transactions" />
          Xóa luôn toàn bộ giao dịch của danh mục này
        </label>
        <p v-if="error" class="form-error" data-testid="delete-error">{{ error }}</p>
        <div class="dialog-actions">
          <button type="button" class="btn" @click="emit('close')">Hủy</button>
          <button type="button" class="btn btn-danger" :disabled="busy" data-testid="confirm-delete-mode" @click="confirmChoice">
            Xác nhận xóa
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 0;
  font-size: 15px;
}
</style>
