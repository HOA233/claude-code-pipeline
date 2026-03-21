# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata git nodejs npm

# Install Claude CLI (for demo - in production use proper installation)
RUN npm install -g @anthropic-ai/claude-code || true

WORKDIR /app

COPY --from=builder /server .
COPY --from=builder /app/config ./config

ENV TZ=Asia/Shanghai

EXPOSE 8080

CMD ["./server"]