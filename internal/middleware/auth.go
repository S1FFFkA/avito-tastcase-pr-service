package middleware

import (
	"encoding/json"
	"net/http"

	"AVITOSAMPISHU/internal/domain"
)

const (
	statusUnauthorized = 401
)

// AuthMiddleware проверяет наличие заголовка Authorization
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusUnauthorized)

			errorResp := domain.NewErrorResponse(domain.ErrorCodeInvalidRequest, "Authorization header is required")
			data, _ := json.MarshalIndent(errorResp, "", "  ")
			_, _ = w.Write(data)
			return
		}

		next.ServeHTTP(w, r)
	})
}
