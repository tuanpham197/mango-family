package storage

import (
	"context"
	"errors"
	"time"

	"household-finance/api/module/category/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("category not found")

type SQLStore struct{ db *gorm.DB }

func NewSQLStore(db *gorm.DB) *SQLStore { return &SQLStore{db: db} }

type ListFilter struct {
	Type          string // "" = cả hai loại
	IncludeHidden bool
}

// List — toàn bộ danh mục của hộ theo filter (FR-002/014/018/020).
func (s *SQLStore) List(ctx context.Context, householdID uuid.UUID, f ListFilter) ([]model.Category, error) {
	q := s.db.WithContext(ctx).Where("household_id = ?", householdID)
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	if !f.IncludeHidden {
		q = q.Where("is_hidden = false")
	}
	var cats []model.Category
	if err := q.Order("type, name").Find(&cats).Error; err != nil {
		return nil, err
	}
	return cats, nil
}

// FindByID — luôn scope theo hộ; ngoài hộ trả ErrNotFound (→ 404, research D5).
func (s *SQLStore) FindByID(ctx context.Context, householdID, id uuid.UUID) (*model.Category, error) {
	var c model.Category
	err := s.db.WithContext(ctx).
		Where("household_id = ? AND id = ?", householdID, id).
		First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (s *SQLStore) Create(ctx context.Context, c *model.Category) error {
	return s.db.WithContext(ctx).Create(c).Error
}

// CountDuplicateName — trùng tên trong (hộ, loại, cùng cấp cha) — FR-017.
func (s *SQLStore) CountDuplicateName(ctx context.Context, householdID uuid.UUID, name, typ string, parentID, excludeID *uuid.UUID) (int64, error) {
	q := s.db.WithContext(ctx).Model(&model.Category{}).
		Where("household_id = ? AND lower(name) = lower(?) AND type = ?", householdID, name, typ)
	if parentID == nil {
		q = q.Where("parent_id IS NULL")
	} else {
		q = q.Where("parent_id = ?", *parentID)
	}
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var n int64
	err := q.Count(&n).Error
	return n, err
}

func (s *SQLStore) ListChildren(ctx context.Context, householdID, parentID uuid.UUID) ([]model.Category, error) {
	var cats []model.Category
	err := s.db.WithContext(ctx).
		Where("household_id = ? AND parent_id = ?", householdID, parentID).
		Find(&cats).Error
	return cats, err
}

// UpdateConditional — mốc lạc quan (D6): UPDATE ... WHERE updated_at = expected.
// Trả về số hàng bị ảnh hưởng; 0 = conflict hoặc đã xóa (biz phân biệt).
func (s *SQLStore) UpdateConditional(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt time.Time, changes map[string]any) (int64, error) {
	res := s.db.WithContext(ctx).Model(&model.Category{}).
		Where("household_id = ? AND id = ? AND updated_at = ?", householdID, id, expectedUpdatedAt).
		Updates(changes)
	return res.RowsAffected, res.Error
}

// DeleteConditional — DELETE ... WHERE updated_at = expected (D6).
func (s *SQLStore) DeleteConditional(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt time.Time) (int64, error) {
	res := s.db.WithContext(ctx).
		Where("household_id = ? AND id = ? AND updated_at = ?", householdID, id, expectedUpdatedAt).
		Delete(&model.Category{})
	return res.RowsAffected, res.Error
}

func (s *SQLStore) DeleteByIDs(ctx context.Context, householdID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).
		Where("household_id = ? AND id IN ?", householdID, ids).
		Delete(&model.Category{}).Error
}

// --- Thao tác trên bảng liên quan trong luồng delete-reassign (D12) ---

func (s *SQLStore) CountTransactionsByCategories(ctx context.Context, householdID uuid.UUID, ids []uuid.UUID) (int64, error) {
	var n int64
	err := s.db.WithContext(ctx).Table("transactions").
		Where("household_id = ? AND category_id IN ?", householdID, ids).
		Count(&n).Error
	return n, err
}

func (s *SQLStore) ReassignTransactions(ctx context.Context, householdID uuid.UUID, fromIDs []uuid.UUID, toID uuid.UUID) error {
	return s.db.WithContext(ctx).Table("transactions").
		Where("household_id = ? AND category_id IN ?", householdID, fromIDs).
		Update("category_id", toID).Error
}

func (s *SQLStore) DeleteTransactionsByCategories(ctx context.Context, householdID uuid.UUID, ids []uuid.UUID) error {
	return s.db.WithContext(ctx).Table("transactions").
		Where("household_id = ? AND category_id IN ?", householdID, ids).
		Delete(nil).Error
}

func (s *SQLStore) DeleteRulesByCategories(ctx context.Context, householdID uuid.UUID, ids []uuid.UUID) error {
	return s.db.WithContext(ctx).Table("categorization_rules").
		Where("household_id = ? AND category_id IN ?", householdID, ids).
		Delete(nil).Error
}

// InTx — chạy toàn bộ luồng delete-reassign trong MỘT DB transaction (D3/D12).
func (s *SQLStore) InTx(ctx context.Context, fn func(txStore *SQLStore) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewSQLStore(tx))
	})
}
