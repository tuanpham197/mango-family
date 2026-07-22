import { expect, test, type Page } from '@playwright/test'
import { USERS, createBudget, createCategory, createTransaction, login, uniqueName, DEV_PASSWORD } from './helpers'

// Màn Tổng quan (feature 004) — kịch bản một ngữ cảnh (quickstart #1–24). Dùng danh
// mục riêng (uniqueName) cho các khẳng định cần cô lập khỏi dữ liệu tích luỹ của hộ.

async function y(page: Page, testId: string): Promise<number> {
  const box = await page.getByTestId(testId).boundingBox()
  return box?.y ?? -1
}

// #1 — đăng nhập không kèm đích → màn Tổng quan (KHÔNG dùng helper vì helper chuyển sang /ledger)
test('#1 đăng nhập → vào thẳng màn Tổng quan', async ({ page }) => {
  await page.goto('/login')
  await page.getByTestId('login-email').fill(USERS.alice.email)
  await page.getByTestId('login-password').fill(DEV_PASSWORD)
  await page.getByTestId('login-submit').click()
  await expect(page).toHaveURL(/\/$/)
  await expect(page.getByTestId('networth-card')).toBeVisible()
})

// #3 — liên kết cụ thể được tôn trọng
test('#3 mở liên kết cụ thể không bị ép về Tổng quan', async ({ page }) => {
  await login(page, USERS.alice.email)
  await page.goto('/budgets')
  await expect(page).toHaveURL(/\/budgets$/)
})

// #4 / #22 / #23 — điều hướng 5 mục + Báo cáo placeholder + lối phụ Danh mục
test('#4,#22,#23 thanh điều hướng 5 mục + Báo cáo + Danh mục lối phụ', async ({ page }) => {
  await login(page, USERS.alice.email)
  await page.goto('/')
  const nav = page.getByTestId('bottom-nav')
  for (const id of ['nav-overview', 'nav-ledger', 'nav-new-transaction', 'nav-budgets', 'nav-reports']) {
    await expect(nav.getByTestId(id)).toBeVisible()
  }
  // "Báo cáo" nay là màn thật (feature 005) — bộ chọn khoảng hiển thị ngay
  await page.getByTestId('nav-reports').click()
  await expect(page.getByTestId('time-range-picker')).toBeVisible()
  await page.goto('/')
  await page.getByTestId('manage-categories').click() // lối phụ vào Danh mục (D32)
  await expect(page).toHaveURL(/\/categories$/)
})

// #5 — Thu/Chi tháng phản ánh giao dịch (kiểm qua delta thẻ Chi phí)
test('#5 thẻ Thu nhập/Chi phí phản ánh giao dịch tháng', async ({ page }) => {
  await login(page, USERS.alice.email)
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await page.goto('/')
  const before = Math.abs(await readMoney(page, 'expense-amount'))
  await createTransaction(page, { amount: 123000, category: cat, description: uniqueName('t') })
  await page.goto('/')
  const after = Math.abs(await readMoney(page, 'expense-amount'))
  expect(after - before).toBe(123000)
})

// #9 — Chi tiêu theo danh mục liệt kê danh mục có chi
test('#9 chi tiêu theo danh mục hiển thị danh mục + số tiền', async ({ page }) => {
  await login(page, USERS.alice.email)
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createTransaction(page, { amount: 250000, category: cat, description: uniqueName('t') })
  await page.goto('/')
  const row = page.getByTestId('category-spending-row').filter({ hasText: cat })
  await expect(row).toHaveCount(1)
  await expect(row).toContainText('250.000')
})

// #8 / #12 / #21 — bố cục đúng thứ tự dashboard.png
test('#8,#12,#21 thứ tự phần: tài sản ròng → Thu/Chi → chi theo danh mục → ngân sách → gần đây', async ({ page }) => {
  await login(page, USERS.alice.email)
  await page.goto('/')
  await expect(page.getByTestId('networth-card')).toBeVisible()
  const nw = await y(page, 'networth-card')
  const ie = await y(page, 'income-card')
  const cs = await y(page, 'category-spending')
  const bg = await y(page, 'overview-budget-widget')
  const rt = await y(page, 'recent-transactions')
  expect(nw).toBeLessThan(ie)
  expect(ie).toBeLessThan(cs)
  expect(cs).toBeLessThan(bg) // Chi tiêu theo danh mục TRƯỚC Ngân sách (#12)
  expect(bg).toBeLessThan(rt)
})

// #14 — Tổng tài sản ròng = tổng số dư tài khoản
test('#14 tài sản ròng khớp tổng số dư tài khoản', async ({ page }) => {
  await login(page, USERS.alice.email)
  await page.goto('/')
  await expect(page.getByTestId('networth-card')).toBeVisible()
  const shown = await readMoney(page, 'networth-amount')
  const accs = (await (await page.request.get('/api/accounts')).json()).data as Array<{ balance: number }>
  const expected = accs.reduce((s, a) => s + a.balance, 0)
  expect(shown).toBe(expected)
})

// #17 — Giao dịch gần đây hiển thị giao dịch mới nhất
test('#17 giao dịch gần đây hiển thị giao dịch vừa nhập', async ({ page }) => {
  await login(page, USERS.alice.email)
  const cat = uniqueName('An')
  const d = uniqueName('recent')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createTransaction(page, { amount: 45000, category: cat, description: d })
  await page.goto('/')
  await expect(page.getByTestId('recent-transactions').filter({ hasText: d })).toBeVisible()
  // "Xem tất cả ›" mở màn Giao dịch (/ledger)
  await page.getByTestId('recent-see-all').click()
  await expect(page).toHaveURL(/\/ledger$/)
})

// #20 — lời chào + tên
test('#20 màn Tổng quan có lời chào + tên thành viên', async ({ page }) => {
  await login(page, USERS.alice.email)
  await page.goto('/')
  await expect(page.getByTestId('greeting')).toContainText(/Chào buổi/)
  await expect(page.getByTestId('current-user')).toHaveText(USERS.alice.name)
})

// #25 — Avatar góc phải → trang cá nhân, và đăng xuất từ đó
test('#25 click avatar mở trang cá nhân + đăng xuất', async ({ page }) => {
  await login(page, USERS.alice.email)
  await page.goto('/')
  await page.getByTestId('avatar').click()
  await expect(page).toHaveURL(/\/profile$/)
  await expect(page.getByTestId('profile-name')).toHaveText(USERS.alice.name)
  await expect(page.getByTestId('profile-email')).toContainText('@')
  await page.getByTestId('profile-logout').click()
  await expect(page).toHaveURL(/\/login$/)
})

// #24 — widget ngân sách + "Xem tất cả"
test('#24 tóm tắt ngân sách + Xem tất cả trên Tổng quan', async ({ page }) => {
  await login(page, USERS.alice.email)
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createBudget(page, { category: cat, limit: 1000000, period: 'MONTHLY' })
  await page.goto('/')
  await expect(page.getByTestId('overview-budget-widget')).toBeVisible()
  await page.getByTestId('budget-see-all').click()
  await expect(page).toHaveURL(/\/budgets$/)
})

async function readMoney(page: Page, testId: string): Promise<number> {
  const text = (await page.getByTestId(testId).textContent()) ?? '0'
  const negative = text.includes('−') || text.includes('-')
  const digits = Number(text.replace(/[^\d]/g, ''))
  return negative ? -digits : digits
}
