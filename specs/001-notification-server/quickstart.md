# Quickstart Guide: Centralized Notification Server

**Feature**: 001-notification-server  
**Purpose**: Step-by-step guide to set up, develop, test, and run the notification server

---

## Prerequisites

### Required Software

- **Go**: Version 1.21 or higher
  ```bash
  go version  # Should show go1.21+ 
  ```

- **Node.js**: Version 18 or higher
  ```bash
  node --version  # Should show v18.x or higher
  npm --version   # Should show 9.x or higher
  ```

- **SQLite**: Version 3.x (usually pre-installed on macOS/Linux)
  ```bash
  sqlite3 --version  # Should show 3.x
  ```

- **Git**: For version control
  ```bash
  git --version
  ```

### Optional Tools

- **golangci-lint**: For Go linting
  ```bash
  brew install golangci-lint  # macOS
  # or
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```

- **Postman or curl**: For API testing
- **Telegram Bot**: Create a bot via [@BotFather](https://t.me/botfather) for testing

---

## Project Setup

### 1. Clone Repository

```bash
git clone https://github.com/developertyrone/notimulti.git
cd notimulti
git checkout 001-notification-server
```

### 2. Backend Setup

```bash
cd backend

# Initialize Go module (if not already done)
go mod init github.com/developertyrone/notimulti

# Install dependencies
go mod tidy

# Verify installation
go build -o notimulti-server ./cmd/server
```

**Expected dependencies** (will be added to go.mod):
```
github.com/gin-gonic/gin
github.com/fsnotify/fsnotify
github.com/go-telegram-bot-api/telegram-bot-api/v5
gopkg.in/gomail.v2
github.com/mattn/go-sqlite3
```

### 3. Frontend Setup

```bash
cd ../frontend

# Install dependencies
npm install

# Verify installation
npm run build
```

**Expected dependencies** (package.json):
```json
{
  "dependencies": {
    "vue": "^3.3.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^4.4.0",
    "vite": "^5.0.0",
    "tailwindcss": "^3.3.0",
    "autoprefixer": "^10.4.0",
    "postcss": "^8.4.0",
    "vitest": "^1.0.0",
    "@vue/test-utils": "^2.4.0"
  }
}
```

---

## Configuration

### 1. Create Configuration Directory

```bash
mkdir -p backend/configs
```

### 2. Create Test Provider Configurations

**Telegram Configuration** (`backend/configs/telegram-test.json`):
```json
{
  "id": "telegram-test",
  "type": "telegram",
  "enabled": true,
  "config": {
    "bot_token": "YOUR_BOT_TOKEN_HERE",
    "default_chat_id": "YOUR_CHAT_ID_HERE",
    "parse_mode": "Markdown",
    "timeout_seconds": 5
  }
}
```

**Email Configuration** (`backend/configs/email-test.json`):
```json
{
  "id": "email-test",
  "type": "email",
  "enabled": true,
  "config": {
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from_address": "your-email@gmail.com",
    "from_name": "Notification Server",
    "use_tls": true,
    "timeout_seconds": 30
  }
}
```

**Note**: For Gmail, you need to create an [App Password](https://support.google.com/accounts/answer/185833).

### 3. Environment Variables

Create `backend/.env`:
```bash
LOG_LEVEL=DEBUG
LOG_FORMAT=json
CONFIG_DIR=./configs
DB_PATH=./notimulti.db
SERVER_PORT=8080
```

---

## Running the Application

### Development Mode

**Terminal 1: Backend**
```bash
cd backend
source .env  # Or use direnv
go run cmd/server/main.go
```

**Expected output**:
```
{"level":"info","msg":"Starting notification server","version":"1.0.0"}
{"level":"info","msg":"Loaded provider","id":"telegram-test","type":"telegram"}
{"level":"info","msg":"Loaded provider","id":"email-test","type":"email"}
{"level":"info","msg":"Server listening","port":"8080"}
```

**Terminal 2: Frontend**
```bash
cd frontend
npm run dev
```

**Expected output**:
```
VITE v5.0.0  ready in 234 ms

➜  Local:   http://localhost:5173/
➜  Network: use --host to expose
```

### Access the Application

- **Backend API**: http://localhost:8080/api/v1
- **Frontend UI**: http://localhost:5173
- **API Docs**: http://localhost:8080/api/v1/docs (if Swagger UI implemented)

---

## Testing the System

### 1. Health Check

```bash
curl http://localhost:8080/api/v1/health
```

**Expected response**:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2025-11-03T10:30:00Z"
}
```

### 2. List Providers

```bash
curl http://localhost:8080/api/v1/providers
```

**Expected response**:
```json
{
  "providers": [
    {
      "id": "telegram-test",
      "type": "telegram",
      "status": "active",
      "last_updated": "2025-11-03T10:00:00Z"
    },
    {
      "id": "email-test",
      "type": "email",
      "status": "active",
      "last_updated": "2025-11-03T10:00:00Z"
    }
  ]
}
```

### 3. Send Test Notification (Telegram)

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "telegram-test",
    "recipient": "YOUR_CHAT_ID",
    "message": "Test notification from notification server!",
    "priority": "normal"
  }'
```

**Expected response**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "created_at": "2025-11-03T10:30:00Z"
}
```

**Check Telegram**: You should receive the message in your configured chat.

### 4. Send Test Notification (Email)

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "email-test",
    "recipient": "your-email@example.com",
    "subject": "Test Notification",
    "message": "This is a test email from the notification server.",
    "priority": "normal"
  }'
```

**Check Email**: You should receive the email shortly.

### 5. Test Dynamic Configuration Reload

**In Terminal 3**:
```bash
cd backend/configs

# Edit telegram-test.json (change parse_mode to HTML)
sed -i '' 's/"Markdown"/"HTML"/' telegram-test.json
```

**Check Backend Logs**: Should see configuration reload message within 30 seconds:
```
{"level":"info","msg":"Configuration changed","file":"telegram-test.json"}
{"level":"info","msg":"Reloaded provider","id":"telegram-test","type":"telegram"}
```

### 6. Test Error Handling

**Invalid Provider**:
```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "invalid-provider",
    "recipient": "test",
    "message": "test"
  }'
```

**Expected response** (404):
```json
{
  "code": "PROVIDER_NOT_FOUND",
  "message": "Provider with ID 'invalid-provider' not found"
}
```

**Missing Required Field**:
```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "telegram-test",
    "recipient": "test"
  }'
```

**Expected response** (400):
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Request validation failed",
  "details": {
    "message": "required field missing"
  }
}
```

---

## Running Tests

### Backend Tests

**All tests**:
```bash
cd backend
go test ./... -v
```

**Contract tests only**:
```bash
go test ./tests/contract/... -v
```

**Integration tests** (requires test credentials):
```bash
TELEGRAM_TEST_TOKEN=xxx TELEGRAM_TEST_CHAT=xxx \
SMTP_TEST_HOST=xxx SMTP_TEST_USER=xxx SMTP_TEST_PASS=xxx \
go test ./tests/integration/... -v
```

**With coverage**:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out  # View in browser
```

**Expected output**:
```
ok      github.com/developertyrone/notimulti/internal/api         0.234s  coverage: 85.2%
ok      github.com/developertyrone/notimulti/internal/providers   0.156s  coverage: 82.7%
```

### Frontend Tests

```bash
cd frontend
npm test
```

**Expected output**:
```
 Test Files  5 passed (5)
      Tests  23 passed (23)
   Start at  10:30:00
   Duration  1.23s
```

---

## Linting and Formatting

### Backend

```bash
cd backend

# Run linter
golangci-lint run

# Format code
go fmt ./...

# Vet code
go vet ./...
```

### Frontend

```bash
cd frontend

# Run linter
npm run lint

# Format code
npm run format
```

---

## Building for Production

### Backend

```bash
cd backend
go build -o notimulti-server -ldflags="-s -w" ./cmd/server
```

**Binary location**: `backend/notimulti-server`

**Run**:
```bash
./notimulti-server
```

### Frontend

```bash
cd frontend
npm run build
```

**Static files location**: `frontend/dist/`

**Serve via backend** (configure Gin to serve static files):
```go
router.Static("/", "./frontend/dist")
```

---

## Troubleshooting

### Backend won't start

**Issue**: `panic: database is locked`
- **Solution**: Close other connections to SQLite database, or enable WAL mode

**Issue**: `Error loading provider: invalid bot token`
- **Solution**: Verify Telegram bot token is correct, test with curl:
  ```bash
  curl https://api.telegram.org/bot<YOUR_TOKEN>/getMe
  ```

**Issue**: `Error loading config: permission denied`
- **Solution**: Check file permissions on configs directory:
  ```bash
  chmod 600 backend/configs/*.json
  ```

### Frontend won't connect to backend

**Issue**: CORS errors in browser console
- **Solution**: Add CORS middleware to backend:
  ```go
  router.Use(cors.Default())
  ```

**Issue**: `net::ERR_CONNECTION_REFUSED`
- **Solution**: Verify backend is running on http://localhost:8080

### Configuration not reloading

**Issue**: Changes to config files not detected
- **Solution**: Check file watcher is working, verify logs show "Watching configs directory"
- **macOS**: May hit file descriptor limit, increase with `ulimit -n 1024`

### Notifications not delivering

**Issue**: Telegram messages not received
- **Solution**: Verify bot token and chat ID, check bot is added to chat/channel

**Issue**: Emails not sending
- **Solution**: Check SMTP credentials, verify port (587 for TLS, 465 for SSL)
- **Gmail**: Ensure "Less secure app access" enabled or use App Password

---

## Next Steps

1. **Implement Backend**:
   - Start with provider interface and registry
   - Implement Telegram provider
   - Implement Email provider
   - Add REST API handlers
   - Implement configuration loader and file watcher

2. **Implement Frontend**:
   - Create Dashboard view
   - Create ProviderCard component
   - Implement API service
   - Add auto-refresh functionality

3. **Write Tests**:
   - Contract tests for API endpoints
   - Integration tests for providers
   - Unit tests for business logic

4. **Deploy**:
   - Build production binaries
   - Set up systemd service (Linux) or launch agent (macOS)
   - Configure reverse proxy (nginx/caddy) if needed
   - Set up log aggregation

---

## Useful Commands Reference

```bash
# Backend development
go run cmd/server/main.go                     # Run server
go test ./... -v                              # Run tests
go test ./... -coverprofile=coverage.out      # Test with coverage
golangci-lint run                             # Lint code
go build -o server ./cmd/server               # Build binary

# Frontend development
npm run dev                                   # Dev server
npm test                                      # Run tests
npm run build                                 # Production build
npm run lint                                  # Lint code

# API testing
curl http://localhost:8080/api/v1/health                    # Health check
curl http://localhost:8080/api/v1/providers                 # List providers
curl -X POST http://localhost:8080/api/v1/notifications ... # Send notification

# Database
sqlite3 backend/notimulti.db "SELECT * FROM notification_logs;"  # Query logs
sqlite3 backend/notimulti.db ".schema"                           # View schema
```

---

## Additional Resources

- [Go Documentation](https://go.dev/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [Vue 3 Documentation](https://vuejs.org/)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [Telegram Bot API](https://core.telegram.org/bots/api)
- [OpenAPI Specification](https://spec.openapis.org/oas/v3.0.3)
