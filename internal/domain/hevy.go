package domain

import (
	"regexp"
	"strings"
)

type HevyWorkout struct {
	Title     string
	Exercises []string
	Time      string
}

func ParseHevy(message string) *HevyWorkout {
	if !strings.Contains(message, "Hevy") {
		return nil
	}

	lines := strings.Split(message, "\n")
	workout := &HevyWorkout{}

	// Regex to match "Today at 9:46 AM · Hevy" or similar
	hevyHeader := regexp.MustCompile(`.*(· Hevy|Logged with Hevy|Hevy)`)
	
	for i, line := range lines {
		if hevyHeader.MatchString(line) {
			// Usually the title is the very next non-empty line
			if i+1 < len(lines) {
				workout.Title = strings.TrimSpace(lines[i+1])
			}
			break
		}
	}

	// If header not found or weird, fallback to finding title before "Logged with Hevy"
	if workout.Title == "" || strings.Contains(workout.Title, "Logged with Hevy") {
		for i, line := range lines {
			if strings.Contains(line, "Logged with Hevy") && i > 0 {
				workout.Title = strings.TrimSpace(lines[i-1])
				break
			}
		}
	}

	// Extract exercises and time
	exerciseMap := make(map[string]bool)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Find Time
		if strings.HasPrefix(trimmed, "Time") && i+1 < len(lines) {
			workout.Time = strings.TrimSpace(lines[i+1])
			continue
		}

		// Exercises are lines followed by "Set 1"
		if strings.HasPrefix(trimmed, "Set 1:") && i > 0 {
			// The line before "Set 1" is the exercise name
			exercise := strings.TrimSpace(lines[i-1])
			if exercise != "" && !exerciseMap[exercise] && !strings.Contains(exercise, "Logged with Hevy") {
				workout.Exercises = append(workout.Exercises, exercise)
				exerciseMap[exercise] = true
			}
		}
	}

	return workout
}
