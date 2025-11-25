# Catatan Pemahaman Proyek WhatsApp API Go

Dokumentasi singkat hasil membaca kode dan dokumen di repositori ini.

## Gambaran Besar
- Aplikasi WhatsApp Web Multi-Device berbasis Go, menyediakan dua mode: REST API + dashboard Fiber/Vue (`src/cmd/rest.go`) dan server MCP berbasis SSE untuk agent AI (`src/cmd/mcp.go`).
- Berbasis library `whatsmeow` untuk koneksi WA multi-device, dengan banyak fitur pengiriman pesan (teks, media, sticker, lokasi, kontak, poll, presence), manajemen grup/newsletter, webhook, auto-reply, auto-mark-read, auto-download media (lihat `readme.md`, `docs/openapi.yaml`).
- UI dashboard dibundel lewat embed FS (`src/main.go`, folder `src/views`) untuk login QR/pairing code, monitoring status, dan eksekusi aksi kirim pesan/kelola grup.

## Cara Aplikasi Berjalan
- Entry point `src/main.go` memanggil Cobra root command (`src/cmd/root.go`) yang membaca konfigurasi via Viper (`utils.LoadConfig`) dari flag/env/.env.
- Dua subcommand utama:
  - `rest`: bootstrap Fiber + template HTML embedded, apply middleware (basic auth opsional, CORS, recovery, logger) lalu registrasi route REST + Swagger (`src/ui/rest/*.go`, `src/ui/rest/swagger.go`). WebSocket hub untuk event UI di `src/ui/websocket`.
  - `mcp`: mengaktifkan server MCP SSE (`github.com/mark3labs/mcp-go/server`) dan mendaftarkan tools WA (`src/ui/mcp`). Endpoint SSE di `/sse`, endpoint message di `/message`.
- Auto reconnect & auto connect after boot di `ui/rest/helpers` berjalan paralel ketika server start.

## Arsitektur & Modul
- **Config** (`src/config/settings.go`): default nilai env (port, debug, base path, basic auth, webhook, batas ukuran file/video, dsb).
- **Domains/Usecase** (`src/domains`, `src/usecase`): kontrak dan implementasi layanan bisnis. Misal:
  - `send.go` memproses validasi + kompresi media, menyimpan message terkirim ke chat storage, support reply/mention/disappearing mode.
  - `session/session_usecase.go` mengelola multi-agent session: simpan ke DB, buat client WA per agent, cache QR, reconnect/delete, metrik sesi.
  - `chat`, `user`, `group`, `newsletter`, `message`, `app` mengemas query/perintah yang di-expose REST/MCP.
- **Infrastructure**:
  - `infrastructure/whatsapp`: inisialisasi klien whatsmeow, handler event (auto-reply, auto-mark-read, webhook forward, download media, presence), cleanup DB/temporary file, dan `ClientManager` untuk memetakan banyak agent -> banyak client dengan DB terpisah (SQLite per file) atau Postgres.
  - `infrastructure/chatstorage`: repository SQLite (bisa Postgres via driver) untuk simpan chats/messages + schema migrasi (`infrastructure/database`).
  - `infrastructure/repository`: repo session & API key untuk kompatibilitas API lama/multi-agent.
- **UI/REST** (`src/ui/rest`): adapter Fiber untuk setiap usecase; menyelaraskan body parser/form data, sanitasi nomor, dan respon JSON (`pkg/utils/response.go`).
- **Utility & Observabilitas**:
  - `pkg/utils`: loader env, sanitasi JID, download/convert media, response helper, dll.
  - `pkg/metrics`: counter pesan, sesi; diekspos di `/metrics` dan `/health`.

## Data & Konfigurasi
- Contoh konfigurasi ada di `src/.env.example`; `.env` aktif dibaca Viper (lihat `src/.env` di repo). Prioritas: flag > env > file.
- DB utama `DB_URI` dan `DB_KEYS_URI` bisa SQLite (`file:...`) atau Postgres (`postgres://...`). Chat storage punya URI sendiri (`CHAT_STORAGE_URI`) untuk menyimpan riwayat chat/pesan.
- Folder storage/media dikontrol variabel `Path*` di `config/settings.go` dan dibuat otomatis saat init (`utils.CreateFolder`).
- Basic auth multi-credential, base path untuk subpath deployment, trusted proxies tersedia lewat flag/env.

## Alur Penting
- **Login**: REST `/app/login` (QR) atau `/app/login-with-code`; MCP menyediakan tools `whatsapp_login_qr` & `whatsapp_login_with_code`. QR disimpan di cache `ClientManager` untuk diambil ulang.
- **Event Inbound**: semua event whatsmeow dialirkan ke handler (`infrastructure/whatsapp/init.go`). Untuk pesan masuk: simpan ke chat storage, opsi auto mark read, opsi auto-reply, lalu forward ke webhook jika diset (`WHATSAPP_WEBHOOK` + HMAC `WHATSAPP_WEBHOOK_SECRET`).
- **Pengiriman Pesan**: setiap endpoint /send/* mem-validasi request (`validations`), memastikan JID valid (`utils.ValidateJidWithLogin`), membangun payload `waE2E.Message`, lalu mengirim via `whatsapp.GetClient()` dan menyimpan message terkirim ke chat storage (asinkron dengan timeout).
- **Multi-Agent**: API kompatibilitas lama di `src/ui/rest/session` & `src/ui/rest/agent` menggunakan `ClientManager` untuk memisahkan client per `agentID` + session table `whatsapp_user`. Setiap agent bisa memiliki DB SQLite sendiri (`storages/whatsapp-<agent>.db`) jika tidak memakai Postgres.

## API & UI Singkat
- REST routes di `src/ui/rest`: `app` (login/reconnect/devices), `send`, `user`, `chat`, `message`, `group`, `newsletter`; swagger tersedia di `/docs`/`/swagger`.
- Webhook payload & spesifikasi API lebih lanjut di `docs/webhook-payload.md`, `docs/openapi.yaml`, dan tabel di `readme.md`.
- Dashboard assets di `src/views` (Vue CDN + CSS premium) dengan static path `/statics`, `/components`, `/assets`.

## Catatan Pengembangan
- Jalankan dengan `go run ./src/main.go rest` atau build `bin/whatsapp-api`; untuk MCP gunakan `... mcp --port/--host`.
- Perhatikan batas ukuran media (`config.WhatsappSettingMax*`), dependency FFmpeg untuk kompres video/gambar, dan set `APP_TRUSTED_PROXIES` jika di belakang reverse proxy.
- Testing/script referensi: `test_api.sh`, `setup_postgres.sh`, dan dokumen `QUICKSTART.md`, `USAGE_GUIDE.md`, `API_TESTING.md` (lihat `DOCUMENTATION_INDEX.md` untuk indeks lengkap).
