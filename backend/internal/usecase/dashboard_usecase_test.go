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

func TestDashboardGetMy_NoHafalanNoMurajaah(t *testing.T) {
	hRepo := newFakeHafalanRepo()
	fRepo := newFakeFamilyGroupRepo()
	mRepo := newFakeMurajaahRepo()
	uc := usecase.NewDashboardUseCase(hRepo, fRepo, mRepo)

	dash, err := uc.GetMyDashboard(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, 0, dash.TotalHafalAyat)
	assert.Equal(t, float64(0), dash.PercentHafal)
	assert.Equal(t, 0, dash.Streak)
}

func TestDashboardGetMy_CalculatesPercentAndMurajaahCount(t *testing.T) {
	hRepo := newFakeHafalanRepo()
	fRepo := newFakeFamilyGroupRepo()
	mRepo := newFakeMurajaahRepo()
	userID := uuid.New()

	hRepo.items[uuid.New()] = &entity.HafalanProgress{UserID: userID, SurahNumber: 1, AyatStart: 1, AyatEnd: 7, Status: entity.StatusHafal}
	hRepo.items[uuid.New()] = &entity.HafalanProgress{UserID: userID, SurahNumber: 2, AyatStart: 1, AyatEnd: 3, Status: entity.StatusSedang}

	now := time.Now()
	mRepo.schedules[uuid.New()] = &entity.MurajaahSchedule{UserID: userID, CompletedAt: &now}
	mRepo.schedules[uuid.New()] = &entity.MurajaahSchedule{UserID: userID, CompletedAt: nil}

	uc := usecase.NewDashboardUseCase(hRepo, fRepo, mRepo)
	dash, err := uc.GetMyDashboard(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 7, dash.TotalHafalAyat)
	assert.Equal(t, 1, dash.TotalHafalSurah)
	assert.InDelta(t, 70.0, dash.PercentHafal, 0.01)
	assert.Equal(t, 2, dash.MurajaahTotal)
	assert.Equal(t, 1, dash.MurajaahDone)
}

func TestDashboardGetAdmin_CountsActiveTodayAndHafalAyat(t *testing.T) {
	hRepo := newFakeHafalanRepo()
	fRepo := newFakeFamilyGroupRepo()
	mRepo := newFakeMurajaahRepo()
	fgID := uuid.New()

	activeUser := uuid.New()
	inactiveUser := uuid.New()
	now := time.Now()
	old := time.Now().AddDate(0, 0, -10)

	fRepo.members[fgID] = []*entity.User{
		{ID: activeUser, Name: "Aktif", IsActive: true, LastSeenAt: &now},
		{ID: inactiveUser, Name: "Tidak Aktif", IsActive: true, LastSeenAt: &old},
	}
	hRepo.items[uuid.New()] = &entity.HafalanProgress{UserID: activeUser, SurahNumber: 1, AyatStart: 1, AyatEnd: 10, Status: entity.StatusHafal}

	uc := usecase.NewDashboardUseCase(hRepo, fRepo, mRepo)
	dash, err := uc.GetAdminDashboard(context.Background(), fgID)
	require.NoError(t, err)
	assert.Equal(t, 2, dash.TotalMembers)
	assert.Equal(t, 1, dash.ActiveToday)

	var activeProgress *usecase.MemberProgress
	for _, mp := range dash.MemberProgress {
		if mp.UserID == activeUser {
			activeProgress = mp
		}
	}
	require.NotNil(t, activeProgress)
	assert.Equal(t, 10, activeProgress.HafalAyat)
}
