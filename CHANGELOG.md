# Changelog

All notable changes to the Centralized Notification Server will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-11-06

### Added

#### User Story 1: Send Notifications via REST API (P1)
- REST API endpoint `POST /api/v1/notifications` for sending notifications
- Telegram provider implementation with exponential backoff retry logic
- Email provider (SMTP) implementation with TLS/STARTTLS support
- Provider interface and registry for managing multiple provider instances
- SQLite database for notification history and audit trail
- Structured JSON logging with configurable levels (DEBUG, INFO, WARN, ERROR)
- Health check endpoint `GET /api/v1/health`
- Request validation with detailed error messages for field-level errors
- Support for notification priorities (low, normal, high)
- Support for custom metadata (up to 10 key-value pairs per notification)

#### User Story 2: Dynamic Provider Configuration (P2)
- File-based configuration using JSON files in `configs/` directory
- File system watching with automatic provider reload on configuration changes
- Configuration validation before applying changes (malformed configs rejected)
- Atomic provider replacement during configuration updates
- Debounced file watching to prevent reload storms during editor saves
- Graceful error handling for invalid configurations (keeps existing provider active)
- Configuration change detection within 30 seconds (typical: <5 seconds)
- Support for enabling/disabling providers without restart

#### User Story 3: View Current Server Configuration (P3)
- Read-only web UI for monitoring provider status
- Dashboard view showing all configured providers
- Provider list endpoint `GET /api/v1/providers`
- Provider detail endpoint `GET /api/v1/providers/:id`
- Sensitive field masking in API responses (tokens, passwords)
- Auto-refresh every 30 seconds for real-time status updates
- Status indicators: active, error, disabled, initializing
- Responsive UI with Tailwind CSS (mobile-friendly)

#### Technical Features
- Go 1.21+ backend with Gin web framework
- Vue 3 + Vite + Tailwind CSS frontend
- SQLite 3 for metadata persistence with WAL mode
- Concurrent request handling (100+ concurrent requests supported)
- Graceful shutdown with proper resource cleanup
- CORS middleware for frontend-backend communication
- Request ID propagation for distributed tracing
- Environment-based configuration via .env files

### Supported Providers

- **Telegram**: Send messages via Telegram Bot API
  - Markdown and HTML parse modes
  - Rate limiting with exponential backoff
  - Timeout configuration (default: 5s)
  - Retry logic for transient failures (max 3 retries)

- **Email (SMTP)**: Send emails via SMTP
  - TLS/STARTTLS support
  - Connection pooling for performance
  - Plain text and HTML email bodies
  - Subject and metadata support
  - Timeout configuration (default: 30s)
  - Retry logic for transient failures (max 3 retries)

### Performance

- API response time: <2s (p95) for 100 concurrent requests
- Configuration reload: <5s (typically <2s)
- UI interaction: <200ms (p95)
- Provider change detection: <30s (typically <5s)

### Security

- Sensitive data redaction in logs (tokens, passwords)
- File permission recommendations for configuration files
- Masked sensitive fields in API responses
- Secure SMTP connection support (TLS/STARTTLS)

### Testing

- Comprehensive test suite with 80%+ code coverage
- Contract tests for all API endpoints
- Integration tests for file watching and configuration reload
- Unit tests for all core components
- End-to-end tests for complete workflows

### Documentation

- README.md with quick start guide and architecture overview
- Quickstart guide with step-by-step setup instructions
- OpenAPI 3.0 specification for REST API
- Environment variable documentation
- Production deployment guide

## Known Limitations

### v1.0.0
- **No authentication**: API endpoints are not authenticated (suitable for internal use only)
- **No WebSocket support**: UI uses polling (30s interval) instead of real-time updates
- **No message queue**: Notifications are sent synchronously (no guaranteed delivery on server crash)
- **Single instance**: No horizontal scaling support (SQLite limitation)
- **No attachment support**: Email provider doesn't support attachments yet
- **No retry queue**: Failed notifications are logged but not automatically retried
- **Limited provider types**: Only Telegram and Email supported (no SMS, Slack, Discord, etc.)
- **No configuration UI**: Provider configurations must be edited as JSON files
- **No notification templates**: Messages must be constructed by the client application
- **No rate limiting**: No per-provider or per-client rate limiting

### Workarounds
- **Authentication**: Use reverse proxy (nginx/caddy) with HTTP Basic Auth or API key validation
- **Real-time updates**: Implement WebSocket endpoint in future version or use Server-Sent Events
- **Guaranteed delivery**: Use external message queue (Redis, RabbitMQ) in future version
- **Horizontal scaling**: Migrate to PostgreSQL for multi-instance support
- **Attachments**: Use base64 encoding in metadata as temporary workaround
- **Retry queue**: Implement background job processor in future version

## Future Roadmap

### v1.1.0 (Planned)
- Additional providers: SMS (Twilio), Slack, Discord, Microsoft Teams
- Notification templates with variable substitution
- API key authentication
- WebSocket support for real-time UI updates
- Email attachment support

### v1.2.0 (Planned)
- Retry queue for failed notifications
- Rate limiting per provider and per client
- Notification scheduling (send at specific time)
- Bulk notification sending
- Configuration UI (web-based provider management)

### v2.0.0 (Planned)
- PostgreSQL support for horizontal scaling
- Message queue integration (Redis, RabbitMQ)
- Prometheus metrics export
- Webhook callbacks for delivery confirmation
- Multi-tenancy support

## Migration Notes

### Upgrading to v1.0.0
This is the initial release. No migration required.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## License

See [LICENSE](LICENSE) for license information.

---

For detailed feature specifications, see [specs/001-notification-server/spec.md](specs/001-notification-server/spec.md).
