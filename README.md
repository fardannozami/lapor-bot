# WhatsApp Activity Tracker Bot

Bot WhatsApp sederhana untuk melacak aktivitas harian grup (cth: "30 Days of Sweat") dan menampilkan leaderboard streak. Bot ini dibangun menggunakan Go dan library [whatsmeow](https://github.com/tulir/whatsmeow), berjalan sebagai spesifik single-session untuk satu grup.

## Fitur

- 📝 **Self-Reporting (`#lapor`)**: Member grup dapat melapor aktivitas harian mereka.
- 🔥 **Streak Tracking**: Menghitung streak harian secara otomatis.
- 🏆 **Leaderboard (`#leaderboard`)**: Menampilkan klasemen streak tertinggi.
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
| `#lapor` | Merekam aktivitas harian user. Menambah streak jika laporan hari ini/kemarin. |
| `#leaderboard` | Menampilkan klasemen streak, daftar yang "Keep Streak" 🔥 dan "Lose Streak" 💔. |

## Struktur Project

- `cmd/bot/main.go`: Entry point aplikasi.
- `internal/config`: Load konfigurasi `.env`.
- `internal/infra/wa`: Service WhatsApp (whatsmeow), handle koneksi & event.
- `internal/infra/sqlite`: Repository database.
- `internal/app/usecase`: Business logic (Lapor, Leaderboard).

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
