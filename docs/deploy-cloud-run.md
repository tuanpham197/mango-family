# Deploy — API (Go) lên Google Cloud Run + DB Supabase + Web (Vue) trên Vercel

Kiến trúc **tách origin**: Web trên **Vercel** ↔ API trên **Cloud Run** ↔ Postgres trên
**Supabase**. Vì khác origin nên API bật **CORS** và cookie phiên dùng **SameSite=None; Secure**.

```
  Trình duyệt ── https ──▶ Vercel (Vue SPA, VITE_API_BASE_URL)
       │
       └──── https/wss (kèm cookie, credentials) ──▶ Cloud Run (Go API) ──▶ Supabase (Postgres)
```

> **Realtime:** pubsub + WebSocket hub đang **in-memory** → Cloud Run chạy **1 instance**
> (`--max-instances 1`). Nếu cần scale ngang, chuyển pubsub sang Postgres LISTEN/NOTIFY (hoặc
> Redis) trước. FE vẫn có fallback refetch khi focus nên mất realtime không làm sai dữ liệu.

---

## 0. Yêu cầu trước

- `gcloud` đã cài + đăng nhập: `gcloud auth login` rồi `gcloud config set project <PROJECT_ID>`
- Bật API cần thiết (một lần):
  ```bash
  gcloud services enable run.googleapis.com cloudbuild.googleapis.com artifactregistry.googleapis.com
  ```
- Có sẵn Supabase project (lấy connection string ở **Project Settings → Database**).

## 1. Chạy migration lên Supabase (goose)

Dùng **direct connection** hoặc **session pooler** cho migration (KHÔNG dùng transaction pooler 6543).

```bash
cd src
# Ví dụ session pooler (cổng 5432):
make migrate-up DATABASE_URL="postgres://postgres.<ref>:<password>@aws-0-<region>.pooler.supabase.com:5432/postgres?sslmode=require"
make migrate-status DATABASE_URL="postgres://postgres.<ref>:...@...:5432/postgres?sslmode=require"
```

> **Seed** (`make seed DATABASE_URL=...`) là dữ liệu DEV (Alice/Bob/Carol). Với prod chỉ chạy nếu
> bạn thực sự muốn tài khoản mẫu; nếu không, tạo user/hộ qua quy trình nghiệp vụ.

## 2. Cấu hình biến môi trường

```bash
cd src
cp deploy/env.example.yaml deploy/env.yaml   # deploy/env.yaml ĐÃ .gitignore
# Sửa deploy/env.yaml:
#   DATABASE_URL   → Supabase POOLER (session, cổng 5432, sslmode=require)
#   JWT_SECRET     → openssl rand -base64 48
#   CORS_ORIGINS   → https://<your-app>.vercel.app  (tạm thời chưa có thì điền sau, redeploy)
#   COOKIE_SECURE=true, COOKIE_SAMESITE=none
```

- App runtime nên dùng **session pooler** (cổng 5432): hỗ trợ prepared statement, không cần đổi code.
- Nếu buộc dùng **transaction pooler** (cổng 6543): thêm `DB_PREFER_SIMPLE_PROTOCOL: "true"`.

## 3. Deploy

```bash
cd src
./deploy/deploy.sh                       # mặc định service=household-finance-api, region=asia-southeast1
# hoặc: REGION=asia-southeast1 SERVICE=hf-api ./deploy/deploy.sh
```

Script build image từ `src/api/Dockerfile` qua Cloud Build rồi deploy, và in ra **URL API**
(vd `https://household-finance-api-xxxx-as.a.run.app`).

Kiểm tra nhanh:
```bash
curl -s https://<cloud-run-url>/healthz     # {"status":"ok","db":"ok"} = app sống + DB kết nối
curl -i https://<cloud-run-url>/api/me      # 401 (chưa đăng nhập) — API sống
# login user giả: 401 INVALID_CREDENTIALS = DB đã migrate; 500 = lỗi DB/migrate
curl -s -X POST https://<cloud-run-url>/api/auth/login \
  -H 'Content-Type: application/json' -d '{"email":"x@x.com","password":"x"}'
```

**Tài liệu API tương tác** (dễ kiểm tra & thử API sau deploy):
- Swagger UI: `https://<cloud-run-url>/docs` — xem toàn bộ endpoint + "Try it out" (đăng nhập
  ngay trong trang là gọi được các API cần auth vì cùng origin, cookie gửi kèm).
- OpenAPI spec: `https://<cloud-run-url>/openapi.yaml`.
- Tắt docs ở môi trường nhạy cảm: đặt env `DOCS_ENABLED=false`.

## 4. Deploy web lên Vercel

1. Import repo vào Vercel; **Root Directory = `src/web`** (Vercel tự nhận Vite; đã có `vercel.json`).
2. **Environment Variables** → thêm:
   - `VITE_API_BASE_URL = https://<cloud-run-url>`  (KHÔNG có dấu `/` ở cuối)
3. Deploy. Lấy domain Vercel (vd `https://<your-app>.vercel.app`).
4. Quay lại `deploy/env.yaml`, đảm bảo `CORS_ORIGINS` chứa đúng domain Vercel, rồi **deploy lại API**:
   `./deploy/deploy.sh`.

## 5. Kiểm thử end-to-end

- Mở web Vercel → đăng nhập → xem sổ giao dịch/ngân sách/báo cáo.
- DevTools → Network: request tới Cloud Run có `Access-Control-Allow-Origin` = domain Vercel và
  cookie `hf_token` được set (`Secure`, `SameSite=None`).

---

## ⚠️ Lưu ý cookie bên-thứ-ba (QUAN TRỌNG)

Vì `*.vercel.app` và `*.run.app` là **hai domain gốc khác nhau**, cookie phiên là **third-party
cookie**. **Safari (ITP) chặn mặc định** và **Chrome đang loại bỏ dần** → đăng nhập cookie cross-site
có thể KHÔNG chạy ổn định lâu dài.

**Cách khắc phục chuẩn production — dùng chung domain gốc:**

- Trỏ web `app.example.com` (Vercel) và API `api.example.com` (Cloud Run **Domain Mapping**).
- Đổi `deploy/env.yaml`:
  ```yaml
  CORS_ORIGINS: "https://app.example.com"
  COOKIE_SAMESITE: "lax"     # cùng site → không cần None
  COOKIE_DOMAIN: ".example.com"
  ```
- Đặt `VITE_API_BASE_URL=https://api.example.com` trên Vercel.

Khi đó cookie là **first-party** (SameSite=Lax) → chạy tốt trên mọi trình duyệt. Code đã hỗ trợ sẵn
qua biến môi trường, không cần sửa lại.

---

## Biến môi trường API (tham chiếu)

| Biến | Ý nghĩa | Mặc định |
|---|---|---|
| `DATABASE_URL` | Chuỗi kết nối Supabase (pooler, sslmode=require) | dev localhost |
| `JWT_SECRET` | Khóa ký JWT | `dev-secret-change-me` |
| `PORT` | Cổng lắng nghe (Cloud Run tự set) | `8080` |
| `CORS_ORIGINS` | Allow-list origin (phân tách bằng `,`) | rỗng (chỉ same-origin) |
| `COOKIE_SECURE` | Cookie chỉ gửi qua HTTPS | `false` |
| `COOKIE_SAMESITE` | `lax` \| `none` \| `strict` | `lax` |
| `COOKIE_DOMAIN` | Domain cookie (vd `.example.com`) | rỗng (host-only) |
| `DOCS_ENABLED` | Bật `/docs` (Swagger UI) + `/openapi.yaml` | `true` |
| `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` | Giới hạn pool | `10` / `5` |
| `DB_CONN_MAX_LIFETIME_SEC` / `DB_CONN_MAX_IDLE_SEC` | Vòng đời kết nối | `300` / `60` |
| `DB_PREFER_SIMPLE_PROTOCOL` | Tắt prepared statement (transaction pooler 6543) | `false` |
