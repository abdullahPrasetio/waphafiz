package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleMember UserRole = "member"
)

type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;not null"       json:"id"`
	Name          string         `gorm:"size:100;not null"                   json:"name"`
	Email         string         `gorm:"size:255;uniqueIndex;not null"       json:"email"`
	Password      string         `gorm:"size:255;not null"                   json:"-"`
	FamilyGroupID *uuid.UUID     `gorm:"type:uuid;index"                     json:"family_group_id"`
	Role          UserRole       `gorm:"size:20;not null;default:'member'"   json:"role"`
	IsActive      bool           `gorm:"not null;default:true"               json:"is_active"`
	LastSeenAt    *time.Time     `                                           json:"last_seen_at"`
	CreatedAt     time.Time      `                                           json:"created_at"`
	UpdatedAt     time.Time      `                                           json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index"                               json:"-"`

	FamilyGroup *FamilyGroup `gorm:"foreignKey:FamilyGroupID" json:"family_group,omitempty"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (User) TableName() string { return "users" }
