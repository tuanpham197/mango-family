<script setup lang="ts">
import { onUnmounted, ref, watch } from 'vue'
import { api } from '../api/client'
import type { CategoryType, Suggestion } from '../api/types'

// Gợi ý danh mục theo mô tả (FR-015, D7): rule-based + học lịch sử chung hộ,
// LUÔN chỉ tham khảo — bấm chip để điền, chọn tay danh mục khác thì thôi.
const props = defineProps<{
  type: CategoryType
  description: string
  selectedId: string | null
}>()
const emit = defineEmits<{ apply: [categoryId: string] }>()

const suggestion = ref<Suggestion | null>(null)
let timer: ReturnType<typeof setTimeout> | null = null

watch(
  () => [props.description, props.type] as const,
  ([description, type]) => {
    if (timer) clearTimeout(timer)
    if (!description.trim()) {
      suggestion.value = null
      return
    }
    timer = setTimeout(async () => {
      try {
        suggestion.value = await api<Suggestion | null>(
          `/api/categories/suggest?description=${encodeURIComponent(description)}&type=${type}`,
        )
      } catch {
        suggestion.value = null
      }
    }, 350)
  },
)

onUnmounted(() => {
  if (timer) clearTimeout(timer)
})
</script>

<template>
  <button
    v-if="suggestion && suggestion.category_id !== selectedId"
    type="button"
    class="chip"
    data-testid="suggestion-chip"
    @click="suggestion && emit('apply', suggestion.category_id)"
  >
    💡 Gợi ý: {{ suggestion.name }}
  </button>
</template>
