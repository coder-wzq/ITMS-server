# ITMS Multi-stage Docker Build
# Stage 1: Build (uses vendor/ for offline builds)
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

# Copy vendor and source
COPY . .

RUN go build -mod=vendor -ldflags="-s -w" -o /build/itms-server ./cmd/gateway

# Stage 2: Runtime
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata curl && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

COPY --from=builder /build/itms-server .
COPY deploy/configs/ /app/etc/

RUN mkdir -p /app/logs /app/data /app/records /app/scenarios /app/resources

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["./itms-server"]
