package usecase

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
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
	"Belum terlambat untuk mulai lagi. Tubuhmu selalu siap, yang perlu dimulai adalah langkah pertamamu! 🚶",
	"Hujan, malas, sibuk? Tetap sempatkan 15 menit gerak. Streak tidak menunggu mood! ⏰",
	"Hari ini adalah hari yang sempurna untuk buktiin ke diri sendiri bahwa kamu bisa! 💪",
	"Liat teman-teman yang sudah lapor duluan. Yuk, join mereka hari ini! 🌟",
	"Badan yang segar dimulai dari keputusan kecil hari ini. Yuk mulai! ☀️",
	"10 menit workout di rumah = tetap on track. Nggak perlu sempurna, yang penting mulai! 🏠",
	"Sapa hari ini dengan satu gerakan kecil. Streak-mu menunggumu! 🔥",
}

var motivationsLong = []string{
	"Sudah lama nih belum lapor. Tapi nggak ada kata terlambat! Comeback-mu bisa jadi inspirasi 🔥",
	"Yang penting bukan seberapa lama kamu berhenti, tapi keberanian untuk mulai lagi! 💪",
	"Banyak yang udah comeback dan makin keren! Giliranmu sekarang 🏆",
	"Kamu pernah streak tinggi, pasti bisa lagi! Ayo buktikan! ⚡",
	"Tubuhmu merindukanmu berolahraga. Yuk mulai hari ini! 🌅",
	"Waktu terbaik untuk restart adalah SEKARANG. Riwayat absen tidak menghapus kemampuanmu. 💫",
	"Versi dirimu yang dulu pernah streak tinggi masih ada di dalam. Panggil dia keluar! 🦁",
	"Hampir semua orang pernah berhenti. Yang membedakan juara adalah mereka yang selalu kembali. 🥇",
	"Jangan bandingkan progressmu dengan orang lain. Bandingkan dengan dirimu 3 bulan lalu. Itu saja ukurannya. ⚖️",
	"Mulai dari 1 hari. Kalau 1 hari terlalu berat, mulai dari 10 menit. Yang penting tidak nol. 🎯",
	"Yang penting bukan berapa kali kamu jatuh, tapi berapa kali kamu berdiri lagi. Ayo berdiri! 🌅",
	"Hutang motivasi ke dirimu yang dulu pernah berjuang untuk streak itu. Bayar dengan satu laporan hari ini. 💎",
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

	weeklyGroups := groupInactiveUsersByWeeks(users)

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

	messageText, mentions := BuildReminderMessage(users, weeklyGroups, hallOfFame, comebackCounts)

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
	weeklyGroups map[int][]inactiveUserInfo,
	hallOfFame []inactiveUserInfo,
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

	writeWeeklyInactiveGroups(&sb, weeklyGroups, &mentions)

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

	// Append health pillar reminders (olahraga, makanan, istirahat, kelola stres)
	sb.WriteString(BuildWellnessReminder())

	mentions = deduplicateMentions(mentions)
	return sb.String(), mentions
}

func groupInactiveUsersByWeeks(users []inactiveUserInfo) map[int][]inactiveUserInfo {
	groups := make(map[int][]inactiveUserInfo)
	for _, user := range users {
		weeksInactive := user.daysInactive / 7
		if weeksInactive < 1 {
			weeksInactive = 1
		}
		if weeksInactive > 3 {
			weeksInactive = 3
		}
		groups[weeksInactive] = append(groups[weeksInactive], user)
	}
	return groups
}

func writeWeeklyInactiveGroups(sb *strings.Builder, groups map[int][]inactiveUserInfo, mentions *[]string) {
	if len(groups) == 0 {
		return
	}

	weeks := make([]int, 0, len(groups))
	for week := range groups {
		weeks = append(weeks, week)
	}
	sort.Ints(weeks)

	sb.WriteString("\n📅 *List belum olahraga / streak terputus:*\n")
	for _, week := range weeks {
		users := groups[week]
		sb.WriteString(fmt.Sprintf("\n%s (%d orang)\n", weeklyReminderTitle(week), len(users)))
		writeMentionList(sb, users, mentions)
	}
}

func weeklyReminderTitle(week int) string {
	if week <= 1 {
		return "🟡 *1 minggu belum lapor*"
	}
	if week == 2 {
		return "🟠 *2 minggu belum lapor*"
	}
	return "🔴 *3+ minggu belum lapor*"
}

// writeMentionList writes a compact comma-separated list of @mentions.
func writeMentionList(sb *strings.Builder, users []inactiveUserInfo, mentions *[]string) {
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
