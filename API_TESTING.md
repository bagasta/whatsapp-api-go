# API Testing Examples

Contoh penggunaan API sesuai dengan spesifikasi API-OLD.MD

## Environment Variables

```bash
export BASE_URL="http://localhost:3000"
export USER_ID="user123"
export AGENT_ID="agent001"
export API_KEY="your-api-key-here"
```

## 1. Session Management

### Create Session
```bash
curl -X POST "$BASE_URL/sessions" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "'"$USER_ID"'",
    "agentId": "'"$AGENT_ID"'",
    "agentName": "Support Bot",
    "apikey": "'"$API_KEY"'"
  }'
```

### Get Session Status
```bash
curl "$BASE_URL/sessions/$AGENT_ID"
```

### Get QR Code
```bash
curl -X POST "$BASE_URL/sessions/$AGENT_ID/qr"
```

### Reconnect Session
```bash
curl -X POST "$BASE_URL/sessions/$AGENT_ID/reconnect"
```

### Delete Session
```bash
curl -X DELETE "$BASE_URL/sessions/$AGENT_ID"
```

## 2. Agent Operations

### Execute AI Agent
```bash
curl -X POST "$BASE_URL/agents/$AGENT_ID/run" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "input": "Hello, how can I help you?",
    "session_id": "6281234567890",
    "parameters": {
      "max_steps": 5
    }
  }'
```

### Send Message
```bash
curl -X POST "$BASE_URL/agents/$AGENT_ID/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "to": "6281234567890",
    "message": "Hello from API!"
  }'
```

### Send Message with Quote
```bash
curl -X POST "$BASE_URL/agents/$AGENT_ID/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "to": "6281234567890",
    "message": "Reply to your message",
    "quotedMessageId": "true_6281234567890@c.us_XXXX"
  }'
```

### Send Media (Base64)
```bash
curl -X POST "$BASE_URL/agents/$AGENT_ID/media" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "to": "6281234567890",
    "data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
    "caption": "Test image",
    "filename": "test.png",
    "mimeType": "image/png"
  }'
```

### Send Media (URL)
```bash
curl -X POST "$BASE_URL/agents/$AGENT_ID/media" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "to": "6281234567890",
    "url": "https://example.com/image.jpg",
    "caption": "Image from URL"
  }'
```

## 3. Health & Metrics

### Health Check
```bash
curl "$BASE_URL/health"
```

### Metrics (Prometheus)
```bash
curl "$BASE_URL/metrics"
```

## Expected Responses

### Success Response (Create Session)
```json
{
  "isReady": false,
  "sessionState": "awaiting_qr",
  "qr": {
    "contentType": "image/png",
    "base64": "..."
  },
  "timestamps": {
    "createdAt": "2025-11-20T07:00:00Z",
    "updatedAt": "2025-11-20T07:00:00Z"
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "INVALID_PAYLOAD",
    "message": "userId, agentId, and agentName are required"
  }
}
```

## Testing with PostgreSQL

```bash
# Start PostgreSQL
docker run -d \
  --name whatsapp-postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=whatsapp \
  -p 5432:5432 \
  postgres:15

# Run application
cd src
./bin/whatsapp-api rest \
  --db-uri="postgres://postgres:password@localhost:5432/whatsapp?sslmode=disable"
```

## Multi-Agent Testing

```bash
# Create Agent 1
curl -X POST "$BASE_URL/sessions" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user123",
    "agentId": "support-bot-1",
    "agentName": "Support Bot 1"
  }'

# Create Agent 2
curl -X POST "$BASE_URL/sessions" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user123",
    "agentId": "support-bot-2",
    "agentName": "Support Bot 2"
  }'

# Both agents can run independently
```
