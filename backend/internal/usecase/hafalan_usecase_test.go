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

// ── fake repository ──────────────────────────────────────────────────────────

type fakeHafalanRepo struct {
	items       map[uuid.UUID]*entity.HafalanProgress
	overlapping []*entity.HafalanProgress
	createErr   error
}

func newFakeHafalanRepo() *fakeHafalanRepo {
	return &fakeHafalanRepo{items: make(map[uuid.UUID]*entity.HafalanProgress)}
}

func (r *fakeHafalanRepo) Create(_ context.Context, h *entity.HafalanProgress) error {
	if r.createErr != nil {
		return r.createErr
	}
	h.ID = uuid.New()
	r.items[h.ID] = h
	return nil
}

func (r *fakeHafalanRepo) Update(_ context.Context, h *entity.HafalanProgress) error {
	r.items[h.ID] = h
	return nil
}

func (r *fakeHafalanRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.HafalanProgress, error) {
	h, ok := r.items[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return h, nil
}

func (r *fakeHafalanRepo) FindByUserID(_ context.Context, userID uuid.UUID) ([]*entity.HafalanProgress, error) {
	var out []*entity.HafalanProgress
	for _, h := range r.items {
		if h.UserID == userID {
			out = append(out, h)
		}
	}
	return out, nil
}

func (r *fakeHafalanRepo) FindByUserAndSurah(_ context.Context, userID uuid.UUID, surahNumber int) ([]*entity.HafalanProgress, error) {
	var out []*entity.HafalanProgress
	for _, h := range r.items {
		if h.UserID == userID && h.SurahNumber == surahNumber {
			out = append(out, h)
		}
	}
	return out, nil
}

func (r *fakeHafalanRepo) FindOverlapping(_ context.Context, _ uuid.UUID, _, _, _ int) ([]*entity.HafalanProgress, error) {
	return r.overlapping, nil
}

func (r *fakeHafalanRepo) FindByUserSince(_ context.Context, userID uuid.UUID, since time.Time) ([]*entity.HafalanProgress, error) {
	var out []*entity.HafalanProgress
	for _, h := range r.items {
		if h.UserID == userID && !h.NotedAt.Before(since) {
			out = append(out, h)
		}
	}
	return out, nil
}

func (r *fakeHafalanRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.items, id)
	return nil
}

func (r *fakeHafalanRepo) FindByFamilyGroup(_ context.Context, _ uuid.UUID) ([]*entity.HafalanProgress, error) {
	var out []*entity.HafalanProgress
	for _, h := range r.items {
		out = append(out, h)
	}
	return out, nil
}

// ── Create ───────────────────────────────────────────────────────────────────

func TestHafalanCreate_OK(t *testing.T) {
	repo := newFakeHafalanRepo()
	uc := usecase.NewHafalanUseCase(repo)
	userID := uuid.New()

	h, err := uc.Create(context.Background(), userID, &usecase.CreateHafalanRequest{
		SurahNumber: 1, AyatStart: 1, AyatEnd: 7, Status: "hafal",
	})
	require.NoError(t, err)
	assert.Equal(t, userID, h.UserID)
	assert.Equal(t, entity.StatusHafal, h.Status)
}

func TestHafalanCreate_InvalidAyatRange(t *testing.T) {
	repo := newFakeHafalanRepo()
	uc := usecase.NewHafalanUseCase(repo)

	_, err := uc.Create(context.Background(), uuid.New(), &usecase.CreateHafalanRequest{
		SurahNumber: 1, AyatStart: 10, AyatEnd: 5, Status: "belum",
	})
	assert.ErrorIs(t, err, usecase.ErrAyatInvalid)
}

func TestHafalanCreate_OverlapRejected(t *testing.T) {
	repo := newFakeHafalanRepo()
	repo.overlapping = []*entity.HafalanProgress{{ID: uuid.New(), SurahNumber: 1, AyatStart: 1, AyatEnd: 5}}
	uc := usecase.NewHafalanUseCase(repo)

	_, err := uc.Create(context.Background(), uuid.New(), &usecase.CreateHafalanRequest{
		SurahNumber: 1, AyatStart: 3, AyatEnd: 8, Status: "sedang",
	})
	assert.ErrorIs(t, err, usecase.ErrAyatOverlap)
}

// ── Update ───────────────────────────────────────────────────────────────────

func TestHafalanUpdate_NotOwner(t *testing.T) {
	repo := newFakeHafalanRepo()
	owner := uuid.New()
	h := &entity.HafalanProgress{ID: uuid.New(), UserID: owner, SurahNumber: 2, AyatStart: 1, AyatEnd: 5, Status: entity.StatusBelum}
	repo.items[h.ID] = h
	uc := usecase.NewHafalanUseCase(repo)

	_, err := uc.Update(context.Background(), uuid.New(), h.ID, &usecase.UpdateHafalanRequest{Status: "hafal"})
	assert.ErrorIs(t, err, usecase.ErrNotOwner)
}

func TestHafalanUpdate_OK(t *testing.T) {
	repo := newFakeHafalanRepo()
	owner := uuid.New()
	h := &entity.HafalanProgress{ID: uuid.New(), UserID: owner, SurahNumber: 2, AyatStart: 1, AyatEnd: 5, Status: entity.StatusBelum}
	repo.items[h.ID] = h
	uc := usecase.NewHafalanUseCase(repo)

	updated, err := uc.Update(context.Background(), owner, h.ID, &usecase.UpdateHafalanRequest{Status: "hafal"})
	require.NoError(t, err)
	assert.Equal(t, entity.StatusHafal, updated.Status)
}

func TestHafalanUpdate_NotFound(t *testing.T) {
	repo := newFakeHafalanRepo()
	uc := usecase.NewHafalanUseCase(repo)

	_, err := uc.Update(context.Background(), uuid.New(), uuid.New(), &usecase.UpdateHafalanRequest{Status: "hafal"})
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

// ── Summary ──────────────────────────────────────────────────────────────────

func TestHafalanSummary_CalculatesPercentAndCounts(t *testing.T) {
	repo := newFakeHafalanRepo()
	userID := uuid.New()
	repo.items[uuid.New()] = &entity.HafalanProgress{UserID: userID, SurahNumber: 1, AyatStart: 1, AyatEnd: 7, Status: entity.StatusHafal}
	repo.items[uuid.New()] = &entity.HafalanProgress{UserID: userID, SurahNumber: 2, AyatStart: 1, AyatEnd: 3, Status: entity.StatusSedang}
	uc := usecase.NewHafalanUseCase(repo)

	summary, err := uc.Summary(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 10, summary.TotalAyat)
	assert.Equal(t, 7, summary.HafalAyat)
	assert.Equal(t, 3, summary.SedangAyat)
	assert.Equal(t, 2, summary.TotalSurah)
	assert.InDelta(t, 70.0, summary.PercentHafal, 0.01)
}

func TestHafalanSummary_EmptyHafalan(t *testing.T) {
	repo := newFakeHafalanRepo()
	uc := usecase.NewHafalanUseCase(repo)

	summary, err := uc.Summary(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, 0, summary.TotalAyat)
	assert.Equal(t, float64(0), summary.PercentHafal)
}

// ── List / ListByFamilyGroup ──────────────────────────────────────────────────

func TestHafalanList_OnlyReturnsOwnRecords(t *testing.T) {
	repo := newFakeHafalanRepo()
	userID := uuid.New()
	repo.items[uuid.New()] = &entity.HafalanProgress{UserID: userID, SurahNumber: 1, AyatStart: 1, AyatEnd: 5}
	repo.items[uuid.New()] = &entity.HafalanProgress{UserID: uuid.New(), SurahNumber: 2, AyatStart: 1, AyatEnd: 5}
	uc := usecase.NewHafalanUseCase(repo)

	list, err := uc.List(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, userID, list[0].UserID)
}

func TestHafalanListByFamilyGroup_ReturnsAll(t *testing.T) {
	repo := newFakeHafalanRepo()
	repo.items[uuid.New()] = &entity.HafalanProgress{UserID: uuid.New(), SurahNumber: 1, AyatStart: 1, AyatEnd: 5}
	repo.items[uuid.New()] = &entity.HafalanProgress{UserID: uuid.New(), SurahNumber: 2, AyatStart: 1, AyatEnd: 5}
	uc := usecase.NewHafalanUseCase(repo)

	list, err := uc.ListByFamilyGroup(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}
