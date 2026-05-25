package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"itms-server/pkg/response"
)

func Recovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %s %s: %v\n%s", r.Method, r.URL.Path, err, debug.Stack())
				response.WriteJSON(w, http.StatusInternalServerError,
					response.Error(response.CodeAuthServerErr, "internal server error"))
			}
		}()
		next(w, r)
	}
}
