<script setup lang="ts">
import { ref, watch } from 'vue'

// Ô nhập số tiền (VND, số nguyên) — tự format dấu ngăn cách hàng nghìn khi gõ.
// v-model là GIÁ TRỊ SỐ (number | null); hiển thị là chuỗi đã format "3.500.000".
// Dùng cho MỌI ô nhập số tiền để hành vi nhất quán (giao dịch, ngân sách, …).
const props = defineProps<{ modelValue: number | null }>()
const emit = defineEmits<{ 'update:modelValue': [value: number | null] }>()

function fmt(n: number): string {
  return new Intl.NumberFormat('vi-VN').format(n)
}

const display = ref(props.modelValue == null ? '' : fmt(props.modelValue))

// Đồng bộ khi giá trị đổi từ ngoài (vd prefill form sửa). Không ghi đè khi giá trị
// số đã khớp (đang gõ) để không nhảy con trỏ.
watch(
  () => props.modelValue,
  (v) => {
    const currentDigits = display.value.replace(/\D/g, '')
    const nextDigits = v == null ? '' : String(v)
    if (currentDigits !== nextDigits) display.value = v == null ? '' : fmt(v)
  },
)

function onInput(e: Event) {
  const digits = (e.target as HTMLInputElement).value.replace(/\D/g, '')
  if (digits === '') {
    display.value = ''
    emit('update:modelValue', null)
    return
  }
  const n = Number(digits)
  display.value = fmt(n)
  emit('update:modelValue', n)
}
</script>

<template>
  <input
    :value="display"
    type="text"
    inputmode="numeric"
    autocomplete="off"
    placeholder="0"
    @input="onInput"
  />
</template>
