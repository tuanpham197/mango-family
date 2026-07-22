package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func d(y int, m time.Month, day int) time.Time {
	return time.Date(y, m, day, 12, 0, 0, 0, time.UTC)
}

func TestResolveRange(t *testing.T) {
	now := d(2026, 7, 15) // Thứ Tư

	from, to := ResolveRange("month", now)
	assert.Equal(t, "2026-07-01", from.Format("2006-01-02"))
	assert.Equal(t, "2026-07-31", to.Format("2006-01-02"))

	from, to = ResolveRange("quarter", now)
	assert.Equal(t, "2026-07-01", from.Format("2006-01-02")) // Q3
	assert.Equal(t, "2026-09-30", to.Format("2006-01-02"))

	from, to = ResolveRange("year", now)
	assert.Equal(t, "2026-01-01", from.Format("2006-01-02"))
	assert.Equal(t, "2026-12-31", to.Format("2006-01-02"))

	from, to = ResolveRange("week", now)
	assert.Equal(t, time.Monday, from.Weekday()) // bắt đầu Thứ Hai
	assert.Equal(t, from.AddDate(0, 0, 6), to)   // 7 ngày
	assert.False(t, now.Before(from))            // now trong tuần
	assert.True(t, now.Before(to.AddDate(0, 0, 1)))
}

func TestChooseGroupUnit(t *testing.T) {
	assert.Equal(t, GroupDay, ChooseGroupUnit(d(2026, 7, 1), d(2026, 7, 31)))   // 31 ngày → day
	assert.Equal(t, GroupWeek, ChooseGroupUnit(d(2026, 7, 1), d(2026, 8, 1)))   // 32 ngày → week
	assert.Equal(t, GroupWeek, ChooseGroupUnit(d(2026, 7, 1), d(2026, 9, 30)))  // 92 ngày → week
	assert.Equal(t, GroupMonth, ChooseGroupUnit(d(2026, 7, 1), d(2026, 10, 1))) // 93 ngày → month
}

func TestBucketKey(t *testing.T) {
	assert.Equal(t, "2026-07-15", BucketKey(GroupDay, d(2026, 7, 15)))
	assert.Equal(t, "2026-07", BucketKey(GroupMonth, d(2026, 7, 15)))
	assert.Regexp(t, `^2026-W\d{2}$`, BucketKey(GroupWeek, d(2026, 7, 15)))
}

func TestBucketSequence(t *testing.T) {
	// day: 2026-07-01..07-05 → 5 mốc liên tục
	days := BucketSequence(GroupDay, d(2026, 7, 1), d(2026, 7, 5))
	assert.Equal(t, []string{"2026-07-01", "2026-07-02", "2026-07-03", "2026-07-04", "2026-07-05"}, days)

	// month: cả năm → 12 mốc
	months := BucketSequence(GroupMonth, d(2026, 1, 1), d(2026, 12, 31))
	assert.Len(t, months, 12)
	assert.Equal(t, "2026-01", months[0])
	assert.Equal(t, "2026-12", months[11])

	// week: quý → 14 tuần (khớp smoke test)
	weeks := BucketSequence(GroupWeek, d(2026, 7, 1), d(2026, 9, 30))
	assert.Len(t, weeks, 14)
}
