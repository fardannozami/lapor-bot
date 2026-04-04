package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	log.Printf("Processing Strava event: %+v", event)

	// Only process new activities
	if event.ObjectType != "activity" || event.AspectType != "create" {
		log.Printf("Ignoring Strava event: not a new activity (ObjectType=%s, AspectType=%s)", event.ObjectType, event.AspectType)
		return nil
	}

	// 1. Get Strava account
	account, err := uc.repo.GetStravaAccountByAthleteID(ctx, event.AthleteID)
	if err != nil {
		return err
	}
	if account == nil {
		log.Printf("Strava athlete %d not linked to any WhatsApp user", event.AthleteID)
		return nil
	}
	log.Printf("Found linked account for user %s", account.UserID)

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
		log.Printf("Failed to get Strava activity %d: %v", event.ObjectID, err)
		return err
	}
	log.Printf("Fetched activity: %s (%s)", activity.Name, activity.Type)

	// 4. Filter activity types (Run, Ride, Swim, Walk, etc.)
	// You can customize this list
	validTypes := map[string]bool{
		"Run": true, "Ride": true, "Walk": true, "Hike": true, "Swim": true,
		"Workout": true, "WeightTraining": true, "Crossfit": true, "Yoga": true,
		"VirtualRun": true, "VirtualRide": true,
	}
	if !validTypes[activity.Type] {
		log.Printf("Ignoring activity type: %s (User: %s)", activity.Type, account.UserID)
		return nil
	}

	// 5. Trigger Report
	var workout *domain.HevyWorkout
	workoutTypes := map[string]bool{
		"workout": true, "weighttraining": true, "yoga": true, "crossfit": true,
		"hiit": true, "pilates": true, "rockclimbing": true,
	}

	activityTypeLower := strings.ToLower(activity.Type)
	isWorkout := workoutTypes[activityTypeLower]
	log.Printf("Activity Type: %s (isWorkout=%v)", activity.Type, isWorkout)

	if isWorkout {
		workout = &domain.HevyWorkout{
			Title: activity.Name,
			Time:  "", // Strava time is usually duration_seconds
		}
		if activity.Description != "" {
			lines := strings.Split(activity.Description, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" {
					workout.Exercises = append(workout.Exercises, trimmed)
				}
			}
		}
	}

	report, err := uc.repo.GetReport(ctx, account.UserID)
	if err != nil {
		return err
	}

	name := account.Name
	if name == "" {
		name = "User"
	}
	if report != nil {
		name = report.Name
	}

	response, err := uc.reportUC.Execute(ctx, account.UserID, name, workout)
	if err != nil {
		log.Printf("Failed to execute report usecase: %v", err)
		return err
	}
	log.Printf("Report triggered successfully for %s. Response: %s", name, response)

	// 6. Send notification to group
	if uc.groupID != "" && waClient != nil && waClient.IsConnected() {
		targetJID, _ := types.ParseJID(uc.groupID)
		log.Printf("Sending Strava auto-report to group JID: %s", targetJID.String())

		// Handle workout-style activities without distance
		activityLabel := "Activity"
		distanceLine := fmt.Sprintf("📏 Distance: %.2f km\n", activity.Distance/1000)

		if isWorkout || activity.Distance == 0 {
			activityLabel = "Workout List"
			if activity.Distance == 0 {
				distanceLine = ""
			}
		}

		notification := fmt.Sprintf("🚴‍♂️ *STRAVA AUTO-REPORT* 🏃‍♂️\n\n🎯 %s: %s\n%s📅 Type: %s\n\n%s",
			activityLabel, activity.Name, distanceLine, activity.Type, response)

		msg := &waE2E.Message{
			Conversation: &notification,
		}
		_, err = waClient.SendMessage(ctx, targetJID, msg)
		if err != nil {
			log.Printf("Failed to send Strava notification to group %s: %v", uc.groupID, err)
		} else {
			log.Printf("Sent Strava auto-report notification to group %s", uc.groupID)
		}
	} else {
		log.Printf("Skipping group notification: groupID=%s, waClient connected=%v", uc.groupID, waClient != nil && waClient.IsConnected())
	}

	return nil
}
