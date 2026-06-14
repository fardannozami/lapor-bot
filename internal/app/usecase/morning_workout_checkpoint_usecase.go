package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type MorningWorkoutCheckpointUsecase struct {
	repo domain.ReportRepository
}

func NewMorningWorkoutCheckpointUsecase(repo domain.ReportRepository) *MorningWorkoutCheckpointUsecase {
	return &MorningWorkoutCheckpointUsecase{repo: repo}
}

func (u *MorningWorkoutCheckpointUsecase) Execute(ctx context.Context, now time.Time) (string, error) {
	reports, err := u.repo.GetAllReports(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get reports: %w", err)
	}

	activeToday, pendingToday := splitReportsByToday(reports, now)
	return BuildMorningWorkoutCheckpointMessage(activeToday, pendingToday), nil
}

var morningWorkoutMotivations = []string{
	"Keringat pagi adalah deposit kecil untuk energi seharian. Mulai dari 10 menit dulu. 💪",
	"Gerak sebelum hari makin ramai: jalan kaki, stretching, squat, atau workout ringan. Yang penting badan diajak hidup. 🌤️",
	"Seperti pepatah lama: air yang mengalir jarang keruh. Tubuh yang bergerak juga lebih siap menghadapi hari. 🌊",
	"Olahraga pagi bukan hukuman untuk tubuh, tapi ucapan terima kasih karena tubuh masih bisa bergerak. 🙌",
	"Jangan tunggu mood. Pakai sepatu, buka matras, mulai satu set. Momentum sering datang setelah gerakan pertama. 🚀",
	"Kesehatan adalah tabungan yang bunganya terasa di masa depan. Setor satu aktivitas sehat hari ini. 🪙",
	"Pagi ini cukup pilih satu kemenangan kecil: 15 menit jalan, 20 push-up, atau stretching sampai badan terasa ringan. ✨",
	"Tubuh yang aktif bikin pikiran lebih terang. Kalau pekerjaan menunggu, olahraga singkat bisa jadi pemanasan fokus. 🧠",
	"Hari yang sibuk tetap bisa diawali dengan pilihan sehat. Tidak harus sempurna; yang penting tidak nol. 🔥",
	"Bangun kesehatan seperti menanam pohon: siram sedikit setiap hari, nanti rindangnya terasa. 🌱",
}

func BuildMorningWorkoutCheckpointMessage(activeToday, pendingToday []*domain.Report) string {
	var sb strings.Builder

	sb.WriteString("🌤️ *09:09 Workout Checkpoint*\n\n")
	sb.WriteString(morningWorkoutMotivations[rand.Intn(len(morningWorkoutMotivations))])
	sb.WriteString("\n\n")

	if len(activeToday) > 0 {
		sb.WriteString(fmt.Sprintf("👏 *Sudah olahraga pagi ini (%d):*\n", len(activeToday)))
		sb.WriteString(formatReportNames(activeToday))
		sb.WriteString("\nKeren! Kalian sudah buka jalan dan jadi pemantik semangat buat grup. Respect! 🔥\n\n")
	} else {
		sb.WriteString("Belum ada yang laporan olahraga pagi ini. Siapa yang mau jadi pembuka dan nyalain semangat grup? 💪\n\n")
	}

	if len(pendingToday) > 0 {
		sb.WriteString(fmt.Sprintf("⏳ *Belum lapor hari ini (%d):*\n", len(pendingToday)))
		sb.WriteString(formatReportNames(pendingToday))
		sb.WriteString("\n\nMasih pagi — ambil 10–30 menit buat jalan kaki, stretching, bodyweight workout, atau olahraga favoritmu. Satu gerakan kecil hari ini tetap menang dari nol. 🚀")
	} else {
		sb.WriteString("Mantap, semua sudah lapor hari ini. Grup sehat full power! 🏆")
	}

	return sb.String()
}

func splitReportsByToday(reports []*domain.Report, now time.Time) (activeToday, pendingToday []*domain.Report) {
	today := domain.GetToday(now)
	activeToday = make([]*domain.Report, 0, len(reports))
	pendingToday = make([]*domain.Report, 0, len(reports))

	for _, report := range reports {
		if report == nil {
			continue
		}
		if domain.GetToday(report.LastReportDate).Equal(today) {
			activeToday = append(activeToday, report)
			continue
		}
		pendingToday = append(pendingToday, report)
	}

	sort.Slice(activeToday, func(i, j int) bool {
		return activeToday[i].LastReportDate.Before(activeToday[j].LastReportDate)
	})
	sort.Slice(pendingToday, func(i, j int) bool {
		return reportDisplayName(pendingToday[i]) < reportDisplayName(pendingToday[j])
	})

	return activeToday, pendingToday
}

func formatReportNames(reports []*domain.Report) string {
	names := make([]string, 0, len(reports))
	for _, report := range reports {
		names = append(names, reportDisplayName(report))
	}
	return strings.Join(names, ", ")
}

func reportDisplayName(report *domain.Report) string {
	if report == nil {
		return "Teman"
	}
	name := strings.TrimSpace(report.Name)
	if name != "" {
		return name
	}
	return report.UserID
}
