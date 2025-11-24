# Panduan Lengkap WhatsApp API Multi-Agent

Dokumentasi lengkap untuk menggunakan WhatsApp API dengan dukungan multi-agent sesuai spesifikasi API-OLD.MD.

## Daftar Isi

1. [Persiapan](#persiapan)
2. [Setup PostgreSQL](#setup-postgresql)
3. [Konfigurasi Aplikasi](#konfigurasi-aplikasi)
4. [Menjalankan Aplikasi](#menjalankan-aplikasi)
5. [API Endpoints](#api-endpoints)
6. [Contoh Penggunaan](#contoh-penggunaan)
7. [Troubleshooting](#troubleshooting)

---

## Persiapan

### Requirements

- Go 1.24+ (sudah terinstall)
- PostgreSQL 12+ (untuk production)
- SQLite3 (untuk development, sudah built-in)

### Build Aplikasi

```bash
cd /home/bagas/Whatsapp-api-go/src
go build -o ../bin/whatsapp-api
```

---

## Setup PostgreSQL

### 1. Install PostgreSQL (Ubuntu/Debian)

```bash
# Install PostgreSQL
sudo apt update
sudo apt install postgresql postgresql-contrib

# Start service
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### 2. Buat Database dan User

```bash
# Masuk ke PostgreSQL
sudo -u postgres psql

# Buat database dan user
CREATE DATABASE whatsapp;
CREATE USER whatsapp_user WITH ENCRYPTED PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE whatsapp TO whatsapp_user;

# Keluar
\q
```

### 3. Alternatif: Menggunakan Docker

```bash
# Run PostgreSQL dengan Docker
docker run -d \
  --name whatsapp-postgres \
  -e POSTGRES_DB=whatsapp \
  -e POSTGRES_USER=whatsapp_user \
  -e POSTGRES_PASSWORD=your_secure_password \
  -p 5432:5432 \
  -v whatsapp_data:/var/lib/postgresql/data \
  postgres:15

# Cek status
docker ps | grep whatsapp-postgres
```

### 4. Test Koneksi

```bash
# Test koneksi ke PostgreSQL
psql -h localhost -U whatsapp_user -d whatsapp -c "SELECT version();"
```

---

## Konfigurasi Aplikasi

### 1. Copy dan Edit .env

```bash
cd /home/bagas/Whatsapp-api-go/src
cp .env.example .env
nano .env
```

### 2. Konfigurasi untuk PostgreSQL

Edit file `.env`:

```bash
# Application Settings
APP_PORT=3000
APP_DEBUG=true
APP_OS=Chrome
APP_BASE_PATH=
APP_TRUSTED_PROXIES=0.0.0.0/0

# Database Settings (PostgreSQL)
DB_URI="postgres://whatsapp_user:your_secure_password@localhost:5432/whatsapp?sslmode=disable"
DB_KEYS_URI=""

# Chat Storage Database (PostgreSQL)
CHAT_STORAGE_URI="postgres://whatsapp_user:your_secure_password@localhost:5432/whatsapp?sslmode=disable"

# WhatsApp Settings
WHATSAPP_AUTO_REPLY=""
WHATSAPP_AUTO_MARK_READ=false
WHATSAPP_AUTO_DOWNLOAD_MEDIA=true
WHATSAPP_WEBHOOK=
WHATSAPP_WEBHOOK_SECRET=super-secret-key
WHATSAPP_ACCOUNT_VALIDATION=true
WHATSAPP_CHAT_STORAGE=true

# AI Backend Settings (optional)
# AI_BACKEND_URL=http://localhost:5000
# DEVELOPER_JID=6281234567890@s.whatsapp.net
```

### 3. Konfigurasi untuk SQLite (Development)

```bash
# Database Settings (SQLite)
DB_URI="file:storages/whatsapp.db?_foreign_keys=on"
DB_KEYS_URI="file::memory:?cache=shared&_foreign_keys=on"

# Chat Storage Database (SQLite)
CHAT_STORAGE_URI="file:storages/chatstorage.db"
```

---

## Menjalankan Aplikasi

### 1. Dengan PostgreSQL

```bash
cd /home/bagas/Whatsapp-api-go

# Jalankan dengan config dari .env
./bin/whatsapp-api rest

# Atau dengan flag manual
./bin/whatsapp-api rest \
  --port=3000 \
  --db-uri="postgres://whatsapp_user:your_secure_password@localhost:5432/whatsapp?sslmode=disable" \
  --debug=true
```

### 2. Dengan SQLite (Development)

```bash
./bin/whatsapp-api rest \
  --port=3000 \
  --db-uri="file:storages/whatsapp.db?_foreign_keys=on" \
  --debug=true
```

### 3. Verifikasi Aplikasi Berjalan

```bash
# Cek health endpoint
curl http://localhost:3000/health

# Expected response:
# {"status":"ok","timestamp":1700000000,"uptime":0}
```

---

## API Endpoints

### Base URL

```
http://localhost:3000
```

### Authentication

Endpoints `/agents/*` memerlukan Bearer token di header:

```
Authorization: Bearer YOUR_API_KEY
```

---

## Contoh Penggunaan

### Setup Environment Variables untuk Testing

```bash
export BASE_URL="http://localhost:3000"
export USER_ID="user_bagas_001"
export AGENT_ID="support_bot_001"
export API_KEY="test_api_key_12345"
export PHONE_NUMBER="6281234567890"
```

---

## 1. Session Management

### 1.1 Create Session (Buat Sesi WhatsApp Baru)

```bash
curl -X POST "$BASE_URL/sessions" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "'"$USER_ID"'",
    "agentId": "'"$AGENT_ID"'",
    "agentName": "Support Bot Indonesia",
    "apikey": "'"$API_KEY"'"
  }' | jq
```

**Response Success:**
```json
{
  "isReady": false,
  "sessionState": "awaiting_qr",
  "qr": {
    "contentType": "image/png",
    "base64": "2@v5pT8DABcl5R..."
  },
  "timestamps": {
    "createdAt": "2025-11-20T07:00:00Z",
    "updatedAt": "2025-11-20T07:00:00Z"
  }
}
```

**Response Error:**
```json
{
  "error": {
    "code": "INVALID_PAYLOAD",
    "message": "userId, agentId, and agentName are required"
  }
}
```

### 1.2 Get Session Status

```bash
curl "$BASE_URL/sessions/$AGENT_ID" | jq
```

**Response:**
```json
{
  "isReady": true,
  "hasClient": true,
  "sessionState": "authenticated"
}
```

### 1.3 Get QR Code

```bash
curl -X POST "$BASE_URL/sessions/$AGENT_ID/qr" | jq
```

**Response:**
```json
{
  "qr": {
    "contentType": "image/png",
    "base64": "iVBORw0KGgoAAAANS..."
  },
  "qrUpdatedAt": "2025-11-20T07:05:00Z"
}
```

**Cara Scan QR Code:**

```bash
# Save QR to file
curl -X POST "$BASE_URL/sessions/$AGENT_ID/qr" | \
  jq -r '.qr.base64' | \
  base64 -d > qr_code.png

# Buka file qr_code.png dan scan dengan WhatsApp
```

### 1.4 Reconnect Session

```bash
curl -X POST "$BASE_URL/sessions/$AGENT_ID/reconnect" | jq
```

### 1.5 Delete Session

```bash
curl -X DELETE "$BASE_URL/sessions/$AGENT_ID" | jq
```

**Response:**
```json
{
  "deleted": true
}
```

---

## 2. Agent Operations (AI & Messaging)

### 2.1 Execute AI Agent

Endpoint ini memanggil AI backend dan mengirim reply ke WhatsApp.

```bash
curl -X POST "$BASE_URL/agents/$AGENT_ID/run" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "input": "Halo, saya butuh bantuan",
    "session_id": "'"$PHONE_NUMBER"'",
    "parameters": {
      "max_steps": 5
    }
  }' | jq
```

**Response:**
```json
{
  "reply": "Halo! Ada yang bisa saya bantu?",
  "replySent": true,
  "traceId": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Error Responses:**

```bash
# 401 - Unauthorized
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid API key"
  }
}

# 404 - Session not found
{
  "error": {
    "code": "SESSION_NOT_FOUND",
    "message": "Session not found"
  }
}

# 409 - Session not ready
{
  "error": {
    "code": "SESSION_NOT_READY",
    "message": "WhatsApp session is not connected"
  }
}

# 504 - AI timeout
{
  "error": {
    "code": "AI_TIMEOUT",
    "message": "AI backend timeout"
  }
}
```

### 2.2 Send Message (Kirim Pesan Teks)

```bash
curl -X POST "$BASE_URL/agents/$AGENT_ID/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "to": "'"$PHONE_NUMBER"'",
    "message": "Halo, ini pesan dari API!"
  }' | jq
```

**Response:**
```json
{
  "delivered": true
}
```

### 2.3 Send Message with Quote (Reply)

```bash
curl -X POST "$BASE_URL/agents/$AGENT_ID/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "to": "'"$PHONE_NUMBER"'",
    "message": "Ini balasan untuk pesan Anda",
    "quotedMessageId": "true_6281234567890@c.us_3EB0C431B23A1D5E2F45"
  }' | jq
```

### 2.4 Send Media - Base64

```bash
# Kirim gambar dari base64
curl -X POST "$BASE_URL/agents/$AGENT_ID/media" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "to": "'"$PHONE_NUMBER"'",
    "data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
    "caption": "Ini gambar test",
    "filename": "test.png",
    "mimeType": "image/png"
  }' | jq
```

**Encode file ke base64:**

```bash
# Encode image
base64 -w 0 /path/to/image.jpg > image_base64.txt

# Gunakan dalam curl
IMAGE_BASE64=$(cat image_base64.txt)
curl -X POST "$BASE_URL/agents/$AGENT_ID/media" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "to": "'"$PHONE_NUMBER"'",
    "data": "'"$IMAGE_BASE64"'",
    "caption": "Foto dari API",
    "filename": "photo.jpg",
    "mimeType": "image/jpeg"
  }' | jq
```

### 2.5 Send Media - URL

```bash
curl -X POST "$BASE_URL/agents/$AGENT_ID/media" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "to": "'"$PHONE_NUMBER"'",
    "url": "https://picsum.photos/200/300",
    "caption": "Random image dari URL"
  }' | jq
```

**Response:**
```json
{
  "delivered": true,
  "previewPath": "/statics/media/preview_1234567890.jpg"
}
```

---

## 3. Health & Monitoring

### 3.1 Health Check

```bash
curl "$BASE_URL/health" | jq
```

**Response:**
```json
{
  "status": "ok",
  "uptime": 3600.5,
  "timestamp": 1700000000
}
```

### 3.2 Metrics (Prometheus)

```bash
curl "$BASE_URL/metrics"
```

**Response:**
```
# Prometheus metrics placeholder
# TODO: Implement actual metrics
```

---

## 4. Multi-Agent Scenario

### Skenario: Mengelola 3 Bot Berbeda

```bash
# Bot 1: Customer Support
curl -X POST "$BASE_URL/sessions" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "company_001",
    "agentId": "support_bot",
    "agentName": "Customer Support Bot"
  }' | jq

# Bot 2: Sales Bot
curl -X POST "$BASE_URL/sessions" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "company_001",
    "agentId": "sales_bot",
    "agentName": "Sales Bot"
  }' | jq

# Bot 3: Notification Bot
curl -X POST "$BASE_URL/sessions" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "company_001",
    "agentId": "notification_bot",
    "agentName": "Notification Bot"
  }' | jq

# Kirim pesan dari bot berbeda
curl -X POST "$BASE_URL/agents/support_bot/messages" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"to": "6281234567890", "message": "Dari Support Bot"}' | jq

curl -X POST "$BASE_URL/agents/sales_bot/messages" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"to": "6281234567890", "message": "Dari Sales Bot"}' | jq
```

---

## 5. Database Management

### Cek Tabel di PostgreSQL

```bash
# Masuk ke PostgreSQL
psql -h localhost -U whatsapp_user -d whatsapp

# Lihat semua tabel
\dt

# Lihat isi tabel whatsapp_user
SELECT * FROM whatsapp_user;

# Lihat isi tabel api_keys
SELECT * FROM api_keys;

# Keluar
\q
```

### Insert API Key Manual

```bash
psql -h localhost -U whatsapp_user -d whatsapp -c "
INSERT INTO api_keys (user_id, access_token, is_active, created_at)
VALUES ('user_bagas_001', 'test_api_key_12345', true, NOW());
"
```

### Lihat Semua Sessions

```bash
psql -h localhost -U whatsapp_user -d whatsapp -c "
SELECT agent_id, agent_name, status, last_connected_at 
FROM whatsapp_user 
ORDER BY updated_at DESC;
"
```

---

## 6. Testing Script

### Script Lengkap untuk Testing

Buat file `test_api.sh`:

```bash
#!/bin/bash

# Configuration
BASE_URL="http://localhost:3000"
USER_ID="test_user_001"
AGENT_ID="test_agent_001"
API_KEY="test_api_key_12345"
PHONE="6281234567890"

echo "=== WhatsApp API Testing ==="
echo ""

# 1. Health Check
echo "1. Testing Health Check..."
curl -s "$BASE_URL/health" | jq
echo ""

# 2. Create Session
echo "2. Creating Session..."
curl -s -X POST "$BASE_URL/sessions" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "'"$USER_ID"'",
    "agentId": "'"$AGENT_ID"'",
    "agentName": "Test Bot"
  }' | jq
echo ""

# 3. Get Session Status
echo "3. Getting Session Status..."
sleep 2
curl -s "$BASE_URL/sessions/$AGENT_ID" | jq
echo ""

# 4. Send Message (will fail if not authenticated)
echo "4. Testing Send Message..."
curl -s -X POST "$BASE_URL/agents/$AGENT_ID/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "to": "'"$PHONE"'",
    "message": "Test message from API"
  }' | jq
echo ""

echo "=== Testing Complete ==="
```

Jalankan:

```bash
chmod +x test_api.sh
./test_api.sh
```

---

## Troubleshooting

### 1. Database Connection Error

```bash
# Error: connection refused
# Solusi: Pastikan PostgreSQL running
sudo systemctl status postgresql
sudo systemctl start postgresql

# Test koneksi
psql -h localhost -U whatsapp_user -d whatsapp -c "SELECT 1;"
```

### 2. Migration Error

```bash
# Error: failed to run migrations
# Solusi: Cek permission database
psql -h localhost -U whatsapp_user -d whatsapp -c "
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO whatsapp_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO whatsapp_user;
"
```

### 3. Session Not Ready

```bash
# Error: SESSION_NOT_READY
# Solusi: Pastikan sudah scan QR code
curl -X POST "$BASE_URL/sessions/$AGENT_ID/qr" | jq -r '.qr.base64' | base64 -d > qr.png
# Scan qr.png dengan WhatsApp
```

### 4. Port Already in Use

```bash
# Error: bind: address already in use
# Solusi: Gunakan port lain
./bin/whatsapp-api rest --port=3001
```

### 5. Check Logs

```bash
# Run dengan debug mode
./bin/whatsapp-api rest --debug=true

# Atau lihat logs
tail -f /var/log/whatsapp-api.log
```

---

## Tips & Best Practices

### 1. Production Deployment

```bash
# Gunakan systemd service
sudo nano /etc/systemd/system/whatsapp-api.service
```

```ini
[Unit]
Description=WhatsApp API Multi-Agent
After=network.target postgresql.service

[Service]
Type=simple
User=whatsapp
WorkingDirectory=/home/bagas/Whatsapp-api-go
ExecStart=/home/bagas/Whatsapp-api-go/bin/whatsapp-api rest
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Enable dan start
sudo systemctl daemon-reload
sudo systemctl enable whatsapp-api
sudo systemctl start whatsapp-api
sudo systemctl status whatsapp-api
```

### 2. Backup Database

```bash
# Backup PostgreSQL
pg_dump -h localhost -U whatsapp_user whatsapp > backup_$(date +%Y%m%d).sql

# Restore
psql -h localhost -U whatsapp_user whatsapp < backup_20251120.sql
```

### 3. Monitor Performance

```bash
# Monitor connections
watch -n 1 'psql -h localhost -U whatsapp_user -d whatsapp -c "SELECT count(*) FROM whatsapp_user WHERE status = '\''connected'\'';"'

# Monitor API
watch -n 1 'curl -s http://localhost:3000/health | jq'
```

---

## Referensi

- **API Spec**: `API-OLD.MD`
- **Implementation Status**: `IMPLEMENTATION_STATUS.md`
- **Testing Examples**: `API_TESTING.md`
- **Environment Config**: `src/.env.example`

---

**Dibuat oleh**: Antigravity AI  
**Tanggal**: 2025-11-20  
**Versi**: 1.0.0
