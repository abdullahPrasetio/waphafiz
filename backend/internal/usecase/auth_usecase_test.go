package usecase_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

// ── fake user repository ──────────────────────────────────────────────────────

type fakeUserRepo struct {
	byID    map[uuid.UUID]*entity.User
	byEmail map[string]*entity.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byID: make(map[uuid.UUID]*entity.User), byEmail: make(map[string]*entity.User)}
}

func (r *fakeUserRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.User, error) {
	u, ok := r.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) FindByEmail(_ context.Context, email string) (*entity.User, error) {
	u, ok := r.byEmail[strings.ToLower(email)]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) FindAll(_ context.Context) ([]*entity.User, error) {
	var out []*entity.User
	for _, u := range r.byID {
		out = append(out, u)
	}
	return out, nil
}

func (r *fakeUserRepo) Create(_ context.Context, user *entity.User) error {
	user.ID = uuid.New()
	r.byID[user.ID] = user
	r.byEmail[strings.ToLower(user.Email)] = user
	return nil
}

func (r *fakeUserRepo) Update(_ context.Context, user *entity.User) error {
	r.byID[user.ID] = user
	r.byEmail[strings.ToLower(user.Email)] = user
	return nil
}

func (r *fakeUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.byID, id)
	return nil
}

func (r *fakeUserRepo) ExistsByEmail(_ context.Context, email string) (bool, error) {
	_, ok := r.byEmail[strings.ToLower(email)]
	return ok, nil
}

// ── fake family group repository ────────────────────────────────────────────

type fakeFamilyGroupRepo struct {
	byID         map[uuid.UUID]*entity.FamilyGroup
	byInviteCode map[string]*entity.FamilyGroup
	members      map[uuid.UUID][]*entity.User
}

func newFakeFamilyGroupRepo() *fakeFamilyGroupRepo {
	return &fakeFamilyGroupRepo{
		byID:         make(map[uuid.UUID]*entity.FamilyGroup),
		byInviteCode: make(map[string]*entity.FamilyGroup),
		members:      make(map[uuid.UUID][]*entity.User),
	}
}

func (r *fakeFamilyGroupRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.FamilyGroup, error) {
	fg, ok := r.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return fg, nil
}

func (r *fakeFamilyGroupRepo) FindByInviteCode(_ context.Context, code string) (*entity.FamilyGroup, error) {
	fg, ok := r.byInviteCode[code]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return fg, nil
}

func (r *fakeFamilyGroupRepo) Create(_ context.Context, fg *entity.FamilyGroup) error {
	fg.ID = uuid.New()
	r.byID[fg.ID] = fg
	r.byInviteCode[fg.InviteCode] = fg
	return nil
}

func (r *fakeFamilyGroupRepo) UpdateInviteCode(_ context.Context, id uuid.UUID, code string) error {
	fg, ok := r.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(r.byInviteCode, fg.InviteCode)
	fg.InviteCode = code
	r.byInviteCode[code] = fg
	return nil
}

func (r *fakeFamilyGroupRepo) ListMembers(_ context.Context, familyGroupID uuid.UUID) ([]*entity.User, error) {
	return r.members[familyGroupID], nil
}

func (r *fakeFamilyGroupRepo) UpdateMemberStatus(_ context.Context, userID uuid.UUID, isActive bool) error {
	return nil
}

func (r *fakeFamilyGroupRepo) UpdateMemberLastSeen(_ context.Context, userID uuid.UUID) error {
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

var testJWTCfg = &auth.Config{
	Secret:   "test-secret-key-at-least-32-bytes-long!",
	Issuer:   "waphafiz-test",
	Audience: "api",
	Expiry:   time.Hour,
}

func seedFamily(t *testing.T, repo *fakeFamilyGroupRepo, code string) *entity.FamilyGroup {
	t.Helper()
	fg := &entity.FamilyGroup{Name: "Keluarga Test", InviteCode: code}
	require.NoError(t, repo.Create(context.Background(), fg))
	return fg
}

// ── Register ─────────────────────────────────────────────────────────────────

func TestAuthRegister_InvalidInviteCode(t *testing.T) {
	uc := usecase.NewAuthUseCase(newFakeUserRepo(), newFakeFamilyGroupRepo(), testJWTCfg)

	_, err := uc.Register(context.Background(), &usecase.RegisterRequest{
		InviteCode: "FAM-NOTEXIST", Name: "Budi", Email: "budi@test.com", Password: "password123",
	})
	assert.ErrorIs(t, err, usecase.ErrInvalidInviteCode)
}

func TestAuthRegister_EmailConflict(t *testing.T) {
	userRepo := newFakeUserRepo()
	familyRepo := newFakeFamilyGroupRepo()
	fg := seedFamily(t, familyRepo, "FAM-AAAA")
	_ = userRepo.Create(context.Background(), &entity.User{Email: "exists@test.com", FamilyGroupID: &fg.ID})
	uc := usecase.NewAuthUseCase(userRepo, familyRepo, testJWTCfg)

	_, err := uc.Register(context.Background(), &usecase.RegisterRequest{
		InviteCode: "FAM-AAAA", Name: "Budi", Email: "exists@test.com", Password: "password123",
	})
	assert.ErrorIs(t, err, usecase.ErrEmailConflict)
}

func TestAuthRegister_OK(t *testing.T) {
	userRepo := newFakeUserRepo()
	familyRepo := newFakeFamilyGroupRepo()
	seedFamily(t, familyRepo, "FAM-BBBB")
	uc := usecase.NewAuthUseCase(userRepo, familyRepo, testJWTCfg)

	resp, err := uc.Register(context.Background(), &usecase.RegisterRequest{
		InviteCode: "FAM-BBBB", Name: "Budi", Email: "Budi@Test.com", Password: "password123",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "member", resp.User.Role)
	assert.Equal(t, "Keluarga Test", resp.User.FamilyGroupName)

	stored, err := userRepo.FindByEmail(context.Background(), "budi@test.com")
	require.NoError(t, err)
	assert.NotEqual(t, "password123", stored.Password, "password must be hashed, not stored in plaintext")
}

// ── RegisterFamily ───────────────────────────────────────────────────────────

func TestAuthRegisterFamily_CreatesAdminAndInviteCode(t *testing.T) {
	userRepo := newFakeUserRepo()
	familyRepo := newFakeFamilyGroupRepo()
	uc := usecase.NewAuthUseCase(userRepo, familyRepo, testJWTCfg)

	resp, err := uc.RegisterFamily(context.Background(), &usecase.RegisterFamilyRequest{
		FamilyName: "Keluarga Baru", Name: "Abdullah", Email: "abdullah@test.com", Password: "password123",
	})
	require.NoError(t, err)
	assert.Equal(t, "admin", resp.User.Role)
	assert.NotEqual(t, uuid.Nil, resp.User.FamilyGroupID)

	fg, err := familyRepo.FindByID(context.Background(), resp.User.FamilyGroupID)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(fg.InviteCode, "FAM-"))
	assert.Equal(t, fg.CreatedBy, resp.User.ID)
}

func TestAuthRegisterFamily_EmailConflict(t *testing.T) {
	userRepo := newFakeUserRepo()
	familyRepo := newFakeFamilyGroupRepo()
	_ = userRepo.Create(context.Background(), &entity.User{Email: "dup@test.com"})
	uc := usecase.NewAuthUseCase(userRepo, familyRepo, testJWTCfg)

	_, err := uc.RegisterFamily(context.Background(), &usecase.RegisterFamilyRequest{
		FamilyName: "Keluarga", Name: "Dup", Email: "dup@test.com", Password: "password123",
	})
	assert.ErrorIs(t, err, usecase.ErrEmailConflict)
}

// ── Login ────────────────────────────────────────────────────────────────────

func TestAuthLogin_OK(t *testing.T) {
	userRepo := newFakeUserRepo()
	familyRepo := newFakeFamilyGroupRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	_ = userRepo.Create(context.Background(), &entity.User{
		Email: "user@test.com", Password: string(hash), Role: entity.RoleMember, IsActive: true,
	})
	uc := usecase.NewAuthUseCase(userRepo, familyRepo, testJWTCfg)

	resp, err := uc.Login(context.Background(), &usecase.LoginRequest{Email: "user@test.com", Password: "password123"})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
}

func TestAuthLogin_WrongPassword(t *testing.T) {
	userRepo := newFakeUserRepo()
	familyRepo := newFakeFamilyGroupRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	_ = userRepo.Create(context.Background(), &entity.User{
		Email: "user@test.com", Password: string(hash), IsActive: true,
	})
	uc := usecase.NewAuthUseCase(userRepo, familyRepo, testJWTCfg)

	_, err := uc.Login(context.Background(), &usecase.LoginRequest{Email: "user@test.com", Password: "wrong-password"})
	assert.ErrorIs(t, err, usecase.ErrInvalidCredentials)
}

func TestAuthLogin_UnknownEmail(t *testing.T) {
	uc := usecase.NewAuthUseCase(newFakeUserRepo(), newFakeFamilyGroupRepo(), testJWTCfg)

	_, err := uc.Login(context.Background(), &usecase.LoginRequest{Email: "nobody@test.com", Password: "whatever"})
	assert.ErrorIs(t, err, usecase.ErrInvalidCredentials)
}

func TestAuthLogin_InactiveUser(t *testing.T) {
	userRepo := newFakeUserRepo()
	familyRepo := newFakeFamilyGroupRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	_ = userRepo.Create(context.Background(), &entity.User{
		Email: "inactive@test.com", Password: string(hash), IsActive: false,
	})
	uc := usecase.NewAuthUseCase(userRepo, familyRepo, testJWTCfg)

	_, err := uc.Login(context.Background(), &usecase.LoginRequest{Email: "inactive@test.com", Password: "password123"})
	assert.ErrorIs(t, err, usecase.ErrUserInactive)
}
