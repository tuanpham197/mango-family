// Đăng ký tree-shake chỉ các thành phần chart.js cần (D39) — donut (phân bổ) + line (xu hướng).
import {
  ArcElement,
  CategoryScale,
  Chart,
  DoughnutController,
  Filler,
  Legend,
  LineController,
  LineElement,
  LinearScale,
  PointElement,
  Tooltip,
} from 'chart.js'

Chart.register(
  DoughnutController,
  ArcElement,
  LineController,
  LineElement,
  PointElement,
  LinearScale,
  CategoryScale,
  Tooltip,
  Legend,
  Filler,
)

export { Chart }
