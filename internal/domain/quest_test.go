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
	if len(tasksFighter) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(tasksFighter))
	}
	if tasksFighter[0].ID != "jalan" || tasksFighter[0].Target != 4000 || tasksFighter[0].Difficulty != "easy" {
		t.Errorf("expected easy walk target, got %+v", tasksFighter[0])
	}
	if tasksFighter[1].ID != "sepeda" || tasksFighter[1].Target != 50 || tasksFighter[1].Difficulty != "easy" {
		t.Errorf("expected easy bike target, got %+v", tasksFighter[1])
	}

	// 3. Medium/hard scale gently with level.
	tasksLv20 := GenerateDailyQuest("fighter", 20, now)
	if tasksLv20[2].Target <= tasksFighter[2].Target {
		t.Errorf("expected medium target to scale up, got lv0=%d lv20=%d", tasksFighter[2].Target, tasksLv20[2].Target)
	}
	if tasksLv20[3].Target <= tasksFighter[3].Target {
		t.Errorf("expected hard target to scale up, got lv0=%d lv20=%d", tasksFighter[3].Target, tasksLv20[3].Target)
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
		// New hard exercises
		{"mountain climber 30 detik", "mountainclimber", true},
		{"gerakan panjat 30", "mountainclimber", true},
		{"lunges 10", "lunges", true},
		{"lunjak 10", "lunges", true},
		{"glute bridge 15", "glutebridge", true},
		{"jembatan pinggul 15", "glutebridge", true},
		{"reverse crunch 10", "reversecrunch", true},
		// Negative cases
		{"bicycle crunch 10", "reversecrunch", false}, // should not match: exclude "bicycle"
		{"random text", "highknees", false},
	}

	for _, tt := range tests {
		got := MatchTask(tt.input, tt.taskID)
		if got != tt.want {
			t.Errorf("MatchTask(%q, %q) = %v, want %v", tt.input, tt.taskID, got, tt.want)
		}
	}
}
