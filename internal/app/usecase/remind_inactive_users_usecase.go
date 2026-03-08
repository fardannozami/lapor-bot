package usecase

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"

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

var motivations = []string{
	"Semangat ya! Jangan lupa olahraga hari ini supaya badan tetap fit! 💪",
	"Kesehatan itu investasi. Yuk, mulai gerak lagi! ✨",
	"Sudah seminggu nih belum lapor. Ayo kalahkan rasa malasnya! 🔥",
	"Olahraga sedikit lebih baik daripada tidak sama sekali. Ditunggu laporannya! 🏃‍♂️",
	"Ingat target awalmu! Ayo bangun dan mulai beraktivitas lagi! 🌟",
	"Jangan biarkan streak-mu hilang begitu saja. Yuk olahraga! 👟",
}

func (u *RemindInactiveUsersUsecase) Execute(ctx context.Context, client *whatsmeow.Client, groupID string) (string, error) {
	if groupID == "" {
		return "", fmt.Errorf("group ID is not configured")
	}

	// 1. Get users who haven't reported in 7 days
	inactiveUsers, err := u.repo.GetInactiveUsers(ctx, 7)
	if err != nil {
		return "", fmt.Errorf("failed to get inactive users: %w", err)
	}

	if len(inactiveUsers) == 0 {
		return "Tidak ada user yang tidak laporan lebih dari seminggu. Mantap! 👍", nil
	}

	// 2. Construct message with tags
	var sb strings.Builder
	sb.WriteString("📢 *PENGUMUMAN OLAHRAGA*\n\n")
	sb.WriteString("Halo teman-teman, sudah seminggu lebih nih beberapa dari kita belum lapor aktivitas:\n\n")

	var mentions []string
	for _, user := range inactiveUsers {
		// Format for tagging: @phone
		sb.WriteString(fmt.Sprintf("- @%s (%s)\n", user.UserID, user.Name))
		mentions = append(mentions, user.UserID+"@s.whatsapp.net")
	}

	// 3. Add random motivation
	sb.WriteString("\n")
	sb.WriteString(motivations[rand.Intn(len(motivations))])

	messageText := sb.String()

	// 4. Send message to group
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
