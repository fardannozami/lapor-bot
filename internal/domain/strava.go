package domain

import (
	"context"
	"time"
)

type StravaAccount struct {
	UserID       string    `json:"user_id" db:"user_id"`
	AthleteID    int64     `json:"athlete_id" db:"athlete_id"`
	AccessToken  string    `json:"access_token" db:"access_token"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	Name         string    `json:"name" db:"name"`
}

type StravaRepository interface {
	UpsertStravaAccount(ctx context.Context, account *StravaAccount) error
	GetStravaAccountByAthleteID(ctx context.Context, athleteID int64) (*StravaAccount, error)
	GetStravaAccountByUserID(ctx context.Context, userID string) (*StravaAccount, error)
}
