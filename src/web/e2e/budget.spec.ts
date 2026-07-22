import { expect, test, type Page } from '@playwright/test'
import { USERS, budgetRow, createBudget, createCategory, createTransaction, login, uniqueName } from './helpers'

// Ngân sách feature 003 — kịch bản một ngữ cảnh (quickstart #1–5, #7–10, #12–14, #20–24, #27–28, #30–34).
// Mỗi test dùng danh mục riêng (uniqueName) để tiến độ được cô lập khỏi dữ liệu khác.

test.beforeEach(async ({ page }) => {
  await login(page, USERS.alice.email)
})

// Xóa giao dịch theo mô tả (mỗi mô tả duy nhất = 1 giao dịch). Chờ sổ nạp xong
// (row hiện ra) trước khi thao tác để tránh đọc danh sách khi fetch chưa hoàn tất.
async function deleteAllTxns(page: Page, desc: string) {
  await page.goto('/ledger')
  const row = page.getByTestId('transaction-row').filter({ hasText: desc }).first()
  await expect(row).toBeVisible()
  const delId = (await row.getByRole('button').getAttribute('data-testid'))!
  await page.getByTestId(delId).click()
  await page.getByTestId('confirm-delete-transaction').click()
  await expect(page.getByTestId('transaction-row').filter({ hasText: desc })).toHaveCount(0)
}

// #1 / #7 / #33 — tạo ngân sách danh mục; tiến độ từ chi tiêu hiện có (từ đầu kỳ)
test('#1,#7,#33 tạo ngân sách danh mục → xuất hiện kèm tiến độ 70%', async ({ page }) => {
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createTransaction(page, { amount: 3500000, category: cat, description: uniqueName('chi') })
  await createBudget(page, { category: cat, limit: 5000000, period: 'MONTHLY' })
  const row = budgetRow(page, cat)
  await expect(row).toContainText('70%')
  await expect(row.getByTestId('progress-amount')).toContainText('3.500.000')
})

// #2 — giới hạn ≤ 0 bị chặn
test('#2 giới hạn 0 bị chặn (LIMIT_INVALID)', async ({ page }) => {
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await page.goto('/budgets/new')
  await page.getByTestId('limit-input').fill('0')
  await page.getByTestId(`category-option-${cat}`).click()
  await page.getByTestId('save-budget').click()
  await expect(page.getByTestId('error-limit')).toBeVisible()
})

// #3 — bắt buộc danh mục; picker chỉ danh mục Chi
test('#3 thiếu danh mục bị chặn; picker chỉ hiện danh mục Chi', async ({ page }) => {
  await page.goto('/budgets/new')
  await page.getByTestId('limit-input').fill('5000000')
  await page.getByTestId('save-budget').click()
  await expect(page.getByTestId('error-category')).toBeVisible()
  // Danh mục Thu (Lương) không xuất hiện trong picker ngân sách
  await expect(page.getByTestId('category-option-Lương')).toHaveCount(0)
})

// #4 — trùng danh mục/kỳ → BUDGET_DUPLICATE, chỉ tới ngân sách hiện có
test('#4 trùng ngân sách danh mục/kỳ bị chặn, chỉ tới cái hiện có', async ({ page }) => {
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createBudget(page, { category: cat, limit: 5000000, period: 'MONTHLY' })
  await page.goto('/budgets/new')
  await page.getByTestId('limit-input').fill('9000000')
  await page.getByTestId(`category-option-${cat}`).click()
  await page.getByTestId('save-budget').click()
  await expect(page.getByTestId('form-error')).toBeVisible()
  await expect(page.getByTestId('goto-existing-budget')).toBeVisible()
})

// #5 — kỳ một lần ngày sai
test('#5 kỳ một lần ngày kết thúc < bắt đầu bị chặn (PERIOD_INVALID)', async ({ page }) => {
  await page.goto('/budgets/new')
  await page.getByTestId('type-total').click()
  await page.getByTestId('limit-input').fill('1000000')
  await page.getByTestId('period-select').selectOption('ONE_TIME')
  await page.getByTestId('start-date-input').fill('2026-07-10')
  await page.getByTestId('end-date-input').fill('2026-07-05')
  await page.getByTestId('save-budget').click()
  await expect(page.getByTestId('error-period')).toBeVisible()
})

// #8 / #9 — thêm rồi xóa giao dịch → tiến độ cập nhật đúng
test('#8,#9 thêm giao dịch → 80%; xóa → về 70%', async ({ page }) => {
  const cat = uniqueName('An')
  const d = uniqueName('them')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createTransaction(page, { amount: 3500000, category: cat, description: uniqueName('base') })
  await createBudget(page, { category: cat, limit: 5000000, period: 'MONTHLY' })

  await createTransaction(page, { amount: 500000, category: cat, description: d })
  await page.goto('/budgets')
  await expect(budgetRow(page, cat)).toContainText('80%')

  await deleteAllTxns(page, d)
  await page.goto('/budgets')
  await expect(budgetRow(page, cat)).toContainText('70%')
})

// #12 — danh mục con tính vào ngân sách cha
test('#12 chi ở danh mục con tính vào ngân sách danh mục cha', async ({ page }) => {
  const parent = uniqueName('Cha')
  const child = uniqueName('Con')
  await createCategory(page, { name: parent, type: 'EXPENSE' })
  await createCategory(page, { name: child, type: 'EXPENSE', parent })
  await createBudget(page, { category: parent, limit: 1000000, period: 'MONTHLY' })
  await createTransaction(page, { amount: 400000, category: child, description: uniqueName('con-chi') })
  await page.goto('/budgets')
  await expect(budgetRow(page, parent)).toContainText('40%') // 400k/1M
})

// #13 — đổi loại Chi → Thu: không còn tính vào ngân sách
test('#13 đổi giao dịch Chi → Thu → không còn tính vào ngân sách', async ({ page }) => {
  const cat = uniqueName('An')
  const d = uniqueName('doi-loai')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createBudget(page, { category: cat, limit: 1000000, period: 'MONTHLY' })
  await createTransaction(page, { amount: 500000, category: cat, description: d })
  await page.goto('/budgets')
  await expect(budgetRow(page, cat)).toContainText('50%')

  // Sửa giao dịch sang Thu (danh mục Lương)
  await page.goto('/ledger')
  await page.getByTestId('transaction-row').filter({ hasText: d }).first().click()
  await page.getByTestId('type-income').click()
  await page.getByTestId('category-option-Lương').click()
  await page.getByTestId('save-transaction').click()
  await expect(page).toHaveURL(/\/ledger$/)
  await page.goto('/budgets')
  await expect(budgetRow(page, cat)).toContainText('0%')
})

// #14 — kỳ chưa có giao dịch → 0%, không cảnh báo
test('#14 ngân sách chưa có chi → 0%, không cảnh báo', async ({ page }) => {
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createBudget(page, { category: cat, limit: 2000000, period: 'MONTHLY' })
  const row = budgetRow(page, cat)
  await expect(row).toContainText('0%')
  await expect(row.getByTestId('alert-threshold')).toHaveCount(0)
  await expect(row.getByTestId('alert-over')).toHaveCount(0)
})

// #15 / #16 / #18 — cảnh báo 80%, vượt (số tiền vượt), phát lại sau khi tụt dưới
test('#15,#16,#18 cảnh báo 80% → vượt 50.000 → re-arm sau khi tụt dưới', async ({ page }) => {
  const cat = uniqueName('An')
  const dA = uniqueName('a')
  const dB = uniqueName('b')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createBudget(page, { category: cat, limit: 1000000, period: 'MONTHLY' })

  // 85% → cảnh báo 80% (realtime qua budgets_changed)
  await createTransaction(page, { amount: 850000, category: cat, description: dA })
  await page.goto('/budgets')
  await expect(budgetRow(page, cat).getByTestId('alert-threshold')).toBeVisible({ timeout: 8000 })

  // vượt 105% → cảnh báo vượt kèm đúng 50.000 (SC-005)
  await createTransaction(page, { amount: 200000, category: cat, description: dB })
  await page.goto('/budgets')
  const over = budgetRow(page, cat).getByTestId('alert-over')
  await expect(over).toBeVisible({ timeout: 8000 })
  await expect(over).toContainText('50.000')

  // tụt về 0 → cảnh báo tự gỡ (re-arm — US3 #4)
  await deleteAllTxns(page, dA)
  await deleteAllTxns(page, dB)
  await page.goto('/budgets')
  await expect(budgetRow(page, cat).getByTestId('alert-over')).toHaveCount(0, { timeout: 8000 })
  await expect(budgetRow(page, cat).getByTestId('alert-threshold')).toHaveCount(0)
})

// #20 / #21 / #22 — ngân sách tổng: tiến độ = tổng chi hộ; chặn trùng; song song NS danh mục
test('#20,#21,#22 ngân sách tổng chi tiêu', async ({ page }) => {
  // Dọn ngân sách tổng cũ (nếu có) để tránh nhiễu
  await removeTotalIfAny(page)

  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createBudget(page, { category: cat, limit: 5000000, period: 'MONTHLY' })
  await createTransaction(page, { amount: 300000, category: cat, description: uniqueName('t') })

  await createBudget(page, { type: 'TOTAL', limit: 50000000, period: 'MONTHLY' })
  const total = await monthlyExpense(page)
  const row = budgetRow(page, 'Tổng chi tiêu')
  await expect(row).toContainText(new Intl.NumberFormat('vi-VN').format(total))

  // #21 trùng ngân sách tổng cùng kỳ → chặn
  await page.goto('/budgets/new')
  await page.getByTestId('type-total').click()
  await page.getByTestId('limit-input').fill('60000000')
  await page.getByTestId('save-budget').click()
  await expect(page.getByTestId('form-error')).toBeVisible()

  // #22 cả hai tồn tại song song
  await page.goto('/budgets')
  await expect(budgetRow(page, 'Tổng chi tiêu')).toHaveCount(1)
  await expect(budgetRow(page, cat)).toHaveCount(1)
})

// #23 / #24 — sửa giới hạn → tiến độ tính lại; sửa giới hạn không hợp lệ bị chặn
test('#23,#24 sửa giới hạn → tiến độ tính lại; giới hạn 0 bị chặn', async ({ page }) => {
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createTransaction(page, { amount: 3500000, category: cat, description: uniqueName('s') })
  await createBudget(page, { category: cat, limit: 5000000, period: 'MONTHLY' })
  await expect(budgetRow(page, cat)).toContainText('70%')

  // sửa giới hạn → 4.000.000 (87,5%) — chờ prefill xong (getById) rồi mới sửa
  await budgetRow(page, cat).click()
  await expect(page.getByRole('heading', { name: 'Sửa ngân sách' })).toBeVisible()
  await expect(page.getByTestId('limit-input')).toHaveValue('5.000.000') // MoneyInput tự format
  await page.getByTestId('limit-input').fill('4000000')
  await page.getByTestId('save-budget').click()
  await expect(page).toHaveURL(/\/budgets$/)
  await expect(budgetRow(page, cat)).toContainText(/87[.,]5%/)

  // #24 sửa giới hạn 0 → chặn
  await budgetRow(page, cat).click()
  await expect(page.getByTestId('limit-input')).toHaveValue('4.000.000') // chờ prefill (đã format)
  await page.getByTestId('limit-input').fill('0')
  await page.getByTestId('save-budget').click()
  await expect(page.getByTestId('error-limit')).toBeVisible()
})

// #27 / #28 — xóa có xác nhận; hủy không đổi
test('#27,#28 xóa qua xác nhận; hủy không đổi gì', async ({ page }) => {
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createBudget(page, { category: cat, limit: 1000000, period: 'MONTHLY' })

  // #28 hủy
  const delId = `delete-budget-${await budgetId(page, cat)}`
  await page.getByTestId(delId).click()
  await expect(page.getByTestId('delete-budget-dialog')).toBeVisible()
  await page.getByTestId('cancel-delete-budget').click()
  await expect(budgetRow(page, cat)).toHaveCount(1)

  // #27 xác nhận
  await page.getByTestId(delId).click()
  await page.getByTestId('confirm-delete-budget').click()
  await expect(budgetRow(page, cat)).toHaveCount(0)
})

// #30 — danh mục có ngân sách bị ẩn → vẫn theo dõi + nhãn "đã ẩn"
test('#30 ẩn danh mục có ngân sách → vẫn theo dõi, hiển thị nhãn đã ẩn', async ({ page }) => {
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createBudget(page, { category: cat, limit: 1000000, period: 'MONTHLY' })
  // Ẩn danh mục qua màn quản lý danh mục
  await hideCategory(page, cat)
  await page.goto('/budgets')
  await expect(budgetRow(page, cat).getByTestId('category-hidden-tag')).toBeVisible()
})

// #31 — ngân sách một lần đã qua kỳ → ENDED (xem ở status=all)
test('#31 ngân sách một lần đã qua end_date → ENDED', async ({ page }) => {
  await createBudget(page, { type: 'TOTAL', limit: 1000000, period: 'ONE_TIME', start: '2026-06-01', end: '2026-06-30' })
  // Danh sách ACTIVE mặc định không còn hiển thị; kiểm tra qua API status=all
  const res = await page.request.get('/api/budgets?status=all')
  const items = (await res.json()).data as Array<{ period_type: string; status: string }>
  const ended = items.find((b) => b.period_type === 'ONE_TIME')
  expect(ended?.status).toBe('ENDED')
})

// #34 — tóm tắt ngân sách trên màn Tổng quan + "Xem tất cả"
test('#34 màn Tổng quan hiển thị tóm tắt ngân sách + Xem tất cả', async ({ page }) => {
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createTransaction(page, { amount: 900000, category: cat, description: uniqueName('ov') })
  await createBudget(page, { category: cat, limit: 1000000, period: 'MONTHLY' })

  await page.goto('/overview')
  await expect(page.getByTestId('overview-budget-widget')).toBeVisible()
  await expect(page.getByTestId('overview-budget-widget')).toContainText('90%')
  await page.getByTestId('budget-see-all').click()
  await expect(page).toHaveURL(/\/budgets$/)
})

// --- helpers cục bộ ---

async function budgetId(page: Page, label: string): Promise<string> {
  const res = await page.request.get('/api/budgets?status=all')
  const items = (await res.json()).data as Array<{ id: string; category_name: string | null; type: string }>
  const b = items.find((x) => x.category_name === label || (label === 'Tổng chi tiêu' && x.type === 'TOTAL'))
  return b!.id
}

async function removeTotalIfAny(page: Page) {
  const res = await page.request.get('/api/budgets?status=all')
  const items = (await res.json()).data as Array<{ id: string; type: string; status: string }>
  for (const b of items.filter((x) => x.type === 'TOTAL' && x.status === 'ACTIVE')) {
    await page.request.delete(`/api/budgets/${b.id}`, { data: {} })
  }
}

async function monthlyExpense(page: Page): Promise<number> {
  const res = await page.request.get('/api/transactions?page_size=1000')
  const txns = (await res.json()).data as Array<{ amount: number; type: string; transaction_date: string }>
  const now = new Date()
  return txns
    .filter((t) => t.type === 'EXPENSE')
    .filter((t) => {
      const d = new Date(t.transaction_date)
      return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth()
    })
    .reduce((s, t) => s + t.amount, 0)
}

async function hideCategory(page: Page, name: string) {
  const res = await page.request.get('/api/categories?include_hidden=true')
  const trees = (await res.json()).data as Array<{ id: string; name: string; updated_at: string }>
  const cat = trees.find((c) => c.name === name)!
  await page.request.patch(`/api/categories/${cat.id}`, {
    data: { is_hidden: true, expected_updated_at: cat.updated_at },
  })
}
