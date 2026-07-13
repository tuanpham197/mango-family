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

export interface Paging {
  page: number
  page_size: number
  total: number
}

export interface PagedResult<T> {
  data: T
  paging?: Paging
}

async function request<T>(path: string, init: RequestInit = {}): Promise<PagedResult<T>> {
  const res = await fetch(path, {
    credentials: 'same-origin',
    ...init,
    headers: {
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
      ...init.headers,
    },
  })
  if (res.status === 204) return { data: undefined as T }
  const json = await res.json().catch(() => null)
  if (!res.ok) {
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
