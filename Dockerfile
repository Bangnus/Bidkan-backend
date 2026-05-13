# ===== Stage 1: Build =====
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY docs/ ./docs/

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/main.go

# ===== Stage 2: Run =====
FROM alpine AS runner

RUN apk add --no-cache ca-certificates tzdata

# สร้าง user แยก ไม่ใช้ root (Security Best Practice)
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

WORKDIR /app

COPY --from=builder /app/server .

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./server"]
