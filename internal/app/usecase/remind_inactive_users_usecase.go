package usecase

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

type RemindInactiveUsersUsecase struct {
	repo domain.ReportRepository
}

func NewRemindInactiveUsersUsecase(repo domain.ReportRepository) *RemindInactiveUsersUsecase {
	return &RemindInactiveUsersUsecase{repo: repo}
}

var motivationsShort = []string{
	"Semangat ya! Jangan lupa olahraga hari ini supaya badan tetap fit! 💪",
	"Kesehatan itu investasi. Yuk, mulai gerak lagi! ✨",
	"Olahraga sedikit lebih baik daripada tidak sama sekali. Ditunggu laporannya! 🏃‍♂️",
	"Ingat target awalmu! Ayo bangun dan mulai beraktivitas lagi! 🌟",
	"Jangan biarkan streak-mu hilang begitu saja. Yuk olahraga! 👟",
}

var motivationsLong = []string{
	"Sudah lama nih belum lapor. Tapi nggak ada kata terlambat! Comeback-mu bisa jadi inspirasi 🔥",
	"Yang penting bukan seberapa lama kamu berhenti, tapi keberanian untuk mulai lagi! 💪",
	"Banyak yang udah comeback dan makin keren! Giliranmu sekarang 🏆",
	"Kamu pernah streak tinggi, pasti bisa lagi! Ayo buktikan! ⚡",
	"Tubuhmu merindukanmu berolahraga. Yuk mulai hari ini! 🌅",
}

func (u *RemindInactiveUsersUsecase) Execute(ctx context.Context, client *whatsmeow.Client, groupID string) (string, error) {
	if groupID == "" {
		return "", fmt.Errorf("group ID is not configured")
	}

	// Get users who haven't reported in 7 days
	inactiveUsers, err := u.repo.GetInactiveUsers(ctx, 7)
	if err != nil {
		return "", fmt.Errorf("failed to get inactive users: %w", err)
	}

	if len(inactiveUsers) == 0 {
		return "Tidak ada user yang tidak laporan lebih dari seminggu. Mantap! 👍", nil
	}

	now := time.Now()
	var sb strings.Builder
	sb.WriteString("📢 *PENGUMUMAN OLAHRAGA*\n\n")
	sb.WriteString("Halo teman-teman, sudah seminggu lebih nih beberapa dari kita belum lapor aktivitas:\n\n")

	var mentions []string
	for _, user := range inactiveUsers {
		lastReport := user.LastReportDate
		lastReportDate := time.Date(lastReport.Year(), lastReport.Month(), lastReport.Day(), 0, 0, 0, 0, time.UTC)
		todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		daysInactive := int(math.Round(todayDate.Sub(lastReportDate).Hours() / 24))

		// Personalized info with lost streak
		sb.WriteString(fmt.Sprintf("- @%s (%s)", user.UserID, user.Name))
		if user.MaxStreak > 0 {
			sb.WriteString(fmt.Sprintf(" — pernah streak %d hari 🔥, sudah %d hari absen", user.MaxStreak, daysInactive))
		} else {
			sb.WriteString(fmt.Sprintf(" — sudah %d hari absen", daysInactive))
		}

		// Mention potential comeback achievement
		for _, a := range domain.AllComebackAchievements {
			if daysInactive >= a.MinInactiveDays && !domain.HasAchievement(user.Achievements, a.ID) {
				sb.WriteString(fmt.Sprintf(" (💡 bisa unlock \"%s\"!)", a.Name))
				break
			}
		}
		sb.WriteString("\n")

		mentions = append(mentions, user.UserID+"@s.whatsapp.net")
	}

	// Add motivation based on how long most users have been inactive
	sb.WriteString("\n")
	maxInactive := 0
	for _, user := range inactiveUsers {
		lastReport := user.LastReportDate
		lastReportDate := time.Date(lastReport.Year(), lastReport.Month(), lastReport.Day(), 0, 0, 0, 0, time.UTC)
		todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		days := int(math.Round(todayDate.Sub(lastReportDate).Hours() / 24))
		if days > maxInactive {
			maxInactive = days
		}
	}

	if maxInactive > 14 {
		sb.WriteString(motivationsLong[rand.Intn(len(motivationsLong))])
	} else {
		sb.WriteString(motivationsShort[rand.Intn(len(motivationsShort))])
	}

	messageText := sb.String()

	// Send message to group
	targetJID, _ := types.ParseJID(groupID)

	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &messageText,
			ContextInfo: &waE2E.ContextInfo{
				MentionedJID: mentions,
			},
		},
	}

	_, err = client.SendMessage(ctx, targetJID, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send reminder message: %w", err)
	}

	log.Printf("Sent inactivity reminder to %d users in group %s", len(inactiveUsers), groupID)
	return fmt.Sprintf("Berhasil mengirim pengingat ke %d user.", len(inactiveUsers)), nil
}
