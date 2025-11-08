package storage

// Schema contains the SQL DDL statements for creating the database schema
const Schema = `
CREATE TABLE IF NOT EXISTS notification_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_id TEXT NOT NULL,
    provider_type TEXT NOT NULL,
    recipient TEXT NOT NULL,
    message TEXT NOT NULL,
    subject TEXT,
    metadata TEXT,
    priority TEXT DEFAULT 'normal',
    status TEXT NOT NULL,
    error_message TEXT,
    attempts INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    delivered_at DATETIME,
    is_test INTEGER NOT NULL DEFAULT 0
);

-- Composite indexes for common query patterns (Phase 2 optimization)
CREATE INDEX IF NOT EXISTS idx_provider_created 
    ON notification_logs(provider_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_status_created 
    ON notification_logs(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_type_created 
    ON notification_logs(provider_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_created_id 
    ON notification_logs(created_at DESC, id DESC);
`

// Status constants for notification logs
const (
	StatusPending  = "pending"
	StatusSent     = "sent"
	StatusFailed   = "failed"
	StatusRetrying = "retrying"
)
