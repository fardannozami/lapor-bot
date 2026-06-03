package usecase

import (
	"math/rand"
	"time"
)

var motivationalQuotes = []string{
	"💪 Tubuh yang kuat dimulai dari kebiasaan kecil. Ayo bergerak hari ini!",
	"🏃‍♂️ Satu langkah kecil hari ini adalah investasi besar untuk masa depanmu!",
	"🔥 Jangan menunggu semangat datang. Mulai bergerak, dan semangat akan mengikuti!",
	"🌟 Setiap keringat adalah bukti bahwa kamu sedang bertransformasi menjadi versi terbaikmu!",
	"⏰ Waktumu terbatas, jangan sia-siakan untuk menunda. Mulai sekarang!",
	"🦾 Tidak ada yang namanya latihan yang sia-sia. Setiap reps counts!",
	"🎯 Fokus pada proses, hasil akan mengikuti. Consistency is the key!",
	"🌈 Hari ini adalah kesempatanmu untuk jadi lebih baik dari kemarin. Gaskeun!",
	"⚡ Tenagamu lebih besar dari alasanmu untuk berhenti. Buktikan!",
	"🏆 Champions are made when no one is watching. Jadilah juara di ruang latihanmu!",
	"🚀 Jangan bandingkan dirimu dengan orang lain. Kalahkan versi dirimu kemarin!",
	"🌱 Perubahan tidak terjadi dalam semalam, tapi setiap hari kamu sedang tumbuh!",
	"💥 Rasa nyamanmu adalah musuh terbesar. Keluar dari zona nyaman dan berkembang!",
	"🏋️‍♀️ Barbell tidak bohong. Usaha mu akan terbayar dengan hasil yang nyata!",
	"🎖️ Pain is temporary, but pride is forever. Tahan sedikit rasa tidak nyaman hari ini!",
	"🌞 Setiap pagi adalah kesempatan baru untuk membuat dirimu lebih kuat!",
	"🔥 Jangan cari motivasi. Jadilah motivasi untuk orang-orang di sekitarmu!",
	"🦵 Kaki yang lelah adalah tanda bahwa kamu telah berusaha keras. Bangga!",
	"💯 1% better every day. Itu saja yang kamu butuhkan untuk mencapai goals-mu!",
	"🚴‍♂️ Perjalanan seribu mil dimulai dari satu langkah. Langkahmu hari ini sudah cukup berharga!",
	"🧠 Pikiranmu akan menyerah sebelum tubuhmu. Jangan biarkan pikiran menang!",
	"⚔️ You vs You. Itu satu-satunya pertarungan yang benar-benar penting!",
	"🎊 Selesaikan latihanmu hari ini, dan besok kamu akan berterima kasih pada dirimu sendiri!",
	"🏃‍♀️ Lari dari alasanmu, bukan dari latihanmu!",
	"💪🏼 Kekuatan bukan datang dari apa yang bisa kamu lakukan. Kekuatan datang dari mengatasi apa yang kamu pikir tidak bisa kamu lakukan!",
	"🌟 Kamu tidak harus hebat untuk memulai, tapi kamu harus memulai untuk menjadi hebat!",
	"🔥 Discipline will take you where motivation won't. Latih disiplinmu hari ini!",
	"🏋️‍♂️ Jangan hitung hari-hari mu, buat setiap hari terhitung. Ayo workout!",
	"🦾 Tubuhmu bisa melakukan hampir semuanya. Yang membatasimu hanya pikiranmu!",
	"🎯 Setiap detik yang kamu investasikan di gym adalah investasi untuk masa depan yang lebih sehat!",
}

type GetMotivationUsecase struct {
	rng *rand.Rand
}

func NewGetMotivationUsecase() *GetMotivationUsecase {
	return &GetMotivationUsecase{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (uc *GetMotivationUsecase) Execute() string {
	idx := uc.rng.Intn(len(motivationalQuotes))
	return "✨ *Motivasi Harian* ✨\n\n" + motivationalQuotes[idx] + "\n\n_Ketik #motivasi untuk quote baru!_"
}
