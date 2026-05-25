package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type OperationLog struct {
	OperatorID   int64     `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	Operation    string    `json:"operation"`
	Resource     string    `json:"resource"`
	RequestParam string    `json:"request_param"`
	ResponseCode int       `json:"response_code"`
	IP           string    `json:"ip"`
	Duration     int64     `json:"duration"`
	CreatedAt    time.Time `json:"created_at"`
}

type logWriter struct {
	http.ResponseWriter
	statusCode int
	buf        *bytes.Buffer
}

func (lw *logWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func (lw *logWriter) Write(b []byte) (int, error) {
	lw.buf.Write(b)
	return lw.ResponseWriter.Write(b)
}

var logChannel = make(chan *OperationLog, 1000)

func init() {
	go func() {
		for entry := range logChannel {
			data, _ := json.Marshal(entry)
			log.Printf("[AUDIT] %s", string(data))
		}
	}()
}

var sensitiveFields = []string{"password", "token", "secret", "refreshToken", "accessToken", "authorization"}

func sanitizeParams(params map[string]interface{}) map[string]interface{} {
	for key := range params {
		lower := strings.ToLower(key)
		for _, sf := range sensitiveFields {
			if strings.Contains(lower, sf) {
				params[key] = "***"
				break
			}
		}
	}
	return params
}

func OperationLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
			next(w, r)
			return
		}

		start := time.Now()

		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		lw := &logWriter{ResponseWriter: w, statusCode: http.StatusOK, buf: &bytes.Buffer{}}
		next(lw, r)

		entry := &OperationLog{
			Operation:    r.Method,
			Resource:     r.URL.Path,
			ResponseCode: lw.statusCode,
			IP:           clientIP(r),
			Duration:     time.Since(start).Milliseconds(),
			CreatedAt:    time.Now(),
		}

		if claims := GetClaims(r.Context()); claims != nil {
			entry.OperatorID = claims.UserID
			entry.OperatorName = claims.Username
		}

		if len(bodyBytes) > 0 && len(bodyBytes) < 8192 {
			var params map[string]interface{}
			if json.Unmarshal(bodyBytes, &params) == nil {
				entry.RequestParam = toJSON(sanitizeParams(params))
			} else {
				entry.RequestParam = sanitizeString(string(bodyBytes))
			}
		}

		select {
		case logChannel <- entry:
		default:
			log.Println("[AUDIT] log channel full, dropping entry")
		}
	}
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func sanitizeString(s string) string {
	lower := strings.ToLower(s)
	for _, sf := range sensitiveFields {
		if strings.Contains(lower, sf) {
			return "***"
		}
	}
	if len(s) > 512 {
		return s[:512]
	}
	return s
}
