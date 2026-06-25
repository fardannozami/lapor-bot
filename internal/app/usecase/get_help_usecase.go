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
		content: "*/lapor* atau *#lapor* — Laporkan workout atau aktivitas harianmu.\nContoh: `/lapor`, `#lapor`, `/lapor Push Day`, atau `#lapor Push Day`\n\n🔄 Setiap laporan yang valid akan menambah streak mingguanmu dan total hari aktif.\n❄️ *Streak Freeze* — kamu punya 1 freeze gratis per season. Freeze otomatis melindungi 1 minggu absen. Dapatkan +1 freeze lagi saat kamu mencapai 4 minggu streak!\n\n📌 *Max 3x laporan per hari*: Kamu bisa lapor maksimal 3x dalam sehari. Laporan ke-2 dan ke-3 tetap dihitung 1 hari tapi XP dibagi 2.\n📌 */lapor-kemarin* atau *#lapor-kemarin* — Laporan khusus untuk hari kemarin. Sama seperti /lapor, XP dibagi 2. Max 3x per hari.",
	},
	{
		emoji:   "✨",
		title:   "Side Quest Harian",
		content: "*/lapor sidequest* atau *#lapor sidequest* — Lihat side quest easy, medium, dan hard hari ini untuk profil yang sudah punya job.\n\nEasy: jalan kaki minimal 4.000 langkah atau sepeda 5 km (pilih salah satu). Medium/hard berisi latihan ringan yang bisa dilakukan di rumah/kantor, dan naik sedikit sesuai level job. XP bonus bervariasi per difficulty, dihitung otomatis di belakang.\n\nLapor dengan `/lapor sidequest [kegiatan] [jumlah]` atau `#lapor sidequest [kegiatan] [jumlah]`. Nama kegiatan harus sesuai yang tertera di daftar quest. Contoh: `/lapor sidequest jalan kaki 4000`. Target harus tercapai dulu; kalau kurang, laporan ditolak dan kamu bisa ulang setelah menambah aktivitas.",
	},
	{
		emoji:   "❌",
		title:   "Membatalkan Laporan",
		content: "*/cancel* atau *#cancel* — Batalkan laporan utama terakhir hari ini jika kamu salah input. Kalau hari ini ada 2-3 laporan utama, hanya laporan paling akhir yang dihapus.\n*/cancel-all* atau *#cancel-all* — Hapus semua laporan utama hari ini dan hitung ulang progresmu.\n*/cancel sidequest* atau *#cancel sidequest* — Batalkan side quest terakhir hari ini.\n*/cancel-all sidequest* atau *#cancel-all sidequest* — Hapus semua side quest hari ini.\nHanya bisa digunakan pada hari yang sama dengan laporan.",
	},
	{
		emoji:   "🏆",
		title:   "Leaderboard & Statistik",
		content: "Command leaderboard dan stats di WhatsApp sudah dipindah ke web supaya grup tidak ramai.\n\n🌐 Buka https://lapor-bot.web.id/ untuk cek klasemen, stats personal, ranking season, achievement, dan progres lain.",
	},
	{
		emoji:   "❓",
		title:   "Bantuan",
		content: "*/help* atau *#help* — Tampilkan list command ringkas.\n*/tutorial* atau *#tutorial* — Tampilkan panduan lengkap penggunaan bot.",
	},
	{
		emoji:   "⚔️",
		title:   "Status RPG (Attributes)",
		content: "Bot membaca laporan `/lapor` atau `#lapor` dengan daftar kata kunci yang sama untuk semua user. Setiap attribute yang cocok naik +1 per laporan valid.\n\n💪 *STR (Strength)*: gym, beban, push-up, pull-up, squat, plank, lunges, glute bridge, calf raises\n🏃‍♂️ *STA (Stamina)*: lari, jogging, sepeda, cardio, renang, hiking, tangga, jumping jacks, skipping\n⚡ *AGI (Agility)*: futsal/sepak bola, basket, badminton, tenis/padel, sprint, HIIT/tabata, boxing/muay thai, high knees, lateral shuffle\n🧘‍♂️ *VIT (Vitality)*: yoga, pilates, stretching, mobility, recovery, jalan santai, meditasi, breathing, balance\n\nJika tidak ada kata kunci spesifik, laporan otomatis masuk *VIT*.",
	},
}

type GetHelpUsecase struct{}

func NewGetHelpUsecase() *GetHelpUsecase {
	return &GetHelpUsecase{}
}

func (uc *GetHelpUsecase) Execute() string {
	return `🤖 *Command Lapor Bot*

📝 /lapor or #lapor — laporan aktivitas harian (max 3x/hari)
✨ /lapor sidequest or #lapor sidequest — lihat side quest hari ini
✨ /lapor sidequest [kegiatan] [jumlah] or #lapor sidequest [kegiatan] [jumlah] — lapor side quest
📌 /lapor-kemarin or #lapor-kemarin — laporan khusus hari kemarin (max 3x/hari)
❌ /cancel or #cancel — batalkan laporan terakhir hari ini
🧹 /cancel-all or #cancel-all — batalkan semua laporan hari ini
❌ /cancel sidequest or #cancel sidequest — batalkan side quest terakhir hari ini
🧹 /cancel-all sidequest or #cancel-all sidequest — batalkan semua side quest hari ini
📚 /tutorial or #tutorial — panduan lengkap penggunaan bot
❓ /help or #help — list command ini

🌐 Klasemen & stats personal: https://lapor-bot.web.id/`
}

func (uc *GetHelpUsecase) ExecuteAttributes() string {
	msg := "⚔️ *Status RPG (Attributes)* ⚔️\n\n"
	msg += "Setiap `/lapor` valid memberi +1 ke setiap attribute yang cocok. Aturannya sama untuk semua user; XP, level, streak, dan bonus tidak mengubah nilai attribute.\n\n"
	msg += "💪 *STR (Strength)*: gym, beban, push-up, pull-up, squat, plank, lunges, glute bridge, calf raises\n"
	msg += "🏃‍♂️ *STA (Stamina)*: lari/jogging, sepeda, cardio, renang, hiking, tangga, jumping jacks, skipping\n"
	msg += "⚡ *AGI (Agility)*: futsal/sepak bola, basket, badminton, tenis/padel, sprint, HIIT/tabata, boxing/muay thai, high knees, lateral shuffle\n"
	msg += "🧘‍♂️ *VIT (Vitality)*: yoga, pilates, stretching, mobility, recovery, jalan santai, meditasi, breathing, balance\n\n"
	msg += "💡 *Tips:* Jika tidak ada kata kunci di atas yang cocok, laporan otomatis masuk *VIT*."
	return msg
}

func (uc *GetHelpUsecase) ExecuteTutorial() string {
	msg := "📚 *Panduan Penggunaan Lapor Bot* 📚\n\n"
	msg += "Halo! Aku adalah bot untuk melacak aktivitas harian workout dan olahraga grup ini.\n"
	msg += "Kamu bisa menggunakan perintah dengan awalan `/` atau `#`:\n\n"

	for i, section := range helpSections {
		msg += fmt.Sprintf("%d. %s *%s*\n%s\n\n", i+1, section.emoji, section.title, section.content)
	}

	msg += "⚔️ *Level Numerik*\n"
	msg += "Level lifetime dimulai dari Lv.0 dan naik dari total points/EXP. Semakin tinggi level, semakin banyak EXP yang dibutuhkan untuk naik level. Season boleh reset, tapi level lifetime tetap lanjut.\n\n"
	msg += "🏅 *Badge*\n"
	msg += "Notifikasi /lapor hanya menampilkan badge terbaru supaya ringkas. Detail lengkap bisa dicek di web.\n\n"
	msg += "✨ *Flow Side Quest Harian*\n"
	msg += "1. Setiap pagi bot memberi reminder di grup untuk hunter yang sudah punya job.\n"
	msg += "2. Cek detail quest kamu dengan `/lapor sidequest` atau `#lapor sidequest`.\n"
	msg += "3. Pilih easy (jalan kaki atau sepeda), medium, hard, atau beberapa sekaligus.\n"
	msg += "4. Lapor dengan format `/lapor sidequest <kegiatan> <jumlah>` atau `#lapor sidequest <kegiatan> <jumlah>`, gunakan nama kegiatan yang tertera di daftar quest. Contoh: `/lapor sidequest jalan kaki 4000`, `/lapor sidequest sepeda 5 km`, `/lapor sidequest chair squat 18`.\n"
	msg += "5. Side quest memberi XP bonus kecil (bervariasi per difficulty), tetap masuk streak, stats, leaderboard, dan total side quest di web.\n\n"
	msg += "_Catatan: Bot hanya merespon di grup yang sudah dikonfigurasi. Semangat terus! 💪_\n\n" +
		"🌐 Klasemen & stats personal: https://lapor-bot.web.id/"
	return msg
}
