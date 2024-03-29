package internal

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	appctx "github.com/brave-intl/bat-go/libs/context"
	"github.com/brave-intl/bat-go/libs/logging"
	batware "github.com/brave-intl/bat-go/libs/middleware"
	"github.com/brave/go-sync/cache"
	syncContext "github.com/brave/go-sync/context"
	"github.com/brave/go-sync/controller"
	"github.com/brave/go-sync/middleware"
	"github.com/go-chi/chi"
	chiware "github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

func setupLogger(ctx context.Context) (context.Context, *zerolog.Logger) {
	ctx = context.WithValue(ctx, appctx.EnvironmentCTXKey, os.Getenv("ENV"))
	ctx = context.WithValue(ctx, appctx.LogLevelCTXKey, zerolog.WarnLevel)
	return logging.SetupLogger(ctx)
}

func setupRouter(ctx context.Context, logger *zerolog.Logger, dbFile string) (context.Context, *chi.Mux) {
	r := chi.NewRouter()

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

	sqliteStore, _ := NewSqliteDatastore(dbFile)
	cache := cache.NewCache(NewFakeRedisClient())

	ctx = context.WithValue(ctx, syncContext.ContextKeyDatastore, sqliteStore)
	ctx = context.WithValue(ctx, syncContext.ContextKeyCache, &cache)

	r.Mount("/litesync", controller.SyncRouter(cache, sqliteStore))

	return ctx, r
}

func StartServer(port string, dbFile string) {
	serverCtx, logger := setupLogger(context.Background())
	serverCtx, r := setupRouter(serverCtx, logger, dbFile)
	srv := http.Server{Addr: port, Handler: chi.ServerBaseContext(serverCtx, r)}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	go func() {
		<-sig
		_ = srv.Shutdown(serverCtx)
	}()
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Info().Msg("HTTP server closed")
	} else if err != nil {
		log.Panic().Err(err).Msg("HTTP server start failed!")
	}
}
