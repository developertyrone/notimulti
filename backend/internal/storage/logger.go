package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/developertyrone/notimulti/internal/providers"
)

// NotificationLogger handles asynchronous logging of notifications to database
type NotificationLogger struct {
	db          *sql.DB
	logQueue    chan LogEntry
	flushTicker *time.Ticker
	wg          sync.WaitGroup
	mu          sync.Mutex
	closed      bool
}

// LogEntry represents a notification log entry
type LogEntry struct {
	Notification *providers.Notification
	Status       string
	ErrorMessage string
	ProviderType string
	Attempts     int
	DeliveredAt  string // ISO8601 timestamp
	IsTest       bool
}

// NewNotificationLogger creates a new notification logger with buffered channel
func NewNotificationLogger(db *sql.DB) (*NotificationLogger, error) {
	logger := &NotificationLogger{
		db:          db,
		logQueue:    make(chan LogEntry, 1000), // Buffer 1000 entries per research.md
		flushTicker: time.NewTicker(5 * time.Second),
		closed:      false,
	}

	// Start background worker
	logger.wg.Add(1)
	go logger.worker()

	return logger, nil
}

// worker processes log entries from the queue in batches
func (nl *NotificationLogger) worker() {
	defer nl.wg.Done()
	batch := make([]LogEntry, 0, 100) // 100 entries per batch per research.md

	for {
		select {
		case entry, ok := <-nl.logQueue:
			if !ok {
				// Channel closed, flush remaining and exit
				if len(batch) > 0 {
					nl.flushBatch(batch)
				}
				return
			}
			batch = append(batch, entry)
			if len(batch) >= 100 {
				nl.flushBatch(batch)
				batch = batch[:0]
			}
		case <-nl.flushTicker.C:
			if len(batch) > 0 {
				nl.flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch writes a batch of log entries to the database using a transaction
func (nl *NotificationLogger) flushBatch(batch []LogEntry) error {
	if len(batch) == 0 {
		return nil
	}

	// Start transaction for batch insert
	tx, err := nl.db.Begin()
	if err != nil {
		log.Printf("ERROR: Failed to start transaction for notification logs: %v", err)
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statement within transaction
	stmt, err := tx.Prepare(`
		INSERT INTO notification_logs (
			provider_id, provider_type, recipient, message, subject,
			metadata, priority, status, error_message, attempts, created_at, delivered_at, is_test
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Printf("ERROR: Failed to prepare statement: %v", err)
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert all entries in batch
	for _, entry := range batch {
		if err := nl.writeEntryWithStmt(stmt, entry); err != nil {
			log.Printf("ERROR: Failed to write notification log entry: %v", err)
			// Continue with other entries - best effort
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("ERROR: Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// writeEntryWithStmt writes a single entry using the provided statement
func (nl *NotificationLogger) writeEntryWithStmt(stmt *sql.Stmt, entry LogEntry) error {
	// Serialize metadata to JSON
	var metadataJSON []byte
	if entry.Notification.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(entry.Notification.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	// Convert IsTest to integer (SQLite boolean)
	isTestInt := 0
	if entry.IsTest {
		isTestInt = 1
	}

	// Handle nullable delivered_at
	var deliveredAt interface{}
	if entry.DeliveredAt != "" {
		deliveredAt = entry.DeliveredAt
	} else {
		deliveredAt = nil
	}

	_, err := stmt.Exec(
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
		deliveredAt,
		isTestInt,
	)

	return err
}

// Log adds a log entry to the queue (non-blocking)
func (nl *NotificationLogger) Log(entry LogEntry) {
	nl.mu.Lock()
	if nl.closed {
		nl.mu.Unlock()
		log.Printf("WARN: Attempted to log to closed NotificationLogger")
		return
	}
	nl.mu.Unlock()

	select {
	case nl.logQueue <- entry:
		// Logged successfully
	default:
		// Queue full, log error but don't block
		log.Printf("ERROR: Notification log queue full, dropping entry for provider %s", entry.Notification.ProviderID)
	}
}

// Close gracefully shuts down the logger with 30s timeout per research.md
func (nl *NotificationLogger) Close() error {
	nl.mu.Lock()
	if nl.closed {
		nl.mu.Unlock()
		return nil
	}
	nl.closed = true
	nl.mu.Unlock()

	// Stop ticker
	nl.flushTicker.Stop()

	// Close channel to signal worker to stop
	close(nl.logQueue)

	// Wait for worker to finish with timeout
	done := make(chan struct{})
	go func() {
		nl.wg.Wait()
		close(done)
	}()

	// 30s timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for logger to close")
	}
}

