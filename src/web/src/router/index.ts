import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

// Route 004: màn Tổng quan là trang chủ (`/`); sổ giao dịch chuyển sang `/ledger` (D31).
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('../views/LoginView.vue'), meta: { public: true } },
    { path: '/', name: 'dashboard', component: () => import('../views/DashboardView.vue') },
    { path: '/ledger', name: 'ledger', component: () => import('../views/LedgerView.vue') },
    { path: '/overview', redirect: '/' }, // tương thích liên kết cũ (feature 003)
    { path: '/reports', name: 'reports', component: () => import('../views/ReportsView.vue') },
    { path: '/profile', name: 'profile', component: () => import('../views/ProfileView.vue') },
    { path: '/transactions/new', name: 'transaction-new', component: () => import('../views/TransactionFormView.vue') },
    { path: '/transactions/:id/edit', name: 'transaction-edit', component: () => import('../views/TransactionFormView.vue') },
    { path: '/categories', name: 'categories', component: () => import('../views/CategoryManageView.vue') },
    { path: '/categories/new', name: 'category-new', component: () => import('../views/CategoryFormView.vue') },
    { path: '/categories/:id/edit', name: 'category-edit', component: () => import('../views/CategoryFormView.vue') },
    { path: '/budgets', name: 'budgets', component: () => import('../views/BudgetListView.vue') },
    { path: '/budgets/new', name: 'budget-new', component: () => import('../views/BudgetFormView.vue') },
    { path: '/budgets/:id/edit', name: 'budget-edit', component: () => import('../views/BudgetFormView.vue') },
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
    return { name: 'dashboard' } // đăng nhập xong → màn Tổng quan (trang mặc định — FR-001)
  }
})

export default router
