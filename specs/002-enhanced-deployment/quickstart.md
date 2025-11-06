# Quickstart Guide: Enhanced Deployment & Operations

**Feature**: 002-enhanced-deployment  
**Date**: 2025-11-06  
**Purpose**: Get the notification server running with Docker in under 5 minutes

## Prerequisites

- Docker 20.10+ and Docker Compose 2.0+
- Git (to clone the repository)
- Basic understanding of notification providers (Telegram bot token or SMTP credentials)

## Quick Start with Docker Compose

### 1. Clone and Navigate

```bash
git clone https://github.com/yourusername/notimulti.git
cd notimulti
```

### 2. Create Provider Configuration

Create a Telegram provider configuration file:

```bash
mkdir -p configs
cat > configs/telegram-example.json <<EOF
{
  "id": "telegram-example",
  "type": "telegram",
  "enabled": true,
  "config": {
    "bot_token": "YOUR_BOT_TOKEN_HERE",
    "default_chat_id": "YOUR_CHAT_ID_HERE",
    "parse_mode": "Markdown",
    "timeout_seconds": 5
  }
}
EOF
```

**Getting Telegram Credentials**:
1. Talk to [@BotFather](https://t.me/BotFather) on Telegram
2. Send `/newbot` and follow instructions to get your `bot_token`
3. Get your chat ID by messaging [@userinfobot](https://t.me/userinfobot)

Or create an Email provider:

```bash
cat > configs/email-example.json <<EOF
{
  "id": "email-example",
  "type": "email",
  "enabled": true,
  "config": {
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from_address": "your-email@gmail.com",
    "from_name": "Notimulti Server",
    "use_tls": true,
    "timeout_seconds": 30
  }
}
EOF
```

### 3. Start the Server

```bash
docker-compose up -d
```

This will:
- Build the Docker image with backend and frontend
- Start the notification server on http://localhost:8080
- Mount `configs/` directory for provider configurations
- Mount `data/` directory for SQLite database persistence

### 4. Verify It's Running

```bash
# Check health endpoint
curl http://localhost:8080/api/v1/health

# Expected output:
# {"status":"ok","version":"2.0.0","timestamp":"2025-11-06T10:30:00Z"}

# List providers
curl http://localhost:8080/api/v1/providers

# Expected output:
# {"providers":[{"id":"telegram-example","type":"telegram","status":"active",...}]}
```

### 5. Open the Web UI

Navigate to http://localhost:8080 in your browser. You should see:
- Dashboard with your configured provider(s)
- "Test" button next to each provider
- Notification history page (empty initially)

### 6. Test Your Provider

**Via UI**:
1. Click the "Test" button next to your provider
2. Wait for the result (success or error message)
3. Check your Telegram chat or email inbox for the test message

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/providers/telegram-example/test

# Expected output:
# {"result":"success","tested_at":"2025-11-06T10:30:00Z","message":"Test notification sent successfully"}
```

### 7. Send a Real Notification

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "telegram-example",
    "recipient": "YOUR_CHAT_ID_HERE",
    "message": "Hello from notimulti! 🚀",
    "priority": "normal"
  }'

# Expected output:
# {"id":1,"status":"pending","created_at":"2025-11-06T10:30:00Z"}
```

### 8. View Notification History

**Via UI**:
- Navigate to http://localhost:8080/history
- See all sent notifications with status, timestamps, errors
- Use filters to narrow results by provider, status, date

**Via API**:
```bash
curl "http://localhost:8080/api/v1/notifications/history?page_size=10"

# Returns paginated notification history
```

## Docker Compose Configuration

The `deploy/docker-compose.yml` file contains:

```yaml
version: '3.8'

services:
  notimulti:
    build:
      context: .
      dockerfile: Dockerfile
    image: notimulti:latest
    container_name: notimulti
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - LOG_LEVEL=info
      - CONFIG_DIR=/app/configs
      - DB_PATH=/app/data/notifications.db
      - LOG_RETENTION_DAYS=90
    volumes:
      - ./configs:/app/configs:ro
      - ./data:/app/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

**Key Configuration**:
- `configs/` mounted read-only for security
- `data/` mounted for database persistence
- Health check monitors server availability
- `restart: unless-stopped` ensures automatic recovery

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | HTTP port for API and UI |
| `LOG_LEVEL` | info | Log level (debug, info, warn, error) |
| `CONFIG_DIR` | /app/configs | Provider configuration directory |
| `DB_PATH` | /app/data/notifications.db | SQLite database file path |
| `LOG_RETENTION_DAYS` | 90 | Days to keep notification logs |

## Directory Structure

```text
.
├── configs/                  # Provider configurations
│   ├── telegram-example.json
│   └── email-example.json
├── data/                     # Database persistence
│   └── notifications.db      # SQLite database (auto-created)
├── deploy/
│   └── docker-compose.yml    # Docker Compose config
└── Dockerfile                # Multi-stage build definition
```

## Common Tasks

### Add a New Provider

1. Create config file in `configs/`:
   ```bash
   cat > configs/telegram-alerts.json <<EOF
   {
     "id": "telegram-alerts",
     "type": "telegram",
     "enabled": true,
     "config": {
       "bot_token": "YOUR_BOT_TOKEN",
       "default_chat_id": "YOUR_CHAT_ID"
     }
   }
   EOF
   ```

2. Server detects new file within 30 seconds (no restart needed)
3. Check UI or API to verify provider is loaded

### Update Provider Configuration

1. Edit the config file:
   ```bash
   nano configs/telegram-example.json
   ```

2. Save changes - server reloads within 30 seconds

### View Logs

```bash
# Follow logs in real-time
docker-compose logs -f notimulti

# View last 100 lines
docker-compose logs --tail=100 notimulti

# Filter for errors
docker-compose logs notimulti | grep ERROR
```

### Backup Database

```bash
# Stop the container
docker-compose stop notimulti

# Backup database
cp data/notifications.db data/notifications.db.backup-$(date +%Y%m%d)

# Restart container
docker-compose start notimulti
```

### Clean Old Logs

The server automatically cleans logs older than retention period (default 90 days) on startup. To manually trigger:

```bash
# Connect to running container
docker-compose exec notimulti sh

# Run manual cleanup (if implemented as CLI command)
./server cleanup-logs --days 30

# Or restart container to trigger startup cleanup
docker-compose restart notimulti
```

## Troubleshooting

### Provider Status Shows "error"

Check the error message in the UI or via API:
```bash
curl http://localhost:8080/api/v1/providers/telegram-example
```

Common issues:
- **Invalid bot token**: Verify token with @BotFather
- **Invalid chat ID**: Check chat ID format (should start with `-` for groups)
- **Network timeout**: Check firewall/network connectivity

### Test Fails but Configuration Looks Correct

Enable debug logging:
```bash
# Edit docker-compose.yml and change LOG_LEVEL=debug
docker-compose up -d

# Check logs for detailed error
docker-compose logs -f notimulti
```

### Database Locked Errors

SQLite uses WAL mode for concurrent access, but if you see locking errors:
```bash
# Check if database file has correct permissions
docker-compose exec notimulti ls -la /app/data/

# If needed, fix permissions
docker-compose exec notimulti chown -R notimulti:notimulti /app/data/
```

### Container Fails to Start

```bash
# Check container logs
docker-compose logs notimulti

# Check if ports are available
lsof -i :8080

# Verify Docker resources
docker info
```

## Next Steps

- **Production Deployment**: See Kubernetes quickstart in `deploy/k8s/README.md`
- **API Documentation**: Full OpenAPI spec at `specs/002-enhanced-deployment/contracts/openapi.yaml`
- **CI/CD Setup**: Configure GitHub Actions workflow with your Docker Hub credentials
- **Monitoring**: Export logs to your logging system (JSON format on stdout)
- **Scaling**: For high volume, consider multiple instances with external database

## Production Checklist

Before deploying to production:

- [ ] Replace example credentials with real provider tokens
- [ ] Set `LOG_LEVEL=info` (not debug)
- [ ] Configure log aggregation (Elasticsearch, CloudWatch, etc.)
- [ ] Set up database backups (volume snapshots or periodic exports)
- [ ] Enable HTTPS with reverse proxy (nginx, Traefik, Ingress)
- [ ] Configure monitoring and alerting
- [ ] Review and adjust `LOG_RETENTION_DAYS` based on compliance requirements
- [ ] Set resource limits in Docker Compose (memory, CPU)
- [ ] Configure provider-specific rate limits if needed
- [ ] Document runbook for common operational tasks

## Getting Help

- **Issues**: Check GitHub Issues for known problems
- **Logs**: Always include relevant log snippets when reporting issues
- **Configuration**: Validate config files with JSON linter before reporting errors
- **Testing**: Use the built-in provider test feature to isolate provider configuration issues
