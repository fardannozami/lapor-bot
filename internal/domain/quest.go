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
	Target       int    `json:"target"`        // numerical target quantity
	Progress     int    `json:"progress"`      // current accumulated progress
	Unit         string `json:"unit"`          // e.g. "x", "menit", "detik", "km", "langkah", "ml"
	RewardPoints int    `json:"reward_points"` // points rewarded upon 100% completion
}

// GenerateDailyQuest deterministically generates a list of 3 tasks based on jobClass, level, and date.
func GenerateDailyQuest(jobClass string, level int, date time.Time) []QuestTask {
	dateStr := date.Format("2006-01-02")
	h := fnv.New32a()
	h.Write([]byte(dateStr))
	h.Write([]byte(jobClass))
	hashValue := int(h.Sum32())

	// Points scaling: base reward + (level/5), capped at base + 5
	var baseReward int
	if jobClass == "" {
		baseReward = 2
	} else {
		baseReward = 3
	}
	reward := baseReward + (level / 5)
	if reward > baseReward+5 {
		reward = baseReward + 5
	}

	scale := func(base int, multiplier int) int {
		return base + (level * multiplier)
	}

	var tasks []QuestTask

	switch jobClass {
	case "fighter":
		templates := [][]QuestTask{
			{
				{ID: "pushup", Name: "Push-up", Target: scale(10, 1), Unit: "x", RewardPoints: reward},
				{ID: "squat", Name: "Squat", Target: scale(15, 1), Unit: "x", RewardPoints: reward},
				{ID: "plank", Name: "Plank", Target: scale(30, 3), Unit: "detik", RewardPoints: reward},
			},
			{
				{ID: "burpee", Name: "Burpee", Target: scale(5, 1), Unit: "x", RewardPoints: reward},
				{ID: "pushup", Name: "Push-up", Target: scale(12, 1), Unit: "x", RewardPoints: reward},
				{ID: "squat", Name: "Squat", Target: scale(20, 1), Unit: "x", RewardPoints: reward},
			},
			{
				{ID: "weight", Name: "Latihan Beban (Weight/Strength)", Target: scale(15, 1), Unit: "menit", RewardPoints: reward},
				{ID: "plank", Name: "Plank", Target: scale(40, 2), Unit: "detik", RewardPoints: reward},
				{ID: "pushup", Name: "Push-up", Target: scale(8, 1), Unit: "x", RewardPoints: reward},
			},
		}
		tasks = templates[hashValue%len(templates)]

	case "tank":
		templates := [][]QuestTask{
			{
				{ID: "air", Name: "Minum Air Putih", Target: scale(2000, 50), Unit: "ml", RewardPoints: reward},
				{ID: "squat", Name: "Squat", Target: scale(10, 1), Unit: "x", RewardPoints: reward},
				{ID: "stretching", Name: "Stretching Ringan", Target: scale(10, 1), Unit: "menit", RewardPoints: reward},
			},
			{
				{ID: "stretching", Name: "Stretching Seluruh Tubuh", Target: scale(15, 1), Unit: "menit", RewardPoints: reward},
				{ID: "plank", Name: "Plank", Target: scale(20, 2), Unit: "detik", RewardPoints: reward},
				{ID: "air", Name: "Minum Air Putih", Target: scale(2500, 50), Unit: "ml", RewardPoints: reward},
			},
			{
				{ID: "deepbreathing", Name: "Latihan Pernapasan Dalam", Target: scale(5, 1), Unit: "menit", RewardPoints: reward},
				{ID: "plank", Name: "Plank", Target: scale(30, 2), Unit: "detik", RewardPoints: reward},
				{ID: "squat", Name: "Squat", Target: scale(12, 1), Unit: "x", RewardPoints: reward},
			},
		}
		tasks = templates[hashValue%len(templates)]

	case "assassin":
		templates := [][]QuestTask{
			{
				{ID: "hiit", Name: "HIIT Workout", Target: scale(10, 1), Unit: "menit", RewardPoints: reward},
				{ID: "jumpingjacks", Name: "Jumping Jacks", Target: scale(30, 2), Unit: "x", RewardPoints: reward},
				{ID: "plank", Name: "Plank", Target: scale(30, 2), Unit: "detik", RewardPoints: reward},
			},
			{
				{ID: "lari", Name: "Sprint Pendek", Target: scale(4, 1), Unit: "kali", RewardPoints: reward},
				{ID: "jumpingjacks", Name: "Jumping Jacks", Target: scale(40, 2), Unit: "x", RewardPoints: reward},
				{ID: "hiit", Name: "HIIT Workout", Target: scale(12, 1), Unit: "menit", RewardPoints: reward},
			},
			{
				{ID: "burpee", Name: "Burpee", Target: scale(8, 1), Unit: "x", RewardPoints: reward},
				{ID: "plank", Name: "Plank", Target: scale(45, 2), Unit: "detik", RewardPoints: reward},
				{ID: "jumpingjacks", Name: "Jumping Jacks", Target: scale(25, 2), Unit: "x", RewardPoints: reward},
			},
		}
		tasks = templates[hashValue%len(templates)]

	case "mage":
		templates := [][]QuestTask{
			{
				{ID: "cardio", Name: "Variasi Latihan Campuran", Target: scale(15, 1), Unit: "menit", RewardPoints: reward},
				{ID: "pushup", Name: "Push-up", Target: scale(8, 1), Unit: "x", RewardPoints: reward},
				{ID: "squat", Name: "Squat", Target: scale(10, 1), Unit: "x", RewardPoints: reward},
			},
			{
				{ID: "stretching", Name: "Coba Gerakan/Latihan Baru", Target: scale(10, 1), Unit: "menit", RewardPoints: reward},
				{ID: "plank", Name: "Plank", Target: scale(30, 2), Unit: "detik", RewardPoints: reward},
				{ID: "shadowboxing", Name: "Shadow Boxing", Target: scale(10, 1), Unit: "menit", RewardPoints: reward},
			},
			{
				{ID: "shadowboxing", Name: "Shadow Boxing", Target: scale(12, 1), Unit: "menit", RewardPoints: reward},
				{ID: "pushup", Name: "Push-up", Target: scale(10, 1), Unit: "x", RewardPoints: reward},
				{ID: "squat", Name: "Squat", Target: scale(12, 1), Unit: "x", RewardPoints: reward},
			},
		}
		tasks = templates[hashValue%len(templates)]

	case "ranger":
		templates := [][]QuestTask{
			{
				{ID: "lari", Name: "Lari / Jogging", Target: scale(20, 2), Unit: "100m", RewardPoints: reward}, // Target km: scale(20, 2) / 10 = 2.0 + 0.2*level km
				{ID: "jalan", Name: "Jalan Kaki Harian", Target: scale(5000, 250), Unit: "langkah", RewardPoints: reward},
				{ID: "squat", Name: "Squat", Target: scale(15, 1), Unit: "x", RewardPoints: reward},
			},
			{
				{ID: "sepeda", Name: "Bersepeda", Target: scale(50, 5), Unit: "100m", RewardPoints: reward}, // scale(50, 5)/10 = 5.0 + 0.5*level km
				{ID: "jalan", Name: "Jalan Kaki Harian", Target: scale(6000, 250), Unit: "langkah", RewardPoints: reward},
				{ID: "plank", Name: "Plank", Target: scale(25, 2), Unit: "detik", RewardPoints: reward},
			},
			{
				{ID: "lari", Name: "Lari / Jogging", Target: scale(30, 2), Unit: "100m", RewardPoints: reward},
				{ID: "squat", Name: "Squat", Target: scale(12, 1), Unit: "x", RewardPoints: reward},
				{ID: "jalan", Name: "Jalan Kaki Harian", Target: scale(4000, 250), Unit: "langkah", RewardPoints: reward},
			},
		}
		// Adjust target values to readable units in description formatting
		tasks = make([]QuestTask, len(templates[hashValue%len(templates)]))
		copy(tasks, templates[hashValue%len(templates)])

	case "healer":
		templates := [][]QuestTask{
			{
				{ID: "yoga", Name: "Yoga / Mobility Flow", Target: scale(10, 1), Unit: "menit", RewardPoints: reward},
				{ID: "deepbreathing", Name: "Latihan Pernapasan & Meditasi", Target: scale(5, 1), Unit: "menit", RewardPoints: reward},
				{ID: "stretching", Name: "Peregangan Statis Pasca Bangun", Target: scale(5, 1), Unit: "menit", RewardPoints: reward},
			},
			{
				{ID: "stretching", Name: "Pereda Ketegangan Leher & Punggung", Target: scale(10, 1), Unit: "menit", RewardPoints: reward},
				{ID: "yoga", Name: "Yoga / Mobility Flow", Target: scale(12, 1), Unit: "menit", RewardPoints: reward},
				{ID: "deepbreathing", Name: "Latihan Pernapasan & Meditasi", Target: scale(8, 1), Unit: "menit", RewardPoints: reward},
			},
			{
				{ID: "meditasi", Name: "Mindfulness Meditation", Target: scale(10, 1), Unit: "menit", RewardPoints: reward},
				{ID: "stretching", Name: "Yoga / Mobility Flow", Target: scale(8, 1), Unit: "menit", RewardPoints: reward},
				{ID: "deepbreathing", Name: "Deep Breathing", Target: scale(5, 1), Unit: "menit", RewardPoints: reward},
			},
		}
		tasks = templates[hashValue%len(templates)]

	case "necromancer":
		templates := [][]QuestTask{
			{
				{ID: "stretching", Name: "Pemanasan Sendi & Otot", Target: scale(5, 1), Unit: "menit", RewardPoints: reward},
				{ID: "situp", Name: "Sit-up", Target: scale(10, 1), Unit: "x", RewardPoints: reward},
				{ID: "jalan", Name: "Jalan Kaki Ringan", Target: scale(15, 1), Unit: "menit", RewardPoints: reward},
			},
			{
				{ID: "jalan", Name: "Jalan Kaki Ringan", Target: scale(20, 1), Unit: "menit", RewardPoints: reward},
				{ID: "squat", Name: "Squat", Target: scale(12, 1), Unit: "x", RewardPoints: reward},
				{ID: "situp", Name: "Sit-up", Target: scale(12, 1), Unit: "x", RewardPoints: reward},
			},
			{
				{ID: "stretching", Name: "Peregangan Pemulihan", Target: scale(10, 1), Unit: "menit", RewardPoints: reward},
				{ID: "situp", Name: "Sit-up", Target: scale(8, 1), Unit: "x", RewardPoints: reward},
				{ID: "jalan", Name: "Jalan Kaki Ringan", Target: scale(15, 1), Unit: "menit", RewardPoints: reward},
			},
		}
		tasks = templates[hashValue%len(templates)]

	default: // general / no job
		templates := [][]QuestTask{
			{
				{ID: "jalan", Name: "Jalan Kaki Santai", Target: scale(15, 1), Unit: "menit", RewardPoints: reward},
				{ID: "pushup", Name: "Push-up / Wall Push-up", Target: scale(5, 1), Unit: "x", RewardPoints: reward},
				{ID: "stretching", Name: "Peregangan Seluruh Tubuh", Target: scale(5, 1), Unit: "menit", RewardPoints: reward},
			},
			{
				{ID: "stretching", Name: "Peregangan Seluruh Tubuh", Target: scale(8, 1), Unit: "menit", RewardPoints: reward},
				{ID: "squat", Name: "Squat Ringan", Target: scale(8, 1), Unit: "x", RewardPoints: reward},
				{ID: "jalan", Name: "Jalan Kaki Santai", Target: scale(20, 1), Unit: "menit", RewardPoints: reward},
			},
			{
				{ID: "jalan", Name: "Jalan Kaki Santai", Target: scale(10, 1), Unit: "menit", RewardPoints: reward},
				{ID: "pushup", Name: "Push-up / Wall Push-up", Target: scale(6, 1), Unit: "x", RewardPoints: reward},
				{ID: "squat", Name: "Squat Ringan", Target: scale(6, 1), Unit: "x", RewardPoints: reward},
			},
		}
		tasks = templates[hashValue%len(templates)]
	}

	return tasks
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
	}
	return false
}

// FormatQuestProgressTask returns a descriptive string for a quest task, resolving Ranger 100m units to km.
func FormatQuestProgressTask(task QuestTask) string {
	if task.Unit == "100m" {
		targetKm := float64(task.Target) / 10.0
		progressKm := float64(task.Progress) / 10.0
		return fmt.Sprintf("%s: %.1f/%.1f km (+%d pts)", task.Name, progressKm, targetKm, task.RewardPoints)
	}
	return fmt.Sprintf("%s: %d/%d %s (+%d pts)", task.Name, task.Progress, task.Target, task.Unit, task.RewardPoints)
}
