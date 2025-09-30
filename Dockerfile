# ---- Build ----
FROM golang:alpine AS builder

RUN apk update && apk add --no-cache ca-certificates git curl && rm -rf /var/cache/apk/*
RUN adduser -D -g '' ackuser

WORKDIR /app
COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download && go mod verify
COPY . .

# Download Tailwind CSS CLI (use v3 for compatibility)
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then TAILWIND_ARCH="x64"; \
    elif [ "$ARCH" = "aarch64" ]; then TAILWIND_ARCH="arm64"; \
    else echo "Unsupported architecture: $ARCH" && exit 1; fi && \
    curl -sL https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.16/tailwindcss-linux-${TAILWIND_ARCH} -o /tmp/tailwindcss && \
    chmod +x /tmp/tailwindcss

# Build CSS
RUN mkdir -p ./static && \
    /tmp/tailwindcss -i ./assets/input.css -o ./static/output.css --minify

ARG VERSION="dev"
ARG COMMIT="unknown"
ARG BUILD_DATE="unknown"

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${BUILD_DATE}" \
    -o ackify ./cmd/community

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o migrate ./cmd/migrate

# ---- Run ----
FROM gcr.io/distroless/static-debian12:nonroot

ARG VERSION="dev"

LABEL maintainer="Benjamin TOUCHARD"
LABEL version="${VERSION}"
LABEL description="Ackify - Document signature validation platform"
LABEL org.opencontainers.image.source="https://github.com/btouchard/ackify-ce"
LABEL org.opencontainers.image.description="Professional solution for validating and tracking document reading"
LABEL org.opencontainers.image.licenses="AGPL-3.0-or-later"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app
COPY --from=builder /app/ackify /app/ackify
COPY --from=builder /app/migrate /app/migrate
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/locales /app/locales
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static

ENV ACKIFY_TEMPLATES_DIR=/app/templates
ENV ACKIFY_LOCALES_DIR=/app/locales
ENV ACKIFY_STATIC_DIR=/app/static

EXPOSE 8080

ENTRYPOINT ["/app/ackify"]
## SPDX-License-Identifier: AGPL-3.0-or-later
