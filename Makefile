.PHONY: run build test lint migrate-up migrate-down docker-up docker-up-fast docker-down integration-tests clean logs init clean-all

# Переменные
GO=go
DOCKER_COMPOSE=docker-compose
MIGRATE=migrate

# Запуск приложения
run:
	$(GO) run ./cmd/main.go

# Сборка приложения
build:
	$(GO) build -o bin/app ./cmd/main.go

# Запуск тестов
test:
	$(GO) test -v -race -coverprofile=coverage.out ./...

# Запуск тестов с покрытием
test-coverage: test
	$(GO) tool cover -html=coverage.out -o coverage.html

# Линтинг
lint:
	golangci-lint run

# Применение миграций
migrate-up:
	$(MIGRATE) -path migrations -database "$$DB_DSN" up

# Откат миграций
migrate-down:
	$(MIGRATE) -path migrations -database "$$DB_DSN" down

# Запуск через docker-compose
docker-up:
	$(DOCKER_COMPOSE) up -d --build

# Остановка docker-compose
docker-down:
	$(DOCKER_COMPOSE) down -v

# Интеграционные тесты
integration-tests:
	$(GO) test -v -tags=integration ./integration_tests/...

# Просмотр логов
logs:
	$(DOCKER_COMPOSE) logs -f backend

# Очистка
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html


