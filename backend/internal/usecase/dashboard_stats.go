package usecase

import (
	"time"

	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
)

// hafalanStats adalah agregat hafalan yang dipakai bersama oleh
// dashboard pribadi dan halaman share publik — satu sumber perhitungan
// supaya angka di keduanya tidak pernah selisih.
type hafalanStats struct {
	TotalHafalAyat  int
	TotalHafalSurah int
	PercentHafal    float64
}

func calcHafalanStats(list []*entity.HafalanProgress) hafalanStats {
	stats := hafalanStats{}
	surahMap := make(map[int]bool)
	totalAyat := 0

	for _, h := range list {
		ayat := h.AyatEnd - h.AyatStart + 1
		totalAyat += ayat
		if h.Status == entity.StatusHafal {
			stats.TotalHafalAyat += ayat
			surahMap[h.SurahNumber] = true
		}
	}
	stats.TotalHafalSurah = len(surahMap)
	if totalAyat > 0 {
		stats.PercentHafal = float64(stats.TotalHafalAyat) / float64(totalAyat) * 100
	}
	return stats
}

// calcStreak menghitung hari berturut-turut (mundur dari now) di mana
// seluruh muraja'ah pada hari itu selesai.
func calcStreak(schedules []*entity.MurajaahSchedule, now time.Time) int {
	if len(schedules) == 0 {
		return 0
	}

	dayTotal := make(map[string]int)
	dayDone := make(map[string]int)
	for _, s := range schedules {
		day := s.ScheduledDate.Format("2006-01-02")
		dayTotal[day]++
		if s.CompletedAt != nil {
			dayDone[day]++
		}
	}
	dayCompleted := make(map[string]bool)
	for day, total := range dayTotal {
		if dayDone[day] == total {
			dayCompleted[day] = true
		}
	}

	streak := 0
	current := now
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
