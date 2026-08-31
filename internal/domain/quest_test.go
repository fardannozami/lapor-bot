package domain

import (
	"testing"
	"time"
)

func TestGenerateDailyQuest(t *testing.T) {
	now := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	// 1. No job means no side quest.
	tasksGeneral := GenerateDailyQuest("", 0, now)
	if len(tasksGeneral) != 0 {
		t.Fatalf("expected no tasks without job, got %d", len(tasksGeneral))
	}

	// 2. Job quest tests
	tasksFighter := GenerateDailyQuest("fighter", 0, now)
	if len(tasksFighter) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasksFighter))
	}
	if tasksFighter[0].ID != "easycardio" || tasksFighter[0].Target != 1 || tasksFighter[0].Difficulty != "easy" {
		t.Errorf("expected easy walk/bike target, got %+v", tasksFighter[0])
	}

	// 3. Medium/hard scale gently with level.
	tasksLv20 := GenerateDailyQuest("fighter", 20, now)
	if tasksLv20[1].Target <= tasksFighter[1].Target {
		t.Errorf("expected medium target to scale up, got lv0=%d lv20=%d", tasksFighter[1].Target, tasksLv20[1].Target)
	}
	if tasksLv20[2].Target <= tasksFighter[2].Target {
		t.Errorf("expected hard target to scale up, got lv0=%d lv20=%d", tasksFighter[2].Target, tasksLv20[2].Target)
	}

	// 4. User-specific hash gives users different medium/hard rotations.
	tasksUserA := GenerateDailyQuestForUser("628111", "fighter", 5, now)
	tasksUserB := GenerateDailyQuestForUser("628222", "fighter", 5, now)
	if tasksUserA[1].ID == tasksUserB[1].ID && tasksUserA[2].ID == tasksUserB[2].ID {
		t.Errorf("expected different users to get randomized medium/hard quests, got %+v and %+v", tasksUserA, tasksUserB)
	}

	// 5. Same user/job/day is deterministic.
	tasksUserAAgain := GenerateDailyQuestForUser("628111", "fighter", 5, now)
	for i := range tasksUserA {
		if tasksUserA[i].ID != tasksUserAAgain[i].ID || tasksUserA[i].Target != tasksUserAAgain[i].Target {
			t.Fatalf("expected deterministic quests, got %+v and %+v", tasksUserA, tasksUserAAgain)
		}
	}
}

func TestGenerateDailyQuest_JobAttributeSpecialists(t *testing.T) {
	now := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		job  string
		want AttributeType
	}{
		{"fighter", AttrStr},
		{"ranger", AttrSta},
		{"assassin", AttrAgi},
		{"healer", AttrVit},
		{"tank", AttrEnd},
		{"necromancer", AttrLuk},
	}

	for _, tt := range tests {
		t.Run(tt.job, func(t *testing.T) {
			tasks := GenerateDailyQuestForUser("628111", tt.job, 5, now)
			if len(tasks) != 3 {
				t.Fatalf("expected 3 tasks, got %d", len(tasks))
			}
			for _, task := range tasks[1:] {
				attrs, _ := ResolveReportAttributes(task.Name, tt.job)
				if !hasAttribute(attrs, tt.want) {
					t.Fatalf("%s task %q classified as %#v, want %s", tt.job, task.Name, attrs, tt.want)
				}
				if !MatchTask(task.Name, task.ID) {
					t.Fatalf("generated task should be matchable by its name: %+v", task)
				}
			}
		})
	}
}

func TestGenerateDailyQuest_AvoidsOverlappingMediumAndHardTasks(t *testing.T) {
	jobs := []string{"fighter", "ranger", "assassin", "mage", "healer", "tank", "necromancer"}
	for day := 1; day <= 14; day++ {
		date := time.Date(2026, 6, day, 0, 0, 0, 0, time.UTC)
		for _, job := range jobs {
			for _, userID := range []string{"628111", "628222", "628333"} {
				tasks := GenerateDailyQuestForUser(userID, job, 8, date)
				if len(tasks) != 3 {
					t.Fatalf("expected 3 tasks for %s/%s/%s, got %d", userID, job, date.Format(time.DateOnly), len(tasks))
				}
				if questTasksConflict(tasks[1], tasks[2]) {
					t.Fatalf("medium and hard tasks should not overlap for %s/%s/%s: %+v", userID, job, date.Format(time.DateOnly), tasks)
				}
			}
		}
	}
}

func TestMatchTask(t *testing.T) {
	tests := []struct {
		input  string
		taskID string
		want   bool
	}{
		// Original exercises
		{"pushup 5", "pushup", true},
		{"push up 10", "pushup", true},
		{"Push-Up 3", "pushup", true},
		{"sit up 10", "situp", true},
		{"situps 12", "situp", true},
		{"jalan 30", "jalan", true},
		{"running 5", "lari", true},
		{"cycling 20", "sepeda", true},
		{"burpee 10", "burpee", true},
		{"jalan kaki 4000", "easycardio", true},
		{"sepeda 5 km", "easycardio", true},
		{"something else 10", "pushup", false},
		// New medium exercises
		{"high knees 1 menit", "highknees", true},
		{"marching 2 menit", "highknees", true},
		{"arm circle 2 menit", "armcircles", true},
		{"putaran lengan 1 menit", "armcircles", true},
		{"calf raises 20", "calfraises", true},
		{"jinjit 20", "calfraises", true},
		{"shoulder shrug 15", "shouldershrugs", true},
		{"shrug bahu 15", "shouldershrugs", true},
		{"naik turun tangga 5 menit", "stairs", true},
		{"desk plank 30 detik", "deskplank", true},
		// New hard exercises
		{"mountain climber 30 detik", "mountainclimber", true},
		{"gerakan panjat 30", "mountainclimber", true},
		{"lunges 10", "lunges", true},
		{"lunjak 10", "lunges", true},
		{"glute bridge 15", "glutebridge", true},
		{"jembatan pinggul 15", "glutebridge", true},
		{"reverse crunch 10", "reversecrunch", true},
		{"squat jump 8", "squatjump", true},
		{"lateral shuffle 45 detik", "lateralshuffle", true},
		// Negative cases
		{"bicycle crunch 10", "reversecrunch", false}, // should not match: exclude "bicycle"
		{"chair squat", "air", false},
		{"rundown meeting", "lari", false},
		{"legacy cleanup", "weight", false},
		{"random text", "highknees", false},
	}

	for _, tt := range tests {
		got := MatchTask(tt.input, tt.taskID)
		if got != tt.want {
			t.Errorf("MatchTask(%q, %q) = %v, want %v", tt.input, tt.taskID, got, tt.want)
		}
	}
}

// TestGenerateDailyQuest_OfficePoolSize guards the variety invariant: every
// attribute must keep at least 4 options per tier so the per-user/per-day hash
// selection stays varied (removing non-office items without replacements
// would shrink the pools and make the daily quests repetitive).
func TestGenerateDailyQuest_OfficePoolSize(t *testing.T) {
	for _, job := range []string{"fighter", "ranger", "assassin", "healer", "tank", "necromancer"} {
		medium, hard := questOptionsForJob(job, 0, 0)
		if len(medium) < 4 {
			t.Errorf("%s medium pool has %d options, want >=4", job, len(medium))
		}
		if len(hard) < 4 {
			t.Errorf("%s hard pool has %d options, want >=4", job, len(hard))
		}
	}
}

// TestGenerateDailyQuest_OfficeFeasiblePool ensures exercises that need
// equipment or a track are not generated as daily office side quests.
func TestGenerateDailyQuest_OfficeFeasiblePool(t *testing.T) {
	removed := []string{"skipping", "ladderdrill", "shuttlerun", "obstaclecourse", "adaptability"}
	for _, job := range []string{"fighter", "ranger", "assassin", "healer", "tank", "necromancer", "mage"} {
		for day := 1; day <= 28; day++ {
			date := time.Date(2026, 6, day, 0, 0, 0, 0, time.UTC)
			for _, userID := range []string{"628111", "628222", "628333"} {
				tasks := GenerateDailyQuestForUser(userID, job, 10, date)
				for _, task := range tasks {
					for _, id := range removed {
						if task.ID == id {
							t.Fatalf("%s/%s/%s generated removed item %q", userID, job, date.Format(time.DateOnly), id)
						}
					}
					if !MatchTask(task.Name, task.ID) {
						t.Fatalf("generated task should be matchable by its name: %+v", task)
					}
				}
			}
		}
	}
}

// TestGenerateDailyQuest_VaryAcrossUsersAndDays expands the randomization
// check: across users and days, each attribute pool must actually produce
// several different medium tasks (the hash spreads over the whole pool).
func TestGenerateDailyQuest_VaryAcrossUsersAndDays(t *testing.T) {
	jobs := []struct {
		job     string
		primary AttributeType
	}{
		{"fighter", AttrStr},
		{"ranger", AttrSta},
		{"assassin", AttrAgi},
		{"healer", AttrVit},
		{"tank", AttrEnd},
		{"necromancer", AttrLuk},
	}
	for _, tt := range jobs {
		medium, hard := questOptionsForJob(tt.job, 0, 0)
		seenMedium := map[string]bool{}
		seenHard := map[string]bool{}
		for day := 1; day <= 14; day++ {
			date := time.Date(2026, 6, day, 0, 0, 0, 0, time.UTC)
			for _, userID := range []string{"628111", "628222", "628333", "628444", "628555"} {
				tasks := GenerateDailyQuestForUser(userID, tt.job, 10, date)
				if len(tasks) != 3 {
					t.Fatalf("%s: expected 3 tasks, got %d", tt.job, len(tasks))
				}
				seenMedium[tasks[1].ID] = true
				seenHard[tasks[2].ID] = true
			}
		}
		if len(seenMedium) < 3 || len(seenMedium) > len(medium) {
			t.Errorf("%s: medium variety = %d (%v), want 3..%d", tt.job, len(seenMedium), seenMedium, len(medium))
		}
		if len(seenHard) < 3 || len(seenHard) > len(hard) {
			t.Errorf("%s: hard variety = %d (%v), want 3..%d", tt.job, len(seenHard), seenHard, len(hard))
		}
	}
}

// TestMatchTaskNewOfficeTasks locks the keyword mapping for the new
// office-feasible side quest items.
func TestMatchTaskNewOfficeTasks(t *testing.T) {
	tests := []struct {
		input  string
		taskID string
		want   bool
	}{
		{"single leg balance 2 menit", "singlelegbalance", true},
		{"balance satu kaki 2", "singlelegbalance", true},
		{"stair walk 10 menit", "stairwalk", true},
		{"jalan di tangga 10", "stairwalk", true},
		{"interval walk 20 menit", "intervalwalk", true},
		{"jalan interval 20", "intervalwalk", true},
		{"new route walk 15", "newroutewalk", true},
		{"rute baru 15 menit", "newroutewalk", true},
		{"timed walk challenge 30", "timedwalk", true},
		// No cross-matches between the new walk items.
		{"stair walk 10", "stairs", false},
		{"stair walk 10", "intervalwalk", false},
		{"interval walk 20", "timedwalk", false},
		{"new route walk 15", "explorationwalk", false},
		{"single leg balance 2", "lateralshuffle", false},
		// Regular stairs input still matches only stairs.
		{"naik turun tangga 5", "stairs", true},
		{"naik turun tangga 5", "stairwalk", false},
	}
	for _, tt := range tests {
		if got := MatchTask(tt.input, tt.taskID); got != tt.want {
			t.Errorf("MatchTask(%q, %q) = %v, want %v", tt.input, tt.taskID, got, tt.want)
		}
	}
}

func hasAttribute(attrs []AttributeType, want AttributeType) bool {
	for _, attr := range attrs {
		if attr == want {
			return true
		}
	}
	return false
}
