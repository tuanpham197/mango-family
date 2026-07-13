package storage

import (
	"context"

	"household-finance/api/module/category/model"

	"github.com/google/uuid"
)

// ListRulesByType — rule của hộ có danh mục cùng loại, ưu tiên match_count (D7).
func (s *SQLStore) ListRulesByType(ctx context.Context, householdID uuid.UUID, typ string) ([]model.RuleSuggestion, error) {
	rules := []model.RuleSuggestion{}
	err := s.db.WithContext(ctx).
		Table("categorization_rules AS r").
		Select("r.keyword, r.category_id, r.match_count, c.name AS category_name, c.is_hidden").
		Joins("JOIN categories c ON c.id = r.category_id").
		Where("r.household_id = ? AND c.type = ?", householdID, typ).
		Order("r.match_count DESC, r.created_at").
		Scan(&rules).Error
	return rules, err
}

// UpsertRule — học từ lịch sử: keyword đã chuẩn hóa, trùng (hộ, keyword, danh mục)
// thì tăng match_count (unique constraint — D7).
func (s *SQLStore) UpsertRule(ctx context.Context, householdID uuid.UUID, keyword string, categoryID uuid.UUID) error {
	return s.db.WithContext(ctx).Exec(`
		INSERT INTO categorization_rules (household_id, keyword, category_id, match_count)
		VALUES (?, ?, ?, 1)
		ON CONFLICT (household_id, keyword, category_id)
		DO UPDATE SET match_count = categorization_rules.match_count + 1`,
		householdID, keyword, categoryID).Error
}
