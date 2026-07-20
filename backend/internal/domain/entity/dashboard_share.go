package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DashboardShare adalah link pantau read-only untuk dashboard satu user.
// Token asli tidak pernah disimpan — hanya hash SHA-256-nya (TokenHash).
// Tidak pakai soft delete: revoke dimodelkan eksplisit lewat RevokedAt,
// dan record dipertahankan sebagai audit trail.
type DashboardShare struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;not null" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index"      json:"user_id"`
	CreatedBy    uuid.UUID  `gorm:"type:uuid;not null"            json:"created_by"`
	TokenHash    string     `gorm:"size:64;uniqueIndex;not null"  json:"-"`
	Label        string     `gorm:"size:100;not null"             json:"label"`
	ExpiresAt    *time.Time `                                     json:"expires_at"`
	RevokedAt    *time.Time `                                     json:"revoked_at"`
	LastViewedAt *time.Time `                                     json:"last_viewed_at"`
	ViewCount    int        `gorm:"not null;default:0"            json:"view_count"`
	CreatedAt    time.Time  `                                     json:"created_at"`
	UpdatedAt    time.Time  `                                     json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (s *DashboardShare) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (DashboardShare) TableName() string { return "dashboard_shares" }

// IsValid melaporkan apakah share masih boleh dilihat pada waktu now.
func (s *DashboardShare) IsValid(now time.Time) bool {
	if s.RevokedAt != nil {
		return false
	}
	if s.ExpiresAt != nil && now.After(*s.ExpiresAt) {
		return false
	}
	return true
}
