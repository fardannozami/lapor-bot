# WhatsApp Activity Tracker Bot

Bot WhatsApp sederhana untuk melacak aktivitas harian grup (cth: "30 Days of Sweat") dan menampilkan leaderboard streak. Bot ini dibangun menggunakan Go dan library [whatsmeow](https://github.com/tulir/whatsmeow), berjalan sebagai spesifik single-session untuk satu grup.

## Fitur

- 📝 **Self-Reporting (`/lapor`)**: Member grup dapat melapor aktivitas harian mereka.
- 🔥 **Streak Tracking**: Menghitung streak harian secara otomatis.
- 🏆 **Web Dashboard**: Menampilkan klasemen, stats personal, ranking season, dan achievement di https://lapor-bot.web.id/.
- 📱 **Multi-Login Support**: Mendukung login menggunakan QR Code atau Pairing Code.
- 💾 **SQLite Database**: Penyimpanan data ringan dan lokal.

## Prasyarat

- [Go](https://go.dev/dl/) 1.22+ terinstall.
- Akun WhatsApp (disarankan nomor sekunder/khusus bot).
- GCC compiler (untuk SQLite driver karena menggunakan CGO, cth: TDM-GCC di Windows).

## Instalasi

1. **Clone Repository**
   ```bash
   git clone https://github.com/yourusername/lapor-bot.git
   cd lapor-bot
   ```

2. **Install Dependencies**
   ```bash
   go mod tidy
   ```

## Konfigurasi

Buat file `.env` di root folder dan sesuaikan konfigurasi:

```ini
# Path database SQLite (otomatis dibuat jika belum ada)
SQLITE_PATH=./data/whatsapp.db

# ID Grup WhatsApp target (Bot hanya merespon di grup ini)
# Cara mendapatkan ID: Jalankan bot, kirim pesan di grup, cek log terminal.
# Format: 12036304xxx@g.us
GROUP_ID=12036xxxx@g.us

# (Opsional) Nomor Bot untuk Login via Pairing Code
# Format: 628xxxxxxxx (Gunakan kode negara, tanpa +)
# Jika dikosongkan, bot akan menampilkan QR Code di terminal.
BOT_PHONE=628123456789
```

## Cara Menjalankan

### Mode Development (Run langsung)
```bash
go run ./cmd/bot/main.go
```

### Build Binary
```bash
go build -o bot.exe ./cmd/bot/main.go
./bot.exe
```

## Login WhatsApp

Bot mendukung dua metode login:

### 1. Pairing Code (Rekomendasi)
1. Isi `BOT_PHONE` di file `.env`.
2. Jalankan bot.
3. Terminal akan menampilkan **Pair Code** (misal: `ABC-DEF-GH`).
4. Buka WhatsApp di HP > **Perangkat Tertaut** > **Link dengan nomor telepon**.
5. Masukkan kode tersebut.

### 2. QR Code
1. Kosongkan `BOT_PHONE` di file `.env`.
2. Jalankan bot.
3. Terminal akan menampilkan instruksi/event QR.
4. Scan QR menggunakan **WhatsApp > Perangkat Tertaut**.

## Daftar Perintah (Commands)

Bot hanya merespon perintah berikut di dalam grup yang telah dikonfigurasi (`GROUP_ID`):

| Perintah | Fungsi |
| --- | --- |
| `/lapor` | Merekam aktivitas harian user. Menambah streak jika laporan hari ini. |
| `/lapor-kemarin` | Merekam laporan khusus hari kemarin. |
| `/lapor sidequest` | Menampilkan side quest harian untuk user yang sudah punya job. |
| `/lapor sidequest [kegiatan] [jumlah]` | Melaporkan side quest. Reward bonus kecil, tetap dihitung ke streak, stats, dan leaderboard. |
| `/cancel` | Membatalkan laporan terakhir hari ini. Hanya bisa digunakan di hari yang sama. |
| `/cancel-all` | Membatalkan semua laporan hari ini. |
| `/help` | Menampilkan list command yang tersedia. |
| `/tutorial` | Menampilkan panduan lengkap cara memakai bot, termasuk link web stats dan klasemen. |

Command lain yang diawali `/` akan mendapat pesan fallback berisi link bantuan dan web dashboard. Pesan biasa tanpa prefix `/` tidak akan dibalas bot.

🌐 Klasemen, stats personal, ranking season, achievement, dan progres lain tersedia di https://lapor-bot.web.id/.

### Fitur Gamifikasi 🏅
- **Season Ranks**: Rank ala hunter dihitung dari seasonal points dan reset setiap season. Cek di web dashboard.
- **Hunter Jobs**: Job profile seperti fighter, tanker, assassin, mage, ranger, healer, atau necromancer tampil di web dashboard dan laporan harian.
- **Side Quest (`/lapor sidequest`)**: Bonus gerak harian easy/medium/hard untuk user yang sudah punya job. Lapor dengan `/lapor sidequest jalan 4000` atau `/lapor sidequest sepeda 5 km`.
- **Goals Tracking**: Set personal dan weekly goals. Bot akan mengirim notifikasi ke grup saat goal terselesaikan!
- **Season Badges**: Badge reset setiap season supaya semua member mulai berburu dari awal.
- **Lifetime Level & EXP**: Total poin dan level numerik (`Lv.0+`) tetap tersimpan lintas season. EXP naik level memakai kurva `5×level² + 50×level + 100` agar makin tinggi level makin lama naiknya.
- **Milestone Notification**: Dapat notifikasi khusus saat mencapai streak tertenu (7, 14, 30 hari, dst).
- **Leaderboard**: Bersaing dengan teman untuk streak tertinggi di https://lapor-bot.web.id/.

## Panduan AI (AI Context)
Untuk AI Agent (Gemini, Claude, Cursor, dll), silakan baca file-file berikut untuk memahami standar arsitektur:
- `CLAUDE.md` / `GEMINI.md` / `AGENTS.md` di root directory.
- `frontend/agents.md` untuk aturan ketat Turborepo monorepo.

### Recent Updates (June 2026)
- **Gamification**: Added daily quests and job-specific tasks.
- **Mobile UI Enhancements**: Implemented stacked stats layout in HunterCard for the lifetime tab, fixed LeaderboardList `NavigationContainer` context errors, resolved NativeWind navigation crash by removing conditional shadows, and added ErrorBoundary for React components in the mobile UI package.

## Struktur Project

Aplikasi ini menggunakan arsitektur monorepo untuk frontend dan backend Go:

### Backend (Go)
- `cmd/bot/main.go`: Entry point aplikasi bot WhatsApp.
- `internal/config`: Load konfigurasi `.env`.
- `internal/infra/wa`: Service WhatsApp (whatsmeow), handle koneksi & event.
- `internal/infra/sqlite`: Repository database.
- `internal/app/usecase`: Business logic (Lapor, Leaderboard, Goals).

### Frontend (Turborepo)
- `frontend/apps/web`: Aplikasi Web Dashboard (React + Vite).
- `frontend/apps/mobile`: Aplikasi Mobile (React Native + Expo).
- `frontend/packages/*`: Shared packages untuk UI, API contracts, dan design system. **Semua logic view dan API calls diletakkan di sini, bukan di dalam apps/.**

## Troubleshooting

- **Database Locked**: Pastikan tidak ada proses lain yang membuka file `.db`.
- **Bot tidak merespon**: Pastikan `GROUP_ID` di `.env` sudah benar sesuai ID grup (bukan nama grup). Cek log terminal saat ada pesan masuk.
- **Login Gagal**: Hapus file database di folder `data/` untuk reset sesi dan login ulang.

## Kontribusi

Kontribusi sangat dipersilakan! Silakan ikuti langkah-langkah berikut:

1.  **Fork** repository ini.
2.  Buat **Branch** baru untuk fitur/fix Anda (`git checkout -b fitur-keren`).
3.  **Commit** perubahan Anda (`git commit -m 'Menambahkan fitur keren'`).
4.  **Push** ke Branch (`git push origin fitur-keren`).
5.  Buat **Pull Request**.

Jangan ragu untuk membuka _Issue_ jika menemukan bug atau memiliki saran fitur baru.
