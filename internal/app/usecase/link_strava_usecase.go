package usecase

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/fardannozami/whatsapp-gateway/internal/config"
	"github.com/fardannozami/whatsapp-gateway/internal/domain"
	"github.com/fardannozami/whatsapp-gateway/internal/infra/strava"
)

type LinkStravaUsecase struct {
	repo         domain.ReportRepository
	stravaClient *strava.Client
	clientID     string
	redirectURI  string
}

func NewLinkStravaUsecase(repo domain.ReportRepository, stravaClient *strava.Client, cfg config.Config) *LinkStravaUsecase {
	return &LinkStravaUsecase{
		repo:         repo,
		stravaClient: stravaClient,
		clientID:     cfg.StravaClientID,
		redirectURI:  fmt.Sprintf("%s/strava/callback", cfg.AppBaseURL),
	}
}

func (uc *LinkStravaUsecase) GetAuthURL(userID, userName string) string {
	baseURL := "https://www.strava.com/oauth/authorize"
	params := url.Values{}
	params.Set("client_id", uc.clientID)
	params.Set("redirect_uri", uc.redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "activity:read_all")
	params.Set("state", fmt.Sprintf("%s|%s", userID, userName))

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

func (uc *LinkStravaUsecase) HandleCallback(ctx context.Context, code, state string) error {
	account, err := uc.stravaClient.ExchangeToken(code)
	if err != nil {
		return err
	}

	parts := strings.Split(state, "|")
	userID := parts[0]
	userName := ""
	if len(parts) > 1 {
		userName = parts[1]
	}

	account.UserID = userID
	account.Name = userName
	return uc.repo.UpsertStravaAccount(ctx, account)
}
