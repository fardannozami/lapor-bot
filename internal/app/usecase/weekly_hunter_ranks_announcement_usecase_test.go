package usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type mockWeeklyRanksRepo struct {
	domain.ReportRepository
	reports []*domain.Report
}

func (m *mockWeeklyRanksRepo) GetAllReports(ctx context.Context) ([]*domain.Report, error) {
	return m.reports, nil
}

func TestWeeklyHunterRanksAnnouncement(t *testing.T) {
	now := time.Date(2026, 6, 15, 7, 0, 0, 0, time.UTC) // A Monday

	reports := []*domain.Report{
		{
			UserID:                "user1",
			Name:                  "Alice",
			JobClass:              "ranger",
			SeasonalPoints:        300,
			SeasonalActivityCount: 15,
		},
		{
			UserID:                "user2",
			Name:                  "Bob",
			JobClass:              "tank",
			SeasonalPoints:        120,
			SeasonalActivityCount: 6,
		},
		{
			UserID:                "user3",
			Name:                  "Charlie",
			JobClass:              "",
			SeasonalPoints:        50,
			SeasonalActivityCount: 2,
		},
		{
			UserID:                "user4",
			Name:                  "Inactive User",
			JobClass:              "",
			SeasonalPoints:        0,
			SeasonalActivityCount: 0,
		},
	}

	repo := &mockWeeklyRanksRepo{reports: reports}
	uc := NewWeeklyHunterRanksAnnouncementUsecase(repo)

	msg, err := uc.Execute(context.Background(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify header and contents
	for _, expected := range []string{
		"PENGUMUMAN MINGGUAN: RANK & JOB HUNTER",
		"1. Alice — C-Rank Hunter 🟦 (Ranger 🏹) | 300 pts | 15 hari",
		"2. Bob — D-Rank Hunter 🟩 (Tanker 🛡️) | 120 pts | 6 hari",
		"3. Charlie — E-Rank Hunter 🟫 (Belum memilih job) | 50 pts | 2 hari",
	} {
		if !strings.Contains(msg, expected) {
			t.Errorf("expected output to contain %q, but it didn't. Got:\n%s", expected, msg)
		}
	}

	// Inactive users with 0 points and 0 activity should be filtered out
	if strings.Contains(msg, "Inactive User") {
		t.Errorf("expected output not to contain 'Inactive User', but it did. Got:\n%s", msg)
	}
}
