# Data Model: Centralized Notification Server

**Date**: 2025-11-03  
**Feature**: 001-notification-server  
**Purpose**: Define entities, relationships, validation rules, and state transitions

## Core Entities

### 1. Provider Configuration

**Description**: Represents a notification provider instance configuration loaded from a file.

**Fields**:
- `id` (string, required): Unique identifier for this provider instance (e.g., "telegram-alerts", "email-prod")
- `type` (string, required): Provider type identifier ("telegram", "email")
- `enabled` (boolean, default: true): Whether this provider is active
- `config` (object, required): Provider-specific configuration

**Telegram-Specific Config**:
- `bot_token` (string, required): Telegram bot API token
- `default_chat_id` (string, required): Default chat/channel ID for notifications
- `parse_mode` (string, optional): Message parse mode ("Markdown", "HTML", default: "Markdown")
- `timeout_seconds` (integer, default: 5): Request timeout in seconds

**Email-Specific Config**:
- `smtp_host` (string, required): SMTP server hostname
- `smtp_port` (integer, required): SMTP server port (25, 587, 465)
- `username` (string, required): SMTP authentication username
- `password` (string, required): SMTP authentication password
- `from_address` (string, required): Sender email address
- `from_name` (string, optional): Sender display name
- `use_tls` (boolean, default: true): Use TLS/STARTTLS
- `timeout_seconds` (integer, default: 30): SMTP operation timeout

**Validation Rules**:
- `id` must be unique across all providers
- `id` must match pattern: `^[a-z0-9-]+$` (lowercase, numbers, hyphens only)
- `type` must be one of: ["telegram", "email"]
- `bot_token` must not be empty (Telegram)
- `default_chat_id` must match Telegram chat ID format (Telegram)
- `smtp_host` must be valid hostname (Email)
- `smtp_port` must be 1-65535 (Email)
- `from_address` must be valid email format (Email)
- All `timeout_seconds` must be positive integers

**File Format** (JSON):
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

**State Transitions**:
1. **Created**: Config file added to configs/ directory
2. **Loading**: Config file read and parsed
3. **Validating**: Config structure and values validated
4. **Active**: Provider initialized and ready to send
5. **Error**: Validation or initialization failed
6. **Disabled**: Provider exists but `enabled: false`
7. **Removed**: Config file deleted

**Relationships**:
- One Provider Configuration can generate many Notification Requests
- One Provider Configuration is stored in one Configuration File

---

### 2. Provider Instance (Runtime)

**Description**: Runtime representation of a loaded and initialized provider.

**Fields**:
- `id` (string): Same as configuration ID
- `type` (string): Provider type
- `status` (string): Current operational status
- `last_updated` (timestamp): When configuration was last loaded
- `error_message` (string, nullable): Current error if status is "error"
- `config_checksum` (string): Hash of configuration for change detection

**Status Values**:
- `"active"`: Provider is operational and ready to send
- `"error"`: Provider has configuration or connectivity issues
- `"disabled"`: Provider is intentionally disabled
- `"initializing"`: Provider is being loaded

**State Transitions**:
```
initializing → active (on successful init)
initializing → error (on init failure)
active → error (on send failure)
active → disabled (on config change: enabled=false)
error → active (on config fix + reload)
disabled → active (on config change: enabled=true)
* → removed (on config file deletion)
```

**Validation Rules**:
- `status` must be one of: ["active", "error", "disabled", "initializing"]
- `error_message` required when status is "error"
- `last_updated` must be valid RFC3339 timestamp

**Relationships**:
- Corresponds to one Provider Configuration
- Generates many Notification Requests

---

### 3. Notification Request

**Description**: A request to send a notification through a specific provider.

**Fields**:
- `id` (string, generated): Unique notification request ID (UUID)
- `provider_id` (string, required): Target provider instance ID
- `recipient` (string, required): Recipient identifier (chat ID, email address)
- `message` (string, required): Notification message content
- `subject` (string, optional): Message subject (Email only)
- `metadata` (object, optional): Additional key-value pairs
- `priority` (string, default: "normal"): Priority level
- `created_at` (timestamp): When request was received
- `status` (string): Current delivery status
- `error_message` (string, nullable): Error details if delivery failed
- `delivered_at` (timestamp, nullable): When notification was successfully delivered
- `attempts` (integer, default: 0): Number of delivery attempts

**Priority Values**:
- `"low"`: Best-effort delivery
- `"normal"`: Standard delivery (default)
- `"high"`: Expedited delivery

**Status Values**:
- `"pending"`: Queued for delivery
- `"sending"`: Currently being sent
- `"delivered"`: Successfully delivered
- `"failed"`: Delivery failed permanently
- `"retrying"`: Temporary failure, will retry

**Validation Rules**:
- `provider_id` must reference an existing provider
- `recipient` must be non-empty
- `message` must be non-empty
- `message` length must be ≤ 4096 characters (Telegram limit)
- `recipient` must be valid email format (Email provider)
- `recipient` must be valid Telegram chat ID format (Telegram provider)
- `priority` must be one of: ["low", "normal", "high"]
- `status` must be one of: ["pending", "sending", "delivered", "failed", "retrying"]

**State Transitions**:
```
pending → sending (when delivery starts)
sending → delivered (on successful send)
sending → failed (on permanent failure)
sending → retrying (on transient failure)
retrying → sending (on retry attempt)
retrying → failed (after max retries)
```

**API Request Format** (JSON):
```json
{
  "provider_id": "telegram-alerts",
  "recipient": "-1001234567890",
  "message": "Server alert: High CPU usage detected",
  "metadata": {
    "source": "monitoring-system",
    "severity": "warning"
  },
  "priority": "high"
}
```

**API Response Format** (JSON):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "created_at": "2025-11-03T10:30:00Z"
}
```

**Relationships**:
- Belongs to one Provider Instance
- Stored in Notification Log (SQLite)

---

### 4. Notification Log (Persistent)

**Description**: SQLite table storing notification history for audit and troubleshooting.

**Schema**:
```sql
CREATE TABLE notification_logs (
    id TEXT PRIMARY KEY,
    provider_id TEXT NOT NULL,
    provider_type TEXT NOT NULL,
    recipient TEXT NOT NULL,
    message TEXT NOT NULL,
    subject TEXT,
    metadata TEXT,  -- JSON string
    priority TEXT NOT NULL DEFAULT 'normal',
    status TEXT NOT NULL,
    error_message TEXT,
    attempts INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,  -- ISO8601
    delivered_at TEXT,  -- ISO8601
    INDEX idx_provider_id (provider_id),
    INDEX idx_created_at (created_at),
    INDEX idx_status (status)
);
```

**Indexes**:
- `idx_provider_id`: For querying notifications by provider
- `idx_created_at`: For time-range queries and cleanup
- `idx_status`: For filtering by delivery status

**Retention Policy**:
- Keep logs for 30 days
- Implement periodic cleanup (daily cron job or startup check)
- Archive old logs to external storage if needed

**Validation Rules**:
- `id` must be unique UUID
- Foreign key: `provider_id` references active provider (soft constraint)
- `created_at` must be valid ISO8601 timestamp
- `delivered_at` must be after `created_at` if present

**Relationships**:
- Each row corresponds to one Notification Request
- Multiple rows can reference same Provider Instance

---

### 5. Configuration File

**Description**: JSON file in `configs/` directory representing a provider configuration.

**Fields**:
- `filename` (string): File name (e.g., "telegram-alerts.json")
- `path` (string): Absolute file path
- `checksum` (string): SHA256 hash of file content
- `last_modified` (timestamp): File system modification time
- `watch_status` (string): File watch state

**Watch Status Values**:
- `"watching"`: Currently being monitored
- `"changed"`: Change detected, pending reload
- `"removed"`: File deleted
- `"error"`: File read or parse error

**State Transitions**:
```
watching → changed (on file modification)
watching → removed (on file deletion)
changed → watching (after successful reload)
changed → error (on reload failure)
error → watching (after manual fix + reload)
removed → (deleted from watch list)
```

**Validation Rules**:
- `filename` must match pattern: `^[a-z0-9-]+\.(json)$`
- File size must be < 1MB (prevent abuse)
- Must contain valid JSON
- JSON must match Provider Configuration schema

**Relationships**:
- Corresponds to one Provider Configuration
- Changes trigger Provider Instance updates

---

## Entity Relationships Diagram

```
┌─────────────────────┐
│ Configuration File  │
│ (File System)       │
└──────────┬──────────┘
           │ 1:1
           ▼
┌─────────────────────┐
│ Provider Config     │
│ (Parsed JSON)       │
└──────────┬──────────┘
           │ 1:1
           ▼
┌─────────────────────┐       1:N        ┌─────────────────────┐
│ Provider Instance   │◄─────────────────┤ Notification Request│
│ (Runtime)           │                  │ (API Input)         │
└─────────────────────┘                  └──────────┬──────────┘
                                                    │ 1:1
                                                    ▼
                                         ┌─────────────────────┐
                                         │ Notification Log    │
                                         │ (SQLite Table)      │
                                         └─────────────────────┘
```

## Validation Summary

### Configuration Validation (Load Time)
1. JSON syntax valid
2. Required fields present
3. Field types correct
4. Provider-specific config valid
5. ID unique across all providers
6. Sensitive fields not logged

### Request Validation (API Time)
1. Provider ID exists and active
2. Required fields present (provider_id, recipient, message)
3. Message length within limits
4. Recipient format valid for provider type
5. Priority value valid
6. Metadata JSON valid (if present)

### Runtime Validation (Continuous)
1. Provider connectivity checks
2. Configuration file integrity
3. Database consistency
4. Log retention enforcement

## Error Handling

### Configuration Errors
- **Invalid JSON**: Log error, skip file, continue with other configs
- **Missing Fields**: Log error with field names, skip file
- **Duplicate ID**: Log error, reject new config, keep existing
- **Invalid Credentials**: Mark provider as "error" status, log details

### Request Errors
- **Unknown Provider**: Return 404 with provider ID
- **Validation Failure**: Return 400 with field-level errors
- **Send Failure**: Mark as "failed", log error with full context
- **Timeout**: Mark as "retrying", schedule retry

### Storage Errors
- **DB Connection**: Log error, continue in memory, retry periodically
- **DB Write Failure**: Log error, continue processing (don't block sends)
- **DB Full**: Implement retention policy, alert operator

## State Consistency

### Provider Registry
- In-memory map: `provider_id → Provider Instance`
- Thread-safe access via `sync.RWMutex`
- Atomic updates on configuration reload

### Configuration Reload
1. Read file
2. Parse and validate
3. Create new provider instance
4. Swap atomically in registry
5. Close old provider instance gracefully
6. Log change event

### Notification Queue
- Buffered channel: `chan NotificationRequest`
- Worker pool processes requests concurrently
- Failed requests logged but don't block queue

## Performance Considerations

- **Provider Registry**: O(1) lookup by ID
- **Configuration Reload**: < 5 seconds (requirement NFR-003)
- **Notification Send**: < 2 seconds p95 (requirement NFR-002)
- **Database Writes**: Batched for efficiency
- **File Watching**: Debounced to prevent reload storms
