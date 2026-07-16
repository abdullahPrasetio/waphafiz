package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	domainrepo "github.com/abdullahPrasetio/waphafiz/internal/domain/repository"
)

type murajaahRepository struct {
	db *gorm.DB
}

func NewMurajaahRepository(db *gorm.DB) domainrepo.MurajaahRepository {
	return &murajaahRepository{db: db}
}

func (r *murajaahRepository) CreateSchedules(ctx context.Context, schedules []*entity.MurajaahSchedule) error {
	if len(schedules) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&schedules).Error; err != nil {
		return fmt.Errorf("create murajaah schedules: %w", err)
	}
	return nil
}

func (r *murajaahRepository) FindTodayByUserID(ctx context.Context, userID uuid.UUID, date time.Time) ([]*entity.MurajaahSchedule, error) {
	var schedules []*entity.MurajaahSchedule
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND scheduled_date = ?", userID, date.Format("2006-01-02")).
		Order("type ASC, surah_number ASC, ayat_start ASC").
		Find(&schedules).Error
	if err != nil {
		return nil, fmt.Errorf("find today schedules: %w", err)
	}
	return schedules, nil
}

func (r *murajaahRepository) ExistsTodayForUser(ctx context.Context, userID uuid.UUID, date time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.MurajaahSchedule{}).
		Where("user_id = ? AND scheduled_date = ?", userID, date.Format("2006-01-02")).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check today schedule: %w", err)
	}
	return count > 0, nil
}

func (r *murajaahRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.MurajaahSchedule, error) {
	var s entity.MurajaahSchedule
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find schedule by id: %w", err)
	}
	return &s, nil
}

func (r *murajaahRepository) MarkComplete(ctx context.Context, id uuid.UUID, completedAt time.Time) error {
	err := r.db.WithContext(ctx).
		Model(&entity.MurajaahSchedule{}).
		Where("id = ?", id).
		Update("completed_at", completedAt).Error
	if err != nil {
		return fmt.Errorf("mark complete: %w", err)
	}
	return nil
}

func (r *murajaahRepository) FindHistoryByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.MurajaahSchedule, int64, error) {
	var schedules []*entity.MurajaahSchedule
	var total int64

	q := r.db.WithContext(ctx).Model(&entity.MurajaahSchedule{}).Where("user_id = ? AND completed_at IS NOT NULL", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count history: %w", err)
	}

	err := q.Order("scheduled_date DESC").Limit(limit).Offset(offset).Find(&schedules).Error
	if err != nil {
		return nil, 0, fmt.Errorf("find history: %w", err)
	}
	return schedules, total, nil
}

func (r *murajaahRepository) CreateLog(ctx context.Context, log *entity.MurajaahLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("create murajaah log: %w", err)
	}
	return nil
}
