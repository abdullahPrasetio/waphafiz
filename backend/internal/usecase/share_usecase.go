package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	domainrepo "github.com/abdullahPrasetio/waphafiz/internal/domain/repository"
	"github.com/abdullahPrasetio/waphafiz/pkg/quranapi"
)

var (
	// ErrShareNotFound dipakai untuk SEMUA kegagalan view publik (token salah,
	// dicabut, kedaluwarsa, user nonaktif) — sengaja tidak dibedakan supaya
	// tidak memberi petunjuk ke penebak token.
	ErrShareNotFound     = errors.New("link tidak ditemukan atau sudah tidak berlaku")
	ErrShareForbidden    = errors.New("tidak berhak mengelola link pantau anggota ini")
	ErrShareLimitReached = errors.New("jumlah link pantau aktif sudah mencapai batas maksimum")
	ErrShareUserNotFound = errors.New("anggota tidak ditemukan")
)

const (
	maxActiveSharesPerUser = 10
	shareTokenBytes        = 32
	shareViewCacheTTL      = 5 * time.Minute
	lastActivityLimit      = 7
	// shareRetentionDays adalah lama share yang sudah dicabut/kedaluwarsa
	// dipertahankan sebagai audit trail sebelum dihapus permanen. Cukup
	// panjang untuk "siapa saja yang pernah punya akses baru-baru ini",
	// tapi tetap membatasi pertumbuhan tabel tanpa batas.
	shareRetentionDays = 180
)

type CreateShareInput struct {
	UserID        *uuid.UUID `json:"user_id"`         // nil = diri sendiri
	Label         string     `json:"label"            validate:"required,max=100"`
	ExpiresInDays int        `json:"expires_in_days"` // 0 = tanpa kedaluwarsa
}

type CreateShareResult struct {
	ID        uuid.UUID  `json:"id"`
	Label     string     `json:"label"`
	ExpiresAt *time.Time `json:"expires_at"`
	Token     string     `json:"token"` // hanya muncul sekali, di sini
}

type ShareInfo struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	UserName     string     `json:"user_name"`
	Label        string     `json:"label"`
	ExpiresAt    *time.Time `json:"expires_at"`
	LastViewedAt *time.Time `json:"last_viewed_at"`
	ViewCount    int        `json:"view_count"`
	Status       string     `json:"status"` // aktif | kedaluwarsa | dicabut
}

type PublicStats struct {
	TotalHafalAyat  int     `json:"total_hafal_ayat"`
	TotalHafalSurah int     `json:"total_hafal_surah"`
	PercentHafal    float64 `json:"percent_hafal"`
	Streak          int     `json:"streak"`
}

type PublicMurajaahToday struct {
	Done  int `json:"done"`
	Total int `json:"total"`
}

type PublicSurahProgress struct {
	SurahNumber int    `json:"surah_number"`
	SurahName   string `json:"surah_name"`
	AyatHafal   int    `json:"ayat_hafal"`
	AyatTotal   int    `json:"ayat_total"`
	Status      string `json:"status"`
}

type PublicActivity struct {
	Date   string `json:"date"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
	Status string `json:"status"`
}

// PublicDashboard adalah payload halaman share publik. Didefinisikan eksplisit
// (bukan reuse struct internal) supaya penambahan field internal di masa depan
// tidak otomatis bocor ke publik. Tidak boleh memuat email, user_id, last_seen,
// atau data anggota lain.
type PublicDashboard struct {
	MemberName     string                 `json:"member_name"`
	Label          string                 `json:"label"`
	Stats          PublicStats            `json:"stats"`
	MurajaahToday  PublicMurajaahToday    `json:"murajaah_today"`
	SurahProgress  []*PublicSurahProgress `json:"surah_progress"`
	LastActivity   []*PublicActivity      `json:"last_activity"`
	GeneratedAt    time.Time              `json:"generated_at"`
}

// SurahNameProvider menyediakan metadata surah (nama + jumlah ayat).
// Diimplementasikan oleh pkg/quranapi.Client (dengan cache Redis 24 jam).
type SurahNameProvider interface {
	GetSurahList(ctx context.Context) ([]quranapi.SurahMeta, error)
}

type ShareUseCase interface {
	CreateShare(ctx context.Context, requesterID uuid.UUID, requesterRole string, requesterFamilyID uuid.UUID, input CreateShareInput) (*CreateShareResult, error)
	ListShares(ctx context.Context, requesterID uuid.UUID, requesterRole string, requesterFamilyID uuid.UUID, targetUserID *uuid.UUID) ([]*ShareInfo, error)
	RevokeShare(ctx context.Context, requesterID uuid.UUID, requesterRole string, requesterFamilyID uuid.UUID, shareID uuid.UUID) error
	ViewShared(ctx context.Context, token string) (*PublicDashboard, error)
}

type shareUseCase struct {
	shareRepo    domainrepo.ShareRepository
	userRepo     domainrepo.UserRepository
	hafalanRepo  domainrepo.HafalanRepository
	murajaahRepo domainrepo.MurajaahRepository
	surahNames   SurahNameProvider
	cache        domainrepo.Cacher
}

func NewShareUseCase(
	shareRepo domainrepo.ShareRepository,
	userRepo domainrepo.UserRepository,
	hafalanRepo domainrepo.HafalanRepository,
	murajaahRepo domainrepo.MurajaahRepository,
	surahNames SurahNameProvider,
	cache domainrepo.Cacher,
) ShareUseCase {
	return &shareUseCase{
		shareRepo:    shareRepo,
		userRepo:     userRepo,
		hafalanRepo:  hafalanRepo,
		murajaahRepo: murajaahRepo,
		surahNames:   surahNames,
		cache:        cache,
	}
}

// resolveTarget memastikan requester berhak mengelola share milik targetUserID:
// dirinya sendiri, atau admin terhadap anggota satu keluarga.
func (u *shareUseCase) resolveTarget(ctx context.Context, requesterID uuid.UUID, requesterRole string, requesterFamilyID, targetUserID uuid.UUID) (*entity.User, error) {
	target, err := u.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, ErrShareUserNotFound
	}
	if targetUserID == requesterID {
		return target, nil
	}
	if requesterRole != string(entity.RoleAdmin) {
		return nil, ErrShareForbidden
	}
	if target.FamilyGroupID == nil || requesterFamilyID == uuid.Nil || *target.FamilyGroupID != requesterFamilyID {
		return nil, ErrShareForbidden
	}
	return target, nil
}

func (u *shareUseCase) CreateShare(ctx context.Context, requesterID uuid.UUID, requesterRole string, requesterFamilyID uuid.UUID, input CreateShareInput) (*CreateShareResult, error) {
	targetID := requesterID
	if input.UserID != nil {
		targetID = *input.UserID
	}
	if _, err := u.resolveTarget(ctx, requesterID, requesterRole, requesterFamilyID, targetID); err != nil {
		return nil, err
	}

	now := time.Now()
	active, err := u.shareRepo.CountActiveByUserID(ctx, targetID, now)
	if err != nil {
		return nil, fmt.Errorf("count active shares: %w", err)
	}
	if active >= maxActiveSharesPerUser {
		return nil, ErrShareLimitReached
	}

	token, tokenHash, err := generateShareToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	share := &entity.DashboardShare{
		UserID:    targetID,
		CreatedBy: requesterID,
		TokenHash: tokenHash,
		Label:     input.Label,
	}
	if input.ExpiresInDays > 0 {
		exp := now.AddDate(0, 0, input.ExpiresInDays)
		share.ExpiresAt = &exp
	}

	if err := u.shareRepo.Create(ctx, share); err != nil {
		return nil, fmt.Errorf("create share: %w", err)
	}

	return &CreateShareResult{
		ID:        share.ID,
		Label:     share.Label,
		ExpiresAt: share.ExpiresAt,
		Token:     token,
	}, nil
}

func (u *shareUseCase) ListShares(ctx context.Context, requesterID uuid.UUID, requesterRole string, requesterFamilyID uuid.UUID, targetUserID *uuid.UUID) ([]*ShareInfo, error) {
	targetID := requesterID
	if targetUserID != nil {
		targetID = *targetUserID
	}
	target, err := u.resolveTarget(ctx, requesterID, requesterRole, requesterFamilyID, targetID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	// Pembersihan lazy: sama seperti lazy generation muraja'ah, tidak perlu
	// cron terpisah. Best-effort — kegagalan purge tidak boleh menggagalkan
	// permintaan lihat daftar.
	_ = u.shareRepo.PurgeInactiveBefore(ctx, targetID, now.AddDate(0, 0, -shareRetentionDays))

	shares, err := u.shareRepo.ListByUserID(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("list shares: %w", err)
	}

	infos := make([]*ShareInfo, 0, len(shares))
	for _, s := range shares {
		infos = append(infos, &ShareInfo{
			ID:           s.ID,
			UserID:       s.UserID,
			UserName:     target.Name,
			Label:        s.Label,
			ExpiresAt:    s.ExpiresAt,
			LastViewedAt: s.LastViewedAt,
			ViewCount:    s.ViewCount,
			Status:       shareStatus(s, now),
		})
	}
	return infos, nil
}

func (u *shareUseCase) RevokeShare(ctx context.Context, requesterID uuid.UUID, requesterRole string, requesterFamilyID uuid.UUID, shareID uuid.UUID) error {
	share, err := u.shareRepo.FindByID(ctx, shareID)
	if err != nil {
		return ErrShareNotFound
	}
	if _, err := u.resolveTarget(ctx, requesterID, requesterRole, requesterFamilyID, share.UserID); err != nil {
		return err
	}
	// Idempotent: revoke share yang sudah dicabut tetap sukses.
	if share.RevokedAt != nil {
		return nil
	}
	if err := u.shareRepo.Revoke(ctx, shareID, time.Now()); err != nil {
		return fmt.Errorf("revoke share: %w", err)
	}
	return nil
}

func (u *shareUseCase) ViewShared(ctx context.Context, token string) (*PublicDashboard, error) {
	tokenHash := hashShareToken(token)

	// Validitas share SELALU dicek ke DB (query ringan via unique index) supaya
	// revoke berlaku seketika; yang di-cache hanya payload dashboard-nya.
	share, err := u.shareRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrShareNotFound
	}
	now := time.Now()
	if !share.IsValid(now) {
		return nil, ErrShareNotFound
	}

	user, err := u.userRepo.FindByID(ctx, share.UserID)
	if err != nil || !user.IsActive {
		return nil, ErrShareNotFound
	}

	_ = u.shareRepo.TouchView(ctx, share.ID, now)

	cacheKey := "share:view:" + tokenHash
	if u.cache != nil {
		var cached PublicDashboard
		if err := u.cache.Get(ctx, cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	dash, err := u.buildPublicDashboard(ctx, user, share.Label, now)
	if err != nil {
		return nil, err
	}

	if u.cache != nil {
		_ = u.cache.Set(ctx, cacheKey, dash, shareViewCacheTTL)
	}
	return dash, nil
}

func (u *shareUseCase) buildPublicDashboard(ctx context.Context, user *entity.User, label string, now time.Time) (*PublicDashboard, error) {
	hafalan, err := u.hafalanRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("fetch hafalan: %w", err)
	}

	stats := calcHafalanStats(hafalan)

	dash := &PublicDashboard{
		MemberName: user.Name,
		Label:      label,
		Stats: PublicStats{
			TotalHafalAyat:  stats.TotalHafalAyat,
			TotalHafalSurah: stats.TotalHafalSurah,
			PercentHafal:    stats.PercentHafal,
		},
		SurahProgress: []*PublicSurahProgress{},
		LastActivity:  []*PublicActivity{},
		GeneratedAt:   now,
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if schedules, err := u.murajaahRepo.FindTodayByUserID(ctx, user.ID, today); err == nil {
		dash.MurajaahToday.Total = len(schedules)
		for _, s := range schedules {
			if s.CompletedAt != nil {
				dash.MurajaahToday.Done++
			}
		}
	}
	if history, _, err := u.murajaahRepo.FindHistoryByUserID(ctx, user.ID, 200, 0); err == nil {
		dash.Stats.Streak = calcStreak(history, now)
	}

	surahMeta := u.fetchSurahMeta(ctx)
	dash.SurahProgress = buildSurahProgress(hafalan, surahMeta)
	dash.LastActivity = buildLastActivity(hafalan, surahMeta)

	return dash, nil
}

// fetchSurahMeta mengambil metadata surah; kegagalan tidak mematikan halaman
// share — nama surah hanya jadi kosong (degrade gracefully).
func (u *shareUseCase) fetchSurahMeta(ctx context.Context) map[int]quranapi.SurahMeta {
	meta := make(map[int]quranapi.SurahMeta)
	if u.surahNames == nil {
		return meta
	}
	list, err := u.surahNames.GetSurahList(ctx)
	if err != nil {
		return meta
	}
	for _, s := range list {
		meta[s.Number] = s
	}
	return meta
}

func buildSurahProgress(hafalan []*entity.HafalanProgress, meta map[int]quranapi.SurahMeta) []*PublicSurahProgress {
	bySurah := make(map[int]*PublicSurahProgress)
	order := []int{}

	for _, h := range hafalan {
		if h.Status == entity.StatusBelum {
			continue
		}
		p, ok := bySurah[h.SurahNumber]
		if !ok {
			p = &PublicSurahProgress{SurahNumber: h.SurahNumber, Status: string(entity.StatusHafal)}
			if m, found := meta[h.SurahNumber]; found {
				p.SurahName = m.EnglishName
				p.AyatTotal = m.NumberOfAyahs
			}
			bySurah[h.SurahNumber] = p
			order = append(order, h.SurahNumber)
		}
		if h.Status == entity.StatusHafal {
			p.AyatHafal += h.AyatEnd - h.AyatStart + 1
		} else {
			// ada rentang yang masih 'sedang' → surah belum sepenuhnya hafal
			p.Status = string(entity.StatusSedang)
		}
	}

	result := make([]*PublicSurahProgress, 0, len(order))
	for _, num := range order {
		p := bySurah[num]
		if p.AyatTotal > 0 && p.AyatHafal < p.AyatTotal && p.Status == string(entity.StatusHafal) {
			p.Status = string(entity.StatusSedang)
		}
		result = append(result, p)
	}
	return result
}

func buildLastActivity(hafalan []*entity.HafalanProgress, meta map[int]quranapi.SurahMeta) []*PublicActivity {
	sorted := make([]*entity.HafalanProgress, len(hafalan))
	copy(sorted, hafalan)
	// urutkan noted_at menurun; list dari repo terurut per surah, bukan waktu
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].NotedAt.After(sorted[j].NotedAt)
	})

	limit := min(lastActivityLimit, len(sorted))
	result := make([]*PublicActivity, 0, limit)
	for _, h := range sorted[:limit] {
		name := fmt.Sprintf("Surah %d", h.SurahNumber)
		if m, ok := meta[h.SurahNumber]; ok {
			name = m.EnglishName
		}
		result = append(result, &PublicActivity{
			Date:   h.NotedAt.Format("2006-01-02"),
			Type:   "sabak",
			Detail: fmt.Sprintf("%s ayat %d–%d", name, h.AyatStart, h.AyatEnd),
			Status: string(h.Status),
		})
	}
	return result
}

func shareStatus(s *entity.DashboardShare, now time.Time) string {
	switch {
	case s.RevokedAt != nil:
		return "dicabut"
	case s.ExpiresAt != nil && now.After(*s.ExpiresAt):
		return "kedaluwarsa"
	default:
		return "aktif"
	}
}

func generateShareToken() (token, tokenHash string, err error) {
	raw := make([]byte, shareTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, hashShareToken(token), nil
}

func hashShareToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
