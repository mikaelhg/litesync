package internal

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	batware "github.com/brave-intl/bat-go/middleware"
	appctx "github.com/brave-intl/bat-go/utils/context"
	"github.com/brave-intl/bat-go/utils/logging"
	"github.com/brave/go-sync/cache"
	syncContext "github.com/brave/go-sync/context"
	"github.com/brave/go-sync/controller"
	"github.com/brave/go-sync/middleware"
	sentry "github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	chiware "github.com/go-chi/chi/middleware"
	"github.com/mikaelhg/litesync/internal/litecache"
	"github.com/mikaelhg/litesync/internal/liteds"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

func setupLogger(ctx context.Context) (context.Context, *zerolog.Logger) {
	ctx = context.WithValue(ctx, appctx.EnvironmentCTXKey, os.Getenv("ENV"))
	ctx = context.WithValue(ctx, appctx.LogLevelCTXKey, zerolog.WarnLevel)
	return logging.SetupLogger(ctx)
}

func setupRouter(ctx context.Context, logger *zerolog.Logger) (context.Context, *chi.Mux) {
	r := chi.NewRouter()

	r.Use(chiware.RequestID)
	r.Use(chiware.RealIP)
	r.Use(chiware.Heartbeat("/"))

	if logger != nil {
		r.Use(hlog.NewHandler(*logger))
		r.Use(hlog.UserAgentHandler("user_agent"))
		r.Use(hlog.RequestIDHandler("req_id", "Request-Id"))
		r.Use(batware.RequestLogger(logger))
	}

	r.Use(chiware.Timeout(60 * time.Second))
	r.Use(batware.BearerToken)
	r.Use(middleware.CommonResponseHeaders)

	sqlite_ds := liteds.NewSqliteDatastore()
	cache := cache.NewCache(&litecache.FakeRedisClient{})

	ctx = context.WithValue(ctx, syncContext.ContextKeyDatastore, sqlite_ds)
	ctx = context.WithValue(ctx, syncContext.ContextKeyCache, &cache)

	r.Mount("/v2", controller.SyncRouter(cache, sqlite_ds))
	r.Get("/metrics", batware.Metrics())

	return ctx, r
}

func StartServer() {
	serverCtx, logger := setupLogger(context.Background())

	serverCtx, r := setupRouter(serverCtx, logger)

	port := ":8295"
	srv := http.Server{Addr: port, Handler: chi.ServerBaseContext(serverCtx, r)}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	go func() {
		<-sig
		srv.Shutdown(serverCtx)
	}()

	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		log.Info().Msg("HTTP server closed")
	} else if err != nil {
		sentry.CaptureException(err)
		log.Panic().Err(err).Msg("HTTP server start failed!")
	}
}
