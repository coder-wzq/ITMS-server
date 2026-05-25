package response

import "net/http"

func StatusCode(code int) int {
	switch {
	case code == CodeSuccess:
		return http.StatusOK
	case code >= 1001 && code <= 1999:
		if code == CodeAuthTokenInvalid || code == CodeAuthTokenMissing {
			return http.StatusUnauthorized
		}
		if code == CodeAuthForbidden || code == CodeAuthNoPermission {
			return http.StatusForbidden
		}
		if code == CodeTooManyRequests {
			return http.StatusTooManyRequests
		}
		return http.StatusBadRequest
	case code >= 2000 && code <= 9999:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
