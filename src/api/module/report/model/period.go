package model

import (
	"fmt"
	"time"
)

// ResolveRange — quy preset (week/month/quarter/year) về [from, to] (inclusive) tính
// từ `now` (D35). Tuần bắt đầu Thứ Hai; tháng/quý/năm theo dương lịch. Dùng cho test
// + tiện ích; FE cũng có thể tự gửi from/to.
func ResolveRange(preset string, now time.Time) (from, to time.Time) {
	loc := now.Location()
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, loc)
	switch preset {
	case "week":
		offset := (int(now.Weekday()) + 6) % 7 // Mon=0
		from = today.AddDate(0, 0, -offset)
		to = from.AddDate(0, 0, 6)
	case "quarter":
		q := (int(m) - 1) / 3
		from = time.Date(y, time.Month(q*3+1), 1, 0, 0, 0, 0, loc)
		to = from.AddDate(0, 3, -1)
	case "year":
		from = time.Date(y, 1, 1, 0, 0, 0, 0, loc)
		to = time.Date(y, 12, 31, 0, 0, 0, 0, loc)
	default: // month
		from = time.Date(y, m, 1, 0, 0, 0, 0, loc)
		to = from.AddDate(0, 1, -1)
	}
	return from, to
}

// EndExclusive — mốc loại trừ cho cửa sổ [from, to] inclusive (đầu ngày sau `to`).
func EndExclusive(to time.Time) time.Time {
	return time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location()).AddDate(0, 0, 1)
}

// ChooseGroupUnit — đơn vị gom nhóm theo độ dài khoảng (D36): ≤ 31 ngày → day;
// ≤ 92 ngày → week; còn lại → month (span inclusive).
func ChooseGroupUnit(from, to time.Time) string {
	spanDays := int(to.Sub(from).Hours()/24) + 1
	switch {
	case spanDays <= 31:
		return GroupDay
	case spanDays <= 92:
		return GroupWeek
	default:
		return GroupMonth
	}
}

// BucketKey — nhãn mốc thời gian (dùng CHUNG cho kết quả SQL lẫn dãy mốc sinh ở Go → khớp định dạng).
func BucketKey(unit string, t time.Time) string {
	switch unit {
	case GroupWeek:
		y, w := t.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", y, w)
	case GroupMonth:
		return t.Format("2006-01")
	default: // day
		return t.Format("2006-01-02")
	}
}

// BucketSequence — dãy nhãn mốc liên tục từ `from` đến `to` theo đơn vị (điền mốc trống = 0).
func BucketSequence(unit string, from, to time.Time) []string {
	var keys []string
	seen := map[string]bool{}
	cur := truncateToUnit(unit, from)
	end := EndExclusive(to)
	for cur.Before(end) {
		k := BucketKey(unit, cur)
		if !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
		cur = stepUnit(unit, cur)
	}
	return keys
}

func truncateToUnit(unit string, t time.Time) time.Time {
	loc := t.Location()
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	switch unit {
	case GroupWeek:
		offset := (int(t.Weekday()) + 6) % 7 // về Thứ Hai
		return day.AddDate(0, 0, -offset)
	case GroupMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
	default:
		return day
	}
}

func stepUnit(unit string, t time.Time) time.Time {
	switch unit {
	case GroupWeek:
		return t.AddDate(0, 0, 7)
	case GroupMonth:
		return t.AddDate(0, 1, 0)
	default:
		return t.AddDate(0, 0, 1)
	}
}

// PgTruncUnit — map group unit → đơn vị date_trunc của Postgres (whitelist an toàn).
func PgTruncUnit(unit string) string {
	switch unit {
	case GroupWeek:
		return "week"
	case GroupMonth:
		return "month"
	default:
		return "day"
	}
}
