package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"

	"itms-server/pkg/response"
)

// wrapWriter buffers the response so we can rewrite it in unified format.
type wrapWriter struct {
	http.ResponseWriter
	status int
	buf    *bytes.Buffer
	wroteHeader bool
}

func (w *wrapWriter) WriteHeader(code int) {
	w.status = code
	// Defer actual WriteHeader until we write the wrapped body.
}

func (w *wrapWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *wrapWriter) flush() {
	// Prevent double-write if the handler already wrote headers.
	// (the go-zero gateway doesn't, but be safe)
}

// WrapResponse is the innermost middleware. It buffers the handler output
// and rewrites it into {code, message, data, requestId} format.
func WrapResponse(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := &wrapWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
			buf:            &bytes.Buffer{},
		}
		next(rw, r)

		body := rw.buf.Bytes()
		status := rw.status
		if status == 0 {
			status = http.StatusOK
		}

		var wrapped []byte

		if status >= 200 && status < 300 {
			// Success: wrap raw JSON body
			var data interface{}
			if len(body) > 0 {
				if err := json.Unmarshal(body, &data); err != nil {
					data = json.RawMessage(body)
				}
			}
			wrapped, _ = json.Marshal(response.Success(data))
		} else {
			// Error: map go-zero error to our format
			code, msg := parseGatewayError(body)
			wrapped, _ = json.Marshal(response.Error(code, msg))
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		w.Write(wrapped)
	}
}

// parseGatewayError extracts a go-zero error and maps it to our error code.
// The gateway writes errors via httpx.ErrorCtx which returns JSON like {"code":...,"message":...}
// or writes plain text from gRPC status.
func parseGatewayError(body []byte) (int, string) {
	if len(body) == 0 {
		return response.CodeAuthServerErr, "internal server error"
	}

	var ge struct {
		Code int    `json:"code"`
		Desc string `json:"desc"`
		Msg  string `json:"message"`
	}
	if err := json.Unmarshal(body, &ge); err == nil {
		msg := ge.Desc
		if msg == "" {
			msg = ge.Msg
		}
		if msg == "" {
			msg = string(body)
		}
		// Map go-zero HTTP-status-based code to our business code
		code := response.CodeAuthServerErr
		if ge.Code > 0 {
			code = ge.Code
		}
		return code, msg
	}

	return response.CodeAuthServerErr, string(body)
}
