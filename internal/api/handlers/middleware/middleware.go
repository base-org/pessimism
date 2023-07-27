package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type HeaderParam = string

const (
	ContentLength HeaderParam = "Content-Length"
	UserAgent     HeaderParam = "User-Agent"
	Host          HeaderParam = "Host"
	ContentType   HeaderParam = "Content-Type"
)

// InjectedLogging uses logging middleware
func InjectedLogging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Error("Failure occurred during request processing")
				}
			}()

			userAgent := r.Header.Get(UserAgent)
			if len(userAgent) == 0 {
				userAgent = "-"
			}

			contentLength := r.Header.Get(ContentLength)
			if len(contentLength) == 0 {
				contentLength = "-"
			}

			host := r.Header.Get(Host)
			if host == "" {
				host = r.RemoteAddr
			}

			contentType := r.Header.Get(ContentType)
			if contentType == "" {
				contentType = "-"
			}

			start := time.Now()
			next.ServeHTTP(w, r)

			logger.Info("HTTP request received",
				zap.String("method", r.Method), zap.String("path", r.URL.EscapedPath()),
				zap.Duration("duration", time.Since(start)), zap.String("content_type", contentType),
				zap.String("content_length", contentLength),
				zap.String("user_agent", userAgent), zap.String("host", host),
			)
		}

		return http.HandlerFunc(fn)
	}
}
