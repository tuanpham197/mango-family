import { expect, test } from '@playwright/test'
import { USERS, createTransaction, login, uniqueName } from './helpers'

test.beforeEach(async ({ page }) => {
  await login(page, USERS.alice.email)
})

// #4 — Nhập giao dịch hợp lệ → số dư giảm (UC-TRK-02 AC-1)
test('#4 nhập Chi 50.000 → xuất hiện trong sổ, số dư Tiền mặt giảm', async ({ page }) => {
  const before = await readBalance(page, 'Tiền mặt')
  await createTransaction(page, { amount: 50000, category: 'Ăn uống', description: uniqueName('bữa trưa') })
  const after = await readBalance(page, 'Tiền mặt')
  expect(before - after).toBe(50000) // Chi làm số dư giảm đúng 50.000
  await expect(page.getByTestId('transaction-row').first()).toContainText('Ăn uống')
})

// #5 — Số tiền không hợp lệ (AMOUNT_INVALID)
test('#5 số tiền 0 bị chặn', async ({ page }) => {
  await page.goto('/transactions/new')
  await page.getByTestId('type-expense').click()
  await page.getByTestId('amount-input').fill('0')
  await page.getByTestId('category-option-Ăn uống').click()
  await page.getByTestId('save-transaction').click()
  await expect(page.getByTestId('error-amount')).toBeVisible()
})

// #6 — Thiếu danh mục (CATEGORY_REQUIRED)
test('#6 thiếu danh mục bị chặn', async ({ page }) => {
  await page.goto('/transactions/new')
  await page.getByTestId('type-expense').click()
  await page.getByTestId('amount-input').fill('10000')
  await page.getByTestId('save-transaction').click()
  await expect(page.getByTestId('error-category')).toBeVisible()
  await expect(page).toHaveURL(/\/transactions\/new$/)
})

// #9 — Chặn ngày tương lai: date input max = hôm nay (FR-005, D15)
test('#9 ô ngày không cho chọn tương lai (max = hôm nay)', async ({ page }) => {
  await page.goto('/transactions/new')
  const max = await page.getByTestId('date-input').getAttribute('max')
  expect(max).toBe(new Date().toISOString().slice(0, 10))
})

// #10 — Lọc danh mục theo loại
test('#10 đổi loại → picker chỉ hiện danh mục cùng loại', async ({ page }) => {
  await page.goto('/transactions/new')
  await page.getByTestId('type-income').click()
  await expect(page.getByTestId('category-option-Lương')).toBeVisible()
  await expect(page.getByTestId('category-option-Ăn uống')).toHaveCount(0)
})

// #14 — Sửa số tiền → số dư khớp (UC-TRK-04 AC-1)
test('#14 sửa 50.000 → 80.000, số dư điều chỉnh đúng 30.000', async ({ page }) => {
  await createTransaction(page, { amount: 50000, category: 'Ăn uống', description: uniqueName('sửa-tiền') })
  const before = await readBalance(page, 'Tiền mặt')
  await page.getByTestId('transaction-row').first().click()
  await expect(page.getByRole('heading', { name: 'Sửa giao dịch' })).toBeVisible()
  await page.getByTestId('amount-input').fill('80000')
  await page.getByTestId('save-transaction').click()
  await expect(page).toHaveURL(/\/ledger$/)
  const after = await readBalance(page, 'Tiền mặt')
  expect(before - after).toBe(30000) // chi tăng thêm 30.000 → số dư giảm thêm 30.000
})

// #15 — Đổi loại buộc chọn lại danh mục (D18)
test('#15 sửa đổi Chi → Thu làm danh mục cũ mất hiệu lực, phải chọn lại', async ({ page }) => {
  await createTransaction(page, { amount: 40000, category: 'Ăn uống', description: uniqueName('đổi-loại') })
  await page.getByTestId('transaction-row').first().click()
  await expect(page.getByRole('heading', { name: 'Sửa giao dịch' })).toBeVisible()
  await page.getByTestId('type-income').click()
  // danh mục Chi cũ không còn được chọn → lưu bị chặn tới khi chọn danh mục Thu
  await page.getByTestId('save-transaction').click()
  await expect(page.getByTestId('error-category')).toBeVisible()
  await page.getByTestId('category-option-Lương').click()
  await page.getByTestId('save-transaction').click()
  await expect(page).toHaveURL(/\/ledger$/)
})

// #19/#20/#21 — Xóa có xác nhận; hủy; số dư hoàn tác (UC-TRK-05)
test('#19–21 xóa qua dialog xác nhận; hủy không đổi; xác nhận thì hoàn số dư', async ({ page }) => {
  await createTransaction(page, { amount: 50000, category: 'Ăn uống', description: uniqueName('sắp-xóa') })
  const withTxn = await readBalance(page, 'Tiền mặt')
  const row = page.getByTestId('transaction-row').first()
  const delId = (await row.getByRole('button').getAttribute('data-testid'))!

  // #19 mở dialog + #20 hủy → không đổi
  await page.getByTestId(delId).click()
  await expect(page.getByTestId('delete-transaction-dialog')).toBeVisible()
  await page.getByTestId('cancel-delete').click()
  await expect(page.getByTestId('delete-transaction-dialog')).toHaveCount(0)
  expect(await readBalance(page, 'Tiền mặt')).toBe(withTxn)

  // #21 xác nhận xóa → số dư +50.000
  await page.getByTestId(delId).click()
  await page.getByTestId('confirm-delete-transaction').click()
  await expect(page.getByTestId('delete-transaction-dialog')).toHaveCount(0)
  const after = await readBalance(page, 'Tiền mặt')
  expect(after - withTxn).toBe(50000)
})

// #22 — Số dư luôn khớp: đối chiếu API sau chuỗi thao tác (SC-004, D14)
test('#22 số dư từ GET /api/accounts khớp tổng bút toán sau chuỗi thao tác', async ({ page }) => {
  await createTransaction(page, { amount: 30000, category: 'Ăn uống' })
  await createTransaction(page, { amount: 20000, type: 'INCOME', category: 'Lương' })

  const accRes = await page.request.get('/api/accounts')
  const acc = (await accRes.json()).data.find((a: { name: string }) => a.name === 'Tiền mặt')
  // Duyệt HẾT trang (API chặn page_size ở 100) để tổng đúng kể cả khi hộ có > 100 giao dịch.
  const txns: Array<{ amount: number; type: string; account_id: string }> = []
  for (let p = 1; ; p++) {
    const res = await page.request.get(`/api/transactions?page=${p}&page_size=100`)
    const body = await res.json()
    txns.push(...body.data)
    if (txns.length >= body.paging.total || body.data.length === 0) break
  }
  const expected = txns
    .filter((t) => t.account_id === acc.id)
    .reduce((s, t) => s + (t.type === 'INCOME' ? t.amount : -t.amount), 0)
  expect(acc.balance).toBe(expected) // view = tổng bút toán có dấu
})

// #24 — Filter tháng/năm ở màn Giao dịch (feature bổ sung 2026-07)
test('#24 filter tháng/năm: đổi năm ẩn giao dịch ngoài khoảng, đổi lại hiện lại', async ({ page }) => {
  const d = uniqueName('filter')
  await createTransaction(page, { amount: 77000, category: 'Ăn uống', description: d }) // tháng hiện tại → /ledger
  await expect(page.getByTestId('txn-filter')).toBeVisible()
  await expect(page.getByTestId('transaction-row').filter({ hasText: d })).toBeVisible()

  // Đổi sang năm trước → giao dịch (năm nay) không nằm trong khoảng
  const now = new Date()
  await page.getByTestId('filter-year').selectOption(String(now.getFullYear() - 1))
  await expect(page.getByTestId('transaction-row').filter({ hasText: d })).toHaveCount(0)

  // Đổi về năm nay → hiển thị lại
  await page.getByTestId('filter-year').selectOption(String(now.getFullYear()))
  await expect(page.getByTestId('transaction-row').filter({ hasText: d })).toBeVisible()
})

// #25 — Ô số tiền tự format có dấu ngăn cách hàng nghìn khi gõ
test('#25 ô số tiền tự format khi gõ (3500000 → 3.500.000)', async ({ page }) => {
  await page.goto('/transactions/new')
  const input = page.getByTestId('amount-input')
  await input.fill('3500000')
  await expect(input).toHaveValue('3.500.000')
  await input.fill('50000')
  await expect(input).toHaveValue('50.000')
})

// #26 — Màn tạo giao dịch mặc định ngày = hôm nay
test('#26 màn tạo mặc định điền sẵn ngày hôm nay', async ({ page }) => {
  await page.goto('/transactions/new')
  await expect(page.getByTestId('date-input')).toHaveValue(new Date().toISOString().slice(0, 10))
})

// Đọc số dư theo SỰ THẬT của server (GET /api/accounts) — tất định, không lệ thuộc thời
// điểm refetch bất đồng bộ của AccountBalanceChip qua WebSocket (tránh flaky). Vẫn kiểm
// đúng "số dư điều chỉnh đúng" sau tạo/sửa/xóa; view số dư = tổng bút toán có dấu (D14).
async function readBalance(page: import('@playwright/test').Page, name: string): Promise<number> {
  const res = await page.request.get('/api/accounts')
  const accounts = (await res.json()).data as Array<{ name: string; balance: number }>
  const acc = accounts.find((a) => a.name === name)
  return acc ? acc.balance : 0
}
