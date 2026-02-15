package supabase

import (
	"context"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
	supa "github.com/nedpals/supabase-go"
)

type ReportRepository struct {
	client *supa.Client
}

type UserReport struct {
	UserID         string `json:"user_id"`
	Name           string `json:"name"`
	Streak         int    `json:"streak"`
	ActivityCount  int    `json:"activity_count"`
	LastReportDate string `json:"last_report_date"`
	MaxStreak      int    `json:"max_streak"`
	TotalPoints    int    `json:"total_points"`
	Achievements   string `json:"achievements"`
}

type LIDMap struct {
	LID string `json:"lid"`
	PN  string `json:"pn"`
}

func NewReportRepository(client *supa.Client) *ReportRepository {
	return &ReportRepository{client: client}
}

func (r *ReportRepository) GetReport(ctx context.Context, userID string) (*domain.Report, error) {
	var results []UserReport

	err := r.client.DB.From("user_reports").
		Select("*").
		Eq("user_id", userID).
		Execute(&results)

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	result := results[0]
	report := &domain.Report{
		UserID:        result.UserID,
		Name:          result.Name,
		Streak:        result.Streak,
		ActivityCount: result.ActivityCount,
		MaxStreak:     result.MaxStreak,
		TotalPoints:   result.TotalPoints,
		Achievements:  result.Achievements,
	}

	if result.LastReportDate != "" {
		report.LastReportDate = parseTime(result.LastReportDate)
	}

	return report, nil
}

func (r *ReportRepository) UpsertReport(ctx context.Context, report *domain.Report) error {
	data := UserReport{
		UserID:         report.UserID,
		Name:           report.Name,
		Streak:         report.Streak,
		ActivityCount:  report.ActivityCount,
		LastReportDate: report.LastReportDate.Format("2006-01-02T15:04:05Z07:00"),
		MaxStreak:      report.MaxStreak,
		TotalPoints:    report.TotalPoints,
		Achievements:   report.Achievements,
	}

	var results []UserReport
	err := r.client.DB.From("user_reports").
		Upsert(data).
		Execute(&results)

	return err
}

func (r *ReportRepository) GetAllReports(ctx context.Context) ([]*domain.Report, error) {
	var results []UserReport

	err := r.client.DB.From("user_reports").
		Select("*").
		Execute(&results)

	if err != nil {
		return nil, err
	}

	var reports []*domain.Report
	for _, result := range results {
		report := &domain.Report{
			UserID:        result.UserID,
			Name:          result.Name,
			Streak:        result.Streak,
			ActivityCount: result.ActivityCount,
			MaxStreak:     result.MaxStreak,
			TotalPoints:   result.TotalPoints,
			Achievements:  result.Achievements,
		}

		if result.LastReportDate != "" {
			report.LastReportDate = parseTime(result.LastReportDate)
		}

		reports = append(reports, report)
	}

	return reports, nil
}

func (r *ReportRepository) InitTable(ctx context.Context) error {
	// Table initialization is handled by the SQL schema in Supabase
	// This method is kept for compatibility but does nothing
	return nil
}

func (r *ReportRepository) ResolveLIDToPhone(ctx context.Context, lid string) string {
	var results []LIDMap

	err := r.client.DB.From("whatsmeow_lid_map").
		Select("pn").
		Eq("lid", lid).
		Execute(&results)

	if err != nil || len(results) == 0 {
		return lid
	}

	if results[0].PN != "" {
		return results[0].PN
	}

	return lid
}

// Helper function to parse time strings
func parseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}
	}
	return t
}
