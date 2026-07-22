import { expect, test, type BrowserContext } from '@playwright/test'
import { USERS, createCategory, createTransaction, login, uniqueName } from './helpers'

// Báo cáo — cô lập hộ (quickstart #16).
test('#16 cô lập hộ: chi hộ A không xuất hiện trong báo cáo Carol (hộ B)', async ({ browser }) => {
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
    await carol.goto('/reports')
    await expect(carol.getByTestId('report-summary')).toBeVisible()
    await expect(carol.getByTestId(`breakdown-row-${cat}`)).toHaveCount(0)
  } finally {
    await a?.close()
    await c?.close()
  }
})
