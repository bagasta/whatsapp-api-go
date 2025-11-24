# Kumpulan cURL WhatsApp API Multi-Agent

Dokumen ini berisi kumpulan perintah cURL siap pakai untuk berinteraksi dengan WhatsApp API.

## 📋 Daftar Isi
1. [Session Management](#1-session-management)
2. [Agent Messaging](#2-agent-messaging)
3. [AI Agent Operations](#3-ai-agent-operations)
4. [System Health](#4-system-health)

---

## Konfigurasi Dasar
Ganti variabel berikut sesuai kebutuhan Anda:
- `BASE_URL`: `http://localhost:8080`
- `API_KEY`: Key rahasia yang Anda buat saat create session (misal: `secret123`)
- `AGENT_ID`: ID unik bot Anda (misal: `bot_cs_01`)
- `PHONE`: Nomor tujuan format internasional tanpa `+` (misal: `628123456789`)

---

## 1. Session Management

### Buat Session Baru (Register Bot)
Mendaftarkan bot baru ke sistem.
```bash
curl -X POST "http://localhost:8080/sessions" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_001",
    "agentId": "bot_cs_01",
    "agentName": "Customer Service Bot",
    "apikey": "secret123"
  }'
```

### Dapatkan QR Code
Mengambil QR code untuk login WhatsApp (jika belum login).
```bash
curl -X POST "http://localhost:8080/sessions/bot_cs_01/qr"
```
*Tips: Copy string base64 dari response dan convert ke image di browser/tools.*

### Cek Status Session
Mengecek apakah bot sudah terhubung (`connected`).
```bash
curl "http://localhost:8080/sessions/bot_cs_01"
```

### Hapus Session (Logout)
Menghapus sesi bot dari sistem.
```bash
curl -X DELETE "http://localhost:8080/sessions/bot_cs_01"
```

---

## 2. Agent Messaging
**Note:** Semua request di bawah ini memerlukan header `Authorization: Bearer <API_KEY>`.

### Kirim Pesan Teks
```bash
curl -X POST "http://localhost:8080/agents/bot_cs_01/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer secret123" \
  -d '{
    "to": "628123456789",
    "message": "Halo! Terima kasih telah menghubungi kami."
  }'
```

### Kirim Gambar (via URL)
```bash
curl -X POST "http://localhost:8080/agents/bot_cs_01/media" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer secret123" \
  -d '{
    "to": "628123456789",
    "url": "https://picsum.photos/400/300",
    "caption": "Ini contoh gambar dari URL",
    "mimeType": "image/jpeg"
  }'
```

### Kirim Gambar (via Base64)
```bash
curl -X POST "http://localhost:8080/agents/bot_cs_01/media" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer secret123" \
  -d '{
    "to": "628123456789",
    "data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
    "caption": "Ini gambar titik merah (Base64)",
    "mimeType": "image/png"
  }'
```

### Kirim Dokumen (PDF/Doc)
```bash
curl -X POST "http://localhost:8080/agents/bot_cs_01/media" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer secret123" \
  -d '{
    "to": "628123456789",
    "url": "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf",
    "filename": "dokumen_penting.pdf",
    "caption": "Silakan cek dokumen terlampir",
    "mimeType": "application/pdf"
  }'
```

---

## 3. AI Agent Operations

### Jalankan AI Agent (Chat)
Mengirim input ke AI Backend (n8n/Python) dan membalas otomatis ke user.
```bash
curl -X POST "http://localhost:8080/agents/bot_cs_01/run" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer secret123" \
  -d '{
    "session_id": "628123456789",
    "input": "Bisa bantu jelaskan produk terbaru?",
    "parameters": {
        "tone": "formal",
        "language": "id"
    }
  }'
```

---

## 4. System Health

### Cek Kesehatan Server
```bash
curl "http://localhost:8080/health"
```
