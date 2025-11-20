package middleware

import (
	"net/http"
	"sync/atomic"
	"time"
)

var throttleCounter int64
var throttleLimit int64 = 20

func ThrottleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		current := atomic.AddInt64(&throttleCounter, 1)

		if current > throttleLimit {
			time.Sleep(500 * time.Millisecond) // delay
		}

		next.ServeHTTP(w, r)
	})
}
