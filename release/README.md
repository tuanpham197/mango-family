# release/

Tài liệu phát hành & vận hành deploy cho **Mango Family**.

| File | Nội dung |
|------|----------|
| [`RELEASE-NOTES.md`](./RELEASE-NOTES.md) | Ghi chú phát hành theo version (mới nhất trước) — phạm vi, thay đổi API/DB, có cần migration không |
| [`DEPLOY.md`](./DEPLOY.md) | Runbook deploy API (Cloud Run) & Web (Vercel) + **chiến lược nhánh deploy** |

Thiết lập hạ tầng gốc (gcloud, Supabase, env, cookie cross-origin): [`../docs/deploy-cloud-run.md`](../docs/deploy-cloud-run.md).

## Nhánh deploy

- `deploy/api` → nguồn deploy **API** lên Cloud Run (quan tâm `src/api/**`, `src/db/migrations/**`).
- `deploy/web` → nguồn deploy **Web** lên Vercel (quan tâm `src/web/**`).
- `develop` → nhánh tích hợp; cắt release bằng cách merge sang `deploy/api` và/hoặc `deploy/web`.

Quy trình đầy đủ ở [`DEPLOY.md`](./DEPLOY.md).
