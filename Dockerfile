FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Обновляем go.sum для всех зависимостей
RUN go mod tidy

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/app ./cmd/main.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Создаем директорию для логов
RUN mkdir -p /app/logs

# Копируем бинарник из builder
COPY --from=builder /app/bin/app .

EXPOSE 8080

CMD ["./app"]

