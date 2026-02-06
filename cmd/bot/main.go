package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fardannozami/whatsapp-gateway/internal/app/usecase"
	"github.com/fardannozami/whatsapp-gateway/internal/config"
	"github.com/fardannozami/whatsapp-gateway/internal/infra/sqlite"
	"github.com/fardannozami/whatsapp-gateway/internal/infra/wa"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	walog "go.mau.fi/whatsmeow/util/log"
	_ "modernc.org/sqlite"
)

func main() {
	// 1. Load Config
	cfg := config.Load()

	// 2. Logger
	logger := walog.Stdout("Client", "INFO", true)

	// 3. Database & Repositories
	// Enable WAL mode and busy timeout to avoid "database is locked" errors
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", cfg.SQLitePath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo := sqlite.NewReportRepository(db)
	// Initialize table if needed (optional but good practice)
	if err := repo.InitTable(context.Background()); err != nil {
		log.Printf("Failed to init table: %v", err)
	}

	// 4. Use Cases
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	handleMessageUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC)

	// 5. WhatsApp Service
	waService := wa.NewService(cfg.SQLitePath, logger)

	// 6. Register Message Handler
	waService.SetMessageHandler(func(ctx context.Context, client *whatsmeow.Client, evt *events.Message) {
		// Only handle messages from groups or specific sources if needed.
		// For now, we filter by GroupID if configured.
		if cfg.GroupID != "" && evt.Info.Chat.String() != cfg.GroupID {
			return
		}

		// Ignore messages from self
		if evt.Info.IsFromMe {
			return
		}

		// Get sender info
		sender := evt.Info.Sender.String()
		pushName := evt.Info.PushName
		if pushName == "" {
			pushName = "Unknown" // Fallback name
		}

		// Get message content
		msg := ""
		if evt.Message.Conversation != nil {
			msg = *evt.Message.Conversation
		} else if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.Text != nil {
			msg = *evt.Message.ExtendedTextMessage.Text
		}

		if msg == "" {
			return
		}

		fmt.Printf("Message from %s (%s): %s\n", pushName, sender, msg)

		// Execute Use Case
		response, err := handleMessageUC.Execute(ctx, sender, pushName, msg)
		if err != nil {
			log.Printf("Error handling message: %v", err)
			return
		}

		if response != "" {
			// Send response
			resp := &waE2E.Message{
				Conversation: &response,
			}
			_, err := waService.GetClient().SendMessage(ctx, evt.Info.Chat, resp)
			if err != nil {
				log.Printf("Failed to send response: %v", err)
			}
		}
	})

	// 7. Connect
	if err := waService.Connect(context.Background()); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Check login status
	if !waService.IsLoggedIn() {
		if cfg.BotPhone != "" {
			log.Println("Not logged in. Attempting to pair with phone:", cfg.BotPhone)
			code, err := waService.Pair(cfg.BotPhone)
			if err != nil {
				log.Printf("Failed to generate pair code: %v", err)
			} else {
				log.Println("==================================================")
				log.Printf("PAIR CODE: %s", code)
				log.Println("==================================================")
				log.Println("Please verify this code on your WhatsApp (Linked Devices > Link with phone number)")
			}
		} else {
			log.Println("Not logged in. BOT_PHONE not set. Printing QR...")
			waService.PrintQR()
		}
	} else {
		log.Println("Client is already logged in.")
	}

	log.Println("Bot is running... Press Ctrl+C to exit.")

	// 8. Wait for OS Signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down...")
	waService.Disconnect()
	os.Exit(0)
}
