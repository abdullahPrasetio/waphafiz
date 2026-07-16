package usecase_test

// Integration test for the critical flow: register family -> login ->
// catat hafalan -> get muraja'ah today. Runs against a real PostgreSQL
// instance (no DB mocking, per project rule in CLAUDE.md).
//
// Requires: `make docker-up` (or an already-running postgres on
// DB_HOST/DB_PORT). Skips automatically if the DB is unreachable, so this
// file does not break `go test ./...` in environments without docker.

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	dbrepo "github.com/abdullahPrasetio/waphafiz/internal/repository/db"
	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func openIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=disable password='%s'",
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "5432"),
		getenv("DB_USER", "temancode"),
		getenv("DB_NAME", "waphafiz"),
		getenv("DB_PASSWORD", ""),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("skipping integration test: cannot connect to postgres: %v", err)
	}
	sqlDB, err := db.DB()
	require.NoError(t, err)
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("skipping integration test: postgres not reachable: %v", err)
	}
	require.NoError(t, db.AutoMigrate(
		&entity.FamilyGroup{},
		&entity.User{},
		&entity.HafalanProgress{},
		&entity.MurajaahSchedule{},
		&entity.MurajaahLog{},
	))
	return db
}

// cleanupRows deletes rows created by this test run, scoped to the given user/family IDs,
// so repeated test runs don't accumulate data or collide on unique constraints.
func cleanupRows(t *testing.T, db *gorm.DB, userIDs []uuid.UUID, familyIDs []uuid.UUID) {
	t.Helper()
	t.Cleanup(func() {
		db.Unscoped().Where("user_id IN ?", userIDs).Delete(&entity.MurajaahLog{})
		db.Unscoped().Where("user_id IN ?", userIDs).Delete(&entity.MurajaahSchedule{})
		db.Unscoped().Where("user_id IN ?", userIDs).Delete(&entity.HafalanProgress{})
		db.Unscoped().Where("id IN ?", userIDs).Delete(&entity.User{})
		db.Unscoped().Where("id IN ?", familyIDs).Delete(&entity.FamilyGroup{})
	})
}

func TestIntegration_RegisterLoginHafalanMurajaah(t *testing.T) {
	db := openIntegrationDB(t)

	userRepo := dbrepo.NewUserRepository(db)
	familyRepo := dbrepo.NewFamilyGroupRepository(db)
	hafalanRepo := dbrepo.NewHafalanRepository(db)
	murajaahRepo := dbrepo.NewMurajaahRepository(db)

	jwtCfg := &auth.Config{
		Secret:   "integration-test-secret-key-32-bytes-min!",
		Issuer:   "waphafiz-test",
		Audience: "api",
		Expiry:   time.Hour,
	}

	authUC := usecase.NewAuthUseCase(userRepo, familyRepo, jwtCfg)
	hafalanUC := usecase.NewHafalanUseCase(hafalanRepo)
	murajaahUC := usecase.NewMurajaahUseCase(murajaahRepo, hafalanRepo)

	ctx := context.Background()
	uniqueEmail := fmt.Sprintf("integration-%s@test.com", uuid.New().String()[:8])

	// 1. Register a new family (creates admin user + family group + invite code)
	registerResp, err := authUC.RegisterFamily(ctx, &usecase.RegisterFamilyRequest{
		FamilyName: "Keluarga Integration Test",
		Name:       "Admin Test",
		Email:      uniqueEmail,
		Password:   "password123",
	})
	require.NoError(t, err)
	require.NotEmpty(t, registerResp.Token)
	userID := registerResp.User.ID
	familyID := registerResp.User.FamilyGroupID
	cleanupRows(t, db, []uuid.UUID{userID}, []uuid.UUID{familyID})

	// 2. Login with the same credentials
	loginResp, err := authUC.Login(ctx, &usecase.LoginRequest{
		Email:    uniqueEmail,
		Password: "password123",
	})
	require.NoError(t, err)
	require.NotEmpty(t, loginResp.Token)
	require.Equal(t, userID, loginResp.User.ID)

	// 3. Catat hafalan baru
	hafalan, err := hafalanUC.Create(ctx, userID, &usecase.CreateHafalanRequest{
		SurahNumber: 1, AyatStart: 1, AyatEnd: 7, Status: "hafal",
	})
	require.NoError(t, err)
	require.Equal(t, userID, hafalan.UserID)

	// Creating an overlapping range must be rejected end-to-end through the real repo.
	_, err = hafalanUC.Create(ctx, userID, &usecase.CreateHafalanRequest{
		SurahNumber: 1, AyatStart: 3, AyatEnd: 10, Status: "sedang",
	})
	require.ErrorIs(t, err, usecase.ErrAyatOverlap)

	// 4. Get muraja'ah hari ini -> lazy generation harus membuat jadwal dari hafalan di atas
	schedules, err := murajaahUC.GetToday(ctx, userID)
	require.NoError(t, err)
	require.NotEmpty(t, schedules, "murajaah schedule must be generated from the hafalan just recorded")

	found := false
	for _, s := range schedules {
		if s.SurahNumber == 1 && s.AyatStart == 1 && s.AyatEnd == 7 {
			found = true
		}
	}
	require.True(t, found, "today's schedule must include the newly recorded hafalan")

	// Calling GetToday again must not duplicate the schedule (idempotent lazy generation).
	schedulesAgain, err := murajaahUC.GetToday(ctx, userID)
	require.NoError(t, err)
	require.Len(t, schedulesAgain, len(schedules), "second call must not regenerate/duplicate today's schedule")

	// 5. Mark one schedule as complete and verify it persists
	err = murajaahUC.MarkComplete(ctx, userID, schedules[0].ID)
	require.NoError(t, err)

	history, total, err := murajaahUC.GetHistory(ctx, userID, 1, 20)
	require.NoError(t, err)
	require.GreaterOrEqual(t, total, int64(1))
	completedCount := 0
	for _, h := range history {
		if h.CompletedAt != nil {
			completedCount++
		}
	}
	require.GreaterOrEqual(t, completedCount, 1)
}
