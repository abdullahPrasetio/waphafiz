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

type familyGroupRepository struct {
	db *gorm.DB
}

func NewFamilyGroupRepository(db *gorm.DB) domainrepo.FamilyGroupRepository {
	return &familyGroupRepository{db: db}
}

func (r *familyGroupRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.FamilyGroup, error) {
	var fg entity.FamilyGroup
	err := r.db.WithContext(ctx).First(&fg, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find family group by id: %w", err)
	}
	return &fg, nil
}

func (r *familyGroupRepository) FindByInviteCode(ctx context.Context, code string) (*entity.FamilyGroup, error) {
	var fg entity.FamilyGroup
	err := r.db.WithContext(ctx).First(&fg, "invite_code = ? AND deleted_at IS NULL", code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find family group by invite code: %w", err)
	}
	return &fg, nil
}

func (r *familyGroupRepository) Create(ctx context.Context, fg *entity.FamilyGroup) error {
	if err := r.db.WithContext(ctx).Create(fg).Error; err != nil {
		return fmt.Errorf("create family group: %w", err)
	}
	return nil
}

func (r *familyGroupRepository) UpdateInviteCode(ctx context.Context, id uuid.UUID, code string) error {
	err := r.db.WithContext(ctx).
		Model(&entity.FamilyGroup{}).
		Where("id = ?", id).
		Updates(map[string]any{"invite_code": code, "updated_at": time.Now()}).Error
	if err != nil {
		return fmt.Errorf("update invite code: %w", err)
	}
	return nil
}

func (r *familyGroupRepository) ListMembers(ctx context.Context, familyGroupID uuid.UUID) ([]*entity.User, error) {
	var users []*entity.User
	err := r.db.WithContext(ctx).
		Where("family_group_id = ? AND deleted_at IS NULL", familyGroupID).
		Order("created_at ASC").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	return users, nil
}

func (r *familyGroupRepository) UpdateMemberStatus(ctx context.Context, userID uuid.UUID, isActive bool) error {
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{"is_active": isActive, "updated_at": time.Now()}).Error
	if err != nil {
		return fmt.Errorf("update member status: %w", err)
	}
	return nil
}

func (r *familyGroupRepository) UpdateMemberLastSeen(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{"last_seen_at": now, "updated_at": now}).Error
	if err != nil {
		return fmt.Errorf("update last seen: %w", err)
	}
	return nil
}
