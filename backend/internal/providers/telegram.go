package providers

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramProvider implements the Provider interface for Telegram Bot API
type TelegramProvider struct {
	id     string
	bot    *tgbotapi.BotAPI
	config *TelegramConfig
}

// NewTelegramProvider creates a new Telegram provider instance
func NewTelegramProvider(id string, config *TelegramConfig) (*TelegramProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.BotToken == "" {
		return nil, fmt.Errorf("bot_token is required")
	}

	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	return &TelegramProvider{
		id:     id,
		bot:    bot,
		config: config,
	}, nil
}

// Send sends a notification via Telegram with retry logic
func (tp *TelegramProvider) Send(ctx context.Context, notification *Notification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	if notification.Recipient == "" {
		return fmt.Errorf("recipient (chat_id) cannot be empty")
	}

	chatID, err := parseChatID(notification.Recipient)
	if err != nil {
		return fmt.Errorf("invalid chat_id: %w", err)
	}

	// Prepare message with optional parse mode
	parseMode := tp.config.ParseMode
	if parseMode == "" {
		parseMode = "HTML"
	}

	message := tgbotapi.NewMessage(chatID, notification.Message)
	message.ParseMode = parseMode

	// Add optional subject as reply_markup or in message
	if notification.Subject != "" {
		message.Text = fmt.Sprintf("<b>%s</b>\n\n%s", notification.Subject, notification.Message)
	}

	// Apply timeout from config (default 5s)
	timeout := 5 * time.Second
	if tp.config.TimeoutSeconds > 0 {
		timeout = time.Duration(tp.config.TimeoutSeconds) * time.Second
	}

	// Create context with timeout for retries
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Retry logic with exponential backoff
	backoffMs := []int{1000, 2000, 4000} // 1s, 2s, 4s
	var lastErr error

	for attempt := 0; attempt < 3; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		_, err := tp.bot.Send(message)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
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
func (tp *TelegramProvider) GetStatus() *ProviderStatus {
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to get bot info to verify connectivity
	user, err := tp.bot.GetMe()
	if err != nil {
		return &ProviderStatus{
			Status:       StatusError,
			LastUpdated:  time.Now(),
			ErrorMessage: fmt.Sprintf("bot connectivity check failed: %v", err),
		}
	}

	return &ProviderStatus{
		Status:       StatusActive,
		LastUpdated:  time.Now(),
		ErrorMessage: fmt.Sprintf("Bot: @%s (%s)", user.UserName, user.FirstName),
	}
}

// GetID returns the provider ID
func (tp *TelegramProvider) GetID() string {
	return tp.id
}

// GetType returns the provider type
func (tp *TelegramProvider) GetType() string {
	return "telegram"
}

// Close performs cleanup (Telegram doesn't require explicit cleanup)
func (tp *TelegramProvider) Close() error {
	return nil
}

// Helper function to parse chat ID from string
func parseChatID(chatIDStr string) (int64, error) {
	var chatID int64
	_, err := fmt.Sscanf(chatIDStr, "%d", &chatID)
	if err != nil {
		return 0, fmt.Errorf("invalid chat_id format: %s", chatIDStr)
	}
	return chatID, nil
}

// Helper function to determine if error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Rate limit (429) or other API errors
	if apiErr, ok := err.(tgbotapi.Error); ok {
		// RetryAfter indicates rate limiting - this is retryable
		if apiErr.RetryAfter > 0 {
			return true
		}
	}

	// Timeout or connection errors
	retryableErrors := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"EOF",
		"temporary failure",
	}

	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}

	return false
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(substr) <= len(s)))
}
