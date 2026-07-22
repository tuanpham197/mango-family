<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { Chart as ChartType } from 'chart.js'
import { Chart } from './registry'

// Wrapper mỏng cho biểu đồ tròn/donut (chart.js). Dữ liệu: nhãn + giá trị + màu.
const props = defineProps<{ labels: string[]; values: number[]; colors: string[] }>()

const canvas = ref<HTMLCanvasElement | null>(null)
let chart: ChartType | null = null

function config() {
  return {
    type: 'doughnut' as const,
    data: { labels: props.labels, datasets: [{ data: props.values, backgroundColor: props.colors }] },
    options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false } } },
  }
}

onMounted(() => {
  if (canvas.value) chart = new Chart(canvas.value, config())
})
watch(
  () => [props.labels, props.values],
  () => {
    if (!chart) return
    chart.data.labels = props.labels
    chart.data.datasets[0].data = props.values
    ;(chart.data.datasets[0] as { backgroundColor?: string[] }).backgroundColor = props.colors
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
  <div class="chart-box" data-testid="donut-chart"><canvas ref="canvas"></canvas></div>
</template>

<style scoped>
.chart-box {
  position: relative;
  height: 220px;
}
</style>
