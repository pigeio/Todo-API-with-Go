package middleware

import (
	"net/http"

	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(5, 10) // 5 requests/sec, burst 10

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !limiter.Allow() {
			http.Error(w, "Too Many Requests (Rate Limited)", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
