package biz

import (
	"context"
	"strings"

	"household-finance/api/common"
	"household-finance/api/module/category/model"

	"github.com/google/uuid"
)

type RuleLister interface {
	ListRulesByType(ctx context.Context, householdID uuid.UUID, typ string) ([]model.RuleSuggestion, error)
}

type SuggestCategoryBiz struct{ store RuleLister }

func NewSuggestCategoryBiz(store RuleLister) *SuggestCategoryBiz {
	return &SuggestCategoryBiz{store: store}
}

// NormalizeKeyword — chuẩn hóa mô tả/keyword: lowercase + gọn khoảng trắng (D7).
func NormalizeKeyword(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}

// Suggest — GET /api/categories/suggest?description=&type= (FR-015, D7):
// keyword chứa-trong-mô-tả, cùng hộ + cùng loại, ưu tiên match_count cao nhất;
// danh mục ẩn không được gợi ý cho giao dịch mới (FR-020). Không match → null.
// Luôn chỉ mang tính tham khảo — người dùng ghi đè được (Clarifications 2026-06-24).
func (b *SuggestCategoryBiz) Suggest(ctx context.Context, householdID uuid.UUID, description, typ string) (*model.Suggestion, error) {
	if typ != common.TypeIncome && typ != common.TypeExpense {
		return nil, common.NewBadRequest("type phải là INCOME hoặc EXPENSE").WithField("type")
	}
	desc := NormalizeKeyword(description)
	if desc == "" {
		return nil, nil
	}
	rules, err := b.store.ListRulesByType(ctx, householdID, typ)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	for _, r := range rules { // đã sắp theo match_count DESC
		if r.IsHidden || r.Keyword == "" {
			continue
		}
		if strings.Contains(desc, r.Keyword) {
			return &model.Suggestion{CategoryID: r.CategoryID, Name: r.CategoryName}, nil
		}
	}
	return nil, nil
}
