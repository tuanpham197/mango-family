<script setup lang="ts">
import { onMounted, ref } from 'vue'

// Bộ chọn khoảng thời gian (FR-001): preset tuần/tháng/quý/năm + tùy chỉnh (end ≥ start).
// Phát sự kiện `change(from, to)` (YYYY-MM-DD) — FE quy preset về [from,to] (D35).
const emit = defineEmits<{ change: [from: string, to: string] }>()

type Preset = 'week' | 'month' | 'quarter' | 'year' | 'custom'
const preset = ref<Preset>('month')
const customFrom = ref('')
const customTo = ref('')
const error = ref('')

function ymd(d: Date): string {
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${d.getFullYear()}-${m}-${day}`
}

function rangeFor(p: Exclude<Preset, 'custom'>): [string, string] {
  const now = new Date()
  const y = now.getFullYear()
  const m = now.getMonth()
  if (p === 'week') {
    const off = (now.getDay() + 6) % 7 // Thứ Hai = 0
    const mon = new Date(now)
    mon.setDate(now.getDate() - off)
    const sun = new Date(mon)
    sun.setDate(mon.getDate() + 6)
    return [ymd(mon), ymd(sun)]
  }
  if (p === 'quarter') {
    const q = Math.floor(m / 3)
    return [ymd(new Date(y, q * 3, 1)), ymd(new Date(y, q * 3 + 3, 0))]
  }
  if (p === 'year') return [`${y}-01-01`, `${y}-12-31`]
  return [ymd(new Date(y, m, 1)), ymd(new Date(y, m + 1, 0))] // month
}

function apply(p: Preset) {
  preset.value = p
  error.value = ''
  if (p === 'custom') {
    applyCustom()
    return
  }
  const [f, t] = rangeFor(p)
  emit('change', f, t)
}

function applyCustom() {
  if (!customFrom.value || !customTo.value) return
  if (customTo.value < customFrom.value) {
    error.value = 'Ngày kết thúc không được trước ngày bắt đầu'
    return
  }
  error.value = ''
  emit('change', customFrom.value, customTo.value)
}

const presets: { key: Preset; label: string }[] = [
  { key: 'week', label: 'Tuần này' },
  { key: 'month', label: 'Tháng này' },
  { key: 'quarter', label: 'Quý này' },
  { key: 'year', label: 'Năm nay' },
  { key: 'custom', label: 'Tùy chỉnh' },
]

onMounted(() => apply('month')) // mặc định tháng này
</script>

<template>
  <div class="range-picker" data-testid="time-range-picker">
    <div class="segments range-presets">
      <button
        v-for="p in presets"
        :key="p.key"
        type="button"
        :class="{ 'active-neutral': preset === p.key }"
        :data-testid="`range-${p.key}`"
        @click="apply(p.key)"
      >
        {{ p.label }}
      </button>
    </div>
    <div v-if="preset === 'custom'" class="custom-range">
      <input v-model="customFrom" type="date" data-testid="custom-from" @change="applyCustom" />
      <span>→</span>
      <input v-model="customTo" type="date" data-testid="custom-to" @change="applyCustom" />
    </div>
    <p v-if="error" class="field-error" data-testid="range-error">{{ error }}</p>
  </div>
</template>

<style scoped>
.range-picker {
  margin-bottom: 14px;
}
.range-presets {
  flex-wrap: wrap;
}
.custom-range {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}
.custom-range input {
  flex: 1;
  padding: 8px 10px;
  border: 1.5px solid var(--border);
  border-radius: 10px;
  font: inherit;
}
</style>
