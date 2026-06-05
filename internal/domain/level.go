package domain

import (
	"fmt"
	"strings"
)

// Level represents a gamification level that users progress through.
type Level struct {
	Tier      int
	Name      string
	Icon      string
	MinPoints int
}

// Rank represents a season-scoped Solo-Leveling-inspired rank.
type Rank struct {
	Tier      int
	Name      string
	Icon      string
	MinPoints int
}

// JobClass represents a selectable RPG hunter job for a user profile.
type JobClass struct {
	ID          string
	Name        string
	Icon        string
	Description string
	Trait       string
}

// NumericLevelProgress represents persistent lifetime RPG level progress.
// The level itself is season-independent; only lifetime points move it forward.
type NumericLevelProgress struct {
	Level       int
	CurrentXP   int
	RequiredXP  int
	TotalPoints int
}

const (
	levelXPQuadratic = 5
	levelXPLinear    = 50
	levelXPBase      = 100
)

// XPForNextNumericLevel returns the EXP required to advance from level N to N+1.
// It follows the battle-tested quadratic chatbot curve: 5L² + 50L + 100.
// Level 0 starts at 100 EXP so new deployments can begin cleanly from Lv.0.
func XPForNextNumericLevel(level int) int {
	if level < 0 {
		level = 0
	}
	return levelXPQuadratic*level*level + levelXPLinear*level + levelXPBase
}

// GetNumericLevelProgress derives the persistent numeric RPG level from lifetime points.
// The existing points/EXP value remains the source of truth, so old data migrates safely.
func GetNumericLevelProgress(totalPoints int) NumericLevelProgress {
	if totalPoints < 0 {
		totalPoints = 0
	}

	level := 0
	remaining := totalPoints
	for {
		required := XPForNextNumericLevel(level)
		if remaining < required {
			return NumericLevelProgress{
				Level:       level,
				CurrentXP:   remaining,
				RequiredXP:  required,
				TotalPoints: totalPoints,
			}
		}
		remaining -= required
		level++
	}
}

// NumericLevelFromTotalPoints returns only the persistent numeric level.
func NumericLevelFromTotalPoints(totalPoints int) int {
	return GetNumericLevelProgress(totalPoints).Level
}

// AllLevels defines the progression tiers in ascending order.
// Curved to reward early engagement while making max tier a long-term achievement.
// With ~17 weeks per 4-month season, dedicated users can reach Tier 5-6 in season 1,
// Tier 7 in season 2-3, and Tier 8 in season 4+.
var AllLevels = []Level{
	{Tier: 1, Name: "Newbie", Icon: "🌱", MinPoints: 0},
	{Tier: 2, Name: "Fighter", Icon: "💪", MinPoints: 50},
	{Tier: 3, Name: "Warrior", Icon: "⚔️", MinPoints: 120},
	{Tier: 4, Name: "Champion", Icon: "🏆", MinPoints: 250},
	{Tier: 5, Name: "Legend", Icon: "👑", MinPoints: 500},
	{Tier: 6, Name: "Immortal", Icon: "🔱", MinPoints: 1000},
	{Tier: 7, Name: "Titan", Icon: "⭐", MinPoints: 2000},
	{Tier: 8, Name: "God", Icon: "⚡", MinPoints: 3500},
}

// AllSeasonRanks defines rank titles for the current season only.
// Level remains lifetime; rank resets with seasonal points.
var AllSeasonRanks = []Rank{
	{Tier: 1, Name: "E-Rank Hunter", Icon: "🟫", MinPoints: 0},
	{Tier: 2, Name: "D-Rank Hunter", Icon: "🟩", MinPoints: 100},
	{Tier: 3, Name: "C-Rank Hunter", Icon: "🟦", MinPoints: 250},
	{Tier: 4, Name: "B-Rank Hunter", Icon: "🟪", MinPoints: 450},
	{Tier: 5, Name: "A-Rank Hunter", Icon: "🟥", MinPoints: 700},
	{Tier: 6, Name: "S-Rank Hunter", Icon: "🟨", MinPoints: 1000},
	{Tier: 7, Name: "Monarch", Icon: "👑", MinPoints: 1500},
}

// AllJobClasses defines jobs inspired by Solo Leveling hunter roles and common
// RPG archetypes. Jobs are cosmetic profile flavor and do not reset per season.
var AllJobClasses = []JobClass{
	{ID: "fighter", Name: "Fighter", Icon: "⚔️", Description: "Melee hunter yang mengandalkan disiplin, stamina, dan daya tahan.", Trait: "cocok untuk yang suka latihan strength/functional"},
	{ID: "tank", Name: "Tanker", Icon: "🛡️", Description: "Frontliner yang kuat bertahan dan konsisten menjaga formasi.", Trait: "cocok untuk yang fokus konsistensi dan habit jangka panjang"},
	{ID: "assassin", Name: "Assassin", Icon: "🗡️", Description: "Hunter cepat, gesit, dan tajam mengeksekusi sesi singkat tapi intens.", Trait: "cocok untuk HIIT, sprint, atau workout cepat"},
	{ID: "mage", Name: "Mage", Icon: "🔥", Description: "Damage dealer jarak jauh dengan energi eksplosif dan variasi latihan.", Trait: "cocok untuk yang suka eksplor banyak jenis olahraga"},
	{ID: "ranger", Name: "Ranger", Icon: "🏹", Description: "Hunter presisi yang unggul di endurance, pace, dan jarak.", Trait: "cocok untuk lari, sepeda, jalan jauh, hiking"},
	{ID: "healer", Name: "Healer", Icon: "💚", Description: "Support hunter yang menjaga recovery, mobilitas, dan kesehatan jangka panjang.", Trait: "cocok untuk yoga, mobility, recovery, pola hidup sehat"},
	{ID: "necromancer", Name: "Necromancer", Icon: "🌑", Description: "Hidden job yang bangkit dari kegagalan dan mengubah comeback jadi kekuatan.", Trait: "cocok untuk comeback setelah absen dan bangun sistem baru"},
}

// GetLevel returns the current level for the given total points.
func GetLevel(totalPoints int) Level {
	current := AllLevels[0]
	for _, l := range AllLevels {
		if totalPoints >= l.MinPoints {
			current = l
		}
	}
	return current
}

// GetSeasonRank returns the current season rank for seasonal points.
func GetSeasonRank(seasonalPoints int) Rank {
	current := AllSeasonRanks[0]
	for _, r := range AllSeasonRanks {
		if seasonalPoints >= r.MinPoints {
			current = r
		}
	}
	return current
}

// FormatSeasonRank returns a display string like "C-Rank Hunter 🟦".
func FormatSeasonRank(seasonalPoints int) string {
	rank := GetSeasonRank(seasonalPoints)
	return fmt.Sprintf("%s %s", rank.Name, rank.Icon)
}

// GetJobClass returns a job class by id.
func GetJobClass(id string) (*JobClass, bool) {
	for _, job := range AllJobClasses {
		if job.ID == id {
			return &job, true
		}
	}
	return nil, false
}

// FormatJobClass returns the display string for a job id.
func FormatJobClass(id string) string {
	job, ok := GetJobClass(id)
	if !ok {
		return "Belum memilih job"
	}
	return fmt.Sprintf("%s %s", job.Name, job.Icon)
}

// GetNextLevel returns the next level and how many points are needed, or nil if max level.
func GetNextLevel(totalPoints int) (*Level, int) {
	for _, l := range AllLevels {
		if totalPoints < l.MinPoints {
			return &l, l.MinPoints - totalPoints
		}
	}
	return nil, 0
}

// FormatLevel returns a display string like "Fighter 💪"
func FormatLevel(totalPoints int) string {
	lvl := GetLevel(totalPoints)
	return fmt.Sprintf("%s %s", lvl.Name, lvl.Icon)
}

// FormatProgressBar returns a progress bar string toward the next level.
// Example: "[████░░░░░░] 85/200 pts"
func FormatProgressBar(totalPoints int) string {
	next, _ := GetNextLevel(totalPoints)
	if next == nil {
		return "MAX LEVEL! 🔱"
	}

	current := GetLevel(totalPoints)
	rangeTotal := next.MinPoints - current.MinPoints
	progress := totalPoints - current.MinPoints
	if rangeTotal <= 0 {
		rangeTotal = 1
	}

	barLen := 10
	filled := (progress * barLen) / rangeTotal
	if filled > barLen {
		filled = barLen
	}

	bar := ""
	for i := 0; i < barLen; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return fmt.Sprintf("[%s] %d/%d pts → %s %s", bar, totalPoints, next.MinPoints, next.Name, next.Icon)
}

// FormatNumericLevelProgressBar returns compact Solo-Leveling-style lifetime progress.
func FormatNumericLevelProgressBar(totalPoints int) string {
	progress := GetNumericLevelProgress(totalPoints)
	barLen := 10
	filled := 0
	if progress.RequiredXP > 0 {
		filled = (progress.CurrentXP * barLen) / progress.RequiredXP
	}
	if filled > barLen {
		filled = barLen
	}

	var bar strings.Builder
	for i := 0; i < barLen; i++ {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}

	return fmt.Sprintf("Lv.%d [%s] %d/%d EXP", progress.Level, bar.String(), progress.CurrentXP, progress.RequiredXP)
}
