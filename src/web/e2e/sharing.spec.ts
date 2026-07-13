import { expect, test, type BrowserContext } from '@playwright/test'
import { USERS, createCategory, login, uniqueName } from './helpers'

// Các kịch bản đa thành viên trong CÙNG hộ (Alice + Bob → hộ A): dùng 2 context
// Playwright riêng để mô phỏng 2 phiên đồng thời.

// #13 — Chia sẻ danh mục trong hộ, đồng bộ ≤ 5s (FR-018, D8)
test('#13 Alice tạo danh mục → Bob thấy trong ≤ 5s (WS invalidation)', async ({ browser }) => {
  const name = uniqueName('ChiaSẻ')
  let aliceCtx: BrowserContext | undefined
  let bobCtx: BrowserContext | undefined
  try {
    aliceCtx = await browser.newContext()
    bobCtx = await browser.newContext()
    const alicePage = await aliceCtx.newPage()
    const bobPage = await bobCtx.newPage()
    await login(alicePage, USERS.alice.email)
    await login(bobPage, USERS.bob.email)

    // Bob mở màn nhập giao dịch Chi (đang lắng nghe categories_changed)
    await bobPage.goto('/transactions/new')
    await bobPage.getByTestId('type-expense').click()
    await expect(bobPage.getByTestId(`category-option-${name}`)).toHaveCount(0)

    // Alice tạo danh mục mới
    await createCategory(alicePage, { name, type: 'EXPENSE' })

    // Bob thấy danh mục mới KHÔNG cần reload, trong ngưỡng đồng bộ (expect timeout 7s)
    await expect(bobPage.getByTestId(`category-option-${name}`)).toBeVisible()
  } finally {
    await aliceCtx?.close()
    await bobCtx?.close()
  }
})

// #14 — Sổ chung + authorship (FR-022)
test('#14 Alice nhập giao dịch → Bob thấy "do Alice nhập"', async ({ browser }) => {
  const cat = uniqueName('SổChung')
  const amount = String(100000 + Math.floor(Math.random() * 99999))
  let aliceCtx: BrowserContext | undefined
  let bobCtx: BrowserContext | undefined
  try {
    aliceCtx = await browser.newContext()
    bobCtx = await browser.newContext()
    const alicePage = await aliceCtx.newPage()
    const bobPage = await bobCtx.newPage()
    await login(alicePage, USERS.alice.email)
    await login(bobPage, USERS.bob.email)

    await createCategory(alicePage, { name: cat, type: 'EXPENSE' })
    await alicePage.goto('/transactions/new')
    await alicePage.getByTestId('type-expense').click()
    await alicePage.getByTestId('amount-input').fill(amount)
    await alicePage.getByTestId(`category-option-${cat}`).click()
    await alicePage.getByTestId('save-transaction').click()
    await expect(alicePage).toHaveURL(/\/$/)

    // Bob mở sổ → thấy giao dịch của Alice kèm authorship (WS refetch)
    await bobPage.goto('/')
    const row = bobPage.getByTestId('transaction-row').filter({ hasText: cat })
    await expect(row.first()).toContainText('do Alice nhập')
  } finally {
    await aliceCtx?.close()
    await bobCtx?.close()
  }
})

// #16 — Đồng thời: không ghi đè thầm lặng (D6, SC-007)
test('#16 hai thành viên sửa cùng danh mục → người sau nhận CONCURRENCY_CONFLICT', async ({ browser }) => {
  const original = uniqueName('ĐồngThời')
  let aliceCtx: BrowserContext | undefined
  let bobCtx: BrowserContext | undefined
  try {
    aliceCtx = await browser.newContext()
    bobCtx = await browser.newContext()
    const alicePage = await aliceCtx.newPage()
    const bobPage = await bobCtx.newPage()
    await login(alicePage, USERS.alice.email)
    await createCategory(alicePage, { name: original, type: 'EXPENSE' })

    await login(bobPage, USERS.bob.email)

    // Cả hai cùng mở form sửa (đều nắm cùng expected_updated_at)
    await alicePage.goto('/categories')
    await alicePage.getByTestId(`edit-${original}`).click()
    await bobPage.goto('/categories')
    await bobPage.getByTestId(`edit-${original}`).click()

    // Alice lưu trước → thành công
    await alicePage.getByTestId('category-name').fill(uniqueName('AliceThắng'))
    await alicePage.getByTestId('save-category').click()
    await expect(alicePage).toHaveURL(/\/categories$/)

    // Bob lưu sau với mốc cũ → nhận xung đột, không ghi đè thầm lặng
    await bobPage.getByTestId('category-name').fill(uniqueName('BobSau'))
    await bobPage.getByTestId('save-category').click()
    await expect(bobPage.getByTestId('form-error')).toContainText(/thành viên khác/i)
  } finally {
    await aliceCtx?.close()
    await bobCtx?.close()
  }
})
