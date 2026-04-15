package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/app/usecase"
	"github.com/fardannozami/whatsapp-gateway/internal/config"
	botHTTP "github.com/fardannozami/whatsapp-gateway/internal/infra/http"
	"github.com/fardannozami/whatsapp-gateway/internal/infra/repository"
	"github.com/fardannozami/whatsapp-gateway/internal/infra/strava"
	"github.com/fardannozami/whatsapp-gateway/internal/infra/wa"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	walog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	// 1. Load Config
	cfg := config.Load()

	// 2. Logger
	logger := walog.Stdout("Client", "INFO", true)

	// 3. Database & Repositories
	repo := repository.NewReportRepository(cfg)

	// 4. Use Cases
	// 4. Use Cases
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	remindInactiveUC := usecase.NewRemindInactiveUsersUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	broadcastUpdateUC := usecase.NewBroadcastUpdateUsecase()

	// Strava Integration
	stravaClient := strava.NewClient(cfg)
	linkStravaUC := usecase.NewLinkStravaUsecase(repo, stravaClient, cfg)
	processStravaUC := usecase.NewProcessStravaWebhookUsecase(repo, stravaClient, reportUC, cfg.GroupID)

	handleMessageUC := usecase.NewHandleMessageUsecase(
		reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, updateNameUC, linkStravaUC, broadcastUpdateUC,
	)

	// 5. WhatsApp Service
	waService := wa.NewService(cfg.SQLitePath, logger)

	// 6. Register Message Handler
	waService.SetMessageHandler(func(ctx context.Context, client *whatsmeow.Client, evt *events.Message) {
		// Log all incoming messages with their Chat ID (useful for getting groupID)
		fmt.Printf("[DEBUG] Incoming message from Chat ID: %s\n", evt.Info.Chat.String())

		// Only handle messages from groups or specific sources if needed.
		// For now, we filter by GroupID if configured.
		if cfg.GroupID != "" && evt.Info.Chat.String() != cfg.GroupID {
			return
		}

		// Ignore messages from self
		if evt.Info.IsFromMe {
			return
		}

		// Get sender info - resolve LID to phone number for consistent user tracking
		senderJID := evt.Info.Sender
		var userID string
		if senderJID.Server == "lid" || senderJID.Server == types.DefaultUserServer && len(senderJID.User) > 15 {
			// Looks like a LID, try to resolve to phone number
			userID = repo.ResolveLIDToPhone(ctx, senderJID.User)
		} else {
			// Already a phone number
			userID = senderJID.User
		}

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
		} else if evt.Message.ImageMessage != nil && evt.Message.ImageMessage.Caption != nil {
			msg = *evt.Message.ImageMessage.Caption
		} else if evt.Message.VideoMessage != nil && evt.Message.VideoMessage.Caption != nil {
			msg = *evt.Message.VideoMessage.Caption
		} else if evt.Message.DocumentMessage != nil && evt.Message.DocumentMessage.Caption != nil {
			msg = *evt.Message.DocumentMessage.Caption
		}

		if msg == "" {
			return
		}

		fmt.Printf("Message from %s (%s): %s\n", pushName, userID, msg)

		// Special handling for check_inactive command
		if strings.HasPrefix(msg, "!check_inactive") {
			log.Printf("[DEBUG] Received !check_inactive command from %s", evt.Info.Chat.String())

			targetID := cfg.GroupID
			if targetID == "" {
				targetID = evt.Info.Chat.String()
				log.Printf("[DEBUG] GROUP_ID is empty, using current chat JID: %s", targetID)
			}

			response, err := remindInactiveUC.Execute(ctx, waService.GetClient(), targetID)
			if err != nil {
				log.Printf("Failed to run manual inactivity check: %v", err)
				response = fmt.Sprintf("Gagal menjalankan pengecekan: %v", err)
			}

			// Send the status response back to the user
			log.Printf("[DEBUG] Sending response for !check_inactive: %s", response)
			resp := &waE2E.Message{
				Conversation: &response,
			}
			_, err = waService.GetClient().SendMessage(ctx, evt.Info.Chat, resp)
			if err != nil {
				log.Printf("Failed to send status response: %v", err)
			}
			return
		}

		// Execute Use Case
		response, err := handleMessageUC.Execute(ctx, userID, pushName, msg)
		if err != nil {
			log.Printf("Error handling message: %v", err)
			return
		}

		if response.Text != "" {
			// Determine response target
			targetChat := evt.Info.Chat
			if response.IsPrivate {
				targetChat = evt.Info.Sender
			}

			// Apply reply delay to appear more human-like
			delayMs := cfg.ReplyDelayMinMs
			if cfg.ReplyDelayMaxMs > cfg.ReplyDelayMinMs {
				// Random delay between min and max
				delayMs = cfg.ReplyDelayMinMs + rand.Intn(cfg.ReplyDelayMaxMs-cfg.ReplyDelayMinMs+1)
			}

			if delayMs > 0 {
				// Show typing indicator if enabled
				if cfg.ShowTyping {
					_ = waService.GetClient().SendChatPresence(ctx, targetChat, types.ChatPresenceComposing, types.ChatPresenceMediaText)
				}

				log.Printf("Delaying reply by %dms", delayMs)
				time.Sleep(time.Duration(delayMs) * time.Millisecond)

				// Clear typing indicator
				if cfg.ShowTyping {
					_ = waService.GetClient().SendChatPresence(ctx, targetChat, types.ChatPresencePaused, types.ChatPresenceMediaText)
				}
			}

			// Send response
			resp := &waE2E.Message{
				Conversation: &response.Text,
			}
			targetChat.Device = 0
			_, err := waService.GetClient().SendMessage(ctx, targetChat, resp)
			if err != nil {
				log.Printf("Failed to send response: %v", err)
			}
		}
	})

	// 7. Initialize Client (DB, Device, etc) - DO NOT CONNECT YET
	if err := waService.Initialize(context.Background()); err != nil {
		log.Fatalf("Failed to initialize WhatsApp service: %v", err)
	}

	// 8. Connect / Login Logic
	if !waService.IsLoggedIn() {
		if cfg.BotPhone != "" {
			// Pair Code Mode
			// Must connect first to pair
			if err := waService.Connect(); err != nil {
				log.Fatalf("Failed to connect for pairing: %v", err)
			}

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
			// QR Code Mode
			log.Println("Not logged in. BOT_PHONE not set. Printing QR...")
			// PrintQR handles GetQRChannel AND Connect() internally to ensure no race condition
			waService.PrintQR()
		}
	} else {
		// Already logged in, just connect
		if err := waService.Connect(); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		log.Println("Client is already logged in.")
	}

	log.Println("Bot is running... Press Ctrl+C to exit.")

	// Background ticker for inactivity check (every day at 12:00 WIB)
	go func() {
		// Ensure logic runs in Asia/Jakarta
		loc, _ := time.LoadLocation("Asia/Jakarta")

		for {
			now := time.Now().In(loc)
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 15, 9, 0, 0, loc)

			if now.After(nextRun) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			delay := time.Until(nextRun)
			log.Printf("Next scheduled inactivity check at: %v (in %v)", nextRun, delay)

			select {
			case <-time.After(delay):
				if cfg.GroupID != "" && waService.IsLoggedIn() && waService.GetClient().IsConnected() {
					log.Println("Running scheduled inactivity check...")
					_, err := remindInactiveUC.Execute(context.Background(), waService.GetClient(), cfg.GroupID)
					if err != nil {
						log.Printf("Scheduled inactivity check failed: %v", err)
					}
				}
			case <-context.Background().Done():
				return
			}
		}
	}()

	// Background ticker for daily leaderboard (every day at 23:58 WIB)
	go func() {
		loc, _ := time.LoadLocation("Asia/Jakarta")
		for {
			now := time.Now().In(loc)
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 23, 58, 0, 0, loc)

			if now.After(nextRun) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			delay := time.Until(nextRun)
			log.Printf("Next scheduled leaderboard at: %v (in %v)", nextRun, delay)

			select {
			case <-time.After(delay):
				if cfg.GroupID != "" && waService.IsLoggedIn() && waService.GetClient().IsConnected() {
					log.Println("Running scheduled leaderboard...")
					response, err := leaderboardUC.Execute(context.Background())
					if err != nil {
						log.Printf("Scheduled leaderboard failed: %v", err)
						continue
					}

					targetJID, _ := types.ParseJID(cfg.GroupID)
					msg := &waE2E.Message{
						Conversation: &response,
					}
					_, err = waService.GetClient().SendMessage(context.Background(), targetJID, msg)
					if err != nil {
						log.Printf("Failed to send scheduled leaderboard: %v", err)
					}
				}
			case <-context.Background().Done():
				return
			}
		}
	}()

	log.Printf("Starting HTTP server on port %s", cfg.Port)

	// 8. HTTP server (Healthcheck + Strava)
	httpServer := botHTTP.NewServer(linkStravaUC, processStravaUC, waService.GetClient(), cfg)
	mux := http.NewServeMux()
	httpServer.RegisterHandlers(mux)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// 9. Wait for OS Signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("Received signal: %v. Shutting down...", sig)
	waService.Disconnect()

	// Shutdown healthcheck server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Healthcheck server shutdown error: %v", err)
	}

	os.Exit(0)
}
