package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/fardannozami/whatsapp-gateway/internal/config"
	"github.com/fardannozami/whatsapp-gateway/internal/domain"
	"github.com/fardannozami/whatsapp-gateway/internal/infra/sqlite"
	_ "modernc.org/sqlite"
)

func NewReportRepository(cfg config.Config) domain.ReportRepository {
	log.Println("Using SQLite database")
	// Enable WAL mode and busy timeout to avoid "database is locked" errors
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", cfg.SQLitePath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	repo := sqlite.NewReportRepository(db)
	// Initialize table if needed
	if err := repo.InitTable(context.Background()); err != nil {
		log.Printf("Failed to init table: %v", err)
	}

	return repo
}
