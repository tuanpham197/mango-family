<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { initials } from '../utils/avatar'

// Trang cá nhân — mở từ avatar góc phải màn Tổng quan.
const auth = useAuthStore()
const router = useRouter()

async function logout() {
  await auth.logout()
  router.push('/login')
}
</script>

<template>
  <div>
    <div class="page-head profile-head">
      <button class="btn btn-outline" data-testid="profile-back" @click="router.push('/')">‹ Quay lại</button>
      <h1>Trang cá nhân</h1>
    </div>

    <div class="card profile-card" data-testid="profile-card">
      <div class="profile-avatar" data-testid="profile-avatar">{{ initials(auth.user?.display_name) }}</div>
      <h2 class="profile-name" data-testid="profile-name">{{ auth.user?.display_name }}</h2>
      <p class="muted" data-testid="profile-email">{{ auth.user?.email }}</p>
    </div>

    <div class="card">
      <div class="info-row">
        <span class="muted">Hộ gia đình</span>
        <span data-testid="profile-household">{{ auth.household?.name ?? '—' }}</span>
      </div>
    </div>

    <button class="btn btn-danger btn-block" data-testid="profile-logout" @click="logout">Đăng xuất</button>
  </div>
</template>

<style scoped>
.profile-head {
  display: flex;
  align-items: center;
  gap: 10px;
}
.profile-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 22px 16px;
}
.profile-avatar {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: var(--green);
  color: #fff;
  font-weight: 700;
  font-size: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
}
.profile-name {
  margin: 0 0 2px;
  font-size: 20px;
}
.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
