package http_server

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewRateLimiterMiddleware() *RateLimiter {
	return NewRateLimiter(10, 20)
}

func NewHTTPServer(lc fx.Lifecycle, mux *http.ServeMux, log *zap.Logger, rateLimiter *RateLimiter, requestLogger *RequestLogger) *http.Server {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	var handler http.Handler = mux
	handler = rateLimiter.Middleware(handler)
	handler = requestLogger.Middleware(handler)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           http.TimeoutHandler(handler, 30*time.Second, "Request timeout"),
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Info("Starting HTTP server",
				zap.String("addr", srv.Addr),
				zap.Duration("read_timeout", srv.ReadTimeout),
				zap.Duration("write_timeout", srv.WriteTimeout),
				zap.String("rate_limit", "10 req/s, burst 20"),
				zap.Bool("request_logging", true),
				zap.Bool("correlation_ids", true),
			)
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down HTTP server gracefully...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Error("HTTP server shutdown error", zap.Error(err))
				return err
			}

			log.Info("HTTP server stopped successfully")
			return nil
		},
	})
	return srv
}
