package storage

import (
	"context"
	"errors"

	"household-finance/api/module/household/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNoHousehold = errors.New("user has no household")

type SQLStore struct{ db *gorm.DB }

func NewSQLStore(db *gorm.DB) *SQLStore { return &SQLStore{db: db} }

// HouseholdOfUser — hộ mà user là thành viên (MVP: mỗi user một hộ).
func (s *SQLStore) HouseholdOfUser(ctx context.Context, userID uuid.UUID) (*model.Household, error) {
	var h model.Household
	err := s.db.WithContext(ctx).
		Joins("JOIN household_members m ON m.household_id = households.id").
		Where("m.user_id = ?", userID).
		First(&h).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoHousehold
		}
		return nil, err
	}
	return &h, nil
}

func (s *SQLStore) Create(ctx context.Context, h *model.Household) error {
	return s.db.WithContext(ctx).Create(h).Error
}

func (s *SQLStore) AddMember(ctx context.Context, householdID, userID uuid.UUID) error {
	return s.db.WithContext(ctx).
		Where(model.Member{HouseholdID: householdID, UserID: userID}).
		FirstOrCreate(&model.Member{HouseholdID: householdID, UserID: userID}).Error
}
