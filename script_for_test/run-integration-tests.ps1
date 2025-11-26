param(
    [string]$ComposeFile = "docker-compose.test.yml"
)

$testDbUser = if ($env:TEST_DB_USER) { $env:TEST_DB_USER } else { "test_user" }
$testDbPassword = if ($env:TEST_DB_PASSWORD) { $env:TEST_DB_PASSWORD } else { "test_password" }
$testDbName = if ($env:TEST_DB_NAME) { $env:TEST_DB_NAME } else { "avito_pr_test" }
$testDbPort = if ($env:TEST_DB_PORT) { $env:TEST_DB_PORT } else { "55432" }
$testDbHost = if ($env:TEST_DB_HOST) { $env:TEST_DB_HOST } else { "localhost" }

$env:TEST_DB_HOST = $testDbHost
$env:TEST_DB_PORT = $testDbPort
$env:TEST_DB_USER = $testDbUser
$env:TEST_DB_PASSWORD = $testDbPassword
$env:TEST_DB_NAME = $testDbName
$env:TEST_DB_SSLMODE = "disable"

Write-Host "Starting integration test dependencies..." -ForegroundColor Yellow
docker compose -f $ComposeFile up -d --build
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to start docker compose services." -ForegroundColor Red
    exit $LASTEXITCODE
}

try {
    Write-Host "Running integration tests..." -ForegroundColor Green
    go test -v -tags=integration ./integration_tests/...
    $exitCode = $LASTEXITCODE
}
finally {
    Write-Host "Stopping integration test dependencies..." -ForegroundColor Yellow
    docker compose -f $ComposeFile down -v | Out-Null
}

if ($exitCode -eq 0) {
    Write-Host "Integration tests completed successfully." -ForegroundColor Green
} else {
    Write-Host "Integration tests failed with exit code $exitCode." -ForegroundColor Red
}

exit $exitCode

