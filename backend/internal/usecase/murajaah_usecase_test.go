package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
)

// ── fake murajaah repository ──────────────────────────────────────────────────

type fakeMurajaahRepo struct {
	schedules     map[uuid.UUID]*entity.MurajaahSchedule
	existsToday   bool
	createCalls   int
	createdBatch  []*entity.MurajaahSchedule
	logs          []*entity.MurajaahLog
}

func newFakeMurajaahRepo() *fakeMurajaahRepo {
	return &fakeMurajaahRepo{schedules: make(map[uuid.UUID]*entity.MurajaahSchedule)}
}

func (r *fakeMurajaahRepo) CreateSchedules(_ context.Context, schedules []*entity.MurajaahSchedule) error {
	r.createCalls++
	for _, s := range schedules {
		s.ID = uuid.New()
		r.schedules[s.ID] = s
	}
	r.createdBatch = schedules
	return nil
}

func (r *fakeMurajaahRepo) FindTodayByUserID(_ context.Context, userID uuid.UUID, date time.Time) ([]*entity.MurajaahSchedule, error) {
	var out []*entity.MurajaahSchedule
	for _, s := range r.schedules {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *fakeMurajaahRepo) ExistsTodayForUser(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
	return r.existsToday, nil
}

func (r *fakeMurajaahRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.MurajaahSchedule, error) {
	s, ok := r.schedules[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return s, nil
}

func (r *fakeMurajaahRepo) MarkComplete(_ context.Context, id uuid.UUID, completedAt time.Time) error {
	s, ok := r.schedules[id]
	if !ok {
		return errors.New("not found")
	}
	s.CompletedAt = &completedAt
	return nil
}

func (r *fakeMurajaahRepo) FindHistoryByUserID(_ context.Context, userID uuid.UUID, limit, offset int) ([]*entity.MurajaahSchedule, int64, error) {
	var out []*entity.MurajaahSchedule
	for _, s := range r.schedules {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return out, int64(len(out)), nil
}

func (r *fakeMurajaahRepo) CreateLog(_ context.Context, log *entity.MurajaahLog) error {
	r.logs = append(r.logs, log)
	return nil
}

// ── GetToday — lazy generation ────────────────────────────────────────────────

func TestMurajaahGetToday_SkipGenerateIfNoHafalan(t *testing.T) {
	mRepo := newFakeMurajaahRepo()
	hRepo := newFakeHafalanRepo()
	uc := usecase.NewMurajaahUseCase(mRepo, hRepo)

	schedules, err := uc.GetToday(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Empty(t, schedules)
	assert.Equal(t, 0, mRepo.createCalls, "must not call CreateSchedules when user has no hafalan")
}

func TestMurajaahGetToday_GeneratesSabqiAndManzil(t *testing.T) {
	mRepo := newFakeMurajaahRepo()
	hRepo := newFakeHafalanRepo()
	userID := uuid.New()

	// hafalan dicatat hari ini -> masuk sabqi
	hRepo.items[uuid.New()] = &entity.HafalanProgress{
		UserID: userID, SurahNumber: 1, AyatStart: 1, AyatEnd: 7, Status: entity.StatusHafal, NotedAt: time.Now(),
	}
	// hafalan lama (>7 hari) -> hanya masuk manzil
	hRepo.items[uuid.New()] = &entity.HafalanProgress{
		UserID: userID, SurahNumber: 2, AyatStart: 1, AyatEnd: 5, Status: entity.StatusHafal, NotedAt: time.Now().AddDate(0, 0, -30),
	}

	uc := usecase.NewMurajaahUseCase(mRepo, hRepo)
	schedules, err := uc.GetToday(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 1, mRepo.createCalls)
	assert.NotEmpty(t, schedules)

	hasSabqi := false
	for _, s := range mRepo.createdBatch {
		if s.Type == entity.TypeSabqi {
			hasSabqi = true
		}
	}
	assert.True(t, hasSabqi, "recent hafalan must produce a sabqi schedule")
}

func TestMurajaahGetToday_SkipGenerateIfAlreadyExists(t *testing.T) {
	mRepo := newFakeMurajaahRepo()
	mRepo.existsToday = true
	hRepo := newFakeHafalanRepo()
	userID := uuid.New()
	hRepo.items[uuid.New()] = &entity.HafalanProgress{UserID: userID, SurahNumber: 1, AyatStart: 1, AyatEnd: 7, Status: entity.StatusHafal, NotedAt: time.Now()}

	uc := usecase.NewMurajaahUseCase(mRepo, hRepo)
	_, err := uc.GetToday(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 0, mRepo.createCalls, "must not regenerate schedule if one already exists for today")
}

func TestMurajaahGetToday_NoDuplicateBetweenSabqiAndManzil(t *testing.T) {
	mRepo := newFakeMurajaahRepo()
	hRepo := newFakeHafalanRepo()
	userID := uuid.New()
	// single recent hafalan record; should not appear twice (sabqi + manzil) for same range
	hRepo.items[uuid.New()] = &entity.HafalanProgress{UserID: userID, SurahNumber: 1, AyatStart: 1, AyatEnd: 7, Status: entity.StatusHafal, NotedAt: time.Now()}

	uc := usecase.NewMurajaahUseCase(mRepo, hRepo)
	_, err := uc.GetToday(context.Background(), userID)
	require.NoError(t, err)

	type key struct {
		surah, start, end int
	}
	seen := make(map[key]int)
	for _, s := range mRepo.createdBatch {
		seen[key{s.SurahNumber, s.AyatStart, s.AyatEnd}]++
	}
	for k, count := range seen {
		assert.Equal(t, 1, count, "ayat range %+v must not be scheduled twice (sabqi+manzil overlap)", k)
	}
}

// ── MarkComplete ──────────────────────────────────────────────────────────────

func TestMurajaahMarkComplete_NotOwner(t *testing.T) {
	mRepo := newFakeMurajaahRepo()
	hRepo := newFakeHafalanRepo()
	sched := &entity.MurajaahSchedule{ID: uuid.New(), UserID: uuid.New()}
	mRepo.schedules[sched.ID] = sched
	uc := usecase.NewMurajaahUseCase(mRepo, hRepo)

	err := uc.MarkComplete(context.Background(), uuid.New(), sched.ID)
	assert.ErrorIs(t, err, usecase.ErrNotOwner)
}

func TestMurajaahMarkComplete_OK(t *testing.T) {
	mRepo := newFakeMurajaahRepo()
	hRepo := newFakeHafalanRepo()
	userID := uuid.New()
	sched := &entity.MurajaahSchedule{ID: uuid.New(), UserID: userID}
	mRepo.schedules[sched.ID] = sched
	uc := usecase.NewMurajaahUseCase(mRepo, hRepo)

	err := uc.MarkComplete(context.Background(), userID, sched.ID)
	require.NoError(t, err)
	assert.NotNil(t, sched.CompletedAt)
	assert.Len(t, mRepo.logs, 1)
}

func TestMurajaahMarkComplete_AlreadyDoneIsNoop(t *testing.T) {
	mRepo := newFakeMurajaahRepo()
	hRepo := newFakeHafalanRepo()
	userID := uuid.New()
	now := time.Now()
	sched := &entity.MurajaahSchedule{ID: uuid.New(), UserID: userID, CompletedAt: &now}
	mRepo.schedules[sched.ID] = sched
	uc := usecase.NewMurajaahUseCase(mRepo, hRepo)

	err := uc.MarkComplete(context.Background(), userID, sched.ID)
	require.NoError(t, err)
	assert.Empty(t, mRepo.logs, "must not create a duplicate log when already completed")
}

func TestMurajaahMarkComplete_NotFound(t *testing.T) {
	mRepo := newFakeMurajaahRepo()
	hRepo := newFakeHafalanRepo()
	uc := usecase.NewMurajaahUseCase(mRepo, hRepo)

	err := uc.MarkComplete(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

// ── GetHistory pagination defaults ─────────────────────────────────────────────

func TestMurajaahGetHistory_DefaultsInvalidPagination(t *testing.T) {
	mRepo := newFakeMurajaahRepo()
	hRepo := newFakeHafalanRepo()
	userID := uuid.New()
	mRepo.schedules[uuid.New()] = &entity.MurajaahSchedule{UserID: userID}
	uc := usecase.NewMurajaahUseCase(mRepo, hRepo)

	_, total, err := uc.GetHistory(context.Background(), userID, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
}
