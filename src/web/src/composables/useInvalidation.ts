import { onMounted, onUnmounted } from 'vue'

// WS invalidation client (research D8): kết nối /ws (cookie same-origin),
// nhận {"type":"categories_changed"|"transactions_changed"} → refetch store
// tương ứng. Tự reconnect (backoff); fallback refetch khi tab focus lại.

type Handler = () => void

const handlers = new Map<string, Set<Handler>>()
let socket: WebSocket | null = null
let attempts = 0
let reconnectTimer: ReturnType<typeof setTimeout> | null = null

function notifyAll() {
  for (const set of handlers.values()) for (const h of set) h()
}

// URL của /ws: prod tách origin dùng VITE_API_BASE_URL (đổi http→ws, https→wss);
// dev/same-origin dùng host hiện tại (Vite proxy /ws → :8080).
function wsUrl(): string {
  const base = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '')
  if (base) return base.replace(/^http/, 'ws') + '/ws'
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  return `${proto}://${location.host}/ws`
}

function connect() {
  if (socket || handlers.size === 0) return
  const ws = new WebSocket(wsUrl())
  socket = ws
  ws.onopen = () => {
    // Sau khi nối lại có thể đã lỡ event — refetch toàn bộ cho chắc.
    if (attempts > 0) notifyAll()
    attempts = 0
  }
  ws.onmessage = (e) => {
    try {
      const { type } = JSON.parse(e.data) as { type: string }
      handlers.get(type)?.forEach((h) => h())
    } catch {
      /* bỏ qua frame không hợp lệ */
    }
  }
  ws.onclose = () => {
    socket = null
    if (handlers.size === 0) return
    attempts = Math.min(attempts + 1, 6)
    reconnectTimer = setTimeout(connect, attempts * 1000)
  }
}

function disconnectIfIdle() {
  if (handlers.size > 0) return
  if (reconnectTimer) clearTimeout(reconnectTimer)
  socket?.close()
  socket = null
}

/**
 * Đăng ký refetch khi server phát `topic`; dùng trong view/store setup.
 * Ví dụ: useInvalidation('categories_changed', () => store.fetch())
 */
export function useInvalidation(topic: string, refetch: Handler) {
  const onFocus = () => {
    if (document.visibilityState === 'visible') refetch()
  }
  onMounted(() => {
    if (!handlers.has(topic)) handlers.set(topic, new Set())
    handlers.get(topic)!.add(refetch)
    connect()
    document.addEventListener('visibilitychange', onFocus)
  })
  onUnmounted(() => {
    const set = handlers.get(topic)
    set?.delete(refetch)
    if (set && set.size === 0) handlers.delete(topic)
    document.removeEventListener('visibilitychange', onFocus)
    disconnectIfIdle()
  })
}
