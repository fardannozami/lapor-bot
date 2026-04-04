package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	SQLitePath      string
	GroupID         string
	BotPhone        string
	ReplyDelayMinMs int  // Minimum delay before reply (milliseconds)
	ReplyDelayMaxMs int  // Maximum delay before reply (milliseconds), 0 = use min as fixed
	ShowTyping      bool // Show typing indicator during delay
	StravaClientID  string
	StravaClientSecret string
	StravaVerifyToken string
	AppBaseURL      string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults/environment variables")
	}

	port := getenv("PORT", "8080")
	sqlitePath := getenv("SQLITE_PATH", "./data/whatsapp.db")
	groupID := getenv("GROUP_ID", "")
	botPhone := getenv("BOT_PHONE", "")
	replyDelayMinMs := getenvInt("REPLY_DELAY_MIN_MS", 0)
	replyDelayMaxMs := getenvInt("REPLY_DELAY_MAX_MS", 0)
	showTyping := getenvBool("SHOW_TYPING", false)
	stravaClientID := getenv("STRAVA_CLIENT_ID", "")
	stravaClientSecret := getenv("STRAVA_CLIENT_SECRET", "")
	stravaVerifyToken := getenv("STRAVA_VERIFY_TOKEN", "")
	appBaseURL := getenv("APP_BASE_URL", "http://localhost:8080")

	return Config{
		Port:            port,
		SQLitePath:      sqlitePath,
		GroupID:         groupID,
		BotPhone:        botPhone,
		ReplyDelayMinMs: replyDelayMinMs,
		ReplyDelayMaxMs: replyDelayMaxMs,
		ShowTyping:      showTyping,
		StravaClientID:  stravaClientID,
		StravaClientSecret: stravaClientSecret,
		StravaVerifyToken: stravaVerifyToken,
		AppBaseURL:      appBaseURL,
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
