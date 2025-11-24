# WhatsApp API Multi-Agent - Quick Start

API WhatsApp multi-agent dengan dukungan PostgreSQL, kompatibel 100% dengan spesifikasi API-OLD.MD.

## 🚀 Quick Start (PostgreSQL)

### 1. Setup Database & Build

```bash
# Jalankan setup otomatis
./setup_postgres.sh
```

Script ini akan:
- ✅ Setup PostgreSQL database
- ✅ Create user dan permissions
- ✅ Generate .env file
- ✅ Build aplikasi

### 2. Jalankan Aplikasi

```bash
./bin/whatsapp-api rest
```

### 3. Test API

```bash
# Test otomatis semua endpoints
./test_api.sh

# Atau manual
curl http://localhost:3000/health
```

## 📚 Dokumentasi Lengkap

- **[USAGE_GUIDE.md](USAGE_GUIDE.md)** - Panduan lengkap dengan semua contoh cURL
- **[IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md)** - Status implementasi
- **[API_TESTING.md](API_TESTING.md)** - Contoh testing tambahan
- **[API-OLD.MD](API-OLD.MD)** - Spesifikasi API original

## 🔧 Konfigurasi Cepat

### PostgreSQL (Production)

```bash
# Edit .env
DB_URI="postgres://user:pass@localhost:5432/whatsapp?sslmode=disable"
```

### SQLite (Development)

```bash
# Edit .env
DB_URI="file:storages/whatsapp.db?_foreign_keys=on"
```

## 📡 API Endpoints

### Session Management
- `POST /sessions` - Create session
- `GET /sessions/:agentId` - Get status
- `DELETE /sessions/:agentId` - Delete session
- `POST /sessions/:agentId/reconnect` - Reconnect
- `POST /sessions/:agentId/qr` - Get QR code

### Agent Operations (Requires Bearer Auth)
- `POST /agents/:agentId/run` - Execute AI
- `POST /agents/:agentId/messages` - Send message
- `POST /agents/:agentId/media` - Send media

### Monitoring
- `GET /health` - Health check
- `GET /metrics` - Metrics

## 💡 Contoh Penggunaan

### Create Session

```bash
curl -X POST http://localhost:3000/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user001",
    "agentId": "bot001",
    "agentName": "Support Bot"
  }'
```

### Send Message

```bash
curl -X POST http://localhost:3000/agents/bot001/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "to": "6281234567890",
    "message": "Hello from API!"
  }'
```

## 🗄️ Database Management

### View Sessions

```bash
PGPASSWORD=whatsapp_pass_2025 psql -h localhost -U whatsapp_user -d whatsapp -c \
  "SELECT agent_id, agent_name, status FROM whatsapp_user;"
```

### Insert API Key

```bash
PGPASSWORD=whatsapp_pass_2025 psql -h localhost -U whatsapp_user -d whatsapp -c \
  "INSERT INTO api_keys (user_id, access_token, is_active) 
   VALUES ('user001', 'your_api_key', true);"
```

## 🐛 Troubleshooting

### PostgreSQL Connection Error

```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Start PostgreSQL
sudo systemctl start postgresql

# Test connection
psql -h localhost -U whatsapp_user -d whatsapp -c "SELECT 1;"
```

### Port Already in Use

```bash
# Use different port
./bin/whatsapp-api rest --port=3001
```

### Check Logs

```bash
# Run with debug
./bin/whatsapp-api rest --debug=true
```

## 🎯 Features

- ✅ Multi-agent support (multiple WhatsApp sessions)
- ✅ PostgreSQL & SQLite support
- ✅ Auto-migration (seperti Prisma)
- ✅ Bearer token authentication
- ✅ AI backend integration
- ✅ Media sending (base64 & URL)
- ✅ Health & metrics endpoints
- ✅ 100% compatible dengan API-OLD.MD

## 📦 Project Structure

```
.
├── bin/
│   └── whatsapp-api          # Binary executable
├── src/
│   ├── cmd/                  # CLI commands
│   ├── config/               # Configuration
│   ├── domains/              # Domain layer
│   │   ├── agent/           # Agent domain
│   │   ├── apikey/          # API key domain
│   │   └── session/         # Session domain
│   ├── infrastructure/       # Infrastructure layer
│   │   ├── database/        # Database migrations
│   │   ├── repository/      # Data repositories
│   │   └── whatsapp/        # WhatsApp client manager
│   └── ui/rest/             # REST handlers
│       ├── agent/           # Agent endpoints
│       └── session/         # Session endpoints
├── setup_postgres.sh        # Quick setup script
├── test_api.sh             # Testing script
└── USAGE_GUIDE.md          # Full documentation
```

## 🔐 Security Notes

- API keys stored in database
- Bearer token authentication for agent endpoints
- SQL injection protection via parameterized queries
- Session isolation per agent

## 🚀 Production Deployment

### Systemd Service

```bash
sudo nano /etc/systemd/system/whatsapp-api.service
```

```ini
[Unit]
Description=WhatsApp API Multi-Agent
After=postgresql.service

[Service]
Type=simple
User=whatsapp
WorkingDirectory=/home/bagas/Whatsapp-api-go
ExecStart=/home/bagas/Whatsapp-api-go/bin/whatsapp-api rest
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable whatsapp-api
sudo systemctl start whatsapp-api
```

## 📞 Support

Untuk pertanyaan dan issue, lihat dokumentasi lengkap di [USAGE_GUIDE.md](USAGE_GUIDE.md).

---

**Version**: 1.0.0  
**Build**: Success ✅  
**Compatibility**: API-OLD.MD 100% ✅
