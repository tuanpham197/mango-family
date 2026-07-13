import { expect, type Page } from '@playwright/test'

// Tài khoản seed dev (quickstart Setup, research D9).
export const USERS = {
  alice: { email: 'alice@dev.local', name: 'Alice' },
  bob: { email: 'bob@dev.local', name: 'Bob' },
  carol: { email: 'carol@dev.local', name: 'Carol' },
}
export const DEV_PASSWORD = 'Password123!'

/** Tên duy nhất cho mỗi lần chạy để test lặp lại không đụng dữ liệu cũ. */
export function uniqueName(prefix: string): string {
  return `${prefix}-${Date.now().toString(36)}-${Math.floor(Math.random() * 1e4)}`
}

/** Chọn <option> theo text chứa `text` (selectOption label yêu cầu chuỗi chính xác). */
export async function selectByText(page: Page, testId: string, text: string) {
  const select = page.getByTestId(testId)
  const value = await select.locator('option', { hasText: text }).first().getAttribute('value')
  await select.selectOption(value ?? '')
}

/** Đăng nhập qua UI và chờ vào sổ giao dịch. */
export async function login(page: Page, email: string, password = DEV_PASSWORD) {
  await page.goto('/login')
  await page.getByTestId('login-email').fill(email)
  await page.getByTestId('login-password').fill(password)
  await page.getByTestId('login-submit').click()
  await expect(page.getByTestId('household-name')).toBeVisible()
}

/** Tạo giao dịch qua UI form (002): account chọn sẵn khi hộ có 1 tài khoản. */
export async function createTransaction(
  page: Page,
  opts: { amount: number; type?: 'EXPENSE' | 'INCOME'; category: string; description?: string },
) {
  await page.goto('/transactions/new')
  if (opts.type === 'INCOME') await page.getByTestId('type-income').click()
  else await page.getByTestId('type-expense').click()
  await page.getByTestId('amount-input').fill(String(opts.amount))
  if (opts.description) await page.getByTestId('description-input').fill(opts.description)
  await page.getByTestId(`category-option-${opts.category}`).click()
  await page.getByTestId('save-transaction').click()
  await expect(page).toHaveURL(/\/$/)
}

/** Tạo danh mục qua UI form; đợi lưu xong (điều hướng về /categories) rồi mới trả về. */
export async function createCategory(
  page: Page,
  opts: { name: string; type?: 'EXPENSE' | 'INCOME'; parent?: string; confirmDuplicate?: boolean },
) {
  await page.goto('/categories/new')
  if (opts.type) await page.getByTestId(opts.type === 'EXPENSE' ? 'type-expense' : 'type-income').click()
  await page.getByTestId('category-name').fill(opts.name)
  if (opts.parent) await selectByText(page, 'category-parent', opts.parent)
  await page.getByTestId('save-category').click()
  if (opts.confirmDuplicate) {
    await expect(page.getByTestId('form-error')).toBeVisible()
    await page.getByTestId('save-category').click()
  }
  // Chờ POST + refetch store hoàn tất, tránh bị goto() kế tiếp hủy request đang bay.
  await expect(page).toHaveURL(/\/categories$/)
}
