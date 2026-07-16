package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
)

type FamilyGroupRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entity.FamilyGroup, error)
	FindByInviteCode(ctx context.Context, code string) (*entity.FamilyGroup, error)
	Create(ctx context.Context, fg *entity.FamilyGroup) error
	UpdateInviteCode(ctx context.Context, id uuid.UUID, code string) error

	ListMembers(ctx context.Context, familyGroupID uuid.UUID) ([]*entity.User, error)
	UpdateMemberStatus(ctx context.Context, userID uuid.UUID, isActive bool) error
	UpdateMemberLastSeen(ctx context.Context, userID uuid.UUID) error
}
