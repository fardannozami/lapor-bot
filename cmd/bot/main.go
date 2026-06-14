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
	"github.com/fardannozami/whatsapp-gateway/internal/queue"
	"github.com/fardannozami/whatsapp-gateway/internal/scheduler"

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
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	remindInactiveUC := usecase.NewRemindInactiveUsersUsecase(repo)
	morningCheckpointUC := usecase.NewMorningWorkoutCheckpointUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	cancelUC := usecase.NewCancelReportUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	broadcastUpdateUC := usecase.NewBroadcastUpdateUsecase()
	resetSessionUC := usecase.NewResetSessionUsecase(repo)
	motivationUC := usecase.NewGetMotivationUsecase()
	helpUC := usecase.NewGetHelpUsecase()
	goalUC := usecase.NewGoalUsecase(repo)

	// Strava Integration
	stravaClient := strava.NewClient(cfg)
	linkStravaUC := usecase.NewLinkStravaUsecase(repo, stravaClient, cfg)
	processStravaUC := usecase.NewProcessStravaWebhookUsecase(repo, stravaClient, reportUC, cfg.GroupID)

	handleMessageUC := usecase.NewHandleMessageUsecase(
		reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, cancelUC, updateNameUC, linkStravaUC, broadcastUpdateUC, motivationUC, helpUC,
	)

	// 5. WhatsApp Service
	waService := wa.NewService(cfg.SQLitePath, logger)

	// Shared context for graceful shutdown of scheduler and queue
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// Declare sender early so the message handler closure can capture it.
	// It will be initialized after the WhatsApp client connects.
	var sender *queue.MessageSender

	// 6. Register Message Handler
	waService.SetMessageHandler(func(ctx context.Context, client *whatsmeow.Client, evt *events.Message) {
		fmt.Printf("[DEBUG] Incoming message from Chat ID: %s\n", evt.Info.Chat.String())

		if cfg.GroupID != "" && evt.Info.Chat.String() != cfg.GroupID {
			return
		}

		if evt.Info.IsFromMe {
			return
		}

		senderJID := evt.Info.Sender
		var userID string
		if senderJID.Server == "lid" || senderJID.Server == types.DefaultUserServer && len(senderJID.User) > 15 {
			userID = repo.ResolveLIDToPhone(ctx, senderJID.User)
		} else {
			userID = senderJID.User
		}

		pushName := senderDisplayName(ctx, client, senderJID, evt.Info.SenderAlt, evt.Info.PushName)

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

			log.Printf("[DEBUG] Sending response for !check_inactive: %s", response)
			resp := &waE2E.Message{
				Conversation: &response,
			}
			if sender != nil {
				_ = sender.SendNormalPriority(ctx, evt.Info.Chat, resp)
			} else {
				_, _ = waService.GetClient().SendMessage(ctx, evt.Info.Chat, resp)
			}
			return
		}

		response, err := handleMessageUC.Execute(ctx, userID, pushName, msg)
		if err != nil {
			log.Printf("Error handling message: %v", err)
			return
		}

		if response.Text != "" {
			targetChat := evt.Info.Chat
			if response.IsPrivate {
				targetChat = evt.Info.Sender
			}

			delayMs := cfg.ReplyDelayMinMs
			if cfg.ReplyDelayMaxMs > cfg.ReplyDelayMinMs {
				delayMs = cfg.ReplyDelayMinMs + rand.Intn(cfg.ReplyDelayMaxMs-cfg.ReplyDelayMinMs+1)
			}

			if delayMs > 0 {
				if cfg.ShowTyping {
					_ = waService.GetClient().SendChatPresence(ctx, targetChat, types.ChatPresenceComposing, types.ChatPresenceMediaText)
				}

				log.Printf("Delaying reply by %dms", delayMs)
				time.Sleep(time.Duration(delayMs) * time.Millisecond)

				if cfg.ShowTyping {
					_ = waService.GetClient().SendChatPresence(ctx, targetChat, types.ChatPresencePaused, types.ChatPresenceMediaText)
				}
			}

			resp := &waE2E.Message{
				Conversation: &response.Text,
			}
			targetChat.Device = 0

			if sender != nil {
				_ = sender.SendNormalPriority(ctx, targetChat, resp)
			} else {
				_, _ = waService.GetClient().SendMessage(ctx, targetChat, resp)
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
			log.Println("Not logged in. BOT_PHONE not set. Printing QR...")
			waService.PrintQR()
		}
	} else {
		if err := waService.Connect(); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		log.Println("Client is already logged in.")
	}

	log.Println("Bot is running... Press Ctrl+C to exit.")

	// 9. Start message sender (serializes all SendMessage calls)
	sender = queue.NewMessageSender(waService.GetClient(), appCtx)
	sender.Start()

	// 10. Schedule seasonal reset every 4 months at 00:00 WIB.
	resetCtx, resetCancel := context.WithCancel(appCtx)
	defer resetCancel()
	usecase.ScheduleSessionReset(resetCtx, resetSessionUC, func() *whatsmeow.Client {
		return waService.GetClient()
	}, func() bool {
		return waService.IsLoggedIn() && waService.GetClient().IsConnected()
	}, cfg.GroupID)

	// 11. Daily scheduled jobs via the scheduler module
	jakartaLoc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatalf("Failed to load Asia/Jakarta timezone: %v", err)
	}

	morningSchedule, err := scheduler.ParseDaily(cfg.NotifyMorningTime, jakartaLoc)
	if err != nil {
		log.Fatalf("Invalid NOTIFY_MORNING_TIME %q: %v", cfg.NotifyMorningTime, err)
	}

	inactiveSchedule, err := scheduler.ParseDaily(cfg.NotifyInactiveTime, jakartaLoc)
	if err != nil {
		log.Fatalf("Invalid NOTIFY_INACTIVE_TIME %q: %v", cfg.NotifyInactiveTime, err)
	}

	leaderboardSchedule, err := scheduler.ParseDaily(cfg.NotifyLeaderboardTime, jakartaLoc)
	if err != nil {
		log.Fatalf("Invalid NOTIFY_LEADERBOARD_TIME %q: %v", cfg.NotifyLeaderboardTime, err)
	}

	goalCleanupSchedule, err := scheduler.ParseDaily("00:10", jakartaLoc)
	if err != nil {
		log.Fatalf("Invalid goal cleanup schedule: %v", err)
	}

	sched := scheduler.NewScheduler(appCtx)

	sched.AddJob(&scheduler.Job{
		Name:    "goal-cleanup",
		Freq:    goalCleanupSchedule,
		Recover: false,
		Fn: func(ctx context.Context) error {
			deleted, err := goalUC.CleanupExpired(ctx, time.Now().In(jakartaLoc))
			if err != nil {
				log.Printf("[SCHEDULER] Goal cleanup failed: %v", err)
				return err
			}
			if deleted > 0 {
				log.Printf("[SCHEDULER] Goal cleanup deleted %d expired goal(s)", deleted)
			}
			return nil
		},
	})

	sched.AddJob(&scheduler.Job{
		Name:    "morning-workout-checkpoint",
		Freq:    morningSchedule,
		Recover: false,
		Fn: func(ctx context.Context) error {
			if cfg.GroupID == "" || !waService.IsLoggedIn() || !waService.GetClient().IsConnected() {
				return fmt.Errorf("not connected or no group configured")
			}

			log.Println("[SCHEDULER] Running morning workout checkpoint...")
			response, err := morningCheckpointUC.Execute(ctx, time.Now().In(jakartaLoc))
			if err != nil {
				log.Printf("[SCHEDULER] Morning workout checkpoint failed: %v", err)
				return err
			}

			targetJID, err := types.ParseJID(cfg.GroupID)
			if err != nil {
				return fmt.Errorf("invalid GroupID: %w", err)
			}
			msg := &waE2E.Message{
				Conversation: &response,
			}
			return sender.SendHighPriority(ctx, targetJID, msg)
		},
	})

	sched.AddJob(&scheduler.Job{
		Name:    "inactivity-check",
		Freq:    inactiveSchedule,
		Recover: false,
		Fn: func(ctx context.Context) error {
			if cfg.GroupID == "" || !waService.IsLoggedIn() || !waService.GetClient().IsConnected() {
				return fmt.Errorf("not connected or no group configured")
			}
			log.Println("[SCHEDULER] Running inactivity check...")
			_, err := remindInactiveUC.ExecuteAt(ctx, waService.GetClient(), cfg.GroupID, time.Now().In(jakartaLoc))
			if err != nil {
				log.Printf("[SCHEDULER] Inactivity check failed: %v", err)
			}
			return err
		},
	})

	sched.AddJob(&scheduler.Job{
		Name:    "leaderboard",
		Freq:    leaderboardSchedule,
		Recover: false,
		Fn: func(ctx context.Context) error {
			if cfg.GroupID == "" || !waService.IsLoggedIn() || !waService.GetClient().IsConnected() {
				return fmt.Errorf("not connected or no group configured")
			}
			log.Println("[SCHEDULER] Running daily leaderboard...")
			response, err := leaderboardUC.Execute(ctx)
			if err != nil {
				log.Printf("[SCHEDULER] Leaderboard failed: %v", err)
				return err
			}

			response += usecase.BuildWellnessReminder()

			targetJID, err := types.ParseJID(cfg.GroupID)
			if err != nil {
				return fmt.Errorf("invalid GroupID: %w", err)
			}
			msg := &waE2E.Message{
				Conversation: &response,
			}
			return sender.SendHighPriority(ctx, targetJID, msg)
		},
	})

	sched.Start()

	log.Printf("Starting HTTP server on port %s", cfg.Port)

	// 12. HTTP server (Healthcheck + Strava)
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

	// 13. Wait for OS signal then graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	sig := <-sigCh

	log.Printf("Received signal: %v. Shutting down...", sig)

	// 1. Cancel all scheduler goroutines
	appCancel()

	// 2. Drain pending notifications with timeout
	sender.Shutdown(5 * time.Second)

	// 3. Disconnect WhatsApp
	waService.Disconnect()

	// 4. Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	os.Exit(0)
}

func senderDisplayName(ctx context.Context, client *whatsmeow.Client, senderJID, senderAlt types.JID, pushName string) string {
	if name := validDisplayName(pushName); name != "" {
		return name
	}

	for _, jid := range []types.JID{senderJID, senderAlt} {
		if jid.IsEmpty() || client == nil || client.Store == nil || client.Store.Contacts == nil {
			continue
		}

		contact, err := client.Store.Contacts.GetContact(ctx, jid.ToNonAD())
		if err != nil || !contact.Found {
			continue
		}
		if name := validDisplayName(contact.FullName); name != "" {
			return name
		}
		if name := validDisplayName(contact.PushName); name != "" {
			return name
		}
		if name := validDisplayName(contact.BusinessName); name != "" {
			return name
		}
	}

	return "Teman"
}

func validDisplayName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "-" || strings.EqualFold(name, "unknown") || strings.EqualFold(name, "user") {
		return ""
	}
	return name
}
