# Notimulti Server Quick Start

## Backend Server is Running! ✅

The notification server is successfully running on `http://localhost:8080`

### API Endpoints

- **Health Check**: `GET /api/v1/health`
- **List Providers**: `GET /api/v1/providers`
- **Get Provider**: `GET /api/v1/providers/:id`
- **Send Notification**: `POST /api/v1/notifications`

### Testing the API

```bash
# Check server health
curl http://localhost:8080/api/v1/health

# List active providers
curl http://localhost:8080/api/v1/providers

# Send a notification (once providers are configured)
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "telegram-demo",
    "message": "Hello from notimulti!",
    "recipient": "optional-chat-id"
  }'
```

### Configuring Providers

Provider configurations are stored in `./configs/` directory. The server watches this directory for changes and automatically reloads configurations.

#### Telegram Provider

Create `configs/telegram-bot.json`:

```json
{
  "id": "telegram-bot",
  "type": "telegram",
  "telegram": {
    "bot_token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
    "default_chat_id": "-1001234567890",
    "parse_mode": "HTML"
  }
}
```

**How to get credentials:**
1. Create a bot with [@BotFather](https://t.me/BotFather)
2. Send `/newbot` and follow instructions
3. Copy the bot token
4. Add your bot to a group/channel or chat with it directly
5. Get chat ID from `https://api.telegram.org/bot<TOKEN>/getUpdates`

#### Email Provider

Create `configs/email-smtp.json`:

```json
{
  "id": "email-smtp",
  "type": "email",
  "email": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from": "your-email@gmail.com",
    "use_tls": true
  }
}
```

**For Gmail:**
1. Enable 2-factor authentication
2. Generate an app password at: https://myaccount.google.com/apppasswords
3. Use the app password (not your regular password)

**To send to multiple recipients**, specify them in the notification request:
```json
{
  "provider_id": "email-smtp",
  "recipient": "user@example.com",
  "subject": "Test Notification",
  "message": "This is a test email"
}
```

### Current Status

- ✅ Server running on port 8080
- ✅ Database initialized (`./notimulti.db`)
- ✅ Configuration watcher active (`./configs/`)
- ⚠️  No providers configured yet (add valid credentials to `configs/*.json`)

### Next Steps

1. **Configure at least one provider** by adding valid credentials to the JSON files in `configs/`
2. **Test the notification endpoint** with curl or the web UI
3. **Start the frontend** to use the web dashboard:
   ```bash
   cd ../frontend
   npm run dev
   # Open http://localhost:5173
   ```

### Troubleshooting

- **Provider not loading?** Check `server.log` for validation errors
- **Telegram bot not working?** Ensure bot token is correct and bot is added to chat
- **Email not sending?** Check SMTP credentials and enable "less secure apps" if needed
- **CORS errors?** Frontend should run on `localhost:5173` (default Vite port)

### Server Management

```bash
# Check if server is running
ps aux | grep notimulti-server

# Stop the server
kill $(pgrep -f notimulti-server)

# Start server in background
./notimulti-server > server.log 2>&1 &

# Watch server logs
tail -f server.log
```

### Production Deployment

See the main `README.md` for production deployment instructions with systemd and nginx.
