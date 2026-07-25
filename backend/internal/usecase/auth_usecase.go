package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	domainrepo "github.com/abdullahPrasetio/waphafiz/internal/domain/repository"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

var (
	ErrInvalidInviteCode  = errors.New("invite code tidak valid")
	ErrInvalidCredentials = errors.New("email atau password salah")
	ErrUserInactive       = errors.New("akun tidak aktif")
	ErrSamePassword       = errors.New("password baru tidak boleh sama dengan password lama")
)

type RegisterRequest struct {
	InviteCode string `json:"invite_code" validate:"required"`
	Name       string `json:"name"        validate:"required,min=2,max=100"`
	Email      string `json:"email"       validate:"required,email"`
	Password   string `json:"password"    validate:"required,min=8,max=72"`
}

type RegisterFamilyRequest struct {
	FamilyName string `json:"family_name" validate:"required,min=2,max=200"`
	Name       string `json:"name"        validate:"required,min=2,max=100"`
	Email      string `json:"email"       validate:"required,email"`
	Password   string `json:"password"    validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=72"`
	NewPassword     string `json:"new_password"     validate:"required,min=8,max=72"`
}

type AuthResponse struct {
	Token  string    `json:"token"`
	User   UserInfo  `json:"user"`
}

type UserInfo struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Role            string    `json:"role"`
	FamilyGroupID   uuid.UUID `json:"family_group_id"`
	FamilyGroupName string    `json:"family_group_name"`
}

type AuthUseCase interface {
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	RegisterFamily(ctx context.Context, req *RegisterFamilyRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error
}

type authUseCase struct {
	userRepo   domainrepo.UserRepository
	familyRepo domainrepo.FamilyGroupRepository
	jwtCfg     *auth.Config
}

func NewAuthUseCase(
	userRepo domainrepo.UserRepository,
	familyRepo domainrepo.FamilyGroupRepository,
	jwtCfg *auth.Config,
) AuthUseCase {
	return &authUseCase{userRepo: userRepo, familyRepo: familyRepo, jwtCfg: jwtCfg}
}

func (u *authUseCase) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	fg, err := u.familyRepo.FindByInviteCode(ctx, strings.TrimSpace(req.InviteCode))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidInviteCode
		}
		return nil, fmt.Errorf("find family group: %w", err)
	}

	exists, err := u.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if exists {
		return nil, ErrEmailConflict
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &entity.User{
		Name:          req.Name,
		Email:         strings.ToLower(req.Email),
		Password:      string(hash),
		FamilyGroupID: &fg.ID,
		Role:          entity.RoleMember,
		IsActive:      true,
	}
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return u.buildAuthResponse(user, fg)
}

func (u *authUseCase) RegisterFamily(ctx context.Context, req *RegisterFamilyRequest) (*AuthResponse, error) {
	exists, err := u.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if exists {
		return nil, ErrEmailConflict
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// Create admin user first (ID needed as created_by for family group)
	user := &entity.User{
		Name:     req.Name,
		Email:    strings.ToLower(req.Email),
		Password: string(hash),
		Role:     entity.RoleAdmin,
		IsActive: true,
	}
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	code, err := generateInviteCode()
	if err != nil {
		return nil, fmt.Errorf("generate invite code: %w", err)
	}

	fg := &entity.FamilyGroup{
		Name:       req.FamilyName,
		InviteCode: code,
		CreatedBy:  user.ID,
	}
	if err := u.familyRepo.Create(ctx, fg); err != nil {
		return nil, fmt.Errorf("create family group: %w", err)
	}

	// Link user to family group
	user.FamilyGroupID = &fg.ID
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("link user to family: %w", err)
	}

	return u.buildAuthResponse(user, fg)
}

func (u *authUseCase) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now()
	user.LastSeenAt = &now
	_ = u.userRepo.Update(ctx, user)

	var fg *entity.FamilyGroup
	if user.FamilyGroupID != nil {
		fg, _ = u.familyRepo.FindByID(ctx, *user.FamilyGroupID)
	}

	return u.buildAuthResponse(user, fg)
}

func (u *authUseCase) buildAuthResponse(user *entity.User, fg *entity.FamilyGroup) (*AuthResponse, error) {
	token, err := auth.Sign(user.ID.String(), []string{string(user.Role)}, u.jwtCfg)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	info := UserInfo{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  string(user.Role),
	}
	if fg != nil {
		info.FamilyGroupID = fg.ID
		info.FamilyGroupName = fg.Name
	}

	return &AuthResponse{Token: token, User: info}, nil
}

func (u *authUseCase) ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error {
	// Find user
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("find user: %w", err)
	}

	// Verify current password matches
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Reject if new password is identical to the current one
	if req.CurrentPassword == req.NewPassword {
		return ErrSamePassword
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	// Update user password
	user.Password = string(hash)
	if err := u.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

func generateInviteCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "FAM-" + strings.ToUpper(hex.EncodeToString(b)), nil
}
