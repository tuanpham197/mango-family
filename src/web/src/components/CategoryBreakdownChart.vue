<script setup lang="ts">
import { computed } from 'vue'
import type { CategoryBreakdown } from '../api/types'
import DonutChart from './charts/DonutChart.vue'

// Phân bổ chi tiêu theo danh mục (BR-RPT-002): donut + chú giải (tên · số tiền · %).
// Bấm một dòng chú giải → drill-in báo cáo chi tiết danh mục.
const props = defineProps<{ items: CategoryBreakdown[] }>()
const emit = defineEmits<{ select: [categoryId: string] }>()

const PALETTE = ['#1f8a5b', '#d98a1f', '#d1453b', '#3b82c4', '#8a5bd9', '#5bc0be', '#c45b8a', '#b0b0a8']
const colors = computed(() => props.items.map((_, i) => PALETTE[i % PALETTE.length]))

function money(n: number) {
  return new Intl.NumberFormat('vi-VN').format(Math.round(n))
}
</script>

<template>
  <div class="card" data-testid="category-breakdown">
    <h2>Chi tiêu theo danh mục</h2>
    <p v-if="items.length === 0" class="muted" data-testid="breakdown-empty">Chưa có chi tiêu trong khoảng này.</p>
    <template v-else>
      <DonutChart :labels="items.map((i) => i.category_name)" :values="items.map((i) => i.amount)" :colors="colors" />
      <div class="legend">
        <button
          v-for="(c, i) in items"
          :key="c.category_id"
          type="button"
          class="legend-row"
          :data-testid="`breakdown-row-${c.category_name}`"
          @click="emit('select', c.category_id)"
        >
          <span class="dot" :style="{ background: colors[i] }"></span>
          <span class="legend-name">{{ c.category_name }}<span v-if="c.category_hidden" class="muted"> (đã ẩn)</span></span>
          <span class="legend-val">{{ money(c.amount) }} đ · {{ c.percent }}%</span>
        </button>
      </div>
    </template>
  </div>
</template>

<style scoped>
.card h2 {
  font-size: 16px;
  margin: 0 0 10px;
}
.legend {
  margin-top: 10px;
}
.legend-row {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px 4px;
  border: none;
  border-bottom: 1px solid var(--border);
  background: transparent;
  font: inherit;
  cursor: pointer;
  text-align: left;
}
.legend-row:last-child {
  border-bottom: none;
}
.dot {
  width: 12px;
  height: 12px;
  border-radius: 3px;
  flex-shrink: 0;
}
.legend-name {
  flex: 1;
}
.legend-val {
  font-weight: 600;
  white-space: nowrap;
}
</style>
