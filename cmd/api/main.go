package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ahorrapp/internal/adapter/crypto"
	"ahorrapp/internal/adapter/events"
	httpapi "ahorrapp/internal/adapter/http"
	"ahorrapp/internal/adapter/jwt"
	"ahorrapp/internal/adapter/ocr"
	"ahorrapp/internal/adapter/postgres"
	"ahorrapp/internal/adapter/redis"
	"ahorrapp/internal/adapter/storage"
	"ahorrapp/internal/config"
	"ahorrapp/internal/usecase"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	pgPool, err := postgres.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pgPool.Close()

	redisClient := redis.NewClient(cfg.RedisAddr, cfg.RedisPassword)
	defer redisClient.Close()

	storageClient, err := storage.NewClient(
		cfg.S3Endpoint,
		cfg.S3AccessKey,
		cfg.S3SecretKey,
		cfg.S3Bucket,
		cfg.S3UseSSL,
	)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	ocrClient := ocr.NewPaddleOCRProvider(cfg.OCRBaseURL)

	healthUC := usecase.NewHealthUseCase(
		postgres.NewChecker(pgPool),
		redis.NewChecker(redisClient),
	)
	receiptRepo := postgres.NewReceiptRepository(pgPool)
	userRepo := postgres.NewUserRepository(pgPool)
	ocrQueue := redis.NewOCRQueue(redisClient, cfg.OCRQueueKey)
	receiptUploadUC := usecase.NewReceiptUploadUseCase(receiptRepo, storageClient, ocrQueue)
	receiptGetUC := usecase.NewReceiptGetUseCase(receiptRepo)
	receiptConfirmUC := usecase.NewReceiptConfirmUseCase(receiptRepo, events.NewLogger())
	receiptProcessUC := usecase.NewReceiptProcessUseCase(receiptRepo, ocrClient)
	worker := usecase.NewReceiptWorker(ocrQueue, receiptProcessUC)

	workerCtx, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()
	go worker.Run(workerCtx)

	// Auth and profile wiring (E2)
	bcryptHasher := crypto.NewBcryptHasher()
	tokenService := jwt.NewTokenService(cfg.JWTSecret)
	authUC := usecase.NewAuthUseCase(userRepo, bcryptHasher, tokenService)
	profileUC := usecase.NewProfileUseCase(userRepo)

	healthHandler := httpapi.NewHealthHandler(healthUC)
	authHandler := httpapi.NewAuthHandler(authUC)
	profileHandler := httpapi.NewProfileHandler(profileUC)
	rankingHandler := httpapi.NewRankingHandler()
	receiptHandler := httpapi.NewReceiptHandler(receiptUploadUC, receiptGetUC, receiptConfirmUC)
	router := httpapi.NewRouter(
		healthHandler,
		authHandler,
		profileHandler,
		rankingHandler,
		receiptHandler.RegisterRoutes,
		httpapi.JWTMiddleware(tokenService),
	)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stopWorker()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func runMigrations(databaseURL string) error {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
