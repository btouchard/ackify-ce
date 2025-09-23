# ---- Build ----
FROM golang:alpine AS builder

RUN apk update && apk add --no-cache ca-certificates git && rm -rf /var/cache/apk/*
RUN adduser -D -g '' ackuser

WORKDIR /app
COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download && go mod verify
COPY . .

ARG VERSION="dev"
ARG COMMIT="unknown"
ARG BUILD_DATE="unknown"

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${BUILD_DATE}" \
    -o ackify ./cmd/community

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
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
LABEL org.opencontainers.image.licenses="SSPL"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app
COPY --from=builder /app/ackify /app/ackify
COPY --from=builder /app/migrate /app/migrate
COPY --from=builder /app/migrations /app/migrations

COPY --from=builder /app/templates /app/templates

ENV ACKIFY_TEMPLATES_DIR=/app/templates

EXPOSE 8080

ENTRYPOINT ["/app/ackify"]
