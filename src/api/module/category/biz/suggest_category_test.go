package biz

import (
	"context"
	"testing"

	"household-finance/api/common"
	"household-finance/api/module/category/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRuleLister struct{ rules []model.RuleSuggestion }

func (m *mockRuleLister) ListRulesByType(_ context.Context, _ uuid.UUID, _ string) ([]model.RuleSuggestion, error) {
	return m.rules, nil
}

func TestNormalizeKeyword(t *testing.T) {
	assert.Equal(t, "grab đi làm", NormalizeKeyword("  Grab   Đi   LÀM  "))
	assert.Equal(t, "", NormalizeKeyword("   "))
}

func TestSuggest_MatchesKeywordInDescription(t *testing.T) {
	catID := uuid.New()
	biz := NewSuggestCategoryBiz(&mockRuleLister{rules: []model.RuleSuggestion{
		{Keyword: "grab", CategoryID: catID, CategoryName: "Di chuyển"},
	}})
	s, err := biz.Suggest(context.Background(), uuid.New(), "đặt Grab về quê", common.TypeExpense)
	require.NoError(t, err)
	require.NotNil(t, s)
	assert.Equal(t, catID, s.CategoryID)
	assert.Equal(t, "Di chuyển", s.Name)
}

func TestSuggest_PrefersHigherMatchCountOrder(t *testing.T) {
	first, second := uuid.New(), uuid.New()
	// store trả về theo match_count DESC — biz lấy rule khớp ĐẦU TIÊN
	biz := NewSuggestCategoryBiz(&mockRuleLister{rules: []model.RuleSuggestion{
		{Keyword: "cà phê", CategoryID: first, CategoryName: "Ăn uống", MatchCount: 9},
		{Keyword: "cà phê", CategoryID: second, CategoryName: "Giải trí", MatchCount: 1},
	}})
	s, err := biz.Suggest(context.Background(), uuid.New(), "cà phê sáng", common.TypeExpense)
	require.NoError(t, err)
	require.NotNil(t, s)
	assert.Equal(t, first, s.CategoryID)
}

func TestSuggest_SkipsHiddenCategory(t *testing.T) {
	visible := uuid.New()
	biz := NewSuggestCategoryBiz(&mockRuleLister{rules: []model.RuleSuggestion{
		{Keyword: "cà phê", CategoryID: uuid.New(), CategoryName: "Ẩn rồi", IsHidden: true, MatchCount: 9},
		{Keyword: "cà phê", CategoryID: visible, CategoryName: "Ăn uống", MatchCount: 1},
	}})
	s, err := biz.Suggest(context.Background(), uuid.New(), "cà phê sáng", common.TypeExpense)
	require.NoError(t, err)
	require.NotNil(t, s)
	assert.Equal(t, visible, s.CategoryID) // FR-020: ẩn không gợi ý cho giao dịch mới
}

func TestSuggest_NoMatchReturnsNil(t *testing.T) {
	biz := NewSuggestCategoryBiz(&mockRuleLister{rules: []model.RuleSuggestion{
		{Keyword: "grab", CategoryID: uuid.New(), CategoryName: "Di chuyển"},
	}})
	s, err := biz.Suggest(context.Background(), uuid.New(), "mua sách", common.TypeExpense)
	require.NoError(t, err)
	assert.Nil(t, s)

	s, err = biz.Suggest(context.Background(), uuid.New(), "   ", common.TypeExpense)
	require.NoError(t, err)
	assert.Nil(t, s) // mô tả rỗng → không gợi ý
}

func TestSuggest_InvalidType(t *testing.T) {
	biz := NewSuggestCategoryBiz(&mockRuleLister{})
	_, err := biz.Suggest(context.Background(), uuid.New(), "grab", "WRONG")
	assert.Equal(t, common.ErrCodeInvalidRequest, appErrCode(t, err))
}
