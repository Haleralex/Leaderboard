package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authhandler "leaderboard-service/internal/auth/handler"
	authrepo "leaderboard-service/internal/auth/repository"
	authservice "leaderboard-service/internal/auth/service"
	"leaderboard-service/internal/handlers"
	leaderboardhandler "leaderboard-service/internal/leaderboard/handler"
	leaderboardrepo "leaderboard-service/internal/leaderboard/repository"
	leaderboardservice "leaderboard-service/internal/leaderboard/service"
	"leaderboard-service/internal/shared/config"
	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/middleware"
	"leaderboard-service/internal/shared/repository/decorators"
	"leaderboard-service/internal/websocket"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Setup logger
	middleware.SetupLogger(cfg.Log.Level)

	log.Info().Msg("Starting Leaderboard Service")

	// Initialize database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer db.Close()

	// Initialize Redis (optional for testing)
	redis, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Warn().Err(err).Msg("Redis not available, running without cache")
		redis = nil
	} else {
		defer redis.Close()
	}

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTMiddleware(cfg)
	rateLimiter := middleware.NewRateLimiter(cfg)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize WebSocket Hub
	wsHub := websocket.NewHub(
		ctx,
		cfg.GetWebSocketBroadcastInterval(),
		cfg.WebSocket.DefaultLimit,
	)
	go wsHub.Run() // Start hub in background goroutine

	// Create shared cache for decorators
	cache := decorators.NewSimpleCache()

	// Initialize base repositories
	baseUserRepo := authrepo.NewPostgresUserRepository(db)
	baseScoreRepo := leaderboardrepo.NewPostgresScoreRepository(db)

	// Wrap repositories with decorators (Decorator Pattern)
	// Order: base → cached → logged (outermost)
	userRepo := decorators.NewLoggedUserRepository(
		decorators.NewCachedUserRepository(baseUserRepo, cache),
	)
	scoreRepo := decorators.NewLoggedScoreRepository(
		decorators.NewCachedScoreRepository(baseScoreRepo, cache),
	)

	log.Info().Msg("✅ Repositories initialized with caching and logging decorators")

	// Initialize services with decorated repositories
	authService := authservice.NewAuthService(userRepo, jwtMiddleware, cfg)
	leaderboardService := leaderboardservice.NewLeaderboardService(scoreRepo, userRepo, redis, cfg)
	leaderboardService.SetHub(wsHub) // Connect WebSocket broadcasting

	// Initialize handlers (wsHandler needs leaderboardService for initial snapshots)
	wsHandler := handlers.NewWebSocketHandler(wsHub, jwtMiddleware, cfg, leaderboardService)
	authHandler := authhandler.NewAuthHandler(authService)
	leaderboardHandler := leaderboardhandler.NewLeaderboardHandler(leaderboardService)
	healthHandler := handlers.NewHealthHandler(db, redis)

	// Setup router
	r := setupRouter(cfg, jwtMiddleware, rateLimiter, authHandler, leaderboardHandler, healthHandler, wsHandler)

	// Create HTTP server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("address", addr).Msg("Server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server stopped")
}

// setupRouter configures all routes and middleware
func setupRouter(
	_ *config.Config,
	jwtMiddleware *middleware.JWTMiddleware,
	rateLimiter *middleware.RateLimiter,
	authHandler *authhandler.AuthHandler,
	leaderboardHandler *leaderboardhandler.LeaderboardHandler,
	healthHandler *handlers.HealthHandler,
	wsHandler *handlers.WebSocketHandler,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.Recoverer)
	/* r.Use(cors.Handler(middleware.GetCORSOptions())) */
	r.Use(chimiddleware.Timeout(30 * time.Second))

	// Health check endpoints (no auth required)
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Readiness)
	r.Get("/live", healthHandler.Liveness)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public auth endpoints
		r.Group(func(r chi.Router) {
			r.Use(rateLimiter.Limit) // Apply rate limiting
			r.Post("/auth/register", authHandler.Register)
			r.Post("/auth/login", authHandler.Login)
		})

		// Protected leaderboard endpoints (JWT DISABLED FOR TESTING)
		r.Group(func(r chi.Router) {
			// r.Use(jwtMiddleware.Authenticate) // Require JWT - DISABLED FOR TESTING
			r.Use(rateLimiter.Limit) // Apply rate limiting

			// Leaderboard operations
			r.Post("/submit-score", leaderboardHandler.SubmitScore)
			r.Get("/leaderboard", leaderboardHandler.GetLeaderboard)
			r.Get("/leaderboard/user/{userID}", leaderboardHandler.GetUserRank)
		})

		// WebSocket endpoints (NO middleware - validates token from query param)
		r.Get("/ws/leaderboard", wsHandler.HandleLeaderboard)

		// Test/Debug endpoints
		r.Group(func(r chi.Router) {
			r.Use(jwtMiddleware.Authenticate)
			r.Post("/test/broadcast", leaderboardHandler.TestBroadcast)
		})

		// WebSocket stats endpoint
		r.Group(func(r chi.Router) {
			r.Use(jwtMiddleware.Authenticate) // Stats requires JWT header
			r.Get("/ws/stats", wsHandler.HandleStats)
		})
	})

	// 404 handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Not Found","message":"endpoint not found","code":404}`))
	})

	return r
}
