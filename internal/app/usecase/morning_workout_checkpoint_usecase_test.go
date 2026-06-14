package usecase

import (
	"strings"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

func TestSplitReportsByToday(t *testing.T) {
	now := time.Date(2026, 6, 14, 9, 9, 0, 0, time.UTC)
	reports := []*domain.Report{
		{UserID: "1", Name: "Alpha", LastReportDate: now.Add(-time.Hour)},
		{UserID: "2", Name: "Beta", LastReportDate: now.AddDate(0, 0, -1)},
		nil,
	}

	active, pending := splitReportsByToday(reports, now)

	if len(active) != 1 || active[0].Name != "Alpha" {
		t.Fatalf("expected Alpha to be active today, got %#v", active)
	}
	if len(pending) != 1 || pending[0].Name != "Beta" {
		t.Fatalf("expected Beta to be pending today, got %#v", pending)
	}
}

func TestBuildMorningWorkoutCheckpointMessage(t *testing.T) {
	active := []*domain.Report{{UserID: "628111111111", Name: "Active One"}}
	pending := []*domain.Report{{UserID: "628222222222", Name: "Pending Two"}}

	msg := BuildMorningWorkoutCheckpointMessage(active, pending)

	for _, want := range []string{
		"09:09 Workout Checkpoint",
		"Sudah olahraga pagi ini",
		"Active One",
		"Belum lapor hari ini",
		"Pending Two",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("expected message to contain %q, got %q", want, msg)
		}
	}
	if strings.Contains(msg, "@") {
		t.Fatalf("morning checkpoint should use plain names without mentions, got %q", msg)
	}
}
