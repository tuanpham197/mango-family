<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useAuthStore } from './stores/auth'
import { useHead } from '@vueuse/head'

const auth = useAuthStore()
const route = useRoute()
// Chỉ đặt title động ở client. Thẻ og:* đã khai báo TĨNH trong index.html
// (crawler không chạy JS nên og phải nằm sẵn trong HTML gốc).
useHead({ title: 'Sổ Thu Chi' })
</script>

<template>
  <div class="app">
    <main class="app-main">
      <RouterView />
    </main>
    <!-- Thanh điều hướng 5 mục theo dashboard.png (feature 004): Tổng quan · Giao dịch · ＋ · Ngân sách · Báo cáo -->
    <nav v-if="auth.isAuthenticated && route.name !== 'login'" class="bottom-nav" data-testid="bottom-nav">
      <RouterLink to="/" class="nav-item" :class="{ active: route.name === 'dashboard' }" data-testid="nav-overview">
        Tổng quan
      </RouterLink>
      <RouterLink to="/ledger" class="nav-item" :class="{ active: route.name === 'ledger' }" data-testid="nav-ledger">
        Giao dịch
      </RouterLink>
      <RouterLink to="/transactions/new" class="nav-fab" data-testid="nav-new-transaction">＋</RouterLink>
      <RouterLink to="/budgets" class="nav-item" :class="{ active: String(route.name).startsWith('budget') }" data-testid="nav-budgets">
        Ngân sách
      </RouterLink>
      <RouterLink to="/reports" class="nav-item" :class="{ active: route.name === 'reports' }" data-testid="nav-reports">
        Báo cáo
      </RouterLink>
    </nav>
  </div>
</template>
