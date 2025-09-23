# ---------- Stage 1: Build ----------
FROM golang:1.24.6-alpine AS builder

WORKDIR /app

# Зависимости для сборки с CGO (SQLite требует C-библиотек)
RUN apk add --no-cache build-base sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем бинарь с CGO
ENV CGO_ENABLED=1
RUN go build -o weather_bot ./main.go

# ---------- Stage 2: Runtime ----------
FROM alpine:3.20

# Runtime зависимости для SQLite и SSL
RUN apk add --no-cache sqlite-libs ca-certificates tzdata \
    && update-ca-certificates

WORKDIR /app

# Копируем бинарь и папку с кодом
COPY --from=builder /app/weather_bot .
COPY --from=builder /app/bot ./bot

# Запуск
CMD ["./weather_bot"]
