package common

import (
	"time"

	"github.com/google/uuid"
)

// SQLModel — base cho mọi GORM entity (id uuid do Postgres sinh).
type SQLModel struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
