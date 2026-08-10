# Deploy Runbook — API (Cloud Run) & Web (Vercel)

Quy trình phát hành cho **Mango Family**. Đây là runbook thao tác theo từng release; **thiết lập hạ tầng gốc** (gcloud, Supabase, biến môi trường, cookie cross-origin) nằm ở [`docs/deploy-cloud-run.md`](../docs/deploy-cloud-run.md) — đọc file đó một lần khi setup lần đầu.

```
develop ──(cắt release)──┬──▶ deploy/api  ──▶ Cloud Run (Go API)  ──▶ Supabase
                         └──▶ deploy/web  ──▶ Vercel (Vue SPA)
```

## Nhánh deploy (branch strategy)

| Nhánh | Deploy cái gì | Kích hoạt | Chỉ quan tâm |
|-------|---------------|-----------|--------------|
| `develop` | (tích hợp — không deploy trực tiếp) | — | toàn repo |
| `deploy/api` | **API** lên Cloud Run | thủ công: `./deploy/deploy.sh` (hoặc Cloud Build trigger theo nhánh này) | `src/api/**`, `src/db/migrations/**` |
| `deploy/web` | **Web** lên Vercel | Vercel Production Branch = `deploy/web` → auto build khi push (hoặc `vercel --prod`) | `src/web/**` |

**Vì sao tách 2 nhánh?** API và Web phát hành **độc lập**: web hotfix (UI) không cần redeploy API; API đổi (có thể kèm migration) không ép build lại web. Mỗi nhánh là "con trỏ phát hành" — chỉ tiến khi chủ đích deploy phần đó.

**Quy tắc vàng**: chỉ đưa code đã ở `develop` (đã xanh test) lên nhánh deploy. Không commit thẳng lên `deploy/*`.

---

## Checklist trước mỗi release

- [ ] `develop` xanh: `cd src && make test` (Go unit + Vitest) · `make test-api-integration` · `make test-e2e`.
- [ ] Cập nhật [`release/RELEASE-NOTES.md`](./RELEASE-NOTES.md): thêm mục version mới, ghi rõ **có migration không** và **có đổi env không**.
- [ ] Xác định thứ tự: nếu release có **migration** hoặc **API mới mà web phụ thuộc** → **deploy API trước, web sau**. Nếu chỉ web → chỉ deploy web.
- [ ] `deploy/env.yaml` (Cloud Run) và Vercel env (`VITE_API_BASE_URL`) còn đúng.

---

## A. Deploy API (Cloud Run)

> Tiền đề: đã setup theo `docs/deploy-cloud-run.md` (gcloud auth + project, `deploy/env.yaml`).

```bash
cd /Users/kozokozo/personal/specs_app

# 1) Cập nhật nhánh deploy/api từ develop
git fetch origin
git switch deploy/api && git merge --ff-only origin/develop   # hoặc: git merge --no-ff develop
git push origin deploy/api

# 2) (CHỈ khi release có migration mới) chạy migration lên Supabase TRƯỚC khi deploy code
#    → Release v0.8.0/008 KHÔNG có migration, BỎ QUA bước này.
cd src
make migrate-status DATABASE_URL="postgres://postgres.<ref>:<pwd>@aws-0-<region>.pooler.supabase.com:5432/postgres?sslmode=require"
make migrate-up     DATABASE_URL="postgres://postgres.<ref>:<pwd>@aws-0-<region>.pooler.supabase.com:5432/postgres?sslmode=require"

# 3) Deploy (Cloud Build từ src/api/Dockerfile → Cloud Run, region asia-southeast1, max-instances 1)
./deploy/deploy.sh          # in ra URL API

# 4) Smoke test
curl -s https://<cloud-run-url>/healthz     # {"status":"ok","db":"ok"}
curl -i https://<cloud-run-url>/api/me      # 401 = API sống
```

Ràng buộc quan trọng (đã cấu hình sẵn trong `deploy.sh`): **`--max-instances 1`** vì realtime (pubsub + WS hub) đang in-memory. Đừng tăng lên >1 nếu chưa chuyển sang Postgres LISTEN/NOTIFY.

## B. Deploy Web (Vercel)

> Tiền đề: dự án Vercel đã import repo, **Root Directory = `src/web`**, env `VITE_API_BASE_URL = https://<cloud-run-url>` (không có `/` cuối). `src/web/vercel.json` đã rewrite `/api/*` → Cloud Run.

```bash
cd /Users/kozokozo/personal/specs_app

# 1) Cập nhật nhánh deploy/web từ develop
git fetch origin
git switch deploy/web && git merge --ff-only origin/develop   # hoặc: git merge --no-ff develop
git push origin deploy/web
```

- Nếu Vercel Production Branch = `deploy/web`: push ở bước 1 **tự trigger** build production.
- Nếu deploy thủ công: `cd src/web && vercel --prod` (Vercel CLI).

```bash
# 2) Smoke test sau khi Vercel build xong
#    - Mở https://<your-app>.vercel.app → đăng nhập → mở /reports → kiểm phần "Theo thành viên".
#    - DevTools → Network: request Cloud Run có Access-Control-Allow-Origin = domain Vercel; cookie hf_token (Secure, SameSite=None).
```

> Nếu API vừa đổi `CORS_ORIGINS`/domain: nhớ deploy lại API (mục A) để allow-list khớp domain Vercel.

---

## Thứ tự cho release v0.8.0 (feature 008)

1. Không migration, không đổi env → an toàn.
2. Có **API mới** (`/api/reports/members`, `/member/:id`) mà web dùng → **deploy API (A) trước, web (B) sau**.
3. Smoke test `/reports` phần theo thành viên trên domain Vercel.

## Rollback

- **API**: Cloud Run giữ revision cũ → `gcloud run services update-traffic household-finance-api --region asia-southeast1 --to-revisions <REV_CŨ>=100`. Feature 008 read-only, không cần rollback DB.
- **Web**: Vercel → Deployments → "Promote to Production" bản trước; hoặc reset nhánh `deploy/web` về commit cũ rồi push.
- **Migration** (khi có): viết sẵn `goose down` nếu bước lên có rủi ro; feature 008 không áp dụng.
