package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type JobUsecase struct {
	repo domain.ReportRepository
}

func NewJobUsecase(repo domain.ReportRepository) *JobUsecase {
	return &JobUsecase{repo: repo}
}

func (uc *JobUsecase) List() string {
	sb := strings.Builder{}
	sb.WriteString("🧭 *Daftar Hunter Jobs*\n")
	sb.WriteString("Job bersifat role/profile permanen dan tidak reset saat season baru.\n")
	sb.WriteString("Pilih dengan: `#job <id>`\n")
	sb.WriteString("Contoh: `#job ranger`\n\n")

	for _, job := range domain.AllJobClasses {
		sb.WriteString(fmt.Sprintf("%s *%s* (`%s`)\n", job.Icon, job.Name, job.ID))
		sb.WriteString(fmt.Sprintf("_%s — %s_\n", job.Description, job.Trait))
	}

	return sb.String()
}

func (uc *JobUsecase) Select(ctx context.Context, userID, name, jobID string) (string, error) {
	jobID = strings.ToLower(strings.TrimSpace(jobID))
	job, ok := domain.GetJobClass(jobID)
	if !ok {
		return fmt.Sprintf("Job `%s` tidak tersedia. Cek daftar job dengan #jobs.", jobID), nil
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
		return fmt.Sprintf("🔒 *Pilihan Job Belum Terbuka!*\n\nKamu harus memiliki minimal %d poin (level up ke Fighter/Tier 2) untuk memilih job. Poinmu saat ini: %d.\nKumpulkan poin dengan melapor latihan menggunakan `#lapor`! 💪", MinPointsToSelectJob, points), nil
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
