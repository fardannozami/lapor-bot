package domain

import (
	"fmt"
	"hash/fnv"
	"strings"
	"time"
)

type QuestTask struct {
	ID           string `json:"id"`            // e.g. "pushup"
	Name         string `json:"name"`          // e.g. "Push-up"
	Difficulty   string `json:"difficulty"`    // "easy", "medium", or "hard"
	Target       int    `json:"target"`        // numerical target quantity
	Progress     int    `json:"progress"`      // current accumulated progress
	Unit         string `json:"unit"`          // e.g. "x", "menit", "detik", "km", "langkah", "ml"
	RewardPoints int    `json:"reward_points"` // points rewarded upon 100% completion
}

// GenerateDailyQuest deterministically generates daily side quest options.
// A selected job is required; users without a job do not receive side quests.
func GenerateDailyQuest(jobClass string, level int, date time.Time) []QuestTask {
	return GenerateDailyQuestForUser("", jobClass, level, date)
}

// GenerateDailyQuestForUser generates a stable daily quest list per user.
// The same user keeps the same list all day, while different users can receive
// different medium/hard starter movements even when they share the same job.
func GenerateDailyQuestForUser(userID, jobClass string, level int, date time.Time) []QuestTask {
	if strings.TrimSpace(jobClass) == "" {
		return nil
	}

	dateStr := date.Format("2006-01-02")
	h := fnv.New32a()
	h.Write([]byte(dateStr))
	h.Write([]byte(userID))
	h.Write([]byte(jobClass))
	hashValue := int(h.Sum32())

	// Medium/hard scale gently with hunter level so higher-ranked jobs get a
	// little more challenge without turning side quests into the main #lapor.
	if level < 0 {
		level = 0
	}
	mediumBonus := level / 5
	if mediumBonus > 3 {
		mediumBonus = 3
	}
	hardBonus := level / 4
	if hardBonus > 5 {
		hardBonus = 5
	}

	// Medium: light exercises doable at home/office as movement reminders.
	// Level bonus gently scales with hunter level, capped at Lv.15.
	mediumOptions := []QuestTask{
		{ID: "squat", Name: "Chair Squat / Sit-to-Stand", Difficulty: "medium", Target: 18 + mediumBonus*2, Unit: "x", RewardPoints: 5},
		{ID: "pushup", Name: "Wall Push-up / Desk Push-up", Difficulty: "medium", Target: 15 + mediumBonus*2, Unit: "x", RewardPoints: 5},
		{ID: "stretching", Name: "Office Stretching / Mobility", Difficulty: "medium", Target: 8 + mediumBonus, Unit: "menit", RewardPoints: 5},
		{ID: "highknees", Name: "High Knees (Marching)", Difficulty: "medium", Target: 1 + mediumBonus, Unit: "menit", RewardPoints: 5},
		{ID: "armcircles", Name: "Arm Circles", Difficulty: "medium", Target: 2 + mediumBonus, Unit: "menit", RewardPoints: 5},
		{ID: "calfraises", Name: "Calf Raises (Jinjit)", Difficulty: "medium", Target: 20 + mediumBonus*2, Unit: "x", RewardPoints: 5},
		{ID: "stairs", Name: "Naik-Turun Tangga", Difficulty: "medium", Target: 5 + mediumBonus, Unit: "menit", RewardPoints: 5},
		{ID: "deskplank", Name: "Desk Plank", Difficulty: "medium", Target: 30 + mediumBonus*5, Unit: "detik", RewardPoints: 5},
	}
	// Hard: slightly more effort, still home-friendly. Level bonus caps at Lv.20.
	hardOptions := []QuestTask{
		{ID: "plank", Name: "Plank", Difficulty: "hard", Target: 45 + hardBonus*5, Unit: "detik", RewardPoints: 5},
		{ID: "jumpingjacks", Name: "Jumping Jacks", Difficulty: "hard", Target: 35 + hardBonus*3, Unit: "x", RewardPoints: 5},
		{ID: "squatjump", Name: "Squat Jump Ringan", Difficulty: "hard", Target: 8 + hardBonus, Unit: "x", RewardPoints: 5},
		{ID: "yoga", Name: "Mobility Flow", Difficulty: "hard", Target: 12 + hardBonus, Unit: "menit", RewardPoints: 5},
		{ID: "mountainclimber", Name: "Mountain Climber", Difficulty: "hard", Target: 30 + hardBonus*5, Unit: "detik", RewardPoints: 5},
		{ID: "lunges", Name: "Lunges (Bergantian)", Difficulty: "hard", Target: 10 + hardBonus, Unit: "x per kaki", RewardPoints: 5},
		{ID: "glutebridge", Name: "Glute Bridge", Difficulty: "hard", Target: 15 + hardBonus*2, Unit: "x", RewardPoints: 5},
		{ID: "reversecrunch", Name: "Reverse Crunch", Difficulty: "hard", Target: 10 + hardBonus, Unit: "x", RewardPoints: 5},
		{ID: "lateralshuffle", Name: "Lateral Shuffle", Difficulty: "hard", Target: 45 + hardBonus*5, Unit: "detik", RewardPoints: 5},
	}

	return []QuestTask{
		{ID: "easycardio", Name: "Jalan Kaki / Bersepeda", Difficulty: "easy", Target: 1, Unit: "opsi", RewardPoints: 5},
		mediumOptions[hashValue%len(mediumOptions)],
		hardOptions[(hashValue/len(mediumOptions))%len(hardOptions)],
	}
}

// MatchTask checks if input contains keywords matching a specific task ID.
func MatchTask(input string, taskID string) bool {
	input = strings.ToLower(strings.TrimSpace(input))
	switch taskID {
	case "pushup":
		return strings.Contains(input, "push")
	case "situp":
		return strings.Contains(input, "sit")
	case "squat":
		return strings.Contains(input, "squat")
	case "plank":
		return strings.Contains(input, "plank")
	case "burpee":
		return strings.Contains(input, "burpee")
	case "easycardio":
		return MatchTask(input, "jalan") || MatchTask(input, "sepeda")
	case "jalan":
		return strings.Contains(input, "jalan") || strings.Contains(input, "walk")
	case "lari":
		return strings.Contains(input, "lari") || strings.Contains(input, "run") || strings.Contains(input, "sprint")
	case "sepeda":
		return strings.Contains(input, "sepeda") || strings.Contains(input, "cycle") || strings.Contains(input, "cycling")
	case "hiit":
		return strings.Contains(input, "hiit") || strings.Contains(input, "tabata")
	case "stretching":
		return strings.Contains(input, "stretch") || strings.Contains(input, "peregangan")
	case "yoga":
		return strings.Contains(input, "yoga") || strings.Contains(input, "mobility")
	case "meditasi":
		return strings.Contains(input, "meditasi") || strings.Contains(input, "meditate") || strings.Contains(input, "meditation")
	case "air":
		return strings.Contains(input, "air") || strings.Contains(input, "minum") || strings.Contains(input, "water")
	case "shadowboxing":
		return strings.Contains(input, "box") || strings.Contains(input, "shadow")
	case "jumpingjacks":
		return strings.Contains(input, "jumping") || strings.Contains(input, "jack")
	case "deepbreathing":
		return strings.Contains(input, "napas") || strings.Contains(input, "breath") || strings.Contains(input, "breathing")
	case "cardio":
		return strings.Contains(input, "cardio") || strings.Contains(input, "kardio")
	case "weight":
		return strings.Contains(input, "beban") || strings.Contains(input, "weight") || strings.Contains(input, "strength")
	// Medium additions — home/office movement reminders
	case "highknees":
		return strings.Contains(input, "high knee") || strings.Contains(input, "marching") || strings.Contains(input, "knee") && strings.Contains(input, "high")
	case "armcircles":
		return strings.Contains(input, "arm circle") || strings.Contains(input, "circle") || strings.Contains(input, "putaran lengan")
	case "calfraises":
		return strings.Contains(input, "calf") || strings.Contains(input, "jinjit")
	case "shouldershrugs":
		return strings.Contains(input, "shrug") || strings.Contains(input, "bahu")
	case "stairs":
		return strings.Contains(input, "tangga") || strings.Contains(input, "stair")
	case "deskplank":
		return strings.Contains(input, "desk plank") || strings.Contains(input, "plank meja")
	// Hard additions — slightly more effort, still home-friendly
	case "mountainclimber":
		return strings.Contains(input, "mountain") || strings.Contains(input, "climber") || strings.Contains(input, "panjat")
	case "lunges":
		return strings.Contains(input, "lunge") || strings.Contains(input, "lunjak")
	case "glutebridge":
		return strings.Contains(input, "glute") || strings.Contains(input, "bridge") || strings.Contains(input, "pinggul") || strings.Contains(input, "jembatan")
	case "reversecrunch":
		return strings.Contains(input, "reverse crunch") || strings.Contains(input, "crunch") && !strings.Contains(input, "bicycle")
	case "squatjump":
		return strings.Contains(input, "squat jump") || strings.Contains(input, "jump squat")
	case "lateralshuffle":
		return strings.Contains(input, "lateral") || strings.Contains(input, "shuffle")
	}
	return false
}

// FormatQuestProgressTask returns a descriptive string for a quest task, resolving Ranger 100m units to km.
func FormatQuestProgressTask(task QuestTask) string {
	difficulty := formatQuestDifficulty(task.Difficulty)
	if task.ID == "easycardio" {
		return fmt.Sprintf("%s %s: %d/%d selesai (jalan kaki 4000 langkah atau sepeda 5 km)", difficulty, task.Name, task.Progress, task.Target)
	}
	if task.Unit == "100m" {
		targetKm := float64(task.Target) / 10.0
		progressKm := float64(task.Progress) / 10.0
		return fmt.Sprintf("%s %s: %.1f/%.1f km", difficulty, task.Name, progressKm, targetKm)
	}
	return fmt.Sprintf("%s %s: %d/%d %s", difficulty, task.Name, task.Progress, task.Target, task.Unit)
}

// FormatQuestTask returns a descriptive string for a quest task without showing progress (showing only target).
func FormatQuestTask(task QuestTask) string {
	difficulty := formatQuestDifficulty(task.Difficulty)
	if task.ID == "easycardio" {
		return fmt.Sprintf("%s %s: jalan kaki 4000 langkah atau sepeda 5 km", difficulty, task.Name)
	}
	if task.Unit == "100m" {
		targetKm := float64(task.Target) / 10.0
		return fmt.Sprintf("%s %s: %.1f km", difficulty, task.Name, targetKm)
	}
	return fmt.Sprintf("%s %s: %d %s", difficulty, task.Name, task.Target, task.Unit)
}

func formatQuestDifficulty(difficulty string) string {
	switch strings.ToLower(strings.TrimSpace(difficulty)) {
	case "easy":
		return "🟢 Easy •"
	case "medium":
		return "🟡 Medium •"
	case "hard":
		return "🔴 Hard •"
	default:
		return ""
	}
}
