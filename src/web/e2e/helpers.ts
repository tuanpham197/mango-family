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
  await expect(page.getByTestId('household-name')).toBeVisible() // đăng nhập xong → màn Tổng quan (/)
  // 001/002/003 thao tác trên sổ giao dịch → điều hướng tới /ledger (route mới 004)
  await page.goto('/ledger')
  await expect(page.getByTestId('ledger')).toBeVisible()
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
  await expect(page).toHaveURL(/\/ledger$/)
}

/** Tạo ngân sách qua UI form; đợi lưu xong (điều hướng về /budgets) rồi trả về. */
export async function createBudget(
  page: Page,
  opts: {
    type?: 'CATEGORY' | 'TOTAL'
    category?: string
    limit: number
    period?: 'MONTHLY' | 'WEEKLY' | 'ONE_TIME'
    start?: string
    end?: string
  },
) {
  await page.goto('/budgets/new')
  if (opts.type === 'TOTAL') await page.getByTestId('type-total').click()
  await page.getByTestId('limit-input').fill(String(opts.limit))
  if (opts.period) await page.getByTestId('period-select').selectOption(opts.period)
  if (opts.type !== 'TOTAL' && opts.category) {
    await page.getByTestId(`category-option-${opts.category}`).click()
  }
  if (opts.period === 'ONE_TIME') {
    if (opts.start) await page.getByTestId('start-date-input').fill(opts.start)
    if (opts.end) await page.getByTestId('end-date-input').fill(opts.end)
  }
  await page.getByTestId('save-budget').click()
  await expect(page).toHaveURL(/\/budgets$/)
}

/** Dòng ngân sách theo nhãn (tên danh mục / "Tổng chi tiêu"). */
export function budgetRow(page: Page, label: string) {
  return page.getByTestId('budget-row').filter({ hasText: label })
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
