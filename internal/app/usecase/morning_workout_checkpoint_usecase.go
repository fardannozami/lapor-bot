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

	sb.WriteString("🌤️ *Selamat pagi! 09:09 Workout Checkpoint*\n\n")
	sb.WriteString(morningWorkoutMotivations[rand.Intn(len(morningWorkoutMotivations))])
	sb.WriteString("\n\n")

	if len(activeToday) > 0 {
		sb.WriteString(fmt.Sprintf("👏 Sebelum jam ini sudah ada %d laporan olahraga. Keren—terima kasih sudah jadi pemantik energi pagi buat grup! 🔥\n\n", len(activeToday)))
	} else {
		sb.WriteString("Belum ada laporan olahraga pagi ini. Tidak apa-apa—mulai dari 10 menit gerak ringan juga sudah menang dari nol. 💪\n\n")
	}

	if len(pendingToday) > 0 {
		sb.WriteString("Masih pagi — ambil 10–30 menit buat jalan kaki, stretching, bodyweight workout, atau olahraga favoritmu. Setelah itu lapor di grup dengan `#lapor`. 🚀")
	} else {
		sb.WriteString("Mantap, semua sudah lapor hari ini. Grup sehat full power! 🏆")
	}

	sb.WriteString("\n\n✨ Kalau sudah punya job, cek bonus gerak harian dengan `#mysidequest`.")

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
