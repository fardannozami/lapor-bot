package usecase

type BroadcastUpdateUsecase struct{}

func NewBroadcastUpdateUsecase() *BroadcastUpdateUsecase {
	return &BroadcastUpdateUsecase{}
}

func (uc *BroadcastUpdateUsecase) Execute() string {
	return `📢 *PENGUMUMAN PENTING: SESSION 1 BERAKHIR!* ⚠️

Halo para pejuang keringat! 🏋️‍♂️

Kami ingin mengumumkan bahwa *Session 1* dari challenge "30 Days of Sweat" akan *resmi berakhir pada tanggal 30 April 2026*.

⏰ *Jadwal Penting:*
• 📅 *30 April 2026* — Hari terakhir Session 1
• 🔄 *1 Mei 2026* — Session 2 dimulai dari NOL

🗑️ *Yang Akan Di-Reset pada Session 2:*
• 🏆 Leaderboard — reset total
• 🔥 Streak mingguan — mulai dari 0
• 📊 Jumlah hari aktif — mulai dari 0
• 🏅 Achievements — reset semua
• ⭐ Points & Level — mulai dari awalZ
• 🛡️ Centurion Cycles — reset

💡 *Artinya:*
Semua data akan *dihapus* dan kita akan mulai dari awal lagi di Session 2. Ini adalah kesempatan baru bagi semua orang untuk bersaing dari titik yang sama! 🎯

📌 Manfaatkan sisa waktu di Session 1 ini sebaik mungkin. Laporkan aktivitas terakhirmu sebelum 30 April!

*Session 2 dimulai 1 Mei 2026. Siap-siap untuk petualangan baru!* 🚀🔥`
}
