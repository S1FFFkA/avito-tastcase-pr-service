# Скрипт для запуска всех тестов с автоматической очисткой старых файлов покрытия

Write-Host "Очистка старых файлов покрытия..." -ForegroundColor Yellow

# Удаляем старые файлы покрытия
Remove-Item coverage.out -ErrorAction SilentlyContinue
Remove-Item coverage_report.out -ErrorAction SilentlyContinue
Remove-Item coverage.txt -ErrorAction SilentlyContinue
Remove-Item coverage.prof -ErrorAction SilentlyContinue
Remove-Item coverage_report -ErrorAction SilentlyContinue
Remove-Item coverage -Recurse -Force -ErrorAction SilentlyContinue

Write-Host "Создание директории для логов..." -ForegroundColor Yellow

# Создаем директорию для логов
New-Item -ItemType Directory -Force -Path logs | Out-Null

Write-Host "Запуск тестов..." -ForegroundColor Green

# Запускаем тесты (используем имя файла без расширения, которое может интерпретироваться как пакет)
go test -v -coverprofile=coverage_report ./internal/... ./pkg/...

# Проверяем результат
if ($LASTEXITCODE -eq 0) {
    Write-Host "`nВсе тесты прошли успешно!" -ForegroundColor Green
    Write-Host "Файл покрытия создан: coverage_report" -ForegroundColor Cyan
} else {
    Write-Host "`nНекоторые тесты не прошли. Код выхода: $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}

