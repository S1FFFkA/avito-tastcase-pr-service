package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL    = "http://localhost:8080"
	authToken  = "Bearer test-token"
	iterations = 100
	workers    = 5
)

var (
	successCount int64
	errorCount   int64
	mu           sync.Mutex
	teamsCreated []string
	prsCreated   []string
	prsMutex     sync.Mutex
	runID        string // Уникальный ID для каждого запуска
)

func main() {
	// Генерируем уникальный ID для этого запуска (timestamp в миллисекундах)
	runID = fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))

	fmt.Println("Starting load test...")
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Iterations: %d\n", iterations)
	fmt.Printf("Workers: %d\n", workers)
	fmt.Printf("Run ID: %s\n", runID)
	fmt.Println()

	// Создаем начальные команды
	createInitialTeams()

	// Запускаем воркеры
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			worker(workerID)
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Println()
	fmt.Println("Load test completed!")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Successful requests: %d\n", successCount)
	fmt.Printf("Failed requests: %d\n", errorCount)
	fmt.Printf("Requests per second: %.2f\n", float64(successCount+errorCount)/duration.Seconds())
	fmt.Println()
	fmt.Printf("Check metrics at: %s/metrics\n", baseURL)
}

func createInitialTeams() {
	fmt.Println("Creating initial teams...")
	for i := 1; i <= 5; i++ {
		// Добавляем runID для уникальности
		teamName := fmt.Sprintf("load-test-team-%s-%d", runID, i)
		members := make([]map[string]interface{}, 0, 7)

		memberCount := 5 + (i % 3)
		for j := 1; j <= memberCount; j++ {
			members = append(members, map[string]interface{}{
				"user_id":   fmt.Sprintf("user-%s-%d-%d", runID, i, j),
				"username":  fmt.Sprintf("User %d-%d", i, j),
				"is_active": true,
			})
		}

		body := map[string]interface{}{
			"team_name": teamName,
			"members":   members,
		}

		result := makeRequest("POST", "/team/add", body, false)
		if result != nil {
			mu.Lock()
			teamsCreated = append(teamsCreated, teamName)
			mu.Unlock()
			fmt.Printf("  Created team: %s\n", teamName)
		} else {
			// Команда уже существует - это нормально при повторном запуске
			mu.Lock()
			teamsCreated = append(teamsCreated, teamName)
			mu.Unlock()
			fmt.Printf("  Team %s already exists or error (will try to use it)\n", teamName)
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println()
}

func worker(workerID int) {
	requestsPerWorker := iterations / workers
	extraRequests := iterations % workers
	if workerID < extraRequests {
		requestsPerWorker++
	}

	for i := 0; i < requestsPerWorker; i++ {
		iter := workerID*requestsPerWorker + i + 1

		// 1. Создание PR
		mu.Lock()
		teamCount := len(teamsCreated)
		mu.Unlock()

		if teamCount > 0 {
			teamIndex := iter % teamCount

			// Извлекаем номер команды из имени (load-test-team-{runID}-{i})
			// Для упрощения используем индекс команды
			prID := fmt.Sprintf("pr-load-%s-%d-%d", runID, workerID, iter)
			prBody := map[string]interface{}{
				"pull_request_id":   prID,
				"pull_request_name": fmt.Sprintf("Load Test PR %d-%d", workerID, iter),
				"author_id":         fmt.Sprintf("user-%s-%d-1", runID, teamIndex+1),
			}

			if makeRequest("POST", "/pullRequest/create", prBody, false) != nil {
				prsMutex.Lock()
				prsCreated = append(prsCreated, prID)
				prsMutex.Unlock()
			}
		}

		// 2. Переназначение ревьювера
		prsMutex.Lock()
		prCount := len(prsCreated)
		prsMutex.Unlock()

		if prCount > 0 && iter%2 == 0 {
			prIndex := (iter / 2) % prCount
			prsMutex.Lock()
			prID := prsCreated[prIndex]
			prsMutex.Unlock()

			// Получаем реальных ревьюверов из PR (но для упрощения используем известного пользователя)
			// В реальности нужно сначала получить PR, но для нагрузки используем существующего пользователя
			mu.Lock()
			teamCount := len(teamsCreated)
			mu.Unlock()

			if teamCount > 0 {
				teamIndex := prIndex % teamCount
				reassignBody := map[string]interface{}{
					"pull_request_id": prID,
					"old_user_id":     fmt.Sprintf("user-%s-%d-2", runID, teamIndex+1),
				}

				makeRequest("POST", "/pullRequest/reassign", reassignBody, true)
			}
		}

		// 3. Слияние PR
		prsMutex.Lock()
		prCount = len(prsCreated)
		prsMutex.Unlock()

		if prCount > 0 && iter%3 == 0 {
			prIndex := (iter / 3) % prCount
			prsMutex.Lock()
			prID := prsCreated[prIndex]
			prsMutex.Unlock()

			mergeBody := map[string]interface{}{
				"pull_request_id": prID,
			}

			if makeRequest("POST", "/pullRequest/merge", mergeBody, true) != nil {
				prsMutex.Lock()
				// Удаляем из списка
				newPRs := make([]string, 0, len(prsCreated)-1)
				for _, p := range prsCreated {
					if p != prID {
						newPRs = append(newPRs, p)
					}
				}
				prsCreated = newPRs
				prsMutex.Unlock()
			}
		}

		// 4. Получение команды
		mu.Lock()
		teamCount = len(teamsCreated)
		mu.Unlock()

		if teamCount > 0 && iter%4 == 0 {
			teamIndex := (iter / 4) % teamCount
			mu.Lock()
			teamName := teamsCreated[teamIndex]
			mu.Unlock()

			_ = makeRequest("GET", fmt.Sprintf("/team/get?team_name=%s", teamName), nil, true)
		}

		// 5. Получение ревью пользователя
		if iter%5 == 0 {
			teamIndex := (iter % teamCount)
			userID := fmt.Sprintf("user-%s-%d-%d", runID, teamIndex+1, (iter%3)+1)
			makeRequest("GET", fmt.Sprintf("/users/getReview?user_id=%s", userID), nil, true)
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func makeRequest(method, path string, body interface{}, ignoreExpectedErrors bool) map[string]interface{} {
	url := baseURL + path
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			if !ignoreExpectedErrors {
				incrementError()
			}
			fmt.Printf("  [ERROR] Failed to marshal body: %v\n", err)
			return nil
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		if !ignoreExpectedErrors {
			incrementError()
		}
		fmt.Printf("  [ERROR] Failed to create request: %v\n", err)
		return nil
	}

	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		if !ignoreExpectedErrors {
			incrementError()
		}
		fmt.Printf("  [ERROR] Request failed: %v (URL: %s)\n", err, url)
		return nil
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		incrementSuccess()
		var result map[string]interface{}
		json.Unmarshal(bodyBytes, &result)
		return result
	}

	// Проверяем, является ли ошибка ожидаемой (TEAM_EXISTS, PR_EXISTS, PR_MERGED)
	var errorResp map[string]interface{}
	json.Unmarshal(bodyBytes, &errorResp)

	isExpectedError := false
	if errorResp != nil {
		if error, ok := errorResp["error"].(map[string]interface{}); ok {
			if code, ok := error["code"].(string); ok {
				// Эти ошибки ожидаемы и не считаются реальными ошибками
				if code == "TEAM_EXISTS" || code == "PR_EXISTS" || code == "PR_MERGED" || code == "NOT_ASSIGNED" {
					isExpectedError = true
				}
			}
		}
	}

	if !ignoreExpectedErrors && !isExpectedError {
		incrementError()
		fmt.Printf("  [ERROR] HTTP %d: %s - Response: %s\n", resp.StatusCode, path, string(bodyBytes))
	} else if !isExpectedError {
		// Логируем только неожиданные ошибки
		fmt.Printf("  [WARN] HTTP %d: %s - Response: %s\n", resp.StatusCode, path, string(bodyBytes))
	}

	return nil
}

func incrementSuccess() {
	mu.Lock()
	successCount++
	mu.Unlock()
}

func incrementError() {
	mu.Lock()
	errorCount++
	mu.Unlock()
}
