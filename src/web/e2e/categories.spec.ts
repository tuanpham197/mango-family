import { expect, test } from '@playwright/test'
import { USERS, createCategory, login, selectByText, uniqueName } from './helpers'

test.beforeEach(async ({ page }) => {
  await login(page, USERS.alice.email)
})

// #1 — Tạo & dùng danh mục mới (AC1, UC-CAT-02)
test('#1 tạo danh mục Chi rồi chọn được khi nhập giao dịch', async ({ page }) => {
  const name = uniqueName('Thú cưng')
  await createCategory(page, { name, type: 'EXPENSE' })
  await expect(page.getByText(name)).toBeVisible() // xuất hiện ở màn quản lý

  await page.goto('/transactions/new')
  await page.getByTestId('type-expense').click()
  await expect(page.getByTestId(`category-option-${name}`)).toBeVisible()
})

// #2 — Loại là bắt buộc (FR-004)
test('#2 tạo danh mục bỏ trống loại bị chặn', async ({ page }) => {
  await page.goto('/categories/new')
  await page.getByTestId('category-name').fill(uniqueName('KhôngLoại'))
  await page.getByTestId('save-category').click()
  await expect(page.getByTestId('error-type')).toBeVisible()
})

// #3 — Cảnh báo trùng tên (FR-017)
test('#3 trùng tên cảnh báo nhưng cho xác nhận tiếp tục', async ({ page }) => {
  const name = uniqueName('Trùng')
  await createCategory(page, { name, type: 'EXPENSE' })

  // tạo lại cùng tên + loại → cảnh báo
  await page.goto('/categories/new')
  await page.getByTestId('type-expense').click()
  await page.getByTestId('category-name').fill(name)
  await page.getByTestId('save-category').click()
  await expect(page.getByTestId('form-error')).toContainText(/danh mục cùng tên/i)

  // bấm Lưu lần nữa → vẫn tạo (xác nhận)
  await page.getByTestId('save-category').click()
  await expect(page).toHaveURL(/\/categories$/)
})

// #7 — Danh mục con một cấp (FR-010/011)
test('#7 con kế thừa loại cha; không cho tạo cấp con thứ hai', async ({ page }) => {
  const parent = uniqueName('Cha')
  const child = uniqueName('Con')
  await createCategory(page, { name: parent, type: 'EXPENSE' })
  await createCategory(page, { name: child, type: 'EXPENSE', parent })

  // con hiển thị dưới cha ở màn quản lý, mang loại Chi
  await page.goto('/categories')
  await expect(page.getByText(child)).toBeVisible()

  // con KHÔNG xuất hiện làm lựa chọn cha (chặn cấp thứ hai — NESTING_TOO_DEEP)
  await page.goto('/categories/new')
  await page.getByTestId('type-expense').click()
  const parentOptions = await page.getByTestId('category-parent').locator('option').allTextContents()
  expect(parentOptions.some((o) => o.includes(parent))).toBeTruthy()
  expect(parentOptions.some((o) => o.includes(child))).toBeFalsy()
})

// #8 — Xóa an toàn, gán lại (FR-008/009/012, SC-007, D12)
test('#8 xóa danh mục còn giao dịch → gán lại sang danh mục cùng loại', async ({ page }) => {
  const victim = uniqueName('SắpXóa')
  const target = uniqueName('Đích')
  await createCategory(page, { name: victim, type: 'EXPENSE' })
  await createCategory(page, { name: target, type: 'EXPENSE' })

  // thêm 1 giao dịch vào victim
  await page.goto('/transactions/new')
  await page.getByTestId('type-expense').click()
  await page.getByTestId('amount-input').fill('75000')
  await page.getByTestId(`category-option-${victim}`).click()
  await page.getByTestId('save-transaction').click()
  await expect(page).toHaveURL(/\/ledger$/)

  // xóa victim → dialog báo còn giao dịch → chọn gán lại
  await page.goto('/categories')
  await page.getByTestId(`delete-${victim}`).click()
  await page.getByTestId('confirm-delete').click()
  await expect(page.getByText(/còn 1 giao dịch/i)).toBeVisible()
  await page.getByTestId('mode-reassign').check()
  await selectByText(page, 'reassign-target', target)
  await page.getByTestId('confirm-delete-mode').click()

  // victim biến mất; giao dịch chuyển sang target
  await expect(page.getByText(victim)).toHaveCount(0)
  await page.goto('/ledger')
  await expect(page.getByTestId('ledger').getByText(target).first()).toBeVisible()
})

// #9 — Gán lại sai loại bị chặn (FR-009) — dropdown chỉ chào đích cùng loại
test('#9 dialog gán lại chỉ liệt kê danh mục cùng loại', async ({ page }) => {
  const victim = uniqueName('XóaChi')
  await createCategory(page, { name: victim, type: 'EXPENSE' })
  await page.goto('/transactions/new')
  await page.getByTestId('type-expense').click()
  await page.getByTestId('amount-input').fill('12000')
  await page.getByTestId(`category-option-${victim}`).click()
  await page.getByTestId('save-transaction').click()
  await expect(page).toHaveURL(/\/ledger$/)

  await page.goto('/categories')
  await page.getByTestId(`delete-${victim}`).click()
  await page.getByTestId('confirm-delete').click()
  await page.getByTestId('mode-reassign').check()
  const options = await page.getByTestId('reassign-target').locator('option').allTextContents()
  // "Lương" là danh mục Thu mặc định — KHÔNG được xuất hiện trong đích Chi
  expect(options.some((o) => o.includes('Lương'))).toBeFalsy()
})

// #10 — Ẩn / bỏ ẩn (FR-020)
test('#10 ẩn danh mục thì không chọn được khi nhập; bỏ ẩn thì hiện lại', async ({ page }) => {
  const name = uniqueName('ẨnThử')
  await createCategory(page, { name, type: 'EXPENSE' })

  await page.goto('/categories')
  await page.getByTestId(`toggle-hidden-${name}`).click()
  await expect(page.getByTestId(`category-row`).filter({ hasText: name }).getByText('Đã ẩn')).toBeVisible()

  // không xuất hiện khi nhập giao dịch
  await page.goto('/transactions/new')
  await page.getByTestId('type-expense').click()
  await expect(page.getByTestId(`category-option-${name}`)).toHaveCount(0)

  // bỏ ẩn → hiện lại
  await page.goto('/categories')
  await page.getByTestId(`toggle-hidden-${name}`).click()
  await page.goto('/transactions/new')
  await page.getByTestId('type-expense').click()
  await expect(page.getByTestId(`category-option-${name}`)).toBeVisible()
})

// #11 — Đổi tên phản ánh mọi nơi (FR-016)
test('#11 đổi tên danh mục → giao dịch lịch sử hiện tên mới', async ({ page }) => {
  const original = uniqueName('TênCũ')
  const renamed = uniqueName('TênMới')
  await createCategory(page, { name: original, type: 'EXPENSE' })

  await page.goto('/transactions/new')
  await page.getByTestId('type-expense').click()
  await page.getByTestId('amount-input').fill('33000')
  await page.getByTestId(`category-option-${original}`).click()
  await page.getByTestId('save-transaction').click()
  await expect(page).toHaveURL(/\/ledger$/)

  // đổi tên
  await page.goto('/categories')
  await page.getByTestId(`edit-${original}`).click()
  await page.getByTestId('category-name').fill(renamed)
  await page.getByTestId('save-category').click()

  // giao dịch lịch sử hiển thị tên mới (tham chiếu id, không sao chép)
  await page.goto('/ledger')
  await expect(page.getByTestId('ledger').getByText(renamed).first()).toBeVisible()
})
