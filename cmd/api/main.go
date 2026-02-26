package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nate/go-boilerplate/internal/config"
	apphttp "github.com/nate/go-boilerplate/internal/http"
	httpmiddleware "github.com/nate/go-boilerplate/internal/http/middleware"
	"github.com/nate/go-boilerplate/internal/repository"
	"github.com/nate/go-boilerplate/internal/service"
	"github.com/nate/go-boilerplate/pkg/jwt"
	"github.com/nate/go-boilerplate/pkg/password"
	"github.com/nate/go-boilerplate/pkg/validation"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	validation.Init()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: parseRedisAddr(cfg.RedisURL),
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	userRepo := repository.NewUserRepository(pool)
	noteRepo := repository.NewNoteRepository(pool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(redisClient)
	rateLimitRepo := repository.NewRateLimitRepository(redisClient)

	jwtManager, err := jwt.NewManagerFromFiles(
		cfg.JWT.AccessPrivateKeyPath,
		cfg.JWT.AccessPublicKeyPath,
		cfg.JWT.RefreshPrivateKeyPath,
		cfg.JWT.RefreshPublicKeyPath,
		cfg.JWT.AccessTTL,
		cfg.JWT.RefreshTTL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", err)
	}

	passwordHasher := password.NewHasher(password.Config{
		Memory:      cfg.Argon2.Memory,
		Iterations:  cfg.Argon2.Iterations,
		Parallelism: cfg.Argon2.Parallelism,
		SaltLength:  cfg.Argon2.SaltLength,
		KeyLength:   cfg.Argon2.KeyLength,
	})

	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtManager, passwordHasher, cfg.JWT.RefreshTTL)
	userService := service.NewUserService(userRepo)
	noteService := service.NewNoteService(noteRepo)

	rateLimiter := httpmiddleware.NewRateLimiter(rateLimitRepo, cfg.RateLimit.Requests, int64(cfg.RateLimit.Window.Seconds()))

	router := apphttp.NewRouter(authService, userService, noteService, jwtManager, rateLimiter)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func runMigrations(databaseURL string) error {
	db, err := goose.OpenDBWithDriver("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, "db/migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func parseRedisAddr(redisURL string) string {
	return redisURL
}
