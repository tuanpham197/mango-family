import { expect, test } from '@playwright/test'
import { DEV_PASSWORD, USERS, createCategory, login, uniqueName } from './helpers'

// #0 — Đăng nhập nền tảng (UC-TRK-01, D4)
test('#0 đăng nhập đúng/sai + chặn khi chưa đăng nhập', async ({ page }) => {
  // chưa đăng nhập → về /login
  await page.goto('/')
  await expect(page).toHaveURL(/\/login/)

  // sai mật khẩu → từ chối an toàn
  await page.getByTestId('login-email').fill(USERS.alice.email)
  await page.getByTestId('login-password').fill('sai-mat-khau')
  await page.getByTestId('login-submit').click()
  await expect(page.getByTestId('login-error')).toBeVisible()

  // đúng → vào hộ A, thấy tên Alice
  await page.getByTestId('login-password').fill(DEV_PASSWORD)
  await page.getByTestId('login-submit').click()
  await expect(page.getByTestId('household-name')).toHaveText('Gia đình A')
  await expect(page.getByTestId('current-user')).toHaveText(USERS.alice.name)
})

// #15 — Cô lập giữa các hộ (FR-018, D5)
test('#15 Carol (hộ B) không thấy dữ liệu hộ A; gọi thẳng API id hộ A → 404', async ({ browser }) => {
  const catName = uniqueName('RiêngHộA')

  // Alice (hộ A) tạo danh mục và lấy id qua API
  const aliceCtx = await browser.newContext()
  const alicePage = await aliceCtx.newPage()
  await login(alicePage, USERS.alice.email)
  await createCategory(alicePage, { name: catName, type: 'EXPENSE' })
  const list = await alicePage.request.get('/api/categories?include_hidden=true')
  const cats = (await list.json()).data as Array<{ id: string; name: string; updated_at: string }>
  const target = cats.find((c) => c.name === catName)!
  expect(target).toBeTruthy()

  // Carol (hộ B) không thấy trong UI
  const carolCtx = await browser.newContext()
  const carolPage = await carolCtx.newPage()
  await login(carolPage, USERS.carol.email)
  await expect(carolPage.getByTestId('household-name')).toHaveText('Gia đình B')
  await carolPage.goto('/categories')
  await expect(carolPage.getByText(catName)).toHaveCount(0)

  // Gọi thẳng API bằng id của hộ A → 404 (không lộ tồn tại)
  const res = await carolPage.request.patch(`/api/categories/${target.id}`, {
    data: { name: 'Hack', expected_updated_at: target.updated_at },
  })
  expect(res.status()).toBe(404)

  await aliceCtx.close()
  await carolCtx.close()
})

// #17 — Quyền ngang nhau (FR-021)
test('#17 Bob sửa được danh mục do Alice tạo (không cổng quyền theo vai trò)', async ({ browser }) => {
  const original = uniqueName('AliceTạo')
  const renamed = uniqueName('BobSửa')

  const aliceCtx = await browser.newContext()
  const alicePage = await aliceCtx.newPage()
  await login(alicePage, USERS.alice.email)
  await createCategory(alicePage, { name: original, type: 'EXPENSE' })

  const bobCtx = await browser.newContext()
  const bobPage = await bobCtx.newPage()
  await login(bobPage, USERS.bob.email)
  await bobPage.goto('/categories')
  await expect(bobPage.getByText(original)).toBeVisible()
  await bobPage.getByTestId(`edit-${original}`).click()
  await bobPage.getByTestId('category-name').fill(renamed)
  await bobPage.getByTestId('save-category').click()
  await expect(bobPage.getByText(renamed)).toBeVisible() // thành công, không bị chặn quyền

  await aliceCtx.close()
  await bobCtx.close()
})
