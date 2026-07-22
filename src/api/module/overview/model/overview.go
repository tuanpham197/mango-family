// Package model — DTO suy ra cho màn Tổng quan (feature 004). KHÔNG map bảng, không
// lưu; tính ở biz khi đọc GET /api/overview (D27). Chỉ đọc dữ liệu 001/002/003.
package model

import (
	transactionmodel "household-finance/api/module/transaction/model"

	"github.com/google/uuid"
)

// MonthCashflow — tổng Thu/Chi & chênh lệch của hộ trong tháng dương lịch hiện tại.
type MonthCashflow struct {
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Net     float64 `json:"net"`
}

// CategorySpending — chi của một danh mục cha (gộp con một cấp) trong tháng (D29).
type CategorySpending struct {
	CategoryID     uuid.UUID `json:"category_id"`
	CategoryName   string    `json:"category_name"`
	CategoryHidden bool      `json:"category_hidden"`
	Amount         float64   `json:"amount"`
	Percent        int       `json:"percent"` // biz tính = round(amount/tổng chi tháng × 100)
}

// OverviewSummary — response GET /api/overview (contracts/overview-api.md, D27).
type OverviewSummary struct {
	NetWorth              float64                     `json:"net_worth"`
	NetWorthChangePercent *float64                    `json:"net_worth_change_percent"` // null khi không tính được (D28)
	Month                 MonthCashflow               `json:"month"`
	CategorySpending      []CategorySpending          `json:"category_spending"`
	RecentTransactions    []transactionmodel.ListItem `json:"recent_transactions"`
}
