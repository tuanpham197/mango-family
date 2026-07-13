import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

// Khung route 001: login + sổ tối thiểu + nhập giao dịch + quản lý danh mục.
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('../views/LoginView.vue'), meta: { public: true } },
    { path: '/', name: 'ledger', component: () => import('../views/LedgerView.vue') },
    { path: '/transactions/new', name: 'transaction-new', component: () => import('../views/TransactionEntryView.vue') },
    { path: '/categories', name: 'categories', component: () => import('../views/CategoryManageView.vue') },
    { path: '/categories/new', name: 'category-new', component: () => import('../views/CategoryFormView.vue') },
    { path: '/categories/:id/edit', name: 'category-edit', component: () => import('../views/CategoryFormView.vue') },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

// Guard: chưa đăng nhập → /login (bootstrap phiên qua GET /api/me — T012).
router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.bootstrap()
  if (!to.meta.public && !auth.isAuthenticated) {
    return { name: 'login', query: to.fullPath !== '/' ? { redirect: to.fullPath } : {} }
  }
  if (to.meta.public && auth.isAuthenticated) {
    return { name: 'ledger' }
  }
})

export default router
