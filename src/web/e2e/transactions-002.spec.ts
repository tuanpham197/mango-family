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
  await expect(page).toHaveURL(/\/$/)
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
  await expect(page).toHaveURL(/\/$/)
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
  const txRes = await page.request.get('/api/transactions?page_size=1000')
  const txns = (await txRes.json()).data as Array<{ amount: number; type: string; account_id: string }>
  const expected = txns
    .filter((t) => t.account_id === acc.id)
    .reduce((s, t) => s + (t.type === 'INCOME' ? t.amount : -t.amount), 0)
  expect(acc.balance).toBe(expected) // view = tổng bút toán có dấu
})

async function readBalance(page: import('@playwright/test').Page, name: string): Promise<number> {
  const text = (await page.getByTestId(`balance-${name}`).textContent()) ?? '0'
  // "−50.000 ₫" → -50000
  const negative = text.includes('−') || text.includes('-')
  const digits = Number(text.replace(/[^\d]/g, ''))
  return negative ? -digits : digits
}
