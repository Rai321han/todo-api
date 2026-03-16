package middlewares

import (
	"net/http"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

// responseWriter is a custom http.ResponseWriter that captures the status code for logging purposes.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newResponseWriter creates a new instance of responseWriter, initializing the status code to 200 OK by default.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code and calls the underlying ResponseWriter's WriteHeader method to send the response to the client.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger is a Beego filter function that logs incoming HTTP requests, including the method, URL, status code, and duration of the request.
// It only logs requests in development mode to avoid performance overhead in production.
func RequestLogger(next beego.FilterFunc) beego.FilterFunc {
	return func(ctx *context.Context) {

		// log only in dev mode to avoid performance overhead in production
		if beego.BConfig.RunMode != "dev" {
			next(ctx)
			return
		}
		start := time.Now()

		// Wrap the response writer to capture status code
		wrapped := newResponseWriter(ctx.ResponseWriter.ResponseWriter)
		ctx.ResponseWriter.ResponseWriter = wrapped

		defer func() {
			duration := time.Since(start)
			logs.Info("%s %s | %d | %v",
				ctx.Input.Method(),
				ctx.Input.URL(),
				wrapped.statusCode,
				duration,
			)
		}()

		next(ctx)
	}
}
