// Chữ viết tắt cho avatar từ tên hiển thị.
// "Alice" → "A"; "Nguyễn Văn A" → "NA" (chữ đầu của từ đầu + từ cuối).
export function initials(name?: string | null): string {
  if (!name) return '?'
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return '?'
  const first = parts[0][0] ?? ''
  const last = parts.length > 1 ? (parts[parts.length - 1][0] ?? '') : ''
  return (first + last).toUpperCase() || '?'
}
