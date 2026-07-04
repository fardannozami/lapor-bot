package usecase

import (
	"context"
	"fmt"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type DisconnectStravaUsecase struct {
	repo domain.ReportRepository
}

func NewDisconnectStravaUsecase(repo domain.ReportRepository) *DisconnectStravaUsecase {
	return &DisconnectStravaUsecase{repo: repo}
}

func (uc *DisconnectStravaUsecase) Execute(ctx context.Context, userID string) (string, error) {
	// Check if user has a linked Strava account
	account, err := uc.repo.GetStravaAccountByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("gagal mengambil data Strava: %w", err)
	}

	if account == nil {
		return "❌ Kamu belum menghubungkan akun Strava. Gunakan /strava untuk menghubungkan.", nil
	}

	// Delete the Strava account
	if err := uc.repo.DeleteStravaAccount(ctx, userID); err != nil {
		return "", fmt.Errorf("gagal memutus koneksi Strava: %w", err)
	}

	return "✅ *Akun Strava berhasil diputus!*\n\nAktivitas Strava-mu tidak akan lagi otomatis dilaporkan.\nGunakan /strava jika ingin menghubungkan kembali.", nil
}
