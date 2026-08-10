# release/

Tài liệu phát hành & vận hành deploy cho **Mango Family**.

| File | Nội dung |
|------|----------|
| [`RELEASE-NOTES.md`](./RELEASE-NOTES.md) | **Chỉ mục** release + quy ước tạo note mới + lịch sử trước v0.8.0 |
| [`notes/`](./notes/) | **Một file / một lần deploy**, đặt tên theo ngày deploy: `YYYY-MM-DD-<slug>.md` |
| [`DEPLOY.md`](./DEPLOY.md) | Runbook deploy API (Cloud Run) & Web (Vercel) + **chiến lược nhánh deploy** |

Thiết lập hạ tầng gốc (gcloud, Supabase, env, cookie cross-origin): [`../docs/deploy-cloud-run.md`](../docs/deploy-cloud-run.md).

## Nhánh deploy

- `deploy/api` → nguồn deploy **API** lên Cloud Run (quan tâm `src/api/**`, `src/db/migrations/**`).
- `deploy/web` → nguồn deploy **Web** lên Vercel (quan tâm `src/web/**`).
- `develop` → nhánh tích hợp; cắt release bằng cách merge sang `deploy/api` và/hoặc `deploy/web`.

Quy trình đầy đủ ở [`DEPLOY.md`](./DEPLOY.md).
