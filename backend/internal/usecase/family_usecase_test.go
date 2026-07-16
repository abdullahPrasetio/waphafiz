package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
)

func TestFamilyListMembers_FormatsLastSeen(t *testing.T) {
	repo := newFakeFamilyGroupRepo()
	fgID := uuid.New()
	now := time.Now()
	repo.members[fgID] = []*entity.User{
		{ID: uuid.New(), Name: "Aisyah", Email: "aisyah@test.com", Role: entity.RoleMember, IsActive: true, LastSeenAt: &now},
		{ID: uuid.New(), Name: "Bilal", Email: "bilal@test.com", Role: entity.RoleAdmin, IsActive: false},
	}
	uc := usecase.NewFamilyUseCase(repo)

	members, err := uc.ListMembers(context.Background(), fgID)
	require.NoError(t, err)
	require.Len(t, members, 2)
	assert.NotNil(t, members[0].LastSeen)
	assert.Nil(t, members[1].LastSeen)
}

func TestFamilyUpdateMemberStatus_NotInFamily(t *testing.T) {
	repo := newFakeFamilyGroupRepo()
	fgID := uuid.New()
	repo.members[fgID] = []*entity.User{{ID: uuid.New()}}
	uc := usecase.NewFamilyUseCase(repo)

	err := uc.UpdateMemberStatus(context.Background(), fgID, uuid.New(), false)
	assert.ErrorIs(t, err, usecase.ErrMemberNotInFamily)
}

func TestFamilyUpdateMemberStatus_OK(t *testing.T) {
	repo := newFakeFamilyGroupRepo()
	fgID := uuid.New()
	memberID := uuid.New()
	repo.members[fgID] = []*entity.User{{ID: memberID}}
	uc := usecase.NewFamilyUseCase(repo)

	err := uc.UpdateMemberStatus(context.Background(), fgID, memberID, false)
	require.NoError(t, err)
}

func TestFamilyRegenerateInviteCode_ReplacesOldCode(t *testing.T) {
	repo := newFakeFamilyGroupRepo()
	fg := seedFamily(t, repo, "FAM-OLD1")
	uc := usecase.NewFamilyUseCase(repo)

	newCode, err := uc.RegenerateInviteCode(context.Background(), fg.ID)
	require.NoError(t, err)
	assert.NotEqual(t, "FAM-OLD1", newCode)

	_, err = repo.FindByInviteCode(context.Background(), "FAM-OLD1")
	assert.Error(t, err, "old invite code must no longer resolve")

	found, err := repo.FindByInviteCode(context.Background(), newCode)
	require.NoError(t, err)
	assert.Equal(t, fg.ID, found.ID)
}
