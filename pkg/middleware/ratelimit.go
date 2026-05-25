package middleware

import (
	"net/http"
	"sync"
	"time"

	"itms-server/pkg/response"
)

type tokenBucket struct {
	tokens   float64
	lastTime time.Time
	rate     float64
	burst    float64
	mu       sync.Mutex
}

func newTokenBucket(rate, burst int) *tokenBucket {
	return &tokenBucket{
		tokens:   float64(burst),
		lastTime: time.Now(),
		rate:     float64(rate),
		burst:    float64(burst),
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastTime).Seconds()
	b.tokens += elapsed * b.rate
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
	b.lastTime = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

func RateLimit(rate, burst int) func(http.HandlerFunc) http.HandlerFunc {
	buckets := sync.Map{}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}

			bucket, _ := buckets.LoadOrStore(ip, newTokenBucket(rate, burst))
			tb := bucket.(*tokenBucket)

			if !tb.allow() {
				response.WriteJSON(w, http.StatusTooManyRequests,
					response.Error(response.CodeTooManyRequests, "too many requests, please try again later"))
				return
			}

			next(w, r)
		}
	}
}
