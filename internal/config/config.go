package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	SQLitePath string
	GroupID    string
	BotPhone   string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults/environment variables")
	}

	sqlitePath := getenv("SQLITE_PATH", "./data/whatsapp.db")
	groupID := getenv("GROUP_ID", "")
	botPhone := getenv("BOT_PHONE", "")

	return Config{
		SQLitePath: sqlitePath,
		GroupID:    groupID,
		BotPhone:   botPhone,
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
