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
	"💪 *Olahraga:* Gabungkan kardio, strength, dan mobility dalam seminggu. Biar stamina, otot, dan sendi sama-sama naik level.",
	"🦵 *Olahraga:* Pemanasan 5-10 menit sebelum latihan dan cooldown setelahnya. Cedera kecil bisa mengganggu streak panjang.",
}

var tipsMakanan = []string{
	"🥗 *Makan:* Utamakan real food: sayur, buah, protein, karbo kompleks, dan lemak sehat. Tubuh butuh nutrisi, bukan cuma kenyang.",
	"🍱 *Makan:* Pakai patokan piring seimbang: ½ sayur/buah, ¼ protein, ¼ karbo kompleks. Simpel dan gampang diingat.",
	"🥚 *Makan:* Tambahkan protein di tiap makan — telur, ikan, ayam, tempe, tahu, atau kacang. Protein bantu kenyang dan recovery.",
	"🌾 *Makan:* Pilih karbo yang lebih utuh seperti nasi, kentang, oats, ubi, atau roti gandum. Energi lebih stabil untuk aktivitas.",
	"🚫 *Makan:* Kurangi makanan ultra-olahan, gorengan berlebihan, snack tinggi gula/garam, dan fast food. Boleh sesekali, jangan jadi default.",
	"🍳 *Makan:* Jangan asal skip makan setelah latihan. Isi ulang energi dengan protein + karbo agar recovery lebih optimal.",
	"🧂 *Makan:* Cek kebiasaan bumbu tinggi garam, saus, dan minuman manis. Yang kecil-kecil sering jadi sumber kalori tersembunyi.",
	"🛒 *Makan:* Kalau belanja, prioritaskan bahan segar dulu sebelum snack kemasan. Lingkungan menentukan pilihan saat lapar.",
}

var tipsMinum = []string{
	"💧 *Minum:* Minum air putih cukup sepanjang hari. Jangan tunggu haus berat — dehidrasi bikin lemas dan fokus turun.",
	"🚰 *Minum:* Bawa botol minum agar intake air lebih konsisten. Target sederhana: urine kuning muda sebagian besar hari.",
	"🥤 *Minum:* Kurangi minuman manis, soda, dan boba harian. Kalori cair cepat masuk tapi sering tidak bikin kenyang.",
	"☕ *Minum:* Kopi boleh, tapi imbangi dengan air putih dan hindari kafein terlalu sore agar tidur tetap berkualitas.",
}

var tipsIstirahat = []string{
	"😴 *Istirahat:* Tidur 7-8 jam berkualitas. Otot tumbuh dan tubuh pulih justru saat kamu tidur.",
	"🌙 *Istirahat:* Kurangi main HP sebelum tidur. Layar bikin susah ngantuk dan ganggu kualitas tidur.",
	"🛌 *Istirahat:* Pemulihan sama pentingnya dengan latihan. Beri tubuhmu waktu untuk recovery.",
	"☕ *Istirahat:* Ambil jeda sejenak di tengah kesibukan. Istirahat singkat bikin fokus kembali segar.",
	"🕗 *Istirahat:* Tidur dan bangun di jam yang teratur. Ritme tubuh yang stabil = energi lebih stabil.",
	"🌌 *Tidur:* Buat ritual 20 menit sebelum tidur: redupkan lampu, jauhkan layar, dan tenangkan napas.",
	"⏰ *Tidur:* Kalau tidur kurang, turunkan intensitas latihan besok. Recovery buruk bukan tanda lemah — itu sinyal tubuh.",
}

var tipsStres = []string{
	"🧘 *Kelola stres:* Tarik napas dalam-dalam beberapa kali. Tenangkan pikiran, lepaskan beban sejenak.",
	"🌳 *Kelola stres:* Luangkan waktu untuk hal yang kamu suka. Pikiran yang bahagia bikin tubuh lebih sehat.",
	"💬 *Kelola stres:* Cerita ke teman atau keluarga kalau lagi penat. Kamu tidak harus memikul semuanya sendiri.",
	"📵 *Kelola stres:* Sesekali jauhkan diri dari notifikasi dan media sosial. Beri ruang untuk pikiranmu bernapas.",
	"🙏 *Kelola stres:* Syukuri hal-hal kecil hari ini. Rasa syukur adalah penawar stres yang ampuh.",
	"🧠 *Kelola stres:* Tulis 3 hal yang bikin pikiran penuh, lalu pilih 1 langkah kecil yang bisa kamu lakukan hari ini.",
	"🚶 *Kelola stres:* Jalan santai 10 menit tanpa scrolling bisa menurunkan tegang dan membantu pikiran lebih jernih.",
}

// BuildWellnessReminder returns a formatted block containing one random
// motivational opener plus a fresh tip for each health pillar: olahraga,
// makanan, minum, istirahat, and mengelola stres. It is appended to scheduled
// notifications.
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
	sb.WriteString(tipsMinum[rand.Intn(len(tipsMinum))])
	sb.WriteString("\n")
	sb.WriteString(tipsIstirahat[rand.Intn(len(tipsIstirahat))])
	sb.WriteString("\n")
	sb.WriteString(tipsStres[rand.Intn(len(tipsStres))])
	sb.WriteString("\n\nJaga olahraga, makan, tidur, dan pikiranmu. Sehat itu paket lengkap! 💚")

	return sb.String()
}
