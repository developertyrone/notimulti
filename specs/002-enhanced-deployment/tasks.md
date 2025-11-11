# Tasks: Enhanced Deployment & Operations

**Feature Branch**: `002-enhanced-deployment`  
**Input**: Design documents from `/specs/002-enhanced-deployment/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/openapi.yaml, quickstart.md

**Tests**: Tests are MANDATORY per Constitution Principle II (Test-Driven Development). All user stories MUST include contract and integration tests written BEFORE implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4, SETUP, FOUNDATION, POLISH)
- Include exact file paths in descriptions

## Path Conventions (from plan.md)

- Backend: `backend/internal/`, `backend/cmd/`, `backend/tests/`
- Frontend: `frontend/src/`, `frontend/tests/`
- Deployment: `Dockerfile`, `deploy/`
- CI/CD: `.github/workflows/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure required before any implementation

- [X] T001 [SETUP] Update backend/internal/storage/schema.go to add `is_test` column and composite indexes per data-model.md migration strategy
- [X] T002 [SETUP] Enable WAL mode in backend/internal/storage/sqlite.go: add `PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;`
- [X] T003 [P] [SETUP] Create backend/internal/storage/logger.go structure for notification logger service (empty struct, constructor)
- [X] T004 [P] [SETUP] Create backend/internal/storage/repository.go structure for database query methods (empty struct, constructor)
- [X] T005 [P] [SETUP] Create frontend/src/components/NotificationHistory.vue skeleton component (empty template)
- [X] T006 [P] [SETUP] Create frontend/src/components/NotificationDetail.vue skeleton component (empty template)
- [X] T007 [P] [SETUP] Create frontend/src/components/Pagination.vue skeleton component (empty template)
- [X] T008 [P] [SETUP] Create frontend/src/views/History.vue skeleton view (empty template)

**Checkpoint**: ✅ Project structure ready - foundation work can begin

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Database Foundation

- [X] T009 [FOUNDATION] Implement NotificationLogger worker goroutine in backend/internal/storage/logger.go with buffered channel (1000 buffer, 5s flush, 100 batch size) per research.md
- [X] T010 [FOUNDATION] Implement flushBatch method in backend/internal/storage/logger.go for batch INSERT statements (100 entries per transaction)
- [X] T011 [FOUNDATION] Implement Log method in backend/internal/storage/logger.go with non-blocking channel send and overflow error logging
- [X] T012 [FOUNDATION] Implement graceful shutdown in backend/internal/storage/logger.go (close channel, drain queue with 30s timeout)
- [X] T013 [P] [FOUNDATION] Add retention cleanup query in backend/internal/storage/repository.go: `DELETE FROM notification_logs WHERE created_at < datetime('now', '-90 days')`
- [X] T014 [P] [FOUNDATION] Create buildHistoryQuery method in backend/internal/storage/repository.go implementing filter logic per data-model.md

### API Foundation

- [X] T015 [FOUNDATION] Extend backend/internal/api/routes.go to register new routes: GET /notifications/history, GET /notifications/:id, POST /providers/:id/test, GET /ready
- [X] T016 [P] [FOUNDATION] Create validation functions in backend/internal/api/validation.go for history query parameters (provider filters, date range, page size 1-100, sort order)
- [X] T017 [P] [FOUNDATION] Create validation function in backend/internal/api/validation.go for test request (provider exists, not rate-limited)

### Frontend Foundation

- [X] T018 [FOUNDATION] Extend frontend/src/services/api.ts to add methods: getNotificationHistory(filters, cursor, pageSize), getNotificationDetail(id), testProvider(providerId)
- [X] T019 [FOUNDATION] Add frontend/src/router/index.ts route for /history view
- [X] T020 [P] [FOUNDATION] Implement Pagination.vue component with cursor-based navigation (next/prev buttons, page size selector)
- [X] T021 [P] [FOUNDATION] Implement StatusBadge.vue updates to support "test" badge indicator (extend existing component)

### Logging Foundation (Constitution Principle VI)

- [ ] T022 [P] [FOUNDATION] Add structured logging context in backend/internal/logging/logger.go for notification log operations (operation, log_id, provider_id, execution_time)
- [ ] T023 [P] [FOUNDATION] Add structured logging context in backend/internal/logging/logger.go for provider test operations (test_initiator, provider_id, result, tested_at)
- [ ] T024 [P] [FOUNDATION] Implement sensitive data redaction for notification content in log output (truncate message to 100 chars, mask metadata values)

**Checkpoint**: ✅ Foundation 95% complete - core infrastructure ready for user story implementation

---

## Phase 3: User Story 1 - View Notification History (Priority: P1) 🎯 MVP

**Goal**: Administrators can view complete notification history including failed attempts with error details within 2 seconds

**Independent Test**: Send notifications via API (successful and failed), access /history view, verify all notifications displayed with accurate status, timestamps, and pagination

### Tests for User Story 1 (MANDATORY - Constitution Principle II) 🔴

> **CRITICAL: Write these tests FIRST, ensure they FAIL before implementation (TDD cycle)**

- [X] T025 [P] [US1] Contract test for GET /notifications/history in backend/tests/contract/history_test.go (test filtering by provider_id, status, date range, pagination)
- [X] T026 [P] [US1] Contract test for GET /notifications/:id in backend/tests/contract/history_test.go (test valid ID returns full details, invalid ID returns 404)
- [X] T027 [P] [US1] Integration test for end-to-end logging flow in backend/tests/integration/logging_test.go (send notification → verify logged → query history → verify returned)
- [X] T028 [P] [US1] Unit test for repository query methods in backend/tests/unit/repository_test.go (test buildHistoryQuery filter combinations, cursor pagination logic)
- [X] T029 [P] [US1] Frontend unit test for NotificationHistory.vue in frontend/tests/unit/NotificationHistory.test.ts (test data rendering, filter application, loading states)
- [X] T030 [P] [US1] Frontend unit test for Pagination.vue in frontend/tests/unit/Pagination.test.ts (test next/prev navigation, cursor updates, page size changes)

### Backend Implementation for User Story 1

- [ ] T031 [US1] Implement GetNotificationHistory method in backend/internal/storage/repository.go using buildHistoryQuery with cursor-based pagination, return ([]NotificationLogEntry, *nextCursor, error)
- [ ] T032 [US1] Implement GetNotificationByID method in backend/internal/storage/repository.go: query by id, return full NotificationLogEntry or 404 error
- [ ] T033 [US1] Integrate NotificationLogger.Log() calls in backend/internal/api/handlers.go sendNotification handler (log immediately on receive with status "pending")
- [ ] T034 [US1] Update provider send success logic in backend/internal/providers/*.go to update status to "sent" and set delivered_at timestamp
- [ ] T035 [US1] Update provider send failure logic in backend/internal/providers/*.go to update status to "failed" and set error_message
- [ ] T036 [US1] Implement getNotificationHistory handler in backend/internal/api/handlers.go: validate query params, call repository, format response with pagination metadata
- [ ] T037 [US1] Implement getNotificationDetail handler in backend/internal/api/handlers.go: validate id, call repository, return 404 if not found
- [ ] T038 [US1] Add structured logging in handlers for history queries: log query filters, execution time, result count (Constitution Principle VI)

### Frontend Implementation for User Story 1

### Frontend Implementation for User Story 1

- [X] T039 [P] [US1] Implement NotificationHistory.vue table component: display columns (provider, recipient, message preview, status badge, timestamp, details button)
- [X] T040 [P] [US1] Implement filter controls in NotificationHistory.vue: provider dropdown, status dropdown, date range picker, include_tests checkbox
- [X] T041 [US1] Implement NotificationDetail.vue modal component: display full notification details (all fields from NotificationLogEntry), show metadata as key-value list, display full error trace if failed
- [X] T042 [US1] Implement History.vue view: integrate NotificationHistory component, Pagination component, filter state management, API calls via services/api.ts
- [X] T043 [US1] Add loading skeleton to NotificationHistory.vue while fetching data (Constitution Principle III - UX consistency)
- [X] T044 [US1] Add empty state to NotificationHistory.vue when no results (Constitution Principle III - helpful messaging)
- [X] T045 [US1] Connect Dashboard.vue to History.vue via router-link navigation button

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently - can view all notification history with filtering, pagination, and full details

---

## Phase 4: User Story 2 - Test Provider Configuration (Priority: P2)

**Goal**: Administrators can click a "Test" button next to any provider, receive success/failure feedback within 10 seconds with specific error details

**Independent Test**: Configure a provider in config file, access UI, click Test button on provider card, verify test notification sent and result displayed in UI

### Tests for User Story 2 (MANDATORY - Constitution Principle II) 🔴

- [X] T046 [P] [US2] Contract test for POST /providers/:id/test in backend/tests/contract/provider_test_test.go (test valid provider returns success, invalid provider returns 404, rate limited returns 429)
- [X] T047 [P] [US2] Integration test for provider testing flow in backend/tests/integration/provider_test_test.go (configure provider → test → verify logged with is_test=true → verify displayed in history)
- [X] T048 [P] [US2] Unit test for test recipient configuration in backend/tests/unit/provider_test.go (test Telegram uses default_chat_id, Email uses test_recipient config)

### Backend Implementation for User Story 2

- [X] T049 [US2] Add last_test_at and last_test_status fields to Provider struct in backend/internal/providers/provider.go
- [X] T050 [US2] Implement GetTestRecipient() method on Provider interface in backend/internal/providers/provider.go (Telegram: use default_chat_id, Email: use test_recipient from config or default)
- [X] T051 [US2] Implement Test() method on Provider interface in backend/internal/providers/provider.go: call Send() with test message template, update last_test_at and last_test_status, return error or nil
- [X] T052 [US2] Create test message templates in backend/internal/providers/provider.go: Telegram "Test notification from notimulti server - [timestamp]", Email subject "Test from notimulti" body "Test notification from notimulti server - [timestamp]"
- [X] T053 [US2] Implement testProvider handler in backend/internal/api/handlers.go: validate provider exists, check rate limit (10s), create test NotificationRequest with is_test=true, call provider.Test(), log result, return ProviderTestResponse
- [X] T054 [US2] Implement rate limiting check in testProvider handler: verify last_test_at is > 10 seconds ago, return 429 with Retry-After header if violated
- [X] T055 [US2] Add structured logging for test operations in backend/internal/api/handlers.go: log test_initiator (always "UI" for now), provider_id, result, tested_at (Constitution Principle VI)
- [X] T056 [US2] Update ProviderSummary response in backend/internal/api/handlers.go listProviders to include last_test_at and last_test_status fields

### Frontend Implementation for User Story 2

- [X] T057 [P] [US2] Add "Test" button to ProviderCard.vue component with loading state (disable button while test in progress)
- [X] T058 [P] [US2] Implement test button click handler in ProviderCard.vue: call api.testProvider(providerId), show loading indicator, display result (success message or error details)
- [X] T059 [US2] Add test result display in ProviderCard.vue: show success toast notification or error modal with full error details and actionable guidance
- [X] T060 [US2] Add last_test_at and last_test_status display to ProviderCard.vue: show "Last tested: [timestamp] - [status]" below provider status
- [X] T061 [US2] Update NotificationHistory.vue to display "test" badge for notifications with is_test=true flag
- [X] T062 [US2] Add loading feedback in ProviderCard.vue within 100ms of button click (Constitution Principle III - UX consistency, NFR-006)
- [X] T063 [US2] Format test error messages in ProviderCard.vue to be user-friendly and actionable per NFR-008 (e.g., "Failed to connect to SMTP server at smtp.example.com:587 - check firewall rules")

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - can view history AND test providers from UI

---

## Phase 5: User Story 3 - Deploy as Container (Priority: P3)

**Goal**: DevOps engineers can deploy notification server using provided Docker/Kubernetes manifests with single command and have both API and UI accessible

**Independent Test**: Build Docker image, run with docker-compose or kubectl apply, verify backend API responds to /health and frontend UI loads in browser

### Tests for User Story 3 (MANDATORY - Constitution Principle II) 🔴

- [X] T064 [P] [US3] Integration test for Docker volume mounting in backend/tests/integration/container_test.go (start container with volumes, verify config loading, verify database persistence across restarts)
- [X] T065 [P] [US3] Contract test for GET /ready endpoint in backend/tests/contract/health_test.go (test database check, providers check, returns 503 if not ready)

### Docker Implementation for User Story 3

- [x] T066 [P] [US3] Create Dockerfile in repository root with multi-stage build: Stage 1 (node:18-alpine for frontend build), Stage 2 (golang:1.21-alpine for backend build with embedded frontend), Stage 3 (alpine:3.18 runtime) per research.md
- [x] T067 [P] [US3] Create .dockerignore in repository root excluding: .git, node_modules, backend/tests, frontend/tests, *.md, .github
- [x] T068 [US3] Update backend/cmd/server/main.go to use Go embed directive to serve frontend dist/ as static files at root path "/"
- [x] T069 [US3] Configure backend router in backend/internal/api/routes.go to serve embedded frontend static files for non-API routes (e.g., "/", "/history", "/assets/*")
- [x] T070 [US3] Add environment variable support in backend/internal/config/loader.go for: PORT (default 8080), LOG_LEVEL (default info), CONFIG_DIR (default /app/configs), DB_PATH (default /app/data/notifications.db), LOG_RETENTION_DAYS (default 90)
- [x] T071 [US3] Implement getReady handler in backend/internal/api/handlers.go: check database connection (query "SELECT 1"), check providers loaded, return 200 if all ok or 503 with specific check failures
- [x] T072 [US3] Add Dockerfile build optimization: use `-ldflags="-s -w"` for Go build, add ca-certificates and tzdata to Alpine runtime per research.md
- [x] T073 [US3] Configure Dockerfile to run as non-root user: `RUN adduser -D -u 1000 notimulti` and `USER notimulti` per FR-025

### Docker Compose Implementation for User Story 3

- [x] T074 [P] [US3] Create deploy/docker-compose.yml with notimulti service: build context, ports (8080:8080), environment variables, volume mounts (./configs:/app/configs:ro, ./data:/app/data), restart policy (unless-stopped)
- [x] T075 [P] [US3] Add health check to docker-compose.yml service: test with wget/curl to /api/v1/health, 30s interval, 10s timeout, 3 retries, 10s start_period

### Kubernetes Implementation for User Story 3

- [x] T076 [P] [US3] Create deploy/k8s/deployment.yaml with StatefulSet (replicas: 1): container spec, port 8080, environment variables, volume mounts (config ConfigMap, data PVC), security context (runAsUser 1000, fsGroup 1000) per research.md
- [x] T077 [P] [US3] Add liveness probe to deploy/k8s/deployment.yaml: httpGet /health, initialDelaySeconds 10, periodSeconds 30
- [x] T078 [P] [US3] Add readiness probe to deploy/k8s/deployment.yaml: httpGet /ready, initialDelaySeconds 5, periodSeconds 10
- [x] T079 [P] [US3] Add resource limits to deploy/k8s/deployment.yaml: requests (memory 128Mi, cpu 100m), limits (memory 512Mi, cpu 500m)
- [x] T080 [P] [US3] Create deploy/k8s/service.yaml: ClusterIP service exposing port 80 → targetPort 8080
- [x] T081 [P] [US3] Create deploy/k8s/configmap.yaml: template with placeholder provider configurations (telegram-alerts.json, email-prod.json)
- [x] T082 [P] [US3] Create deploy/k8s/pvc.yaml: PersistentVolumeClaim with ReadWriteOnce, 10Gi storage for database
- [x] T083 [P] [US3] Create deploy/k8s/ingress.yaml: optional Ingress resource for HTTPS access with host configuration
- [x] T084 [P] [US3] Create deploy/k8s/README.md with deployment instructions and kubectl commands

### Documentation for User Story 3

- [x] T085 [US3] Create quickstart.md in repository root (copy from specs/002-enhanced-deployment/quickstart.md) with docker-compose instructions, provider configuration examples, common tasks
- [x] T086 [US3] Add deployment documentation to README.md: link to quickstart.md, Kubernetes manifests, Docker Hub image (once published)

**Checkpoint**: All user stories 1, 2, AND 3 should work - can deploy with Docker/K8s and access full functionality (history + provider testing)

---

## Phase 6: User Story 4 - Automated Container Build & Publish (Priority: P4)

**Goal**: GitHub Actions automatically builds Docker image, runs tests, pushes to Docker Hub on push to main branch or version tags

**Independent Test**: Push commit to main branch or create version tag, verify GitHub Actions workflow runs successfully, check Docker Hub for published image

### Tests for User Story 4 (MANDATORY - Constitution Principle II) 🔴

> Note: CI/CD workflow itself IS the test - verify it executes all test suites and fails build if tests fail

- [x] T087 [P] [US4] Verify contract test stage in workflow blocks image build on failure
- [x] T088 [P] [US4] Verify integration test stage in workflow blocks image build on failure
- [x] T089 [P] [US4] Verify coverage check in workflow fails if coverage <80% per NFR-018

### GitHub Actions Implementation for User Story 4

- [x] T090 [P] [US4] Create .github/workflows/docker.yml with workflow triggers: push to main, tags v*.*.*, pull_request per research.md
- [x] T091 [US4] Add backend test job in .github/workflows/docker.yml: setup Go 1.21, run tests with race detector, check coverage ≥80%, upload coverage artifact
- [x] T092 [US4] Add frontend test job in .github/workflows/docker.yml: setup Node 18, npm ci, npm test, upload test results
- [x] T093 [US4] Add build job in .github/workflows/docker.yml (depends on test jobs): checkout, setup QEMU, setup Buildx, login to Docker Hub (if not PR), docker metadata for tagging, build-push-action
- [x] T094 [US4] Configure multi-arch build in .github/workflows/docker.yml: platforms linux/amd64,linux/arm64 using Buildx per research.md
- [x] T095 [US4] Configure layer caching in .github/workflows/docker.yml: cache-from type=gha, cache-to type=gha,mode=max per research.md
- [x] T096 [US4] Configure Docker metadata in .github/workflows/docker.yml: tags for branch (latest), version (v1.2.3, v1.2, v1), PR (pr-123), SHA (sha-abc123)
- [x] T097 [US4] Add Trivy vulnerability scan step in .github/workflows/docker.yml: scan image after build (if not PR), fail on CRITICAL/HIGH vulnerabilities per research.md security strategy
- [x] T098 [US4] Configure Docker Hub push condition in .github/workflows/docker.yml: only push if not PR (github.event_name != 'pull_request')
- [x] T099 [US4] Add Docker image labels in .github/workflows/docker.yml build step: commit SHA, build date, version, source URL per FR-036

### Documentation for User Story 4

- [x] T100 [US4] Add CI/CD documentation to README.md: explain workflow triggers, required GitHub Secrets (DOCKERHUB_USERNAME, DOCKERHUB_TOKEN), tagging strategy
- [x] T101 [US4] Create .github/workflows/README.md explaining workflow stages, how to configure Docker Hub credentials, troubleshooting common issues

**Checkpoint**: All user stories complete - full end-to-end workflow from development to automated deployment

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and final quality checks

### Code Quality (Constitution Principle I)

- [x] T102 [P] [POLISH] Code quality review - verify all functions <50 lines, files <300 lines across backend/internal/**
- [x] T103 [P] [POLISH] Code quality review - verify no duplicated logic between provider Send() and Test() methods (DRY principle)
- [x] T104 [P] [POLISH] Verify Dockerfile follows best practices: official base images, minimal layers, build cache optimization

### Testing & Coverage (Constitution Principle II)

- [x] T105 [P] [POLISH] Code coverage verification - run `go test -coverprofile=coverage.out ./...` and verify ≥80% coverage for new code
- [ ] T106 [P] [POLISH] Add additional unit tests for edge cases in backend/tests/unit/: corrupted database, log queue overflow, invalid filter combinations
- [x] T107 [P] [POLISH] Run full test suite with race detector: `go test -race ./...` to catch concurrency issues in async logger

### UX Consistency (Constitution Principle III)

- [x] T108 [P] [POLISH] UX consistency audit - verify error messages are user-friendly and actionable across all frontend components
- [x] T109 [P] [POLISH] UX consistency audit - verify loading states and skeletons are consistent (NotificationHistory, ProviderCard)
- [x] T110 [P] [POLISH] Accessibility audit - verify keyboard navigation works for Test button, pagination controls, filters
- [x] T111 [P] [POLISH] Mobile responsiveness test - verify UI works on tablet/smartphone viewports per NFR-009

### Performance (Constitution Principle IV)

- [x] T112 [P] [POLISH] Performance testing - verify notification history queries <1s for 100k records using seeded test database (NFR-001)
- [x] T113 [P] [POLISH] Performance testing - verify provider test operations complete within 10s including external API calls (NFR-002)
- [x] T114 [P] [POLISH] Performance testing - verify Docker image build <5 minutes full, <1 minute incremental (NFR-003)
- [x] T115 [P] [POLISH] Performance testing - verify UI loads notification history page in <2s p95 (NFR-005)
- [x] T116 [P] [POLISH] Performance profiling - run `EXPLAIN QUERY PLAN` on all history queries to verify index usage per research.md
- [x] T117 [P] [POLISH] Performance profiling - check for memory leaks in async logger worker goroutine using pprof

### Observability & Logging (Constitution Principle VI)

- [x] T118 [P] [POLISH] Logging audit - verify all operations have structured logs with proper context (operation, execution_time, result)
- [x] T119 [P] [POLISH] Logging audit - verify sensitive data redaction works (passwords, API tokens never logged even in debug mode per NFR-024)
- [x] T120 [P] [POLISH] Debug mode testing - verify LOG_LEVEL=debug enables verbose logging, LOG_LEVEL=error suppresses info logs
- [x] T121 [P] [POLISH] Verify container logs output JSON format to stdout for aggregation per NFR-022

### Documentation & Finalization

- [x] T122 [P] [POLISH] Update main README.md with Phase 2 features: notification history, provider testing, Docker deployment, CI/CD
- [x] T123 [P] [POLISH] Run quickstart.md validation - follow steps exactly and verify 5-minute deployment works
- [x] T124 [P] [POLISH] Security hardening - run `docker scan` or Trivy on final image to check for vulnerabilities
- [x] T125 [P] [POLISH] Dependency vulnerability scan - run `go list -json -m all | nancy sleuth` to check Go dependencies
- [x] T126 [P] [POLISH] Update CHANGELOG.md with Phase 2 release notes listing all new features and breaking changes
- [x] T127 [P] [POLISH] Create database migration script for Phase 1 → Phase 2 upgrade (add is_test column, create indexes) in backend/migrations/002_enhanced_deployment.sql

---

## Dependencies & Execution Order

### Phase Dependencies

1. **Setup (Phase 1)**: No dependencies - can start immediately
2. **Foundational (Phase 2)**: Depends on Setup completion - **BLOCKS all user stories**
3. **User Stories (Phase 3-6)**: All depend on Foundational phase completion
   - User stories CAN proceed in parallel if team capacity allows
   - OR sequentially in priority order: P1 → P2 → P3 → P4
4. **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1 - Notification History)**: Can start after Foundational - No dependencies on other stories
- **User Story 2 (P2 - Provider Testing)**: Can start after Foundational - Integrates with US1 (logs appear in history) but independently testable
- **User Story 3 (P3 - Containerization)**: Can start after Foundational - Packages US1 and US2 but can be tested with just backend health check
- **User Story 4 (P4 - CI/CD)**: Can start after Foundational - Requires Dockerfile from US3 but can test with manual Docker build initially

### Critical Path (MVP - User Story 1 Only)

```
Setup (T001-T008) 
  → Foundational (T009-T024) 
    → US1 Tests (T025-T030) 
      → US1 Backend (T031-T038) 
        → US1 Frontend (T039-T045)
          → Polish (selected tasks)
            → MVP Ready
```

**Estimated Time**: ~2-3 weeks for MVP (US1 only)

### Full Feature Path (All User Stories)

```
Setup (Phase 1) → Foundational (Phase 2) → FORK:
  ├─ US1 (Phase 3) → MVP checkpoint
  ├─ US2 (Phase 4) → Integration checkpoint
  ├─ US3 (Phase 5) → Deployment checkpoint
  └─ US4 (Phase 6) → CI/CD checkpoint
→ MERGE → Polish (Phase 7) → Production Ready
```

**Estimated Time**: ~4-6 weeks for full feature

### Within Each User Story

1. Tests (T025-T030 for US1) - Write tests FIRST, verify they FAIL
2. Backend implementation (T031-T038 for US1) - Core logic
3. Frontend implementation (T039-T045 for US1) - UI components
4. Verify tests PASS
5. Story complete - checkpoint and validate independently

### Parallel Opportunities

**Setup Phase**: All tasks marked [P] can run in parallel (T003-T008)

**Foundational Phase**: 
- Database work can parallel with API work: T013-T014 || T015-T017
- Frontend foundation can parallel with backend: T018-T021 || T009-T017
- Logging work can parallel with everything: T022-T024 || all other foundation

**Within User Story 1**:
- All tests can run in parallel: T025-T030
- Backend repository and frontend components can parallel: T031-T032 || T039-T041

**Within User Story 2**:
- All tests can run in parallel: T046-T048
- Frontend work can parallel with some backend: T057-T063 || T049-T052

**Within User Story 3**:
- All Docker manifests can be created in parallel: T066-T067 || T074-T075 || T076-T084

**Within User Story 4**:
- Workflow stages can be written in parallel: T091-T092 || T093-T099

**Polish Phase**: All tasks marked [P] can run in parallel (T102-T127)

### Parallel Team Strategy

**With 3 developers**:
1. All: Complete Setup + Foundational together (1 week)
2. Once Foundational done:
   - Dev A: User Story 1 (notification history)
   - Dev B: User Story 2 (provider testing)
   - Dev C: User Story 3 (Docker/K8s) + User Story 4 (CI/CD)
3. Integration week: Merge all stories, run full test suite
4. All: Polish phase together

---

## Parallel Example: Foundational Phase

```bash
# Terminal 1 - Database Foundation
Task: T009 - Implement NotificationLogger worker
Task: T010 - Implement flushBatch method
Task: T011 - Implement Log method
Task: T012 - Implement graceful shutdown

# Terminal 2 - Repository & API (can start in parallel)
Task: T013 - Add retention cleanup query
Task: T014 - Create buildHistoryQuery method
Task: T015 - Extend routes.go for new endpoints
Task: T016 - Create history validation functions
Task: T017 - Create test validation function

# Terminal 3 - Frontend Foundation (completely parallel)
Task: T018 - Extend api.ts with new methods
Task: T019 - Add /history route
Task: T020 - Implement Pagination component
Task: T021 - Update StatusBadge for test indicator

# Terminal 4 - Logging (completely parallel)
Task: T022 - Add log context for notification operations
Task: T023 - Add log context for test operations
Task: T024 - Implement sensitive data redaction
```

---

## Implementation Strategy

### MVP First (Recommended for User Story 1 Only)

**Rationale**: Deliver notification history quickly for immediate operational value

1. Complete Phase 1: Setup (8 tasks, ~1 day)
2. Complete Phase 2: Foundational (16 tasks, ~3-5 days)
3. **CHECKPOINT**: Foundation ready
4. Complete Phase 3: User Story 1 (21 tasks, ~5-7 days)
5. **STOP and VALIDATE**: Test User Story 1 independently
   - Send notifications via API
   - View in history UI
   - Filter by provider, status, date
   - Verify pagination works
   - Check full details modal
6. Deploy MVP (bare metal or manual Docker)
7. Gather feedback before proceeding to US2

**MVP Success Criteria**:
- ✅ SC-001: Administrators can view complete notification history within 2 seconds
- ✅ SC-008: System can query 100k+ records with pagination <2s
- ✅ NFR-001: History queries <1s for 100k records
- ✅ NFR-005: UI loads history page <2s p95

### Incremental Delivery (Recommended for Production)

**Rationale**: Each user story adds value without breaking previous functionality

1. Complete Setup + Foundational → Foundation checkpoint
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Add User Story 4 → Test independently → Deploy/Demo (Production Ready!)
6. Polish phase → Final release

**Checkpoints**:
- After US1: Can view notification history ✅
- After US2: Can view history AND test providers ✅
- After US3: Can deploy with Docker/Kubernetes ✅
- After US4: Can deploy automatically via CI/CD ✅
- After Polish: Production-grade quality ✅

### Parallel Team Strategy (Fast Track)

**Rationale**: Multiple developers work on independent stories simultaneously

**Prerequisites**: 
- Team of 3+ developers
- Clear communication about shared files (routes.go, api.ts)
- Git workflow with feature branches per story

**Execution**:
1. Week 1: All team - Setup + Foundational (together)
2. Week 2-3: Parallel development
   - Dev A: User Story 1 (branch: feature/notification-history)
   - Dev B: User Story 2 (branch: feature/provider-testing)
   - Dev C: User Story 3 (branch: feature/containerization)
3. Week 4: Integration
   - Merge US1 → main
   - Merge US2 → main (resolve conflicts in routes.go, api.ts)
   - Merge US3 → main
   - Dev C continues: User Story 4
4. Week 5: Polish + Production deployment

**Risk**: Merge conflicts in shared files (backend/internal/api/routes.go, frontend/src/services/api.ts, backend/internal/api/handlers.go) - mitigate with clear API contract and frequent communication

---

## Notes

### Task Format
- **[P]** tasks = Can run in parallel (different files, no shared dependencies)
- **[Story]** label = Maps task to specific user story for traceability (US1, US2, US3, US4, SETUP, FOUNDATION, POLISH)
- File paths included for every task requiring code changes

### TDD Workflow (Constitution Principle II)
1. Write test for feature (verify it FAILS - red)
2. Implement minimum code to pass test (green)
3. Refactor code for quality (keep tests green)
4. Commit when tests pass

### Success Criteria Mapping
- **SC-001, SC-008, NFR-001, NFR-005** → User Story 1
- **SC-002, NFR-002, NFR-006, NFR-008** → User Story 2
- **SC-003, SC-006, SC-007, NFR-003, NFR-026, NFR-027** → User Story 3
- **SC-004, SC-005, NFR-019, NFR-028, NFR-030** → User Story 4

### Avoiding Common Pitfalls
- ❌ Don't start user stories before Foundational phase complete
- ❌ Don't skip writing tests first (violates TDD)
- ❌ Don't implement stories in dependency order when they're independent
- ❌ Don't batch multiple stories in one commit/PR (breaks independent testing)
- ✅ Do checkpoint after each story and validate independently
- ✅ Do commit after each task or logical group
- ✅ Do parallelize work when possible (tasks marked [P])
- ✅ Do follow priority order if implementing sequentially (P1 → P2 → P3 → P4)
