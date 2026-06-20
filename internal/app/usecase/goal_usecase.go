package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

const (
	maxWeeklyGoalDays = 7
	goalWindowDays    = 7
)

var shortIndonesianDays = []string{"Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"}
var indonesianMonths = []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
var goalLocation = time.FixedZone("WIB", 7*60*60)

type GoalUsecase struct {
	repo domain.ReportRepository
	now  func() time.Time
}

func NewGoalUsecase(repo domain.ReportRepository) *GoalUsecase {
	return &GoalUsecase{repo: repo, now: time.Now}
}

func (uc *GoalUsecase) Execute(ctx context.Context, userID, name, message string) (string, error) {
	args := strings.Fields(strings.TrimSpace(message))
	if len(args) >= 2 {
		switch strings.ToLower(args[1]) {
		case "set":
			return uc.set(ctx, userID, args[2:])
		case "reset":
			return uc.reset(ctx, userID)
		}
	}

	return uc.status(ctx, userID)
}

func (uc *GoalUsecase) set(ctx context.Context, userID string, args []string) (string, error) {
	if len(args) == 0 {
		return "Format goal: #goal set <1-7> [aktivitas]\nContoh: #goal set 3 Olahraga", nil
	}

	targetDays, err := strconv.Atoi(args[0])
	if err != nil || targetDays < 1 {
		return "Target goal harus angka 1 sampai 7. Contoh: #goal set 3 Olahraga", nil
	}
	if targetDays > maxWeeklyGoalDays {
		return "Target goal maksimal 7 hari. 1 minggu = 7 hari ya 🙏", nil
	}

	now := uc.now()
	activity := strings.TrimSpace(strings.Join(args[1:], " "))
	if activity == "" {
		activity = "Olahraga"
	}

	return uc.setGoalWithStart(ctx, userID, targetDays, activity, now)
}

// SetWithStart allows setting a goal with a custom start time (used by web dashboard).
// start time should already be normalized (e.g. desired midnight or chosen hour in the user's TZ).
func (uc *GoalUsecase) SetWithStart(ctx context.Context, userID string, targetDays int, activity string, start time.Time) (string, error) {
	if targetDays < 1 || targetDays > maxWeeklyGoalDays {
		return "Target goal harus angka 1 sampai 7.", nil
	}
	if strings.TrimSpace(activity) == "" {
		activity = "Olahraga"
	}
	return uc.setGoalWithStart(ctx, userID, targetDays, activity, start)
}

func (uc *GoalUsecase) setGoalWithStart(ctx context.Context, userID string, targetDays int, activity string, start time.Time) (string, error) {
	existing, err := uc.repo.GetActiveGoal(ctx, userID, start)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return "Kamu masih punya goal aktif. Pakai #goal reset dulu kalau mau menggantinya. 🎯", nil
	}

	goal := &domain.WeeklyGoal{
		UserID:     userID,
		TargetDays: targetDays,
		Activity:   activity,
		StartAt:    start,
		EndAt:      start.AddDate(0, 0, goalWindowDays),
		CreatedAt:  start,
	}
	if err := uc.repo.SetGoal(ctx, goal); errors.Is(err, domain.ErrActiveGoalExists) {
		return "Kamu masih punya goal aktif. Pakai #goal reset dulu kalau mau menggantinya. 🎯", nil
	} else if err != nil {
		return "", err
	}

	return fmt.Sprintf("🎯 Goal aktif diset: %dx %s\n%s\n\nLaporkan aktivitas dengan #lapor. Laporan lebih dari 1x di hari yang sama tetap dihitung 1 untuk goal.", targetDays, activity, formatGoalPeriod(goal.StartAt, goal.EndAt)), nil
}

func (uc *GoalUsecase) reset(ctx context.Context, userID string) (string, error) {
	now := uc.now()
	goal, err := uc.repo.GetActiveGoal(ctx, userID, now)
	if err != nil {
		return "", err
	}
	if goal == nil {
		return "Belum ada goal aktif untuk di-reset. Buat dengan #goal set <1-7> [aktivitas].", nil
	}
	if err := uc.repo.DeleteActiveGoal(ctx, userID, now); err != nil {
		return "", err
	}
	return "Goal aktif sudah dihapus. Kamu bisa set ulang dengan #goal set <1-7> [aktivitas]. 🔄", nil
}

func (uc *GoalUsecase) status(ctx context.Context, userID string) (string, error) {
	now := uc.now()
	goal, err := uc.repo.GetActiveGoal(ctx, userID, now)
	if err != nil {
		return "", err
	}
	if goal == nil {
		return "Belum ada goal aktif. Buat dengan #goal set <1-7> [aktivitas].\nContoh: #goal set 3 Olahraga", nil
	}

	activities, err := uc.repo.GetGoalActivities(ctx, userID, goal.StartAt, goal.EndAt)
	if err != nil {
		return "", err
	}

	return formatGoalStatus(goal, activities, now), nil
}

func (uc *GoalUsecase) RecordActivity(ctx context.Context, userID string, activityAt time.Time, activityText string) (bool, error) {
	return uc.repo.RecordGoalActivity(ctx, userID, activityAt, activityText)
}

func (uc *GoalUsecase) CleanupExpired(ctx context.Context, now time.Time) (int64, error) {
	return uc.repo.DeleteExpiredGoals(ctx, now)
}

func formatGoalStatus(goal *domain.WeeklyGoal, activities []domain.GoalActivity, now time.Time) string {
	activityByDate := make(map[string]string, len(activities))
	for _, activity := range activities {
		activityByDate[activity.Date.Format(time.DateOnly)] = strings.TrimSpace(activity.Activity)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🎯 GOAL AKTIF: %dx %s\n", goal.TargetDays, goal.Activity))
	sb.WriteString(formatGoalPeriod(goal.StartAt, goal.EndAt))
	sb.WriteString("\n>\n")
	sb.WriteString("| Hari | Status | Aktivitas |\n")
	sb.WriteString("|:---|:---:|---:|\n")

	completedDays := len(activityByDate)
	startDate := goalReportDate(goal.StartAt)
	endDate := goalReportDate(goal.EndAt.Add(-time.Nanosecond))
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		activity, ok := activityByDate[date.Format(time.DateOnly)]
		status := "⬜"
		if ok {
			status = "✅"
		}
		if activity == "" {
			activity = "—"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", shortIndonesianDays[date.Weekday()], status, activity))
	}

	if completedDays > goal.TargetDays {
		completedDays = goal.TargetDays
	}
	remaining := goal.TargetDays - completedDays
	sb.WriteString(">\n")
	sb.WriteString(fmt.Sprintf("Progress: %s %d/%d\n>\n", formatGoalProgressBar(completedDays), completedDays, goal.TargetDays))
	if remaining <= 0 {
		sb.WriteString("🏆 Goal tercapai! Disiplinmu naik level — pertahankan momentumnya. 🎯")
	} else {
		sb.WriteString(fmt.Sprintf("🚀 Kurang %d hari lagi buat capai goal!", remaining))
	}

	if now.Before(goal.EndAt) {
		sb.WriteString(fmt.Sprintf("\n⏳ Sisa waktu: %s", formatGoalRemaining(goal.EndAt.Sub(now))))
	}

	return sb.String()
}

func formatGoalProgressBar(done int) string {
	if done < 0 {
		done = 0
	}
	if done > 7 {
		done = 7
	}
	return strings.Repeat("▓", done) + strings.Repeat("░", 7-done)
}

func formatGoalPeriod(startAt, endAt time.Time) string {
	return fmt.Sprintf("📅 Mulai: %s\n🏁 Berakhir: %s", formatGoalTime(startAt), formatGoalTime(endAt))
}

func formatGoalTime(t time.Time) string {
	t = t.In(goalLocation)
	return fmt.Sprintf("%02d %s %d %02d:%02d", t.Day(), indonesianMonths[int(t.Month())], t.Year(), t.Hour(), t.Minute())
}

func formatGoalRemaining(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(time.Minute)
	days := int(d / (24 * time.Hour))
	d -= time.Duration(days) * 24 * time.Hour
	hours := int(d / time.Hour)
	d -= time.Duration(hours) * time.Hour
	minutes := int(d / time.Minute)

	parts := make([]string, 0, 3)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d hari", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d jam", hours))
	}
	if minutes > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%d menit", minutes))
	}
	return strings.Join(parts, " ")
}

func goalReportDate(t time.Time) time.Time {
	return domain.GetToday(t.UTC())
}

func goalActivityText(workout *domain.HevyWorkout) string {
	return goalActivityTextWithFallback(workout, "")
}

func goalActivityTextWithFallback(workout *domain.HevyWorkout, fallback string) string {
	if workout == nil || strings.TrimSpace(workout.Title) == "" {
		if text := strings.TrimSpace(fallback); text != "" {
			return text
		}
		return "Olahraga"
	}
	return strings.TrimSpace(workout.Title)
}
