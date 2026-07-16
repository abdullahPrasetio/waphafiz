package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	domainrepo "github.com/abdullahPrasetio/waphafiz/internal/domain/repository"
)

type MurajaahUseCase interface {
	GetToday(ctx context.Context, userID uuid.UUID) ([]*entity.MurajaahSchedule, error)
	MarkComplete(ctx context.Context, userID, scheduleID uuid.UUID) error
	GetHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.MurajaahSchedule, int64, error)
}

type murajaahUseCase struct {
	murajaahRepo domainrepo.MurajaahRepository
	hafalanRepo  domainrepo.HafalanRepository
}

func NewMurajaahUseCase(murajaahRepo domainrepo.MurajaahRepository, hafalanRepo domainrepo.HafalanRepository) MurajaahUseCase {
	return &murajaahUseCase{murajaahRepo: murajaahRepo, hafalanRepo: hafalanRepo}
}

func (u *murajaahUseCase) GetToday(ctx context.Context, userID uuid.UUID) ([]*entity.MurajaahSchedule, error) {
	today := todayDate()

	exists, err := u.murajaahRepo.ExistsTodayForUser(ctx, userID, today)
	if err != nil {
		return nil, fmt.Errorf("check existing schedule: %w", err)
	}

	if !exists {
		if err := u.generateSchedule(ctx, userID, today); err != nil {
			return nil, err
		}
	}

	schedules, err := u.murajaahRepo.FindTodayByUserID(ctx, userID, today)
	if err != nil {
		return nil, fmt.Errorf("find today schedules: %w", err)
	}
	return schedules, nil
}

func (u *murajaahUseCase) MarkComplete(ctx context.Context, userID, scheduleID uuid.UUID) error {
	schedule, err := u.murajaahRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return ErrNotFound
	}
	if schedule.UserID != userID {
		return ErrNotOwner
	}
	if schedule.CompletedAt != nil {
		return nil // already done
	}

	now := time.Now()
	if err := u.murajaahRepo.MarkComplete(ctx, scheduleID, now); err != nil {
		return fmt.Errorf("mark complete: %w", err)
	}

	log := &entity.MurajaahLog{
		ScheduleID:  scheduleID,
		UserID:      userID,
		CompletedAt: now,
	}
	_ = u.murajaahRepo.CreateLog(ctx, log)
	return nil
}

func (u *murajaahUseCase) GetHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.MurajaahSchedule, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return u.murajaahRepo.FindHistoryByUserID(ctx, userID, limit, offset)
}

// generateSchedule builds sabqi + manzil entries for today.
// Skipped silently if user has no hafalan.
func (u *murajaahUseCase) generateSchedule(ctx context.Context, userID uuid.UUID, date time.Time) error {
	allHafalan, err := u.hafalanRepo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetch hafalan: %w", err)
	}
	if len(allHafalan) == 0 {
		return nil
	}

	var schedules []*entity.MurajaahSchedule

	// Sabqi: hafalan dalam 7 hari terakhir
	since := date.AddDate(0, 0, -7)
	sabqiItems, err := u.hafalanRepo.FindByUserSince(ctx, userID, since)
	if err != nil {
		return fmt.Errorf("fetch sabqi hafalan: %w", err)
	}
	for _, h := range sabqiItems {
		schedules = append(schedules, &entity.MurajaahSchedule{
			UserID:        userID,
			Type:          entity.TypeSabqi,
			SurahNumber:   h.SurahNumber,
			AyatStart:     h.AyatStart,
			AyatEnd:       h.AyatEnd,
			ScheduledDate: date,
		})
	}

	// Manzil: bagi seluruh hafalan ke 7 slot, rotate berdasarkan hari
	daySlot := int(date.Weekday()) // 0-6
	chunkSize := (len(allHafalan) + 6) / 7
	start := daySlot * chunkSize
	if start >= len(allHafalan) {
		start = 0
	}
	end := start + chunkSize
	if end > len(allHafalan) {
		end = len(allHafalan)
	}
	for _, h := range allHafalan[start:end] {
		// skip duplicate with sabqi
		alreadySabqi := false
		for _, s := range schedules {
			if s.SurahNumber == h.SurahNumber && s.AyatStart == h.AyatStart && s.AyatEnd == h.AyatEnd {
				alreadySabqi = true
				break
			}
		}
		if alreadySabqi {
			continue
		}
		schedules = append(schedules, &entity.MurajaahSchedule{
			UserID:        userID,
			Type:          entity.TypeManzil,
			SurahNumber:   h.SurahNumber,
			AyatStart:     h.AyatStart,
			AyatEnd:       h.AyatEnd,
			ScheduledDate: date,
		})
	}

	if len(schedules) == 0 {
		return nil
	}
	return u.murajaahRepo.CreateSchedules(ctx, schedules)
}

func todayDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
