package domain

import (
	"reflect"
	"testing"
)

func TestParseHevy(t *testing.T) {
	message := `programmer telo
Today at 9:46 AM · Hevy
Flexibility
Logged with Hevy

Warm Up
Set 1: 1min 30s

Butterfly Stretch
Set 1: 1min 44s
Set 2: 54s
Set 3: 22s

Leg Stretch
Set 1: 1min 15s
Set 2: 1min 21s
Set 3: 1min 1s

Front Wide Stretch
Set 1: 52s
Set 2: 54s
Set 3: 51s

Forward Fold
Set 1: 27s
Set 2: 31s
Set 3: 37s

Happy Baby Pose
Set 1: 27s
Set 43: 43s

Supine Twist R
Set 1: 27s
Set 2: 38s

Supine Twist L
Set 1: 37s
Set 2: 50s

Time
37m 11s`

	expected := &HevyWorkout{
		Title: "Flexibility",
		Exercises: []string{
			"Warm Up",
			"Butterfly Stretch",
			"Leg Stretch",
			"Front Wide Stretch",
			"Forward Fold",
			"Happy Baby Pose",
			"Supine Twist R",
			"Supine Twist L",
		},
		Time: "37m 11s",
	}

	got := ParseHevy(message)

	if got == nil {
		t.Fatal("Expected HevyWorkout, got nil")
	}

	if got.Title != expected.Title {
		t.Errorf("Title: expected %s, got %s", expected.Title, got.Title)
	}

	if got.Time != expected.Time {
		t.Errorf("Time: expected %s, got %s", expected.Time, got.Time)
	}

	if !reflect.DeepEqual(got.Exercises, expected.Exercises) {
		t.Errorf("Exercises: expected %v, got %v", expected.Exercises, got.Exercises)
	}
}
