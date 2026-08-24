package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/spdeepak/aegis/server/api"
	"github.com/spdeepak/aegis/server/internal/config"
	"github.com/spdeepak/aegis/server/internal/db"
	"github.com/spdeepak/aegis/server/internal/jwt_secret"
	"github.com/spdeepak/aegis/server/internal/middleware"
	"github.com/spdeepak/aegis/server/internal/permissions"
	"github.com/spdeepak/aegis/server/internal/ratelimit"
	"github.com/spdeepak/aegis/server/internal/roles"
	"github.com/spdeepak/aegis/server/internal/tokens"
	"github.com/spdeepak/aegis/server/internal/twoFA"
	"github.com/spdeepak/aegis/server/internal/users"
	"github.com/spdeepak/aegis/server/pkg/logging"
)

func main() {
	slog.SetDefault(slog.New(logging.NewDefaultHandler()))
	cfg := config.NewConfiguration()

	err := db.RunMigrations(cfg.Postgres)
	if err != nil {
		slog.Error("Failed to run migrations", "error", err)
	}
	dbConnection := db.Connect(cfg.Postgres)

	//JWT SecretKey
	jwtSecretRepository := jwt_secret.New(dbConnection)
	jwtSecretStorage := jwt_secret.NewStorage(jwtSecretRepository)
	//JWT Token
	tokenRepository := tokens.New(dbConnection)
	tokenService := tokens.NewService(tokenRepository, jwt_secret.GetOrCreateSecret(cfg.Token, jwtSecretStorage), cfg.Token.Issuer)
	//2FA
	twoFAQuery := twoFA.New(dbConnection)
	twoFAService := twoFA.NewService(cfg.TwoFA.AppName, twoFAQuery)
	// Rate-limit store (centralized brute-force state: Postgres or Redis)
	store, err := ratelimit.NewStore(cfg.Auth.RateLimit, dbConnection)
	if err != nil {
		slog.Error("failed to build rate-limit store", "error", err)
		os.Exit(1)
	}
	// Account-aware login throttling is opt-in; when disabled the service gets a
	// no-op store so the (IP-based) middleware limiter keeps working.
	accountStore := ratelimit.Store(store)
	if !cfg.Auth.RateLimit.AccountAware {
		accountStore = &ratelimit.NoopStore{}
	}

	//Users
	userRepository := users.New(dbConnection)
	userService := users.NewService(userRepository, twoFAService, tokenService, accountStore)
	//Roles
	roleQuery := roles.New(dbConnection)
	roleService := roles.NewService(roleQuery)
	//Permissions
	permissionQuery := permissions.New(dbConnection)
	permissionsService := permissions.NewService(permissionQuery)
	//Admin
	adminQuery := users.New(dbConnection)
	adminService := users.NewAdminService(adminQuery)

	//oapi-codegen implementation handler
	server := NewServer(userService, roleService, permissionsService, tokenService, twoFAService, adminService)

	swagger, err := api.GetSwagger()
	if err != nil {
		slog.Error(fmt.Sprintf("Error loading swagger spec\n: %v", os.Stderr), "error", err)
		os.Exit(1)
	}
	swagger.Servers = nil

	authMiddleware := middleware.JWTAuthMiddleware(jwt_secret.GetOrCreateSecret(cfg.Token, jwtSecretStorage), cfg.Auth.SkipPaths, cfg.Token.Issuer)
	rateLimiter := middleware.NewRateLimiter(cfg.Auth.RateLimit, store)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.ContextWithFallback = true

	// Trusted proxies: only these CIDRs may assert the real client IP via
	// X-Forwarded-For. Empty => gin uses RemoteAddr (ignores XFF, safe against
	// spoofing). Operators behind a CDN/ingress must set the proxy CIDRs.
	if cidrs, err := parseTrustedProxies(cfg.Auth.RateLimit.TrustedProxies); err != nil {
		slog.Error("invalid trustedProxies configuration", "error", err)
		os.Exit(1)
	} else if err := router.SetTrustedProxies(cidrs); err != nil {
		slog.Error("failed to set trusted proxies", "error", err)
		os.Exit(1)
	}

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.Use(middleware.MetricHandler(),
		rateLimiter.Middleware(),
		middleware.RequestValidator(swagger),
		authMiddleware,
		gin.Recovery(),
		middleware.ErrorMiddleware,
		middleware.GinLogger())
	api.RegisterHandlers(router, server)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: router,
	}

	chanErrors := make(chan error)
	// Initializing the Server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		chanErrors <- srv.ListenAndServe()
	}()

	chanSignals := make(chan os.Signal, 1)
	signal.Notify(chanSignals, os.Interrupt, syscall.SIGTERM)

	select {
	case err = <-chanErrors:
		slog.Error(fmt.Sprintf("Unable to run server. Error: %s", err))
		os.Exit(1)
	case s := <-chanSignals:
		slog.Warn(fmt.Sprintf("Warning: Received %s signal, aborting in 5 seconds...", s))
		// The context is used to inform the Server it has 5 seconds to finish the request it is currently handling
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		dbConnection.Close()
		if err = srv.Shutdown(ctx); err != nil {
			slog.Error(fmt.Sprintf("Server forced to shutdown. Error: %s", err))
			os.Exit(1)
		}
		slog.Info("Server exiting gracefully")
	}
}

// parseTrustedProxies validates and returns the configured trusted-proxy CIDRs.
// An empty list is valid and means gin will use RemoteAddr (ignore XFF).
func parseTrustedProxies(proxies []string) ([]string, error) {
	cidrs := make([]string, 0, len(proxies))
	for _, p := range proxies {
		if _, _, err := net.ParseCIDR(p); err != nil {
			return nil, err
		}
		cidrs = append(cidrs, p)
	}
	return cidrs, nil
}
