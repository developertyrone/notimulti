package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/developertyrone/notimulti/internal/providers"
)

// NotificationLogger handles asynchronous logging of notifications to database
type NotificationLogger struct {
	db         *sql.DB
	buffer     []LogEntry
	bufferMu   sync.Mutex
	stmt       *sql.Stmt
	ticker     *time.Ticker
	done       chan struct{}
	bufferSize int
	flushTime  time.Duration
}

// LogEntry represents a notification log entry
type LogEntry struct {
	Notification *providers.Notification
	Status       string
	ErrorMessage string
	ProviderType string
	Attempts     int
}

// NewNotificationLogger creates a new notification logger
func NewNotificationLogger(db *sql.DB) (*NotificationLogger, error) {
	// Prepare the insert statement (omit id as it's AUTOINCREMENT)
	stmt, err := db.Prepare(`
		INSERT INTO notification_logs (
			provider_id, provider_type, recipient, message, subject,
			metadata, priority, status, error_message, attempts, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	logger := &NotificationLogger{
		db:         db,
		buffer:     make([]LogEntry, 0, 100),
		stmt:       stmt,
		ticker:     time.NewTicker(5 * time.Second),
		done:       make(chan struct{}),
		bufferSize: 100,
		flushTime:  5 * time.Second,
	}

	// Start background flusher
	go logger.flushLoop()

	return logger, nil
}

// LogNotification adds a notification to the log buffer
func (nl *NotificationLogger) LogNotification(notif *providers.Notification, status, errorMsg, providerType string, attempts int) error {
	if notif == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	entry := LogEntry{
		Notification: notif,
		Status:       status,
		ErrorMessage: errorMsg,
		ProviderType: providerType,
		Attempts:     attempts,
	}

	nl.bufferMu.Lock()
	nl.buffer = append(nl.buffer, entry)
	shouldFlush := len(nl.buffer) >= nl.bufferSize
	nl.bufferMu.Unlock()

	// Flush immediately if buffer is full
	if shouldFlush {
		return nl.Flush()
	}

	return nil
}

// Flush writes all buffered entries to the database
func (nl *NotificationLogger) Flush() error {
	nl.bufferMu.Lock()
	if len(nl.buffer) == 0 {
		nl.bufferMu.Unlock()
		return nil
	}

	// Copy buffer and clear it
	entries := make([]LogEntry, len(nl.buffer))
	copy(entries, nl.buffer)
	nl.buffer = nl.buffer[:0]
	nl.bufferMu.Unlock()

	// Write to database
	for _, entry := range entries {
		if err := nl.writeEntry(entry); err != nil {
			// Log error but don't block - this is best effort
			fmt.Printf("Failed to write notification log: %v\n", err)
		}
	}

	return nil
}

// writeEntry writes a single entry to the database
func (nl *NotificationLogger) writeEntry(entry LogEntry) error {
	// Serialize metadata to JSON
	var metadataJSON []byte
	if entry.Notification.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(entry.Notification.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	_, err := nl.stmt.Exec(
		entry.Notification.ProviderID,
		entry.ProviderType,
		entry.Notification.Recipient,
		entry.Notification.Message,
		entry.Notification.Subject,
		string(metadataJSON),
		entry.Notification.Priority,
		entry.Status,
		entry.ErrorMessage,
		entry.Attempts,
		entry.Notification.Timestamp,
	)

	return err
}

// flushLoop periodically flushes the buffer
func (nl *NotificationLogger) flushLoop() {
	for {
		select {
		case <-nl.ticker.C:
			nl.Flush()
		case <-nl.done:
			// Final flush before shutdown
			nl.Flush()
			return
		}
	}
}

// Close stops the logger and flushes remaining entries
func (nl *NotificationLogger) Close() error {
	close(nl.done)
	nl.ticker.Stop()

	// Final flush
	nl.Flush()

	// Close prepared statement
	if nl.stmt != nil {
		return nl.stmt.Close()
	}

	return nil
}
