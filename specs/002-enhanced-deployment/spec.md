# Feature Specification: Enhanced Deployment & Operations

**Feature Branch**: `002-enhanced-deployment`  
**Created**: 2025-11-06  
**Status**: Draft  
**Input**: User description: "notification log to store in the sqlite db for all notifications, even the send is not successful; one click button in each notification instance, so we can test it immediately and get success / error; dockerfile to build the all-in-one image to include backend and frontend; ci action github action pipeline to push dockerhub; k8s and docker compose sample in 'deploy' folder"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Notification History (Priority: P1)

An administrator or developer needs to troubleshoot notification delivery issues or audit notification activity. They access the web UI to view a complete history of all notification attempts, including successful sends and failures, with full details of errors that occurred.

**Why this priority**: This is critical for operational visibility and troubleshooting. Without notification history, it's impossible to diagnose delivery failures or track notification activity. This directly addresses a gap from Phase 1 where notifications had no persistence.

**Independent Test**: Can be fully tested by sending various notifications (successful and failed) via the API and verifying the UI displays all attempts with accurate status, timestamps, and error details. Delivers immediate value for debugging and auditing.

**Acceptance Scenarios**:

1. **Given** notifications have been sent via the API, **When** a user accesses the notification history view in the UI, **Then** all notification attempts are displayed with provider name, recipient, timestamp, and delivery status (sent/failed/pending)
2. **Given** a notification delivery failed, **When** viewing the notification history, **Then** the failed entry displays the specific error message explaining why delivery failed
3. **Given** the notification history contains 100+ entries, **When** viewing the history, **Then** results are paginated with 50 entries per page and include navigation controls
4. **Given** multiple providers exist, **When** viewing notification history, **Then** users can filter notifications by provider ID or provider type
5. **Given** notifications were sent over multiple days, **When** viewing history, **Then** users can filter by date range (last 24 hours, last 7 days, custom range)
6. **Given** a notification was never sent (validation failure), **When** viewing history, **Then** the attempt is still logged with the validation error details

---

### User Story 2 - Test Provider Configuration (Priority: P2)

An administrator configures a new notification provider (e.g., adding a new Telegram bot or SMTP server) and needs to verify it works correctly before applications start using it. From the web UI, they click a "Test" button next to the provider configuration, which immediately sends a test notification and displays success or failure with specific error details.

**Why this priority**: Essential for validating provider configurations during setup and maintenance. Prevents applications from encountering provider errors in production. Can be tested independently of notification history.

**Independent Test**: Can be tested by configuring a provider in the config files, accessing the UI, clicking the test button, and verifying a test notification is sent and the result is displayed. Works even without notification history feature.

**Acceptance Scenarios**:

1. **Given** a Telegram provider is configured, **When** an administrator clicks the "Test" button on that provider in the UI, **Then** a test message is sent to the configured Telegram chat and success status is displayed within 5 seconds
2. **Given** an Email provider is configured, **When** an administrator clicks the "Test" button, **Then** a test email is sent to a default test recipient and success status is displayed within 10 seconds
3. **Given** a provider has invalid credentials, **When** the test button is clicked, **Then** the UI displays a clear error message indicating authentication failed with the external service
4. **Given** a provider's external service is unreachable, **When** the test button is clicked, **Then** the UI displays a connectivity error with timeout details
5. **Given** a test notification is in progress, **When** waiting for results, **Then** the UI shows a loading indicator and prevents multiple simultaneous tests
6. **Given** a test notification succeeds or fails, **When** results are displayed, **Then** the test attempt is logged in the notification history with status "test"

---

### User Story 3 - Deploy as Container (Priority: P3)

A DevOps engineer needs to deploy the notification server in a containerized environment (Docker, Kubernetes). They use the provided Dockerfile to build a single container image that includes both the backend API server and frontend UI, then deploy using sample docker-compose.yml or Kubernetes manifests.

**Why this priority**: Enables modern deployment patterns and simplifies infrastructure setup. Can be implemented independently as it's purely packaging/deployment focused.

**Independent Test**: Can be tested by building the Docker image, running it with docker-compose or Kubernetes, and verifying both backend API and frontend UI are accessible and functional. Delivers value for production-ready deployments.

**Acceptance Scenarios**:

1. **Given** the Dockerfile exists in the repository root, **When** a DevOps engineer runs `docker build -t notimulti:latest .`, **Then** the image builds successfully in under 5 minutes and includes both backend and frontend
2. **Given** a docker-compose.yml file exists in the deploy/ folder, **When** running `docker-compose up`, **Then** the notification server starts with backend API on port 8080 and frontend UI on port 80
3. **Given** Kubernetes manifests exist in deploy/k8s/, **When** applying them with `kubectl apply -f deploy/k8s/`, **Then** the notification server deploys with backend and frontend services accessible via ingress
4. **Given** the container is running, **When** configuration files are mounted as a volume to /app/configs, **Then** the server loads provider configurations from the mounted directory
5. **Given** the container is running, **When** the SQLite database is mounted as a volume, **Then** notification logs persist across container restarts
6. **Given** environment variables are provided (LOG_LEVEL, PORT), **When** starting the container, **Then** the server respects these configuration overrides

---

### User Story 4 - Automated Container Build & Publish (Priority: P4)

A developer merges code changes to the main branch and needs the container image automatically built and published to Docker Hub for easy deployment. GitHub Actions CI pipeline automatically builds the Docker image, runs tests, and pushes the tagged image to Docker Hub on successful builds.

**Why this priority**: Automates the release process but not essential for manual deployments. Can be implemented after containerization works manually.

**Independent Test**: Can be tested by triggering the GitHub Actions workflow (push to main or tag) and verifying the image appears on Docker Hub with correct tags. Delivers value for continuous deployment.

**Acceptance Scenarios**:

1. **Given** code is pushed to the main branch, **When** the GitHub Actions workflow runs, **Then** it builds the Docker image, runs all tests, and pushes the image with "latest" tag to Docker Hub
2. **Given** a version tag is created (e.g., v1.2.0), **When** the GitHub Actions workflow runs, **Then** it builds and pushes the Docker image with both the version tag and "latest" tag
3. **Given** tests fail during the workflow, **When** the build completes, **Then** the workflow fails and no image is pushed to Docker Hub
4. **Given** Docker Hub credentials are configured as GitHub secrets, **When** the workflow attempts to push, **Then** authentication succeeds and the image is published
5. **Given** the workflow completes successfully, **When** checking Docker Hub, **Then** the image is available with build metadata (commit SHA, build date) in labels
6. **Given** a pull request is created, **When** the GitHub Actions workflow runs, **Then** it builds the image and runs tests but does not push to Docker Hub

---

### Edge Cases

- What happens when the SQLite database file becomes corrupted or locked?
- How does the system handle notification history queries when the database contains millions of records?
- What happens when clicking "Test" on a provider while a configuration reload is in progress?
- How does the Docker container behave when the config volume is unmounted or becomes read-only?
- What happens when GitHub Actions workflow attempts to push to Docker Hub but credentials are invalid or expired?
- How does the UI handle displaying notifications with extremely long error messages or message content?
- What happens when the Docker build fails due to frontend build errors but backend compiles successfully?
- How does Kubernetes handle rolling updates when the new image version has database schema changes?

## Requirements *(mandatory)*

### Functional Requirements

**Notification History & Logging**:

- **FR-001**: System MUST store all notification attempts in SQLite database, including both successful and failed sends
- **FR-002**: System MUST store the following fields for each notification attempt: unique ID, provider ID, provider type, recipient, message content, subject (if applicable), metadata, priority, status (sent/failed/pending/retrying), error message (if failed), number of attempts, created timestamp, delivered timestamp (if successful)
- **FR-003**: UI MUST display a notification history page showing all logged notification attempts
- **FR-004**: Notification history UI MUST show: provider name, recipient, message preview (first 100 characters), status badge, timestamp, and details button
- **FR-005**: Notification history UI MUST support pagination with configurable page size (default 50 entries per page)
- **FR-006**: Notification history UI MUST support filtering by: provider ID, provider type, status, date range
- **FR-007**: Notification history UI MUST support sorting by timestamp (newest first, oldest first)
- **FR-008**: System MUST display full notification details when user clicks on a history entry, including complete message, all metadata, full error trace if failed
- **FR-009**: System MUST retain notification logs for 90 days by default with configurable retention period
- **FR-010**: System MUST log notifications that fail API validation (before sending) with the validation error details

**Provider Testing**:

- **FR-011**: UI MUST display a "Test" button next to each configured provider in the provider list
- **FR-012**: System MUST send a test notification when the Test button is clicked, using the provider's configuration
- **FR-013**: Test notification for Telegram MUST send message "Test notification from notimulti server - [timestamp]" to the default chat ID
- **FR-014**: Test notification for Email MUST send an email with subject "Test from notimulti" and body "Test notification from notimulti server - [timestamp]" to a configurable test recipient address
- **FR-015**: UI MUST display test result within 10 seconds showing either success or specific error message
- **FR-016**: UI MUST show loading indicator while test is in progress and disable the Test button to prevent duplicate tests
- **FR-017**: System MUST log test notification attempts in the notification history with a "test" label or tag
- **FR-018**: System MUST validate provider configuration before attempting test send and display validation errors immediately

**Containerization & Deployment**:

- **FR-019**: System MUST provide a Dockerfile that builds a single container image including both backend Go application and frontend Vue.js application
- **FR-020**: Docker image MUST serve frontend static files through the backend server (eliminate separate frontend server)
- **FR-021**: Docker image MUST expose a single HTTP port (default 8080) serving both API endpoints and frontend UI
- **FR-022**: Docker image MUST support configuration via environment variables: PORT, LOG_LEVEL, CONFIG_DIR, DB_PATH
- **FR-023**: System MUST provide a docker-compose.yml file in deploy/ folder demonstrating how to run the service with volume mounts for configs and database
- **FR-024**: System MUST provide Kubernetes manifests in deploy/k8s/ folder including: Deployment, Service, ConfigMap for provider configs, PersistentVolumeClaim for database
- **FR-025**: Docker container MUST use non-root user for running the application process
- **FR-026**: Docker image MUST use multi-stage build to minimize final image size (target < 100MB)
- **FR-027**: Kubernetes deployment MUST include health check endpoints (/health, /ready) and configure readiness/liveness probes

**CI/CD Pipeline**:

- **FR-028**: System MUST provide GitHub Actions workflow file (.github/workflows/docker.yml) for automated Docker builds
- **FR-029**: GitHub Actions workflow MUST trigger on: push to main branch, version tags (v*.*.*), pull requests
- **FR-030**: Workflow MUST build Docker image using the Dockerfile
- **FR-031**: Workflow MUST run all backend unit tests, integration tests, and contract tests before building final image
- **FR-032**: Workflow MUST run frontend unit tests and build frontend before final image build
- **FR-033**: Workflow MUST push Docker image to Docker Hub with tags: "latest" (main branch), version number (version tags), PR number (pull requests)
- **FR-034**: Workflow MUST use GitHub Secrets for Docker Hub credentials (DOCKERHUB_USERNAME, DOCKERHUB_TOKEN)
- **FR-035**: Workflow MUST fail if any tests fail and prevent image push
- **FR-036**: Workflow MUST add Docker image labels with metadata: commit SHA, build date, version, source URL
- **FR-037**: Workflow for pull requests MUST build and test but NOT push to Docker Hub

### Non-Functional Requirements (Constitution-mandated)

**Performance** (Principle IV):
- **NFR-001**: Notification history queries MUST return results in <1 second for datasets up to 100,000 records
- **NFR-002**: Provider test operations MUST complete within 10 seconds including external API calls
- **NFR-003**: Docker image build MUST complete in <5 minutes for full rebuild, <1 minute for incremental builds
- **NFR-004**: Database write operations for logging MUST not block notification sending (async writes)
- **NFR-005**: UI MUST load notification history page in <2 seconds (p95)

**User Experience** (Principle III):
- **NFR-006**: Test button MUST provide immediate feedback (loading state) within 100ms of click
- **NFR-007**: Notification history MUST show loading skeleton while fetching data
- **NFR-008**: Error messages from test operations MUST be user-friendly and actionable (e.g., "Failed to connect to SMTP server at smtp.example.com:587 - check firewall rules" instead of "dial tcp: i/o timeout")
- **NFR-009**: UI MUST be mobile-responsive for viewing notification history on tablets and smartphones
- **NFR-010**: Pagination controls MUST be intuitive with page numbers, next/prev buttons, and jump-to-page input

**Code Quality** (Principle I):
- **NFR-011**: Functions MUST not exceed 50 lines
- **NFR-012**: Files MUST not exceed 300 lines (excluding tests)
- **NFR-013**: MUST follow DRY principle with no duplicated logic between provider test and normal send operations
- **NFR-014**: Dockerfile MUST follow best practices: use official base images, minimize layers, leverage build cache

**Testing** (Principle II):
- **NFR-015**: Contract tests MUST cover new API endpoints for notification history and provider testing
- **NFR-016**: Integration tests MUST verify end-to-end notification logging flow (API request → database → UI display)
- **NFR-017**: Integration tests MUST verify Docker container startup and configuration loading from volumes
- **NFR-018**: Code coverage MUST be ≥80% for new notification history and testing features
- **NFR-019**: GitHub Actions workflow MUST include test stage that fails the build if coverage drops below threshold

**Observability & Logging** (Principle VI):
- **NFR-020**: MUST log all provider test attempts with test initiator (user ID or "UI"), timestamp, provider ID, result
- **NFR-021**: MUST log database operations (writes, queries) with execution time for performance monitoring
- **NFR-022**: Docker container MUST output logs to stdout in JSON format for aggregation by logging systems
- **NFR-023**: MUST log container startup events including configuration loading and database initialization
- **NFR-024**: MUST NOT log sensitive data (passwords, API tokens) even in debug mode
- **NFR-025**: Kubernetes manifests MUST configure log aggregation labels for filtering logs by component

**Security & Reliability**:
- **NFR-026**: Docker container MUST run as non-root user (UID >= 1000)
- **NFR-027**: Database file MUST have restricted permissions (0600) when created in container
- **NFR-028**: GitHub Actions workflow MUST NOT expose Docker Hub credentials in logs
- **NFR-029**: Kubernetes secrets MUST be used for sensitive configuration data (API tokens, SMTP passwords)
- **NFR-030**: Docker image MUST be scanned for vulnerabilities in CI pipeline (fail build on critical CVEs)

### Key Entities

- **Notification Log Entry**: Represents a persisted record of a notification attempt in SQLite. Contains all notification details, delivery status, error information, and timestamps. Primary key is auto-incrementing ID. Indexed by provider_id, status, and created_at for efficient querying.

- **Test Notification Request**: Represents a user-initiated test of a provider configuration. Contains provider ID, test timestamp, initiator, and result (success/failure with details). Logged in notification history with special "test" type indicator.

- **Docker Container Configuration**: Represents runtime configuration of the containerized application. Includes environment variables (PORT, LOG_LEVEL, CONFIG_DIR, DB_PATH), volume mounts (config directory, database file), and exposed ports.

- **GitHub Actions Workflow**: Represents the CI/CD pipeline definition. Contains build steps, test execution, image tagging strategy, and publish conditions. Configured via YAML file in .github/workflows/.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can view complete notification history including failed attempts with error details within 2 seconds of accessing the history page
- **SC-002**: Administrators can test any provider configuration and receive success/failure feedback within 10 seconds
- **SC-003**: DevOps engineers can deploy the notification server using docker-compose with a single command (`docker-compose up`) and have both API and UI accessible
- **SC-004**: Docker image builds successfully in GitHub Actions CI pipeline in under 5 minutes with all tests passing
- **SC-005**: Docker images are automatically published to Docker Hub within 10 minutes of merging to main branch
- **SC-006**: Kubernetes deployment succeeds using provided manifests and service is accessible within 2 minutes of applying manifests
- **SC-007**: Notification logs persist across container restarts when using volume mounts
- **SC-008**: System can query and display notification history containing 100,000+ records with pagination loading in under 2 seconds per page
- **SC-009**: 100% of provider test failures display actionable error messages that identify the specific configuration or connectivity issue
- **SC-010**: Docker image size is under 100MB and contains no critical security vulnerabilities
