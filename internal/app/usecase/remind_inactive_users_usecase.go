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

// inactiveUserInfo holds pre-computed info for an inactive user.
type inactiveUserInfo struct {
	user         *domain.Report
	daysInactive int
}

// inactivity tier thresholds
const (
	tierCriticalDays = 60 // 2+ months
	tierWarningDays  = 30 // 1-2 months
	// everything else is 1-4 weeks (7-29 days)
)

func (u *RemindInactiveUsersUsecase) Execute(ctx context.Context, client *whatsmeow.Client, groupID string) (string, error) {
	if groupID == "" {
		return "", fmt.Errorf("group ID is not configured")
	}

	inactiveUsers, err := u.repo.GetInactiveUsers(ctx, 7)
	if err != nil {
		return "", fmt.Errorf("failed to get inactive users: %w", err)
	}

	if len(inactiveUsers) == 0 {
		return "Tidak ada user yang tidak laporan lebih dari seminggu. Mantap! 👍", nil
	}

	now := time.Now()
	todayDate := domain.GetToday(now)

	// Pre-compute days inactive for each user
	users := make([]inactiveUserInfo, 0, len(inactiveUsers))
	for _, user := range inactiveUsers {
		lastReportDate := domain.GetToday(user.LastReportDate)
		daysInactive := int(math.Round(todayDate.Sub(lastReportDate).Hours() / 24))
		users = append(users, inactiveUserInfo{user: user, daysInactive: daysInactive})
	}

	// Bucket into tiers
	var critical, warning, mild []inactiveUserInfo
	for _, u := range users {
		switch {
		case u.daysInactive >= tierCriticalDays:
			critical = append(critical, u)
		case u.daysInactive >= tierWarningDays:
			warning = append(warning, u)
		default:
			mild = append(mild, u)
		}
	}

	// Collect hall of fame - users with max streak >= 4 weeks (notable achievers)
	var hallOfFame []inactiveUserInfo
	for _, u := range users {
		if u.user.MaxStreak >= 4 {
			hallOfFame = append(hallOfFame, u)
		}
	}

	// Count users eligible for each comeback achievement
	comebackCounts := make(map[string]int)
	for _, u := range users {
		for _, a := range domain.AllComebackAchievements {
			if u.daysInactive >= a.MinInactiveDays && !domain.HasAchievement(u.user.Achievements, a.ID) {
				comebackCounts[a.Name]++
				break
			}
		}
	}

	messageText, mentions := BuildReminderMessage(users, critical, warning, mild, hallOfFame, comebackCounts)

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

// BuildReminderMessage builds the formatted reminder message and mention list.
// Extracted as a pure function for testability.
func BuildReminderMessage(
	users []inactiveUserInfo,
	critical, warning, mild, hallOfFame []inactiveUserInfo,
	comebackCounts map[string]int,
) (string, []string) {
	var sb strings.Builder
	var mentions []string

	sb.WriteString("📢 *PENGUMUMAN OLAHRAGA*\n\n")
	sb.WriteString(fmt.Sprintf("Ada *%d orang* yang belum lapor aktivitas lebih dari seminggu:\n", len(users)))

	// Hall of fame callout (if any)
	if len(hallOfFame) > 0 {
		sb.WriteString("\n⭐ *Pernah jadi juara streak:*\n")
		for _, u := range hallOfFame {
			sb.WriteString(fmt.Sprintf("  @%s (%s) - streak %d minggu 🔥\n", u.user.UserID, u.user.Name, u.user.MaxStreak))
			mentions = append(mentions, u.user.UserID+"@s.whatsapp.net")
		}
		sb.WriteString("Kalian pernah buktiin bisa konsisten, pasti bisa lagi!\n")
	}

	// Critical tier: 2+ months
	if len(critical) > 0 {
		sb.WriteString(fmt.Sprintf("\n🔴 *Sudah 2+ bulan absen* (%d orang)\n", len(critical)))
		writeTierUsers(&sb, critical, &mentions)
	}

	// Warning tier: 1-2 months
	if len(warning) > 0 {
		sb.WriteString(fmt.Sprintf("\n🟠 *Sudah 1-2 bulan absen* (%d orang)\n", len(warning)))
		writeTierUsers(&sb, warning, &mentions)
	}

	// Mild tier: 1-4 weeks
	if len(mild) > 0 {
		sb.WriteString(fmt.Sprintf("\n🟡 *Sudah 1-4 minggu absen* (%d orang)\n", len(mild)))
		writeTierUsers(&sb, mild, &mentions)
	}

	// Grouped comeback achievement hints
	if len(comebackCounts) > 0 {
		sb.WriteString("\n💡 *Comeback challenge tersedia:*\n")
		for _, a := range domain.AllComebackAchievements {
			if count, ok := comebackCounts[a.Name]; ok {
				sb.WriteString(fmt.Sprintf("  \"%s\" - %d orang bisa unlock! (%s)\n", a.Name, count, a.Description))
			}
		}
	}

	// Motivational closing
	sb.WriteString("\n")
	maxInactive := 0
	for _, u := range users {
		if u.daysInactive > maxInactive {
			maxInactive = u.daysInactive
		}
	}
	if maxInactive > 14 {
		sb.WriteString(motivationsLong[rand.Intn(len(motivationsLong))])
	} else {
		sb.WriteString(motivationsShort[rand.Intn(len(motivationsShort))])
	}

	mentions = deduplicateMentions(mentions)
	return sb.String(), mentions
}

// writeTierUsers writes a compact comma-separated list of @mentions for a tier.
func writeTierUsers(sb *strings.Builder, users []inactiveUserInfo, mentions *[]string) {
	for i, u := range users {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("@%s", u.user.UserID))
		*mentions = append(*mentions, u.user.UserID+"@s.whatsapp.net")
	}
	sb.WriteString("\n")
}

// deduplicateMentions removes duplicate JIDs while preserving order.
func deduplicateMentions(mentions []string) []string {
	seen := make(map[string]bool, len(mentions))
	result := make([]string, 0, len(mentions))
	for _, m := range mentions {
		if !seen[m] {
			seen[m] = true
			result = append(result, m)
		}
	}
	return result
}
