# Implementation Plan: Enhanced Deployment & Operations

**Branch**: `002-enhanced-deployment` | **Date**: 2025-11-06 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `/specs/002-enhanced-deployment/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Phase 2 enhances the notification server with operational capabilities: persistent notification logging in SQLite for troubleshooting, one-click provider testing from the UI, containerized deployment with Docker/Kubernetes manifests, and automated CI/CD pipeline for publishing to Docker Hub. This builds on Phase 1 (001-notification-server) to make the system production-ready.

## Technical Context

**Language/Version**: Go 1.21+ (backend), Vue 3.3+ with TypeScript (frontend)  
**Primary Dependencies**: 
- Backend: Chi router, SQLite driver (modernc.org/sqlite), existing provider implementations
- Frontend: Vue Router, Axios, Tailwind CSS, Vite build system
- Deployment: Docker multi-stage build, GitHub Actions
- Database: SQLite with async write buffering

**Storage**: 
- SQLite database file for notification logs (90-day retention, indexed by provider_id, status, created_at)
- Existing config file watching from Phase 1
- Volume mounts for persistent data in containers

**Testing**: 
- Backend: Go testing stdlib, existing contract/integration/unit test structure
- Frontend: Vitest for unit tests
- Container: Integration tests for docker-compose and volume mounting
- CI: Full test suite execution before image build

**Target Platform**: 
- Linux containers (Docker) for deployment
- Kubernetes clusters (1.20+) 
- Local development via docker-compose
- Multi-arch support (amd64, arm64) via GitHub Actions

**Project Type**: Web application (existing backend + frontend structure)

**Performance Goals**: 
- Notification history queries: <1s for 100k records (indexed queries)
- Provider tests: <10s including external API calls
- Docker build: <5min full, <1min incremental (multi-stage caching)
- Database writes: Async, non-blocking to notification sending

**Constraints**: 
- Docker image size: <100MB (Alpine base, multi-stage build)
- Database write latency: Must not block notification API (buffered writes)
- UI response: <2s for history page load with pagination
- Zero downtime: Volume mounts preserve data across container restarts

**Scale/Scope**: 
- Support 100k+ notification log entries with performant pagination
- Handle concurrent provider tests without race conditions
- CI pipeline completes in <10min end-to-end
- Support multiple deployment targets (docker-compose, K8s, bare metal)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Code Quality First**: Design follows Single Responsibility (separate concerns: logging service, test handler, container build), no duplication (reuse existing provider send logic for tests), functions <50 lines, files <300 lines
- [x] **Test-Driven Development**: TDD cycle will be followed - write contract tests for new API endpoints, integration tests for logging flow and container deployment, unit tests for database operations
- [x] **UX Consistency**: Response time budgets defined (<100ms test button feedback, <1s history queries, <2s page load p95), clear error messages for test failures, consistent UI patterns extending Phase 1 design
- [x] **Performance is a Feature**: Performance budgets defined (history queries <1s for 100k records via indexes, async DB writes non-blocking, Docker build <5min), will monitor query performance and build times
- [x] **KISS Principle**: Using SQLite (already in use), standard Docker multi-stage builds, GitHub Actions (platform standard), no complex orchestration - straightforward implementation extending Phase 1 patterns
- [x] **Observability & Debug-First Logging**: Structured logging for all new operations (test attempts, DB operations, container startup), JSON format for container logs, configurable log levels via env vars, sensitive data redaction maintained

**Gate Status**: ✅ PASSED - All principles satisfied with clear justifications

## Project Structure

### Documentation (this feature)

```text
specs/002-enhanced-deployment/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── openapi.yaml     # Extended API spec with history & test endpoints
├── checklists/
│   └── requirements.md  # Already created
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Existing structure from Phase 1 (001-notification-server)
backend/
├── cmd/
│   └── server/
│       └── main.go           # Add static file serving for frontend
├── internal/
│   ├── api/
│   │   ├── handlers.go       # Extend: Add history & test endpoints
│   │   ├── middleware.go     # Existing
│   │   ├── routes.go         # Extend: Register new routes
│   │   └── validation.go     # Existing
│   ├── config/               # Existing
│   ├── logging/              # Existing
│   ├── providers/            # Existing
│   └── storage/
│       ├── logger.go         # NEW: Notification logger service
│       ├── repository.go     # NEW: Query/persistence methods
│       ├── schema.go         # Extend: Add indexes, update schema
│       └── sqlite.go         # Extend: Connection management
└── tests/
    ├── contract/
    │   ├── health_test.go           # Existing
    │   ├── notifications_test.go    # Existing
    │   ├── history_test.go          # NEW: Test history endpoints
    │   └── provider_test_test.go    # NEW: Test provider testing endpoint
    ├── integration/
    │   ├── logging_test.go          # NEW: End-to-end logging flow
    │   └── container_test.go        # NEW: Docker volume/config tests
    └── unit/
        └── repository_test.go       # NEW: Database query tests

frontend/
├── src/
│   ├── components/
│   │   ├── ProviderCard.vue        # Extend: Add Test button
│   │   ├── NotificationHistory.vue # NEW: History table component
│   │   ├── NotificationDetail.vue  # NEW: Modal for full details
│   │   └── Pagination.vue          # NEW: Pagination controls
│   ├── services/
│   │   └── api.ts                  # Extend: Add history & test methods
│   └── views/
│       ├── Dashboard.vue           # Existing: Shows providers
│       └── History.vue             # NEW: History page with filters
└── tests/
    └── unit/
        ├── NotificationHistory.test.ts  # NEW
        └── Pagination.test.ts           # NEW

# NEW: Deployment artifacts
Dockerfile                    # Multi-stage: frontend build + backend + static serving
.dockerignore
deploy/
├── docker-compose.yml       # Single service with volume mounts
└── k8s/
    ├── deployment.yaml      # Deployment with health checks
    ├── service.yaml         # Service exposing port 8080
    ├── configmap.yaml       # Provider configs
    ├── pvc.yaml            # PersistentVolumeClaim for database
    └── ingress.yaml        # Optional ingress

# NEW: CI/CD
.github/
└── workflows/
    └── docker.yml          # Build, test, publish pipeline
```

**Structure Decision**: Extending Phase 1 web application structure. Backend adds storage layer for notification logs and API endpoints for history/testing. Frontend adds new views and components for history display. New deployment artifacts (Dockerfile, manifests, CI workflow) are added at repository root and deploy/ folder.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitution violations. All complexity is justified:
- SQLite for logging: Already used in Phase 1 schema, simple extension
- Docker multi-stage build: Industry standard for minimizing image size
- GitHub Actions: Platform-native CI/CD, no additional complexity
- Async database writes: Necessary to meet NFR-004 (non-blocking), simple buffered channel pattern

---

## Phase 0: Research & Technical Decisions

### Research Tasks

1. **Docker Multi-Stage Build Optimization**
   - Research: Best practices for Go + Vue multi-stage builds
   - Decision needed: Base image selection (Alpine vs Distroless)
   - Decision needed: Static file embedding vs runtime serving
   - Output: Build strategy with size/security tradeoffs

2. **Async Database Writing Pattern**
   - Research: Go patterns for buffered database writes
   - Decision needed: Channel buffer size and flush strategy
   - Decision needed: Error handling for failed writes
   - Output: Implementation pattern with guarantees

3. **GitHub Actions Docker Build**
   - Research: Best practices for Docker Hub publishing
   - Decision needed: Buildx for multi-arch support
   - Decision needed: Caching strategy for faster builds
   - Output: Workflow structure with optimization

4. **SQLite Performance at Scale**
   - Research: Index strategies for 100k+ records
   - Decision needed: Query pagination approach
   - Decision needed: Retention cleanup strategy
   - Output: Schema design with performance guarantees

5. **Kubernetes Deployment Patterns**
   - Research: Best practices for stateful app with SQLite
   - Decision needed: PVC strategy, backup approach
   - Decision needed: Health check configuration
   - Output: Manifest templates with production considerations
