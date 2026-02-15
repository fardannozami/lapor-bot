package domain

import (
	"context"
	"time"
)

type Report struct {
	UserID         string    `json:"user_id" db:"user_id"`
	Name           string    `json:"name" db:"name"`
	Streak         int       `json:"streak" db:"streak"`
	ActivityCount  int       `json:"activity_count" db:"activity_count"`
	LastReportDate time.Time `json:"last_report_date" db:"last_report_date"`
	MaxStreak      int       `json:"max_streak" db:"max_streak"`
	TotalPoints    int       `json:"total_points" db:"total_points"`
	Achievements   string    `json:"achievements" db:"achievements"`
}

type ReportRepository interface {
	GetReport(ctx context.Context, userID string) (*Report, error)
	UpsertReport(ctx context.Context, report *Report) error
	GetAllReports(ctx context.Context) ([]*Report, error)
	InitTable(ctx context.Context) error
	ResolveLIDToPhone(ctx context.Context, lid string) string
}
