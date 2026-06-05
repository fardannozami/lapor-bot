package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// SessionResetMonths defines the months when a new session starts (every 4 months).
// Reset happens at 00:00 WIB on the 1st of each of these months.
var SessionResetMonths = []time.Month{time.January, time.May, time.September}

type ResetSessionUsecase struct {
	repo domain.ReportRepository
}

func NewResetSessionUsecase(repo domain.ReportRepository) *ResetSessionUsecase {
	return &ResetSessionUsecase{repo: repo}
}

// GetCurrentSessionInfo returns the current season number and its start date.
// Seasons cycle every 4 months. This bot's public season counter starts from
// the current launch season so members see Season 1 first, then Season 2 after
// the next reset.
// Season 1: May 1, 2026 - Aug 31, 2026
// Season 2: Sep 1, 2026 - Dec 31, 2026
// Season 3: Jan 1, 2027 - Apr 30, 2027
// ...and so on.
func GetCurrentSessionInfo(now time.Time) (sessionNumber int, sessionStart time.Time) {
	loc := time.FixedZone("WIB", 7*3600)
	now = now.In(loc)

	// Base: Season 1 starts with this bot's launch season.
	baseYear := 2026
	baseMonth := time.May
	baseStart := time.Date(baseYear, baseMonth, 1, 0, 0, 0, 0, loc)
	baseSession := 1

	// Find which session period we're in
	year := now.Year()

	// Determine which session start month we're currently in or past
	var currentStart time.Time
	for i := len(SessionResetMonths) - 1; i >= 0; i-- {
		m := SessionResetMonths[i]
		candidate := time.Date(year, m, 1, 0, 0, 0, 0, loc)
		if now.Equal(candidate) || now.After(candidate) {
			currentStart = candidate
			break
		}
	}

	// If none found in this year (before January 1), go to previous year's September
	if currentStart.IsZero() {
		currentStart = time.Date(year-1, time.September, 1, 0, 0, 0, 0, loc)
	}
	if currentStart.Before(baseStart) {
		currentStart = baseStart
	}

	// Calculate session number from base
	// Each year has 3 sessions. Calculate total sessions elapsed.
	yearDiff := currentStart.Year() - baseYear
	monthIndex := 0
	baseMonthIndex := 0
	for i, m := range SessionResetMonths {
		if currentStart.Month() == m {
			monthIndex = i
		}
		if baseMonth == m {
			baseMonthIndex = i
		}
	}
	sessionNumber = baseSession + (yearDiff * len(SessionResetMonths)) + (monthIndex - baseMonthIndex)

	return sessionNumber, currentStart
}

// GetNextResetTime returns the next session reset time after 'now'.
func GetNextResetTime(now time.Time) time.Time {
	loc := time.FixedZone("WIB", 7*3600)
	now = now.In(loc)

	year := now.Year()

	// Check each reset month in the current year
	for _, m := range SessionResetMonths {
		candidate := time.Date(year, m, 1, 0, 0, 0, 0, loc)
		if candidate.After(now) {
			return candidate
		}
	}

	// If all reset months this year have passed, next is January of next year
	return time.Date(year+1, SessionResetMonths[0], 1, 0, 0, 0, 0, loc)
}

// Execute resets seasonal report data and sends an announcement to the group.
func (uc *ResetSessionUsecase) Execute(ctx context.Context, client *whatsmeow.Client, groupID string, sessionNumber int) error {
	log.Printf("[SESSION RESET] Starting Season %d reset — clearing seasonal data...", sessionNumber)

	// Reset all reports in the database
	if err := uc.repo.ResetAllReports(ctx); err != nil {
		return fmt.Errorf("failed to reset all reports: %w", err)
	}

	log.Printf("[SESSION RESET] Seasonal data has been reset for Season %d!", sessionNumber)

	// Send announcement to the group
	if groupID != "" && client != nil && client.IsConnected() {
		announcement := buildSeasonResetAnnouncement(sessionNumber)

		targetJID, _ := types.ParseJID(groupID)
		msg := &waE2E.Message{
			Conversation: &announcement,
		}
		_, err := client.SendMessage(ctx, targetJID, msg)
		if err != nil {
			log.Printf("[SESSION RESET] Failed to send announcement: %v", err)
			return fmt.Errorf("reset succeeded but failed to send announcement: %w", err)
		}
		log.Printf("[SESSION RESET] Season %d announcement sent to group!", sessionNumber)
	}

	return nil
}

func buildSeasonResetAnnouncement(sessionNumber int) string {
	seasonTransition := "Season perdana telah resmi dimulai. Semua hunter mulai berburu dari titik yang sama! 🎉"
	if sessionNumber > 1 {
		seasonTransition = fmt.Sprintf("Season %d telah resmi berakhir! 🎉", sessionNumber-1)
	}

	return fmt.Sprintf(`🔄 *SEASON %d TELAH DIMULAI!* 🔄

Halo para pejuang keringat! 🏋️‍♂️

%s

✅ *Yang di-reset:*
• 📊 Seasonal Points — mulai dari 0
• 📅 Seasonal Activity — mulai dari 0
• ⚔️ Seasonal Max Streak — mulai dari 0
• 🏅 Season Badges — mulai berburu dari awal
• 🏆 Rank & Seasonal Leaderboard — reset

💾 *Yang tetap tersimpan:*
• 🔥 Streak mingguan — lanjutkan!
• ⭐ Total Points, EXP & Level — lifetime progress aman
• 🏅 Achievement archive — yang sudah pernah unlock tetap tersimpan
• 🛡️ Centurion Cycles — tetap berlaku

❄️ *Streak Freeze* — reset ke 1 tiap season. Dapat +1 lagi saat capai 4 minggu streak!

🆕 *Season %d dimulai SEKARANG!*
Semua peserta mulai dari titik yang sama untuk seasonal ranking. Tapi level dan EXP lifetime kamu tetap ada! 💪

📌 Langsung laporkan aktivitas pertamamu di Season %d dengan mengirim #lapor!

*Semangat Season %d!* 🚀🔥`, sessionNumber, seasonTransition, sessionNumber, sessionNumber, sessionNumber)
}

// ScheduleSessionReset starts a background goroutine that automatically resets
// season data every 4 months (Jan 1, May 1, Sep 1) at 00:00 WIB.
// It loops forever, scheduling the next reset after each one completes.
func ScheduleSessionReset(ctx context.Context, uc *ResetSessionUsecase, client func() *whatsmeow.Client, isConnected func() bool, groupID string) {
	go func() {
		for {
			now := time.Now()
			sessionNum, sessionStart := GetCurrentSessionInfo(now)

			// Startup check: did we miss a reset?
			// If there's data but it's all from a previous session, trigger a reset now.
			reports, err := uc.repo.GetAllReports(context.Background())
			if err == nil && len(reports) > 0 {
				isAnyFromCurrentSession := false
				for _, r := range reports {
					if !r.LastReportDate.Before(sessionStart) {
						isAnyFromCurrentSession = true
						break
					}
				}
				if !isAnyFromCurrentSession {
					log.Printf("[SESSION RESET] Missed reset detected for Season %d. Executing now...", sessionNum)
					_ = uc.Execute(context.Background(), client(), groupID, sessionNum)
				}
			}

			nextReset := GetNextResetTime(now)
			nextSession, _ := GetCurrentSessionInfo(nextReset)

			delay := time.Until(nextReset)
			log.Printf("[SESSION RESET] Next season reset (Season %d) scheduled at: %v (in %v)", nextSession, nextReset, delay)

			select {
			case <-time.After(delay):
				log.Printf("[SESSION RESET] Reset time reached! Executing Season %d reset...", nextSession)

				// Wait a moment for stable connection
				time.Sleep(5 * time.Second)

				if isConnected() {
					err := uc.Execute(context.Background(), client(), groupID, nextSession)
					if err != nil {
						log.Printf("[SESSION RESET] Reset failed: %v", err)
					} else {
						log.Printf("[SESSION RESET] Season %d reset completed successfully!", nextSession)
					}
				} else {
					log.Println("[SESSION RESET] WARNING: Bot is not connected. Reset will be retried in 1 minute.")
					// Retry loop if not connected
					for i := 0; i < 10; i++ {
						time.Sleep(1 * time.Minute)
						if isConnected() {
							err := uc.Execute(context.Background(), client(), groupID, nextSession)
							if err != nil {
								log.Printf("[SESSION RESET] Retry %d failed: %v", i+1, err)
							} else {
								log.Printf("[SESSION RESET] Season %d reset completed on retry %d!", nextSession, i+1)
								break
							}
						}
					}
				}

				// Sleep a bit before scheduling the next one to avoid edge cases
				time.Sleep(1 * time.Minute)

			case <-ctx.Done():
				log.Println("[SESSION RESET] Scheduler cancelled.")
				return
			}
		}
	}()
}
