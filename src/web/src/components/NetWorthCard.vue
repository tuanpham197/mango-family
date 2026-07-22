<script setup lang="ts">
// Thẻ Tổng tài sản ròng (dashboard.png — thẻ đầu tiên): Σ số dư tài khoản + % so
// tháng trước (ẩn khi null — D28).
defineProps<{ netWorth: number; changePercent: number | null }>()

function money(n: number) {
  return new Intl.NumberFormat('vi-VN').format(Math.round(n))
}
</script>

<template>
  <div class="networth" :class="{ negative: netWorth < 0 }" data-testid="networth-card">
    <div class="nw-label">Tổng tài sản ròng</div>
    <div class="nw-amount" data-testid="networth-amount">{{ money(netWorth) }} đ</div>
    <div
      v-if="changePercent !== null"
      class="nw-change"
      :class="{ up: changePercent >= 0, down: changePercent < 0 }"
      data-testid="networth-change"
    >
      {{ changePercent >= 0 ? '▲' : '▼' }} {{ Math.abs(changePercent) }}% so với tháng trước
    </div>
  </div>
</template>

<style scoped>
.networth {
  background: var(--green);
  color: #fff;
  border-radius: 16px;
  padding: 18px 20px;
  margin-bottom: 12px;
}
/* Tài sản ròng âm → thẻ đỏ (cảnh báo). % thay đổi so tháng trước vẫn hiển ▲/▼. */
.networth.negative {
  background: var(--red);
}
.nw-label {
  font-size: 14px;
  opacity: 0.9;
}
.nw-amount {
  font-size: 30px;
  font-weight: 700;
  margin: 4px 0;
}
.nw-change {
  font-size: 13px;
  opacity: 0.95;
}
</style>
