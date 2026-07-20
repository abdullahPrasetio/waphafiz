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

type shareRepository struct {
	db *gorm.DB
}

func NewShareRepository(db *gorm.DB) domainrepo.ShareRepository {
	return &shareRepository{db: db}
}

func (r *shareRepository) Create(ctx context.Context, s *entity.DashboardShare) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return fmt.Errorf("create share: %w", err)
	}
	return nil
}

func (r *shareRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.DashboardShare, error) {
	var s entity.DashboardShare
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find share by id: %w", err)
	}
	return &s, nil
}

func (r *shareRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.DashboardShare, error) {
	var s entity.DashboardShare
	err := r.db.WithContext(ctx).First(&s, "token_hash = ?", tokenHash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find share by token hash: %w", err)
	}
	return &s, nil
}

func (r *shareRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.DashboardShare, error) {
	var list []*entity.DashboardShare
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list shares by user: %w", err)
	}
	return list, nil
}

func (r *shareRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID, now time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.DashboardShare{}).
		Where("user_id = ? AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at > ?)", userID, now).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count active shares: %w", err)
	}
	return count, nil
}

func (r *shareRepository) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	err := r.db.WithContext(ctx).
		Model(&entity.DashboardShare{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", revokedAt).Error
	if err != nil {
		return fmt.Errorf("revoke share: %w", err)
	}
	return nil
}

func (r *shareRepository) PurgeInactiveBefore(ctx context.Context, userID uuid.UUID, before time.Time) error {
	err := r.db.WithContext(ctx).
		Unscoped().
		Where("user_id = ? AND ((revoked_at IS NOT NULL AND revoked_at < ?) OR (expires_at IS NOT NULL AND expires_at < ?))",
			userID, before, before).
		Delete(&entity.DashboardShare{}).Error
	if err != nil {
		return fmt.Errorf("purge inactive shares: %w", err)
	}
	return nil
}

func (r *shareRepository) TouchView(ctx context.Context, id uuid.UUID, viewedAt time.Time) error {
	err := r.db.WithContext(ctx).
		Model(&entity.DashboardShare{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_viewed_at": viewedAt,
			"view_count":     gorm.Expr("view_count + 1"),
		}).Error
	if err != nil {
		return fmt.Errorf("touch share view: %w", err)
	}
	return nil
}
