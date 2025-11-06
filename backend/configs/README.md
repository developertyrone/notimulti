# Provider Configuration Guide

## Why Don't I See Providers in the Dashboard?

Providers appear in the dashboard only if they:
1. ✅ Have valid JSON structure with `"config": {...}` (not `"telegram": {...}`)
2. ✅ Have `"enabled": true` field
3. ✅ Pass validation (required fields present)
4. ✅ **Successfully initialize** (this is the key requirement!)

## Current Status

### Email Provider (email-demo.json)
- ✅ Loads successfully
- ✅ Appears in dashboard
- ⚠️  Shows "error" status because `smtp.example.com` is not a real SMTP server
- **To fix**: Replace with real SMTP credentials (see below)

### Telegram Provider (telegram-demo.json)
- ❌ Currently disabled (`"enabled": false`)
- **Reason**: The Telegram Bot API validates tokens during initialization
- Fake tokens like `"1234567890:ABCdefGHI..."` are rejected immediately
- **To fix**: Use a real Telegram bot token (see below)

## How to Add Real Credentials

### For Telegram Bot

1. **Create a bot**:
   - Open Telegram and search for `@BotFather`
   - Send `/newbot` and follow instructions
   - Copy the bot token (format: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

2. **Get chat ID**:
   - Start a chat with your bot or add it to a group
   - Send a message to the bot
   - Visit: `https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates`
   - Look for `"chat":{"id": -1001234567890}` in the response
   - Copy the chat ID (can be positive or negative number)

3. **Update config**:
   ```json
   {
     "id": "telegram-prod",
     "type": "telegram",
     "enabled": true,
     "config": {
       "bot_token": "YOUR_REAL_BOT_TOKEN_HERE",
       "default_chat_id": "YOUR_CHAT_ID_HERE",
       "parse_mode": "HTML"
     }
   }
   ```

4. Save the file - the server will auto-reload and the provider will appear in the dashboard!

### For Email/SMTP

#### Gmail Example

1. **Enable 2-factor authentication** on your Google account

2. **Generate app password**:
   - Visit: https://myaccount.google.com/apppasswords
   - Create a new app password
   - Copy the 16-character password

3. **Update config**:
   ```json
   {
     "id": "email-prod",
     "type": "email",
     "enabled": true,
     "config": {
       "host": "smtp.gmail.com",
       "port": 587,
       "username": "your-email@gmail.com",
       "password": "your-16-char-app-password",
       "from": "your-email@gmail.com",
       "use_tls": true
     }
   }
   ```

#### Other SMTP Providers

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
