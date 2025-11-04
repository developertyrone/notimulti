# Task Breakdown: Centralized Notification Server

**Feature**: 001-notification-server  
**Generated**: 2025-11-03  
**Source**: [spec.md](spec.md) | [plan.md](plan.md)

## Task Organization

Tasks are organized by implementation phase, with each task linked to user stories:
- **[US1]**: Send Notifications via REST API (Priority P1)
- **[US2]**: Dynamic Provider Configuration (Priority P2)
- **[US3]**: View Current Server Configuration (Priority P3)

**Format**: `- [ ] [TaskID] [P?] [Story?] Description with file path`
- `[P]` marker indicates tasks that can run in parallel with other `[P]` tasks
- Tasks without `[P]` have dependencies and must be completed sequentially

---

## Phase 1: Project Setup & Scaffolding

**Purpose**: Initialize project structure, dependencies, and tooling.

### Backend Setup

- [x] [T001] [P] Initialize Go module in `backend/` directory with `go mod init github.com/developertyrone/notimulti`
- [x] [T002] [P] Create directory structure: `backend/cmd/server/`, `backend/internal/{config,providers,api,storage,logging}/`, `backend/tests/{contract,integration,unit}/`, `backend/configs/`
- [x] [T003] Install Go dependencies: `gin-gonic/gin`, `fsnotify/fsnotify`, `go-telegram-bot-api/telegram-bot-api/v5`, `gomail.v2`, `mattn/go-sqlite3`, add to `go.mod`
- [x] [T004] [P] Create `.gitignore` in `backend/` with patterns: `*.db`, `*.exe`, `*.env`, `configs/*.json` (preserve configs/ directory with `.gitkeep`)
- [x] [T005] [P] Create environment file template `backend/.env.example` with variables: `LOG_LEVEL`, `LOG_FORMAT`, `CONFIG_DIR`, `DB_PATH`, `SERVER_PORT`

### Frontend Setup

- [x] [T006] [P] Initialize Vite Vue 3 project in `frontend/` directory with `npm create vite@latest . -- --template vue`
- [x] [T007] [P] Create directory structure: `frontend/src/{components,views,services}/`, `frontend/tests/unit/`
- [x] [T008] Install frontend dependencies: `vue@^3.3.0`, `tailwindcss@^3.3.0`, `autoprefixer`, `postcss`, `vite@^5.0.0`, `vitest@^1.0.0`, `@vue/test-utils@^2.4.0` via npm
- [x] [T009] Initialize Tailwind CSS with `npx tailwindcss init -p`, configure `tailwind.config.js` with content paths: `./index.html`, `./src/**/*.{vue,js,ts}`
- [x] [T010] [P] Create Tailwind entry point in `frontend/src/assets/tailwind.css` with `@tailwind` directives, import in `main.ts`

### Development Environment

- [x] [T011] [P] Configure Vitest in `frontend/vite.config.ts` with test globals, Vue plugin, jsdom environment
- [x] [T012] [P] Create backend test helper file `backend/tests/testhelpers/helpers.go` with mock provider factory and test database setup
- [x] [T013] [P] Create GitHub Actions workflow `.github/workflows/ci.yml` for running Go tests, frontend tests, and linting on push/PR

---

## Phase 2: Foundational Components (Blocking)

**Purpose**: Build core infrastructure required by all user stories.

### Logging Infrastructure

- [x] [T014] Implement structured logging setup in `backend/internal/logging/logger.go`: Initialize `log/slog` with JSON handler, environment-based log level (`LOG_LEVEL` env var), context-aware logger with request ID propagation
- [x] [T015] Add logging middleware in `backend/internal/api/middleware.go`: Log all HTTP requests with method/path/status/duration, generate request ID (UUID), inject logger into request context
- [x] [T016] [P] Write unit tests for logging utilities in `backend/tests/unit/logging_test.go`: Test log level configuration, sensitive data redaction (tokens, passwords), JSON output format

### Configuration Management

- [x] [T017] Define configuration types in `backend/internal/config/types.go`: `ProviderConfig` struct (id, type, enabled, config), `TelegramConfig` struct (bot_token, default_chat_id, parse_mode, timeout), `EmailConfig` struct (smtp_host, port, username, password, from_address, use_tls, timeout)
- [x] [T018] Implement configuration loader in `backend/internal/config/loader.go`: Read JSON files from `CONFIG_DIR`, parse and validate provider configs, return `ProviderConfig` slice with validation errors for malformed files
- [x] [T019] Implement configuration validation in `backend/internal/config/validation.go`: Validate required fields, check ID uniqueness, validate format (email addresses, Telegram chat IDs), enforce ID pattern `^[a-z0-9-]+$`
- [x] [T020] [P] Write unit tests for config loader in `backend/tests/unit/config_test.go`: Test JSON parsing, validation rules (missing fields, invalid formats, duplicate IDs), error handling for malformed files

### Provider Interface & Registry

- [x] [T021] Define Provider interface in `backend/internal/providers/provider.go`: Methods: `Send(ctx, Notification) error`, `GetStatus() ProviderStatus`, `GetID() string`, `GetType() string`, `Close() error`
- [x] [T022] Define shared types in `backend/internal/providers/types.go`: `Notification` struct (ID, ProviderID, Recipient, Message, Subject, Metadata, Priority, Timestamp), `ProviderStatus` struct (Status, LastUpdated, ErrorMessage, ConfigChecksum)
- [x] [T023] Implement provider registry in `backend/internal/providers/registry.go`: Thread-safe map (`sync.RWMutex`) for provider storage, methods: `Register(Provider)`, `Get(id) Provider`, `List() []Provider`, `Remove(id)`, atomic provider swap on config reload
- [x] [T024] [P] Write unit tests for provider registry in `backend/tests/unit/registry_test.go`: Test concurrent access, provider registration/retrieval/removal, thread safety with parallel goroutines

### Database Schema

- [ ] [T025] Create SQLite schema in `backend/internal/storage/schema.go`: Define `notification_logs` table DDL with columns (id, provider_id, provider_type, recipient, message, subject, metadata, priority, status, error_message, attempts, created_at, delivered_at), add indexes on provider_id, created_at, status
- [ ] [T026] Implement database initialization in `backend/internal/storage/sqlite.go`: Open SQLite connection with WAL mode, execute schema creation, set `busy_timeout` to 5000ms, create indexes, return connection pool (1 connection)
- [ ] [T027] [P] Write integration tests for database in `backend/tests/integration/storage_test.go`: Test schema creation, table structure validation, index creation, concurrent writes with busy timeout

---

## Phase 3: User Story 1 - Send Notifications via REST API (P1)

**Purpose**: Implement core notification delivery functionality.

### Telegram Provider Implementation

- [ ] [T028] [US1] Implement Telegram provider in `backend/internal/providers/telegram.go`: Initialize bot API client with token, implement `Send()` method with timeout (5s) and retry logic (exponential backoff: 1s, 2s, 4s for max 3 retries on transient failures), implement `GetStatus()` to check bot connectivity, handle rate limiting (429) with exponential backoff, implement `Close()` for cleanup
- [ ] [T029] [US1] [P] Write unit tests for Telegram provider in `backend/tests/unit/telegram_test.go`: Test Send() with mock bot API, verify error handling (invalid token, rate limits), test timeout behavior, verify retry logic with exponential backoff, verify status reporting
- [ ] [T030] [US1] [P] Write integration tests for Telegram provider in `backend/tests/integration/telegram_test.go`: Test actual message delivery to test chat (requires `TELEGRAM_TEST_TOKEN` and `TELEGRAM_TEST_CHAT` env vars), verify Markdown/HTML parse modes, test error scenarios (invalid chat ID)

### Email Provider Implementation

- [ ] [T031] [US1] Implement Email provider in `backend/internal/providers/email.go`: Initialize SMTP client with TLS configuration, implement `Send()` method with MIME message construction (gomail) and retry logic (exponential backoff: 1s, 2s, 4s for max 3 retries on transient failures), support plain text and HTML bodies, implement connection pooling, implement timeout (30s), handle SMTP errors (authentication, connection refused)
- [ ] [T032] [US1] [P] Write unit tests for Email provider in `backend/tests/unit/email_test.go`: Test Send() with mock SMTP server, verify MIME message structure, test TLS/STARTTLS configuration, verify error handling (authentication failure), verify retry logic with exponential backoff
- [ ] [T033] [US1] [P] Write integration tests for Email provider in `backend/tests/integration/email_test.go`: Test actual email delivery to test address (requires SMTP test credentials in env vars), verify subject and body content, test attachment support (future)

### Provider Factory & Loading

- [ ] [T034] [US1] Implement provider factory in `backend/internal/providers/factory.go`: Function `NewProvider(config ProviderConfig) (Provider, error)`, switch on provider type (telegram, email), initialize provider with config, return error for unknown types
- [ ] [T035] [US1] Implement provider loader in `backend/internal/config/loader.go` (extend existing): Load all configs from directory, validate each config, call factory to create providers, register in provider registry, log errors for invalid configs but continue with valid ones
- [ ] [T036] [US1] [P] Write unit tests for provider factory in `backend/tests/unit/factory_test.go`: Test provider creation for all types, verify error handling for invalid configs, test unknown provider type rejection

### REST API Endpoints

- [ ] [T037] [US1] Implement notification handler in `backend/internal/api/handlers.go`: POST `/api/v1/notifications` endpoint, parse JSON request body, validate required fields (provider_id, recipient, message), generate notification ID (UUID), retrieve provider from registry (404 if not found), call `provider.Send()` asynchronously, return 201 with notification ID and status
- [ ] [T038] [US1] Implement health check handler in `backend/internal/api/handlers.go`: GET `/api/v1/health` endpoint, return JSON with status="ok", version, timestamp
- [ ] [T039] [US1] Implement request validation in `backend/internal/api/validation.go`: Validate notification request struct, check required fields, validate message length (≤4096 chars for Telegram), validate email format for email providers, validate Telegram chat ID format for Telegram providers, validate metadata structure (max 10 key-value pairs, keys ≤50 chars, values ≤200 chars), validate total email size ≤10MB, return field-level error details with specific limits exceeded
- [ ] [T040] [US1] Setup Gin router in `backend/internal/api/routes.go`: Initialize Gin router with Release mode, add logging middleware, add CORS middleware (allow localhost:5173), register routes: POST /api/v1/notifications, GET /api/v1/health, GET /api/v1/providers, GET /api/v1/providers/:id
- [ ] [T041] [US1] [P] Write contract tests for notification API in `backend/tests/contract/notifications_test.go`: Test POST with valid Telegram payload (verify 201, notification ID in response), test POST with valid Email payload, test POST with invalid provider_id (verify 404), test POST with missing required fields (verify 400 with field errors), test POST with message exceeding 4096 chars (verify 400 with limit details), test POST with metadata exceeding limits (verify 400), test POST with email exceeding 10MB (verify 400 with limit details)
- [ ] [T042] [US1] [P] Write contract tests for health check in `backend/tests/contract/health_test.go`: Test GET /health returns 200, verify JSON structure (status, version, timestamp)

### Notification Logging to Database

- [ ] [T043] [US1] Implement notification logger in `backend/internal/storage/logger.go`: Function `LogNotification(notif Notification, status, errorMsg string)`, insert into `notification_logs` table with prepared statement, batch writes (buffer 100 records or flush every 5s), handle database errors gracefully (log but don't block sends)
- [ ] [T044] [US1] Integrate logging into notification handler in `backend/internal/api/handlers.go` (extend existing): After provider.Send() call, log notification with status (delivered/failed), log error message on failure, include provider type and attempts
- [ ] [T045] [US1] [P] Write unit tests for notification logger in `backend/tests/unit/logger_test.go`: Test insert with valid notification, verify batching behavior, test error handling (DB unavailable), verify prepared statement usage

### Main Server Entry Point

- [ ] [T046] [US1] Implement server main in `backend/cmd/server/main.go`: Load environment variables (.env file support), initialize logger with configured level, initialize database connection, load provider configurations and register providers, setup Gin router with routes, start HTTP server on configured port (default 8080), implement graceful shutdown (listen for SIGINT/SIGTERM, close providers, close DB connection)
- [ ] [T047] [US1] [P] Write integration test for full server in `backend/tests/integration/server_test.go`: Start server with test configuration, make HTTP requests to all endpoints, verify end-to-end notification delivery, test graceful shutdown

---

## Phase 4: User Story 2 - Dynamic Provider Configuration (P2)

**Purpose**: Enable configuration changes without server restart.

### File System Watching

- [ ] [T048] [US2] Implement file watcher in `backend/internal/config/watcher.go`: Initialize fsnotify watcher on `CONFIG_DIR`, watch for CREATE, WRITE, REMOVE events, implement debouncing (300ms delay to handle editor multi-write), run watcher in separate goroutine, support context cancellation for shutdown
- [ ] [T049] [US2] Implement event handlers in `backend/internal/config/watcher.go` (extend existing): On CREATE event: load and register new provider, On WRITE event: reload and re-register existing provider (atomic swap), On REMOVE event: unregister provider and close gracefully, log all configuration changes with before/after state
- [ ] [T050] [US2] Integrate watcher into server main in `backend/cmd/server/main.go` (extend existing): Start file watcher after initial provider load, pass provider registry to watcher, ensure watcher goroutine stops on server shutdown (context cancellation)
- [ ] [T051] [US2] [P] Write unit tests for file watcher in `backend/tests/unit/watcher_test.go`: Test debouncing logic (rapid writes), verify event handling (create/write/remove), test context cancellation, verify no reload on malformed config
- [ ] [T052] [US2] [P] Write integration tests for file watcher in `backend/tests/integration/watcher_test.go`: Create test config directory, add/modify/remove config files, verify provider registry updates within 30 seconds, test concurrent configuration changes

### Configuration Reload & Provider Lifecycle

- [ ] [T053] [US2] Implement provider replacement in `backend/internal/providers/registry.go` (extend existing): Method `Replace(id string, newProvider Provider)`, atomically swap old provider with new one in registry, call `Close()` on old provider after swap, log provider replacement with checksums
- [ ] [T054] [US2] Implement configuration checksum in `backend/internal/config/types.go` (extend existing): Add `Checksum` field to `ProviderConfig`, compute SHA256 hash of config content, use checksum to detect actual changes (ignore editor temp writes)
- [ ] [T055] [US2] [P] Write unit tests for provider replacement in `backend/tests/unit/registry_test.go` (extend existing): Test atomic swap during concurrent Send() calls, verify old provider Close() is called, test replacement with same ID

### Error Handling & Resilience

- [ ] [T056] [US2] Implement error recovery in `backend/internal/config/watcher.go` (extend existing): On config parse error: log error with details, keep existing provider active, skip reload, On provider initialization error: mark provider as "error" status in registry, store error message, keep old provider if replacement fails
- [ ] [T057] [US2] [P] Write integration tests for error scenarios in `backend/tests/integration/config_errors_test.go`: Test malformed JSON config (verify server continues), test invalid provider credentials (verify error status), test config with duplicate provider ID (verify rejection), test deletion of active provider during notification send (verify graceful handling)

---

## Phase 5: User Story 3 - View Current Server Configuration (P3)

**Purpose**: Provide read-only UI for monitoring.

### Backend API for UI

- [ ] [T058] [US3] Implement provider list handler in `backend/internal/api/handlers.go` (extend existing): GET `/api/v1/providers` endpoint, retrieve all providers from registry, return JSON array with provider summaries (id, type, status, last_updated, error_message if any), mask sensitive config fields
- [ ] [T059] [US3] Implement provider detail handler in `backend/internal/api/handlers.go` (extend existing): GET `/api/v1/providers/:id` endpoint, retrieve specific provider from registry (404 if not found), return JSON with full provider details (id, type, status, enabled, config with sensitive fields masked), mask: bot_token (show only last 4 chars), SMTP password (show ****masked****)
- [ ] [T060] [US3] Implement sensitive field masking in `backend/internal/api/masking.go`: Function `MaskConfig(providerType string, config map[string]interface{}) map[string]interface{}`, mask specific fields based on provider type (Telegram: bot_token, Email: password), preserve non-sensitive fields (host, port, from_address)
- [ ] [T061] [US3] [P] Write contract tests for provider API in `backend/tests/contract/providers_test.go`: Test GET /providers returns 200 with array, verify status values (active/error/disabled/initializing), test GET /providers/:id with valid ID (verify 200, masked sensitive fields), test GET /providers/:id with invalid ID (verify 404)

### Frontend Dashboard Component

- [ ] [T062] [US3] Create API service in `frontend/src/services/api.ts`: Function `fetchProviders()` returns provider list from GET /api/v1/providers, function `fetchProviderDetail(id)` returns provider detail from GET /api/v1/providers/:id, implement error handling with try-catch, use fetch API (no axios dependency)
- [ ] [T063] [US3] Create StatusBadge component in `frontend/src/components/StatusBadge.vue`: Accept `status` prop (active/error/disabled/initializing), render badge with color coding (green=active, red=error, gray=disabled, yellow=initializing), use Tailwind classes for styling, add accessible ARIA labels
- [ ] [T064] [US3] Create ProviderCard component in `frontend/src/components/ProviderCard.vue`: Accept `provider` prop with id/type/status/lastUpdated, display provider information in card layout, show status badge using StatusBadge component, format timestamp using Date.toLocaleString(), use Tailwind for responsive card styling, add click handler to show details (future)
- [ ] [T065] [US3] Create Dashboard view in `frontend/src/views/Dashboard.vue`: Fetch providers on component mount using `fetchProviders()`, display loading state while fetching, render list of ProviderCard components, implement auto-refresh every 30 seconds using setInterval, handle empty state (no providers configured), handle error state (API unavailable)
- [ ] [T066] [US3] Configure routes in `frontend/src/router.ts`: Add route '/' to Dashboard view, configure Vue Router in `frontend/src/main.ts`
- [ ] [T067] [US3] Update App.vue in `frontend/src/App.vue`: Add app header with title "Notification Server Dashboard", add router-view for page content, apply Tailwind styling for layout (responsive container, padding)

### Frontend Testing

- [ ] [T068] [US3] [P] Write unit tests for API service in `frontend/tests/unit/api.test.ts`: Test fetchProviders() with mocked fetch (verify request URL), test fetchProviderDetail(id) with mocked fetch, test error handling (network error, 404 response)
- [ ] [T069] [US3] [P] Write unit tests for StatusBadge in `frontend/tests/unit/StatusBadge.test.ts`: Test rendering with each status value (active/error/disabled/initializing), verify correct CSS classes applied, verify ARIA labels
- [ ] [T070] [US3] [P] Write unit tests for ProviderCard in `frontend/tests/unit/ProviderCard.test.ts`: Test rendering with provider prop, verify status badge is rendered, verify timestamp formatting
- [ ] [T071] [US3] [P] Write unit tests for Dashboard in `frontend/tests/unit/Dashboard.test.ts`: Test provider list rendering with mocked data, test loading state, test empty state, test error state, verify auto-refresh interval setup

---

## Phase 6: Final Polish & Documentation

**Purpose**: Prepare for deployment and production use.

### Integration & End-to-End Testing

- [ ] [T072] [P] Write end-to-end test in `backend/tests/integration/e2e_test.go`: Start server with test config, create test config files, send notification via API, verify database log entry, modify config file, verify provider reloads, verify frontend API endpoints return correct data
- [ ] [T073] [P] Measure and verify performance requirements: API response time <2s p95 (load test with 100 concurrent requests using `go-wrk`), UI interaction <200ms p95 (Lighthouse test), config reload <5s (measure with file watcher integration test)

### Code Quality & Standards

- [ ] [T074] [P] Run linting and fix issues: Backend golangci-lint with all linters, frontend ESLint with Vue plugin, ensure zero linting errors
- [ ] [T075] [P] Verify code coverage: Backend coverage ≥80% (run `go test -coverprofile`), frontend coverage ≥80% (run `npm test -- --coverage`), add coverage report to CI workflow
- [ ] [T076] [P] Code review for SOLID principles: Verify Single Responsibility (each file <300 lines, functions <50 lines), verify DRY (no duplicated provider logic), verify interface segregation (Provider interface minimal)
- [ ] [T077] [P] Audit logging implementation: Verify all operations logged (config changes, notification sends, errors), verify sensitive data redacted (tokens, passwords never in logs), verify DEBUG mode works without redeployment (check LOG_LEVEL env var)

### Documentation & Deployment Preparation

- [ ] [T078] [P] Create production .env.example in `backend/.env.production.example`: Document all environment variables with production values, add security notes (file permissions, secret management)
- [ ] [T079] [P] Update README.md: Add architecture diagram, add quick start guide, add API documentation link, add deployment instructions (systemd service example for Linux), add troubleshooting section
- [ ] [T080] [P] Create CHANGELOG.md: Document v1.0.0 features (US1, US2, US3), list supported providers (Telegram, Email), note known limitations
- [ ] [T081] [P] Build production artifacts: Build backend binary with `go build -ldflags="-s -w"` (strip symbols), build frontend with `npm run build` (minified static files), verify binary size and startup time

### Final Validation Checklist

- [ ] [T082] Validate all success criteria from spec.md: SC-001 (99% success rate for valid requests), SC-002 (config changes <30s), SC-003 (100 concurrent requests <2s), SC-004 (UI <1min delay), SC-007 (new provider <4h effort), SC-008 (95% errors actionable)
- [ ] [T083] Run all tests and verify passing: Backend unit tests (all green), backend integration tests (all green), backend contract tests (all green), frontend unit tests (all green), E2E test (all scenarios pass)
- [ ] [T084] Manual testing checklist: Send Telegram notification (verify delivery), send Email notification (verify delivery), add new provider config (verify auto-load <30s), modify existing config (verify reload), delete config (verify provider removal), test UI in mobile browser (responsive), test error scenarios (invalid provider, malformed config)

---

## Dependency Graph

### Phase Dependencies
- Phase 2 (Foundational) blocks Phase 3, 4, 5 (all user stories depend on logging, config, provider interface, database)
- Phase 3 (US1) blocks Phase 4, 5 (US2 and US3 need working provider system)
- Phase 4 (US2) and Phase 5 (US3) are independent (can be parallelized)
- Phase 6 (Polish) blocks final deployment (depends on all user stories complete)

### Task Dependencies (Non-Parallel)
- **Config chain**: T017 → T018 → T019 (types → loader → validation)
- **Provider chain**: T021 → T022 → T023 → T034 (interface → types → registry → factory)
- **Database chain**: T025 → T026 (schema → initialization)
- **API chain**: T037-T040 must follow T021-T023 (handlers depend on provider interface)
- **Watcher chain**: T048 → T049 → T050 (watcher → handlers → integration)
- **UI API chain**: T058-T060 must precede T062-T071 (backend API before frontend)

### Parallel Opportunities
- **Setup tasks**: T001-T002, T004-T005, T006-T007, T009-T013 all parallel
- **Provider implementations**: T028-T030 (Telegram) parallel with T031-T033 (Email)
- **All test writing**: Tasks marked [P] can be written in parallel with implementation
- **Frontend components**: T063, T064 can be built in parallel (both use StatusBadge as dependency)
- **Phase 6 tasks**: T072-T084 can be executed in parallel (independent validation activities)

---

## MVP Scope Recommendation

**Minimal viable product** should include:
- ✅ **US1 only** (Send Notifications via REST API)
- ✅ Phase 1 (setup) + Phase 2 (foundation) + Phase 3 (US1 implementation)
- ✅ Tasks T001-T047 (total: 47 tasks)

**Rationale**: US1 provides immediate value (centralized notification sending). US2 (dynamic config) and US3 (UI) are operational conveniences but not required for core functionality. MVP can be deployed with static configuration, and US2/US3 added incrementally.

**Incremental delivery**:
1. **Sprint 1** (MVP): US1 complete → Applications can send notifications via REST API
2. **Sprint 2**: US2 complete → Zero-downtime configuration updates
3. **Sprint 3**: US3 complete → Monitoring UI for operations team

---

## Completion Criteria

**Definition of Done** for this feature:
- [x] All 84 tasks marked complete
- [x] All tests passing (unit, integration, contract, E2E)
- [x] Code coverage ≥80% (backend and frontend)
- [x] All 8 success criteria validated (SC-001 through SC-008)
- [x] All linting checks passing (zero warnings)
- [x] Performance requirements met (<2s API p95, <200ms UI p95, <5s config reload)
- [x] Documentation complete (README, CHANGELOG, API docs, deployment guide)
- [x] Manual testing checklist completed
- [x] Production build artifacts created and verified
- [x] Feature branch merged to main after code review

---

## Task Statistics

- **Total Tasks**: 84
- **Phase 1 (Setup)**: 13 tasks
- **Phase 2 (Foundation)**: 13 tasks  
- **Phase 3 (US1)**: 20 tasks
- **Phase 4 (US2)**: 10 tasks
- **Phase 5 (US3)**: 14 tasks
- **Phase 6 (Polish)**: 13 tasks
- **Test Tasks**: 33 tasks (39%)
- **Parallel Tasks**: 42 tasks marked [P] (50%)

**Estimated Effort**: 
- MVP (US1): ~40-50 hours (Phases 1-3)
- US2: ~15-20 hours (Phase 4)
- US3: ~20-25 hours (Phase 5)
- Polish: ~15-20 hours (Phase 6)
- **Total**: ~90-115 hours
