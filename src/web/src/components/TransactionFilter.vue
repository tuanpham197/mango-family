<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

// Filter tháng/năm cho màn Giao dịch. Phát `change(from, to)` (YYYY-MM-DD, inclusive).
// "Cả năm" → toàn bộ năm; mặc định tháng/năm hiện tại.
const emit = defineEmits<{ change: [from: string, to: string] }>()

const now = new Date()
const year = ref(now.getFullYear())
const month = ref(now.getMonth() + 1) // 1–12; 0 = Cả năm

const years = computed(() => Array.from({ length: 5 }, (_, i) => now.getFullYear() - i))

function ymd(d: Date): string {
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${d.getFullYear()}-${m}-${day}`
}

function emitRange() {
  if (month.value === 0) {
    emit('change', `${year.value}-01-01`, `${year.value}-12-31`)
    return
  }
  const from = new Date(year.value, month.value - 1, 1)
  const to = new Date(year.value, month.value, 0) // ngày cuối tháng
  emit('change', ymd(from), ymd(to))
}

onMounted(emitRange) // phát khoảng mặc định (tháng này) khi mở màn
</script>

<template>
  <div class="txn-filter" data-testid="txn-filter">
    <select v-model.number="month" class="filter-select" data-testid="filter-month" @change="emitRange">
      <option :value="0">Cả năm</option>
      <option v-for="m in 12" :key="m" :value="m">Tháng {{ m }}</option>
    </select>
    <select v-model.number="year" class="filter-select" data-testid="filter-year" @change="emitRange">
      <option v-for="y in years" :key="y" :value="y">Năm {{ y }}</option>
    </select>
  </div>
</template>

<style scoped>
.txn-filter {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.filter-select {
  flex: 1;
  padding: 8px 10px;
  border: 1.5px solid var(--border);
  border-radius: 10px;
  background: var(--card);
  font: inherit;
}
</style>
