package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HafalanStatus string

const (
	StatusBelum  HafalanStatus = "belum"
	StatusSedang HafalanStatus = "sedang"
	StatusHafal  HafalanStatus = "hafal"
)

type HafalanProgress struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;not null" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index"      json:"user_id"`
	SurahNumber int            `gorm:"not null"                      json:"surah_number"`
	AyatStart   int            `gorm:"not null"                      json:"ayat_start"`
	AyatEnd     int            `gorm:"not null"                      json:"ayat_end"`
	Status      HafalanStatus  `gorm:"size:20;not null"              json:"status"`
	NotedAt     time.Time      `gorm:"not null"                      json:"noted_at"`
	CreatedAt   time.Time      `                                     json:"created_at"`
	UpdatedAt   time.Time      `                                     json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                         json:"-"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (h *HafalanProgress) BeforeCreate(_ *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}

func (HafalanProgress) TableName() string { return "hafalan_progress" }
