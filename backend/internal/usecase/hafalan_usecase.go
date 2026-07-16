package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	domainrepo "github.com/abdullahPrasetio/waphafiz/internal/domain/repository"
)

var (
	ErrAyatOverlap    = errors.New("rentang ayat overlap dengan hafalan yang sudah ada")
	ErrAyatInvalid    = errors.New("ayat_start harus lebih kecil atau sama dengan ayat_end")
	ErrNotOwner       = errors.New("bukan milik user ini")
)

type CreateHafalanRequest struct {
	SurahNumber int    `json:"surah_number" validate:"required,min=1,max=114"`
	AyatStart   int    `json:"ayat_start"   validate:"required,min=1"`
	AyatEnd     int    `json:"ayat_end"     validate:"required,min=1"`
	Status      string `json:"status"       validate:"required,oneof=belum sedang hafal"`
}

type UpdateHafalanRequest struct {
	Status string `json:"status" validate:"required,oneof=belum sedang hafal"`
}

type HafalanSummary struct {
	TotalAyat    int     `json:"total_ayat"`
	HafalAyat    int     `json:"hafal_ayat"`
	SedangAyat   int     `json:"sedang_ayat"`
	TotalSurah   int     `json:"total_surah"`
	HafalSurah   int     `json:"hafal_surah"`
	PercentHafal float64 `json:"percent_hafal"`
}

type MemberStatusCounts struct {
	Hafal  int `json:"hafal"`
	Sedang int `json:"sedang"`
	Belum  int `json:"belum"`
}

type RecentHafalanItem struct {
	ID          uuid.UUID `json:"id"`
	SurahNumber int       `json:"surah_number"`
	AyatStart   int       `json:"ayat_start"`
	AyatEnd     int       `json:"ayat_end"`
}

type MemberProgressResult struct {
	UserID        uuid.UUID          `json:"user_id"`
	Name          string             `json:"name"`
	Email         string             `json:"email"`
	TotalAyat     int                `json:"total_ayat"`
	TotalSurah    int                `json:"total_surah"`
	TotalJuz      float64            `json:"total_juz"`
	ActiveToday   bool               `json:"active_today"`
	LastSeenAt    *time.Time         `json:"last_seen_at"`
	StatusCounts  MemberStatusCounts `json:"status_counts"`
	RecentHafalan []RecentHafalanItem `json:"recent_hafalan"`
}

type HafalanUseCase interface {
	Create(ctx context.Context, userID uuid.UUID, req *CreateHafalanRequest) (*entity.HafalanProgress, error)
	List(ctx context.Context, userID uuid.UUID) ([]*entity.HafalanProgress, error)
	Update(ctx context.Context, userID uuid.UUID, hafalanID uuid.UUID, req *UpdateHafalanRequest) (*entity.HafalanProgress, error)
	Delete(ctx context.Context, userID uuid.UUID, hafalanID uuid.UUID) error
	Summary(ctx context.Context, userID uuid.UUID) (*HafalanSummary, error)
	ListByFamilyGroup(ctx context.Context, familyGroupID uuid.UUID) ([]*MemberProgressResult, error)
	AdminListByMember(ctx context.Context, memberID uuid.UUID) ([]*entity.HafalanProgress, error)
	AdminUpdate(ctx context.Context, hafalanID uuid.UUID, req *UpdateHafalanRequest) (*entity.HafalanProgress, error)
	AdminDelete(ctx context.Context, hafalanID uuid.UUID) error
}

type hafalanUseCase struct {
	repo domainrepo.HafalanRepository
}

func NewHafalanUseCase(repo domainrepo.HafalanRepository) HafalanUseCase {
	return &hafalanUseCase{repo: repo}
}

func (u *hafalanUseCase) Create(ctx context.Context, userID uuid.UUID, req *CreateHafalanRequest) (*entity.HafalanProgress, error) {
	if req.AyatStart > req.AyatEnd {
		return nil, ErrAyatInvalid
	}

	overlaps, err := u.repo.FindOverlapping(ctx, userID, req.SurahNumber, req.AyatStart, req.AyatEnd)
	if err != nil {
		return nil, fmt.Errorf("check overlap: %w", err)
	}
	if len(overlaps) > 0 {
		return nil, ErrAyatOverlap
	}

	h := &entity.HafalanProgress{
		UserID:      userID,
		SurahNumber: req.SurahNumber,
		AyatStart:   req.AyatStart,
		AyatEnd:     req.AyatEnd,
		Status:      entity.HafalanStatus(req.Status),
		NotedAt:     time.Now(),
	}
	if err := u.repo.Create(ctx, h); err != nil {
		return nil, fmt.Errorf("create hafalan: %w", err)
	}
	return h, nil
}

func (u *hafalanUseCase) List(ctx context.Context, userID uuid.UUID) ([]*entity.HafalanProgress, error) {
	list, err := u.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list hafalan: %w", err)
	}
	return list, nil
}

func (u *hafalanUseCase) Update(ctx context.Context, userID, hafalanID uuid.UUID, req *UpdateHafalanRequest) (*entity.HafalanProgress, error) {
	h, err := u.repo.FindByID(ctx, hafalanID)
	if err != nil {
		return nil, ErrNotFound
	}
	if h.UserID != userID {
		return nil, ErrNotOwner
	}

	h.Status = entity.HafalanStatus(req.Status)
	if err := u.repo.Update(ctx, h); err != nil {
		return nil, fmt.Errorf("update hafalan: %w", err)
	}
	return h, nil
}

func (u *hafalanUseCase) AdminListByMember(ctx context.Context, memberID uuid.UUID) ([]*entity.HafalanProgress, error) {
	list, err := u.repo.FindByUserID(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("admin list hafalan by member: %w", err)
	}
	return list, nil
}

func (u *hafalanUseCase) Delete(ctx context.Context, userID, hafalanID uuid.UUID) error {
	h, err := u.repo.FindByID(ctx, hafalanID)
	if err != nil {
		return ErrNotFound
	}
	if h.UserID != userID {
		return ErrNotOwner
	}
	return u.repo.Delete(ctx, hafalanID)
}

func (u *hafalanUseCase) AdminUpdate(ctx context.Context, hafalanID uuid.UUID, req *UpdateHafalanRequest) (*entity.HafalanProgress, error) {
	h, err := u.repo.FindByID(ctx, hafalanID)
	if err != nil {
		return nil, ErrNotFound
	}
	h.Status = entity.HafalanStatus(req.Status)
	if err := u.repo.Update(ctx, h); err != nil {
		return nil, fmt.Errorf("admin update hafalan: %w", err)
	}
	return h, nil
}

func (u *hafalanUseCase) AdminDelete(ctx context.Context, hafalanID uuid.UUID) error {
	if _, err := u.repo.FindByID(ctx, hafalanID); err != nil {
		return ErrNotFound
	}
	return u.repo.Delete(ctx, hafalanID)
}

func (u *hafalanUseCase) Summary(ctx context.Context, userID uuid.UUID) (*HafalanSummary, error) {
	list, err := u.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("summary hafalan: %w", err)
	}

	summary := &HafalanSummary{}
	surahMap := make(map[int]bool)

	for _, h := range list {
		ayat := h.AyatEnd - h.AyatStart + 1
		summary.TotalAyat += ayat
		if h.Status == entity.StatusHafal {
			summary.HafalAyat += ayat
		} else if h.Status == entity.StatusSedang {
			summary.SedangAyat += ayat
		}
		surahMap[h.SurahNumber] = true
	}
	summary.TotalSurah = len(surahMap)

	if summary.TotalAyat > 0 {
		summary.PercentHafal = float64(summary.HafalAyat) / float64(summary.TotalAyat) * 100
	}
	return summary, nil
}

func (u *hafalanUseCase) ListByFamilyGroup(ctx context.Context, familyGroupID uuid.UUID) ([]*MemberProgressResult, error) {
	list, err := u.repo.FindByFamilyGroup(ctx, familyGroupID)
	if err != nil {
		return nil, fmt.Errorf("list hafalan by family: %w", err)
	}

	today := time.Now()
	todayStr := today.Format("2006-01-02")

	// group hafalan by user_id
	type userBucket struct {
		name   string
		email  string
		items  []*entity.HafalanProgress
	}
	// preserve order of first appearance
	order := []uuid.UUID{}
	buckets := map[uuid.UUID]*userBucket{}

	for _, h := range list {
		if _, ok := buckets[h.UserID]; !ok {
			name, email := "", ""
			if h.User != nil {
				name = h.User.Name
				email = h.User.Email
			}
			buckets[h.UserID] = &userBucket{name: name, email: email}
			order = append(order, h.UserID)
		}
		buckets[h.UserID].items = append(buckets[h.UserID].items, h)
	}

	const ayatPerJuz = 208.0
	results := make([]*MemberProgressResult, 0, len(order))

	for _, uid := range order {
		b := buckets[uid]
		res := &MemberProgressResult{
			UserID: uid,
			Name:   b.name,
			Email:  b.email,
		}
		surahSet := map[int]bool{}
		var lastSeen *time.Time

		for i, h := range b.items {
			ayat := h.AyatEnd - h.AyatStart + 1
			res.TotalAyat += ayat
			switch h.Status {
			case entity.StatusHafal:
				res.StatusCounts.Hafal += ayat
			case entity.StatusSedang:
				res.StatusCounts.Sedang += ayat
			default:
				res.StatusCounts.Belum += ayat
			}
			surahSet[h.SurahNumber] = true

			if lastSeen == nil || h.NotedAt.After(*lastSeen) {
				t := h.NotedAt
				lastSeen = &t
			}
			if h.NotedAt.Format("2006-01-02") == todayStr {
				res.ActiveToday = true
			}

			// last 3 items (list is already ordered by surah/ayat; take first 3 for recent)
			if i < 3 {
				res.RecentHafalan = append(res.RecentHafalan, RecentHafalanItem{
					ID:          h.ID,
					SurahNumber: h.SurahNumber,
					AyatStart:   h.AyatStart,
					AyatEnd:     h.AyatEnd,
				})
			}
		}

		res.TotalSurah = len(surahSet)
		res.TotalJuz = float64(res.TotalAyat) / ayatPerJuz
		res.LastSeenAt = lastSeen
		results = append(results, res)
	}

	return results, nil
}
