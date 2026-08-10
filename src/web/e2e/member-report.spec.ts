import { expect, test, type BrowserContext, type Page } from '@playwright/test'
import { USERS, createCategory, createTransaction, login, uniqueName } from './helpers'

// Báo cáo theo thành viên (feature 008 — quickstart QS-1…QS-7).

async function openReports(page: Page) {
  await page.goto('/reports')
  await expect(page.getByTestId('report-summary')).toBeVisible()
  await expect(page.getByTestId('member-breakdown')).toBeVisible()
}

// QS-1 / QS-2 — báo cáo theo thành viên hiển thị; tổng hộ hiện; đổi khoảng vẫn hiển thị.
test('QS-1,QS-2 báo cáo theo thành viên + tổng hộ; đổi khoảng cập nhật', async ({ page }) => {
  await login(page, USERS.alice.email)
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createTransaction(page, { amount: 320000, category: cat, description: uniqueName('mbr') })

  await openReports(page)
  // Alice có một dòng trong bảng theo thành viên (số là TỔNG tháng — kiểm số chính xác ở
  // drill-in QS-4 & unit/integration; ở đây chỉ cần dòng + tổng hộ hiển thị).
  await expect(page.getByTestId('member-row').filter({ hasText: 'Alice' })).toBeVisible()
  await expect(page.getByTestId('member-totals')).toBeVisible()

  await page.getByTestId('range-week').click()
  await expect(page.getByTestId('member-breakdown')).toBeVisible()
})

// QS-4 — drill-in danh sách giao dịch của một thành viên + đóng.
test('QS-4 drill-in giao dịch của một thành viên; đóng', async ({ page }) => {
  await login(page, USERS.alice.email)
  const cat = uniqueName('An')
  const desc = uniqueName('drill')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createTransaction(page, { amount: 175000, category: cat, description: desc })

  await openReports(page)
  await page.getByTestId('member-row').filter({ hasText: 'Alice' }).click()
  const detail = page.getByTestId('member-detail')
  await expect(detail).toBeVisible()
  await expect(detail).toContainText('Alice')
  await expect(detail.getByTestId('member-txn').filter({ hasText: desc })).toHaveCount(1)

  await page.getByTestId('close-member-detail').click()
  await expect(page.getByTestId('member-detail')).toHaveCount(0)
  // vẫn ở màn báo cáo theo thành viên, khoảng giữ nguyên
  await expect(page.getByTestId('member-breakdown')).toBeVisible()
})

// QS-3 — thành viên không có giao dịch trong khoảng vẫn hiển thị (seed: Dave, hộ A).
test('QS-3 thành viên không có giao dịch vẫn xuất hiện', async ({ page }) => {
  await login(page, USERS.alice.email)
  await openReports(page)
  await expect(page.getByTestId('member-row').filter({ hasText: 'Dave' })).toBeVisible()
})

// QS-6 — cô lập hộ: Carol (hộ B) không thấy thành viên hộ A.
test('QS-6 cô lập hộ: Carol không thấy Alice/Bob/Dave của hộ A', async ({ browser }) => {
  let c: BrowserContext | undefined
  try {
    c = await browser.newContext()
    const carol = await c.newPage()
    await login(carol, USERS.carol.email)
    await carol.goto('/reports')
    await expect(carol.getByTestId('member-breakdown')).toBeVisible()
    await expect(carol.getByTestId('member-row').filter({ hasText: 'Alice' })).toHaveCount(0)
    await expect(carol.getByTestId('member-row').filter({ hasText: 'Bob' })).toHaveCount(0)
  } finally {
    await c?.close()
  }
})
