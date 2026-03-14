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
		ID:          "streak_7",
		Name:        "Konsisten",
		Description: "7 hari berturut-turut",
		Points:      25,
		Check:       func(r *Report) bool { return r.MaxStreak >= 7 },
	},
	{
		ID:          "streak_14",
		Name:        "On Fire",
		Description: "14 hari berturut-turut",
		Points:      50,
		Check:       func(r *Report) bool { return r.MaxStreak >= 14 },
	},
	{
		ID:          "streak_21",
		Name:        "Gigih",
		Description: "21 hari berturut-turut",
		Points:      75,
		Check:       func(r *Report) bool { return r.MaxStreak >= 21 },
	},
	{
		ID:          "streak_30",
		Name:        "Spartan",
		Description: "30 hari berturut-turut",
		Points:      100,
		Check:       func(r *Report) bool { return r.MaxStreak >= 30 },
	},
	{
		ID:          "streak_50",
		Name:        "Titan",
		Description: "50 hari berturut-turut",
		Points:      150,
		Check:       func(r *Report) bool { return r.MaxStreak >= 50 },
	},
	{
		ID:          "streak_100",
		Name:        "Centurion",
		Description: "100 hari berturut-turut",
		Points:      300,
		Check:       func(r *Report) bool { return r.MaxStreak >= 100 },
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
		ID:               "comeback_7",
		Name:             "Comeback Kid",
		Description:      "Kembali dan raih 7 streak setelah absen >7 hari",
		Points:           30,
		MinInactiveDays:  7,
		MinComebackStreak: 7,
	},
	{
		ID:               "comeback_hero",
		Name:             "Comeback Hero",
		Description:      "Kembali dan raih 14 streak setelah absen >14 hari",
		Points:           75,
		MinInactiveDays:  14,
		MinComebackStreak: 14,
	},
	{
		ID:               "phoenix",
		Name:             "Phoenix",
		Description:      "Kembali dan raih 30 streak setelah absen >30 hari",
		Points:           150,
		MinInactiveDays:  30,
		MinComebackStreak: 30,
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
