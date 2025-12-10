# Solution: Dockerfile with PostgreSQL Client Tools for Module 8
# This Dockerfile extends the production-ready Dockerfile from Module 7
# to include PostgreSQL client tools (pg_dump, psql) needed for backup/restore functionality
#
# Use this Dockerfile if your operator performs backups/restores using pg_dump/psql

# Build stage
FROM golang:1.24 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Cache deps before building and copying source
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/ internal/

# Build the manager binary
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -a -o manager cmd/main.go

# Runtime stage - use distroless base but add PostgreSQL client tools
# Note: distroless doesn't have package manager, so we use a minimal Debian base
FROM debian:bookworm-slim

# Install PostgreSQL client tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    postgresql-client \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user and group
RUN groupadd -r -g 65532 nonroot && \
    useradd -r -u 65532 -g nonroot -m -s /bin/bash nonroot

WORKDIR /
COPY --from=builder /workspace/manager .

# Switch to non-root user
USER 65532:65532

ENTRYPOINT ["/manager"]

