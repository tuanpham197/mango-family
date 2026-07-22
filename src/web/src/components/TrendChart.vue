<script setup lang="ts">
import { computed } from 'vue'
import type { TrendPoint } from '../api/types'
import LineChart from './charts/LineChart.vue'

// Xu hướng thu/chi theo thời gian (BR-RPT-003): đường Thu (xanh) + Chi (đỏ).
const props = defineProps<{ trend: TrendPoint[] }>()

const labels = computed(() => props.trend.map((t) => t.bucket))
const series = computed(() => [
  { label: 'Thu', data: props.trend.map((t) => t.income), color: '#1f8a5b' },
  { label: 'Chi', data: props.trend.map((t) => t.expense), color: '#d1453b' },
])
</script>

<template>
  <div class="card" data-testid="trend-chart-block">
    <h2>Xu hướng thu / chi</h2>
    <p v-if="trend.length === 0" class="muted">Chưa có dữ liệu.</p>
    <LineChart v-else :labels="labels" :series="series" />
  </div>
</template>

<style scoped>
.card h2 {
  font-size: 16px;
  margin: 0 0 10px;
}
</style>
