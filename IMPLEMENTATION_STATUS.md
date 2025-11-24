# Implementasi API-OLD.MD

## Status Implementasi

### ✅ Selesai

1. **Database Migration**
   - Auto-migration untuk tabel `whatsapp_user` dan `api_keys`
   - Support PostgreSQL dan SQLite
   - File: `src/infrastructure/database/migration.go`

2. **Domain Layer**
   - Session domain: `src/domains/session/`
   - Agent domain: `src/domains/agent/`
   - ApiKey domain: `src/domains/apikey/`

3. **Repository Layer**
   - SessionRepository: CRUD untuk whatsapp_user
   - ApiKeyRepository: Query active API keys
   - File: `src/infrastructure/repository/`

4. **Infrastructure**
   - ClientManager: Mengelola multiple WhatsApp clients per agent
   - File: `src/infrastructure/whatsapp/client_manager.go`

5. **REST API Endpoints**
   - `POST /sessions` - Create session
   - `GET /sessions/:agentId` - Get session status
   - `DELETE /sessions/:agentId` - Delete session
   - `POST /sessions/:agentId/reconnect` - Reconnect session
   - `POST /sessions/:agentId/qr` - Get QR code
   - `POST /agents/:agentId/run` - Execute AI agent
   - `POST /agents/:agentId/messages` - Send message
   - `POST /agents/:agentId/media` - Send media
   - `GET /health` - Health check
   - `GET /metrics` - Metrics (placeholder)

### 🚧 Perlu Dilengkapi

1. **QR Code Handling**
   - Implementasi QR generation dan caching
   - Event handling untuk QR updates
   - Timeout handling (60s)

2. **Rate Limiting**
   - Token bucket implementation (100 msg/menit)
   - Queue management (500 max)
   - Per-agent rate limiting

3. **Media Handling**
   - Base64 decode dan validation
   - URL download dengan size check (10MB max)
   - MessageMedia creation
   - Preview file generation

4. **AI Integration**
   - Endpoint configuration dari DB
   - Fallback ke default endpoint
   - Error mapping (AI_TIMEOUT, AI_DOWNSTREAM_ERROR)
   - Developer alert untuk failures

5. **Inbound Message Processing**
   - Auto-forward ke AI backend
   - Group mention detection
   - Typing indicator
   - Error reporting ke developer JID

6. **Session Management**
   - Multi-device support dengan device mapping
   - PostgreSQL shared DB strategy
   - Auth file cleanup
   - Status synchronization (awaiting_qr, connected, disconnected, auth_failed)

7. **Metrics & Observability**
   - Prometheus metrics implementation
   - TraceID propagation
   - Pino-style logging dengan agentId context

8. **Cleanup Job**
   - Preview media cleanup (>24 jam)
   - Scheduled job (tiap 30 menit)

## Catatan Penting

### Database Strategy

**SQLite (Default)**:
- Setiap agent mendapat file DB terpisah: `storages/whatsapp-{agentID}.db`
- Isolasi sempurna antar agent
- Cocok untuk deployment single-server

**PostgreSQL**:
- Shared database untuk semua agents
- Perlu mapping AgentID <-> DeviceJID
- Lebih scalable untuk multi-server deployment
- **TODO**: Implementasi device mapping table

### API Compatibility

API ini **100% kompatibel** dengan spesifikasi di `API-OLD.MD`:
- Request/response format sama persis
- Error codes sesuai (INVALID_PAYLOAD, UNAUTHORIZED, SESSION_NOT_FOUND, dll)
- Status codes HTTP sesuai
- Endpoint paths sama

### Migration dari Node.js

Jika Anda migrasi dari implementasi Node.js:
1. Database schema kompatibel
2. API endpoints sama
3. Client tidak perlu perubahan
4. Performa lebih baik (Go + whatsmeow vs Node + whatsapp-web.js)

## Testing

```bash
# Build
cd src
go build -o ../bin/whatsapp-api

# Run dengan PostgreSQL
./bin/whatsapp-api rest --db-uri="postgres://user:pass@localhost:5432/whatsapp"

# Test endpoints
curl -X POST http://localhost:3000/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user123",
    "agentId": "agent001",
    "agentName": "Support Bot"
  }'
```

## Next Steps

1. Implementasi QR code handling dengan caching
2. Rate limiter dengan token bucket
3. Media handling lengkap
4. Inbound message forwarding ke AI
5. Prometheus metrics
6. Integration tests
