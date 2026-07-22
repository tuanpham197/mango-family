package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 12, 0, 0, 0, time.UTC)
}

func TestPeriodKey_Monthly(t *testing.T) {
	assert.Equal(t, "2026-07", PeriodKey(PeriodMonthly, date(2026, 7, 14)))
	assert.Equal(t, "2026-07", PeriodKey(PeriodMonthly, date(2026, 7, 31)))
	assert.Equal(t, "2026-12", PeriodKey(PeriodMonthly, date(2026, 12, 1)))
}

func TestPeriodKey_WeeklyISOFormat(t *testing.T) {
	// Định dạng IYYY-IW (D20); tuần ISO 1 của năm nào đó có 2 chữ số zero-pad.
	y, w := date(2026, 1, 5).ISOWeek()
	assert.Equal(t, PeriodKey(PeriodWeekly, date(2026, 1, 5)), formatISO(y, w))
	assert.Regexp(t, `^\d{4}-\d{2}$`, PeriodKey(PeriodWeekly, date(2026, 1, 5)))
}

func formatISO(y, w int) string {
	// Đối chiếu độc lập với cài đặt.
	return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006") + "-" + pad2(w)
}
func pad2(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

func TestPeriodKey_OneTime(t *testing.T) {
	assert.Equal(t, "once", PeriodKey(PeriodOneTime, date(2026, 7, 14)))
}

func TestResolvePeriod_MonthlyWindow(t *testing.T) {
	p := ResolvePeriod(PeriodMonthly, nil, nil, date(2026, 7, 14))
	assert.Equal(t, "2026-07", p.Key)
	assert.Equal(t, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), p.Start)
	assert.Equal(t, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), p.EndExcl)
	// Ngày cuối tháng vẫn cùng cửa sổ (biên kỳ).
	assert.Equal(t, p.Start, ResolvePeriod(PeriodMonthly, nil, nil, date(2026, 7, 31)).Start)
}

func TestResolvePeriod_WeeklyStartsMonday(t *testing.T) {
	// Với mọi ngày trong tuần, Start là Thứ Hai và cửa sổ dài đúng 7 ngày, chứa `now`.
	for d := 13; d <= 19; d++ { // 2026-07-13 (Mon) … 2026-07-19 (Sun)
		now := date(2026, 7, d)
		p := ResolvePeriod(PeriodWeekly, nil, nil, now)
		assert.Equal(t, time.Monday, p.Start.Weekday(), "Start phải là Thứ Hai (ngày %d)", d)
		assert.Equal(t, p.Start.AddDate(0, 0, 7), p.EndExcl)
		assert.False(t, now.Before(p.Start))
		assert.True(t, now.Before(p.EndExcl))
	}
}

func TestResolvePeriod_OneTimeInclusiveEnd(t *testing.T) {
	start := date(2026, 7, 5)
	end := date(2026, 7, 10)
	p := ResolvePeriod(PeriodOneTime, &start, &end, date(2026, 7, 7))
	require.Equal(t, "once", p.Key)
	assert.Equal(t, time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC), p.Start)
	// EndExcl = end_date + 1 ngày → bao trọn ngày kết thúc.
	assert.Equal(t, time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC), p.EndExcl)
}
