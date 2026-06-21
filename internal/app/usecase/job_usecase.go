package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type JobUsecase struct {
	repo domain.ReportRepository
}

func NewJobUsecase(repo domain.ReportRepository) *JobUsecase {
	return &JobUsecase{repo: repo}
}

func (uc *JobUsecase) List(ctx context.Context) ([]domain.JobClass, error) {
	return uc.repo.GetAllJobClasses(ctx)
}

func (uc *JobUsecase) Select(ctx context.Context, userID, name, jobID string) (string, error) {
	job, err := uc.repo.GetJobClass(ctx, jobID)
	if err != nil {
		return "", err
	}
	if job == nil {
		return "", fmt.Errorf("Job '%s' tidak tersedia. Cek daftar job dengan #jobs.", jobID)
	}

	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	const MinPointsToSelectJob = 50
	points := 0
	if report != nil {
		points = report.TotalPoints
	}
	if points < MinPointsToSelectJob {
		return "", fmt.Errorf("🔒 Pilihan Job Belum Terbuka!\n\nKamu harus memiliki minimal %d poin (level up ke Fighter/Tier 2) untuk memilih job. Poinmu saat ini: %d.\nKumpulkan poin dengan melapor latihan menggunakan #lapor! 💪", MinPointsToSelectJob, points)
	}

	if report == nil {
		report = &domain.Report{
			UserID:         userID,
			Name:           name,
			LastReportDate: time.Now().AddDate(-1, 0, 0),
			Achievements:   "",
		}
	}

	report.JobClass = job.ID
	if err := uc.repo.UpsertReport(ctx, report); err != nil {
		return "", err
	}

	// Clear today's daily quest cache so it gets regenerated for the new job
	todayStr := domain.GetToday(time.Now()).Format("2006-01-02")
	_ = uc.repo.SaveDailyQuest(ctx, userID, todayStr, "")

	return fmt.Sprintf("✅ Job dipilih: %s *%s*\n_%s_\n\nJob ini akan tampil di #mystats dan laporan #lapor berikutnya.", job.Icon, job.Name, job.Description), nil
}
