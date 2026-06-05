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
		content: "*#leaderboard* — Lihat klasemen lifetime (total hari aktif).\n*#leaderboard-weekly* — Lihat klasemen total hari aktif minggu ini.\n*#leaderboard-seasonal* — Lihat klasemen seasonal points.\n*#ranks* — Lihat ranking hunter selama season ini.\n*#mystats* — Cek statistik personal ringkas.",
	},
	{
		emoji:   "🎖️",
		title:   "Achievements",
		content: "*#achievements* — Lihat daftar badge season yang tersedia. Badge reset tiap season; level dan EXP lifetime tetap aman.\n*#comeback* — Cek progress comeback challenge-mu setelah absen.\n\n🏅 Kumpulkan badge dengan menjaga streak dan total hari aktif selama season!",
	},
	{
		emoji:   "🧭",
		title:   "Hunter Jobs",
		content: "*#jobs* — Lihat daftar job yang bisa dipilih.\n*#job [id]* — Pilih job untuk profilmu. Contoh: `#job ranger`\n\nJob tersedia: fighter, tank, assassin, mage, ranger, healer, necromancer. Job tampil di #mystats dan laporan #lapor.",
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
	return `🤖 *Command Lapor Bot*

📝 #lapor — laporan aktivitas harian
❌ #cancel — batalkan laporan hari ini
📊 #mystats — statistik personal
🏆 #leaderboard — leaderboard lifetime
📅 #leaderboard-weekly — leaderboard minggu ini
🏹 #leaderboard-seasonal — leaderboard season
🎖️ #ranks — rank hunter season
🏅 #achievements — detail badge dan syarat unlock
🔄 #comeback — progress comeback challenge
🧭 #jobs — daftar job
🧭 #job [id] — pilih job hunter
✨ #motivasi — quote motivasi
🏃 #strava — hubungkan Strava via chat pribadi
✏️ #setname [nama] — ubah nama tampil
📚 #tutorial — cara pakai bot lengkap
❓ #help — list command ini`
}

func (uc *GetHelpUsecase) ExecuteTutorial() string {
	msg := "📚 *Panduan Penggunaan Lapor Bot* 📚\n\n"
	msg += "Halo! Aku adalah bot untuk melacak aktivitas harian workout dan olahraga grup ini.\n"
	msg += "Kamu bisa menggunakan perintah berikut:\n\n"

	for i, section := range helpSections {
		msg += fmt.Sprintf("%d. %s *%s*\n%s\n\n", i+1, section.emoji, section.title, section.content)
	}

	msg += "⚔️ *Level Numerik*\n"
	msg += "Level lifetime dimulai dari Lv.0 dan naik dari total points/EXP. Semakin tinggi level, EXP yang dibutuhkan makin besar: `5×level² + 50×level + 100`. Season boleh reset, tapi level lifetime tetap lanjut.\n\n"
	msg += "🏅 *Badge*\n"
	msg += "Notifikasi #lapor hanya menampilkan badge terbaru supaya ringkas. Untuk syarat, poin, dan cerita unlock lengkap, buka #achievements.\n\n"
	msg += "_Catatan: Bot hanya merespon di grup yang sudah dikonfigurasi. Semangat terus! 💪_"
	return msg
}
