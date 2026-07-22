import { expect, test } from '@playwright/test'
import { USERS, createCategory, login, uniqueName } from './helpers'

test.beforeEach(async ({ page }) => {
  await login(page, USERS.alice.email)
})

// #5 — Lọc theo loại khi nhập (FR-014, SC-003)
test('#5 picker chỉ hiện danh mục đúng loại đang chọn', async ({ page }) => {
  await page.goto('/transactions/new')

  // Chi: có danh mục Chi mặc định (Ăn uống), không có Thu (Lương)
  await page.getByTestId('type-expense').click()
  await expect(page.getByTestId('category-option-Ăn uống')).toBeVisible()
  await expect(page.getByTestId('category-option-Lương')).toHaveCount(0)

  // Thu: đảo lại
  await page.getByTestId('type-income').click()
  await expect(page.getByTestId('category-option-Lương')).toBeVisible()
  await expect(page.getByTestId('category-option-Ăn uống')).toHaveCount(0)
})

// #6 — Chặn lưu nếu chưa chọn danh mục (FR-013, SC-002)
test('#6 nhập đủ thông tin nhưng thiếu danh mục → bị chặn', async ({ page }) => {
  await page.goto('/transactions/new')
  await page.getByTestId('type-expense').click()
  await page.getByTestId('amount-input').fill('50000')
  await page.getByTestId('save-transaction').click()
  await expect(page.getByTestId('error-category')).toBeVisible()
  await expect(page).toHaveURL(/\/transactions\/new$/) // không điều hướng đi
})

// #4 — Tạo nhanh kế thừa loại (UC-CAT-07)
test('#4 tạo nhanh danh mục trong luồng nhập → mang loại đang chọn & chọn ngay', async ({ page }) => {
  const quick = uniqueName('TạoNhanh')
  await page.goto('/transactions/new')
  await page.getByTestId('type-expense').click()
  await page.getByTestId('quick-create-toggle').click()
  await page.getByTestId('quick-create-name').fill(quick)
  await page.getByTestId('quick-create-submit').click()

  // danh mục mới được chọn ngay cho giao dịch
  await expect(page.getByTestId(`category-option-${quick}`)).toHaveClass(/selected/)

  // lưu được luôn
  await page.getByTestId('amount-input').fill('20000')
  await page.getByTestId('save-transaction').click()
  await expect(page).toHaveURL(/\/ledger$/)

  // danh mục mới là loại Chi (hiện ở tab Chi của màn quản lý)
  await page.goto('/categories')
  await page.getByTestId('tab-expense').click()
  await expect(page.getByText(quick)).toBeVisible()
})

// #12 — Gợi ý danh mục (FR-015, SC-004, D7)
test('#12 học từ lịch sử → gợi ý danh mục cùng loại theo mô tả', async ({ page }) => {
  const keyword = uniqueName('càphê')
  const cat = uniqueName('CàPhê')
  await createCategory(page, { name: cat, type: 'EXPENSE' })

  // dạy hệ thống: giao dịch có mô tả chứa keyword → gán cat
  await page.goto('/transactions/new')
  await page.getByTestId('type-expense').click()
  await page.getByTestId('amount-input').fill('45000')
  await page.getByTestId('description-input').fill(keyword)
  await page.getByTestId(`category-option-${cat}`).click()
  await page.getByTestId('save-transaction').click()
  await expect(page).toHaveURL(/\/ledger$/)

  // lần nhập sau: mô tả chứa keyword → hiện chip gợi ý, bấm để điền
  await page.goto('/transactions/new')
  await page.getByTestId('type-expense').click()
  await page.getByTestId('description-input').fill(`mua ${keyword} sáng`)
  await expect(page.getByTestId('suggestion-chip')).toContainText(cat)
  await page.getByTestId('suggestion-chip').click()
  await expect(page.getByTestId(`category-option-${cat}`)).toHaveClass(/selected/)
})
