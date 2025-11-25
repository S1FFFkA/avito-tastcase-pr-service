package middleware

import "net/http"

func LoggingMiddleware(next http.Handler) http.Handler {
	return next
}
