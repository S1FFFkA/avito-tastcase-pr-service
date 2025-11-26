.PHONY: help run build mod-tidy test test-unit test-coverage lint \
	migrate-up migrate-down docker-up docker-down \
	docker-test-up docker-test-down integration-tests integration-tests-docker

# Переменные
GO=go
DOCKER_COMPOSE=docker-compose
DOCKER_COMPOSE_TEST=docker-compose -f docker-compose.test.yml
MIGRATE=migrate

# Справка по командам
help:
	@echo "Доступные команды:"
	@echo ""
	@echo "Разработка:"
	@echo "  make run              - Запустить приложение локально"
	@echo "  make build            - Собрать приложение"
	@echo "  make mod-tidy         - Обновить зависимости (go mod tidy)"
	@echo ""
	@echo "Тестирование:"
	@echo "  make test             - Запустить все тесты с покрытием"
	@echo "  make test-unit        - Запустить только unit тесты"
	@echo "  make test-coverage    - Показать покрытие кода в браузере"
	@echo "  make integration-tests - Запустить интеграционные тесты (требует запущенную БД)"
	@echo "  make integration-tests-docker - Запустить интеграционные тесты с Docker"
	@echo ""
	@echo "Качество кода:"
	@echo "  make lint             - Запустить линтер (golangci-lint)"
	@echo ""
	@echo "Миграции:"
	@echo "  make migrate-up      - Применить миграции (требует DB_DSN)"
	@echo "  make migrate-down    - Откатить миграции (требует DB_DSN)"
	@echo ""
	@echo "Docker (основное окружение):"
	@echo "  make docker-up       - Запустить все сервисы с пересборкой"
	@echo "  make docker-down     - Остановить все сервисы и удалить volumes"
	@echo ""
	@echo "Docker (тестовое окружение):"
	@echo "  make docker-test-up  - Запустить тестовую БД"
	@echo "  make docker-test-down - Остановить тестовую БД"

# Запуск приложения
run:
	$(GO) run ./cmd/main.go

# Сборка приложения
build:
	@mkdir -p bin
	$(GO) build -o bin/app ./cmd/main.go

# Обновление зависимостей
mod-tidy:
	$(GO) mod tidy
	$(GO) mod verify

# Запуск всех тестов
test:
	$(GO) test -v -race -coverprofile=coverage.out ./...

# Запуск только unit тестов (исключая интеграционные)
test-unit:
	$(GO) test -v -race -coverprofile=coverage.out $$($(GO) list ./... | grep -v integration_tests)

# Запуск тестов с покрытием
test-coverage: test
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Отчет о покрытии сохранен в coverage.html"

# Линтинг
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint не установлен. Установите: https://golangci-lint.run/usage/install/"; \
	fi

# Применение миграций
migrate-up:
	$(MIGRATE) -path migrations -database "$$DB_DSN" up

# Откат миграций
migrate-down:
	$(MIGRATE) -path migrations -database "$$DB_DSN" down

# Запуск через docker-compose с пересборкой
docker-up:
	$(DOCKER_COMPOSE) up -d --build

# Остановка docker-compose
docker-down:
	$(DOCKER_COMPOSE) down -v

# Запуск тестовой БД
docker-test-up:
	$(DOCKER_COMPOSE_TEST) up -d --build

# Остановка тестовой БД
docker-test-down:
	$(DOCKER_COMPOSE_TEST) down -v

# Интеграционные тесты (требует запущенную БД)
integration-tests:
	$(GO) test -v -tags=integration ./integration_tests/...

# Интеграционные тесты с автоматическим управлением Docker
integration-tests-docker:
	@echo "Запуск тестовой БД..."
	@$(DOCKER_COMPOSE_TEST) up -d --build
	@echo "Ожидание готовности БД..."
	@sleep 5
	@echo "Запуск интеграционных тестов..."
	@$(GO) test -v -tags=integration ./integration_tests/...; \
	EXIT_CODE=$$?; \
	$(DOCKER_COMPOSE_TEST) down -v; \
	exit $$EXIT_CODE
