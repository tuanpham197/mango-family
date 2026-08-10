<script setup lang="ts">
import type { MembersReport } from '../api/types'

// Báo cáo theo thành viên (008 — FR-001/004/005/010): bảng Thu/Chi/ròng mỗi thành viên
// hiện tại (0/0/0 nếu không có giao dịch) + dòng "Thành viên cũ" khi có. Bảng là dạng
// trình bày chính (Assumptions); biểu đồ cột so sánh là tùy chọn — hoãn (xem spec A1).
defineProps<{ report: MembersReport }>()
const emit = defineEmits<{ select: [id: string] }>()

function money(n: number) {
  return new Intl.NumberFormat('vi-VN').format(Math.round(n))
}
</script>

<template>
  <div class="card" data-testid="member-breakdown">
    <div class="mb-head">
      <h2>Theo thành viên</h2>
      <span class="muted">Ròng = Thu − Chi</span>
    </div>

    <table class="mb-table">
      <thead>
        <tr>
          <th>Thành viên</th>
          <th class="num">Thu</th>
          <th class="num">Chi</th>
          <th class="num">Ròng</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="m in report.members"
          :key="m.member_id"
          class="mb-row"
          :class="{ former: m.is_former }"
          data-testid="member-row"
          @click="emit('select', m.member_id)"
        >
          <td class="name" data-testid="member-name">{{ m.display_name }}</td>
          <td class="num amount-income">+{{ money(m.income) }}</td>
          <td class="num amount-expense">−{{ money(m.expense) }}</td>
          <td class="num" :class="m.net >= 0 ? 'amount-income' : 'amount-expense'">{{ money(m.net) }}</td>
        </tr>
      </tbody>
      <tfoot>
        <tr class="mb-total" data-testid="member-totals">
          <td>Tổng hộ</td>
          <td class="num">+{{ money(report.totals.income) }}</td>
          <td class="num">−{{ money(report.totals.expense) }}</td>
          <td class="num">{{ money(report.totals.net) }}</td>
        </tr>
      </tfoot>
    </table>
  </div>
</template>

<style scoped>
.mb-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 10px;
}
.mb-head h2 {
  font-size: 16px;
  margin: 0;
}
.mb-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}
.mb-table th,
.mb-table td {
  padding: 8px 6px;
  border-bottom: 1px solid var(--border);
  text-align: left;
}
.mb-table th.num,
.mb-table td.num {
  text-align: right;
  white-space: nowrap;
}
.mb-table thead th {
  color: var(--muted);
  font-weight: 600;
  font-size: 12px;
}
.mb-row {
  cursor: pointer;
}
.mb-row:hover {
  background: var(--hover, rgba(0, 0, 0, 0.03));
}
.mb-row.former .name {
  font-style: italic;
  color: var(--muted);
}
.mb-total td {
  font-weight: 700;
  border-bottom: none;
  border-top: 2px solid var(--border);
}
</style>
