package main

import (
	"ads_backend/internal/ads_service"
	http_server "ads_backend/internal/http"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(
			fx.Annotate(http_server.NewServeMux, fx.ParamTags(`group:"routes"`)),
			http_server.AsRoute(http_server.NewAdsHandler),
			http_server.NewRateLimiterMiddleware,
			http_server.NewRequestLogger,
			http_server.NewHTTPServer,
			ads_service.NewService,
			zap.NewExample,
		),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}
