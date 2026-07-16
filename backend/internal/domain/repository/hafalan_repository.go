package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
)

type HafalanRepository interface {
	Create(ctx context.Context, h *entity.HafalanProgress) error
	Update(ctx context.Context, h *entity.HafalanProgress) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.HafalanProgress, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.HafalanProgress, error)
	FindByUserAndSurah(ctx context.Context, userID uuid.UUID, surahNumber int) ([]*entity.HafalanProgress, error)
	FindOverlapping(ctx context.Context, userID uuid.UUID, surahNumber, ayatStart, ayatEnd int) ([]*entity.HafalanProgress, error)
	FindByUserSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]*entity.HafalanProgress, error)
	FindByFamilyGroup(ctx context.Context, familyGroupID uuid.UUID) ([]*entity.HafalanProgress, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
