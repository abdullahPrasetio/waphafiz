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

type hafalanRepository struct {
	db *gorm.DB
}

func NewHafalanRepository(db *gorm.DB) domainrepo.HafalanRepository {
	return &hafalanRepository{db: db}
}

func (r *hafalanRepository) Create(ctx context.Context, h *entity.HafalanProgress) error {
	if err := r.db.WithContext(ctx).Create(h).Error; err != nil {
		return fmt.Errorf("create hafalan: %w", err)
	}
	return nil
}

func (r *hafalanRepository) Update(ctx context.Context, h *entity.HafalanProgress) error {
	if err := r.db.WithContext(ctx).Save(h).Error; err != nil {
		return fmt.Errorf("update hafalan: %w", err)
	}
	return nil
}

func (r *hafalanRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.HafalanProgress, error) {
	var h entity.HafalanProgress
	err := r.db.WithContext(ctx).First(&h, "id = ? AND deleted_at IS NULL", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find hafalan by id: %w", err)
	}
	return &h, nil
}

func (r *hafalanRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.HafalanProgress, error) {
	var list []*entity.HafalanProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("surah_number ASC, ayat_start ASC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("find hafalan by user: %w", err)
	}
	return list, nil
}

func (r *hafalanRepository) FindByUserAndSurah(ctx context.Context, userID uuid.UUID, surahNumber int) ([]*entity.HafalanProgress, error) {
	var list []*entity.HafalanProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND surah_number = ? AND deleted_at IS NULL", userID, surahNumber).
		Order("ayat_start ASC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("find hafalan by user and surah: %w", err)
	}
	return list, nil
}

func (r *hafalanRepository) FindOverlapping(ctx context.Context, userID uuid.UUID, surahNumber, ayatStart, ayatEnd int) ([]*entity.HafalanProgress, error) {
	var list []*entity.HafalanProgress
	// Two ranges [a1,a2] and [b1,b2] overlap when a1 <= b2 AND a2 >= b1
	err := r.db.WithContext(ctx).
		Where(`user_id = ? AND surah_number = ? AND deleted_at IS NULL
			AND ayat_start <= ? AND ayat_end >= ?`,
			userID, surahNumber, ayatEnd, ayatStart).
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("find overlapping hafalan: %w", err)
	}
	return list, nil
}

func (r *hafalanRepository) FindByUserSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]*entity.HafalanProgress, error) {
	var list []*entity.HafalanProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND noted_at >= ? AND deleted_at IS NULL", userID, since).
		Order("noted_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("find hafalan since: %w", err)
	}
	return list, nil
}

func (r *hafalanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&entity.HafalanProgress{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error; err != nil {
		return fmt.Errorf("delete hafalan: %w", err)
	}
	return nil
}

func (r *hafalanRepository) FindByFamilyGroup(ctx context.Context, familyGroupID uuid.UUID) ([]*entity.HafalanProgress, error) {
	var list []*entity.HafalanProgress
	err := r.db.WithContext(ctx).
		Joins("JOIN users ON users.id = hafalan_progress.user_id").
		Where("users.family_group_id = ? AND hafalan_progress.deleted_at IS NULL AND users.deleted_at IS NULL", familyGroupID).
		Preload("User").
		Order("hafalan_progress.surah_number ASC, hafalan_progress.ayat_start ASC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("find hafalan by family: %w", err)
	}
	return list, nil
}
