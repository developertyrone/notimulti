# Notimulti - Centralized Notification Server

> A simple, lightweight notification server that routes messages through multiple provider channels (Telegram, Email) via REST API.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Vue Version](https://img.shields.io/badge/Vue-3.3+-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- SQLite 3

### Installation

```bash
# Clone the repository
git clone https://github.com/developertyrone/notimulti.git
cd notimulti

# Set up backend
cd backend
go mod download
cp .env.example .env
# Edit .env with your configuration

# Set up frontend
cd ../frontend
npm install

# Start backend (terminal 1)
cd backend
go run cmd/server/main.go

# Start frontend (terminal 2)
cd frontend
npm run dev
```

### Send Your First Notification

```bash
# Create a Telegram provider config
cat > backend/configs/telegram-test.json << EOF
{
  "id": "telegram-test",
  "type": "telegram",
  "enabled": true,
  "config": {
    "bot_token": "YOUR_BOT_TOKEN",
    "default_chat_id": "YOUR_CHAT_ID",
    "parse_mode": "Markdown"
  }
}
EOF

# Send a notification
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "telegram-test",
    "recipient": "YOUR_CHAT_ID",
    "message": "Hello from Notimulti! 🎉"
  }'
```

## 📖 Features

### ✅ User Story 1: Send Notifications via REST API
- 🔌 **REST API** for sending notifications to multiple providers
- 📱 **Telegram support** with Markdown/HTML formatting
- 📧 **Email support** with SMTP/TLS
- 🔄 **Automatic retry** with exponential backoff
- 📊 **SQLite logging** for audit trail
- ⚡ **High throughput** (100+ concurrent requests)

### ✅ User Story 2: Dynamic Provider Configuration
- 📁 **File-based config** (JSON files in `configs/` directory)
- 🔄 **Auto-reload** on configuration changes (<30s detection)
- ✅ **Validation** before applying changes
- 🔀 **Atomic updates** without downtime
- 🛡️ **Error resilience** (keeps old config on failure)

### ✅ User Story 3: View Current Configuration
- 🎨 **Web dashboard** for monitoring provider status
- 🔒 **Sensitive data masking** (tokens, passwords)
- 📊 **Real-time status** (auto-refresh every 30s)
- 📱 **Mobile-responsive** UI with Tailwind CSS

## 🏗️ Architecture

```
┌─────────────┐          ┌──────────────────┐          ┌────────────┐
│   Client    │   HTTP   │   REST API       │          │  Providers │
│ Application │─────────▶│   (Gin Router)   │─────────▶│            │
└─────────────┘          └──────────────────┘          │  Telegram  │
                                  │                     │   Email    │
                                  │                     └────────────┘
                                  ▼                            ▲
                         ┌──────────────────┐                 │
                         │  Provider        │─────────────────┘
                         │  Registry        │
                         └──────────────────┘
                                  │
                         ┌────────┴────────┐
                         ▼                 ▼
                ┌──────────────┐  ┌──────────────┐
                │  File        │  │  SQLite      │
                │  Watcher     │  │  Database    │
                └──────────────┘  └──────────────┘
                         ▲
                         │
                    JSON Configs
```

### Technology Stack

**Backend:**
- Go 1.21+ with Gin web framework
- fsnotify for file watching
- go-telegram-bot-api for Telegram integration
- gomail for SMTP email
- SQLite 3 for persistence

**Frontend:**
- Vue 3 (Composition API)
- Vite 5 (dev server & build)
- Tailwind CSS 3 (styling)

## 📚 API Documentation

### Endpoints

#### Health Check
```http
GET /api/v1/health
```

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2025-11-06T10:30:00Z"
}
```

#### Send Notification
```http
POST /api/v1/notifications
Content-Type: application/json
```

**Request Body:**
```json
{
  "provider_id": "telegram-alerts",
  "recipient": "-1001234567890",
  "message": "Server alert: High CPU usage",
  "priority": "high",
  "metadata": {
    "source": "monitoring-system",
    "severity": "warning"
  }
}
```

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "created_at": "2025-11-06T10:30:00Z"
}
```

#### List Providers
```http
GET /api/v1/providers
```

**Response:**
```json
{
  "providers": [
    {
      "id": "telegram-alerts",
      "type": "telegram",
      "status": "active",
      "last_updated": "2025-11-06T10:00:00Z"
    }
  ]
}
```

#### Get Provider Details
```http
GET /api/v1/providers/:id
```

**Response:**
```json
{
  "id": "telegram-alerts",
  "type": "telegram",
  "status": "active",
  "enabled": true,
  "last_updated": "2025-11-06T10:00:00Z",
  "config": {
    "default_chat_id": "-1001234567890",
    "parse_mode": "Markdown",
    "bot_token": "****masked****"
  }
}
```

For full API specification, see [specs/001-notification-server/contracts/openapi.yaml](specs/001-notification-server/contracts/openapi.yaml).

## ⚙️ Configuration

### Provider Configuration

Create JSON files in `backend/configs/` directory:

**Telegram Provider (`telegram-alerts.json`):**
```json
{
  "id": "telegram-alerts",
  "type": "telegram",
  "enabled": true,
  "config": {
    "bot_token": "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz",
    "default_chat_id": "-1001234567890",
    "parse_mode": "Markdown",
    "timeout_seconds": 5
  }
}
```

**Email Provider (`email-prod.json`):**
```json
{
  "id": "email-prod",
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

### Environment Variables

Create `backend/.env` from `.env.example`:

```bash
LOG_LEVEL=INFO          # DEBUG, INFO, WARN, ERROR
LOG_FORMAT=json         # json, text
CONFIG_DIR=./configs    # Path to provider configs
DB_PATH=./notimulti.db  # SQLite database path
SERVER_PORT=8080        # HTTP server port
```

## 🧪 Testing

### Run All Tests

```bash
# Backend tests
cd backend
go test ./... -v

# Backend tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Frontend tests
cd frontend
npm test

# Frontend tests with coverage
npm test -- --coverage
```

### Run Specific Test Suites

```bash
# Contract tests (API endpoints)
go test ./tests/contract/... -v

# Integration tests (file watching, database)
go test ./tests/integration/... -v

# Unit tests
go test ./tests/unit/... -v
```

## 🚀 Production Deployment

### Build Production Artifacts

```bash
# Build backend binary
cd backend
go build -o notimulti-server -ldflags="-s -w" ./cmd/server

# Build frontend static files
cd frontend
npm run build
```

### Systemd Service (Linux)

Create `/etc/systemd/system/notimulti.service`:

```ini
[Unit]
Description=Notimulti Notification Server
After=network.target

[Service]
Type=simple
User=notimulti
Group=notimulti
WorkingDirectory=/opt/notimulti
ExecStart=/opt/notimulti/notimulti-server
Restart=on-failure
RestartSec=5s

# Environment
Environment="LOG_LEVEL=INFO"
Environment="LOG_FORMAT=json"
Environment="CONFIG_DIR=/etc/notimulti/configs"
Environment="DB_PATH=/var/lib/notimulti/notifications.db"
Environment="SERVER_PORT=8080"

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/notimulti /etc/notimulti/configs

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable notimulti
sudo systemctl start notimulti
sudo systemctl status notimulti
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name notifications.example.com;

    location /api/ {
        proxy_pass http://localhost:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    location / {
        root /opt/notimulti/frontend/dist;
        try_files $uri $uri/ /index.html;
    }
}
```

## 🔒 Security Considerations

1. **File Permissions:**
   ```bash
   chmod 600 backend/.env
   chmod 600 backend/configs/*.json
   chmod 600 backend/notimulti.db
   ```

2. **Secret Management:**
   - Never commit `.env` or provider configs to version control
   - Use secret management tools (Vault, AWS Secrets Manager) for production

3. **Network Security:**
   - Use reverse proxy with TLS/SSL
   - Configure firewall rules
   - Consider API authentication for public exposure

4. **Monitoring:**
   - Monitor logs for errors and security events
   - Set up alerts for critical failures
   - Track disk usage for database growth

## 🐛 Troubleshooting

### Backend Won't Start

**Issue:** `panic: database is locked`
```bash
# Solution: Enable WAL mode (already configured in code)
# Or close other connections to the database
```

**Issue:** `Error loading provider: invalid bot token`
```bash
# Solution: Verify Telegram bot token
curl https://api.telegram.org/bot<YOUR_TOKEN>/getMe
```

### Configuration Not Reloading

**Issue:** Changes to config files not detected
```bash
# Check file watcher is running
# macOS: Increase file descriptor limit
ulimit -n 1024
```

### Notifications Not Delivering

**Issue:** Telegram messages not received
```bash
# Verify bot token and chat ID
# Ensure bot is added to the chat/channel
```

**Issue:** Emails not sending
```bash
# Check SMTP credentials and port
# Gmail: Use App Password instead of regular password
# Verify TLS/STARTTLS configuration
```

## 📊 Performance

- **API Response Time:** <2s (p95) for 100 concurrent requests
- **Configuration Reload:** <5s (typical: <2s)
- **UI Interaction:** <200ms (p95)
- **Provider Change Detection:** <30s (typical: <5s)

## 🗺️ Roadmap

See [CHANGELOG.md](CHANGELOG.md) for detailed feature roadmap.

**v1.1.0 (Planned):**
- SMS provider (Twilio)
- Slack, Discord, Microsoft Teams providers
- Notification templates
- API key authentication

**v1.2.0 (Planned):**
- Retry queue for failed notifications
- Rate limiting
- Notification scheduling
- Configuration UI

**v2.0.0 (Planned):**
- PostgreSQL support
- Message queue integration
- Prometheus metrics
- Webhook callbacks

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📧 Support

- **Issues:** [GitHub Issues](https://github.com/developertyrone/notimulti/issues)
- **Documentation:** [specs/001-notification-server/](specs/001-notification-server/)
- **API Docs:** [OpenAPI Specification](specs/001-notification-server/contracts/openapi.yaml)

---

Made with ❤️ by [developertyrone](https://github.com/developertyrone)
