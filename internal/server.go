package internal

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	appctx "github.com/brave-intl/bat-go/libs/context"
	"github.com/brave-intl/bat-go/libs/logging"
	"github.com/brave/go-sync/cache"
	syncContext "github.com/brave/go-sync/context"
	"github.com/brave/go-sync/controller"
	"github.com/brave/go-sync/middleware"
	syncMiddleware "github.com/brave/go-sync/middleware"
	"github.com/go-chi/chi/v5"
	chiware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

const (
	defaultTimeout  = 60 * time.Second
	shutdownTimeout = 30 * time.Second
)

// StartServer initializes and starts the HTTP server with graceful shutdown handling.
func StartServer(bindAddr, dbPath string) error {
	ctx := context.Background()
	ctx, logger := setupLogger(ctx)

	ctx, router, err := setupRouter(ctx, logger, dbPath)
	if err != nil {
		return fmt.Errorf("failed to setup router: %w", err)
	}

	server := &http.Server{
		Addr:    bindAddr,
		Handler: router,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine so we can listen for signals concurrently
	errChan := make(chan error, 1)
	go func() {
		logger.Info().Str("address", bindAddr).Msg("Starting HTTP server")
		errChan <- server.ListenAndServe()
	}()

	// Wait for either server error or shutdown signal
	select {
	case err := <-errChan:
		if errors.Is(err, http.ErrServerClosed) {
			logger.Info().Msg("HTTP server closed gracefully")
			return nil
		}
		return fmt.Errorf("HTTP server error: %w", err)
	case sig := <-sigChan:
		logger.Info().Str("signal", sig.String()).Msg("Received shutdown signal")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error().Err(err).Msg("Failed to shutdown server gracefully, forcing close")
			return server.Close()
		}

		logger.Info().Msg("HTTP server shutdown complete")
		return nil
	}
}

// setupLogger configures the application logger with environment-specific settings.
func setupLogger(ctx context.Context) (context.Context, *zerolog.Logger) {
	ctx = context.WithValue(ctx, appctx.EnvironmentCTXKey, os.Getenv("ENV"))
	ctx = context.WithValue(ctx, appctx.LogLevelCTXKey, zerolog.WarnLevel)
	return logging.SetupLogger(ctx)
}

// setupRouter configures the HTTP router with middleware and routes.
func setupRouter(ctx context.Context, logger *zerolog.Logger, dbPath string) (context.Context, chi.Router, error) {
	router := chi.NewRouter()

	// Middleware setup
	router.Use(chiware.RealIP)
	router.Use(chiware.Heartbeat("/"))

	if logger != nil {
		router.Use(hlog.NewHandler(*logger))
		router.Use(hlog.UserAgentHandler("user_agent"))
		router.Use(hlog.RequestIDHandler("req_id", "Request-Id"))
		// router.Use(batware.RequestLogger(logger))
		// router.Use(httplog.RequestLogger(logger))
	}

	router.Use(chiware.Timeout(defaultTimeout))
	router.Use(bearerToken)
	router.Use(middleware.CommonResponseHeaders)

	// Data store initialization
	sqliteStore, err := NewSqliteDatastore(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create sqlite datastore: %w", err)
	}

	// Cache initialization
	cacheInstance := cache.NewCache(NewFakeRedisClient())

	// Context value injection
	ctx = context.WithValue(ctx, syncContext.ContextKeyDatastore, sqliteStore)
	ctx = context.WithValue(ctx, syncContext.ContextKeyCache, &cacheInstance)

	r := chi.NewRouter()
	r.Use(syncMiddleware.Auth)
	r.Use(syncMiddleware.DisabledChain)
	r.Method("POST", "/command/", controller.Command(cacheInstance, sqliteStore))
	router.Mount("/litesync", r)

	return ctx, router, nil
}

type bearerTokenKey struct{}

// BearerToken is a middleware that adds the bearer token included in a request's headers to context
func bearerToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		bearer := r.Header.Get("Authorization")

		if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
			token = bearer[7:]
		}
		ctx := context.WithValue(r.Context(), bearerTokenKey{}, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
