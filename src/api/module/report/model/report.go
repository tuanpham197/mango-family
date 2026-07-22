// Package model — DTO suy ra cho báo cáo (feature 005). KHÔNG map bảng, không lưu;
// tổng hợp ở biz/storage khi đọc (D34/D41). Chỉ đọc transactions/categories.
package model

import (
	"time"

	transactionmodel "household-finance/api/module/transaction/model"

	"github.com/google/uuid"
)

// TrendRow / CategoryTrendRow — hàng thô từ storage (GROUP BY date_trunc); biz định
// dạng bucket + điền mốc trống (BucketKey/BucketSequence).
type TrendRow struct {
	Bucket  time.Time
	Income  float64
	Expense float64
}

type CategoryTrendRow struct {
	Bucket time.Time
	Amount float64
}

// Group unit cho biểu đồ xu hướng (D36).
const (
	GroupDay   = "day"
	GroupWeek  = "week"
	GroupMonth = "month"
)

// CategoryBreakdown — chi của một danh mục cha (gộp con) trong khoảng (D37).
type CategoryBreakdown struct {
	CategoryID     uuid.UUID `json:"category_id"`
	CategoryName   string    `json:"category_name"`
	CategoryHidden bool      `json:"category_hidden"`
	Amount         float64   `json:"amount"`
	Percent        int       `json:"percent"`
}

// TrendPoint — một mốc thời gian của xu hướng thu/chi (overview).
type TrendPoint struct {
	Bucket  string  `json:"bucket"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

// ReportOverview — response GET /api/reports/overview (contracts).
type ReportOverview struct {
	From              string              `json:"from"`
	To                string              `json:"to"`
	GroupUnit         string              `json:"group_unit"`
	Income            float64             `json:"income"`
	Expense           float64             `json:"expense"`
	Net               float64             `json:"net"`
	CategoryBreakdown []CategoryBreakdown `json:"category_breakdown"`
	Trend             []TrendPoint        `json:"trend"`
}

// CategoryTrendPoint — một mốc xu hướng chi của một danh mục.
type CategoryTrendPoint struct {
	Bucket string  `json:"bucket"`
	Amount float64 `json:"amount"`
}

// CategoryReport — response GET /api/reports/category/:id (contracts).
type CategoryReport struct {
	CategoryID   uuid.UUID                   `json:"category_id"`
	CategoryName string                      `json:"category_name"`
	From         string                      `json:"from"`
	To           string                      `json:"to"`
	GroupUnit    string                      `json:"group_unit"`
	Total        float64                     `json:"total"`
	Transactions []transactionmodel.ListItem `json:"transactions"`
	Trend        []CategoryTrendPoint        `json:"trend"`
}
