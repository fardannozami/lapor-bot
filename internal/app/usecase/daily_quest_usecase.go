package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
	"github.com/fardannozami/whatsapp-gateway/internal/queue"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

type DailyQuestUsecase struct {
	repo domain.ReportRepository
}

func NewDailyQuestUsecase(repo domain.ReportRepository) *DailyQuestUsecase {
	return &DailyQuestUsecase{repo: repo}
}

// GetOrGenerateQuestList retrieves today's quest list from database, or generates a new one.
func (u *DailyQuestUsecase) GetOrGenerateQuestList(ctx context.Context, userID, jobClass string, level int, now time.Time) ([]domain.QuestTask, error) {
	todayStr := domain.GetToday(now).Format("2006-01-02")
	tasksJSON, err := u.repo.GetDailyQuest(ctx, userID, todayStr)
	if err != nil {
		return nil, err
	}

	if tasksJSON != "" {
		var tasks []domain.QuestTask
		if err := json.Unmarshal([]byte(tasksJSON), &tasks); err == nil {
			return tasks, nil
		}
	}

	// Generate and save new quest
	tasks := domain.GenerateDailyQuest(jobClass, level, now)
	bytes, err := json.Marshal(tasks)
	if err != nil {
		return nil, err
	}

	if err := u.repo.SaveDailyQuest(ctx, userID, todayStr, string(bytes)); err != nil {
		return nil, err
	}

	return tasks, nil
}

// SendDailyQuests sends the consolidated daily quest list to the configured group at 04:00.
func (u *DailyQuestUsecase) SendDailyQuests(ctx context.Context, now time.Time, client *whatsmeow.Client, sender *queue.MessageSender, groupID string) error {
	if groupID == "" {
		return fmt.Errorf("group ID is not configured")
	}

	reports, err := u.repo.GetAllReports(ctx)
	if err != nil {
		return fmt.Errorf("failed to get reports: %w", err)
	}

	dateStr := domain.GetToday(now).Format("02-01-2006")

	var sb strings.Builder
	sb.WriteString("⚔️ *DAILY QUEST HARIAN HUNTER* ⚔️\n")
	sb.WriteString(fmt.Sprintf("Tanggal: %s\n\n", dateStr))
	sb.WriteString("Berikut adalah quest harian untuk para Hunter hari ini:\n\n")

	var mentions []string
	hasQuests := false
	anyUserWithoutJob := false

	for _, r := range reports {
		if r == nil {
			continue
		}

		tasks, err := u.GetOrGenerateQuestList(ctx, r.UserID, r.JobClass, r.Level, now)
		if err != nil {
			log.Printf("[DAILY-QUEST] Failed to generate quest for %s: %v", r.UserID, err)
			continue
		}

		hasQuests = true

		sb.WriteString(fmt.Sprintf("👤 @%s\n", r.UserID))
		mentions = append(mentions, r.UserID+"@s.whatsapp.net")

		if r.JobClass == "" {
			sb.WriteString("Job: - (General Quest)\n")
			anyUserWithoutJob = true
		} else {
			sb.WriteString(fmt.Sprintf("Job: %s (Lv.%d)\n", domain.FormatJobClass(r.JobClass), r.Level))
		}

		sb.WriteString("📜 *Daftar Quest:*\n")
		for i, t := range tasks {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, domain.FormatQuestProgressTask(t)))
		}

		sb.WriteString(fmt.Sprintf("Progress: %s\n\n", formatQuestProgressBar(tasks)))
	}

	if !hasQuests {
		return nil
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("Cara melapor:\n")
	sb.WriteString("Kirim pesan di GRUP dengan format `#lapor-quest [nama-quest] [jumlah]`.\n")
	sb.WriteString("Kamu juga bisa melaporkan beberapa quest sekaligus secara bersamaan:\n")
	sb.WriteString("#lapor-quest\n")
	sb.WriteString("- push up 10\n")
	sb.WriteString("- squat 15\n")

	if anyUserWithoutJob {
		sb.WriteString("\n💡 *Tip:* Bagi Hunter yang belum memilih job, kumpulkan minimal 50 poin agar bisa memilih job khusus (Fighter, Tanker, Assassin, Mage, Ranger, Healer, Necromancer) untuk mendapatkan quest khusus dengan reward yang lebih besar!\n")
	}

	sb.WriteString("\nSemangat berlatih dan terus tingkatkan rank-mu! 🔥")

	targetJID, err := types.ParseJID(groupID)
	if err != nil {
		return fmt.Errorf("invalid group ID: %w", err)
	}

	msgText := sb.String()
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &msgText,
			ContextInfo: &waE2E.ContextInfo{
				MentionedJID: mentions,
			},
		},
	}

	// Send to group
	if err := sender.SendHighPriority(ctx, targetJID, msg); err != nil {
		return fmt.Errorf("failed to send daily quest to group: %w", err)
	}

	return nil
}

// ViewQuest displays the user's quest checklist.
func (u *DailyQuestUsecase) ViewQuest(ctx context.Context, userID, name string, now time.Time) (string, error) {
	report, err := u.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	if report == nil {
		return fmt.Sprintf("Halo %s, kamu belum terdaftar di database. Silakan lakukan laporan pertama dengan `#lapor` terlebih dahulu! 💪", name), nil
	}

	tasks, err := u.GetOrGenerateQuestList(ctx, userID, report.JobClass, report.Level, now)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📜 *Quest Harian - %s* 🏹\n", report.Name))
	if report.JobClass == "" {
		sb.WriteString("Job: - (General Quest)\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("Job: %s (Lv.%d)\n\n", domain.FormatJobClass(report.JobClass), report.Level))
	}

	for i, t := range tasks {
		status := "⏳"
		if t.Progress >= t.Target {
			status = "✅"
		}
		sb.WriteString(fmt.Sprintf("%s %d. %s\n", status, i+1, domain.FormatQuestProgressTask(t)))
	}

	sb.WriteString(fmt.Sprintf("\nProgress: %s\n\n", formatQuestProgressBar(tasks)))
	sb.WriteString("Ketik `#lapor-quest [nama] [jumlah]` untuk update progres!\n")
	sb.WriteString("Contoh:\n")
	sb.WriteString("#lapor-quest push-up 5\n\n")
	sb.WriteString("Atau laporkan sekaligus:\n")
	sb.WriteString("#lapor-quest\n")
	sb.WriteString("- push up 10\n")
	sb.WriteString("- squat 15")

	return sb.String(), nil
}

// UpdateProgress updates the progressive reps or minutes for matched tasks.
func (u *DailyQuestUsecase) UpdateProgress(ctx context.Context, userID, name string, inputLines []string, reportUC *ReportActivityUsecase, now time.Time) (string, error) {
	report, err := u.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	if report == nil {
		return fmt.Sprintf("Halo %s, kamu belum terdaftar di database. Silakan lakukan laporan pertama dengan `#lapor` terlebih dahulu! 💪", name), nil
	}

	tasks, err := u.GetOrGenerateQuestList(ctx, userID, report.JobClass, report.Level, now)
	if err != nil {
		return "", err
	}

	updated := false
	pointsAwarded := 0
	var completedTasks []string

	for _, line := range inputLines {
		namePart, val := parseLineFloat(line)
		if namePart == "" || val <= 0 {
			continue
		}

		for idx, task := range tasks {
			if domain.MatchTask(namePart, task.ID) {
				added := 0
				if task.Unit == "100m" {
					if val < 50.0 { // e.g. "lari 2.5" is 2.5 km -> 25 units of 100m
						added = int(val * 10.0)
					} else {
						added = int(val)
					}
				} else {
					added = int(val)
				}

				wasCompleted := task.Progress >= task.Target
				tasks[idx].Progress += added
				updated = true

				if !wasCompleted && tasks[idx].Progress >= task.Target {
					pointsAwarded += task.RewardPoints
					completedTasks = append(completedTasks, task.Name)
				}
			}
		}
	}

	if !updated {
		return "Gagal memperbarui progres. Pastikan format penulisan benar.\nContoh: `#lapor-quest push-up 5`", nil
	}

	// Save updated tasks list
	bytes, err := json.Marshal(tasks)
	if err != nil {
		return "", err
	}
	todayStr := domain.GetToday(now).Format("2006-01-02")
	if err := u.repo.SaveDailyQuest(ctx, userID, todayStr, string(bytes)); err != nil {
		return "", err
	}

	// Update user report points
	if pointsAwarded > 0 {
		report.TotalPoints += pointsAwarded
		report.SeasonalPoints += pointsAwarded
		report.Level = domain.NumericLevelFromTotalPoints(report.TotalPoints)
		if err := u.repo.UpsertReport(ctx, report); err != nil {
			return "", err
		}
	}

	// Check if this is their first quest update of today to trigger auto-#lapor
	var autoLaporResult string
	today := domain.GetToday(now)
	reportedToday := domain.GetToday(report.LastReportDate).Equal(today)

	if !reportedToday {
		autoLaporResult, err = reportUC.Execute(ctx, userID, name, nil)
		if err != nil {
			log.Printf("[DAILY-QUEST] Failed to auto-trigger #lapor: %v", err)
		}
	}

	// Format response message
	var sb strings.Builder
	if len(completedTasks) > 0 {
		sb.WriteString("🎉 *QUEST BERHASIL DISELESAIKAN!* 🏆\n\n")
		sb.WriteString("Selamat, kamu menyelesaikan:\n")
		for _, name := range completedTasks {
			sb.WriteString(fmt.Sprintf("- %s\n", name))
		}
		sb.WriteString(fmt.Sprintf("\n💰 Reward: +%d poin ditambahkan ke profilmu!\n\n", pointsAwarded))
	} else {
		sb.WriteString("📈 *PROGRES QUEST HARIAN DIUPDATE*\n\n")
	}

	sb.WriteString("📜 *Daftar Quest Saat Ini:*\n")
	for i, t := range tasks {
		status := "⏳"
		if t.Progress >= t.Target {
			status = "✅"
		}
		sb.WriteString(fmt.Sprintf("%s %d. %s\n", status, i+1, domain.FormatQuestProgressTask(t)))
	}

	sb.WriteString(fmt.Sprintf("\nProgress: %s\n", formatQuestProgressBar(tasks)))

	if autoLaporResult != "" {
		sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		sb.WriteString("🤖 *AUTO #LAPOR DETECTED:*\n")
		sb.WriteString(autoLaporResult)
	}

	return sb.String(), nil
}

func parseLineFloat(line string) (string, float64) {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "-")
	line = strings.TrimPrefix(line, "*")
	line = strings.TrimPrefix(line, "•")
	line = strings.TrimSpace(line)

	re := regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) == 0 {
		return "", 0
	}

	numStr := matches[1]
	namePart := strings.Replace(line, numStr, "", 1)
	namePart = strings.TrimSpace(namePart)
	for _, suffix := range []string{"x", "menit", "km", "m", "detik", "langkah", "steps", "ml"} {
		namePart = strings.TrimSuffix(namePart, suffix)
	}
	namePart = strings.TrimSpace(namePart)

	var val float64
	fmt.Sscanf(numStr, "%f", &val)
	return namePart, val
}

func formatQuestProgressBar(tasks []domain.QuestTask) string {
	totalTarget := 0
	totalProgress := 0
	for _, t := range tasks {
		prog := t.Progress
		if prog > t.Target {
			prog = t.Target
		}
		totalTarget += t.Target
		totalProgress += prog
	}

	percentage := 0
	if totalTarget > 0 {
		percentage = (totalProgress * 100) / totalTarget
	}

	barLen := 10
	filled := (percentage * barLen) / 100

	var bar strings.Builder
	for i := 0; i < barLen; i++ {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}
	return fmt.Sprintf("[%s] %d%%", bar.String(), percentage)
}
