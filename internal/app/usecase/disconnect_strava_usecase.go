package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
	"github.com/fardannozami/whatsapp-gateway/internal/infra/strava"
)

type DisconnectStravaUsecase struct {
	repo         domain.ReportRepository
	stravaClient *strava.Client
}

func NewDisconnectStravaUsecase(repo domain.ReportRepository, stravaClient *strava.Client) *DisconnectStravaUsecase {
	return &DisconnectStravaUsecase{
		repo:         repo,
		stravaClient: stravaClient,
	}
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

	// Refresh token if it is expired or close to expiring (within 5 minutes)
	if time.Now().After(account.ExpiresAt.Add(-5 * time.Minute)) {
		log.Printf("[DISCONNECT] Refreshing Strava token for user %s before deauthorizing", account.UserID)
		refreshed, err := uc.stravaClient.RefreshToken(account.RefreshToken)
		if err != nil {
			log.Printf("[DISCONNECT] Failed to refresh token: %v. Proceeding to force delete locally.", err)
		} else {
			refreshed.UserID = account.UserID
			refreshed.AthleteID = account.AthleteID
			if err := uc.repo.UpsertStravaAccount(ctx, refreshed); err != nil {
				log.Printf("[DISCONNECT] Failed to save refreshed token: %v. Proceeding.", err)
			}
			account = refreshed
		}
	}

	// Deauthorize from Strava
	if err := uc.stravaClient.Deauthorize(account.AccessToken); err != nil {
		log.Printf("[DISCONNECT] Failed to deauthorize token from Strava server: %v. Proceeding to delete locally.", err)
	}

	// Delete the Strava account locally
	if err := uc.repo.DeleteStravaAccount(ctx, userID); err != nil {
		return "", fmt.Errorf("gagal memutus koneksi Strava di database: %w", err)
	}

	return "✅ *Akun Strava berhasil diputus!*\n\nAktivitas Strava-mu tidak akan lagi otomatis dilaporkan.\nGunakan /strava jika ingin menghubungkan kembali.", nil
}
