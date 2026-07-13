package storage

import (
	"context"
	"errors"

	"household-finance/api/module/user/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("user not found")

type SQLStore struct{ db *gorm.DB }

func NewSQLStore(db *gorm.DB) *SQLStore { return &SQLStore{db: db} }

func (s *SQLStore) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *SQLStore) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
