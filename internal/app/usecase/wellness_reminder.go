package usecase

import (
	"math/rand"
	"strings"
)

// wellnessOpeners are short, varied motivational lines used to kick off the
// wellness reminder block on scheduled notifications (leaderboard & reminder).
var wellnessOpeners = []string{
	"Tubuh yang bergerak adalah tubuh yang hidup. Yuk rawat dirimu hari ini! 🌟",
	"Sehat itu bukan tujuan akhir, tapi cara kita menjalani hidup. Mulai dari sekarang! 💫",
	"Sekecil apa pun langkah hari ini, itu tetap kemenangan untuk tubuhmu. 🏃",
	"Investasi terbaik bukan di dompet, tapi di tubuh dan pikiranmu. ✨",
	"Kamu cuma punya satu tubuh. Perlakukan dia seperti juara! 🏆",
	"Konsistensi kecil hari ini menabung kesehatan untuk masa depan. 🌱",
	"Gerak sedikit hari ini jauh lebih baik daripada rebahan seharian. 💪",
	"Versi terbaik dirimu sedang menunggu di balik kebiasaan sehat. Gas! 🔥",
}

// Each wellness pillar holds several variations so the reminder feels fresh
// every time it is sent.

var tipsOlahraga = []string{
	"🏃 *Olahraga:* Sempatkan minimal 30 menit gerak hari ini — jalan kaki, lari, atau workout. Tubuhmu akan berterima kasih!",
	"🏋️ *Olahraga:* Tidak harus ke gym. Push-up, squat, atau stretching di rumah pun sudah hitung. Yang penting bergerak!",
	"🚶 *Olahraga:* Jangan duduk terlalu lama. Bangun, regangkan badan, dan jalan kaki sebentar tiap jam.",
	"🧘 *Olahraga:* Mulai dari yang ringan kalau sudah lama vakum. Konsistensi lebih penting daripada intensitas.",
	"🚴 *Olahraga:* Pilih aktivitas yang kamu nikmati supaya bertahan lama. Sepeda, renang, badminton — bebas, asal gerak!",
}

var tipsMakanan = []string{
	"🥗 *Makanan:* Perbanyak sayur, buah, dan protein. Kurangi gorengan dan minuman manis ya!",
	"💧 *Makanan:* Jangan lupa minum air putih cukup, minimal 8 gelas sehari. Dehidrasi bikin lemas.",
	"🍚 *Makanan:* Makan secukupnya, jangan berlebihan. Porsi seimbang = energi seimbang.",
	"🍳 *Makanan:* Sarapan itu penting buat bahan bakar harimu. Jangan dilewatkan!",
	"🥦 *Makanan:* Isi setengah piringmu dengan sayur dan buah. Tubuhmu butuh nutrisi, bukan cuma kenyang.",
}

var tipsIstirahat = []string{
	"😴 *Istirahat:* Tidur 7-8 jam berkualitas. Otot tumbuh dan tubuh pulih justru saat kamu tidur.",
	"🌙 *Istirahat:* Kurangi main HP sebelum tidur. Layar bikin susah ngantuk dan ganggu kualitas tidur.",
	"🛌 *Istirahat:* Pemulihan sama pentingnya dengan latihan. Beri tubuhmu waktu untuk recovery.",
	"☕ *Istirahat:* Ambil jeda sejenak di tengah kesibukan. Istirahat singkat bikin fokus kembali segar.",
	"🕗 *Istirahat:* Tidur dan bangun di jam yang teratur. Ritme tubuh yang stabil = energi lebih stabil.",
}

var tipsStres = []string{
	"🧘 *Kelola stres:* Tarik napas dalam-dalam beberapa kali. Tenangkan pikiran, lepaskan beban sejenak.",
	"🌳 *Kelola stres:* Luangkan waktu untuk hal yang kamu suka. Pikiran yang bahagia bikin tubuh lebih sehat.",
	"💬 *Kelola stres:* Cerita ke teman atau keluarga kalau lagi penat. Kamu tidak harus memikul semuanya sendiri.",
	"📵 *Kelola stres:* Sesekali jauhkan diri dari notifikasi dan media sosial. Beri ruang untuk pikiranmu bernapas.",
	"🙏 *Kelola stres:* Syukuri hal-hal kecil hari ini. Rasa syukur adalah penawar stres yang ampuh.",
}

// BuildWellnessReminder returns a formatted block containing one random
// motivational opener plus a fresh tip for each of the four health pillars:
// olahraga (exercise), makanan (food), istirahat (rest), and mengelola stres
// (stress management). It is appended to scheduled notifications.
func BuildWellnessReminder() string {
	var sb strings.Builder

	sb.WriteString("\n\n━━━━━━━━━━━━━━━\n")
	sb.WriteString("💡 *Pengingat Sehat Hari Ini*\n\n")
	sb.WriteString(wellnessOpeners[rand.Intn(len(wellnessOpeners))])
	sb.WriteString("\n\n")
	sb.WriteString(tipsOlahraga[rand.Intn(len(tipsOlahraga))])
	sb.WriteString("\n")
	sb.WriteString(tipsMakanan[rand.Intn(len(tipsMakanan))])
	sb.WriteString("\n")
	sb.WriteString(tipsIstirahat[rand.Intn(len(tipsIstirahat))])
	sb.WriteString("\n")
	sb.WriteString(tipsStres[rand.Intn(len(tipsStres))])
	sb.WriteString("\n\nJaga olahraga, makan, tidur, dan pikiranmu. Sehat itu paket lengkap! 💚")

	return sb.String()
}
