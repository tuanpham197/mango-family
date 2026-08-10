package model

import (
	transactionmodel "household-finance/api/module/transaction/model"

	"github.com/google/uuid"
)

// DTO suy ra cho báo cáo THEO THÀNH VIÊN (feature 008). KHÔNG map bảng, không lưu;
// tổng hợp ở biz/storage khi đọc (D42–D44). Chỉ đọc transactions/users/household_members.

// FormerMemberID — sentinel cho dòng gộp "Thành viên cũ / Đã rời hộ" (D44/FR-010).
const FormerMemberID = "former"

// FormerMemberLabel — nhãn hiển thị dòng gộp người đã rời hộ.
const FormerMemberLabel = "Thành viên cũ / Đã rời hộ"

// MemberAggRow — hàng thô từ storage: Σ thu/chi của một `created_by` trong khoảng (D42).
type MemberAggRow struct {
	CreatedBy uuid.UUID
	Income    float64
	Expense   float64
}

// MemberInfo — thành viên hiện tại của hộ (household_members JOIN users). Email dùng để
// fallback tên hiển thị (D46/FR-009).
type MemberInfo struct {
	ID          uuid.UUID
	DisplayName string
	Email       string
}

// MemberRow — một dòng thành viên trong báo cáo tổng hợp (contracts §/members).
type MemberRow struct {
	MemberID    string  `json:"member_id"` // UUID; hoặc sentinel "former"
	DisplayName string  `json:"display_name"`
	IsFormer    bool    `json:"is_former"`
	Income      float64 `json:"income"`
	Expense     float64 `json:"expense"`
	Net         float64 `json:"net"`
}

// MemberTotals — tổng toàn hộ trong khoảng (phải bằng Σ MemberRow — đối soát SC-001).
type MemberTotals struct {
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Net     float64 `json:"net"`
}

// MembersReport — response GET /api/reports/members.
type MembersReport struct {
	From    string       `json:"from"`
	To      string       `json:"to"`
	Members []MemberRow  `json:"members"`
	Totals  MemberTotals `json:"totals"`
}

// MemberReport — response GET /api/reports/member/:id (drill-in, phân trang).
type MemberReport struct {
	MemberID     string                      `json:"member_id"`
	DisplayName  string                      `json:"display_name"`
	IsFormer     bool                        `json:"is_former"`
	From         string                      `json:"from"`
	To           string                      `json:"to"`
	Income       float64                     `json:"income"`
	Expense      float64                     `json:"expense"`
	Net          float64                     `json:"net"`
	Transactions []transactionmodel.ListItem `json:"transactions"`
	Page         int                         `json:"page"`
	PageSize     int                         `json:"page_size"`
	Total        int64                       `json:"total"`
}

// DisplayNameOrEmail — quy tắc định danh FR-009/BR-MBR-009: tên hiển thị, fallback email.
func (m MemberInfo) DisplayNameOrEmail() string {
	if m.DisplayName != "" {
		return m.DisplayName
	}
	return m.Email
}
