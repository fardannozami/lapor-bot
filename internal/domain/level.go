package domain

import "fmt"

// Level represents a gamification level that users progress through.
type Level struct {
	Tier      int
	Name      string
	Icon      string
	MinPoints int
}

// AllLevels defines the progression tiers in ascending order.
var AllLevels = []Level{
	{Tier: 1, Name: "Newbie", Icon: "🌱", MinPoints: 0},
	{Tier: 2, Name: "Fighter", Icon: "💪", MinPoints: 30},
	{Tier: 3, Name: "Warrior", Icon: "⚔️", MinPoints: 80},
	{Tier: 4, Name: "Champion", Icon: "🏆", MinPoints: 200},
	{Tier: 5, Name: "Legend", Icon: "👑", MinPoints: 400},
	{Tier: 6, Name: "Immortal", Icon: "🔱", MinPoints: 700},
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
