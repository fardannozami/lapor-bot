package usecase

import "fmt"

var helpSections = []struct {
	emoji   string
	title   string
	content string
}{
	{
		emoji:   "📝",
		title:   "Melaporkan Aktivitas",
		content: "*#lapor* — Laporkan workout atau aktivitas harianmu.\nContoh: `#lapor` atau `#lapor Push Day`\n\n🔄 Setiap laporan yang valid akan menambah streak mingguanmu dan total hari aktif.\n❄️ *Streak Freeze* — kamu punya 1 freeze gratis per season. Freeze otomatis melindungi 1 minggu absen. Dapatkan +1 freeze lagi saat kamu mencapai 4 minggu streak!",
	},
	{
		emoji:   "❌",
		title:   "Membatalkan Laporan",
		content: "*#cancel* — Batalkan laporan hari ini jika kamu salah lapor.\nHanya bisa digunakan pada hari yang sama dengan laporan.",
	},
	{
		emoji:   "🏆",
		title:   "Leaderboard & Statistik",
		content: "*#leaderboard* — Lihat klasemen lifetime (total hari aktif).\n*#leaderboard-weekly* — Lihat klasemen total hari aktif minggu ini.\n*#leaderboard-seasonal* — Lihat klasemen seasonal points.\n*#mystats* — Cek statistik personal (streak, poin, achievements, seasonal progress).",
	},
	{
		emoji:   "🎖️",
		title:   "Achievements",
		content: "*#achievements* — Lihat daftar semua achievement yang tersedia.\n*#comeback* — Cek progress comeback challenge-mu setelah absen.\n\n🏅 Kumpulkan badge dengan menjaga streak dan total hari aktif!",
	},
	{
		emoji:   "✨",
		title:   "Motivasi",
		content: "*#motivasi* — Dapatkan quote motivasi acak untuk semangat berolahraga!",
	},
	{
		emoji:   "🏃‍♂️",
		title:   "Integrasi Strava",
		content: "*#strava* — Hubungkan akun Strava untuk laporan otomatis.\nAktivitas lari dan bersepedamu akan tercatat secara otomatis!",
	},
	{
		emoji:   "✏️",
		title:   "Pengaturan Profil",
		content: "*#setname [nama]* — Ubah nama yang tampil di leaderboard.\nContoh: `#setname Budi`",
	},
	{
		emoji:   "❓",
		title:   "Bantuan",
		content: "*#help* — Tampilkan panduan ini kapan saja!",
	},
}

type GetHelpUsecase struct{}

func NewGetHelpUsecase() *GetHelpUsecase {
	return &GetHelpUsecase{}
}

func (uc *GetHelpUsecase) Execute() string {
	msg := "📚 *Panduan Penggunaan Lapor Bot* 📚\n\n"
	msg += "Halo! Aku adalah bot untuk melacak aktivitas harian workout dan olahraga grup ini.\n"
	msg += "Kamu bisa menggunakan perintah berikut:\n\n"

	for i, section := range helpSections {
		msg += fmt.Sprintf("%d. %s *%s*\n%s\n\n", i+1, section.emoji, section.title, section.content)
	}

	msg += "_Catatan: Bot hanya merespon di grup yang sudah dikonfigurasi. Semangat terus! 💪_"
	return msg
}
