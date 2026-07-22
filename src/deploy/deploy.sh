#!/usr/bin/env bash
# Deploy API Go lên Google Cloud Run (build từ source qua Cloud Build).
#
# Yêu cầu trước:
#   - gcloud đã đăng nhập + set project:  gcloud auth login && gcloud config set project <ID>
#   - đã bật API:  gcloud services enable run.googleapis.com cloudbuild.googleapis.com artifactregistry.googleapis.com
#   - đã chạy migration lên Supabase (xem docs/deploy-cloud-run.md)
#   - đã điền deploy/env.yaml (copy từ deploy/env.example.yaml)
#
# Cách dùng:  ./deploy/deploy.sh        (hoặc REGION=... SERVICE=... ./deploy/deploy.sh)
set -euo pipefail

SERVICE="${SERVICE:-household-finance-api}"
REGION="${REGION:-asia-southeast1}"   # Singapore — gần VN
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_DIR="$(cd "$HERE/../api" && pwd)"
ENV_FILE="$HERE/env.yaml"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "✗ Thiếu $ENV_FILE — copy từ deploy/env.example.yaml rồi điền giá trị thật." >&2
  exit 1
fi

# --max-instances 1: realtime (pubsub + WS hub) đang IN-MEMORY nên phải 1 instance
#   (nếu scale >1, event ở instance A không tới client ở instance B). FE có fallback
#   refetch khi focus. Muốn scale ngang → chuyển pubsub sang Postgres LISTEN/NOTIFY.
# --timeout 3600: cho phép WebSocket sống lâu (Cloud Run tối đa 60 phút/request).
gcloud run deploy "$SERVICE" \
  --source "$SOURCE_DIR" \
  --region "$REGION" \
  --platform managed \
  --allow-unauthenticated \
  --port 8080 \
  --min-instances 0 \
  --max-instances 1 \
  --cpu 1 \
  --memory 512Mi \
  --timeout 3600 \
  --env-vars-file "$ENV_FILE"

echo ""
echo "✓ Đã deploy. URL API:"
gcloud run services describe "$SERVICE" --region "$REGION" --format 'value(status.url)'
echo "→ Đặt VITE_API_BASE_URL = URL trên trong Vercel, và thêm domain Vercel vào CORS_ORIGINS."
