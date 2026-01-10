package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"woragis-posts-service/internal/config"
	"woragis-posts-service/internal/database"
	"woragis-posts-service/pkg/health"
	applogger "woragis-posts-service/pkg/logger"
	appmetrics "woragis-posts-service/pkg/metrics"
	apptracing "woragis-posts-service/pkg/tracing"
	appsecurity "woragis-posts-service/pkg/security"
	apptimeout "woragis-posts-service/pkg/timeout"
	
	postsdomain "woragis-posts-service/internal/domains"
)

func main() {
	// Validate required environment variables early
	validateRequiredEnvVars()

	// Load configuration first to get environment
	cfg := config.Load()
	env := cfg.Env
	if env == "" {
		env = os.Getenv("ENV")
		if env == "" {
			env = "development"
		}
	}

	// Setup structured logger with trace ID support
	slogLogger := applogger.New(env)

	// Log all environment variables for debugging
	logEnvironmentVariables()

	// Initialize OpenTelemetry tracing
	tracingShutdown, err := apptracing.Init(apptracing.Config{
		ServiceName:    cfg.AppName,
		ServiceVersion: "1.0.0", // TODO: Get from build info
		Environment:    env,
		JaegerEndpoint: os.Getenv("JAEGER_ENDPOINT"), // Defaults to http://jaeger:4318
	})
	if err != nil {
		slogLogger.Warn("failed to initialize tracing", "error", err)
	} else {
		slogLogger.Info("tracing initialized", "service", cfg.AppName)
		defer func() {
			if tracingShutdown != nil {
				tracingShutdown()
			}
		}()
	}

	// Load database and Redis configs
	dbCfg := config.LoadDatabaseConfig()
	redisCfg := config.LoadRedisConfig()
	
	// Initialize database manager
	dbManager, err := database.NewFromConfig(dbCfg, redisCfg)
	if err != nil {
		slogLogger.Error("failed to initialize database manager", "error", err)
		os.Exit(1)
	}
	defer dbManager.Close()
	
	// Perform initial health check
	if err := dbManager.HealthCheck(); err != nil {
		slogLogger.Warn("Database health check failed", "error", err)
	} else {
		slogLogger.Info("All database connections are healthy")
	}

	// Run migrations
	if err := postsdomain.MigratePostsTables(dbManager.GetPostgres()); err != nil {
		slogLogger.Error("failed to run posts migrations", "error", err)
		os.Exit(1)
	}

	// Create Fiber app
	app := config.CreateFiberApp(cfg)

	// Recovery middleware (early in chain)
	app.Use(recover.New())

	// Security headers middleware (must be early, before other middlewares)
	app.Use(appsecurity.SecurityHeadersMiddleware())

	// Request timeout middleware (30 seconds default)
	app.Use(apptimeout.Middleware(apptimeout.DefaultConfig()))

	// Add OpenTelemetry tracing middleware (must be first to extract trace context)
	app.Use(apptracing.Middleware(cfg.AppName))
	// Add request ID middleware for distributed tracing (works with tracing, preserves trace_id)
	app.Use(applogger.RequestIDMiddleware(slogLogger))
	// Add structured request logging middleware
	app.Use(applogger.RequestLoggerMiddleware(slogLogger))
	// Add Prometheus metrics middleware
	app.Use(appmetrics.Middleware())

	// Request size limit (10MB)
	app.Use(appsecurity.RequestSizeLimitMiddleware(10 * 1024 * 1024))

	// Input sanitization
	app.Use(appsecurity.InputSanitizationMiddleware())

	// CSRF protection (for state-changing requests)
	csrfCfg := appsecurity.DefaultCSRFConfig(dbManager.GetRedis())
	app.Use(appsecurity.CSRFMiddleware(csrfCfg))

	// Rate limiting (100 requests per minute per IP/user)
	app.Use(appsecurity.RateLimitMiddleware(100, time.Minute))

	// CORS middleware (if enabled)
	corsCfg := config.LoadCORSConfig()
	if corsCfg.Enabled {
		config.SetupCORS(app, corsCfg)
	}

	// Initialize health checker
	healthChecker := health.NewHealthChecker(dbManager.GetPostgres(), dbManager.GetRedis(), slogLogger)

	// Health check endpoints (before API routes, no auth required)
	app.Get("/healthz", healthChecker.Handler())           // Combined health check
	app.Get("/healthz/live", healthChecker.LivenessHandler())   // Liveness probe (Kubernetes)
	app.Get("/healthz/ready", healthChecker.ReadinessHandler()) // Readiness probe (Kubernetes)

	// Prometheus metrics endpoint (before API routes, no auth required)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// API routes group
	api := app.Group("/api/v1")

	// Load auth service URL for JWT validation
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:3000"
	}

	// Setup posts domain routes
	postsdomain.SetupRoutes(api, dbManager.GetPostgres(), authServiceURL, slogLogger)

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		slogLogger.Info("starting posts service", "addr", addr, "env", env)
		if err := app.Listen(addr); err != nil {
			slogLogger.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	slogLogger.Info("shutting down posts service gracefully")

	// Give ongoing requests time to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		slogLogger.Error("error during shutdown", "error", err)
	}

	slogLogger.Info("posts service stopped")
}

// maskValue returns a masked version of the value for display purposes
func maskValue(val string) string {
	if val == "" {
		return "<not set>"
	}
	if len(val) <= 4 {
		return "****"
	}
	return val[:4] + strings.Repeat("*", len(val)-4)
}

// logEnvironmentVariables logs all environment variables with their status and masked values
func logEnvironmentVariables() {
	slog.Info("====== ENVIRONMENT VARIABLES ======")

	// Required variables
	slog.Info("Required Variables:")
	slog.Info("  DATABASE_URL", "status", getVarStatus("DATABASE_URL"), "value", maskValue(os.Getenv("DATABASE_URL")))
	slog.Info("  REDIS_URL", "status", getVarStatus("REDIS_URL"), "value", maskValue(os.Getenv("REDIS_URL")))
	slog.Info("  AES_KEY", "status", getVarStatus("AES_KEY"), "value", maskValue(os.Getenv("AES_KEY")))
	slog.Info("  HASH_SALT", "status", getVarStatus("HASH_SALT"), "value", maskValue(os.Getenv("HASH_SALT")))

	// Application settings
	slog.Info("Application Variables:")
	slog.Info("  APP_NAME", "status", getVarStatus("APP_NAME"), "value", os.Getenv("APP_NAME"))
	slog.Info("  APP_PORT", "status", getVarStatus("APP_PORT"), "value", os.Getenv("APP_PORT"))
	slog.Info("  APP_ENV", "status", getVarStatus("APP_ENV"), "value", os.Getenv("APP_ENV"))
	slog.Info("  APP_PUBLIC_URL", "status", getVarStatus("APP_PUBLIC_URL"), "value", os.Getenv("APP_PUBLIC_URL"))

	// Database settings
	slog.Info("Database Variables:")
	slog.Info("  POSTGRES_USER", "status", getVarStatus("POSTGRES_USER"), "value", os.Getenv("POSTGRES_USER"))
	slog.Info("  POSTGRES_DB", "status", getVarStatus("POSTGRES_DB"), "value", os.Getenv("POSTGRES_DB"))
	slog.Info("  POSTGRES_PASSWORD", "status", getVarStatus("POSTGRES_PASSWORD"), "value", maskValue(os.Getenv("POSTGRES_PASSWORD")))
	slog.Info("  POSTGRES_PORT", "status", getVarStatus("POSTGRES_PORT"), "value", os.Getenv("POSTGRES_PORT"))

	// Redis settings
	slog.Info("Redis Variables:")
	slog.Info("  REDIS_DB", "status", getVarStatus("REDIS_DB"), "value", os.Getenv("REDIS_DB"))
	slog.Info("  REDIS_PORT", "status", getVarStatus("REDIS_PORT"), "value", os.Getenv("REDIS_PORT"))

	// JWT/Auth settings
	slog.Info("JWT/Auth Variables:")
	slog.Info("  AUTH_JWT_SECRET", "status", getVarStatus("AUTH_JWT_SECRET"), "value", maskValue(os.Getenv("AUTH_JWT_SECRET")))
	slog.Info("  AUTH_JWT_TTL", "status", getVarStatus("AUTH_JWT_TTL"), "value", os.Getenv("AUTH_JWT_TTL"))

	// RabbitMQ settings
	slog.Info("RabbitMQ Variables:")
	slog.Info("  RABBITMQ_URL", "status", getVarStatus("RABBITMQ_URL"), "value", maskValue(os.Getenv("RABBITMQ_URL")))
	slog.Info("  RABBITMQ_USER", "status", getVarStatus("RABBITMQ_USER"), "value", os.Getenv("RABBITMQ_USER"))
	slog.Info("  RABBITMQ_PASSWORD", "status", getVarStatus("RABBITMQ_PASSWORD"), "value", maskValue(os.Getenv("RABBITMQ_PASSWORD")))
	slog.Info("  RABBITMQ_HOST", "status", getVarStatus("RABBITMQ_HOST"), "value", os.Getenv("RABBITMQ_HOST"))
	slog.Info("  RABBITMQ_PORT", "status", getVarStatus("RABBITMQ_PORT"), "value", os.Getenv("RABBITMQ_PORT"))
	slog.Info("  RABBITMQ_VHOST", "status", getVarStatus("RABBITMQ_VHOST"), "value", os.Getenv("RABBITMQ_VHOST"))

	// SMTP settings
	slog.Info("SMTP Variables:")
	slog.Info("  SMTP_HOST", "status", getVarStatus("SMTP_HOST"), "value", os.Getenv("SMTP_HOST"))
	slog.Info("  SMTP_PORT", "status", getVarStatus("SMTP_PORT"), "value", os.Getenv("SMTP_PORT"))
	slog.Info("  SMTP_USERNAME", "status", getVarStatus("SMTP_USERNAME"), "value", maskValue(os.Getenv("SMTP_USERNAME")))
	slog.Info("  SMTP_PASSWORD", "status", getVarStatus("SMTP_PASSWORD"), "value", maskValue(os.Getenv("SMTP_PASSWORD")))
	slog.Info("  SMTP_FROM", "status", getVarStatus("SMTP_FROM"), "value", os.Getenv("SMTP_FROM"))
	slog.Info("  SMTP_TLS", "status", getVarStatus("SMTP_TLS"), "value", os.Getenv("SMTP_TLS"))

	// CORS settings
	slog.Info("CORS Variables:")
	slog.Info("  CORS_ENABLED", "status", getVarStatus("CORS_ENABLED"), "value", os.Getenv("CORS_ENABLED"))
	slog.Info("  CORS_ALLOWED_ORIGINS", "status", getVarStatus("CORS_ALLOWED_ORIGINS"), "value", os.Getenv("CORS_ALLOWED_ORIGINS"))
	slog.Info("  CORS_ALLOWED_METHODS", "status", getVarStatus("CORS_ALLOWED_METHODS"), "value", os.Getenv("CORS_ALLOWED_METHODS"))

	// Observability settings
	slog.Info("Observability Variables:")
	slog.Info("  MONITORING_ENABLED", "status", getVarStatus("MONITORING_ENABLED"), "value", os.Getenv("MONITORING_ENABLED"))
	slog.Info("  METRICS_NAMESPACE", "status", getVarStatus("METRICS_NAMESPACE"), "value", os.Getenv("METRICS_NAMESPACE"))
	slog.Info("  OTLP_ENDPOINT", "status", getVarStatus("OTLP_ENDPOINT"), "value", os.Getenv("OTLP_ENDPOINT"))
	slog.Info("  JAEGER_ENDPOINT", "status", getVarStatus("JAEGER_ENDPOINT"), "value", maskValue(os.Getenv("JAEGER_ENDPOINT")))

	// Service URLs
	slog.Info("Service URLs (Optional):")
	slog.Info("  AUTH_SERVICE_URL", "status", getVarStatus("AUTH_SERVICE_URL"), "value", os.Getenv("AUTH_SERVICE_URL"))
	slog.Info("  AI_SERVICE_URL", "status", getVarStatus("AI_SERVICE_URL"), "value", os.Getenv("AI_SERVICE_URL"))

	slog.Info("====== END ENVIRONMENT VARIABLES ======")
}

// getVarStatus returns a symbol indicating the status of an environment variable
func getVarStatus(key string) string {
	val := os.Getenv(key)
	if val == "" {
		return "✗"
	}
	return "✓"
}

// validateRequiredEnvVars ensures all required environment variables are set
func validateRequiredEnvVars() {
	requiredVars := []string{
		"DATABASE_URL",
		"REDIS_URL",
		"AES_KEY",
		"HASH_SALT",
	}

	// JWT_SECRET is only required in production
	if os.Getenv("APP_ENV") == "production" {
		requiredVars = append(requiredVars, "AUTH_JWT_SECRET")
	}

	var missingVars []string
	for _, varName := range requiredVars {
		if os.Getenv(varName) == "" {
			missingVars = append(missingVars, varName)
		}
	}

	if len(missingVars) > 0 {
		slog.Error("missing required environment variables",
			"missing", strings.Join(missingVars, ", "),
			"count", len(missingVars),
		)
		os.Exit(1)
	}
}

