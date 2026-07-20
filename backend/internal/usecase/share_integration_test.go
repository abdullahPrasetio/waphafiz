package usecase_test

// Integration test alur share dashboard: register -> catat hafalan ->
// buat share -> view publik via token -> angka cocok dengan dashboard
// pribadi -> revoke -> view gagal. Real PostgreSQL, tanpa mock DB.

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	dbrepo "github.com/abdullahPrasetio/waphafiz/internal/repository/db"
	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

func TestIntegration_ShareDashboardFlow(t *testing.T) {
	db := openIntegrationDB(t)

	userRepo := dbrepo.NewUserRepository(db)
	familyRepo := dbrepo.NewFamilyGroupRepository(db)
	hafalanRepo := dbrepo.NewHafalanRepository(db)
	murajaahRepo := dbrepo.NewMurajaahRepository(db)
	shareRepo := dbrepo.NewShareRepository(db)

	jwtCfg := &auth.Config{
		Secret:   "integration-test-secret-key-32-bytes-min!",
		Issuer:   "waphafiz-test",
		Audience: "api",
		Expiry:   time.Hour,
	}
	authUC := usecase.NewAuthUseCase(userRepo, familyRepo, jwtCfg)
	hafalanUC := usecase.NewHafalanUseCase(hafalanRepo)
	dashboardUC := usecase.NewDashboardUseCase(hafalanRepo, familyRepo, murajaahRepo)
	// surahNames nil: nama surah kosong di payload, tapi alur tetap jalan.
	// cache nil: tiap view hit DB langsung — cukup untuk test.
	shareUC := usecase.NewShareUseCase(shareRepo, userRepo, hafalanRepo, murajaahRepo, nil, nil)

	ctx := context.Background()
	email := fmt.Sprintf("share-int-%s@test.com", uuid.New().String()[:8])

	reg, err := authUC.RegisterFamily(ctx, &usecase.RegisterFamilyRequest{
		FamilyName: "Keluarga Share Test",
		Name:       "Ayah Share",
		Email:      email,
		Password:   "password123",
	})
	require.NoError(t, err)
	userID := reg.User.ID
	familyID := reg.User.FamilyGroupID
	cleanupRows(t, db, []uuid.UUID{userID}, []uuid.UUID{familyID})

	_, err = hafalanUC.Create(ctx, userID, &usecase.CreateHafalanRequest{
		SurahNumber: 67, AyatStart: 1, AyatEnd: 30, Status: "hafal",
	})
	require.NoError(t, err)

	// Buat share untuk diri sendiri (role admin karena pembuat keluarga).
	created, err := shareUC.CreateShare(ctx, userID, "admin", familyID, usecase.CreateShareInput{
		Label:         "Ustadz Integration",
		ExpiresInDays: 30,
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.Token)

	// View publik tanpa auth apapun — hanya bermodal token.
	dash, err := shareUC.ViewShared(ctx, created.Token)
	require.NoError(t, err)
	require.Equal(t, "Ayah Share", dash.MemberName)

	// Angka di halaman share harus identik dengan dashboard pribadi.
	my, err := dashboardUC.GetMyDashboard(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, my.TotalHafalAyat, dash.Stats.TotalHafalAyat)
	require.Equal(t, my.TotalHafalSurah, dash.Stats.TotalHafalSurah)
	require.InDelta(t, my.PercentHafal, dash.Stats.PercentHafal, 0.001)

	// View tercatat di DB.
	stored, err := shareRepo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, 1, stored.ViewCount)

	// Revoke → view publik langsung mati.
	require.NoError(t, shareUC.RevokeShare(ctx, userID, "admin", familyID, created.ID))
	_, err = shareUC.ViewShared(ctx, created.Token)
	require.ErrorIs(t, err, usecase.ErrShareNotFound)
}
