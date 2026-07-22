import { expect, test, type Page } from '@playwright/test'
import { USERS, createCategory, createTransaction, login, uniqueName } from './helpers'

// Báo cáo (feature 005) — kịch bản một ngữ cảnh (quickstart #1–15). Dùng danh mục riêng
// (uniqueName) cho khẳng định cần cô lập khỏi dữ liệu tích luỹ của hộ.

test.beforeEach(async ({ page }) => {
  await login(page, USERS.alice.email)
})

async function openReports(page: Page) {
  await page.goto('/reports')
  await expect(page.getByTestId('report-summary')).toBeVisible()
}

// #1 / #2 — tổng quan + đổi khoảng
test('#1,#2 báo cáo tổng quan hiển thị Thu/Chi/Ròng; đổi khoảng cập nhật', async ({ page }) => {
  await openReports(page)
  await expect(page.getByTestId('report-income')).toBeVisible()
  await expect(page.getByTestId('report-expense')).toBeVisible()
  await expect(page.getByTestId('report-net')).toBeVisible()
  await page.getByTestId('range-week').click()
  await expect(page.getByTestId('report-summary')).toBeVisible() // vẫn hiển thị sau khi đổi khoảng
})

// #4 — khoảng tùy chỉnh end < start bị chặn
test('#4 khoảng tùy chỉnh end < start hiện lỗi', async ({ page }) => {
  await openReports(page)
  await page.getByTestId('range-custom').click()
  await page.getByTestId('custom-from').fill('2026-07-31')
  await page.getByTestId('custom-to').fill('2026-07-01')
  await page.getByTestId('custom-to').blur()
  await expect(page.getByTestId('range-error')).toBeVisible()
})

// #6 / #11 / #15 — phân bổ danh mục + drill-in chi tiết + xu hướng danh mục
test('#6,#11,#15 phân bổ theo danh mục → drill-in chi tiết + danh sách giao dịch', async ({ page }) => {
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createTransaction(page, { amount: 250000, category: cat, description: uniqueName('rpt') })

  await openReports(page)
  const row = page.getByTestId(`breakdown-row-${cat}`)
  await expect(row).toBeVisible()
  await expect(row).toContainText('250.000')

  await row.click()
  const detail = page.getByTestId('category-detail')
  await expect(detail).toBeVisible()
  await expect(detail).toContainText(cat)
  await expect(detail).toContainText('250.000')
  await expect(detail.getByTestId('detail-txn')).toHaveCount(1)
  // đóng chi tiết
  await page.getByTestId('close-category-detail').click()
  await expect(page.getByTestId('category-detail')).toHaveCount(0)
})

// #9 — biểu đồ xu hướng hiển thị
test('#9 khối xu hướng thu/chi hiển thị', async ({ page }) => {
  const cat = uniqueName('An')
  await createCategory(page, { name: cat, type: 'EXPENSE' })
  await createTransaction(page, { amount: 100000, category: cat, description: uniqueName('t') })
  await openReports(page)
  await expect(page.getByTestId('trend-chart-block')).toBeVisible()
  await expect(page.getByTestId('line-chart')).toBeVisible()
  await expect(page.getByTestId('donut-chart')).toBeVisible() // phân bổ dạng donut
})
