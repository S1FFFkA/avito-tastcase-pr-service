

Write-Host "Очистка старых файлов покрытия..." -ForegroundColor Yellow


Remove-Item coverage.out -ErrorAction SilentlyContinue
Remove-Item coverage_report.out -ErrorAction SilentlyContinue
Remove-Item coverage.txt -ErrorAction SilentlyContinue
Remove-Item coverage.prof -ErrorAction SilentlyContinue
Remove-Item coverage_report -ErrorAction SilentlyContinue
Remove-Item coverage -Recurse -Force -ErrorAction SilentlyContinue

Write-Host "Создание директории для логов..." -ForegroundColor Yellow


New-Item -ItemType Directory -Force -Path logs | Out-Null

Write-Host "Запуск тестов..." -ForegroundColor Green


go test -v -coverprofile=coverage_report ./internal/... ./pkg/...


if ($LASTEXITCODE -eq 0) {
    Write-Host "`nВсе тесты прошли успешно!" -ForegroundColor Green
    Write-Host "Файл покрытия создан: coverage_report" -ForegroundColor Cyan
} else {
    Write-Host "`nНекоторые тесты не прошли. Код выхода: $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}

