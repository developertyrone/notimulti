package storage

import (
	"database/sql"
)

// Repository handles database queries for notification history
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// Ping checks database connectivity using a simple SELECT 1 query
func (r *Repository) Ping() error {
	_, err := r.db.Exec("SELECT 1")
	return err
}

// HistoryFilters represents query filters for notification history
type HistoryFilters struct {
	ProviderID   string
	ProviderType string
	Status       string
	DateFrom     string
	DateTo       string
	IncludeTests bool
	Cursor       int
	PageSize     int
	SortOrder    string
}

// NotificationLogEntry represents a notification log record from the database
type NotificationLogEntry struct {
	ID           int            `json:"id"`
	ProviderID   string         `json:"provider_id"`
	ProviderType string         `json:"provider_type"`
	Recipient    string         `json:"recipient"`
	Message      string         `json:"message"`
	Subject      sql.NullString `json:"subject"`
	Metadata     sql.NullString `json:"metadata"`
	Priority     string         `json:"priority"`
	Status       string         `json:"status"`
	ErrorMessage sql.NullString `json:"error_message"`
	Attempts     int            `json:"attempts"`
	CreatedAt    string         `json:"created_at"`
	DeliveredAt  sql.NullString `json:"delivered_at"`
	IsTest       bool           `json:"is_test"`
}

// buildHistoryQuery constructs the SQL query with filters
func (r *Repository) buildHistoryQuery(filters HistoryFilters) (string, []interface{}) {
	query := `SELECT * FROM notification_logs WHERE 1=1`
	args := []interface{}{}

	if filters.ProviderID != "" {
		query += " AND provider_id = ?"
		args = append(args, filters.ProviderID)
	}
	if filters.ProviderType != "" {
		query += " AND provider_type = ?"
		args = append(args, filters.ProviderType)
	}
	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}
	if filters.DateFrom != "" {
		query += " AND created_at >= ?"
		args = append(args, filters.DateFrom)
	}
	if filters.DateTo != "" {
		query += " AND created_at <= ?"
		args = append(args, filters.DateTo)
	}
	if !filters.IncludeTests {
		query += " AND is_test = 0"
	}
	if filters.Cursor > 0 {
		query += " AND id < ?"
		args = append(args, filters.Cursor)
	}

	// Sort by created_at and id for cursor pagination
	query += " ORDER BY created_at " + filters.SortOrder + ", id " + filters.SortOrder
	query += " LIMIT ?"
	args = append(args, filters.PageSize+1) // Fetch one extra to determine "has more"

	return query, args
}

// GetNotificationHistory retrieves notification history with filters and pagination
func (r *Repository) GetNotificationHistory(filters HistoryFilters) ([]NotificationLogEntry, *int, error) {
	query, args := r.buildHistoryQuery(filters)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var entries []NotificationLogEntry
	for rows.Next() {
		var entry NotificationLogEntry
		var isTestInt int
		err := rows.Scan(
			&entry.ID,
			&entry.ProviderID,
			&entry.ProviderType,
			&entry.Recipient,
			&entry.Message,
			&entry.Subject,
			&entry.Metadata,
			&entry.Priority,
			&entry.Status,
			&entry.ErrorMessage,
			&entry.Attempts,
			&entry.CreatedAt,
			&entry.DeliveredAt,
			&isTestInt,
		)
		if err != nil {
			return nil, nil, err
		}
		entry.IsTest = isTestInt != 0
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	// Determine if there are more results
	var nextCursor *int
	if len(entries) > filters.PageSize {
		// Remove the extra entry
		entries = entries[:filters.PageSize]
		lastID := entries[len(entries)-1].ID
		nextCursor = &lastID
	}

	return entries, nextCursor, nil
}

// GetNotificationByID retrieves a specific notification by ID
func (r *Repository) GetNotificationByID(id int) (*NotificationLogEntry, error) {
	query := `SELECT * FROM notification_logs WHERE id = ?`

	var entry NotificationLogEntry
	var isTestInt int
	err := r.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.ProviderID,
		&entry.ProviderType,
		&entry.Recipient,
		&entry.Message,
		&entry.Subject,
		&entry.Metadata,
		&entry.Priority,
		&entry.Status,
		&entry.ErrorMessage,
		&entry.Attempts,
		&entry.CreatedAt,
		&entry.DeliveredAt,
		&isTestInt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	entry.IsTest = isTestInt != 0
	return &entry, nil
}

// CleanupOldLogs removes notifications older than the retention period
func (r *Repository) CleanupOldLogs(retentionDays int) error {
	query := `DELETE FROM notification_logs WHERE created_at < datetime('now', '-' || ? || ' days')`
	_, err := r.db.Exec(query, retentionDays)
	return err
}
