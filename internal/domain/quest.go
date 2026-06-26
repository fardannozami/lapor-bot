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

	mediumOptions, hardOptions := questOptionsForJob(jobClass, mediumBonus, hardBonus)

	mediumTask := mediumOptions[hashValue%len(mediumOptions)]
	hardTask := selectNonConflictingQuest(mediumTask, hardOptions, (hashValue/len(mediumOptions))%len(hardOptions))

	return []QuestTask{
		{ID: "easycardio", Name: "Jalan Kaki / Bersepeda", Difficulty: "easy", Target: 1, Unit: "opsi", RewardPoints: 5},
		mediumTask,
		hardTask,
	}
}

func selectNonConflictingQuest(medium QuestTask, hardOptions []QuestTask, start int) QuestTask {
	for offset := 0; offset < len(hardOptions); offset++ {
		candidate := hardOptions[(start+offset)%len(hardOptions)]
		if !questTasksConflict(medium, candidate) {
			return candidate
		}
	}
	return hardOptions[start%len(hardOptions)]
}

func questTasksConflict(a, b QuestTask) bool {
	return a.ID == b.ID || MatchTask(a.Name, b.ID) || MatchTask(b.Name, a.ID)
}

func questOptionsForJob(jobClass string, mediumBonus, hardBonus int) ([]QuestTask, []QuestTask) {
	mediumByAttr := map[AttributeType][]QuestTask{
		AttrStr: {
			{ID: "squat", Name: "Chair Squat / Sit-to-Stand", Difficulty: "medium", Target: 18 + mediumBonus*2, Unit: "x", RewardPoints: 5},
			{ID: "pushup", Name: "Wall Push-up / Desk Push-up", Difficulty: "medium", Target: 15 + mediumBonus*2, Unit: "x", RewardPoints: 5},
			{ID: "calfraises", Name: "Calf Raises (Jinjit)", Difficulty: "medium", Target: 20 + mediumBonus*2, Unit: "x", RewardPoints: 5},
			{ID: "deskplank", Name: "Desk Plank", Difficulty: "medium", Target: 30 + mediumBonus*5, Unit: "detik", RewardPoints: 5},
			{ID: "wallsit", Name: "Wall Sit", Difficulty: "medium", Target: 30 + mediumBonus*5, Unit: "detik", RewardPoints: 5},
		},
		AttrSta: {
			{ID: "stairs", Name: "Naik-Turun Tangga", Difficulty: "medium", Target: 5 + mediumBonus, Unit: "menit", RewardPoints: 5},
			{ID: "stepups", Name: "Step-up", Difficulty: "medium", Target: 16 + mediumBonus*2, Unit: "x", RewardPoints: 5},
			{ID: "jumpingjacks", Name: "Jumping Jacks", Difficulty: "medium", Target: 25 + mediumBonus*3, Unit: "x", RewardPoints: 5},
			{ID: "skipping", Name: "Skipping / Rope Jump", Difficulty: "medium", Target: 2 + mediumBonus, Unit: "menit", RewardPoints: 5},
		},
		AttrAgi: {
			{ID: "highknees", Name: "High Knees (Marching)", Difficulty: "medium", Target: 1 + mediumBonus, Unit: "menit", RewardPoints: 5},
			{ID: "shadowboxing", Name: "Shadow Boxing", Difficulty: "medium", Target: 2 + mediumBonus, Unit: "menit", RewardPoints: 5},
			{ID: "lateralshuffle", Name: "Lateral Shuffle", Difficulty: "medium", Target: 30 + mediumBonus*5, Unit: "detik", RewardPoints: 5},
			{ID: "ladderdrill", Name: "Ladder Drill", Difficulty: "medium", Target: 2 + mediumBonus, Unit: "menit", RewardPoints: 5},
		},
		AttrVit: {
			{ID: "stretching", Name: "Office Stretching / Mobility", Difficulty: "medium", Target: 8 + mediumBonus, Unit: "menit", RewardPoints: 5},
			{ID: "armcircles", Name: "Arm Circles", Difficulty: "medium", Target: 2 + mediumBonus, Unit: "menit", RewardPoints: 5},
			{ID: "deepbreathing", Name: "Deep Breathing", Difficulty: "medium", Target: 3 + mediumBonus, Unit: "menit", RewardPoints: 5},
			{ID: "birddog", Name: "Bird Dog", Difficulty: "medium", Target: 12 + mediumBonus*2, Unit: "x", RewardPoints: 5},
			{ID: "balance", Name: "Balance Drill", Difficulty: "medium", Target: 2 + mediumBonus, Unit: "menit", RewardPoints: 5},
		},
	}

	hardByAttr := map[AttributeType][]QuestTask{
		AttrStr: {
			{ID: "plank", Name: "Plank", Difficulty: "hard", Target: 45 + hardBonus*5, Unit: "detik", RewardPoints: 5},
			{ID: "lunges", Name: "Lunges (Bergantian)", Difficulty: "hard", Target: 10 + hardBonus, Unit: "x per kaki", RewardPoints: 5},
			{ID: "glutebridge", Name: "Glute Bridge", Difficulty: "hard", Target: 15 + hardBonus*2, Unit: "x", RewardPoints: 5},
			{ID: "reversecrunch", Name: "Reverse Crunch", Difficulty: "hard", Target: 10 + hardBonus, Unit: "x", RewardPoints: 5},
			{ID: "deadbug", Name: "Dead Bug", Difficulty: "hard", Target: 12 + hardBonus*2, Unit: "x", RewardPoints: 5},
		},
		AttrSta: {
			{ID: "mountainclimber", Name: "Mountain Climber", Difficulty: "hard", Target: 30 + hardBonus*5, Unit: "detik", RewardPoints: 5},
			{ID: "jumpingjacks", Name: "Jumping Jacks", Difficulty: "hard", Target: 35 + hardBonus*3, Unit: "x", RewardPoints: 5},
			{ID: "stairs", Name: "Naik-Turun Tangga", Difficulty: "hard", Target: 8 + hardBonus, Unit: "menit", RewardPoints: 5},
			{ID: "shuttlerun", Name: "Shuttle Run", Difficulty: "hard", Target: 4 + hardBonus, Unit: "menit", RewardPoints: 5},
		},
		AttrAgi: {
			{ID: "squatjump", Name: "Squat Jump Ringan", Difficulty: "hard", Target: 8 + hardBonus, Unit: "x", RewardPoints: 5},
			{ID: "mountainclimber", Name: "Mountain Climber", Difficulty: "hard", Target: 30 + hardBonus*5, Unit: "detik", RewardPoints: 5},
			{ID: "lateralshuffle", Name: "Lateral Shuffle", Difficulty: "hard", Target: 45 + hardBonus*5, Unit: "detik", RewardPoints: 5},
			{ID: "shadowboxing", Name: "Shadow Boxing", Difficulty: "hard", Target: 4 + hardBonus, Unit: "menit", RewardPoints: 5},
		},
		AttrVit: {
			{ID: "yoga", Name: "Mobility Flow", Difficulty: "hard", Target: 12 + hardBonus, Unit: "menit", RewardPoints: 5},
			{ID: "stretching", Name: "Full Body Stretching", Difficulty: "hard", Target: 12 + hardBonus, Unit: "menit", RewardPoints: 5},
			{ID: "deepbreathing", Name: "Deep Breathing", Difficulty: "hard", Target: 6 + hardBonus, Unit: "menit", RewardPoints: 5},
			{ID: "balance", Name: "Balance Drill", Difficulty: "hard", Target: 4 + hardBonus, Unit: "menit", RewardPoints: 5},
		},
	}

	primary := JobClassPrimaryAttribute(jobClass)
	if primary != "" {
		return mediumByAttr[primary], hardByAttr[primary]
	}

	var mediumOptions []QuestTask
	var hardOptions []QuestTask
	for _, attr := range []AttributeType{AttrStr, AttrSta, AttrAgi, AttrVit} {
		mediumOptions = append(mediumOptions, mediumByAttr[attr]...)
		hardOptions = append(hardOptions, hardByAttr[attr]...)
	}
	return mediumOptions, hardOptions
}

var questTaskKeywords = map[string][]string{
	"pushup":          {"push up", "pushup", "wall push up", "desk push up"},
	"situp":           {"sit up", "situp", "situps"},
	"squat":           {"squat", "chair squat", "sit to stand"},
	"plank":           {"plank"},
	"burpee":          {"burpee", "burpees"},
	"deskplank":       {"desk plank", "plank meja"},
	"easycardio":      {"jalan", "walk", "walking", "jalan kaki", "sepeda", "bersepeda", "cycle", "cycling", "bike"},
	"jalan":           {"jalan", "walk", "walking", "jalan kaki"},
	"lari":            {"lari", "run", "running", "jogging", "sprint"},
	"sepeda":          {"sepeda", "bersepeda", "cycle", "cycling", "bike"},
	"hiit":            {"hiit", "tabata"},
	"stretching":      {"stretch", "stretching", "peregangan", "mobility", "full body stretching"},
	"yoga":            {"yoga", "mobility flow"},
	"meditasi":        {"meditasi", "meditate", "meditation"},
	"air":             {"air putih", "minum air", "water"},
	"shadowboxing":    {"shadow boxing", "shadowboxing", "boxing", "tinju"},
	"jumpingjacks":    {"jumping jacks", "jumpingjacks"},
	"deepbreathing":   {"deep breathing", "breathing", "napas"},
	"cardio":          {"cardio", "kardio"},
	"weight":          {"beban", "weight", "strength", "weightlifting"},
	"highknees":       {"high knees", "highknees", "marching"},
	"armcircles":      {"arm circle", "arm circles", "armcircles", "putaran lengan"},
	"calfraises":      {"calf raises", "calf raise", "calfraises", "jinjit"},
	"shouldershrugs":  {"shoulder shrug", "shoulder shrugs", "shrug bahu"},
	"stairs":          {"stairs", "tangga", "naik turun tangga"},
	"mountainclimber": {"mountain climber", "mountainclimber", "gerakan panjat", "panjat"},
	"lunges":          {"lunges", "lunge", "lunjak"},
	"glutebridge":     {"glute bridge", "glutebridge", "jembatan pinggul"},
	"reversecrunch":   {"reverse crunch", "reversecrunch"},
	"squatjump":       {"squat jump", "jump squat"},
	"lateralshuffle":  {"lateral shuffle", "lateralshuffle"},
	"wallsit":         {"wall sit", "wallsit"},
	"stepups":         {"step up", "stepup"},
	"skipping":        {"skipping", "rope jump"},
	"ladderdrill":     {"ladder drill", "ladderdrill"},
	"shuttlerun":      {"shuttle run", "shuttlerun"},
	"deadbug":         {"dead bug", "deadbug"},
	"birddog":         {"bird dog", "birddog"},
	"balance":         {"balance", "balance drill", "keseimbangan"},
}

// MatchTask checks if input contains keywords matching a specific task ID.
func MatchTask(input string, taskID string) bool {
	return containsAnyActivityKeyword(normalizeActivityText(input), questTaskKeywords[taskID])
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
