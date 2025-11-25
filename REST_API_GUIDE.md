# Panduan Lengkap REST API WhatsApp Go

Panduan ringkas namun lengkap untuk mode REST (termasuk kompatibilitas API-OLD). Gunakan placeholder environment saat mengetik di terminal:
```bash
export BASE_URL="http://localhost:3000"
export AGENT_ID="your-agent-id"
export USER_ID="your-user-id"
export API_KEY="your-api-key"      # untuk endpoint agent
export PHONE="6281234567890"       # tujuan WA (atau JID lengkap)
```

> Jika Basic Auth diaktifkan, tambahkan `-u user:pass` pada setiap `curl`. Untuk agent endpoints, tambahkan header `-H "Authorization: Bearer $API_KEY"` (atau `-H "X-Api-Key: $API_KEY"`).

## 1) Koneksi & Status
### Health
```bash
curl "$BASE_URL/health"
```
### Metrics (Prometheus text)
```bash
curl "$BASE_URL/metrics"
```
### Daftar perangkat/login status
```bash
curl "$BASE_URL/app/devices"
```

## 2) Session Lifecycle (API-OLD kompatibel)
### Buat/rekoneksi session (QR otomatis bila belum login)
```bash
curl -X POST "$BASE_URL/sessions" \
  -H "Content-Type: application/json" \
  -d '{"userId":"'"$USER_ID"'","agentId":"'"$AGENT_ID"'","agentName":"Support Bot","apikey":"'"$API_KEY"'"}'
```
### Cek status session
```bash
curl "$BASE_URL/sessions/$AGENT_ID"
```
### Ambil QR (PNG base64)
```bash
curl -X POST "$BASE_URL/sessions/$AGENT_ID/qr"
```
### Reconnect paksa
```bash
curl -X POST "$BASE_URL/sessions/$AGENT_ID/reconnect"
```
### Hapus session
```bash
curl -X DELETE "$BASE_URL/sessions/$AGENT_ID"
```

## 3) Login / Logout (REST UI)
### Login dengan QR
```bash
curl "$BASE_URL/app/login"
```
### Login dengan pairing code
```bash
curl "$BASE_URL/app/login-with-code?phone=$PHONE"
```
### Logout & cleanup
```bash
curl "$BASE_URL/app/logout"
```

## 4) Pengiriman Pesan (REST modern `/send/*`)
> Wajib sudah login; Basic Auth jika diaktifkan. Tidak butuh API_KEY.

### Teks
```bash
curl -X POST "$BASE_URL/send/message" \
  -H "Content-Type: application/json" \
  -d '{"phone":"'"$PHONE"'","message":"Hello from REST"}'
```
### Gambar dari URL
```bash
curl -X POST "$BASE_URL/send/image" \
  -H "Content-Type: application/json" \
  -d '{"phone":"'"$PHONE"'","imageUrl":"https://picsum.photos/200","caption":"Random image"}'
``]
### File upload (form-data)
```bash
curl -X POST "$BASE_URL/send/file" \
  -F "phone=$PHONE" \
  -F "file=@/path/to/file.pdf"
```
### Video (form-data)
```bash
curl -X POST "$BASE_URL/send/video" \
  -F "phone=$PHONE" \
  -F "video=@/path/to/video.mp4"
```
### Stiker (form-data atau URL)
```bash
curl -X POST "$BASE_URL/send/sticker" \
  -F "phone=$PHONE" \
  -F "sticker=@/path/to/image.png"
```
### Lokasi
```bash
curl -X POST "$BASE_URL/send/location" \
  -H "Content-Type: application/json" \
  -d '{"phone":"'"$PHONE"'","lat":-6.2,"lng":106.8,"address":"Jakarta"}'
```
### Chat presence (typing indicator) manual
```bash
curl -X POST "$BASE_URL/send/chat-presence" \
  -H "Content-Type: application/json" \
  -d '{"phone":"'"$PHONE"'","action":"start"}'
```

## 5) Agent (API-OLD) — butuh API_KEY
Tambahkan header auth:
```
-H "Authorization: Bearer $API_KEY"
```

### Kirim teks via agent
```bash
curl -X POST "$BASE_URL/agents/$AGENT_ID/messages" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"to":"'"$PHONE"'","message":"Hello via agent"}'
```
### Kirim media via agent (base64 singkat)
```bash
TINY_PNG="iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
curl -X POST "$BASE_URL/agents/$AGENT_ID/media" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"to":"'"$PHONE"'","data":"'"$TINY_PNG"'","filename":"tiny.png","mimeType":"image/png","caption":"tiny"}'
```
### Eksekusi AI agent (auto-reply ke WA jika berhasil)
```bash
curl -X POST "$BASE_URL/agents/$AGENT_ID/run" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"input":"Hai, bisa bantu?","session_id":"'"$PHONE"'","parameters":{"max_steps":3}}'
```

## 6) User Info (REST)
```bash
curl "$BASE_URL/user/info"
curl "$BASE_URL/user/my/contacts"
curl "$BASE_URL/user/my/groups"
curl "$BASE_URL/user/my/newsletters"
```

## 7) Chat & Message
```bash
curl "$BASE_URL/chats?limit=20"
curl "$BASE_URL/chat/$PHONE/messages?limit=20"
curl -X POST "$BASE_URL/message/12345/revoke"
curl -X POST "$BASE_URL/message/12345/reaction" -H "Content-Type: application/json" -d '{"emoji":"👍"}'
```

## 8) Group (contoh singkat)
```bash
curl -X POST "$BASE_URL/group" \
  -H "Content-Type: application/json" \
  -d '{"subject":"Test Group","participants":["'"$PHONE"'"]}'
curl "$BASE_URL/group/info?groupId=1203630xxxxx@g.us"
```

## 9) Webhook
- Set env `WHATSAPP_WEBHOOK=https://your.url/webhook` (bisa lebih dari satu, pisahkan koma).
- Set `WHATSAPP_WEBHOOK_SECRET` untuk validasi HMAC.
- Semua pesan masuk + receipt dikirim sebagai JSON ke webhook Anda.

## 10) Swagger & OpenAPI
- UI: `http://<host>:<port>/docs/swagger/`
- Spec YAML: `http://<host>:<port>/docs/openapi.yaml`

## 11) Troubleshooting Cepat
- 401 di endpoint agent: pastikan `API_KEY` benar atau Basic Auth (jika diaktifkan).
- QR tidak muncul: session sudah login; coba `DELETE /sessions/{agentId}` lalu buat ulang.
- AI tidak menjawab: cek `AI_BACKEND_URL`, log error `AI_DOWNSTREAM_ERROR`, dan format respons (`response`/`reply`).
