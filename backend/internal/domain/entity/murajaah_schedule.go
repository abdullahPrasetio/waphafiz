package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MurajaahType string

const (
	TypeSabqi  MurajaahType = "sabqi"
	TypeManzil MurajaahType = "manzil"
)

type MurajaahSchedule struct {
	ID            uuid.UUID    `gorm:"type:uuid;primaryKey;not null" json:"id"`
	UserID        uuid.UUID    `gorm:"type:uuid;not null;index"      json:"user_id"`
	Type          MurajaahType `gorm:"size:20;not null"              json:"type"`
	SurahNumber   int          `gorm:"not null"                      json:"surah_number"`
	AyatStart     int          `gorm:"not null"                      json:"ayat_start"`
	AyatEnd       int          `gorm:"not null"                      json:"ayat_end"`
	ScheduledDate time.Time    `gorm:"type:date;not null;index"      json:"scheduled_date"`
	CompletedAt   *time.Time   `                                     json:"completed_at"`
	CreatedAt     time.Time    `                                     json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (m *MurajaahSchedule) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (MurajaahSchedule) TableName() string { return "murajaah_schedule" }

type MurajaahLog struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;not null" json:"id"`
	ScheduleID  uuid.UUID  `gorm:"type:uuid;not null;index"      json:"schedule_id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index"      json:"user_id"`
	CompletedAt time.Time  `gorm:"not null"                      json:"completed_at"`
	Notes       string     `gorm:"size:500"                      json:"notes"`
}

func (m *MurajaahLog) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (MurajaahLog) TableName() string { return "murajaah_log" }
