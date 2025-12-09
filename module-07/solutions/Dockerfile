# Solution: Production-Ready Dockerfile from Module 7
# This matches the kubebuilder-generated Dockerfile structure
# Multi-stage build for smaller, secure images

# Build stage
FROM golang:1.21 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Cache deps before building and copying source
# This allows us to cache dependencies
RUN go mod download

# Copy the go source (kubebuilder structure)
# Note: Copy entire internal/ directory to include both controllers AND webhooks
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/ internal/

# Build the manager binary
# Note: kubebuilder puts main.go in cmd/ directory
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -a -o manager cmd/main.go

# Runtime stage - use distroless for security
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]

