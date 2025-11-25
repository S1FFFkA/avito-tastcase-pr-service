param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$AuthToken = "Bearer test-token",
    [int]$Iterations = 50
)

$ErrorActionPreference = "Continue"

Write-Host "Starting load test..." -ForegroundColor Green
Write-Host "Base URL: $BaseUrl" -ForegroundColor Yellow
Write-Host "Iterations: $Iterations" -ForegroundColor Yellow
Write-Host ""

$headers = @{
    "Authorization" = $AuthToken
    "Content-Type" = "application/json"
}

$successCount = 0
$errorCount = 0
$teamsCreated = @()
$prsCreated = @()

# Функция для отправки запроса
function Invoke-APIRequest {
    param(
        [string]$Method,
        [string]$Url,
        [object]$Body = $null
    )
    
    try {
        $params = @{
            Method = $Method
            Uri = $Url
            Headers = $headers
            TimeoutSec = 30
        }
        
        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json -Compress)
        }
        
        $response = Invoke-RestMethod @params
        $script:successCount++
        return $response
    }
    catch {
        $script:errorCount++
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# Создаем несколько команд для начала
Write-Host "Creating initial teams..." -ForegroundColor Cyan
for ($i = 1; $i -le 5; $i++) {
    $teamName = "load-test-team-$i"
    $members = @()
    
    # Создаем команду с 5-7 участниками
    $memberCount = 5 + ($i % 3)
    for ($j = 1; $j -le $memberCount; $j++) {
        $members += @{
            user_id = "user-$i-$j"
            username = "User $i-$j"
            is_active = $true
        }
    }
    
    $body = @{
        team_name = $teamName
        members = $members
    }
    
    $result = Invoke-APIRequest -Method "POST" -Url "$BaseUrl/team/add" -Body $body
    if ($result) {
        $teamsCreated += $teamName
        Write-Host "  Created team: $teamName" -ForegroundColor Green
    }
    Start-Sleep -Milliseconds 100
}

Write-Host ""
Write-Host "Starting load test iterations..." -ForegroundColor Cyan

# Основной цикл нагрузки
for ($iter = 1; $iter -le $Iterations; $iter++) {
    Write-Host "Iteration $iter/$Iterations..." -ForegroundColor Yellow
    
    # 1. Создание PR (генерирует метрики)
    if ($teamsCreated.Count -gt 0) {
        $teamIndex = $iter % $teamsCreated.Count
        $teamName = $teamsCreated[$teamIndex]
        $prId = "pr-load-$iter"
        
        $prBody = @{
            pull_request_id = $prId
            pull_request_name = "Load Test PR $iter"
            author_id = "user-$($teamIndex + 1)-1"
        }
        
        $pr = Invoke-APIRequest -Method "POST" -Url "$BaseUrl/pullRequest/create" -Body $prBody
        if ($pr) {
            $prsCreated += $prId
            Write-Host "  Created PR: $prId" -ForegroundColor Green
        }
    }
    
    # 2. Переназначение ревьювера (генерирует метрики)
    if ($prsCreated.Count -gt 0 -and $iter % 2 -eq 0) {
        $prIndex = ($iter / 2) % $prsCreated.Count
        $prId = $prsCreated[$prIndex]
        
        $reassignBody = @{
            pull_request_id = $prId
            old_user_id = "user-$($prIndex % 5 + 1)-2"
        }
        
        $result = Invoke-APIRequest -Method "POST" -Url "$BaseUrl/pullRequest/reassign" -Body $reassignBody
        if ($result) {
            Write-Host "  Reassigned reviewer for PR: $prId" -ForegroundColor Green
        }
    }
    
    # 3. Слияние PR (генерирует метрики)
    if ($prsCreated.Count -gt 0 -and $iter % 3 -eq 0) {
        $prIndex = ($iter / 3) % $prsCreated.Count
        $prId = $prsCreated[$prIndex]
        
        $mergeBody = @{
            pull_request_id = $prId
        }
        
        $result = Invoke-APIRequest -Method "POST" -Url "$BaseUrl/pullRequest/merge" -Body $mergeBody
        if ($result) {
            Write-Host "  Merged PR: $prId" -ForegroundColor Green
            # Удаляем из списка, чтобы не пытаться мержить снова
            $prsCreated = $prsCreated | Where-Object { $_ -ne $prId }
        }
    }
    
    # 4. Получение команды
    if ($teamsCreated.Count -gt 0 -and $iter % 4 -eq 0) {
        $teamIndex = ($iter / 4) % $teamsCreated.Count
        $teamName = $teamsCreated[$teamIndex]
        
        $result = Invoke-APIRequest -Method "GET" -Url "$BaseUrl/team/get?team_name=$teamName"
        if ($result) {
            Write-Host "  Retrieved team: $teamName" -ForegroundColor Green
        }
    }
    
    # 5. Получение ревью пользователя
    if ($iter % 5 -eq 0) {
        $userId = "user-$($iter % 5 + 1)-$($iter % 3 + 1)"
        $result = Invoke-APIRequest -Method "GET" -Url "$BaseUrl/users/getReview?user_id=$userId"
        if ($result) {
            Write-Host "  Retrieved reviews for user: $userId" -ForegroundColor Green
        }
    }
    
    # Небольшая задержка между итерациями
    Start-Sleep -Milliseconds 200
}

Write-Host ""
Write-Host "Load test completed!" -ForegroundColor Green
Write-Host "Successful requests: $successCount" -ForegroundColor Green
Write-Host "Failed requests: $errorCount" -ForegroundColor $(if ($errorCount -gt 0) { "Red" } else { "Green" })
Write-Host ""
Write-Host "Check metrics at: $BaseUrl/metrics" -ForegroundColor Cyan

