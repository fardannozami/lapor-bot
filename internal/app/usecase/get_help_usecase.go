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
		content: "*#lapor* — Laporkan workout atau aktivitas harianmu.\nContoh: `#lapor` atau `#lapor Push Day`\n\n*#lapor sidequest [kegiatan] [jumlah]* — Laporkan side quest dari `#mysidequest`. Gunakan nama kegiatan sesuai yang tertera di daftar quest. Contoh: `#lapor sidequest jalan kaki 4000`. Side quest memberi XP bonus kecil, tapi tetap dihitung untuk streak, stats, leaderboard, dan goal.\n\n🔄 Setiap laporan yang valid akan menambah streak mingguanmu dan total hari aktif.\n❄️ *Streak Freeze* — kamu punya 1 freeze gratis per season. Freeze otomatis melindungi 1 minggu absen. Dapatkan +1 freeze lagi saat kamu mencapai 4 minggu streak!\n\n📌 *Max 3x laporan per hari*: Kamu bisa lapor maksimal 3x dalam sehari. Laporan ke-2 dan ke-3 tetap dihitung 1 hari tapi XP dibagi 2.\n📌 *#lapor-kemarin* — Laporan khusus untuk hari kemarin. Sama seperti #lapor, XP dibagi 2. Max 3x per hari.",
	},
	{
		emoji:   "✨",
		title:   "Side Quest Harian",
		content: "*#mysidequest* — Lihat side quest easy, medium, dan hard hari ini. Wajib pilih job dulu lewat `#job <id>` agar mendapat side quest.\n\nEasy: jalan kaki minimal 4.000 langkah atau sepeda 5 km (pilih salah satu). Medium/hard berisi latihan ringan yang bisa dilakukan di rumah/kantor, dan naik sedikit sesuai level job. XP bonus bervariasi per difficulty, dihitung otomatis di belakang.\n\nLapor dengan `#lapor sidequest [kegiatan] [jumlah]`. Nama kegiatan harus sesuai yang tertera di `#mysidequest`. Target harus tercapai dulu; kalau kurang, laporan ditolak dan kamu bisa ulang setelah menambah aktivitas.",
	},
	{
		emoji:   "❌",
		title:   "Membatalkan Laporan",
		content: "*#cancel* — Batalkan laporan terakhir hari ini jika kamu salah input. Kalau hari ini ada 2-3 laporan, hanya laporan paling akhir yang dihapus.\n*#cancel-all* — Hapus semua laporan hari ini dan hitung ulang progresmu.\nHanya bisa digunakan pada hari yang sama dengan laporan.",
	},
	{
		emoji:   "🏆",
		title:   "Leaderboard & Statistik",
		content: "*#leaderboard* — Lihat klasemen lifetime (total hari aktif).\n*#leaderboard-weekly* — Lihat klasemen total hari aktif minggu ini (Senin—Minggu).\n*#leaderboard-seasonal* — Lihat klasemen seasonal points.\n*#ranks* — Lihat ranking hunter selama season ini.\n*#mystats* — Cek statistik personal ringkas.",
	},
	{
		emoji:   "🎯",
		title:   "Goal Mingguan",
		content: "*#goal set [1-7] [aktivitas]* — Tetapkan target hari aktif untuk 7 hari ke depan sejak command dikirim. Contoh: `#goal set 3 Olahraga`.\n*#goal* — Lihat progress, waktu mulai, dan waktu berakhir goal aktif.\n*#goal reset* — Hapus goal aktif kalau ingin set ulang.\n\nAlur: set goal → lapor aktivitas dengan #lapor → cek progress dengan #goal → jika target tercapai, total goals di #mystats bertambah.\nLaporan dobel di hari yang sama tetap dihitung 1. Data goal yang sudah lewat dibersihkan otomatis setiap 00:10.",
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
	{
		emoji:   "⚔️",
		title:   "Status RPG (Attributes)",
		content: "Bot akan membaca laporanmu dan menaikkan status tertentu berdasarkan kata kunci aktivitas.\n\n💪 *STR (Strength)*: beban, weight, strength, gym, angkat, powerlifting, push, pull, leg\n🏃‍♂️ *STA (Stamina)*: lari, run, running, sepeda, cycle, hiit, kardio, cardio, renang, swim\n⚡ *AGI (Agility)*: bola, futsal, basket, bulutangkis, tenis, sprint, muaythai, boxing, calisthenics, padel, padle\n🧘‍♂️ *VIT (Vitality)*: yoga, pilates, stretching, recovery, jalan, walk, meditasi\n\nJika tidak ada kata kunci spesifik, laporan akan otomatis meningkatkan *VIT*.",
	},
}

type GetHelpUsecase struct{}

func NewGetHelpUsecase() *GetHelpUsecase {
	return &GetHelpUsecase{}
}

func (uc *GetHelpUsecase) Execute() string {
	return `🤖 *Command Lapor Bot*

📝 #lapor — laporan aktivitas harian (max 3x/hari)
✨ #mysidequest — lihat side quest hari ini
✨ #lapor sidequest [kegiatan] [jumlah] — lapor side quest ½ XP
📌 #lapor-kemarin — laporan khusus hari kemarin (max 3x/hari)
❌ #cancel — batalkan laporan hari ini
🧹 #cancel-all — batalkan semua laporan hari ini
📊 #mystats — statistik personal
🎯 #goal set [1-7] [aktivitas] — set goal 7 hari sejak sekarang
🎯 #goal — progress goal aktif
🔄 #goal reset — reset goal aktif
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
⚔️ #attributes — info kata kunci status RPG
❓ #help — list command ini`
}

func (uc *GetHelpUsecase) ExecuteAttributes() string {
	msg := "⚔️ *Status RPG (Attributes)* ⚔️\n\n"
	msg += "Setiap kali kamu `#lapor`, bot akan membaca kata kunci dari laporanmu dan memberikan atribut ke status tertentu. Berikut daftar kata kuncinya:\n\n"
	msg += "💪 *STR (Strength)*: beban, weight, strength, gym, angkat, powerlifting, push, pull, leg\n"
	msg += "🏃‍♂️ *STA (Stamina)*: lari, run, running, sepeda, cycle, hiit, kardio, cardio, renang, swim\n"
	msg += "⚡ *AGI (Agility)*: bola, futsal, basket, bulutangkis, tenis, sprint, muaythai, boxing, calisthenics, padel, padle\n"
	msg += "🧘‍♂️ *VIT (Vitality)*: yoga, pilates, stretching, recovery, jalan, walk, meditasi\n\n"
	msg += "💡 *Tips:* Jika tidak ada kata kunci di atas yang cocok dalam teks laporanmu, secara default atributmu akan otomatis masuk ke *VIT*."
	return msg
}

func (uc *GetHelpUsecase) ExecuteTutorial() string {
	msg := "📚 *Panduan Penggunaan Lapor Bot* 📚\n\n"
	msg += "Halo! Aku adalah bot untuk melacak aktivitas harian workout dan olahraga grup ini.\n"
	msg += "Kamu bisa menggunakan perintah berikut:\n\n"

	for i, section := range helpSections {
		msg += fmt.Sprintf("%d. %s *%s*\n%s\n\n", i+1, section.emoji, section.title, section.content)
	}

	msg += "⚔️ *Level Numerik*\n"
	msg += "Level lifetime dimulai dari Lv.0 dan naik dari total points/EXP. Semakin tinggi level, semakin banyak EXP yang dibutuhkan untuk naik level. Season boleh reset, tapi level lifetime tetap lanjut.\n\n"
	msg += "🏅 *Badge*\n"
	msg += "Notifikasi #lapor hanya menampilkan badge terbaru supaya ringkas. Untuk syarat, poin, dan cerita unlock lengkap, buka #achievements.\n\n"
	msg += "🎯 *Flow Goal Mingguan*\n"
	msg += "1. Set target: `#goal set 3 Olahraga` (maksimal 7).\n"
	msg += "2. Window goal berjalan 7 hari dari waktu kamu set, bukan kalender Senin—Minggu.\n"
	msg += "3. Lapor aktivitas dengan `#lapor`; laporan dobel di hari yang sama tetap dihitung 1 untuk goal.\n"
	msg += "4. Cek progress kapan saja dengan `#goal`. Jika ingin ganti target saat masih aktif, pakai `#goal reset` dulu.\n"
	msg += "5. Saat target tercapai, total goals di `#mystats` dan `#achievements` bertambah.\n\n"
	msg += "✨ *Flow Side Quest Harian*\n"
	msg += "1. Pilih job dulu dengan `#job <id>` agar side quest terbuka.\n"
	msg += "2. Setiap pagi bot memberi reminder di grup; cek detail quest kamu dengan `#mysidequest`.\n"
	msg += "3. Pilih easy (jalan kaki atau sepeda), medium, hard, atau beberapa sekaligus.\n"
	msg += "4. Lapor dengan format `#lapor sidequest <kegiatan> <jumlah>`, gunakan nama kegiatan yang tertera di `#mysidequest`. Contoh: `#lapor sidequest jalan kaki 4000`, `#lapor sidequest sepeda 5 km`, `#lapor sidequest chair squat 18`.\n"
	msg += "5. Side quest memberi XP bonus kecil (bervariasi per difficulty), tetap masuk streak, stats, leaderboard, goal, dan total side quest di `#mystats`/`#achievements`.\n\n"
	msg += "_Catatan: Bot hanya merespon di grup yang sudah dikonfigurasi. Semangat terus! 💪_"
	return msg
}
