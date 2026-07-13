<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError } from '../api/client'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const error = ref('')
const busy = ref(false)

async function submit() {
  error.value = ''
  busy.value = true
  try {
    await auth.login(email.value.trim(), password.value)
    router.push(typeof route.query.redirect === 'string' ? route.query.redirect : '/')
  } catch (e) {
    if (e instanceof ApiError && e.code === 'NO_HOUSEHOLD') {
      error.value = 'Tài khoản chưa thuộc hộ gia đình nào — hãy liên hệ để được thêm vào một hộ.'
    } else if (e instanceof ApiError) {
      error.value = e.message
    } else {
      error.value = 'Không kết nối được máy chủ, vui lòng thử lại.'
    }
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="login">
    <h1>Đăng nhập</h1>
    <p class="muted">Sổ thu chi chung của hộ gia đình</p>
    <form @submit.prevent="submit">
      <div v-if="error" class="form-error" data-testid="login-error">{{ error }}</div>
      <div class="field">
        <label for="email">Email</label>
        <input id="email" v-model="email" type="email" autocomplete="username" required data-testid="login-email" />
      </div>
      <div class="field">
        <label for="password">Mật khẩu</label>
        <input
          id="password"
          v-model="password"
          type="password"
          autocomplete="current-password"
          required
          data-testid="login-password"
        />
      </div>
      <button class="btn btn-primary btn-block" type="submit" :disabled="busy" data-testid="login-submit">
        {{ busy ? 'Đang đăng nhập…' : 'Đăng nhập' }}
      </button>
    </form>
  </div>
</template>

<style scoped>
.login {
  padding-top: 15vh;
}
</style>
