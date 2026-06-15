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

// SendDailyQuests prepares today's side quests and sends a group-only morning prompt at 04:00.
func (u *DailyQuestUsecase) SendDailyQuests(ctx context.Context, now time.Time, client *whatsmeow.Client, sender *queue.MessageSender, groupID string) error {
	_ = client
	if groupID == "" {
		return fmt.Errorf("group ID is not configured")
	}

	reports, err := u.repo.GetAllReports(ctx)
	if err != nil {
		return fmt.Errorf("failed to get reports: %w", err)
	}

	eligibleCount := 0

	for _, r := range reports {
		if r == nil || strings.TrimSpace(r.JobClass) == "" {
			continue
		}

		tasks, err := u.GetOrGenerateQuestList(ctx, r.UserID, r.JobClass, r.Level, now)
		if err != nil {
			log.Printf("[SIDE-QUEST] Failed to generate quest for %s: %v", r.UserID, err)
			continue
		}
		if len(tasks) > 0 {
			eligibleCount++
		}
	}

	if eligibleCount == 0 {
		return nil
	}

	dateStr := domain.GetToday(now).Format("02 Jan 2006")
	msgText := fmt.Sprintf("🌅 *Selamat pagi, Hunters!*\n\nSide quest hari ini sudah terbuka untuk %d hunter yang sudah memilih job. Buka `#mysidequest` untuk lihat pilihan easy, medium, dan hard hari ini.\n\nPilih yang ringan dulu juga boleh—jalan 4.000 langkah, sepeda 5 km, atau gerakan singkat di rumah/kantor. Side quest cuma bonus ½ XP; misi utama tetap konsisten olahraga dan lapor di grup. ✨\n\n📅 %s\n📝 Lapor side quest: `#lapor sidequest <kegiatan> <jumlah>`", eligibleCount, dateStr)

	targetJID, err := types.ParseJID(groupID)
	if err != nil {
		return fmt.Errorf("invalid group ID: %w", err)
	}

	msg := &waE2E.Message{
		Conversation: &msgText,
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
	if strings.TrimSpace(report.JobClass) == "" {
		return "🔒 *Side quest belum terbuka.*\n\nKamu perlu memilih job dulu agar bisa mendapat side quest harian. Cek `#jobs`, lalu pilih dengan `#job <id>` setelah syarat poin terpenuhi.", nil
	}

	tasks, err := u.GetOrGenerateQuestList(ctx, userID, report.JobClass, report.Level, now)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📜 *Side Quest Hari Ini - %s* 🏹\n", report.Name))
	sb.WriteString(fmt.Sprintf("Job: %s (Lv.%d)\n", domain.FormatJobClass(report.JobClass), report.Level))
	sb.WriteString("Reward: ½ XP per side quest yang valid. #lapor utama tetap prioritas.\n\n")

	completed := 0
	for i, t := range tasks {
		status := "⏳"
		if t.Progress >= t.Target {
			status = "✅"
			completed++
		}
		sb.WriteString(fmt.Sprintf("%s %d. %s\n", status, i+1, domain.FormatQuestProgressTask(t)))
	}

	sb.WriteString(fmt.Sprintf("\nProgress hari ini: %s (%d/%d selesai)\n", formatQuestProgressBar(tasks), completed, len(tasks)))
	sb.WriteString(fmt.Sprintf("Total side quest selesai: %d lifetime • %d season\n\n", report.TotalSideQuests, report.SeasonalSideQuests))
	sb.WriteString("Cara lapor: `#lapor sidequest [nama kegiatan] [jumlah]`\n")
	sb.WriteString("Contoh:\n")
	sb.WriteString("#lapor sidequest jalan 4000\n")
	sb.WriteString("#lapor sidequest sepeda 5 km\n")
	sb.WriteString("#lapor sidequest plank 60 detik")

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
	if strings.TrimSpace(report.JobClass) == "" {
		return "🔒 Side quest hanya tersedia setelah kamu memilih job. Cek `#jobs`, lalu pilih dengan `#job <id>` saat syarat poin sudah terpenuhi.", nil
	}
	today := domain.GetToday(now)
	dailyCount, err := u.repo.GetDailyActivityCount(ctx, userID, today)
	if err != nil {
		return "", err
	}
	if dailyCount >= MaxDailyReports {
		return fmt.Sprintf("Batas laporan harian sudah penuh (%dx). Side quest belum diterima agar progres quest tidak kepisah dari stats. Kalau salah input, pakai #cancel dulu lalu lapor ulang. 🙏", MaxDailyReports), nil
	}

	tasks, err := u.GetOrGenerateQuestList(ctx, userID, report.JobClass, report.Level, now)
	if err != nil {
		return "", err
	}

	var completedTasks []string
	var rejected []string

	for _, line := range inputLines {
		namePart, val := parseLineFloat(line)
		if namePart == "" || val <= 0 {
			continue
		}

		matched := false
		for idx, task := range tasks {
			if domain.MatchTask(namePart, task.ID) {
				matched = true
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

				if task.Progress >= task.Target {
					return fmt.Sprintf("%s sudah selesai hari ini. Pilih side quest lain di `#mysidequest` kalau masih mau lanjut. ✅", task.Name), nil
				}
				if added < task.Target {
					rejected = append(rejected, fmt.Sprintf("%s butuh minimal %s, laporanmu baru %s", task.Name, formatQuestTarget(task), formatQuestValue(task, added)))
					continue
				}

				tasks[idx].Progress = task.Target
				completedTasks = append(completedTasks, task.Name)
			}
		}
		if !matched {
			rejected = append(rejected, fmt.Sprintf("%q tidak cocok dengan side quest hari ini", namePart))
		}
	}

	if len(completedTasks) == 0 {
		if len(rejected) > 0 {
			return "Side quest belum diterima. Targetnya harus tercapai dulu ya. 💪\n\n- " + strings.Join(rejected, "\n- ") + "\n\nCek target hari ini dengan `#mysidequest`, lalu lapor ulang: `#lapor sidequest <kegiatan> <jumlah>`", nil
		}
		return "Gagal membaca laporan side quest. Contoh: `#lapor sidequest jalan 4000` atau `#lapor sidequest sepeda 5 km`", nil
	}

	// Save updated tasks list
	bytes, err := json.Marshal(tasks)
	if err != nil {
		return "", err
	}
	todayStr := today.Format("2006-01-02")
	if err := u.repo.SaveDailyQuest(ctx, userID, todayStr, string(bytes)); err != nil {
		return "", err
	}

	activityText := "Side quest: " + strings.Join(completedTasks, ", ")
	reportResult, err := reportUC.ExecuteSideQuest(ctx, userID, name, activityText, len(completedTasks), now)
	if err != nil {
		return "", err
	}

	// Format response message
	var sb strings.Builder
	if len(completedTasks) > 0 {
		sb.WriteString("🎉 *SIDE QUEST BERHASIL DISELESAIKAN!* 🏆\n\n")
		sb.WriteString("Selamat, kamu menyelesaikan:\n")
		for _, name := range completedTasks {
			sb.WriteString(fmt.Sprintf("- %s\n", name))
		}
		sb.WriteString("\n💰 Reward: ½ XP per side quest valid.\n\n")
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
	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(reportResult)

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

func formatQuestTarget(task domain.QuestTask) string {
	return formatQuestValue(task, task.Target)
}

func formatQuestValue(task domain.QuestTask, value int) string {
	if task.Unit == "100m" {
		return fmt.Sprintf("%.1f km", float64(value)/10.0)
	}
	return fmt.Sprintf("%d %s", value, task.Unit)
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
