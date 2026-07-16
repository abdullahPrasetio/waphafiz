package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	domainrepo "github.com/abdullahPrasetio/waphafiz/internal/domain/repository"
)

var ErrMemberNotInFamily = errors.New("member tidak ada di keluarga ini")

type MemberResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	IsActive bool      `json:"is_active"`
	LastSeen *string   `json:"last_seen"`
}

type FamilyUseCase interface {
	ListMembers(ctx context.Context, familyGroupID uuid.UUID) ([]*MemberResponse, error)
	UpdateMemberStatus(ctx context.Context, familyGroupID, memberID uuid.UUID, isActive bool) error
	RegenerateInviteCode(ctx context.Context, familyGroupID uuid.UUID) (string, error)
}

type familyUseCase struct {
	familyRepo domainrepo.FamilyGroupRepository
}

func NewFamilyUseCase(familyRepo domainrepo.FamilyGroupRepository) FamilyUseCase {
	return &familyUseCase{familyRepo: familyRepo}
}

func (u *familyUseCase) ListMembers(ctx context.Context, familyGroupID uuid.UUID) ([]*MemberResponse, error) {
	members, err := u.familyRepo.ListMembers(ctx, familyGroupID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	result := make([]*MemberResponse, 0, len(members))
	for _, m := range members {
		mr := &MemberResponse{
			ID:       m.ID,
			Name:     m.Name,
			Email:    m.Email,
			Role:     string(m.Role),
			IsActive: m.IsActive,
		}
		if m.LastSeenAt != nil {
			s := m.LastSeenAt.Format("2006-01-02T15:04:05Z07:00")
			mr.LastSeen = &s
		}
		result = append(result, mr)
	}
	return result, nil
}

func (u *familyUseCase) UpdateMemberStatus(ctx context.Context, familyGroupID, memberID uuid.UUID, isActive bool) error {
	members, err := u.familyRepo.ListMembers(ctx, familyGroupID)
	if err != nil {
		return fmt.Errorf("list members: %w", err)
	}

	found := false
	for _, m := range members {
		if m.ID == memberID {
			found = true
			break
		}
	}
	if !found {
		return ErrMemberNotInFamily
	}

	if err := u.familyRepo.UpdateMemberStatus(ctx, memberID, isActive); err != nil {
		return fmt.Errorf("update member status: %w", err)
	}
	return nil
}

func (u *familyUseCase) RegenerateInviteCode(ctx context.Context, familyGroupID uuid.UUID) (string, error) {
	code, err := generateInviteCode()
	if err != nil {
		return "", fmt.Errorf("generate code: %w", err)
	}

	if err := u.familyRepo.UpdateInviteCode(ctx, familyGroupID, code); err != nil {
		return "", fmt.Errorf("update invite code: %w", err)
	}
	return code, nil
}

// GetFamilyByID is a helper for handlers that need family info.
func GetFamilyByID(ctx context.Context, repo domainrepo.FamilyGroupRepository, id uuid.UUID) (*entity.FamilyGroup, error) {
	return repo.FindByID(ctx, id)
}
