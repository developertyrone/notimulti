# Technical Research: Centralized Notification Server

**Date**: 2025-11-03  
**Feature**: 001-notification-server  
**Purpose**: Document technology choices, patterns, and best practices for implementation

## Technology Stack Decisions

### Backend: Go with Gin Framework

**Decision**: Use Go 1.21+ with Gin web framework for REST API implementation.

**Rationale**:
- **Performance**: Go's goroutine-based concurrency naturally handles 100+ concurrent requests (requirement NFR-004)
- **Simplicity**: Gin provides minimalist REST framework aligned with KISS principle
- **Standard Library**: Strong stdlib for HTTP, JSON, file I/O reduces external dependencies
- **Binary Distribution**: Compiles to single binary, simplifying deployment
- **Type Safety**: Static typing catches errors at compile time, improving code quality

**Alternatives Considered**:
- **Python + FastAPI**: More external dependencies, GIL limits true concurrency
- **Node.js + Express**: Callback complexity, less type safety without TypeScript
- **Rust + Actix**: Steeper learning curve, overkill for this use case

**Best Practices**:
- Use `context.Context` for request cancellation and timeouts
- Implement graceful shutdown for clean provider cleanup
- Use `sync.RWMutex` for thread-safe provider registry access
- Leverage `errgroup` for concurrent provider operations

### Frontend: Vue 3 + Vite + Tailwind CSS

**Decision**: Vue 3.3+ with Composition API, Vite 5 build tool, Tailwind CSS 3 for styling.

**Rationale**:
- **Simplicity**: Vue's progressive framework philosophy aligns with KISS principle
- **Composition API**: Better code organization for read-only dashboard
- **Vite**: Fast dev server and build times (<200ms HMR requirement)
- **Tailwind**: Utility-first CSS reduces custom CSS, improves consistency
- **Minimal Dependencies**: Core stack only, no heavy state management libraries needed

**Alternatives Considered**:
- **React**: More boilerplate, ecosystem complexity
- **Svelte**: Smaller ecosystem, less mature tooling
- **Plain HTML/JS**: Too much manual DOM manipulation for dynamic updates

**Best Practices**:
- Use `<script setup>` syntax for conciseness
- Implement auto-refresh with `setInterval` or WebSocket (start with polling)
- Use `fetch` API for backend communication (no axios needed)
- Leverage Tailwind's responsive utilities for mobile support

### Storage: SQLite for Metadata

**Decision**: SQLite 3 for storing provider status history and notification logs.

**Rationale**:
- **Zero Configuration**: No separate database server required (KISS principle)
- **Embedded**: Runs in-process, reduces deployment complexity
- **Performance**: Fast enough for 100s of providers and 1000s of notifications
- **Transactions**: ACID compliance ensures data consistency
- **File-Based**: Easy backup and migration

**Alternatives Considered**:
- **PostgreSQL/MySQL**: Overkill for this scale, requires separate server
- **In-Memory Only**: Loses history on restart, no audit trail
- **JSON Files**: No query capabilities, concurrency issues

**Best Practices**:
- Use Write-Ahead Logging (WAL) mode for better concurrency
- Create indexes on `provider_id` and `timestamp` columns
- Set `busy_timeout` for handling concurrent writes
- Implement connection pooling (1 connection sufficient for SQLite)

### Configuration: File-Based with File Watching

**Decision**: JSON configuration files in `configs/` directory, watched via `fsnotify` library.

**Rationale**:
- **Simplicity**: Human-readable, version-controllable, no database required
- **Standard Format**: JSON parsing built into Go stdlib
- **Dynamic Reload**: `fsnotify` provides cross-platform file system events
- **Atomicity**: File writes are atomic at OS level
- **Validation**: Easy to validate structure before applying

**Alternatives Considered**:
- **Environment Variables**: Not suitable for multiple provider instances
- **Database Configuration**: Adds unnecessary complexity and dependencies
- **YAML/TOML**: More dependencies, JSON is simpler and sufficient

**Best Practices**:
- Debounce file system events (avoid rapid reload on editor writes)
- Validate configuration before applying (keep old config on error)
- Use file naming convention: `<provider-type>-<instance-id>.json`
- Implement atomic configuration updates (load → validate → swap)

## Provider Architecture

### Interface-Based Design

**Pattern**: Define `Provider` interface, implement per provider type.

```go
type Provider interface {
    Send(ctx context.Context, notification Notification) error
    GetStatus() ProviderStatus
    GetID() string
    GetType() string
    Close() error
}
```

**Rationale**:
- **Extensibility**: New providers implement same interface
- **Polymorphism**: Registry can manage all providers uniformly
- **Testing**: Easy to mock providers for contract tests
- **Simplicity**: Clear contract, no complex plugin system

**Best Practices**:
- Use `context.Context` for cancellation and timeout control
- Return structured errors with provider-specific details
- Implement exponential backoff for transient failures
- Log all provider operations with request ID correlation

### Telegram Integration

**Decision**: Use `go-telegram-bot-api` library.

**Rationale**:
- Well-maintained, idiomatic Go API
- Supports both bot API methods and webhooks
- Handles rate limiting automatically
- Active community and good documentation

**Best Practices**:
- Set reasonable timeout (5s for Send operations)
- Handle rate limit errors (429) with exponential backoff
- Validate chat/channel ID format before sending
- Log Telegram error codes for troubleshooting

### Email Integration

**Decision**: Use Go stdlib `net/smtp` + `gomail` for message construction.

**Rationale**:
- Stdlib `net/smtp` for SMTP protocol (minimal dependencies)
- `gomail` for MIME message construction (attachments, HTML)
- Standard SMTP protocol ensures broad compatibility
- No heavy email framework needed

**Best Practices**:
- Support TLS/STARTTLS for secure connections
- Implement connection pooling for performance
- Set reasonable timeouts (10s for SMTP handshake, 30s for send)
- Validate email addresses before sending
- Support both plain text and HTML email bodies

## File Watching Strategy

**Implementation**: Use `fsnotify` with debouncing.

**Pattern**:
1. Watch `configs/` directory for CREATE, WRITE, REMOVE events
2. Debounce events (300ms) to avoid rapid reloads during editor saves
3. On event: read file → validate JSON → validate provider config → update registry
4. On error: log error, keep existing configuration

**Rationale**:
- `fsnotify` is cross-platform (Linux, macOS, Windows)
- Debouncing prevents reload storms
- Validation before apply prevents partial updates
- Graceful degradation on errors

**Best Practices**:
- Use separate goroutine for file watching
- Implement context cancellation for clean shutdown
- Log all configuration changes with before/after state
- Handle file system race conditions (read after delete)

## Logging Strategy

**Decision**: Structured JSON logging with configurable levels.

**Implementation**:
- Use `log/slog` (Go 1.21+ stdlib structured logging)
- Environment variable `LOG_LEVEL` (DEBUG, INFO, WARN, ERROR)
- Environment variable `LOG_FORMAT` (json, text)
- Include context fields: request_id, provider_id, operation

**Rationale**:
- Stdlib solution (no external dependencies)
- JSON format machine-parseable for log aggregation
- Configurable via environment (no code changes)
- Context propagation via `slog.Logger` with attributes

**Best Practices**:
- Redact sensitive fields (tokens, passwords) before logging
- Log at appropriate levels (DEBUG=verbose, INFO=operations, WARN=degraded, ERROR=failures)
- Include operation duration for performance tracking
- Use structured fields (not string interpolation)

## API Design

**Pattern**: RESTful API following OpenAPI 3.0 specification.

**Endpoints**:
- `POST /api/v1/notifications` - Send notification
- `GET /api/v1/providers` - List all providers
- `GET /api/v1/providers/:id` - Get provider details
- `GET /api/v1/health` - Health check

**Rationale**:
- REST is simple, widely understood
- Versioning (`/v1/`) allows future changes
- Standard HTTP status codes
- JSON request/response bodies

**Best Practices**:
- Use HTTP method semantics correctly (POST for actions)
- Return 201 Created for successful notification submission
- Return 404 for unknown provider IDs
- Return 400 with validation details for bad requests
- Include request ID in responses for tracing
- Implement CORS headers for frontend access

## Error Handling

**Strategy**: Structured errors with codes and context.

**Pattern**:
```go
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

**Rationale**:
- Machine-readable error codes
- Human-readable messages
- Additional context in details
- Consistent error format across API

**Best Practices**:
- Use specific error codes (INVALID_PROVIDER, VALIDATION_ERROR, etc.)
- Include field-level validation errors in details
- Log errors with full context before returning to client
- Never expose internal errors to client (wrap with generic message)

## Performance Considerations

### Concurrency

- Use goroutines for concurrent provider sends
- Limit concurrent sends with worker pool (default: 10 workers)
- Use buffered channels for notification queue
- Implement context-based timeout (default: 30s per notification)

### Caching

- Cache provider configurations in memory
- Invalidate cache on configuration file change
- No HTTP caching (UI fetches fresh data each time)

### Database

- Use prepared statements for repeated queries
- Create indexes on frequently queried columns
- Batch insert notification logs (flush every 100 records or 5s)
- Implement periodic cleanup of old logs (retention: 30 days)

## Security Considerations

- **API Authentication**: Start without auth (internal tool), add API key auth if needed
- **Configuration Files**: Set restrictive file permissions (600) for config files
- **Secrets**: Never log sensitive values (tokens, passwords)
- **Input Validation**: Validate all user inputs before processing
- **CORS**: Configure allowed origins for frontend access
- **Dependencies**: Regularly scan for vulnerabilities with `go mod` security tools

## Testing Strategy

### Contract Tests (Backend)
- Test all API endpoints with various inputs
- Verify correct status codes and response structure
- Use `httptest` package for testing handlers
- Mock provider implementations

### Integration Tests (Providers)
- Test actual Telegram/Email delivery (requires test accounts)
- Verify error handling for invalid credentials
- Test retry logic and timeout handling
- Use test containers if needed

### Unit Tests (Frontend)
- Test component rendering with various props
- Verify API client error handling
- Use Vitest with Vue Test Utils
- Aim for 80%+ coverage

## Development Workflow

1. **Setup**: Install Go 1.21+, Node 18+, SQLite
2. **Backend**: Run `go run cmd/server/main.go`
3. **Frontend**: Run `npm run dev` (Vite dev server)
4. **Tests**: Run `go test ./...` (backend), `npm test` (frontend)
5. **Linting**: `golangci-lint` (backend), `eslint` (frontend)
6. **Build**: `go build` (backend), `npm run build` (frontend)

## Deployment Considerations

- **Backend**: Single Go binary, runs on Linux/macOS/Windows
- **Frontend**: Static files, serve via backend or separate static host
- **Configuration**: Mount configs directory as volume
- **Database**: SQLite file, persist via volume mount
- **Ports**: Backend listens on configurable port (default: 8080)
- **Environment**: Set `LOG_LEVEL`, `LOG_FORMAT`, `CONFIG_DIR`, `DB_PATH` via env vars

## Future Extensibility

- **New Providers**: Implement `Provider` interface, add to registry
- **Webhooks**: Add webhook receiver for async delivery confirmation
- **Metrics**: Add Prometheus metrics endpoint
- **UI Write Operations**: Add configuration CRUD endpoints and UI forms
- **Multi-Instance**: Add horizontal scaling with shared SQLite (or migrate to PostgreSQL)
- **Message Queue**: Add Redis/RabbitMQ for reliable delivery queuing
