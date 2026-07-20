package usecase_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
	"github.com/abdullahPrasetio/waphafiz/pkg/quranapi"
)

// ── fake share repository ────────────────────────────────────────────────────

type fakeShareRepo struct {
	byID map[uuid.UUID]*entity.DashboardShare
}

func newFakeShareRepo() *fakeShareRepo {
	return &fakeShareRepo{byID: make(map[uuid.UUID]*entity.DashboardShare)}
}

func (r *fakeShareRepo) Create(_ context.Context, s *entity.DashboardShare) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	s.CreatedAt = time.Now()
	r.byID[s.ID] = s
	return nil
}

func (r *fakeShareRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.DashboardShare, error) {
	s, ok := r.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return s, nil
}

func (r *fakeShareRepo) FindByTokenHash(_ context.Context, tokenHash string) (*entity.DashboardShare, error) {
	for _, s := range r.byID {
		if s.TokenHash == tokenHash {
			return s, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeShareRepo) ListByUserID(_ context.Context, userID uuid.UUID) ([]*entity.DashboardShare, error) {
	var out []*entity.DashboardShare
	for _, s := range r.byID {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *fakeShareRepo) CountActiveByUserID(_ context.Context, userID uuid.UUID, now time.Time) (int64, error) {
	var count int64
	for _, s := range r.byID {
		if s.UserID == userID && s.IsValid(now) {
			count++
		}
	}
	return count, nil
}

func (r *fakeShareRepo) Revoke(_ context.Context, id uuid.UUID, revokedAt time.Time) error {
	if s, ok := r.byID[id]; ok && s.RevokedAt == nil {
		s.RevokedAt = &revokedAt
	}
	return nil
}

func (r *fakeShareRepo) TouchView(_ context.Context, id uuid.UUID, viewedAt time.Time) error {
	if s, ok := r.byID[id]; ok {
		s.LastViewedAt = &viewedAt
		s.ViewCount++
	}
	return nil
}

func (r *fakeShareRepo) PurgeInactiveBefore(_ context.Context, userID uuid.UUID, before time.Time) error {
	for id, s := range r.byID {
		if s.UserID != userID {
			continue
		}
		inactiveSince := s.RevokedAt
		if s.ExpiresAt != nil && (inactiveSince == nil || s.ExpiresAt.Before(*inactiveSince)) {
			inactiveSince = s.ExpiresAt
		}
		if inactiveSince != nil && inactiveSince.Before(before) {
			delete(r.byID, id)
		}
	}
	return nil
}

// ── fake surah name provider ─────────────────────────────────────────────────

type fakeSurahNames struct{}

func (fakeSurahNames) GetSurahList(_ context.Context) ([]quranapi.SurahMeta, error) {
	return []quranapi.SurahMeta{
		{Number: 67, EnglishName: "Al-Mulk", NumberOfAyahs: 30},
		{Number: 1, EnglishName: "Al-Fatihah", NumberOfAyahs: 7},
	}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

type shareEnv struct {
	shareRepo *fakeShareRepo
	userRepo  *fakeUserRepo
	hRepo     *fakeHafalanRepo
	mRepo     *fakeMurajaahRepo
	uc        usecase.ShareUseCase
}

func newShareEnv() *shareEnv {
	env := &shareEnv{
		shareRepo: newFakeShareRepo(),
		userRepo:  newFakeUserRepo(),
		hRepo:     newFakeHafalanRepo(),
		mRepo:     newFakeMurajaahRepo(),
	}
	env.uc = usecase.NewShareUseCase(env.shareRepo, env.userRepo, env.hRepo, env.mRepo, fakeSurahNames{}, nil)
	return env
}

func (e *shareEnv) addUser(name string, familyID *uuid.UUID, role entity.UserRole, active bool) *entity.User {
	u := &entity.User{ID: uuid.New(), Name: name, Email: name + "@test.local", FamilyGroupID: familyID, Role: role, IsActive: active}
	e.userRepo.byID[u.ID] = u
	return u
}

func sha256hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// ── create ───────────────────────────────────────────────────────────────────

func TestShareCreate_SelfStoresHashNotToken(t *testing.T) {
	env := newShareEnv()
	famID := uuid.New()
	user := env.addUser("ahmad", &famID, entity.RoleMember, true)

	res, err := env.uc.CreateShare(context.Background(), user.ID, "member", famID,
		usecase.CreateShareInput{Label: "Ustadz Ahmad", ExpiresInDays: 90})
	require.NoError(t, err)
	require.NotEmpty(t, res.Token)
	require.NotNil(t, res.ExpiresAt)

	stored := env.shareRepo.byID[res.ID]
	require.NotNil(t, stored)
	assert.NotEqual(t, res.Token, stored.TokenHash, "token asli tidak boleh tersimpan")
	assert.Equal(t, sha256hex(res.Token), stored.TokenHash)
	assert.Equal(t, user.ID, stored.UserID)
	assert.Equal(t, user.ID, stored.CreatedBy)
}

func TestShareCreate_NoExpiryWhenZeroDays(t *testing.T) {
	env := newShareEnv()
	famID := uuid.New()
	user := env.addUser("ahmad", &famID, entity.RoleMember, true)

	res, err := env.uc.CreateShare(context.Background(), user.ID, "member", famID,
		usecase.CreateShareInput{Label: "Nenek", ExpiresInDays: 0})
	require.NoError(t, err)
	assert.Nil(t, res.ExpiresAt)
}

func TestShareCreate_AdminForFamilyMember(t *testing.T) {
	env := newShareEnv()
	famID := uuid.New()
	admin := env.addUser("ayah", &famID, entity.RoleAdmin, true)
	child := env.addUser("anak", &famID, entity.RoleMember, true)

	res, err := env.uc.CreateShare(context.Background(), admin.ID, "admin", famID,
		usecase.CreateShareInput{UserID: &child.ID, Label: "Guru ngaji"})
	require.NoError(t, err)
	assert.Equal(t, child.ID, env.shareRepo.byID[res.ID].UserID)
	assert.Equal(t, admin.ID, env.shareRepo.byID[res.ID].CreatedBy)
}

func TestShareCreate_MemberForOtherForbidden(t *testing.T) {
	env := newShareEnv()
	famID := uuid.New()
	m1 := env.addUser("m1", &famID, entity.RoleMember, true)
	m2 := env.addUser("m2", &famID, entity.RoleMember, true)

	_, err := env.uc.CreateShare(context.Background(), m1.ID, "member", famID,
		usecase.CreateShareInput{UserID: &m2.ID, Label: "x"})
	assert.ErrorIs(t, err, usecase.ErrShareForbidden)
}

func TestShareCreate_AdminOtherFamilyForbidden(t *testing.T) {
	env := newShareEnv()
	famA, famB := uuid.New(), uuid.New()
	admin := env.addUser("adminA", &famA, entity.RoleAdmin, true)
	other := env.addUser("memberB", &famB, entity.RoleMember, true)

	_, err := env.uc.CreateShare(context.Background(), admin.ID, "admin", famA,
		usecase.CreateShareInput{UserID: &other.ID, Label: "x"})
	assert.ErrorIs(t, err, usecase.ErrShareForbidden)
}

func TestShareCreate_LimitReached(t *testing.T) {
	env := newShareEnv()
	famID := uuid.New()
	user := env.addUser("ahmad", &famID, entity.RoleMember, true)

	for range 10 {
		_, err := env.uc.CreateShare(context.Background(), user.ID, "member", famID,
			usecase.CreateShareInput{Label: "x"})
		require.NoError(t, err)
	}
	_, err := env.uc.CreateShare(context.Background(), user.ID, "member", famID,
		usecase.CreateShareInput{Label: "kesebelas"})
	assert.ErrorIs(t, err, usecase.ErrShareLimitReached)
}

// ── view ─────────────────────────────────────────────────────────────────────

func TestShareView_ValidTokenReturnsStats(t *testing.T) {
	env := newShareEnv()
	famID := uuid.New()
	user := env.addUser("Ahmad", &famID, entity.RoleMember, true)
	env.hRepo.items[uuid.New()] = &entity.HafalanProgress{
		UserID: user.ID, SurahNumber: 67, AyatStart: 1, AyatEnd: 30,
		Status: entity.StatusHafal, NotedAt: time.Now(),
	}

	res, err := env.uc.CreateShare(context.Background(), user.ID, "member", famID,
		usecase.CreateShareInput{Label: "Ustadz"})
	require.NoError(t, err)

	dash, err := env.uc.ViewShared(context.Background(), res.Token)
	require.NoError(t, err)
	assert.Equal(t, "Ahmad", dash.MemberName)
	assert.Equal(t, "Ustadz", dash.Label)
	assert.Equal(t, 30, dash.Stats.TotalHafalAyat)
	assert.Equal(t, 1, dash.Stats.TotalHafalSurah)
	require.Len(t, dash.SurahProgress, 1)
	assert.Equal(t, "Al-Mulk", dash.SurahProgress[0].SurahName)
	assert.Equal(t, "hafal", dash.SurahProgress[0].Status)
	require.Len(t, dash.LastActivity, 1)
	assert.Contains(t, dash.LastActivity[0].Detail, "Al-Mulk")

	// view tercatat
	stored := env.shareRepo.byID[res.ID]
	assert.Equal(t, 1, stored.ViewCount)
	assert.NotNil(t, stored.LastViewedAt)
}

func TestShareView_PayloadNeverLeaksEmailOrUserID(t *testing.T) {
	env := newShareEnv()
	famID := uuid.New()
	user := env.addUser("Ahmad", &famID, entity.RoleMember, true)

	res, err := env.uc.CreateShare(context.Background(), user.ID, "member", famID,
		usecase.CreateShareInput{Label: "Ustadz"})
	require.NoError(t, err)

	dash, err := env.uc.ViewShared(context.Background(), res.Token)
	require.NoError(t, err)

	raw, err := json.Marshal(dash)
	require.NoError(t, err)
	payload := strings.ToLower(string(raw))
	assert.NotContains(t, payload, user.Email, "payload publik bocor email")
	assert.NotContains(t, payload, strings.ToLower(user.ID.String()), "payload publik bocor user_id")
	assert.NotContains(t, payload, "email")
	assert.NotContains(t, payload, "last_seen")
}

func TestShareView_UniformFailures(t *testing.T) {
	env := newShareEnv()
	famID := uuid.New()
	activeUser := env.addUser("aktif", &famID, entity.RoleMember, true)
	inactiveUser := env.addUser("nonaktif", &famID, entity.RoleMember, false)

	// token salah
	_, err := env.uc.ViewShared(context.Background(), "token-ngawur")
	assert.ErrorIs(t, err, usecase.ErrShareNotFound)

	// revoked
	revoked, err := env.uc.CreateShare(context.Background(), activeUser.ID, "member", famID,
		usecase.CreateShareInput{Label: "revoked"})
	require.NoError(t, err)
	require.NoError(t, env.uc.RevokeShare(context.Background(), activeUser.ID, "member", famID, revoked.ID))
	_, err = env.uc.ViewShared(context.Background(), revoked.Token)
	assert.ErrorIs(t, err, usecase.ErrShareNotFound)

	// expired: mundurkan expires_at langsung di store
	expired, err := env.uc.CreateShare(context.Background(), activeUser.ID, "member", famID,
		usecase.CreateShareInput{Label: "expired", ExpiresInDays: 7})
	require.NoError(t, err)
	past := time.Now().Add(-time.Hour)
	env.shareRepo.byID[expired.ID].ExpiresAt = &past
	_, err = env.uc.ViewShared(context.Background(), expired.Token)
	assert.ErrorIs(t, err, usecase.ErrShareNotFound)

	// user nonaktif
	inactive, err := env.uc.CreateShare(context.Background(), inactiveUser.ID, "member", famID,
		usecase.CreateShareInput{Label: "inactive"})
	require.NoError(t, err)
	_, err = env.uc.ViewShared(context.Background(), inactive.Token)
	assert.ErrorIs(t, err, usecase.ErrShareNotFound)
}

// ── revoke & list ────────────────────────────────────────────────────────────

func TestShareRevoke_Permissions(t *testing.T) {
	env := newShareEnv()
	famID := uuid.New()
	admin := env.addUser("ayah", &famID, entity.RoleAdmin, true)
	owner := env.addUser("anak", &famID, entity.RoleMember, true)
	other := env.addUser("lain", &famID, entity.RoleMember, true)

	res, err := env.uc.CreateShare(context.Background(), owner.ID, "member", famID,
		usecase.CreateShareInput{Label: "x"})
	require.NoError(t, err)

	// member lain → ditolak
	err = env.uc.RevokeShare(context.Background(), other.ID, "member", famID, res.ID)
	assert.ErrorIs(t, err, usecase.ErrShareForbidden)

	// pemilik → sukses; ulangi → tetap sukses (idempotent)
	require.NoError(t, env.uc.RevokeShare(context.Background(), owner.ID, "member", famID, res.ID))
	require.NoError(t, env.uc.RevokeShare(context.Background(), owner.ID, "member", famID, res.ID))

	// admin merevoke share member lain → sukses
	res2, err := env.uc.CreateShare(context.Background(), owner.ID, "member", famID,
		usecase.CreateShareInput{Label: "y"})
	require.NoError(t, err)
	require.NoError(t, env.uc.RevokeShare(context.Background(), admin.ID, "admin", famID, res2.ID))
}

func TestShareList_StatusComputed(t *testing.T) {
	env := newShareEnv()
	famID := uuid.New()
	user := env.addUser("ahmad", &famID, entity.RoleMember, true)

	aktif, err := env.uc.CreateShare(context.Background(), user.ID, "member", famID,
		usecase.CreateShareInput{Label: "aktif"})
	require.NoError(t, err)
	_ = aktif

	dicabut, err := env.uc.CreateShare(context.Background(), user.ID, "member", famID,
		usecase.CreateShareInput{Label: "dicabut"})
	require.NoError(t, err)
	require.NoError(t, env.uc.RevokeShare(context.Background(), user.ID, "member", famID, dicabut.ID))

	kadaluwarsa, err := env.uc.CreateShare(context.Background(), user.ID, "member", famID,
		usecase.CreateShareInput{Label: "kedaluwarsa", ExpiresInDays: 1})
	require.NoError(t, err)
	past := time.Now().Add(-time.Hour)
	env.shareRepo.byID[kadaluwarsa.ID].ExpiresAt = &past

	list, err := env.uc.ListShares(context.Background(), user.ID, "member", famID, nil)
	require.NoError(t, err)
	require.Len(t, list, 3)

	statusByLabel := map[string]string{}
	for _, s := range list {
		statusByLabel[s.Label] = s.Status
	}
	assert.Equal(t, "aktif", statusByLabel["aktif"])
	assert.Equal(t, "dicabut", statusByLabel["dicabut"])
	assert.Equal(t, "kedaluwarsa", statusByLabel["kedaluwarsa"])
}

func TestShareList_PurgesOldInactiveShares(t *testing.T) {
	env := newShareEnv()
	famID := uuid.New()
	user := env.addUser("ahmad", &famID, entity.RoleMember, true)

	// Dicabut 200 hari lalu (> retensi 180 hari) → harus terhapus permanen
	// saat ListShares dipanggil berikutnya (lazy purge).
	oldRevoked, err := env.uc.CreateShare(context.Background(), user.ID, "member", famID,
		usecase.CreateShareInput{Label: "lama"})
	require.NoError(t, err)
	longAgo := time.Now().AddDate(0, 0, -200)
	env.shareRepo.byID[oldRevoked.ID].RevokedAt = &longAgo

	// Dicabut baru-baru ini → harus tetap ada.
	recentRevoked, err := env.uc.CreateShare(context.Background(), user.ID, "member", famID,
		usecase.CreateShareInput{Label: "baru"})
	require.NoError(t, err)
	require.NoError(t, env.uc.RevokeShare(context.Background(), user.ID, "member", famID, recentRevoked.ID))

	list, err := env.uc.ListShares(context.Background(), user.ID, "member", famID, nil)
	require.NoError(t, err)
	require.Len(t, list, 1, "share yang dicabut >180 hari lalu harus sudah terhapus")
	assert.Equal(t, "baru", list[0].Label)
}
