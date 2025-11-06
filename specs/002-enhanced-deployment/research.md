# Research: Enhanced Deployment & Operations

**Feature**: 002-enhanced-deployment  
**Date**: 2025-11-06  
**Purpose**: Research technical decisions for notification logging, provider testing, containerization, and CI/CD

## 1. Docker Multi-Stage Build Optimization

### Decision: Alpine + Multi-Stage Build with Static Embedding

**Rationale**:
- **Alpine base**: Minimal size (~5MB base) vs Distroless (~20MB), security-focused with minimal attack surface
- **Multi-stage approach**: Separate build stages for frontend (Node), backend (Go), and final runtime
- **Static embedding**: Use Go embed directive to bundle frontend dist/ into binary, eliminates need for runtime file serving complexity
- **Target size**: <100MB final image (requirement FR-026)

**Build Strategy**:
```dockerfile
# Stage 1: Frontend build
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --only=production
COPY frontend/ ./
RUN npm run build

# Stage 2: Backend build with embedded frontend
FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
COPY backend/go.* ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-builder /app/frontend/dist ./internal/web/dist
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o server cmd/server/main.go

# Stage 3: Runtime
FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 1000 notimulti
USER notimulti
WORKDIR /app
COPY --from=backend-builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

**Alternatives Considered**:
- **Distroless**: Rejected - larger size, less tooling for debugging
- **Scratch base**: Rejected - no shell for troubleshooting, no ca-certificates
- **Runtime file serving**: Rejected - adds volume mount complexity, cache invalidation issues

**Size Optimization**:
- Go build flags: `-ldflags="-s -w"` (strip debug info) saves ~30%
- Alpine vs Ubuntu: ~80MB savings
- Embed vs separate frontend: Simpler deployment, no NGINX needed

---

## 2. Async Database Writing Pattern

### Decision: Buffered Channel with Batch Flushing

**Rationale**:
- **Non-blocking requirement**: NFR-004 mandates DB writes don't block notification sending
- **Simple Go pattern**: Buffered channel + worker goroutine is idiomatic and well-tested
- **Reliability**: Flush on shutdown, periodic flush for durability
- **Error handling**: Failed writes logged but don't crash application

**Implementation Pattern**:
```go
type NotificationLogger struct {
    db          *sql.DB
    logQueue    chan LogEntry
    flushTicker *time.Ticker
    wg          sync.WaitGroup
}

func NewNotificationLogger(db *sql.DB) *NotificationLogger {
    nl := &NotificationLogger{
        db:          db,
        logQueue:    make(chan LogEntry, 1000), // Buffer 1000 entries
        flushTicker: time.NewTicker(5 * time.Second),
    }
    nl.wg.Add(1)
    go nl.worker()
    return nl
}

func (nl *NotificationLogger) worker() {
    defer nl.wg.Done()
    batch := make([]LogEntry, 0, 100)
    
    for {
        select {
        case entry, ok := <-nl.logQueue:
            if !ok {
                nl.flushBatch(batch)
                return
            }
            batch = append(batch, entry)
            if len(batch) >= 100 {
                nl.flushBatch(batch)
                batch = batch[:0]
            }
        case <-nl.flushTicker.C:
            if len(batch) > 0 {
                nl.flushBatch(batch)
                batch = batch[:0]
            }
        }
    }
}

func (nl *NotificationLogger) Log(entry LogEntry) {
    select {
    case nl.logQueue <- entry:
        // Logged successfully
    default:
        // Queue full, log error but don't block
        log.Error().Msg("notification log queue full, dropping entry")
    }
}
```

**Configuration**:
- **Channel buffer**: 1000 entries (handles bursts)
- **Batch size**: 100 entries (SQLite INSERT efficiency)
- **Flush interval**: 5 seconds (durability vs performance tradeoff)
- **Shutdown**: Close channel, wait for worker to drain queue

**Alternatives Considered**:
- **Synchronous writes**: Rejected - blocks notification sending (violates NFR-004)
- **Write-through cache**: Rejected - added complexity, no clear benefit
- **External queue (Redis)**: Rejected - violates KISS principle, adds dependency

**Error Handling**:
- Database write failures: Log error, continue processing (don't crash)
- Queue overflow: Drop entry with error log (prevents memory exhaustion)
- Shutdown: Graceful drain with timeout (max 30s wait)

---

## 3. GitHub Actions Docker Build & Publish

### Decision: Buildx Multi-Arch with Layer Caching

**Rationale**:
- **Multi-arch support**: Build for amd64 and arm64 (Apple Silicon, AWS Graviton)
- **Buildx**: Docker's official multi-platform build tool
- **Layer caching**: Use GitHub Actions cache to speed up rebuilds
- **Security**: Scan images with Trivy before publishing

**Workflow Structure**:
```yaml
name: Docker Build & Publish

on:
  push:
    branches: [main]
    tags: ['v*.*.*']
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run backend tests
        run: cd backend && go test -v -race -coverprofile=coverage.out ./...
      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=backend/coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "Coverage $coverage% below 80% threshold"
            exit 1
          fi
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - name: Run frontend tests
        run: cd frontend && npm ci && npm test

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up QEMU
        uses: docker/setup-qemu-action@v3
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      - name: Login to Docker Hub
        if: github.event_name != 'pull_request'
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}
      - name: Docker meta
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ secrets.DOCKERHUB_USERNAME }}/notimulti
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
      - name: Run Trivy vulnerability scanner
        if: github.event_name != 'pull_request'
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ secrets.DOCKERHUB_USERNAME }}/notimulti:${{ steps.meta.outputs.version }}
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'
      - name: Fail on critical vulnerabilities
        if: github.event_name != 'pull_request'
        run: |
          if grep -q '"level": "error"' trivy-results.sarif; then
            echo "Critical vulnerabilities found"
            exit 1
          fi
```

**Optimization Strategies**:
- **Layer caching**: `cache-from: type=gha` uses GitHub Actions cache (~60% faster rebuilds)
- **Multi-stage benefits**: Only frontend/backend changes invalidate respective stages
- **Parallel platforms**: Buildx builds amd64/arm64 in parallel
- **Test before build**: Fail fast if tests fail (save build time)

**Tagging Strategy**:
- **main branch**: `latest` and `sha-<commit>`
- **Version tags**: `v1.2.3`, `v1.2`, `v1`
- **Pull requests**: `pr-<number>` (build only, don't push)

**Security**:
- **Trivy scanning**: Fail build on CRITICAL/HIGH vulnerabilities
- **Secrets handling**: Use GitHub Secrets, never expose in logs
- **Non-root user**: Enforced in Dockerfile (FR-025)

**Alternatives Considered**:
- **Jenkins**: Rejected - requires self-hosting, more complex
- **GitLab CI**: Rejected - not using GitLab
- **Single-arch builds**: Rejected - misses ARM market (AWS Graviton, Apple Silicon)

---

## 4. SQLite Performance at Scale

### Decision: Composite Indexes + Cursor Pagination

**Rationale**:
- **100k+ records requirement**: NFR-001 demands <1s queries for 100k records
- **Composite indexes**: Optimize for common query patterns (filter + sort)
- **Cursor pagination**: Better than OFFSET for large datasets (O(1) vs O(n))
- **Write-Ahead Logging**: Enabled for concurrent reads during writes

**Schema Design**:
```sql
CREATE TABLE notification_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_id TEXT NOT NULL,
    provider_type TEXT NOT NULL,
    recipient TEXT NOT NULL,
    message TEXT NOT NULL,
    subject TEXT,
    metadata TEXT,  -- JSON string
    priority TEXT DEFAULT 'normal',
    status TEXT NOT NULL,
    error_message TEXT,
    attempts INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,  -- ISO8601
    delivered_at TEXT,
    
    -- Composite indexes for common queries
    INDEX idx_provider_created (provider_id, created_at DESC),
    INDEX idx_status_created (status, created_at DESC),
    INDEX idx_type_created (provider_type, created_at DESC),
    INDEX idx_created_id (created_at DESC, id DESC)  -- Cursor pagination
);

-- Enable WAL mode for concurrent reads
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;  -- Balance durability vs performance
```

**Query Patterns**:
```go
// Cursor-based pagination (efficient for large offsets)
func (r *Repository) GetNotificationHistory(filters Filters, pageSize int, cursor *int) ([]Notification, *int, error) {
    query := `
        SELECT id, provider_id, provider_type, recipient, message, status, created_at
        FROM notification_logs
        WHERE 1=1
    `
    args := []interface{}{}
    
    if filters.ProviderID != "" {
        query += " AND provider_id = ?"
        args = append(args, filters.ProviderID)
    }
    if filters.Status != "" {
        query += " AND status = ?"
        args = append(args, filters.Status)
    }
    if cursor != nil {
        query += " AND id < ?"
        args = append(args, *cursor)
    }
    
    query += " ORDER BY created_at DESC, id DESC LIMIT ?"
    args = append(args, pageSize+1)
    
    // Execute and return results with next cursor
    // ...
}
```

**Pagination Approach**:
- **Cursor-based**: Use `id < ?` for consistent results even with concurrent writes
- **Page size**: Default 50 (configurable), fetch pageSize+1 to detect "has more"
- **Sorting**: Always include `id` in sort for deterministic ordering

**Performance Guarantees**:
- Composite indexes: O(log n) for filtered queries
- Cursor pagination: O(log n + pageSize) vs OFFSET O(n)
- WAL mode: Concurrent reads don't block during writes
- Query cache: SQLite automatically caches hot queries

**Retention Strategy**:
```go
// Cleanup old logs (run daily via cron or at startup)
func (r *Repository) CleanupOldLogs(retentionDays int) error {
    _, err := r.db.Exec(`
        DELETE FROM notification_logs
        WHERE created_at < datetime('now', '-' || ? || ' days')
    `, retentionDays)
    return err
}
```

**Alternatives Considered**:
- **PostgreSQL**: Rejected - violates KISS (adds deployment complexity), overkill for scale
- **Offset pagination**: Rejected - slow for large offsets, inconsistent with concurrent writes
- **No indexes**: Rejected - table scans too slow for 100k records

---

## 5. Kubernetes Deployment Patterns

### Decision: StatefulSet + PVC with ReadWriteOnce

**Rationale**:
- **SQLite constraint**: Single-writer requirement necessitates single replica
- **StatefulSet**: Guarantees stable pod identity and storage attachment
- **PVC**: Persistent volume for database file, survives pod restarts
- **Health checks**: Separate readiness (DB healthy) and liveness (process alive) probes

**Manifest Structure**:

**deployment.yaml** (using StatefulSet for stability):
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: notimulti
  labels:
    app: notimulti
spec:
  serviceName: notimulti
  replicas: 1  # Single replica due to SQLite
  selector:
    matchLabels:
      app: notimulti
  template:
    metadata:
      labels:
        app: notimulti
    spec:
      securityContext:
        runAsUser: 1000
        runAsNonRoot: true
        fsGroup: 1000
      containers:
      - name: notimulti
        image: username/notimulti:latest
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: PORT
          value: "8080"
        - name: LOG_LEVEL
          value: "info"
        - name: CONFIG_DIR
          value: "/app/configs"
        - name: DB_PATH
          value: "/app/data/notifications.db"
        volumeMounts:
        - name: config
          mountPath: /app/configs
        - name: data
          mountPath: /app/data
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: config
        configMap:
          name: notimulti-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
```

**service.yaml**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: notimulti
spec:
  selector:
    app: notimulti
  ports:
  - port: 80
    targetPort: 8080
    name: http
  type: ClusterIP
```

**configmap.yaml**:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: notimulti-config
data:
  telegram-alerts.json: |
    {
      "id": "telegram-alerts",
      "type": "telegram",
      "enabled": true,
      "config": {
        "bot_token": "placeholder",
        "default_chat_id": "placeholder"
      }
    }
```

**Health Check Strategy**:
- `/health`: Liveness probe - checks if process is alive (simple HTTP 200)
- `/ready`: Readiness probe - checks if DB connection is healthy, configs loaded
- Separate endpoints allow pod restart (liveness) vs traffic routing (readiness)

**Backup Strategy**:
- **Volume snapshots**: Use K8s VolumeSnapshot for point-in-time backups
- **Periodic export**: Cron job to export SQLite to S3/GCS
- **Retention**: 7 daily, 4 weekly, 12 monthly backups

**Scaling Considerations**:
- **Current**: Single replica (SQLite limitation)
- **Future**: If scale requires, migrate to PostgreSQL or use distributed SQLite (LiteFS/rqlite)
- **Read replicas**: Not needed - notification history is not high-traffic

**Alternatives Considered**:
- **Deployment**: Rejected - no stable pod identity, PVC attachment issues on reschedule
- **Multiple replicas**: Rejected - SQLite doesn't support concurrent writes
- **External database**: Rejected - adds complexity, violates KISS for current scale

---

## Summary of Decisions

| Area | Decision | Key Benefit |
|------|----------|-------------|
| Docker Build | Alpine + Multi-stage + Go embed | <100MB image, simple deployment |
| Async Logging | Buffered channel + batch flush | Non-blocking writes (NFR-004) |
| CI/CD | GitHub Actions + Buildx | Multi-arch, fast builds, native integration |
| Database | SQLite + composite indexes + cursor pagination | <1s queries for 100k records |
| Kubernetes | StatefulSet + PVC + health checks | Stable storage, production-ready |

All decisions align with KISS principle, extend Phase 1 patterns, and meet non-functional requirements.
