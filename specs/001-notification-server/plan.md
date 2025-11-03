# Implementation Plan: Centralized Notification Server

**Branch**: `001-notification-server` | **Date**: 2025-11-03 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-notification-server/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a centralized notification server that accepts notification requests via REST API and delivers them through pluggable providers (Telegram, Email). The server dynamically loads provider configurations from files without restart, supports multiple instances of the same provider type, and includes a read-only web UI for monitoring. Technical approach: Go backend with Gin framework for REST API, Vue 3 + Tailwind CSS frontend, SQLite for metadata persistence, file-based configuration with file-watching for dynamic reload.

## Technical Context

**Language/Version**: Go 1.21+ (backend), JavaScript/TypeScript with Vue 3.3+ (frontend)
**Primary Dependencies**: 
- Backend: Gin (REST framework), fsnotify (file watching), go-telegram-bot-api (Telegram), gomail/smtp (Email)
- Frontend: Vue 3, Vite 5, Tailwind CSS 3, minimal additional libraries
**Storage**: SQLite 3 (metadata: provider status, notification history)
**Testing**: Go testing package + testify (backend), Vitest (frontend)
**Target Platform**: Linux/macOS/Windows server (backend), Modern browsers (frontend)
**Project Type**: Web application (backend + frontend)
**Performance Goals**: 100+ concurrent requests, <2s API response (p95), <200ms UI interactions (p95)
**Constraints**: <200ms UI p95, <2s API p95, config reload <5s, 30-second change detection
**Scale/Scope**: Small to medium scale (100s of applications), ~10-50 provider instances

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Code Quality First**: Design follows Single Responsibility (separate provider interfaces, config loader, API handlers), no duplication planned (provider interface abstracts common behavior)
- [x] **Test-Driven Development**: TDD cycle will be followed (tests → fail → implement → refactor) with contract tests for REST API, integration tests for providers
- [x] **UX Consistency**: Response time budgets defined (<200ms UI, <2s API p95), error messages standardized, mobile-responsive UI with Tailwind
- [x] **Performance is a Feature**: Performance budgets defined (100 concurrent requests, <2s API, <5s config reload), monitoring via structured logging
- [x] **KISS Principle**: Simple approach chosen - file-based config (no complex orchestration), SQLite (no heavyweight DB), minimal frontend libraries, standard REST patterns
- [x] **Observability & Debug-First Logging**: JSON structured logging via environment-configurable levels, all provider operations logged, sensitive data redacted

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── internal/
│   ├── config/
│   │   ├── loader.go         # Config file loading & validation
│   │   ├── watcher.go        # File system watching
│   │   └── types.go          # Config data structures
│   ├── providers/
│   │   ├── provider.go       # Provider interface
│   │   ├── telegram.go       # Telegram implementation
│   │   ├── email.go          # Email/SMTP implementation
│   │   └── registry.go       # Provider registry/factory
│   ├── api/
│   │   ├── handlers.go       # REST endpoint handlers
│   │   ├── middleware.go     # Logging, error handling
│   │   └── routes.go         # Route definitions
│   ├── storage/
│   │   └── sqlite.go         # SQLite metadata storage
│   └── logging/
│       └── logger.go         # Structured logging setup
├── tests/
│   ├── contract/             # API contract tests
│   ├── integration/          # Provider integration tests
│   └── unit/                 # Unit tests
├── configs/                  # Provider configuration files
└── go.mod

frontend/
├── src/
│   ├── components/
│   │   ├── ProviderCard.vue  # Provider display component
│   │   └── StatusBadge.vue   # Status indicator component
│   ├── views/
│   │   └── Dashboard.vue     # Main dashboard view
│   ├── services/
│   │   └── api.ts            # Backend API client
│   ├── App.vue
│   └── main.ts
├── public/
├── tests/
│   └── unit/
├── index.html
├── vite.config.ts
├── tailwind.config.js
└── package.json
```

**Structure Decision**: Web application structure (Option 2) chosen because the feature requires both a REST API backend and a web-based UI frontend. Backend handles notification routing and provider management. Frontend provides read-only visualization of server state. Separation allows independent deployment and testing of each layer.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations - all constitution principles satisfied. KISS principle upheld through:
- File-based configuration (simpler than database-driven config)
- SQLite for lightweight metadata (simpler than PostgreSQL/MySQL)
- Minimal frontend libraries (Vue 3 + Tailwind only)
- Standard REST patterns (no GraphQL complexity)
- Interface-based provider abstraction (simple polymorphism, not complex plugin system)
