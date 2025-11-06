# Failed Provider Display Feature

## What Changed

Previously, providers that failed to initialize were **not registered** and would not appear in the dashboard at all. This made it difficult to debug configuration issues.

Now, providers with configuration errors **are registered and displayed** in the dashboard with an "error" status and a clear error message.

## Implementation

### New Component: FailedProvider

Created `internal/providers/failed.go` - a stub provider that:
- Implements the Provider interface
- Always returns errors for Send() operations
- Displays initialization error in status
- Allows provider to be visible in UI even when broken

### Modified Components

1. **cmd/server/main.go**:
   - Changed from `continue` (skip) to creating FailedProvider when initialization fails
   - Now registers all providers, even those with errors

2. **internal/config/watcher.go**:
   - Updated `handleCreate()` to register failed providers
   - Updated `handleWrite()` to replace with failed providers if update fails
   - Ensures file watcher also shows errors in UI

## Behavior

### Before
```
Provider fails → Not registered → Invisible in dashboard → User confused
```

### After
```
Provider fails → Registered as FailedProvider → Shows in dashboard with error → User sees problem
```

## Example Error Messages

**Telegram with invalid token**:
```json
{
  "id": "telegram-demo",
  "type": "telegram",
  "status": "error",
  "error_message": "Initialization failed: failed to create bot API: Unauthorized"
}
```

**Email with invalid SMTP host**:
```json
{
  "id": "email-demo",
  "type": "email",
  "status": "error",
  "error_message": "SMTP connectivity check failed: dial tcp: lookup smtp.example.com: no such host"
}
```

## Benefits

1. ✅ **Visibility**: Users can see all configured providers, even broken ones
2. ✅ **Debuggability**: Clear error messages explain what's wrong
3. ✅ **Completeness**: Dashboard shows full picture of system state
4. ✅ **User Experience**: No silent failures or missing providers

## Testing

1. Configure provider with invalid credentials
2. Server starts and registers provider with error status
3. Dashboard shows provider with red/error indicator
4. Error message visible in provider details
5. Fix credentials in config file
6. File watcher detects change
7. Provider automatically transitions from "error" to "active"

This matches standard DevOps practices where observability of failure states is critical.
