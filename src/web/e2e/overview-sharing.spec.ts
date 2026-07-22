import { expect, test, type BrowserContext } from '@playwright/test'
import { USERS, createCategory, createTransaction, login, uniqueName } from './helpers'

// Màn Tổng quan đa thành viên (quickstart #25 realtime, #26 cô lập hộ).

// #25 — Bob nhập giao dịch → màn Tổng quan Alice cập nhật ≤ 5s
test('#25 realtime: Bob nhập chi → chi tiêu theo danh mục trên Tổng quan Alice cập nhật', async ({ browser }) => {
  const cat = uniqueName('An')
  let a: BrowserContext | undefined
  let b: BrowserContext | undefined
  try {
    a = await browser.newContext()
    b = await browser.newContext()
    const alice = await a.newPage()
    const bob = await b.newPage()
    await login(alice, USERS.alice.email)
    await login(bob, USERS.bob.email)
    await createCategory(alice, { name: cat, type: 'EXPENSE' })

    await alice.goto('/') // Alice mở màn Tổng quan
    await expect(alice.getByTestId('networth-card')).toBeVisible()

    await createTransaction(bob, { amount: 321000, category: cat, description: uniqueName('bob') })
    // Realtime (D33): chi theo danh mục của Alice hiện danh mục Bob vừa chi
    await expect(alice.getByTestId('category-spending-row').filter({ hasText: cat })).toBeVisible({ timeout: 8000 })
  } finally {
    await a?.close()
    await b?.close()
  }
})

// #26 — cô lập hộ: Carol (hộ B) không thấy dữ liệu hộ A
test('#26 cô lập hộ: chi tiêu hộ A không xuất hiện trên Tổng quan Carol', async ({ browser }) => {
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
    await createTransaction(alice, { amount: 456000, category: cat, description: uniqueName('alice') })

    await login(carol, USERS.carol.email)
    await carol.goto('/')
    await expect(carol.getByTestId('networth-card')).toBeVisible()
    await expect(carol.getByTestId('category-spending-row').filter({ hasText: cat })).toHaveCount(0)
  } finally {
    await a?.close()
    await c?.close()
  }
})
