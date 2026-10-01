# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source tree
COPY . .

# Compile static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./main.go

# Production runtime stage
FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata && \
    mkdir -p /data && \
    addgroup -S app && adduser -S app -G app && \
    chown -R app:app /app /data

COPY --from=builder /app/server /app/server

ENV PORT=8080
ENV DATABASE_PATH=/data/terrafertil.db
ENV AI_SERVICE_URL=http://ai-module:8001

USER app

EXPOSE 8080

CMD ["/app/server"]
