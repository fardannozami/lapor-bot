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
		{"tank", AttrVit},
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

func hasAttribute(attrs []AttributeType, want AttributeType) bool {
	for _, attr := range attrs {
		if attr == want {
			return true
		}
	}
	return false
}
