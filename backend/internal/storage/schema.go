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
    delivered_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_notification_logs_provider_id 
    ON notification_logs(provider_id);

CREATE INDEX IF NOT EXISTS idx_notification_logs_created_at 
    ON notification_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_notification_logs_status 
    ON notification_logs(status);

CREATE INDEX IF NOT EXISTS idx_notification_logs_provider_type 
    ON notification_logs(provider_type);
`

// Status constants for notification logs
const (
	StatusPending  = "pending"
	StatusSent     = "sent"
	StatusFailed   = "failed"
	StatusRetrying = "retrying"
)
