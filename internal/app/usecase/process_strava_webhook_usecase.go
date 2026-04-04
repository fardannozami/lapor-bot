package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
	"github.com/fardannozami/whatsapp-gateway/internal/infra/strava"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

type ProcessStravaWebhookUsecase struct {
	repo         domain.ReportRepository
	stravaClient *strava.Client
	reportUC     *ReportActivityUsecase
	groupID      string
}

func NewProcessStravaWebhookUsecase(
	repo domain.ReportRepository,
	stravaClient *strava.Client,
	reportUC *ReportActivityUsecase,
	groupID string,
) *ProcessStravaWebhookUsecase {
	return &ProcessStravaWebhookUsecase{
		repo:         repo,
		stravaClient: stravaClient,
		reportUC:     reportUC,
		groupID:      groupID,
	}
}

type StravaWebhookEvent struct {
	ObjectType     string            `json:"object_type"`
	ObjectID       int64             `json:"object_id"`
	AspectType     string            `json:"aspect_type"`
	AthleteID      int64             `json:"owner_id"`
	SubscriptionID int               `json:"subscription_id"`
	EventTime      int64             `json:"event_time"`
	Updates        map[string]string `json:"updates"`
}

func (uc *ProcessStravaWebhookUsecase) Execute(ctx context.Context, waClient *whatsmeow.Client, event StravaWebhookEvent) error {
	// Only process new activities
	if event.ObjectType != "activity" || event.AspectType != "create" {
		return nil
	}

	// 1. Get Strava account
	account, err := uc.repo.GetStravaAccountByAthleteID(ctx, event.AthleteID)
	if err != nil {
		return err
	}
	if account == nil {
		log.Printf("Strava athlete %d not linked to any user", event.AthleteID)
		return nil
	}

	// 2. Check if token needs refresh
	if time.Now().After(account.ExpiresAt.Add(-5 * time.Minute)) {
		log.Printf("Refreshing Strava token for user %s", account.UserID)
		refreshed, err := uc.stravaClient.RefreshToken(account.RefreshToken)
		if err != nil {
			return err
		}
		refreshed.UserID = account.UserID
		refreshed.AthleteID = account.AthleteID
		if err := uc.repo.UpsertStravaAccount(ctx, refreshed); err != nil {
			return err
		}
		account = refreshed
	}

	// 3. Get activity details
	activity, err := uc.stravaClient.GetActivity(account.AccessToken, event.ObjectID)
	if err != nil {
		return err
	}

	// 4. Filter activity types (Run, Ride, Swim, Walk, etc.)
	// You can customize this list
	validTypes := map[string]bool{
		"Run": true, "Ride": true, "Walk": true, "Hike": true, "Swim": true,
	}
	if !validTypes[activity.Type] {
		log.Printf("Ignoring activity type: %s", activity.Type)
		return nil
	}

	// 5. Trigger Report
	report, err := uc.repo.GetReport(ctx, account.UserID)
	if err != nil {
		return err
	}

	name := "User"
	if report != nil {
		name = report.Name
	}

	response, err := uc.reportUC.Execute(ctx, account.UserID, name)
	if err != nil {
		return err
	}

	// 6. Send notification to group
	if uc.groupID != "" && waClient != nil && waClient.IsConnected() {
		targetJID, _ := types.ParseJID(uc.groupID)
		
		notification := fmt.Sprintf("🚴‍♂️ *STRAVA AUTO-REPORT* 🏃‍♂️\n\n🎯 Activity: %s\n📏 Distance: %.2f km\n📅 Type: %s\n\n%s", 
			activity.Name, activity.Distance/1000, activity.Type, response)

		msg := &waE2E.Message{
			Conversation: &notification,
		}
		_, err = waClient.SendMessage(ctx, targetJID, msg)
		if err != nil {
			log.Printf("Failed to send Strava notification: %v", err)
		}
	}

	return nil
}
