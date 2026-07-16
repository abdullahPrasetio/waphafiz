package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FamilyGroup struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;not null" json:"id"`
	Name       string         `gorm:"size:200;not null"             json:"name"`
	InviteCode string         `gorm:"size:50;uniqueIndex;not null"  json:"invite_code"`
	CreatedBy  uuid.UUID      `gorm:"type:uuid;not null"            json:"created_by"`
	CreatedAt  time.Time      `                                     json:"created_at"`
	UpdatedAt  time.Time      `                                     json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                         json:"-"`
}

func (f *FamilyGroup) BeforeCreate(_ *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

func (FamilyGroup) TableName() string { return "family_groups" }
