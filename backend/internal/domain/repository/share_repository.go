package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
)

type ShareRepository interface {
	Create(ctx context.Context, s *entity.DashboardShare) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.DashboardShare, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (*entity.DashboardShare, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.DashboardShare, error)
	CountActiveByUserID(ctx context.Context, userID uuid.UUID, now time.Time) (int64, error)
	Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	TouchView(ctx context.Context, id uuid.UUID, viewedAt time.Time) error
	// PurgeInactiveBefore menghapus permanen share milik userID yang sudah
	// tidak aktif (dicabut atau kedaluwarsa) sejak sebelum waktu 'before'.
	PurgeInactiveBefore(ctx context.Context, userID uuid.UUID, before time.Time) error
}
