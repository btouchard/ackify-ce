# ---- Build stage ----
FROM golang:alpine AS builder

# Install security updates and ca-certificates
RUN apk update && apk add --no-cache ca-certificates git && rm -rf /var/cache/apk/*

# Create non-root user for build
RUN adduser -D -g '' ackuser

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Set GOTOOLCHAIN to auto to allow Go toolchain updates
ENV GOTOOLCHAIN=auto

RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments for metadata
ARG VERSION="dev"
ARG COMMIT="unknown"
ARG BUILD_DATE="unknown"

# Build the application with optimizations and metadata
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${BUILD_DATE}" \
    -o ackify ./cmd/community

# Build the migrate binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o migrate ./cmd/migrate

# ---- Runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot

# Re-declare ARG for runtime stage
ARG VERSION="dev"

# Add metadata labels
LABEL maintainer="Benjamin TOUCHARD"
LABEL version="${VERSION}"
LABEL description="Ackify - Document signature validation platform"
LABEL org.opencontainers.image.source="https://github.com/btouchard/ackify-ce"
LABEL org.opencontainers.image.description="Professional solution for validating and tracking document reading"
LABEL org.opencontainers.image.licenses="SSPL"

# Copy certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set working directory and copy application files
WORKDIR /app
COPY --from=builder /app/ackify /app/ackify
COPY --from=builder /app/migrate /app/migrate
COPY --from=builder /app/migrations /app/migrations

# Copy templates for filesystem loading
COPY --from=builder /app/templates /app/templates

# Set default environment variable for templates directory
ENV ACKIFY_TEMPLATES_DIR=/app/templates

# Use non-root user (already set in distroless image)
# USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/app/ackify"]
