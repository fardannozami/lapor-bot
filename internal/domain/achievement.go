package domain

import "strings"

// Achievement represents a gamification achievement that users can unlock.
type Achievement struct {
	ID          string
	Name        string
	Description string
	Points      int
	Check       func(report *Report) bool
}

// AllAchievements defines all available achievements in order.
var AllAchievements = []Achievement{
	{
		ID:          "first_report",
		Name:        "Pemula",
		Description: "Laporan pertama",
		Points:      10,
		Check:       func(r *Report) bool { return r.ActivityCount >= 1 },
	},
	{
		ID:          "streak_1",
		Name:        "Konsisten",
		Description: "1 minggu berturut-turut",
		Points:      25,
		Check:       func(r *Report) bool { return r.MaxStreak >= 1 },
	},
	{
		ID:          "streak_2",
		Name:        "On Fire",
		Description: "2 minggu berturut-turut",
		Points:      50,
		Check:       func(r *Report) bool { return r.MaxStreak >= 2 },
	},
	{
		ID:          "streak_3",
		Name:        "Gigih",
		Description: "3 minggu berturut-turut",
		Points:      75,
		Check:       func(r *Report) bool { return r.MaxStreak >= 3 },
	},
	{
		ID:          "streak_4",
		Name:        "Spartan",
		Description: "4 minggu berturut-turut",
		Points:      100,
		Check:       func(r *Report) bool { return r.MaxStreak >= 4 },
	},
	{
		ID:          "streak_8",
		Name:        "Titan",
		Description: "8 minggu berturut-turut",
		Points:      150,
		Check:       func(r *Report) bool { return r.MaxStreak >= 8 },
	},
	{
		ID:          "streak_12",
		Name:        "Centurion",
		Description: "12 minggu berturut-turut",
		Points:      300,
		Check:       func(r *Report) bool { return r.MaxStreak >= 12 },
	},
	{
		ID:          "activity_10",
		Name:        "10 Hari",
		Description: "Total 10 hari aktif",
		Points:      20,
		Check:       func(r *Report) bool { return r.ActivityCount >= 10 },
	},
	{
		ID:          "activity_25",
		Name:        "25 Hari",
		Description: "Total 25 hari aktif",
		Points:      50,
		Check:       func(r *Report) bool { return r.ActivityCount >= 25 },
	},
	{
		ID:          "activity_50",
		Name:        "Half Century",
		Description: "Total 50 hari aktif",
		Points:      100,
		Check:       func(r *Report) bool { return r.ActivityCount >= 50 },
	},
	{
		ID:          "activity_100",
		Name:        "Century",
		Description: "Total 100 hari aktif",
		Points:      200,
		Check:       func(r *Report) bool { return r.ActivityCount >= 100 },
	},
}

// ComebackAchievement represents achievements earned by returning after inactivity.
// These are checked separately because they need InactiveDays context.
type ComebackAchievement struct {
	ID               string
	Name             string
	Description      string
	Points           int
	MinInactiveDays  int // must have been inactive this many days
	MinComebackStreak int // must rebuild streak to this many days
}

// AllComebackAchievements defines achievements for users who return after inactivity.
var AllComebackAchievements = []ComebackAchievement{
	{
		ID:               "comeback_4",
		Name:             "Comeback Kid",
		Description:      "Kembali dan raih 4 minggu streak setelah absen lama",
		Points:           30,
		MinInactiveDays:  7,
		MinComebackStreak: 4,
	},
	{
		ID:               "comeback_hero",
		Name:             "Comeback Hero",
		Description:      "Kembali dan raih 8 minggu streak setelah absen lama",
		Points:           75,
		MinInactiveDays:  14,
		MinComebackStreak: 8,
	},
	{
		ID:               "phoenix",
		Name:             "Phoenix",
		Description:      "Kembali dan raih 12 minggu streak setelah absen lama",
		Points:           150,
		MinInactiveDays:  30,
		MinComebackStreak: 12,
	},
}

// CheckComebackAchievements evaluates comeback achievements against the report.
func CheckComebackAchievements(report *Report) []ComebackAchievement {
	var newlyUnlocked []ComebackAchievement
	for _, a := range AllComebackAchievements {
		if !HasAchievement(report.Achievements, a.ID) &&
			report.InactiveDays >= a.MinInactiveDays &&
			report.ComebackStreak >= a.MinComebackStreak {
			newlyUnlocked = append(newlyUnlocked, a)
		}
	}
	return newlyUnlocked
}

// HasAchievement checks if a report's achievements string contains the given achievement ID.
func HasAchievement(achievements string, id string) bool {
	if achievements == "" {
		return false
	}
	for _, a := range strings.Split(achievements, ",") {
		if strings.TrimSpace(a) == id {
			return true
		}
	}
	return false
}

// AddAchievement appends an achievement ID to the achievements string.
func AddAchievement(achievements string, id string) string {
	if achievements == "" {
		return id
	}
	return achievements + "," + id
}

// CheckNewAchievements evaluates all achievements against the report and returns newly unlocked ones.
func CheckNewAchievements(report *Report) []Achievement {
	var newlyUnlocked []Achievement
	for _, a := range AllAchievements {
		if !HasAchievement(report.Achievements, a.ID) && a.Check(report) {
			newlyUnlocked = append(newlyUnlocked, a)
		}
	}
	return newlyUnlocked
}
