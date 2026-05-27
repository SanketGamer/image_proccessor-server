# ── Stage 1: Build ──────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go

# ── Stage 2: Run ────────────────────────────────────────────────
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

ENV TZ=Asia/Kolkata

WORKDIR /app

COPY --from=builder /app/server .
COPY .env .

EXPOSE 8080

CMD ["./server"]