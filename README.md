# Avito PR API

Сервис для автоматического назначения ревьюеров на Pull Request'ы.

## Запуск

### Настройка переменных окружения

**Важно:** Перед запуском необходимо создать файл `.env` в корневой директории проекта.

1. Скопируйте пример конфигурации:
   ```powershell
   # Windows PowerShell
   Copy-Item .env.example .env
   ```
   ```bash
   # Linux/macOS
   cp .env.example .env
   ```

2. Отредактируйте `.env` файл при необходимости (значения по умолчанию уже настроены):
   ```env
   # API Configuration
   API_PORT=8080

   # Database Configuration
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=avito_user
   DB_PASSWORD=avito_password
   DB_NAME=avito_pr_db
   DB_SSLMODE=disable

   # For Docker Compose, use:
   # DB_HOST=postgres
   ```

### Локальный запуск приложения (через docker-compose)

**Клонирование репозитория**

```bash
git clone <repository-url>
cd AVITOSAMPISHU
```

**Запуск приложения**

```bash
docker-compose up -d --build
```

> **Примечание:** Если у вас установлен `make`, можно использовать `make docker-up` (работает на Linux/macOS)

После выполнения этой команды у вас соберутся все необходимые образы и поднимутся следующие контейнеры:

| Сервис       | Имя контейнера          | Доступный порт на хосте |
| ------------ | ----------------------- | ----------------------- |
| Основной API | avito-pr-api-backend    | localhost:8080          |
| PostgreSQL   | avito-pr-api-postgres   | localhost:5432          |
| Migrate      | avito-pr-api-migrate    | (падает после выполнения миграций) |
| Prometheus   | avito-pr-api-prometheus | localhost:9090          |

Миграции применяются автоматически при запуске через контейнер migrate.

**Остановка всех контейнеров с удалением volumes**

```bash
docker-compose down -v
```

### Локальный запуск приложения

Перед запуском необходимо установить переменные окружения:
- `DB_HOST` - хост БД (по умолчанию `localhost`)
- `DB_PORT` - порт БД (по умолчанию `5432`)
- `DB_USER` - пользователь БД (по умолчанию `avito_user`)
- `DB_PASSWORD` - пароль БД (по умолчанию `avito_password`)
- `DB_NAME` - имя БД (по умолчанию `avito_pr_db`)
- `DB_SSLMODE` - режим SSL (по умолчанию `disable`)
- `API_PORT` - порт для API сервера (по умолчанию `8080`)

**Применение миграций**

```bash
make migrate-up
```

**Запуск приложения**

```bash
make run
```

## Структура проекта

```
AVITOSAMPISHU/
├── cmd/                    # Точка входа приложения
├── internal/
│   ├── app/               # Инициализация приложения
│   ├── domain/            # Доменные модели и ошибки
│   ├── handlers/          # HTTP handlers
│   ├── infrastructure/    # Подключение к БД и другим системам
│   │   └── database/      # Подключение к PostgreSQL
│   ├── middleware/        # HTTP middleware
│   ├── repository/        # Слой доступа к данным
│   ├── server/            # HTTP сервер
│   └── service/           # Бизнес-логика
├── pkg/
│   ├── helpers/           # Вспомогательные функции
│   ├── logger/            # Логирование
│   └── metrics/           # Prometheus метрики
├── migrations/            # SQL миграции
├── integration_tests/     # Интеграционные тесты
└── prometheus/            # Конфигурация Prometheus
```

## Быстрый старт

**Windows:**
```powershell
.\docker-up.ps1
```

**Linux/macOS:**
```bash
make docker-up
```

После запуска сервисы будут доступны:
- API: http://localhost:8080
- Prometheus: http://localhost:9090

## Требования

- Docker и Docker Compose (для запуска через docker-compose)
- Go 1.25+ (для локальной разработки)
- PostgreSQL 15+ (уже включен в docker-compose)

## Установка и запуск

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd AVITOSAMPISHU
```

### 2. Настройка окружения

Создайте файл `.env` на основе `.env.example`:

```bash
cp .env.example .env
```

Отредактируйте `.env` и укажите параметры подключения к БД:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=avito_user
DB_PASSWORD=avito_password
DB_NAME=avito_pr_db
DB_SSLMODE=disable

# API Configuration
API_PORT=8080
```

### 3. Установка зависимостей

```bash
go mod download
```

### 4. Применение миграций

```bash
make migrate-up
```

### 5. Запуск приложения

```bash
make run
```

Приложение будет доступно по адресу `http://localhost:8080`


## Тестирование

### Unit тесты

```bash
make test
```

### Тесты с покрытием

```bash
make test-coverage
```

### Интеграционные тесты

Перед запуском убедитесь, что порт 5432 свободен:

```bash
make integration-tests
```

## Линтинг

```bash
make lint
```

## API Endpoints

### Команды

- `POST /team/add` - Создание команды
- `GET /team/get?team_name=<name>` - Получение команды

### Пользователи

- `POST /users/setIsActive` - Изменение статуса активности пользователя
- `GET /users/getReview?user_id=<id>` - Получение PR пользователя
- `POST /users/deactivateTeamMembers` - Деактивация пользователей команды

### Pull Requests

- `POST /pullRequest/create` - Создание PR
- `POST /pullRequest/merge` - Мердж PR
- `POST /pullRequest/reassign` - Переназначение ревьюера

### Метрики

- `GET /metrics` - Prometheus метрики

## Авторизация

В проекте реализована простая заглушка авторизации для демонстрации. Для **всех эндпоинтов** требуется заголовок `Authorization`.

### Настройка авторизации в Postman

#### Способ 1: Глобальная авторизация для всех запросов (рекомендуется)

1. Откройте Postman
2. Нажмите на иконку **"Environments"** (или `Ctrl+E`)
3. Создайте новое окружение или выберите существующее
4. Добавьте переменную:
   - **Variable**: `auth_token`
   - **Initial Value**: `Bearer test-token` (или любой другой токен)
   - **Current Value**: `Bearer test-token`
5. Сохраните окружение
6. Выберите это окружение в выпадающем списке в правом верхнем углу

7. Настройте авторизацию для коллекции:
   - Откройте вашу коллекцию
   - Перейдите на вкладку **"Authorization"**
   - Выберите тип: **"Bearer Token"**
   - В поле **Token** введите: `{{auth_token}}` (или просто `Bearer test-token`)
   - Нажмите **"Save"**

Теперь все запросы в коллекции будут автоматически использовать этот токен.

#### Способ 2: Авторизация для отдельного запроса

1. Откройте нужный запрос в Postman
2. Перейдите на вкладку **"Authorization"**
3. Выберите тип: **"Bearer Token"**
4. В поле **Token** введите: `test-token` (или любой другой токен)
5. Postman автоматически добавит заголовок `Authorization: Bearer test-token`

#### Способ 3: Ручное добавление заголовка

1. Откройте нужный запрос
2. Перейдите на вкладку **"Headers"**
3. Добавьте новый заголовок:
   - **Key**: `Authorization`
   - **Value**: `Bearer test-token` (или просто `test-token`)

### Важно

- Токен может быть **любым** - проверка только на наличие заголовка
- Это заглушка для демонстрации - в продакшене нужно реализовать реальную проверку токенов
- Рекомендуется использовать формат `Bearer <token>`, но можно и просто `<token>`

### Пример запроса с curl

```bash
curl -X POST http://localhost:8080/team/add \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{"team_name": "backend", "members": [{"user_id": "user1", "username": "User 1", "is_active": true}]}'
```

## Логирование

Приложение использует структурированное логирование через `zap` (SugaredLogger). Логи записываются в:
- `logs/log.txt` - общие логи
- `logs/error.txt` - ошибки
- `stdout` - консольный вывод

Логи включают:
- Успешные операции (создание, получение, обновление)
- Ошибки с контекстом (ID запроса, параметры)
- Старт и остановка сервера
- SQL запросы и транзакции (только ошибки)
- Бизнес-транзакции и правила

## Метрики

### reviewer_load_distribution

Отслеживает распределение нагрузки между ревьюверами (количество открытых PR на ревьювера).

Примеры PromQL запросов:
```promql
# Средняя нагрузка на ревьювера
rate(reviewer_load_distribution_sum[5m]) / rate(reviewer_load_distribution_count[5m])

# 95-й перцентиль нагрузки
histogram_quantile(0.95, rate(reviewer_load_distribution_bucket[5m]))
```

## Тестирование

### Запуск всех тестов

**Windows (PowerShell/CMD):**
### Простой способ (рекомендуется)

Используйте готовый скрипт для автоматической очистки и запуска тестов:

```powershell
.\run-tests.ps1
```

Скрипт автоматически:
- Удаляет старые файлы покрытия
- Создает директорию для логов
- Запускает все тесты
- Показывает результат

### Ручной запуск

Если нужно запустить тесты вручную:

```powershell
# Сначала обновите зависимости
go mod tidy

# Удалите старые файлы покрытия и директорию (если есть)
Remove-Item coverage.out -ErrorAction SilentlyContinue
Remove-Item coverage_report.out -ErrorAction SilentlyContinue
Remove-Item coverage.txt -ErrorAction SilentlyContinue
Remove-Item coverage.prof -ErrorAction SilentlyContinue
Remove-Item coverage_report -ErrorAction SilentlyContinue
Remove-Item coverage -Recurse -Force -ErrorAction SilentlyContinue

# Создайте директорию для логов (если её нет)
New-Item -ItemType Directory -Force -Path logs | Out-Null

# Затем запустите тесты
go test -v -coverprofile=coverage_report ./internal/... ./pkg/...
```

> **Примечание:** Флаг `-race` (race detector) требует cgo и может не работать на Windows без установленного компилятора C. Для проверки гонок данных используйте Linux/macOS.

> **Примечание для Windows:** Если видите ошибку `no required module provides package .out`, это значит, что Go пытается интерпретировать файл покрытия как пакет. Убедитесь, что вы удалили старый файл `coverage.out` перед запуском тестов.
> ```powershell
> go test -v -coverprofile=./coverage.out ./internal/... ./pkg/...
> ```

> **Примечание:** Флаг `-race` (race detector) требует cgo и может не работать на Windows без установленного компилятора C. Для проверки гонок данных используйте Linux/macOS.

**Linux/macOS:**
```bash
make test
# или с race detector
go test -v -race -coverprofile=coverage.out ./...
```

### Запуск тестов с покрытием

**Windows:**
```powershell
# Сначала обновите зависимости
go mod tidy

# Удалите старый файл покрытия (если есть)
Remove-Item coverage.out -ErrorAction SilentlyContinue

# Затем запустите тесты с покрытием
go test -v -coverprofile=coverage.out ./internal/... ./pkg/...
go tool cover -html=coverage.out -o coverage.html
```

**Linux/macOS:**
```bash
make test-coverage
# или с race detector
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

После выполнения откройте `coverage.html` в браузере для просмотра покрытия кода.

### Запуск интеграционных тестов

Интеграционные тесты требуют запущенную базу данных PostgreSQL. Убедитесь, что Docker Compose запущен:

**Windows:**
```powershell
# Убедитесь, что БД запущена
docker-compose ps

# Установите переменную окружения для подключения к БД
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="avito_user"
$env:DB_PASSWORD="avito_password"
$env:DB_NAME="avito_pr_db"
$env:DB_SSLMODE="disable"

# Запустите интеграционные тесты
go test -v -tags=integration ./integration_tests/...
```

**Linux/macOS:**
```bash
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="avito_user"
export DB_PASSWORD="avito_password"
export DB_NAME="avito_pr_db"
export DB_SSLMODE="disable"
make integration-tests
```

### Типы тестов

1. **Unit тесты сервисов** (`internal/service/*/`) - тестируют бизнес-логику с моками репозиториев
2. **Интеграционные тесты репозиториев** (`internal/repository/*/`) - тестируют работу с реальной БД
3. **Unit тесты хелперов** (`pkg/helpers/`) - тестируют вспомогательные функции
4. **Интеграционные тесты** (`integration_tests/`) - тестируют полный флоу работы приложения

## Разработка

### Добавление новой миграции

1. Создайте файлы миграций в `migrations/`:
   - `XXXX_description_up.sql` - применение миграции
   - `XXXX_description_down.sql` - откат миграции

2. Примените миграцию:
```bash
make migrate-up
```

### Структура кода

- **Domain** - доменные модели, ошибки, константы, response структуры
- **Repository** - работа с БД, возвращает доменные ошибки или технические ошибки
- **Service** - бизнес-логика, использует репозитории
- **Handlers** - HTTP слой, валидация, обработка ошибок, логирование

### Логирование

Логгер инициализируется через `logger.InitLogger()` и использует `zap.SugaredLogger` для удобного логирования с ключ-значение парами.

Конфигурация логгера:
- Формат: JSON
- Уровень: Info (можно изменить через переменную окружения)
- Выходы: `logs/log.txt`, `stdout` (общие логи), `logs/error.txt`, `stderr` (ошибки)
- Включает PID процесса в каждую запись

## Лицензия

MIT
