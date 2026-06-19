package usecase

import (
	"fmt"
	"time"
)

type BroadcastUpdateUsecase struct{}

func NewBroadcastUpdateUsecase() *BroadcastUpdateUsecase {
	return &BroadcastUpdateUsecase{}
}

func (uc *BroadcastUpdateUsecase) Execute() string {
	seasonNumber, _ := GetCurrentSessionInfo(time.Now())
	nextSeason, _ := GetCurrentSessionInfo(GetNextResetTime(time.Now()))

	return fmt.Sprintf(`📢 *PENGUMUMAN SEASON CHALLENGE* ⚠️

Halo para pejuang keringat! 🏋️‍♂️

Season %d sedang berjalan. Setelah season berjalan selesai, bot akan otomatis lanjut ke Season %d dan seterusnya setiap 4 bulan.

⏰ *Jadwal Penting:*
• 📅 Season berjalan: *Season %d*
• 🔄 Reset berikutnya: *Season %d* dimulai dari seasonal progress baru

🗑️ *Yang Akan Di-Reset Saat Season Baru:*
• 📊 Seasonal Points
• 📅 Seasonal Activity
• ⚔️ Seasonal Max Streak
• 🏅 Season Badges
• 🏆 Rank & Seasonal Leaderboard

💾 *Yang Tetap Aman:*
• ⭐ Total Points, EXP & Level lifetime
• 🔥 Streak mingguan
• 🏅 Achievement archive
• 🛡️ Centurion Cycles

💡 *Artinya:*
Seasonal ranking mulai dari awal, tapi progres lifetime tetap lanjut. Ini kesempatan baru untuk berburu rank tanpa kehilangan level! 🎯

📌 Manfaatkan Season %d ini sebaik mungkin. Laporkan aktivitasmu dengan /lapor!

*Semangat Season %d!* 🚀🔥`, seasonNumber, nextSeason, seasonNumber, nextSeason, seasonNumber, seasonNumber)
}
