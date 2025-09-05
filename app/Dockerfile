FROM golang:1.24.6-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости для сборки с CGO
RUN apk add --no-cache build-base sqlite-dev

# Копируем зависимости Go
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Сборка бинарника с включенным CGO
ENV CGO_ENABLED=1
RUN GOOS=linux GOARCH=amd64 go build -o weather_bot ./main.go

FROM alpine:3.20

# Runtime зависимости для sqlite
RUN apk add --no-cache sqlite-libs ca-certificates tzdata

WORKDIR /app

# Копируем бинарь и даём права
COPY --from=builder /app/weather_bot .
RUN chmod +x /app/weather_bot

# Копируем нужные папки с кодом
COPY --from=builder /app/bot ./bot

# Создаём директории для volumes
RUN mkdir -p /app/bot/storage/bd \
    /app/bot/logger/logs \
    /app/bot/storage/cache/now \
    /app/bot/storage/cache/today \
    /app/bot/storage/cache/tomorrow

# Старт приложения
CMD ["./weather_bot"]
