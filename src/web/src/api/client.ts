// Fetch wrapper — parse app_response của Go API (contracts/category-api.md):
// thành công {data, paging?}; lỗi {error:{code,message,field?,...extra}}.

export interface ApiErrorBody {
  code: string
  message: string
  field?: string
  [key: string]: unknown
}

export class ApiError extends Error {
  status: number
  body: ApiErrorBody

  constructor(status: number, body: ApiErrorBody) {
    super(body.message)
    this.status = status
    this.body = body
  }

  get code(): string {
    return this.body.code
  }

  get field(): string | undefined {
    return this.body.field
  }
}

// Xử lý toàn cục khi phiên không hợp lệ (401): app đăng ký ở main.ts để
// xoá phiên + đưa về /login (vd cookie trỏ user đã bị xoá sau reset DB).
type UnauthorizedHandler = () => void
let unauthorizedHandler: UnauthorizedHandler | null = null
export function setUnauthorizedHandler(h: UnauthorizedHandler) {
  unauthorizedHandler = h
}

export interface Paging {
  page: number
  page_size: number
  total: number
}

export interface PagedResult<T> {
  data: T
  paging?: Paging
}

// Base URL của API. Rỗng (dev/same-origin) → path tương đối qua Vite proxy.
// Prod tách origin (web Vercel ↔ API Cloud Run): đặt VITE_API_BASE_URL=https://<cloud-run-url>.
const API_BASE = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '')

async function request<T>(path: string, init: RequestInit = {}): Promise<PagedResult<T>> {
  const res = await fetch(API_BASE + path, {
    credentials: 'include', // gửi cookie phiên cả khi cross-origin (Vercel↔Cloud Run)
    ...init,
    headers: {
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
      ...init.headers,
    },
  })
  if (res.status === 204) return { data: undefined as T }
  const json = await res.json().catch(() => null)
  if (!res.ok) {
    // 401 ở endpoint đã đăng nhập = phiên hết hạn/không hợp lệ → xử lý toàn cục.
    // Bỏ qua /api/me (bootstrap) & /api/auth/* (đăng nhập/đăng xuất) để không
    // phá luồng khởi động và thông báo "sai mật khẩu".
    if (res.status === 401 && !path.startsWith('/api/me') && !path.startsWith('/api/auth/')) {
      unauthorizedHandler?.()
    }
    throw new ApiError(res.status, json?.error ?? { code: 'INTERNAL', message: 'Đã có lỗi xảy ra' })
  }
  return json as PagedResult<T>
}

/** Gọi API, trả về phần `data`. */
export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  return (await request<T>(path, init)).data
}

/** Gọi API, trả về cả `data` + `paging` (danh sách phân trang). */
export function apiPaged<T>(path: string, init: RequestInit = {}): Promise<PagedResult<T>> {
  return request<T>(path, init)
}
