package providers

import (
	"context"
	"fmt"
	"net"
	"time"

	"gopkg.in/gomail.v2"
)

// EmailProvider implements the Provider interface for SMTP email
type EmailProvider struct {
	id             string
	config         *EmailConfig
	dailer         *gomail.Dialer
	lastTestAt     *time.Time
	lastTestStatus string
}

// NewEmailProvider creates a new Email provider instance
func NewEmailProvider(id string, config *EmailConfig) (*EmailProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.Host == "" {
		return nil, fmt.Errorf("host is required")
	}

	if config.Port == 0 {
		return nil, fmt.Errorf("port is required")
	}

	if config.From == "" {
		return nil, fmt.Errorf("from address is required")
	}

	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)

	if config.UseTLS {
		dialer.TLSConfig = nil // Use system's root CAs
	}

	return &EmailProvider{
		id:     id,
		config: config,
		dailer: dialer,
	}, nil
}

// Send sends a notification via SMTP email with retry logic
func (ep *EmailProvider) Send(ctx context.Context, notification *Notification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	if notification.Recipient == "" {
		return fmt.Errorf("recipient email cannot be empty")
	}

	// Validate email format
	if !isValidEmail(notification.Recipient) {
		return fmt.Errorf("invalid email format: %s", notification.Recipient)
	}

	message := gomail.NewMessage()
	message.SetHeader("From", ep.config.From)
	message.SetHeader("To", notification.Recipient)

	if notification.Subject != "" {
		message.SetHeader("Subject", notification.Subject)
	} else {
		message.SetHeader("Subject", "Notification")
	}

	message.SetBody("text/plain", notification.Message)

	// Retry logic with exponential backoff
	backoffMs := []int{1000, 2000, 4000} // 1s, 2s, 4s
	var lastErr error

	for attempt := 0; attempt < 3; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		// Send with timeout
		err := ep.dailer.DialAndSend(message)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableEmailError(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Sleep with exponential backoff if not the last attempt
		if attempt < 2 {
			backoff := time.Duration(backoffMs[attempt]) * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			}
		}
	}

	return fmt.Errorf("failed after 3 retries: %w", lastErr)
}

// GetStatus returns the current status of the provider
func (ep *EmailProvider) GetStatus() *ProviderStatus {
	// Try to connect to verify configuration
	conn, err := ep.dailer.Dial()
	if err != nil {
		return &ProviderStatus{
			Status:         StatusError,
			LastUpdated:    time.Now(),
			ErrorMessage:   fmt.Sprintf("SMTP connectivity check failed: %v", err),
			LastTestAt:     ep.lastTestAt,     // T049
			LastTestStatus: ep.lastTestStatus, // T049
		}
	}
	defer conn.Close()

	return &ProviderStatus{
		Status:         StatusActive,
		LastUpdated:    time.Now(),
		ErrorMessage:   fmt.Sprintf("SMTP: %s:%d", ep.config.Host, ep.config.Port),
		LastTestAt:     ep.lastTestAt,     // T049
		LastTestStatus: ep.lastTestStatus, // T049
	}
}

// GetID returns the provider ID
func (ep *EmailProvider) GetID() string {
	return ep.id
}

// GetType returns the provider type
func (ep *EmailProvider) GetType() string {
	return "email"
}

// GetTestRecipient returns the test_recipient or from address as fallback (T050)
func (ep *EmailProvider) GetTestRecipient() (string, error) {
	// Use test_recipient if configured
	if ep.config.TestRecipient != "" {
		return ep.config.TestRecipient, nil
	}
	
	// Fallback to from address
	if ep.config.From != "" {
		return ep.config.From, nil
	}
	
	return "", fmt.Errorf("test_recipient not configured and from address empty for Email provider %s", ep.id)
}

// Test sends a test email and updates last test metadata (T051, T052)
func (ep *EmailProvider) Test(ctx context.Context) error {
	// Get test recipient
	recipient, err := ep.GetTestRecipient()
	if err != nil {
		return fmt.Errorf("failed to get test recipient: %w", err)
	}

	// Create test notification with timestamp (T052)
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 MST")
	testNotification := &Notification{
		ID:         fmt.Sprintf("test-%s-%d", ep.id, time.Now().Unix()),
		ProviderID: ep.id,
		Recipient:  recipient,
		Subject:    "Test from notimulti", // T052: Email subject
		Message:    fmt.Sprintf("Test notification from notimulti server - %s", timestamp),
		Priority:   PriorityNormal,
		Timestamp:  time.Now(),
	}

	// Send test notification
	err = ep.Send(ctx, testNotification)
	
	// Update last test metadata (T051)
	now := time.Now()
	ep.lastTestAt = &now
	
	if err != nil {
		ep.lastTestStatus = "failed"
		return fmt.Errorf("test notification failed: %w", err)
	}
	
	ep.lastTestStatus = "success"
	return nil
}

// Close performs cleanup
func (ep *EmailProvider) Close() error {
	return nil
}

// Helper function to validate email format
func isValidEmail(email string) bool {
	// Basic email validation
	if email == "" {
		return false
	}

	// Check for @ symbol
	atIndex := -1
	for i, r := range email {
		if r == '@' {
			if atIndex != -1 {
				return false // Multiple @ symbols
			}
			atIndex = i
		}
	}

	if atIndex == -1 || atIndex == 0 || atIndex == len(email)-1 {
		return false
	}

	// Check for valid domain
	domain := email[atIndex+1:]
	return isValidDomain(domain)
}

// Helper function to validate domain format
func isValidDomain(domain string) bool {
	if domain == "" {
		return false
	}

	// Must contain at least one dot and a TLD
	hasDot := false
	for _, r := range domain {
		if r == '.' {
			hasDot = true
			break
		}
	}

	return hasDot
}

// Helper function to determine if SMTP error is retryable
func isRetryableEmailError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Network errors are typically retryable
	if _, ok := err.(net.Error); ok {
		return true
	}

	// Check for SMTP transient failures
	retryableErrors := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"temporary failure",
		"service unavailable",
		"try again later",
	}

	for _, retryable := range retryableErrors {
		if len(errStr) >= len(retryable) {
			match := true
			for i := 0; i < len(retryable); i++ {
				if errStr[i] != retryable[i] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}

	return false
}
