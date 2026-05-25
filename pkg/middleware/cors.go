package middleware

import "net/http"

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

var defaultCORS = &CORSConfig{
	AllowedOrigins: []string{"*"},
	AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
	AllowedHeaders: []string{"Authorization", "Content-Type", "X-Request-ID", "X-Requested-With"},
	MaxAge:         3600,
}

var corsConfig *CORSConfig

func InitCORS(cfg *CORSConfig) {
	if cfg == nil {
		corsConfig = defaultCORS
	} else {
		corsConfig = cfg
	}
}

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowOrigin := "*"

		if len(corsConfig.AllowedOrigins) > 0 && corsConfig.AllowedOrigins[0] != "*" {
			for _, allowed := range corsConfig.AllowedOrigins {
				if allowed == origin {
					allowOrigin = origin
					break
				}
			}
		}

		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", joinStrings(corsConfig.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", joinStrings(corsConfig.AllowedHeaders, ", "))
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
