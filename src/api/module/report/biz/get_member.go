package biz

import (
	"context"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/report/model"
	transactionmodel "household-finance/api/module/transaction/model"

	"github.com/google/uuid"
)

// MemberReader — nguồn đọc cho drill-in giao dịch của một thành viên (D47/D48).
type MemberReader interface {
	ListMembers(ctx context.Context, householdID uuid.UUID) ([]model.MemberInfo, error)
	FindMember(ctx context.Context, householdID, userID uuid.UUID) (model.MemberInfo, bool, error)
	MemberSums(ctx context.Context, householdID uuid.UUID, from, endExcl time.Time) ([]model.MemberAggRow, error)
	MemberTransactions(ctx context.Context, householdID, userID uuid.UUID, from, endExcl time.Time, paging common.Paging) ([]transactionmodel.ListItem, int64, error)
	FormerMemberTransactions(ctx context.Context, householdID uuid.UUID, currentIDs []uuid.UUID, from, endExcl time.Time, paging common.Paging) ([]transactionmodel.ListItem, int64, error)
}

type GetMemberBiz struct{ store MemberReader }

func NewGetMemberBiz(store MemberReader) *GetMemberBiz { return &GetMemberBiz{store: store} }

// Get — GET /api/reports/member/:id (D47/D48): danh sách giao dịch (phân trang) của một
// thành viên hoặc bucket "former" trong khoảng + tổng Thu/Chi/ròng của thành viên đó.
// `idParam` = UUID (phải là thành viên hiện tại của hộ) hoặc sentinel "former"; ngược lại 404.
func (b *GetMemberBiz) Get(ctx context.Context, householdID uuid.UUID, idParam string, from, to time.Time, paging common.Paging) (*model.MemberReport, error) {
	if err := validateRange(from, to); err != nil {
		return nil, err
	}
	endExcl := model.EndExclusive(to)

	out := &model.MemberReport{
		From:     dateStr(from),
		To:       dateStr(to),
		Page:     paging.Page,
		PageSize: paging.PageSize,
	}

	if idParam == model.FormerMemberID {
		members, err := b.store.ListMembers(ctx, householdID)
		if err != nil {
			return nil, common.NewInternal(err)
		}
		currentIDs, current := memberIDSet(members)

		sums, err := b.store.MemberSums(ctx, householdID, from, endExcl)
		if err != nil {
			return nil, common.NewInternal(err)
		}
		for _, r := range sums {
			if !current[r.CreatedBy] {
				out.Income += r.Income
				out.Expense += r.Expense
			}
		}
		txns, total, err := b.store.FormerMemberTransactions(ctx, householdID, currentIDs, from, endExcl, paging)
		if err != nil {
			return nil, common.NewInternal(err)
		}
		out.MemberID = model.FormerMemberID
		out.DisplayName = model.FormerMemberLabel
		out.IsFormer = true
		out.Transactions = txns
		out.Net = out.Income - out.Expense
		out.Total = total
		return out, nil
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		return nil, common.NewNotFound("thành viên không tồn tại")
	}
	info, found, err := b.store.FindMember(ctx, householdID, id)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	if !found {
		return nil, common.NewNotFound("thành viên không tồn tại")
	}

	sums, err := b.store.MemberSums(ctx, householdID, from, endExcl)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	for _, r := range sums {
		if r.CreatedBy == id {
			out.Income = r.Income
			out.Expense = r.Expense
			break
		}
	}
	txns, total, err := b.store.MemberTransactions(ctx, householdID, id, from, endExcl, paging)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	out.MemberID = id.String()
	out.DisplayName = info.DisplayNameOrEmail()
	out.Transactions = txns
	out.Net = out.Income - out.Expense
	out.Total = total
	return out, nil
}

func memberIDSet(members []model.MemberInfo) ([]uuid.UUID, map[uuid.UUID]bool) {
	ids := make([]uuid.UUID, 0, len(members))
	set := make(map[uuid.UUID]bool, len(members))
	for _, m := range members {
		ids = append(ids, m.ID)
		set[m.ID] = true
	}
	return ids, set
}
