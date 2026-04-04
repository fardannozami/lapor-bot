package strava

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/config"
	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type Client struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		clientID:     cfg.StravaClientID,
		clientSecret: cfg.StravaClientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	Athlete      struct {
		ID int64 `json:"id"`
	} `json:"athlete"`
}

func (c *Client) ExchangeToken(code string) (*domain.StravaAccount, error) {
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")

	resp, err := c.httpClient.PostForm("https://www.strava.com/oauth/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("strava oauth error: %s", string(body))
	}

	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}

	return &domain.StravaAccount{
		AthleteID:    tr.Athlete.ID,
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		ExpiresAt:    time.Unix(tr.ExpiresAt, 0),
	}, nil
}

func (c *Client) RefreshToken(refreshToken string) (*domain.StravaAccount, error) {
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	resp, err := c.httpClient.PostForm("https://www.strava.com/oauth/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("strava refresh error: %s", string(body))
	}

	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}

	return &domain.StravaAccount{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		ExpiresAt:    time.Unix(tr.ExpiresAt, 0),
	}, nil
}

type Activity struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Distance  float64   `json:"distance"` // meters
	Type      string    `json:"type"`
	StartDate time.Time `json:"start_date"`
}

func (c *Client) GetActivity(accessToken string, activityID int64) (*Activity, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.strava.com/api/v3/activities/%d", activityID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("strava get activity error: %s", string(body))
	}

	var activity Activity
	if err := json.NewDecoder(resp.Body).Decode(&activity); err != nil {
		return nil, err
	}

	return &activity, nil
}
