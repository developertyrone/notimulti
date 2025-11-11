# Provider Configuration Guide

## Quick Start

Sample configuration files are provided in this directory:
- `telegram-example.json` - Telegram bot configuration template
- `email-example.json` - Email/SMTP configuration template

**To use them:**
1. Copy and rename (e.g., `cp telegram-example.json telegram-prod.json`)
2. Replace placeholder values with your real credentials
3. Save - the server auto-detects and loads the provider!

## Configuration Examples

### Telegram Bot

```json
{
  "id": "telegram-main",
  "type": "telegram",
  "enabled": true,
  "config": {
    "bot_token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
    "default_chat_id": "-1001234567890",
    "parse_mode": "HTML",
    "timeout_seconds": 30
  }
}
```

**How to get credentials:**
1. Message `@BotFather` on Telegram and send `/newbot`
2. Follow instructions to get your bot token
3. Start chat with your bot or add to a group
4. Visit `https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates`
5. Find `"chat":{"id": ...}` in the response for your chat ID

### Email (Gmail)

```json
{
  "id": "email-smtp",
  "type": "email",
  "enabled": true,
  "config": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from": "your-email@gmail.com",
    "use_tls": true,
    "timeout_seconds": 30
  }
}
```

**For Gmail:**
1. Enable 2FA on your Google account
2. Visit https://myaccount.google.com/apppasswords
3. Generate app password (16 characters)
4. Use that as the password

### Other SMTP Providers

**SendGrid**:
```json
{
  "host": "smtp.sendgrid.net",
  "port": 587,
  "username": "apikey",
  "password": "YOUR_SENDGRID_API_KEY",
  "from": "noreply@yourdomain.com"
}
```

**Mailgun**:
```json
{
  "host": "smtp.mailgun.org",
  "port": 587,
  "username": "postmaster@your-domain.mailgun.org",
  "password": "YOUR_MAILGUN_PASSWORD",
  "from": "noreply@yourdomain.com"
}
```

**AWS SES**:
```json
{
  "host": "email-smtp.us-east-1.amazonaws.com",
  "port": 587,
  "username": "YOUR_SMTP_USERNAME",
  "password": "YOUR_SMTP_PASSWORD",
  "from": "verified-email@yourdomain.com"
}
```

## Testing Without Real Credentials

If you just want to test the UI without sending real notifications:

1. **Use Mailtrap.io** (free email testing):
   - Sign up at https://mailtrap.io
   - Get SMTP credentials from inbox settings
   - Emails will be captured in Mailtrap instead of being sent

2. **Use a test Telegram bot**:
   - You still need a real bot token from @BotFather
   - But you can use your own chat as the recipient
   - This is safe for testing

## File Watcher

The server watches this `configs/` directory for changes:
- ✅ **Create** new `.json` file → provider appears immediately
- ✅ **Edit** existing file → provider reloads automatically
- ✅ **Delete** file → provider removed from registry
- ✅ **Set `"enabled": false`** → provider removed from active list

**No server restart needed!** Just save your config changes.

## Troubleshooting

### Provider doesn't appear in dashboard

1. **Check server logs**:
   ```bash
   tail -f server.log
   ```

2. **Look for errors**:
   - `"Failed to load configuration"` → JSON syntax error or validation failed
   - `"Failed to create provider"` → Invalid credentials or initialization failed
   - `"Skipping disabled provider"` → Provider has `"enabled": false`

3. **Validate JSON**:
   ```bash
   jq . configs/your-file.json
   ```

### Provider shows "error" status

This means the provider loaded but connectivity check failed:
- **Telegram**: Invalid token, bot deleted, or API unreachable
- **Email**: SMTP host unreachable, wrong port, or auth failed

Check the `error_message` field in the API response for details.

### Changes not picked up

1. **Check file watcher is running**:
   ```bash
   # Should see: "Configuration watcher started"
   tail server.log | grep watcher
   ```

2. **Force reload**:
   ```bash
   # Touch the file to trigger file watcher
   touch configs/your-file.json
   ```

3. **Restart server** (last resort):
   ```bash
   pkill notimulti-server
   ./notimulti-server > server.log 2>&1 &
   ```

## Quick Test

To verify the system is working, enable the email provider with Mailtrap credentials or use a real Telegram bot. The dashboard should update within 1-2 seconds of saving the config file.
