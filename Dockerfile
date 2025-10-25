FROM node:22-alpine AS spa-builder

WORKDIR /app/webapp
COPY webapp/package*.json ./
RUN npm ci
COPY webapp/ ./
RUN npm run build

FROM golang:alpine AS builder

RUN apk update && apk add --no-cache ca-certificates git curl && rm -rf /var/cache/apk/*
RUN adduser -D -g '' ackuser

WORKDIR /app
COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download && go mod verify
COPY backend/ ./backend/

RUN mkdir -p backend/cmd/community/web/dist
COPY --from=spa-builder /app/webapp/dist ./backend/cmd/community/web/dist

ARG VERSION="dev"
ARG COMMIT="unknown"
ARG BUILD_DATE="unknown"

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${BUILD_DATE}" \
    -o ackify ./backend/cmd/community

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o migrate ./backend/cmd/migrate

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
COPY --from=builder /app/backend/migrations /app/migrations
COPY --from=builder /app/backend/locales /app/locales
COPY --from=builder /app/backend/templates /app/templates
COPY --from=builder /app/backend/openapi.yaml /app/openapi.yaml

ENV ACKIFY_TEMPLATES_DIR=/app/templates
ENV ACKIFY_LOCALES_DIR=/app/locales

EXPOSE 8080

ENTRYPOINT ["/app/ackify"]
