# REST Autotest Guide (`scripts/rest_autotest.sh`)

Panduan singkat untuk menjalankan smoke test REST & API-OLD secara otomatis.

## Prasyarat
- Server berjalan (`go run ./src/main.go rest` atau binari/Docker).
- `curl`, `jq`, dan `bash` tersedia.
- API Key (jika endpoint agent membutuhkan auth).

## Variabel Lingkungan (opsional, ada default)
- `BASE_URL` (default `http://localhost:3000`)
- `USER_ID` (default `test_user_001`)
- `AGENT_ID` (default `test_bot_001`)
- `AGENT_NAME` (default `Smoke Bot`)
- `API_KEY` (Bearer / X-Api-Key untuk endpoint agent)
- `PHONE` (nomor/WA JID untuk kirim pesan; kosong = skip kirim)
- `MESSAGE` (default `Hello from rest_autotest.sh`)
- `QR_FILE` (default `qr_autotest.png`)

## Cara Menjalankan
```bash
cd /home/bagas/Whatsapp-api-go
BASE_URL=http://localhost:3000 \
USER_ID=user1 \
AGENT_ID=agent1 \
API_KEY=your_api_key_here \
PHONE=628xxxxxx \
./scripts/rest_autotest.sh
```

## Apa yang Dites
1) `/health`, `/metrics`, `/app/devices`
2) Session lifecycle (create/get) + fetch QR (disimpan ke `QR_FILE`)
3) REST send text (`/send/message`) jika `PHONE` diisi
4) API-OLD compatibility:
   - `/agents/{agentId}/messages`
   - `/agents/{agentId}/run`

## Output
- JSON hasil setiap call dicetak ke terminal.
- QR code (jika tersedia) disimpan ke `QR_FILE` (scan untuk login).
- Status tiap langkah ditandai ✓ atau ✗.

## Tips
- Jika belum login, scan QR dulu lalu ulangi script untuk uji kirim pesan.
- Gunakan `API_KEY` yang valid agar endpoint agent tidak `401`.
- Ubah `MESSAGE` jika ingin payload berbeda saat menguji AI. 
