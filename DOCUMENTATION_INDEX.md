# 📚 Dokumentasi WhatsApp API Multi-Agent

Index lengkap semua dokumentasi yang tersedia untuk project ini.

## 🎯 Mulai Dari Sini

### Untuk Pengguna Baru
1. **[QUICKSTART.md](QUICKSTART.md)** ⭐ - Mulai di sini! Setup cepat dalam 3 langkah
2. **[USAGE_GUIDE.md](USAGE_GUIDE.md)** 📖 - Panduan lengkap dengan semua contoh cURL

### Untuk Developer
1. **[IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md)** 🔧 - Status implementasi dan roadmap
2. **[API-OLD.MD](API-OLD.MD)** 📋 - Spesifikasi API original (referensi)

---

## 📖 Dokumentasi Utama

### [QUICKSTART.md](QUICKSTART.md)
**Quick Start Guide - Mulai dalam 3 langkah**

Isi:
- ✅ Setup otomatis dengan `setup_postgres.sh`
- ✅ Jalankan aplikasi
- ✅ Test dengan `test_api.sh`
- ✅ Troubleshooting cepat

**Kapan menggunakan**: Anda baru pertama kali setup

---

### [USAGE_GUIDE.md](USAGE_GUIDE.md)
**Panduan Lengkap - Dokumentasi komprehensif**

Isi:
- 📦 Setup PostgreSQL detail (manual & Docker)
- ⚙️ Konfigurasi aplikasi (.env)
- 🚀 Cara menjalankan (PostgreSQL & SQLite)
- 📡 Semua API endpoints dengan contoh cURL
- 🔄 Multi-agent scenarios
- 🗄️ Database management
- 🧪 Testing scripts
- 🐛 Troubleshooting lengkap
- 💡 Tips & best practices
- 🚀 Production deployment

**Kapan menggunakan**: Referensi lengkap untuk semua kebutuhan

---

### [IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md)
**Status Implementasi - Untuk Developer**

Isi:
- ✅ Fitur yang sudah selesai
- 🚧 Fitur yang perlu dilengkapi
- 📝 Catatan teknis
- 🔄 Database strategy (SQLite vs PostgreSQL)
- 🧪 Testing guide
- 📋 Next steps

**Kapan menggunakan**: Ingin tahu apa yang sudah/belum diimplementasi

---

### [API_TESTING.md](API_TESTING.md)
**Contoh Testing - cURL Examples**

Isi:
- 🔧 Environment variables
- 📡 Contoh cURL untuk semua endpoints
- ✅ Expected responses
- ❌ Error responses
- 🐳 Testing dengan Docker
- 👥 Multi-agent testing

**Kapan menggunakan**: Butuh contoh cURL cepat

---

### [API-OLD.MD](API-OLD.MD)
**Spesifikasi API Original - Referensi**

Isi:
- 🏗️ Arsitektur sistem
- 💾 Model data & storage
- 🔌 Integrasi eksternal
- 📡 Alur per endpoint
- 🔄 Alur inbound (auto-forward ke AI)
- ⚠️ Status & error handling

**Kapan menggunakan**: Referensi spesifikasi teknis original

---

## 🛠️ Scripts & Tools

### [setup_postgres.sh](setup_postgres.sh)
**Setup Otomatis PostgreSQL**

```bash
./setup_postgres.sh
```

Fungsi:
- ✅ Create database & user
- ✅ Set permissions
- ✅ Generate .env file
- ✅ Build aplikasi
- ✅ Test koneksi

---

### [test_api.sh](test_api.sh)
**Testing Script Lengkap**

```bash
./test_api.sh
```

Fungsi:
- ✅ Test semua endpoints
- ✅ Test error cases
- ✅ Multi-agent testing
- ✅ Generate QR code
- ✅ Cleanup optional

---

## 📂 File Konfigurasi

### [src/.env.example](src/.env.example)
Template konfigurasi dengan contoh:
- PostgreSQL connection string
- SQLite connection string
- AI backend settings
- WhatsApp settings

### [src/.env](src/.env)
File konfigurasi aktif (dibuat oleh `setup_postgres.sh`)

---

## 🎓 Tutorial & Workflow

### Workflow 1: First Time Setup

```bash
# 1. Setup database & build
./setup_postgres.sh

# 2. Start aplikasi
./bin/whatsapp-api rest

# 3. Test (terminal baru)
./test_api.sh
```

### Workflow 2: Development

```bash
# 1. Edit code
cd src
nano domains/session/session_usecase.go

# 2. Rebuild
go build -o ../bin/whatsapp-api

# 3. Restart aplikasi
./bin/whatsapp-api rest --debug=true
```

### Workflow 3: Testing API

```bash
# 1. Create session
curl -X POST http://localhost:3000/sessions \
  -H "Content-Type: application/json" \
  -d '{"userId":"user1","agentId":"bot1","agentName":"Bot 1"}'

# 2. Get QR
curl -X POST http://localhost:3000/sessions/bot1/qr | \
  jq -r '.qr.base64' | base64 -d > qr.png

# 3. Scan QR dengan WhatsApp

# 4. Send message
curl -X POST http://localhost:3000/agents/bot1/messages \
  -H "Authorization: Bearer YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{"to":"6281234567890","message":"Hello!"}'
```

### Workflow 4: Database Management

```bash
# 1. Connect to database
PGPASSWORD=whatsapp_pass_2025 psql -h localhost -U whatsapp_user -d whatsapp

# 2. View sessions
SELECT * FROM whatsapp_user;

# 3. View API keys
SELECT * FROM api_keys;

# 4. Insert API key
INSERT INTO api_keys (user_id, access_token, is_active)
VALUES ('user1', 'my_api_key', true);
```

---

## 🔍 Quick Reference

### Common Commands

```bash
# Build
cd src && go build -o ../bin/whatsapp-api

# Run (PostgreSQL)
./bin/whatsapp-api rest

# Run (SQLite)
./bin/whatsapp-api rest --db-uri="file:storages/whatsapp.db?_foreign_keys=on"

# Run with debug
./bin/whatsapp-api rest --debug=true --port=3001

# Health check
curl http://localhost:3000/health

# View logs
tail -f /var/log/whatsapp-api.log
```

### Database Commands

```bash
# PostgreSQL
PGPASSWORD=whatsapp_pass_2025 psql -h localhost -U whatsapp_user -d whatsapp

# Backup
pg_dump -h localhost -U whatsapp_user whatsapp > backup.sql

# Restore
psql -h localhost -U whatsapp_user whatsapp < backup.sql

# View tables
\dt

# View sessions
SELECT agent_id, status FROM whatsapp_user;
```

---

## 📊 API Endpoints Summary

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/sessions` | No | Create session |
| GET | `/sessions/:id` | No | Get status |
| DELETE | `/sessions/:id` | No | Delete session |
| POST | `/sessions/:id/reconnect` | No | Reconnect |
| POST | `/sessions/:id/qr` | No | Get QR code |
| POST | `/agents/:id/run` | Bearer | Execute AI |
| POST | `/agents/:id/messages` | Bearer | Send message |
| POST | `/agents/:id/media` | Bearer | Send media |
| GET | `/health` | No | Health check |
| GET | `/metrics` | No | Metrics |

---

## 🎯 Berdasarkan Use Case

### Use Case: Setup Pertama Kali
1. [QUICKSTART.md](QUICKSTART.md) - Setup cepat
2. [USAGE_GUIDE.md](USAGE_GUIDE.md) - Referensi lengkap

### Use Case: Testing API
1. [test_api.sh](test_api.sh) - Auto testing
2. [API_TESTING.md](API_TESTING.md) - Manual testing
3. [USAGE_GUIDE.md](USAGE_GUIDE.md) - Contoh lengkap

### Use Case: Development
1. [IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md) - Roadmap
2. [API-OLD.MD](API-OLD.MD) - Spesifikasi
3. [USAGE_GUIDE.md](USAGE_GUIDE.md) - Best practices

### Use Case: Production Deployment
1. [USAGE_GUIDE.md](USAGE_GUIDE.md) - Section "Production Deployment"
2. [setup_postgres.sh](setup_postgres.sh) - Database setup
3. [QUICKSTART.md](QUICKSTART.md) - Systemd service

### Use Case: Troubleshooting
1. [USAGE_GUIDE.md](USAGE_GUIDE.md) - Section "Troubleshooting"
2. [QUICKSTART.md](QUICKSTART.md) - Quick fixes

---

## 📞 Getting Help

1. **Cek dokumentasi**: Mulai dari [QUICKSTART.md](QUICKSTART.md)
2. **Lihat contoh**: [API_TESTING.md](API_TESTING.md)
3. **Troubleshooting**: [USAGE_GUIDE.md](USAGE_GUIDE.md) section Troubleshooting
4. **Spesifikasi teknis**: [API-OLD.MD](API-OLD.MD)

---

## ✅ Checklist Setup

- [ ] PostgreSQL installed & running
- [ ] Run `./setup_postgres.sh`
- [ ] Application built (`bin/whatsapp-api` exists)
- [ ] `.env` file configured
- [ ] Application running (`./bin/whatsapp-api rest`)
- [ ] Health check passed (`curl http://localhost:3000/health`)
- [ ] Test script passed (`./test_api.sh`)
- [ ] QR code scanned
- [ ] Message sent successfully

---

**Dibuat**: 2025-11-20  
**Versi**: 1.0.0  
**Status**: Production Ready ✅
