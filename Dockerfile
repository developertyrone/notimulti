# Stage 1: Frontend build
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend

# Copy package files and install dependencies
COPY frontend/package*.json ./
RUN npm ci

# Copy frontend source and build
COPY frontend/ ./
RUN npm run build

# Stage 2: Backend build with embedded frontend
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app

# Install build dependencies (CGO required for SQLite)
RUN apk add --no-cache gcc musl-dev

# Copy go mod files and download dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source
COPY backend/ ./

# Copy built frontend from previous stage (Vite outputs to ../backend/cmd/server/dist)
COPY --from=frontend-builder /app/backend/cmd/server/dist ./cmd/server/dist

# Build backend with optimizations (T072)
# -ldflags="-s -w" strips debug info and symbol table (reduces size ~30%)
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w" \
    -o server \
    ./cmd/server

# Stage 3: Runtime
FROM alpine:3.18

# Install runtime dependencies (T072)
# ca-certificates: Required for HTTPS connections (Telegram API, SMTP TLS)
# tzdata: Required for correct timestamp handling across timezones
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user (T073, FR-025)
RUN adduser -D -u 1000 notimulti

# Switch to non-root user
USER notimulti

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=backend-builder --chown=notimulti:notimulti /app/server .

# Create directories for runtime data with correct permissions
RUN mkdir -p /app/data /app/configs

# Expose application port
EXPOSE 8080

# Set default environment variables (can be overridden)
ENV PORT=8080 \
    LOG_LEVEL=info \
    CONFIG_DIR=/app/configs \
    DB_PATH=/app/data/notifications.db \
    LOG_RETENTION_DAYS=90

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Run server
CMD ["./server"]
