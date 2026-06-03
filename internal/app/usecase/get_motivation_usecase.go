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

	// === Quote dari atlet & tokoh dunia tentang olahraga & tubuh bergerak ===
	"🥊 \"Aku benci setiap menit latihan, tapi aku bilang: jangan berhenti. Menderitalah sekarang dan jalani sisa hidupmu sebagai juara.\" — Muhammad Ali",
	"🏀 \"Aku gagal berkali-kali dalam hidupku, dan itulah sebabnya aku berhasil.\" — Michael Jordan",
	"⚽ \"Bakat tanpa kerja keras itu bukan apa-apa.\" — Cristiano Ronaldo",
	"⚽ \"Kamu harus berjuang untuk meraih mimpimu. Berkorban dan bekerja keras untuk itu.\" — Lionel Messi",
	"🏋️ \"Kekuatan tidak datang dari menang. Justru perjuanganmu yang membentuk kekuatanmu.\" — Arnold Schwarzenegger",
	"🥋 \"Aku tidak takut pada orang yang berlatih 10.000 tendangan sekali, tapi aku takut pada yang berlatih satu tendangan 10.000 kali.\" — Bruce Lee",
	"🏒 \"Kamu kehilangan 100% tembakan yang tidak pernah kamu coba.\" — Wayne Gretzky",
	"🏃 \"Sukses bukan kebetulan. Itu kerja keras, ketekunan, dan cinta pada apa yang kamu lakukan.\" — Pelé",
	"🏀 \"Kerja keras mengalahkan bakat ketika bakat tidak bekerja keras.\" — Kobe Bryant",
	"🎾 \"Juara bukan ditentukan oleh kemenangannya, tapi oleh caranya bangkit saat terjatuh.\" — Serena Williams",
	"⚡ \"Aku tidak memikirkan batas.\" — Usain Bolt",
	"🏈 \"Satu-satunya tempat sukses datang sebelum kerja keras hanyalah di kamus.\" — Vince Lombardi",
	"⚾ \"Sulit mengalahkan orang yang tidak pernah menyerah.\" — Babe Ruth",
	"🚶 \"Berjalan kaki adalah obat terbaik bagi manusia.\" — Hippocrates",
	"🧬 \"Jagalah tubuhmu. Itu satu-satunya tempat yang kamu punya untuk hidup.\" — Jim Rohn",
	"🏔️ \"Bukan gunung yang kita taklukkan, melainkan diri kita sendiri.\" — Edmund Hillary",
	"💪 \"Tubuh mencapai apa yang diyakini oleh pikiran.\" — Napoleon Hill",
	"🏃‍♀️ \"Kebugaran fisik adalah kunci tubuh yang sehat sekaligus dasar dari aktivitas intelektual yang dinamis dan kreatif.\" — John F. Kennedy",

	// === Ajakan bergerak & hidup sehat dari para tokoh ===
	"🤸 \"Gerakan adalah obat untuk menciptakan perubahan pada kondisi fisik, emosi, dan mental seseorang.\" — Carol Welch",
	"🏛️ \"Kurangnya aktivitas merusak kondisi baik setiap manusia, sedangkan gerakan dan olahraga teratur menjaga dan memeliharanya.\" — Plato",
	"📜 \"Hanya olahraga yang mampu menopang semangat dan menjaga pikiran tetap bugar.\" — Cicero",
	"🌅 \"Kebugaran bukan soal membangun tubuh yang lebih baik, tapi membangun hidup yang lebih baik.\" — Jillian Michaels",
	"👑 \"Olahraga adalah raja. Nutrisi adalah ratu. Satukan keduanya, dan kamu punya kerajaan.\" — Jack LaLanne",
	"🚀 \"Untuk jadi yang terbaik, kamu harus terus menantang diri. Jangan diam, melompatlah ke depan.\" — Ronda Rousey",
	"⚽ \"Pemenang adalah orang yang bangkit satu kali lebih banyak daripada saat ia terjatuh.\" — Mia Hamm",
	"🥇 \"Juara tidak dibentuk di gym. Juara dibentuk dari sesuatu di dalam diri — hasrat, mimpi, dan visi.\" — Muhammad Ali",
	"🧘 \"Kebugaran fisik adalah syarat pertama dari kebahagiaan.\" — Joseph Pilates",
	"🏺 \"Sungguh memalukan bila seseorang menjadi tua tanpa pernah melihat keindahan dan kekuatan yang mampu dicapai tubuhnya.\" — Socrates",
	"🐢 \"Tidak masalah seberapa lambat kamu melangkah, asalkan kamu tidak berhenti.\" — Confucius",
	"💎 \"Sukses bukan selalu soal kehebatan. Ia soal konsistensi. Kerja keras yang konsisten mengantarkan pada kesuksesan.\" — Dwayne Johnson",
	"🔑 \"Satu-satunya perjalanan yang mustahil adalah perjalanan yang tidak pernah kamu mulai.\" — Tony Robbins",
	"🌬️ \"Saat kamu ingin sukses sekuat kamu ingin bernapas, saat itulah kamu akan sukses.\" — Eric Thomas",
	"🎾 \"Kamu harus percaya pada dirimu sendiri bahkan ketika tidak ada orang lain yang percaya.\" — Venus Williams",
	"🩺 \"Bila kita bisa memberi setiap orang takaran gizi dan olahraga yang tepat — tak kurang, tak lebih — kita telah menemukan jalan teraman menuju sehat.\" — Hippocrates",
	"⛸️ \"Olahraga menguji kita dalam banyak hal: keterampilan, hati, dan kemampuan bangkit setelah jatuh.\" — Peggy Fleming",
	"🏀 \"Kalau aku punya masalah, setelah bermain pikiranku jadi lebih jernih. Olahraga itu seperti terapi.\" — Michael Jordan",
	"🌿 \"Hidup sehat bukan tujuan yang harus dicapai, melainkan cara untuk menjalani hidup setiap hari.\" — Anonim",
	"🔁 \"Hasil datang seiring waktu, bukan dalam semalam. Kerja keras, tetap konsisten.\" — Anonim",
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
