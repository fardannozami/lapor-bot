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

// Rank represents a season-scoped competitive rank.
type Rank struct {
	Tier      int
	Name      string
	Icon      string
	MinPoints int
}

// JobClass represents a selectable RPG hunter job for a user profile.
type JobClass struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	Trait       string `json:"trait"`
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

// AllLevels defines lifetime progression tiers in ascending order.
// Lifetime points never reset, so S-tier is intentionally a long-term journey
// across many seasons instead of a single-season finish line.
var AllLevels = []Level{
	{Tier: 1, Name: "E-Tier Hunter", Icon: "🟫", MinPoints: 0},
	{Tier: 2, Name: "D-Tier Hunter", Icon: "🟩", MinPoints: 1150},
	{Tier: 3, Name: "C-Tier Hunter", Icon: "🟦", MinPoints: 4675},
	{Tier: 4, Name: "B-Tier Hunter", Icon: "🟪", MinPoints: 11825},
	{Tier: 5, Name: "A-Tier Hunter", Icon: "🟥", MinPoints: 23850},
	{Tier: 6, Name: "S-Tier Hunter", Icon: "🟨", MinPoints: 42000},
}

// AllSeasonRanks defines Mobile-Legends-inspired rank titles for the current
// season only. Level remains lifetime; rank resets with seasonal points.
var AllSeasonRanks = []Rank{
	{Tier: 1, Name: "Warrior", Icon: "🛡️", MinPoints: 0},
	{Tier: 2, Name: "Elite", Icon: "⚔️", MinPoints: 150},
	{Tier: 3, Name: "Master", Icon: "🏹", MinPoints: 350},
	{Tier: 4, Name: "Grandmaster", Icon: "🔥", MinPoints: 650},
	{Tier: 5, Name: "Epic", Icon: "💎", MinPoints: 900},
	{Tier: 6, Name: "Legend", Icon: "👑", MinPoints: 1250},
	{Tier: 7, Name: "Mythic", Icon: "🌙", MinPoints: 1700},
	{Tier: 8, Name: "Mythical Honor", Icon: "🐉", MinPoints: 2350},
	{Tier: 9, Name: "Mythical Glory", Icon: "✨", MinPoints: 3200},
	{Tier: 10, Name: "Mythical Immortal", Icon: "🌟", MinPoints: 4500},
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

// JobClassPrimaryAttribute returns the attribute specialty used for daily
// side-quest rotation. Mage stays mixed by returning an empty attribute.
func JobClassPrimaryAttribute(jobClass string) AttributeType {
	switch strings.ToLower(strings.TrimSpace(jobClass)) {
	case "fighter":
		return AttrStr
	case "ranger":
		return AttrSta
	case "assassin":
		return AttrAgi
	case "tank", "healer", "necromancer":
		return AttrVit
	default:
		return ""
	}
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

// GetNextSeasonRank returns the next season rank and missing seasonal points, or nil at max rank.
func GetNextSeasonRank(seasonalPoints int) (*Rank, int) {
	for _, r := range AllSeasonRanks {
		if seasonalPoints < r.MinPoints {
			return &r, r.MinPoints - seasonalPoints
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
		current := GetLevel(totalPoints)
		return fmt.Sprintf("MAX LEVEL! %s", current.Icon)
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
