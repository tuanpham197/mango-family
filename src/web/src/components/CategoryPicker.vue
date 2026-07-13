<script setup lang="ts">
import { computed, ref } from 'vue'
import { ApiError } from '../api/client'
import type { CategoryType } from '../api/types'
import { useCategoriesStore } from '../stores/categories'

// Chọn danh mục cho giao dịch: lọc đúng loại, nhóm cha → con, loại trừ hidden
// (FR-014/019/020). Gán được cả danh mục cha lẫn con. Kèm tạo nhanh danh mục
// kế thừa loại đang chọn và chọn ngay cho giao dịch (UC-CAT-07, quickstart #4).
const props = defineProps<{
  type: CategoryType
  modelValue: string | null
}>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const categories = useCategoriesStore()
const groups = computed(() => categories.pickable(props.type))

const quickOpen = ref(false)
const quickName = ref('')
const quickError = ref('')
const quickDupPending = ref(false)
const quickBusy = ref(false)

async function quickCreate() {
  if (!quickName.value.trim()) {
    quickError.value = 'Nhập tên danh mục'
    return
  }
  quickBusy.value = true
  quickError.value = ''
  try {
    const cat = await categories.create({
      name: quickName.value.trim(),
      type: props.type,
      confirm_duplicate: quickDupPending.value,
    })
    emit('update:modelValue', cat.id)
    quickOpen.value = false
    quickName.value = ''
    quickDupPending.value = false
  } catch (e) {
    if (e instanceof ApiError && e.code === 'NAME_DUPLICATE_WARNING') {
      quickDupPending.value = true
      quickError.value = `${e.message} — bấm Tạo lần nữa để xác nhận.`
    } else if (e instanceof ApiError) {
      quickError.value = e.message
    } else {
      quickError.value = 'Không kết nối được máy chủ.'
    }
  } finally {
    quickBusy.value = false
  }
}

function iconOf(icon: string | null) {
  return icon ?? '🏷️'
}
</script>

<template>
  <div class="picker" data-testid="category-picker">
    <p v-if="groups.length === 0" class="muted">Chưa có danh mục nào cho loại này.</p>
    <div v-for="root in groups" :key="root.id" class="picker-group">
      <button
        type="button"
        class="picker-item"
        :class="{ selected: modelValue === root.id }"
        :data-testid="`category-option-${root.name}`"
        @click="emit('update:modelValue', root.id)"
      >
        <span class="row-icon">{{ iconOf(root.icon) }}</span>
        <span>{{ root.name }}</span>
      </button>
      <button
        v-for="child in root.children"
        :key="child.id"
        type="button"
        class="picker-item picker-child"
        :class="{ selected: modelValue === child.id }"
        :data-testid="`category-option-${child.name}`"
        @click="emit('update:modelValue', child.id)"
      >
        <span class="row-icon">{{ iconOf(child.icon) }}</span>
        <span>{{ child.name }}</span>
      </button>
    </div>

    <div class="quick-create">
      <button v-if="!quickOpen" type="button" class="btn btn-outline btn-block" data-testid="quick-create-toggle" @click="quickOpen = true">
        ＋ Tạo danh mục mới
      </button>
      <template v-else>
        <div class="quick-row">
          <input
            v-model="quickName"
            type="text"
            placeholder="Tên danh mục mới"
            data-testid="quick-create-name"
            @keydown.enter.prevent="quickCreate"
          />
          <button type="button" class="btn btn-primary" :disabled="quickBusy" data-testid="quick-create-submit" @click="quickCreate">
            Tạo
          </button>
        </div>
        <p v-if="quickError" class="field-error" data-testid="quick-create-error">{{ quickError }}</p>
      </template>
    </div>
  </div>
</template>

<style scoped>
.picker {
  max-height: 260px;
  overflow-y: auto;
  border: 1.5px solid var(--border);
  border-radius: 12px;
  background: var(--card);
  padding: 6px;
}

.picker-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 8px 10px;
  border: none;
  background: transparent;
  border-radius: 10px;
  font: inherit;
  font-size: 15px;
  cursor: pointer;
  text-align: left;
}

.picker-item.selected {
  background: #eef7f2;
  outline: 1.5px solid var(--green);
}

.picker-child {
  padding-left: 34px;
  font-size: 14px;
}

.picker-child .row-icon {
  width: 28px;
  height: 28px;
  font-size: 13px;
}

.quick-create {
  padding: 6px 4px 4px;
}

.quick-row {
  display: flex;
  gap: 8px;
}

.quick-row input {
  flex: 1;
  padding: 8px 10px;
  border: 1.5px solid var(--border);
  border-radius: 10px;
  font: inherit;
}
</style>
