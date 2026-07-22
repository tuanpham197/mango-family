package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Budget domain enums (003).
const (
	TypeCategory = "CATEGORY"
	TypeTotal    = "TOTAL"

	PeriodMonthly = "MONTHLY"
	PeriodWeekly  = "WEEKLY"
	PeriodOneTime = "ONE_TIME"

	StatusActive = "ACTIVE"
	StatusEnded  = "ENDED"

	LevelThreshold80 = "THRESHOLD_80"
	LevelOver100     = "OVER_100"
)

// Budget — giới hạn chi tiêu của hộ trong một kỳ (data-model 003).
// Tiến độ (spent/percent) KHÔNG lưu — suy ra ở biz khi đọc (D21).
// updated_at làm mốc lạc quan; GORM autoUpdateTime làm mới mỗi lần Updates (D25).
type Budget struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	HouseholdID uuid.UUID  `json:"household_id" gorm:"type:uuid;not null;index"`
	Type        string     `json:"type" gorm:"not null"` // CATEGORY | TOTAL
	CategoryID  *uuid.UUID `json:"category_id" gorm:"type:uuid"`
	LimitAmount float64    `json:"limit_amount" gorm:"type:numeric(14,2);not null"`
	PeriodType  string     `json:"period_type" gorm:"not null"` // MONTHLY | WEEKLY | ONE_TIME
	StartDate   *time.Time `json:"start_date" gorm:"type:date"`
	EndDate     *time.Time `json:"end_date" gorm:"type:date"`
	Status      string     `json:"status" gorm:"not null;default:ACTIVE"`
	CreatedBy   uuid.UUID  `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"` // mốc lạc quan (D25)
}

func (Budget) TableName() string { return "budgets" }

// BudgetAlert — một lần phát cảnh báo của ngân sách trong một kỳ (D23).
type BudgetAlert struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BudgetID   uuid.UUID `json:"budget_id" gorm:"type:uuid;not null;index"`
	PeriodKey  string    `json:"period_key" gorm:"not null"`
	Level      string    `json:"level" gorm:"not null"` // THRESHOLD_80 | OVER_100
	OverAmount *float64  `json:"over_amount" gorm:"type:numeric(14,2)"`
	FiredAt    time.Time `json:"fired_at" gorm:"autoCreateTime"`
}

func (BudgetAlert) TableName() string { return "budget_alerts" }

// AlertView — cảnh báo đang hoạt động kỳ hiện tại, embed vào BudgetView (contracts).
type AlertView struct {
	Level      string    `json:"level"`
	OverAmount *float64  `json:"over_amount"`
	FiredAt    time.Time `json:"fired_at"`
}

// BudgetView — response GET /api/budgets: Budget + tên hiển thị + tiến độ suy ra (D21).
// category_name/hidden/created_by_name nạp qua JOIN; period_key/spent/percent/alerts
// biz tính khi đọc (gorm:"-" — không map cột).
type BudgetView struct {
	Budget
	CategoryName   *string     `json:"category_name" gorm:"column:category_name"`
	CategoryHidden *bool       `json:"category_hidden" gorm:"column:category_hidden"`
	CreatedByName  string      `json:"created_by_name" gorm:"column:created_by_name"`
	PeriodKey      string      `json:"period_key" gorm:"-"`
	Spent          float64     `json:"spent" gorm:"-"`
	Percent        int         `json:"percent" gorm:"-"`
	Alerts         []AlertView `json:"alerts" gorm:"-"`
}

// Period — kỳ hiện tại suy ra (D20): khóa + cửa sổ [Start, EndExcl).
type Period struct {
	Key     string
	Start   time.Time
	EndExcl time.Time
}

// PeriodKey — khóa kỳ hiện tại (D20): MONTHLY 'YYYY-MM' · WEEKLY 'IYYY-IW' (ISO,
// bắt đầu Thứ Hai) · ONE_TIME 'once'.
func PeriodKey(periodType string, now time.Time) string {
	switch periodType {
	case PeriodWeekly:
		y, w := now.ISOWeek()
		return fmt.Sprintf("%04d-%02d", y, w)
	case PeriodOneTime:
		return "once"
	default: // MONTHLY
		return now.Format("2006-01")
	}
}

// ResolvePeriod — kỳ hiện tại + cửa sổ ngày để tổng chi (D20). ONE_TIME dùng
// [start_date, end_date] cố định; MONTHLY/WEEKLY tính từ `now` (tự lặp mỗi kỳ).
func ResolvePeriod(periodType string, startDate, endDate *time.Time, now time.Time) Period {
	loc := now.Location()
	switch periodType {
	case PeriodWeekly:
		// Thứ Hai đầu tuần ISO của `now`.
		daysSinceMon := (int(now.Weekday()) + 6) % 7 // Sun=0 → 6, Mon=1 → 0
		day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		start := day.AddDate(0, 0, -daysSinceMon)
		return Period{Key: PeriodKey(periodType, now), Start: start, EndExcl: start.AddDate(0, 0, 7)}
	case PeriodOneTime:
		var start, endExcl time.Time
		if startDate != nil {
			start = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
		}
		if endDate != nil {
			e := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, loc)
			endExcl = e.AddDate(0, 0, 1) // bao gồm trọn end_date
		}
		return Period{Key: "once", Start: start, EndExcl: endExcl}
	default: // MONTHLY — tháng dương lịch
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		return Period{Key: PeriodKey(periodType, now), Start: start, EndExcl: start.AddDate(0, 1, 0)}
	}
}
