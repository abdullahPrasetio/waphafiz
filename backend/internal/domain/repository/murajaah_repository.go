package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
)

type MurajaahRepository interface {
	CreateSchedules(ctx context.Context, schedules []*entity.MurajaahSchedule) error
	FindTodayByUserID(ctx context.Context, userID uuid.UUID, date time.Time) ([]*entity.MurajaahSchedule, error)
	ExistsTodayForUser(ctx context.Context, userID uuid.UUID, date time.Time) (bool, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.MurajaahSchedule, error)
	MarkComplete(ctx context.Context, id uuid.UUID, completedAt time.Time) error
	FindHistoryByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.MurajaahSchedule, int64, error)

	CreateLog(ctx context.Context, log *entity.MurajaahLog) error
}
