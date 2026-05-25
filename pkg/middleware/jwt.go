package middleware

import (
	"context"
	"net/http"
	"strings"

	"itms-server/pkg/jwt"
	"itms-server/pkg/response"
)

type contextKey string

const ClaimsKey contextKey = "claims"

var jwtConfig *jwt.Config

var excludePaths = map[string]bool{
	"/api/auth/login":   true,
	"/api/auth/refresh": true,
	"/health":           true,
}

func InitJWTMiddleware(cfg *jwt.Config) {
	jwtConfig = cfg
}

func JWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if excludePaths[r.URL.Path] {
			next(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.WriteJSON(w, http.StatusUnauthorized,
				response.Error(response.CodeAuthTokenMissing, response.GetMessage(response.CodeAuthTokenMissing)))
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			response.WriteJSON(w, http.StatusUnauthorized,
				response.Error(response.CodeAuthTokenMissing, response.GetMessage(response.CodeAuthTokenMissing)))
			return
		}

		claims, err := jwt.ParseToken(jwtConfig.Secret, tokenStr)
		if err != nil {
			response.WriteJSON(w, http.StatusUnauthorized,
				response.Error(response.CodeAuthTokenInvalid, response.GetMessage(response.CodeAuthTokenInvalid)))
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

func GetClaims(ctx context.Context) *jwt.Claims {
	if claims, ok := ctx.Value(ClaimsKey).(*jwt.Claims); ok {
		return claims
	}
	return nil
}
