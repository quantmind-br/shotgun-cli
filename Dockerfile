# Multi-stage build for Go application
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

# Set build arguments
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64

# Build the application
RUN CGO_ENABLED=$CGO_ENABLED GOOS=$GOOS GOARCH=$GOARCH go build \
    -ldflags="-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
    -a -installsuffix cgo \
    -o shotgun-cli \
    .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/shotgun-cli .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Add entrypoint
ENTRYPOINT ["./shotgun-cli"]

# Default command
CMD ["--help"]

# Labels
LABEL maintainer="quantmind-br <maintainer@example.com>"
LABEL version="${VERSION}"
LABEL description="Shotgun CLI - LLM-optimized codebase context generator"