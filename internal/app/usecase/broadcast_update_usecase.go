package usecase

type BroadcastUpdateUsecase struct{}

func NewBroadcastUpdateUsecase() *BroadcastUpdateUsecase {
	return &BroadcastUpdateUsecase{}
}

func (uc *BroadcastUpdateUsecase) Execute() string {
	return `📢 *BOT UPDATE: SEASON 1 - THE CENTURION QUEST* 🛡️

Halo para pejuang keringat! Bot telah diperbarui dengan sistem *Prestige* baru yang lebih seru:

🔥 *Sistem Centurion Cycle (Siklus)*
* Setiap kali kamu mencapai *Hari ke-100*, kamu akan mendapatkan gelar *Centurion* 🎖️.
* Hitungan hari akan di-reset ke *Hari 1*, tapi kamu mendapatkan badge permanen [S1-C2] (Cycle 2) di sebelah namamu.
* *Week Streak* kamu tetap berlanjut dan tidak akan pernah di-reset!

🏆 *Leaderboard Race*
* Leaderboard sekarang menjadi ajang balapan dinamis! 
* User yang baru naik ke Cycle berikutnya akan memulai "balapan" lagi dari bawah. Ini kesempatan bagi anggota lain untuk menyalip dan menduduki puncak klasemen sementara!

📈 *Statistik Permanen*
* Total kontribusimu tetap tercatat di sistem sebagai bukti dedikasi jangka panjangmu.

*Ayo, siapa yang akan jadi orang pertama di grup ini yang mencapai Cycle 2?* ⚡`
}
