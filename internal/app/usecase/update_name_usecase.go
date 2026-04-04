package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type UpdateNameUsecase struct {
	repo domain.ReportRepository
}

func NewUpdateNameUsecase(repo domain.ReportRepository) *UpdateNameUsecase {
	return &UpdateNameUsecase{repo: repo}
}

func (uc *UpdateNameUsecase) Execute(ctx context.Context, userID, newName string) (string, error) {
	if newName == "" {
		return "Nama tidak boleh kosong. Gunakan format: #setname <nama_kamu>", nil
	}

	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	if report != nil {
		oldName := report.Name
		report.Name = newName
		if err := uc.repo.UpsertReport(ctx, report); err != nil {
			return "", err
		}
		return fmt.Sprintf("Berhasil mengubah nama dari %s menjadi %s! ✅", oldName, newName), nil
	}

	// Create skeleton report for new user
	report = &domain.Report{
		UserID:         userID,
		Name:           newName,
		LastReportDate: time.Now().AddDate(-1, 0, 0), // Set to long time ago
		Streak:         0,
		ActivityCount:  0,
		MaxStreak:      0,
		TotalPoints:    0,
		Achievements:   "",
		ComebackStreak: 0,
		InactiveDays:   0,
	}

	if err := uc.repo.UpsertReport(ctx, report); err != nil {
		return "", err
	}

	return fmt.Sprintf("Selamat datang! Namamu telah diatur sebagai %s. Silakan mulai laporan dengan #lapor! ✅", newName), nil
}
