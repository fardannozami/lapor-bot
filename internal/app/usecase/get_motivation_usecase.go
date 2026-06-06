package usecase

import (
	"math/rand/v2"
	"strings"
	"sync"
)

// recentWindow controls how many recently-shown quotes are remembered so the
// random picker can avoid repeating the same quote within a short window.
// With 300 quotes and a window of 30, the chance of an immediate repeat drops
// from 1/300 to <10% even under heavy notification traffic.
const recentWindow = 30

var (
	motivationalQuotes = []string{
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
	"🫀 \"Mereka yang tidak menyediakan waktu untuk berolahraga pada akhirnya harus menyediakan waktu untuk sakit.\" — Edward Stanley",
	"🏃 \"Jika olahraga bisa dikemas dalam pil, itu akan menjadi obat paling banyak diresepkan di dunia.\" — Robert Butler",
	"🌤️ \"Bergeraklah bukan karena membenci tubuhmu, tapi karena kamu menghargai hidup yang ditopangnya.\" — Anonim",
	"🧱 \"Latihan hari ini adalah suara kecil yang berkata: aku sedang membangun diriku, bukan menghukum diriku.\" — Anonim",
	"🥗 \"Kesehatan bukan segalanya, tapi tanpa kesehatan segalanya terasa lebih berat.\" — Arthur Schopenhauer",
	"🧭 \"Tubuh yang aktif menolong pikiran tetap jernih. Saat ragu, mulai dari berjalan kaki.\" — Anonim",

	// === Tambahan: Semangat bergerak & olahraga singkat ===
	"🏃‍♂️ Gerak hari ini, bebas dari rasa sakit esok hari.",
	"💪 Tubuhmu adalah satu-satunya tempat kamu tinggal seumur hidup. Rawat dia!",
	"🔥 Tidak ada kata terlambat untuk mulai olahraga. Mulailah sekarang!",
	"🌅 Bangun, bergerak, syukuri — rutinitas kecil yang mengubah hidup.",
	"🏋️ Latihan bukan beban, tapi hadiah yang kamu berikan untuk tubuhmu sendiri.",
	"⚡ Energi tidak datang dari motivasi, tapi dari aksi. Bergeraklah dulu!",
	"🚶 10.000 langkah terasa jauh, tapi kamu hanya perlu memulainya dengan satu langkah.",
	"🥗 Kamu tidak bisa latihan mengisi bensin di pompa yang kosong. Makan & istirahat dulu!",
	"💦 Keringat adalah air mata lemak yang menangis bahagia.",
	"⏱️ Butuh waktu 21 hari untuk membentuk kebiasaan. Hari ini hari ke-1-mu!",
	"🧠 Tubuh yang bugar menopang pikiran yang jernih dan hati yang gembira.",
	"🎽 Jangan tanya apa yang bisa kamu dapatkan dari olahraga. Tanya apa yang bisa kamu berikan untuk tubuhmu.",
	"🦋 Metamorfosis butuh waktu. Prosesnya tidak nyaman, tapi hasilnya menakjubkan.",
	"🏔️ Setiap kali kamu merasa ingin menyerah, ingatlah kenapa kamu memulai.",
	"⛰️ Puncak gunung tidak dicapai dengan cara pintas. Sama seperti tubuh idealmu.",
	"🌊 Aliran sungai mengukir batu hanya dengan konsistensi. Kamu lebih kuat dari batu.",
	"🔥 Rasa sakit hari ini adalah kekuatan yang akan kamu syukuri besok.",
	"🪴 Tanam investasi di tubuh sehat hari ini, panen vitalitas di masa tua nanti.",
	"🎯 Targetnya bukan tubuh sempurna. Targetnya tubuh yang kuat, sehat, dan bisa diandalkan.",
	"🌤️ Jangan tunggu hari Senin, bulan baru, atau tahun baru. Hari ini adalah hari terbaik untuk mulai.",
	"🏃‍♀️ Kamu tidak harus cepat. Kamu hanya harus tidak berhenti.",
	"💎 Disiplin di gym adalah cermin disiplin di kehidupan. Bentuk tubuhmu, bentuk karaktermu.",
	"🎵 Buat olahragamu jadi bagian dari ritme hidup. Bukan paksaan, tapi kebutuhan.",
	"📆 Latihan 1 jam = 4% harimu. 4% untuk tubuhmu, 96% untuk semuanya yang lain.",
	"🛠️ Tubuhmu adalah kendaraanmu. Servis dia secara rutin dengan olahraga dan nutrisi.",

	// === Tambahan: Kata bijak & quotes tokoh terkenal tentang tubuh & kesehatan ===
	"🏛️ \"Penyakit tidak datang tiba-tiba; ia muncul dari kebiasaan buruk yang berkepanjangan.\" — Hippocrates",
	"🌿 \"Tubuh manusia adalah taman, dan gerak adalah matahari, air, dan pupuknya.\" — Anonim",
	"🍎 \"Biarkan makanan menjadi obatmu dan obat menjadi makananmu.\" — Hippocrates",
	"💧 \"Ribuan orang bertahan tanpa cinta, tapi tidak ada yang bertahan tanpa air.\" — W.H. Auden",
	"🩺 \"Kesehatan terbaik adalah rasa sakit yang tidak kamu rasakan.\" — Anonim",
	"🌅 \"Tubuh yang kuat membuat pikiran yang kuat. Beri dia latihan, istirahat, dan nutrisi.\" — Anonim",
	"💪 \"Kekuatanmu tidak diukur dari apa yang bisa kamu angkat, tapi dari berapa kali kamu bangkit.\" — Anonim",
	"⚔️ \"Otot tidak dibangun di zona nyaman. Karakter pun demikian.\" — Anonim",
	"🧠 \"Jaga tubuhmu agar pikiranmu tetap utuh.\" — Anonim",
	"🌬️ \"Bernyawa, bergerak, dan bersyukur — itu tiga anugerah yang sering terlupakan.\" — Anonim",
	"🦴 \"Tulang yang kuat tidak terbentuk dari kenyamanan. Ia terbentuk dari tekanan.\" — Anonim",
	"💗 \"Sayangi jantungmu. Ia berdetak tanpa minta izin, jangan sia-siakan kerjanya.\" — Anonim",
	"🛁 \"Setelah seharian bekerja keras, tubuhmu butuh dipulihkan — bukan dihukum lagi.\" — Anonim",
	"🧘 \"Kebugaran sejati adalah ketenangan pikiran yang didukung oleh tubuh yang aktif.\" — Anonim",
	"📖 \"Tubuh sehat adalah gudang harta yang paling berharga.\" — Pepatah Arab",
	"🌳 \"Merawat tubuh adalah ibadah, karena tubuh adalah amanah yang akan dimintai pertanggungjawaban.\" — Pepatah",
	"🙏 \"Syukuri tubuh yang masih bisa bergerak. Banyak yang rela tukar apa pun untuk bisa melangkah lagi.\" — Anonim",
	"🛤️ \"Hidup yang panjang dan sehat adalah perjalanan, bukan tujuan akhir. Nikmati setiap langkahnya.\" — Anonim",
	"🌟 \"Jadilah orang yang di usia 70 masih bisa naik tangga tanpa ngos-ngosan. Itu dimulai dari hari ini.\" — Anonim",
	"⏳ \"Waktu yang kamu investasikan untuk kesehatan hari ini adalah waktu yang kamu hemat dari rumah sakit di masa depan.\" — Anonim",
	"🔄 \"Tubuh memperbarui dirinya setiap beberapa tahun. Sel-sel baru, otot baru, kesempatan baru.\" — Anonim",
	"💤 \"Tidur cukup bukan kemalasan; itu adalah perbaikan tubuh yang paling underrated.\" — Matthew Walker",
	"🌅 \"Pagi yang baik dimulai dari tidur yang cukup, sarapan yang nyata, dan niat untuk bergerak.\" — Anonim",
	"🩸 \"Kebugaran bukan tentang terlihat bagus. Ini tentang terasa bagus dari dalam ke luar.\" — Anonim",
	"🥦 \"Makan dengan pikiran. Bukan dengan mata, emosi, atau kebiasaan.\" — Anonim",
	"🧂 \"Gula, garam, dan lemak berlebihan: trio pembunuh diam-diam yang sering kita rayakan.\" — Anonim",
	"🍽️ \"Kamu bukan apa yang kamu makan. Kamu adalah hasil dari apa yang kamu lakukan secara konsisten.\" — Anonim",
	"🏃 \"Berlari bukan soal kecepatan. Ini soal keberanian untuk terus melangkah saat tubuh ingin berhenti.\" — Anonim",
	"🚴 \"Bersepeda mengajarkan keseimbangan. Dalam hidup, kamu harus terus mengayuh untuk tidak jatuh.\" — Anonim",
	"🏊 \"Berenang adalah meditasi bergerak. Tubuh lelah, pikiran jernih.\" — Anonim",
	"🥊 \"Setiap pukulan di karung pasir adalah pelepasan beban yang tidak perlu kamu simpan.\" — Anonim",
	"🧗 \"Mendaki gunung adalah refleksi kehidupan: capek, jatuh, bangkit, dan terus naik.\" — Anonim",
	"🧘‍♀️ \"Yoga bukan cuma soal lentur. Ini soal mengenal setiap inci tubuhmu dengan sadar.\" — Anonim",
	"🤸 \"Stretching adalah percakapan jujur dengan tubuh. Dengarkan apa yang dia katakan.\" — Anonim",
	"🏐 \"Olahraga bukan paksaan sosial. Ia adalah bentuk rasa syukur karena masih diberi tubuh yang berfungsi.\" — Anonim",

	// === Tambahan: Atlet dunia & tokoh terkenal (putaran kedua) ===
	"🥇 \"Saya tidak pernah bermimpi tentang sukses. Saya bekerja untuk itu.\" — Estée Lauder",
	"⛹️ \"Kamu bisa melakukan apa pun yang kamu putuskan untuk dilakukan.\" — Michael Jordan",
	"🏈 \"Coba, coba, dan coba lagi sampai kamu berhasil.\" — Tom Brady",
	"⚽ \"Kesempatan tidak datang dengan sendirinya. Kamu yang menciptakannya.\" — Chris Grosser",
	"🎾 \"Saya tidak khawatir dengan apa yang akan saya lakukan besok. Saya hanya fokus pada hari ini.\" — Novak Djokovic",
	"🏃 \"Saya tidak berlatih hanya untuk menjadi yang terbaik. Saya berlatih karena saya mencintai prosesnya.\" — Eliud Kipchoge",
	"🥊 \"Seorang juara bukan tentang berapa banyak dia menang, tapi seberapa dia bangkit dari kekalahan.\" — Lennox Lewis",
	"🏋️ \"Jangan takut gagal. Takutlah tidak pernah mencoba.\" — Michael Jordan",
	"🏆 \"Tidak ada jalan pintas untuk tempat di mana puncaknya.\" — Abraham Lincoln",
	"🎖️ \"Kesuksesan adalah jumlah dari usaha kecil yang diulang hari demi hari.\" — Robert Collier",
	"🏅 \"Saya pikir segala sesuatu mungkin dilakukan jika Anda memiliki cukup tekad.\" — Usain Bolt",
	"🚴 \"Jika kamu ingin sukses, kamu harus siap menerima rasa sakit.\" — Lance Armstrong",
	"🏃‍♀️ \"Lari itu cuma soal menempatkan satu kaki di depan kaki yang lain. Simpel, bukan?\" — Anonim",
	"🏌️ \"Golf adalah olahraga pikiran. Tubuh hanya mengikuti.\" — Anonim",
	"🏈 \"Pemain terbaik adalah mereka yang bermain bukan karena mereka harus, tapi karena mereka mau.\" — Anonim",

	// === Tambahan: Mindset & kebiasaan sehat ===
	"🌅 \"Mulailah hari ini dengan satu keputusan sehat. Besok akan lebih mudah dari yang kamu kira.\" — Anonim",
	"🔁 \"Konsistensi mengalahkan intensitas. Lebih baik 10 menit setiap hari daripada 2 jam sekali seminggu.\" — Anonim",
	"📈 \"Kamu tidak akan melihat hasil hari pertama, minggu pertama, atau bahkan bulan pertama. Tapi di tahun pertama, semua orang akan melihat perubahannya.\" — Anonim",
	"🧭 \"Disiplin adalah melakukan apa yang harus dilakukan, meskipun kamu tidak ingin melakukannya.\" — Anonim",
	"🌟 \"Kebiasaan menentukan karakter. Karakter menentukan nasib. Mulailah dari satu kebiasaan sehat.\" — Anonim",
	"🛌 \"Tidur cukup, makan benar, bergerak cukup, stres minim. Itu saja. Itu sudah cukup.\" — Anonim",
	"🧠 \"Pikiran yang sehat tidak datang dari membaca buku motivasi. Ia datang dari tubuh yang sehat.\" — Anonim",
	"💆 \"Kebugaran adalah investasi paling tenang dan paling menguntungkan yang pernah ada.\" — Anonim",
	"🏅 \"Jangan biarkan tubuhmu menjadi hambatan di usia tua. Rawat dia dari sekarang.\" — Anonim",
	"🪑 \"Duduk terlalu lama itu seperti merokok pelan-pelan. Berdiri, regangkan badanmu.\" — Anonim",
	"📵 \"Letakkan HP-mu, matikan Netflix-mu, dan jalan kaki 30 menit. Itu saja sudah cukup mengubah harimu.\" — Anonim",
	"🥤 \"Beda antara soda dan air putih: yang satu memberi kalori kosong, yang satu memberi kehidupan.\" — Anonim",
	"🌡️ \"Tubuhmu bicara lewat sinyal. Jangan abaikan. Dengarkan lelah, lapar, dan rasa sakitnya.\" — Anonim",
	"🫀 \"Jantung yang sehat memompa lebih dari 100.000 kali sehari. Apa yang kamu lakukan untuknya hari ini?\" — Anonim",
	"🛡️ \"Sistem imun tidak dibangun dalam sehari. Ia dibangun dari tidur, makan, gerak, dan pikiran yang tenang.\" — Anonim",
	"🌤️ \"Sinar matahari pagi selama 10 menit: gratis, menyehatkan, dan memperbaiki mood. Kenapa tidak?\" — Anonim",
	"🚿 \"Setelah latihan, rasa lelah terasa mahal dan berharga. Bangga dengan setiap tetes keringat.\" — Anonim",
	"💪 \"Otot tidak terbentuk dari mengangkat beban ringan. Ia terbentuk dari mengangkat beban yang membuatmu ragu.\" — Anonim",
	"🪞 \"Cermin tidak berbohong. Tapi angka di timbangan juga tidak menceritakan seluruh cerita.\" — Anonim",
	"👖 \"Celana yang dulu sempit bukan musuhmu. Ia adalah pengingat betapa jauh kamu sudah berjalan.\" — Anonim",

	// === Round 3: Atlet dunia & mindset juara (1) ===
	"🥊 \"Saya tidak menghitung repetisi saya. Saya hanya mulai dan tidak berhenti.\" — Jackie Chan",
	"🏆 \"Hidup ini bukan soal seberapa keras pukulannya. Soal seberapa kuat kau bisa bertahan.\" — Rocky Balboa (Sylvester Stallone)",
	"🎾 \"Untuk jadi nomor satu, kau harus berlatih seperti kau bukan siapa-siapa.\" — Maria Sharapova",
	"🏊 \"Pelajaran dari air: kalau kau melawan, kau kalah. Kalau kau mengalir, kau menang.\" — Anonim",
	"🥋 \"Kemenangan terbesar bukan tidak pernah jatuh, tapi bangkit setiap kali kau jatuh.\" — Nelson Mandela",
	"🏋️ \"Otot adalah organ terbesar yang bisa kau bentuk. Bentuk dia setiap hari dengan benar.\" — Mark Rippetoe",
	"🏃 \"Lari pagi bukan untuk orang yang kuat. Lari pagi yang membentuk orang yang kuat.\" — Anonim",
	"⛹️ \"Kau tidak bermain dengan lawan. Kau bermain dengan dirimu sendiri.\" — Anonim",
	"🏒 \"Bakat yang tidak diasah sama dengan kayu bakar yang tidak dinyalakan.\" — Anonim",
	"🏈 \"Keringat di latihan adalah perak di pertandingan.\" — Vince Lombardi",
	"🥇 \"Saya tidak pernah melihat ke belakang. Hanya ada satu arah — ke depan.\" — Usain Bolt",
	"⚽ \"Lebih baik jadi yang terbaik di levelmu sendiri daripada yang terburuk di level tertinggi.\" — Pep Guardiola",
	"🏀 \"Kau tidak bisa mengubah masa lalu, tapi kau bisa mengubah masa depan lewat usaha hari ini.\" — Stephen Curry",
	"🎾 \"Kemenangan terbesar adalah bangkit dari titik terendahmu.\" — Naomi Osaka",
	"🏆 \"Saya datang ke gym bukan untuk jadi orang lain. Saya datang untuk jadi versi terkuat dari diri saya.\" — Anonim",
	"⚽ \"Kerja keras adalah bakat yang sebenarnya.\" — Jürgen Klopp",
	"🏈 \"Kau akan lebih sering gagal daripada berhasil. Itu bukan alasan berhenti.\" — Anonim",
	"🥋 \"Kemenangan bukan segalanya, tapi usaha untuk menang adalah segalanya.\" — Vince Lombardi",
	"🏀 \"Batasan ada di pikiran, bukan di tubuh.\" — Henry Ford",
	"🏃 \"Saya berlari bukan karena saya harus. Saya berlari karena saya bisa, dan karena itu hadiah.\" — Anonim",
	"🏋️ \"Angkat beban bukan untuk tubuhmu. Angkat beban untuk mentalmu.\" — Anonim",
	"⚾ \"Bakat itu umum. Kerja keras itu langka. Itu yang memisahkan yang hebat dari yang biasa.\" — Anonim",
	"🏈 \"Sakitnya hari ini adalah kekuatanmu besok. Ingat itu setiap kali ingin menyerah.\" — Anonim",
	"🥊 \"Satu-satunya knockout yang perlu dikhawatirkan adalah mengalahkan versi lama dirimu.\" — Anonim",
	"🎾 \"Kau tidak harus sempurna untuk memulai. Tapi kau harus memulai untuk jadi lebih baik.\" — Arthur Ashe",
	"⚽ \"Latihan tidak pernah sia-sia. Bahkan saat merasa tidak berkembang, tubuh sedang beradaptasi.\" — Anonim",
	"🏃 \"Berlari adalah terapi paling jujur. Tidak ada topeng, tidak ada alasan, hanya kau dan langkahmu.\" — Anonim",
	"🏀 \"Setiap pagi adalah kesempatan menambah satu lembar cerita yang belum pernah ditulis orang lain.\" — Anonim",
	"🏆 \"Orang sukses bukan yang tidak pernah jatuh. Mereka yang selalu bangun dengan satu tekad: coba lagi.\" — Anonim",
	"🥋 \"Saya tidak takut pada orang yang berlatih 10.000 tendangan sekali. Saya takut pada yang berlatih satu tendangan 10.000 kali.\" — Bruce Lee",

	// === Round 3: Mindset & growth (2) ===
	"🧠 \"Pertumbuhan terjadi di luar zona nyaman. Kalau kau betah, kau tidak berkembang.\" — Anonim",
	"🌱 \"Kau tidak akan melihat hasil hari ini. Tapi suatu hari nanti, kau akan berterima kasih pada dirimu yang tidak menyerah.\" — Anonim",
	"🎯 \"Tujuan bukan hanya untuk dicapai. Tujuan adalah kompas yang mengarahkan setiap langkah.\" — Anonim",
	"🔄 \"Versi terbaik dirimu bukan yang sempurna. Versi terbaikmu adalah yang terus mencoba.\" — Anonim",
	"📈 \"1% lebih baik setiap hari terdengar kecil. Dalam setahun, kau 37 kali lebih baik.\" — James Clear",
	"🌟 \"Kau adalah produk dari kebiasaanmu. Ubah satu kebiasaan, ubah hidupmu.\" — James Clear",
	"🪞 \"Jangan tanya apa yang bisa tubuhmu lakukan untukmu. Tanya apa yang sudah kau lakukan untuk tubuhmu.\" — Anonim",
	"🧭 \"Kau tidak harus melihat seluruh tangga. Ambil satu langkah pertama.\" — Martin Luther King Jr.",
	"💎 \"Karakter dibangun dari hari-hari biasa saat kau memilih untuk tidak menyerah.\" — Anonim",
	"🪴 \"Seperti pohon yang butuh waktu untuk tumbuh, tubuhmu juga butuh waktu. Sabar, konsisten, ulangi.\" — Anonim",
	"🌅 \"Mulai dari yang kecil, mulai dari sekarang. Mulai dari yang kau punya.\" — Anonim",
	"🔁 \"Motivasi datang dan pergi. Rutinitas tetap. Bangun rutinitas, bukan ketergantungan mood.\" — Anonim",
	"🧠 \"Pikiranmu adalah otot pertama yang perlu kau latih. Tubuh akan mengikuti.\" — Anonim",
	"⏳ \"Hasil bukan hadiah instan. Hasil adalah tagihan yang jatuh tempo untuk kerja keras yang konsisten.\" — Anonim",
	"🌳 \"Akar yang kuat butuh waktu. Jangan cabut pohon hanya karena belum berbuah.\" — Anonim",
	"🎢 \"Hidup ini maraton, bukan sprint. Pace yourself.\" — Anonim",
	"🪴 \"Konsistensi kecil setiap hari mengalahkan sesi heroik sekali seminggu.\" — Anonim",
	"🔑 \"Kebiasaan baik tidak dibangun dengan motivasi. Mereka dibangun dengan sistem.\" — James Clear",
	"🧠 \"Mindset yang bertumbuh melihat tantangan sebagai peluang, bukan ancaman.\" — Carol Dweck",
	"🚶 \"Kau tidak harus sempurna untuk jadi lebih baik. Kau hanya harus lebih baik dari kemarin.\" — Anonim",
	"💡 \"Kemauan untuk berubah adalah langkah pertama. Langkah kedua adalah bertahan saat rasanya ingin berhenti.\" — Anonim",
	"🌿 \"Setiap hari adalah halaman kosong. Tulis cerita yang membuatmu bangga.\" — Anonim",
	"🎯 \"Fokus pada langkah, bukan pada gunung. Gunung akan terurai dengan setiap langkah.\" — Anonim",
	"🏗️ \"Karakter dibangun pelan-pelan, di saat tidak ada yang melihat.\" — Anonim",
	"🪞 \"Perubahan kecil yang konsisten adalah satu-satunya yang bertahan dalam jangka panjang.\" — Anonim",

	// === Round 3: Recovery, tidur, nutrisi (3) ===
	"💤 \"Tidur adalah suplemen paling kuat, gratis, dan legal di dunia.\" — Anonim",
	"🥦 \"Protein adalah bahan bakar otot. Sayur adalah vitamin pemulihan. Air adalah pelumas sendi.\" — Anonim",
	"🛌 \"Kau tidak tumbuh saat latihan. Kau tumbuh saat tidur setelah latihan.\" — Anonim",
	"🍌 \"Makan setelah workout dalam 60 menit adalah kasih nutrisi yang tubuhmu minta.\" — Anonim",
	"💧 \"Dehidrasi ringan saja bisa turunkan performa 25%. Minum air, jangan tunggu haus.\" — Anonim",
	"🍠 \"Karbo kompleks = energi stabil. Karbo sederhana = lonjakan singkat lalu jatuh.\" — Anonim",
	"🥑 \"Lemak sehat dari alpukat, kacang, dan ikan = bahan bakar otak & hormon.\" — Anonim",
	"🍳 \"Sarapan dengan protein cukup = tidak lapar 2 jam kemudian, tidak ngemil sembarangan.\" — Anonim",
	"🥗 \"Warna di piring = variasi nutrisi. Makin berwarna, makin lengkap.\" — Anonim",
	"🧊 \"Mandi air dingin setelah latihan bukan siksaan. Itu pemulihan yang sangat efektif.\" — Wim Hof",
	"🛁 \"Foam rolling dan stretching bukan kemewahan. Itu perawatan tubuh dasar.\" — Anonim",
	"🧘 \"Active recovery seperti jalan kaki atau yoga meningkatkan aliran darah tanpa menambah kelelahan.\" — Anonim",
	"💆 \"Massage bukan hanya untuk atlet. Itu investasi pemeliharaan tubuh yang sering diabaikan.\" — Anonim",
	"🛀 \"Mandi air panas setelah latihan berat = rilekskan otot, bantu tidur lebih nyenyak.\" — Anonim",
	"🌙 \"Matikan layar 30 menit sebelum tidur. Tidurmu = recoverymu = performa esok.\" — Anonim",
	"🛏️ \"Tidur 7-9 jam bukan kemalasan. Itu upgrade performa gratis yang tidak ada suplemen yang bisa menyaingi.\" — Matthew Walker",
	"☕ \"Kafein setelah jam 2 siang = tidur tidak nyenyak = bangun capek = performa turun. Pilih.\" — Anonim",
	"🧠 \"Ruangan gelap + dingin = melatonin naik = tidur optimal.\" — Matthew Walker",
	"🧂 \"Elektrolit hilang saat keringat. Isi ulang dengan air mineral, bukan cuma air biasa.\" — Anonim",
	"🍫 \"Cokelat hitam 70%+ = antioksidan tinggi + mood boost tanpa gula berlebihan.\" — Anonim",
	"🥒 \"5-7 porsi sayur & buah per hari. Bukan diet, itu standar minimal tubuhmu.\" — Anonim",
	"🐟 \"Omega-3 dari ikan = antiinflamasi alami = sendi & jantung lebih sehat jangka panjang.\" — Anonim",
	"🌿 \"Antioksidan dari teh hijau, berry, dan dark chocolate = perlindungan sel dari kerusakan.\" — Anonim",
	"🦴 \"Kalsium + vitamin D = tulang kuat. Berjemur pagi 15 menit cukup untuk D harian.\" — Anonim",
	"🍵 \"Minum air lemon di pagi hari = hidrasi + sedikit vitamin C + rutinitas sehat.\" — Anonim",

	// === Round 3: Micro-habits & konsistensi (4) ===
	"🪥 \"Sikat gigi dua kali sehari sudah jadi kebiasaan. Gym tiga kali seminggu? Juga bisa, dengan repetisi yang sama.\" — Anonim",
	"☀️ \"Setelah bangun, langsung pakai sepatu lari. Jangan pikirkan. Pakai dan keluar.\" — Anonim",
	"📅 \"Jadwalkan olahraga seperti jadwal meeting. Non-negotiable.\" — Anonim",
	"⏰ \"Alarm pagi 30 menit lebih awal = 30 menit gym, leluasa.\" — Anonim",
	"🛏️ \"Taruh matras yoga di samping tempat tidur. Kalau kau melihatnya, kau ingat.\" — Anonim",
	"📱 \"Letakkan HP di luar kamar tidur. 10 menit pertama pagi = milikmu, bukan notifikasi.\" — Anonim",
	"🧦 \"Siapkan baju olahraga malam sebelumnya. Kurangi friksi keputusan = lebih konsisten.\" — Anonim",
	"🏠 \"Latihan 10 menit di rumah lebih baik daripada tidak ada latihan sama sekali.\" — Anonim",
	"📓 \"Tulis 'Workout: ✅' di jurnal setiap hari. Streak tulisan = streak disiplin.\" — Anonim",
	"🎒 \"Bawa tas gym di mobil. Kalau sudah siap, keputusan jadi lebih mudah.\" — Anonim",
	"🍎 \"Cuci & potong buah setelah belanja. Camilan sehat = pilihan default.\" — Anonim",
	"🚶 \"Parkir jauh dari pintu masuk kantor. 300 langkah tambahan per hari = 22 km per tahun.\" — Anonim",
	"🪜 \"Naik tangga, bukan lift. Setiap anak tangga = kalori yang tidak terasa.\" — Anonim",
	"📞 \"Lakukan walking meeting. Produktivitas + gerakan.\" — Anonim",
	"🎧 \"Buat playlist khusus olahraga. Trigger musik = trigger aksi.\" — Anonim",
	"💧 \"Botol air 1L di meja kerja. Refill 2-3x sehari = target minum tercapai.\" — Anonim",
	"🌙 \"Ritual malam 30 menit: matikan layar, redupkan lampu, baca buku, tidur.\" — Anonim",
	"🧘 \"5 menit meditasi pagi sebelum cek HP = pikiran lebih jernih sepanjang hari.\" — Anonim",
	"📋 \"Malam sebelumnya, tulis 3 hal yang akan kau lakukan besok. Besok pagi = tinggal eksekusi.\" — Anonim",
	"⏱️ \"Pakai timer 25 menit untuk stretching, plank, atau bodyweight circuit di rumah.\" — Anonim",

	// === Round 3: Mental toughness & resilience (5) ===
	"🛡️ \"Kau lebih tangguh dari yang kau kira. Ujian terberat hidupmu belum membuatmu menyerah.\" — Anonim",
	"⚡ \"Ketahanan bukan tidak pernah jatuh. Ketahanan adalah bangkit setiap kali jatuh, lebih cepat dari sebelumnya.\" — Anonim",
	"🌊 \"Badai akan datang. Pertanyaannya bukan 'apakah', tapi 'kapan'. Siapkan dirimu, dan nikmati perjalanannya.\" — Anonim",
	"🏔️ \"Pendaki gunung yang pernah ke puncak tahu: rasa takut tetap ada. Yang membedakan adalah mereka tetap melangkah.\" — Anonim",
	"🧊 \"Masuk ke air dingin. Tubuhmu akan menjerit 'keluar'. Bertahan 30 detik. Selamat, kau sudah lebih kuat dari 10 menit lalu.\" — Wim Hof",
	"🪨 \"Kapasitasmu menanggung beban jauh lebih besar dari beban yang kau kira tidak bisa kau tanggung.\" — David Goggins",
	"🐺 \"Suara di kepalamu yang bilang 'cukup' adalah mental barrier. Tubuhmu punya 30% cadangan. Pakai.\" — David Goggins",
	"🔥 \"Disiplin adalah melakukan apa yang harus kau lakukan, bukan apa yang ingin kau lakukan.\" — Jocko Willink",
	"🌟 \"Rasa sakit itu sementara. Kalau kau berhenti, rasa sakit itu sia-sia. Kalau kau terus, itu jadi trofi.\" — David Goggins",
	"💎 \"Kau tidak harus termotivasi. Kau hanya harus disiplin. Disiplin mengalahkan mood setiap hari.\" — Jocko Willink",
	"🪖 \"Saat semuanya tidak berjalan sesuai rencana, tetap tenang dan lanjut.\" — Jocko Willink",
	"🌪️ \"Tenang bukan berarti lemah. Tenang adalah kekuatan yang dikontrol.\" — Anonim",
	"🗻 \"Gunung yang ada di depanmu mungkin lebih kecil dari yang kau bayangkan. Atau lebih besar. Either way, kau akan tahu setelah kau mulai.\" — Anonim",
	"🦅 \"Berani bukan berarti tidak takut. Berani adalah bertindak meskipun takut.\" — Anonim",
	"🪖 \"Kalau tidak ada yang percaya padamu, jangan apa-apain. Justru jadikan itu bensin.\" — David Goggins",

	// === Round 3: Buku terkenal & stoik (6) ===
	"📚 \"Kau bukan produk dari apa yang telah terjadi padamu. Kau adalah produk dari apa yang kau pilih untuk dilakukan selanjutnya.\" — Viktor Frankl",
	"📚 \"Kebiasaan adalah berat di pagi hari, tapi makin ringan seiring waktu. Tangga yang kau panjat setiap hari bukan beban, tapi jalan.\" — James Clear",
	"📚 \"Setiap hari adalah kesempatan untuk menjadi lebih kuat, lebih sabar, lebih terampil, lebih murah hati.\" — Ryan Holiday",
	"📚 \"Hambatan bukan yang menghentikanmu. Hambatan adalah jalannya.\" — Ryan Holiday",
	"📚 \"Berani bukan berarti tidak punya rasa takut. Berani adalah mengakui rasa takut tapi tetap maju.\" — Brené Brown",
	"📚 \"Sistem mengalahkan tujuan. Bangun sistem, capai tujuan secara otomatis.\" — James Clear",
	"📚 \"Grit adalah bakat + ketekunan. Ketekunan jauh lebih penting dari bakat untuk sukses jangka panjang.\" — Angela Duckworth",
	"📚 \"Rasa sakit itu nyata. Tapi kau juga nyata. Dan kau lebih kuat dari rasa sakitnya.\" — David Goggins",
	"📚 \"Cara terbaik memprediksi masa depan adalah menciptakannya.\" — Peter Drucker",
	"📚 \"Apa yang kau pikirkan berulang-ulang akan menjadi norma barumu.\" — James Clear",
	"📚 \"Mulai dari tempatmu berada. Gunakan apa yang kau punya. Lakukan apa yang kau bisa.\" — Arthur Ashe",
	"📚 \"Konsistensi kecil yang tampak sepele adalah dasar dari transformasi yang tampak mustahil.\" — James Clear",
	"📚 \"Tidur itu bukan kemewahan. Tidur itu kebutuhan biologis yang menentukan performa, mood, dan umur panjang.\" — Matthew Walker",
	"📚 \"Kamu adalah rata-rata dari 5 orang yang paling sering kamu habiskan waktumu. Pilih yang gerak, bukan yang malas.\" — Jim Rohn",
	"📚 \"Apa yang kita lakukan berulang-ulang membentuk siapa kita. Karena itu, keunggulan bukan tindakan, tapi kebiasaan.\" — Aristoteles",
	"📚 \"Tubuh yang sehat adalah tamu pelayan; jiwa yang jernih adalah tuan rumah. Rawat keduanya, dan hidupmu akan lengkap.\" — Anonim",
	}
)

// rngMu guards recentIdx for the package-level picker. math/rand/v2 is itself
// safe for concurrent use, but the recent-shown buffer is a slice that needs
// explicit synchronization.
var (
	rngMu     sync.Mutex
	recentIdx []int
)

type GetMotivationUsecase struct{}

// NewGetMotivationUsecase returns a usecase that prints a single random
// motivational quote formatted for the #motivasi command.
func NewGetMotivationUsecase() *GetMotivationUsecase {
	return &GetMotivationUsecase{}
}

func (uc *GetMotivationUsecase) Execute() string {
	var sb strings.Builder
	sb.WriteString("✨ *Motivasi Harian* ✨\n\n")
	sb.WriteString(RandomQuote())
	sb.WriteString("\n\n_Ketik #motivasi untuk quote baru!_")
	return sb.String()
}

// RandomQuote returns a single random motivational quote without formatting.
// Safe for concurrent use; avoids repeating any of the last `recentWindow`
// quotes when more than `recentWindow+1` quotes are available.
func RandomQuote() string {
	n := len(motivationalQuotes)
	if n == 0 {
		return ""
	}

	rngMu.Lock()
	defer rngMu.Unlock()

	window := recentWindow
	if window >= n-1 {
		window = n - 1
	}
	if window < 1 {
		window = 1
	}

	// Try a handful of times to find a non-recent index before giving up and
	// falling back to any random pick. With 300 quotes and a 30-quote window
	// the success rate per attempt is ~90%, so retries are rare.
	for attempt := 0; attempt < 5; attempt++ {
		idx := rand.IntN(n)
		if !containsInt(recentIdx, idx) {
			recordRecent(idx)
			return motivationalQuotes[idx]
		}
	}
	idx := rand.IntN(n)
	recordRecent(idx)
	return motivationalQuotes[idx]
}

func recordRecent(idx int) {
	recentIdx = append(recentIdx, idx)
	if len(recentIdx) > recentWindow {
		recentIdx = recentIdx[len(recentIdx)-recentWindow:]
	}
}

func containsInt(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
