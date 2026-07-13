import { defineConfig, devices } from '@playwright/test'

// E2E theo quickstart.md (18 kịch bản #0–#17). Yêu cầu Go API chạy ở :8080
// (make api) + Postgres đã migrate & seed (make up && make migrate-up && make seed).
// Vite dev server (:5173) proxy /api + /ws → :8080; config tự khởi động nếu chưa chạy.
export default defineConfig({
  testDir: './e2e',
  fullyParallel: false, // dùng chung DB seed — chạy tuần tự cho ổn định
  workers: 1,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: [['list']],
  timeout: 30_000,
  expect: { timeout: 7_000 }, // ≥ ngưỡng đồng bộ 5s cho kịch bản WS (#13)
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'retain-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: true,
    timeout: 120_000,
  },
})
