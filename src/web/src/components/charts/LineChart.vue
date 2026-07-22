<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { Chart as ChartType } from 'chart.js'
import { Chart } from './registry'

// Wrapper mỏng cho biểu đồ đường (chart.js). Nhiều chuỗi: {label, data, color}.
interface Series {
  label: string
  data: number[]
  color: string
}
const props = defineProps<{ labels: string[]; series: Series[] }>()

const canvas = ref<HTMLCanvasElement | null>(null)
let chart: ChartType | null = null

function datasets() {
  return props.series.map((s) => ({
    label: s.label,
    data: s.data,
    borderColor: s.color,
    backgroundColor: s.color,
    tension: 0.3,
    pointRadius: 2,
  }))
}
function config() {
  return {
    type: 'line' as const,
    data: { labels: props.labels, datasets: datasets() },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: { legend: { display: props.series.length > 1 } },
      scales: { y: { beginAtZero: true } },
    },
  }
}

onMounted(() => {
  if (canvas.value) chart = new Chart(canvas.value, config())
})
watch(
  () => [props.labels, props.series],
  () => {
    if (!chart) return
    chart.data.labels = props.labels
    chart.data.datasets = datasets()
    chart.update()
  },
  { deep: true },
)
onBeforeUnmount(() => {
  chart?.destroy()
  chart = null
})
</script>

<template>
  <div class="chart-box" data-testid="line-chart"><canvas ref="canvas"></canvas></div>
</template>

<style scoped>
.chart-box {
  position: relative;
  height: 220px;
}
</style>
