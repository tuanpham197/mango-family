package biz

import (
	"context"
	"sort"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/report/model"

	"github.com/google/uuid"
)

// MembersReader — nguồn đọc cho báo cáo theo thành viên (chỉ SELECT/SUM — D42/D43).
type MembersReader interface {
	ListMembers(ctx context.Context, householdID uuid.UUID) ([]model.MemberInfo, error)
	MemberSums(ctx context.Context, householdID uuid.UUID, from, endExcl time.Time) ([]model.MemberAggRow, error)
}

type GetMembersBiz struct{ store MembersReader }

func NewGetMembersBiz(store MembersReader) *GetMembersBiz { return &GetMembersBiz{store: store} }

// Get — GET /api/reports/members (D42–D44): tổng Thu/Chi/ròng mỗi thành viên hiện tại
// (0/0/0 nếu không có giao dịch — FR-004); giao dịch của người đã rời hộ gộp dòng
// "Thành viên cũ" (FR-010); tổng các dòng khớp tổng hộ (FR-005/SC-001). Validate to ≥ from.
func (b *GetMembersBiz) Get(ctx context.Context, householdID uuid.UUID, from, to time.Time) (*model.MembersReport, error) {
	if err := validateRange(from, to); err != nil {
		return nil, err
	}
	endExcl := model.EndExclusive(to)

	members, err := b.store.ListMembers(ctx, householdID)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	sums, err := b.store.MemberSums(ctx, householdID, from, endExcl)
	if err != nil {
		return nil, common.NewInternal(err)
	}

	sumByUser := make(map[uuid.UUID]model.MemberAggRow, len(sums))
	for _, r := range sums {
		sumByUser[r.CreatedBy] = r
	}

	current := make(map[uuid.UUID]bool, len(members))
	rows := make([]model.MemberRow, 0, len(members)+1)
	var totIncome, totExpense float64

	// Thành viên hiện tại — luôn xuất hiện, 0/0/0 nếu không có giao dịch (D43/FR-004).
	for _, m := range members {
		current[m.ID] = true
		agg := sumByUser[m.ID]
		rows = append(rows, model.MemberRow{
			MemberID:    m.ID.String(),
			DisplayName: m.DisplayNameOrEmail(),
			Income:      agg.Income,
			Expense:     agg.Expense,
			Net:         agg.Income - agg.Expense,
		})
		totIncome += agg.Income
		totExpense += agg.Expense
	}

	// Sắp thành viên hiện tại theo Chi giảm dần, tie-break tên (Assumptions/D43).
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Expense != rows[j].Expense {
			return rows[i].Expense > rows[j].Expense
		}
		return rows[i].DisplayName < rows[j].DisplayName
	})

	// Người ĐÃ RỜI hộ — gộp một dòng "Thành viên cũ", chỉ khi có dữ liệu (D44/FR-010).
	var formerIncome, formerExpense float64
	hasFormer := false
	for _, r := range sums {
		if !current[r.CreatedBy] {
			formerIncome += r.Income
			formerExpense += r.Expense
			hasFormer = true
		}
	}
	if hasFormer {
		rows = append(rows, model.MemberRow{
			MemberID:    model.FormerMemberID,
			DisplayName: model.FormerMemberLabel,
			IsFormer:    true,
			Income:      formerIncome,
			Expense:     formerExpense,
			Net:         formerIncome - formerExpense,
		})
		totIncome += formerIncome
		totExpense += formerExpense
	}

	return &model.MembersReport{
		From:    dateStr(from),
		To:      dateStr(to),
		Members: rows,
		Totals:  model.MemberTotals{Income: totIncome, Expense: totExpense, Net: totIncome - totExpense},
	}, nil
}
