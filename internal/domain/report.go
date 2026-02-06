package domain

import (
	"context"
	"time"
)

type Report struct {
	UserID         string    `json:"user_id" db:"user_id"`
	Name           string    `json:"name" db:"name"`
	Streak         int       `json:"streak" db:"streak"`
	LastReportDate time.Time `json:"last_report_date" db:"last_report_date"`
}

type ReportRepository interface {
	GetReport(ctx context.Context, userID string) (*Report, error)
	UpsertReport(ctx context.Context, report *Report) error
	GetAllReports(ctx context.Context) ([]*Report, error)
}
