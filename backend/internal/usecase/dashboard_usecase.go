package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	domainrepo "github.com/abdullahPrasetio/waphafiz/internal/domain/repository"
)

type MyDashboard struct {
	TotalHafalAyat  int     `json:"total_hafal_ayat"`
	TotalHafalSurah int     `json:"total_hafal_surah"`
	PercentHafal    float64 `json:"percent_hafal"`
	MurajaahDone    int     `json:"murajaah_done"`
	MurajaahTotal   int     `json:"murajaah_total"`
	Streak          int     `json:"streak"`
}

type MemberProgress struct {
	UserID   uuid.UUID `json:"user_id"`
	UserName string    `json:"user_name"`
	HafalAyat int     `json:"hafal_ayat"`
	IsActive  bool    `json:"is_active"`
	LastSeen  *string `json:"last_seen"`
}

type AdminDashboard struct {
	TotalMembers    int               `json:"total_members"`
	ActiveToday     int               `json:"active_today"`
	MemberProgress  []*MemberProgress `json:"member_progress"`
}

type DashboardUseCase interface {
	GetMyDashboard(ctx context.Context, userID uuid.UUID) (*MyDashboard, error)
	GetAdminDashboard(ctx context.Context, familyGroupID uuid.UUID) (*AdminDashboard, error)
}

type dashboardUseCase struct {
	hafalanRepo  domainrepo.HafalanRepository
	familyRepo   domainrepo.FamilyGroupRepository
	murajaahRepo domainrepo.MurajaahRepository
}

func NewDashboardUseCase(
	hafalanRepo domainrepo.HafalanRepository,
	familyRepo domainrepo.FamilyGroupRepository,
	murajaahRepo domainrepo.MurajaahRepository,
) DashboardUseCase {
	return &dashboardUseCase{hafalanRepo: hafalanRepo, familyRepo: familyRepo, murajaahRepo: murajaahRepo}
}

func (u *dashboardUseCase) GetMyDashboard(ctx context.Context, userID uuid.UUID) (*MyDashboard, error) {
	hafalan, err := u.hafalanRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch hafalan: %w", err)
	}

	dash := &MyDashboard{}
	surahMap := make(map[int]bool)
	totalAyat := 0

	for _, h := range hafalan {
		ayat := h.AyatEnd - h.AyatStart + 1
		totalAyat += ayat
		if h.Status == entity.StatusHafal {
			dash.TotalHafalAyat += ayat
			surahMap[h.SurahNumber] = true
		}
	}
	dash.TotalHafalSurah = len(surahMap)
	if totalAyat > 0 {
		dash.PercentHafal = float64(dash.TotalHafalAyat) / float64(totalAyat) * 100
	}

	today := todayDate()
	schedules, err := u.murajaahRepo.FindTodayByUserID(ctx, userID, today)
	if err == nil {
		dash.MurajaahTotal = len(schedules)
		for _, s := range schedules {
			if s.CompletedAt != nil {
				dash.MurajaahDone++
			}
		}
	}

	// Simple streak: count consecutive days with completed murajaah
	dash.Streak = u.calculateStreak(ctx, userID)

	return dash, nil
}

func (u *dashboardUseCase) calculateStreak(ctx context.Context, userID uuid.UUID) int {
	// Get recent history (last 30 days max)
	schedules, _, err := u.murajaahRepo.FindHistoryByUserID(ctx, userID, 200, 0)
	if err != nil || len(schedules) == 0 {
		return 0
	}

	// Build set of days where all murajaah were completed
	dayCompleted := make(map[string]bool)
	dayTotal := make(map[string]int)
	dayDone := make(map[string]int)

	for _, s := range schedules {
		day := s.ScheduledDate.Format("2006-01-02")
		dayTotal[day]++
		if s.CompletedAt != nil {
			dayDone[day]++
		}
	}
	for day, total := range dayTotal {
		if dayDone[day] == total {
			dayCompleted[day] = true
		}
	}

	streak := 0
	current := time.Now()
	for {
		day := current.Format("2006-01-02")
		if !dayCompleted[day] {
			break
		}
		streak++
		current = current.AddDate(0, 0, -1)
	}
	return streak
}

func (u *dashboardUseCase) GetAdminDashboard(ctx context.Context, familyGroupID uuid.UUID) (*AdminDashboard, error) {
	members, err := u.familyRepo.ListMembers(ctx, familyGroupID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	dash := &AdminDashboard{
		TotalMembers: len(members),
	}

	today := time.Now()
	activeThreshold := today.AddDate(0, 0, -1)

	for _, m := range members {
		mp := &MemberProgress{
			UserID:   m.ID,
			UserName: m.Name,
			IsActive: m.IsActive,
		}
		if m.LastSeenAt != nil {
			s := m.LastSeenAt.Format("2006-01-02T15:04:05Z07:00")
			mp.LastSeen = &s
			if m.LastSeenAt.After(activeThreshold) {
				dash.ActiveToday++
			}
		}

		hafalan, _ := u.hafalanRepo.FindByUserID(ctx, m.ID)
		for _, h := range hafalan {
			if h.Status == entity.StatusHafal {
				mp.HafalAyat += h.AyatEnd - h.AyatStart + 1
			}
		}
		dash.MemberProgress = append(dash.MemberProgress, mp)
	}

	return dash, nil
}
