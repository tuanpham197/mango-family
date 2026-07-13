<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useAuthStore } from './stores/auth'

const auth = useAuthStore()
const route = useRoute()
</script>

<template>
  <div class="app">
    <main class="app-main">
      <RouterView />
    </main>
    <nav v-if="auth.isAuthenticated && route.name !== 'login'" class="bottom-nav" data-testid="bottom-nav">
      <RouterLink to="/" class="nav-item" :class="{ active: route.name === 'ledger' }">Sổ giao dịch</RouterLink>
      <RouterLink to="/transactions/new" class="nav-fab" data-testid="nav-new-transaction">＋</RouterLink>
      <RouterLink to="/categories" class="nav-item" :class="{ active: String(route.name).startsWith('categor') }">
        Danh mục
      </RouterLink>
    </nav>
  </div>
</template>
