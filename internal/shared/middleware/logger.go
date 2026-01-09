package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger is a middleware that logs HTTP requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Process request
		next.ServeHTTP(ww, r)

		// Log request details
		duration := time.Since(start)

		event := log.Info()
		if ww.Status() >= 400 {
			event = log.Error()
		}

		event.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Int("status", ww.Status()).
			Int("bytes", ww.BytesWritten()).
			Dur("duration", duration).
			Str("user_agent", r.UserAgent()).
			Msg("HTTP request")
	})
}

// SetupLogger configures the global logger
func SetupLogger(level string) {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	// Use console writer for better readability in development
	zerolog.TimeFieldFormat = time.RFC3339

	log.Logger = log.With().Caller().Logger()

	log.Info().
		Str("level", logLevel.String()).
		Msg("Logger initialized")
}
