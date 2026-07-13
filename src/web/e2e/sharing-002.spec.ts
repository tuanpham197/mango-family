import { expect, test, type BrowserContext, type Page } from '@playwright/test'
import { USERS, createTransaction, login, uniqueName } from './helpers'

// Kịch bản đa thành viên trong cùng hộ A (Alice + Bob) và cô lập hộ B (Carol).

// #11 — Sổ chung + tên người nhập + realtime ≤ 5s (UC-TRK-03 AC-1, SC-006)
test('#11 Alice nhập → Bob thấy trong ≤ 5s kèm "do Alice nhập"', async ({ browser }) => {
  const desc = uniqueName('chung')
  let a: BrowserContext | undefined
  let b: BrowserContext | undefined
  try {
    a = await browser.newContext()
    b = await browser.newContext()
    const alice = await a.newPage()
    const bob = await b.newPage()
    await login(alice, USERS.alice.email)
    await login(bob, USERS.bob.email)

    await bob.goto('/') // đang mở sổ, lắng nghe transactions_changed
    await createTransaction(alice, { amount: 60000, category: 'Ăn uống', description: desc })

    const row = bob.getByTestId('transaction-row').filter({ hasText: desc })
    await expect(row.first()).toContainText('do Alice nhập') // tên, không phải mã
  } finally {
    await a?.close()
    await b?.close()
  }
})

// #13 — Cô lập hộ: Carol không thấy giao dịch hộ A; API id hộ A → 404
test('#13 Carol (hộ B) không thấy giao dịch hộ A; gọi API id hộ A → 404', async ({ browser }) => {
  const desc = uniqueName('bí-mật-A')
  let a: BrowserContext | undefined
  let c: BrowserContext | undefined
  try {
    a = await browser.newContext()
    c = await browser.newContext()
    const alice = await a.newPage()
    const carol = await c.newPage()
    await login(alice, USERS.alice.email)
    await createTransaction(alice, { amount: 12345, category: 'Ăn uống', description: desc })
    const list = await alice.request.get('/api/transactions?page_size=1')
    const txnId = (await list.json()).data[0].id

    await login(carol, USERS.carol.email)
    await carol.goto('/')
    await expect(carol.getByText(desc)).toHaveCount(0)
    const res = await carol.request.get(`/api/transactions/${txnId}`)
    expect(res.status()).toBe(404)
  } finally {
    await a?.close()
    await c?.close()
  }
})

// #16 — Ngang quyền sửa: Bob sửa giao dịch Alice nhập
test('#16 Bob sửa được giao dịch do Alice nhập', async ({ browser }) => {
  const desc = uniqueName('của-Alice')
  let a: BrowserContext | undefined
  let b: BrowserContext | undefined
  try {
    a = await browser.newContext()
    b = await browser.newContext()
    const alice = await a.newPage()
    const bob = await b.newPage()
    await login(alice, USERS.alice.email)
    await createTransaction(alice, { amount: 70000, category: 'Ăn uống', description: desc })

    await login(bob, USERS.bob.email)
    await bob.goto('/')
    await bob.getByTestId('transaction-row').filter({ hasText: desc }).first().click()
    await expect(bob.getByRole('heading', { name: 'Sửa giao dịch' })).toBeVisible()
    await bob.getByTestId('amount-input').fill('75000')
    await bob.getByTestId('save-transaction').click()
    await expect(bob).toHaveURL(/\/$/) // thành công — không cổng quyền
  } finally {
    await a?.close()
    await b?.close()
  }
})

// #17 — Không ghi đè thầm lặng: người sau nhận CONCURRENCY_CONFLICT (D17)
test('#17 hai thành viên cùng sửa → người sau nhận xung đột, thấy dữ liệu mới', async ({ browser }) => {
  const desc = uniqueName('đồng-thời')
  let a: BrowserContext | undefined
  let b: BrowserContext | undefined
  try {
    a = await browser.newContext()
    b = await browser.newContext()
    const alice = await a.newPage()
    const bob = await b.newPage()
    await login(alice, USERS.alice.email)
    await createTransaction(alice, { amount: 50000, category: 'Ăn uống', description: desc })

    await login(bob, USERS.bob.email)
    // Cả hai mở form sửa cùng giao dịch (cùng expected_updated_at)
    await openEditByDesc(alice, desc)
    await openEditByDesc(bob, desc)

    // Alice lưu trước → thành công
    await alice.getByTestId('amount-input').fill('55000')
    await alice.getByTestId('save-transaction').click()
    await expect(alice).toHaveURL(/\/$/)

    // Bob lưu sau với mốc cũ → xung đột, không ghi đè
    await bob.getByTestId('amount-input').fill('99000')
    await bob.getByTestId('save-transaction').click()
    await expect(bob.getByTestId('form-error')).toContainText(/thành viên khác/i)
  } finally {
    await a?.close()
    await b?.close()
  }
})

// #18 — Sửa trong lúc bị xóa → RECORD_GONE (UC-TRK-04 E3)
test('#18 Bob xóa trong khi Alice đang sửa → Alice lưu nhận "không còn tồn tại"', async ({ browser }) => {
  const desc = uniqueName('sẽ-bị-xóa')
  let a: BrowserContext | undefined
  let b: BrowserContext | undefined
  try {
    a = await browser.newContext()
    b = await browser.newContext()
    const alice = await a.newPage()
    const bob = await b.newPage()
    await login(alice, USERS.alice.email)
    await createTransaction(alice, { amount: 33000, category: 'Ăn uống', description: desc })

    await login(bob, USERS.bob.email)
    await openEditByDesc(alice, desc) // Alice mở form sửa

    // Bob xóa giao dịch đó
    await bob.goto('/')
    const row = bob.getByTestId('transaction-row').filter({ hasText: desc }).first()
    const delId = (await row.getByRole('button').getAttribute('data-testid'))!
    await bob.getByTestId(delId).click()
    await bob.getByTestId('confirm-delete-transaction').click()
    await expect(bob.getByTestId('delete-transaction-dialog')).toHaveCount(0)

    // Alice lưu → RECORD_GONE
    await alice.getByTestId('amount-input').fill('34000')
    await alice.getByTestId('save-transaction').click()
    await expect(alice.getByTestId('form-error')).toContainText(/đã bị xóa/i)
  } finally {
    await a?.close()
    await b?.close()
  }
})

// #23 — Mất kết nối khi lưu → thử lại chỉ tạo MỘT giao dịch (Edge case)
test('#23 offline khi lưu, bật lại mạng thử lại → chỉ một giao dịch', async ({ browser }) => {
  const desc = uniqueName('offline')
  const ctx = await browser.newContext()
  try {
    const page = await ctx.newPage()
    await login(page, USERS.alice.email)

    await page.goto('/transactions/new')
    await page.getByTestId('type-expense').click()
    await page.getByTestId('amount-input').fill('25000')
    await page.getByTestId('description-input').fill(desc)
    await page.getByTestId('category-option-Ăn uống').click()

    await ctx.setOffline(true)
    await page.getByTestId('save-transaction').click()
    await expect(page.getByTestId('form-error')).toBeVisible() // báo lỗi mạng rõ ràng
    await expect(page).toHaveURL(/\/transactions\/new$/)

    await ctx.setOffline(false)
    await page.getByTestId('save-transaction').click()
    await expect(page).toHaveURL(/\/$/)

    // Chỉ đúng một giao dịch có mô tả này
    const res = await page.request.get('/api/transactions?page_size=1000')
    const matches = (await res.json()).data.filter((t: { description: string }) => t.description === desc)
    expect(matches).toHaveLength(1)
  } finally {
    await ctx.close()
  }
})

async function openEditByDesc(page: Page, desc: string) {
  await page.goto('/')
  await page.getByTestId('transaction-row').filter({ hasText: desc }).first().click()
  await expect(page.getByRole('heading', { name: 'Sửa giao dịch' })).toBeVisible()
}
