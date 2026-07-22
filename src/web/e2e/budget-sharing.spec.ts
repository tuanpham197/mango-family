import { expect, test, type BrowserContext, type Page } from '@playwright/test'
import { USERS, budgetRow, createBudget, createCategory, createTransaction, login, uniqueName } from './helpers'

// Ngân sách đa thành viên (hộ A: Alice + Bob) & cô lập hộ B (Carol) — quickstart #6, #11, #19, #25, #26, #29.

async function twoMembers(browser: import('@playwright/test').Browser) {
  const a = await browser.newContext()
  const b = await browser.newContext()
  const alice = await a.newPage()
  const bob = await b.newPage()
  await login(alice, USERS.alice.email)
  await login(bob, USERS.bob.email)
  return { a, b, alice, bob }
}

// #6 — dùng chung trong hộ, cô lập giữa hộ
test('#6 Bob (cùng hộ) thấy ngân sách Alice tạo; Carol (hộ khác) không; API hộ A → 404', async ({ browser }) => {
  const cat = uniqueName('An')
  let a: BrowserContext | undefined
  let c: BrowserContext | undefined
  try {
    a = await browser.newContext()
    c = await browser.newContext()
    const alice = await a.newPage()
    const carol = await c.newPage()
    await login(alice, USERS.alice.email)
    await createCategory(alice, { name: cat, type: 'EXPENSE' })
    await createBudget(alice, { category: cat, limit: 5000000, period: 'MONTHLY' })
    const res = await alice.request.get('/api/budgets')
    const bid = (await res.json()).data.find((x: { category_name: string }) => x.category_name === cat).id

    // Bob cùng hộ A thấy
    const bob = await (await browser.newContext()).newPage()
    await login(bob, USERS.bob.email)
    await bob.goto('/budgets')
    await expect(budgetRow(bob, cat)).toHaveCount(1)

    // Carol hộ B không thấy + API 404
    await login(carol, USERS.carol.email)
    await carol.goto('/budgets')
    await expect(carol.getByText(cat)).toHaveCount(0)
    const carolRes = await carol.request.get(`/api/budgets/${bid}`)
    expect(carolRes.status()).toBe(404)
  } finally {
    await a?.close()
    await c?.close()
  }
})

// #11 — realtime tiến độ giữa thành viên
test('#11 Bob nhập giao dịch → tiến độ trên màn Alice cập nhật ≤ 5s', async ({ browser }) => {
  const cat = uniqueName('An')
  let a: BrowserContext | undefined
  let b: BrowserContext | undefined
  try {
    const m = await twoMembers(browser)
    a = m.a
    b = m.b
    await createCategory(m.alice, { name: cat, type: 'EXPENSE' })
    await createBudget(m.alice, { category: cat, limit: 1000000, period: 'MONTHLY' })

    await m.alice.goto('/budgets') // Alice đang mở màn ngân sách
    await expect(budgetRow(m.alice, cat)).toContainText('0%')

    await createTransaction(m.bob, { amount: 500000, category: cat, description: uniqueName('bob') })
    await expect(budgetRow(m.alice, cat)).toContainText('50%', { timeout: 8000 }) // realtime
  } finally {
    await a?.close()
    await b?.close()
  }
})

// #19 — cảnh báo tới mọi thành viên
test('#19 Bob gây vượt ngưỡng → Alice (đang mở app) cũng thấy cảnh báo ≤ 5s', async ({ browser }) => {
  const cat = uniqueName('An')
  let a: BrowserContext | undefined
  let b: BrowserContext | undefined
  try {
    const m = await twoMembers(browser)
    a = m.a
    b = m.b
    await createCategory(m.alice, { name: cat, type: 'EXPENSE' })
    await createBudget(m.alice, { category: cat, limit: 1000000, period: 'MONTHLY' })
    await m.alice.goto('/budgets')

    await createTransaction(m.bob, { amount: 900000, category: cat, description: uniqueName('bob') })
    await expect(budgetRow(m.alice, cat).getByTestId('alert-threshold')).toBeVisible({ timeout: 8000 })
  } finally {
    await a?.close()
    await b?.close()
  }
})

// #25 — ngang quyền: Bob sửa ngân sách Alice tạo
test('#25 Bob sửa được ngân sách do Alice tạo', async ({ browser }) => {
  const cat = uniqueName('An')
  let a: BrowserContext | undefined
  let b: BrowserContext | undefined
  try {
    const m = await twoMembers(browser)
    a = m.a
    b = m.b
    await createCategory(m.alice, { name: cat, type: 'EXPENSE' })
    await createBudget(m.alice, { category: cat, limit: 5000000, period: 'MONTHLY' })

    await m.bob.goto('/budgets')
    await budgetRow(m.bob, cat).click()
    await expect(m.bob.getByRole('heading', { name: 'Sửa ngân sách' })).toBeVisible()
    await m.bob.getByTestId('limit-input').fill('6000000')
    await m.bob.getByTestId('save-budget').click()
    await expect(m.bob).toHaveURL(/\/budgets$/) // thành công — không cổng quyền
  } finally {
    await a?.close()
    await b?.close()
  }
})

// #26 — không ghi đè thầm lặng
test('#26 hai thành viên cùng sửa → người sau nhận CONCURRENCY_CONFLICT', async ({ browser }) => {
  const cat = uniqueName('An')
  let a: BrowserContext | undefined
  let b: BrowserContext | undefined
  try {
    const m = await twoMembers(browser)
    a = m.a
    b = m.b
    await createCategory(m.alice, { name: cat, type: 'EXPENSE' })
    await createBudget(m.alice, { category: cat, limit: 5000000, period: 'MONTHLY' })

    await openEdit(m.alice, cat)
    await openEdit(m.bob, cat)

    await m.alice.getByTestId('limit-input').fill('4000000')
    await m.alice.getByTestId('save-budget').click()
    await expect(m.alice).toHaveURL(/\/budgets$/)

    await m.bob.getByTestId('limit-input').fill('9000000')
    await m.bob.getByTestId('save-budget').click()
    await expect(m.bob.getByTestId('form-error')).toContainText(/thành viên khác/i)
  } finally {
    await a?.close()
    await b?.close()
  }
})

// #29 — xóa ngân sách khi thành viên khác đang mở form
test('#29 Bob xóa trong khi Alice đang sửa → Alice lưu nhận RECORD_GONE', async ({ browser }) => {
  const cat = uniqueName('An')
  let a: BrowserContext | undefined
  let b: BrowserContext | undefined
  try {
    const m = await twoMembers(browser)
    a = m.a
    b = m.b
    await createCategory(m.alice, { name: cat, type: 'EXPENSE' })
    await createBudget(m.alice, { category: cat, limit: 5000000, period: 'MONTHLY' })

    await openEdit(m.alice, cat) // Alice mở form sửa

    // Bob xóa ngân sách đó
    await m.bob.goto('/budgets')
    const delId = `delete-budget-${await budgetIdByLabel(m.bob, cat)}`
    await m.bob.getByTestId(delId).click()
    await m.bob.getByTestId('confirm-delete-budget').click()
    await expect(budgetRow(m.bob, cat)).toHaveCount(0)

    // Alice lưu → RECORD_GONE
    await m.alice.getByTestId('limit-input').fill('7000000')
    await m.alice.getByTestId('save-budget').click()
    await expect(m.alice.getByTestId('form-error')).toContainText(/đã bị xóa/i)
  } finally {
    await a?.close()
    await b?.close()
  }
})

async function openEdit(page: Page, label: string) {
  await page.goto('/budgets')
  await budgetRow(page, label).click()
  await expect(page.getByRole('heading', { name: 'Sửa ngân sách' })).toBeVisible()
}

async function budgetIdByLabel(page: Page, label: string): Promise<string> {
  const res = await page.request.get('/api/budgets?status=all')
  const items = (await res.json()).data as Array<{ id: string; category_name: string | null }>
  return items.find((x) => x.category_name === label)!.id
}
