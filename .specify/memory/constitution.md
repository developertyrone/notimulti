<!--
Sync Impact Report
==================
Version change: v1.0.0 → v1.1.0
Modified principles: None
Added sections:
  - Principle VI: Observability & Debug-First Logging (structured logging, configurable
    levels, context-rich logs, flexible configuration, sensitive data redaction)
Removed sections: None
Templates status:
  ✅ .specify/templates/plan-template.md - Added logging principle to Constitution Check
  ✅ .specify/templates/spec-template.md - Added NFR-014 through NFR-019 for logging requirements
  ✅ .specify/templates/tasks-template.md - Added logging infrastructure tasks to Foundational phase,
     added logging audit tasks to Polish phase
Follow-up TODOs: None
-->

# notimulti Constitution

## Core Principles

### I. Code Quality First

All code MUST be clean, readable, and maintainable. Code reviews MUST verify:
- Clear, descriptive naming for variables, functions, and classes
- Single Responsibility Principle adherence
- DRY (Don't Repeat Yourself) - no duplicated logic
- Maximum function length: 50 lines (excluding tests)
- Maximum file length: 300 lines (excluding comprehensive test files)

**Rationale**: Technical debt compounds exponentially. Quality-first prevents
maintenance burden and ensures long-term velocity. Readable code is debuggable
code.

### II. Test-Driven Development (NON-NEGOTIABLE)

Testing is mandatory for all production code. The TDD cycle MUST be followed:
1. Write tests that capture acceptance criteria
2. Verify tests fail (red state)
3. Implement minimum code to pass tests (green state)
4. Refactor while keeping tests green

All tests MUST include:
- Contract tests for all public APIs/interfaces
- Integration tests for critical user journeys
- Unit tests only when justified for complex logic

**Rationale**: Tests written after implementation often miss edge cases and
don't verify actual requirements. TDD ensures testability by design and
provides living documentation.

### III. User Experience Consistency

All user-facing features MUST deliver consistent, intuitive experiences:
- Response times <200ms for UI interactions, <2s for API responses (p95)
- Clear error messages that guide users to resolution
- Consistent UI patterns, terminology, and interaction models
- Accessibility compliance (WCAG 2.1 Level AA minimum)
- Mobile-responsive design for all web interfaces

**Rationale**: Inconsistent UX creates cognitive load, reduces adoption, and
generates support burden. Users should never have to "learn" multiple patterns
within the same product.

### IV. Performance is a Feature

Performance MUST be measurable and monitored:
- Define performance budgets during spec phase
- Load test critical paths before production
- Monitor p50, p95, p99 latency for all operations
- Database queries MUST use indexes (verified via EXPLAIN)
- Memory leaks MUST be prevented (verified via profiling)

No performance regressions allowed without explicit justification.

**Rationale**: Performance issues compound with scale. Prevention costs less
than remediation. Slow software loses users.

### V. Keep It Simple, Stupid (KISS)

Simplicity MUST be the default. For any design decision:
- Choose boring, proven technology over trendy options
- Prefer standard library over external dependencies
- Solve today's problem, not tomorrow's hypothetical
- YAGNI (You Aren't Gonna Need It) - no speculative features

Complexity requires explicit justification documenting:
- What simpler alternatives were considered
- Why simpler approaches are insufficient
- What risks the complexity introduces

**Rationale**: Every dependency is a liability. Every abstraction has cost.
Simple systems are debuggable, maintainable, and resilient.

### VI. Observability & Debug-First Logging

All systems MUST be observable and debuggable in production. Logging
requirements:
- Structured logging with consistent format (JSON recommended for machine
  parsing)
- Configurable log levels (DEBUG, INFO, WARN, ERROR) without code changes
- Context-rich logs including: request IDs, user IDs, timestamps, operation
  names
- Critical paths MUST log: entry/exit points, decision branches, external calls
- Errors MUST log: stack traces, input parameters, system state
- Performance metrics logged: operation duration, resource usage

Flexibility requirements:
- Log levels adjustable via environment variables or config files
- Debug mode enabling without redeployment
- Log output destinations configurable (stdout, files, aggregation services)
- Sensitive data MUST be redacted from logs (passwords, tokens, PII)

**Rationale**: Production issues are inevitable. Without comprehensive logging,
debugging becomes archaeological guesswork. Flexible logging configuration
enables rapid troubleshooting without deployment cycles.

## Development Standards

### Code Style

- Automated formatting enforced via pre-commit hooks
- Linting with zero warnings tolerated
- Type hints/annotations required for all public interfaces
- Comments explain "why", not "what" (code should be self-documenting)

### Version Control

- Atomic commits with descriptive messages following conventional commits
- Feature branches from main, squash merge after approval
- No direct commits to main
- Branch naming: `###-feature-name` where ### is issue number

### Documentation

- README.md MUST include setup, usage, and contribution guide
- API documentation auto-generated from code annotations
- Architecture Decision Records (ADRs) for significant design choices
- Quickstart guides that can be executed as tests

## Quality Gates

All pull requests MUST pass:

1. **Automated Gates**:
   - All tests passing (contract, integration, unit if present)
   - Code coverage ≥80% for new code
   - Linting with zero warnings
   - No security vulnerabilities in dependencies

2. **Manual Review Gates**:
   - Code review approval from maintainer
   - Constitution compliance verified
   - Performance impact assessed
   - Documentation updated

3. **Pre-Deployment Gates**:
   - Smoke tests in staging environment
   - Load test results within performance budgets
   - Rollback plan documented

## Governance

### Amendment Process

Constitution changes require:
1. Proposal documenting rationale and impact
2. Review by project maintainers
3. Update of affected templates and documentation
4. Version increment following semantic versioning

### Versioning Policy

- **MAJOR**: Backward-incompatible principle changes or removals
- **MINOR**: New principles added or existing principles expanded
- **PATCH**: Clarifications, wording improvements, non-semantic fixes

### Compliance Reviews

All feature specifications and implementation plans MUST verify compliance
with these principles. Non-compliance requires documented justification
approved by project maintainers.

### Enforcement

Constitution supersedes all other documentation. When conflicts arise,
constitution principles take precedence. Repeated violations may result in
contribution privileges being revoked.

**Version**: 1.1.0 | **Ratified**: 2025-11-03 | **Last Amended**: 2025-11-03
