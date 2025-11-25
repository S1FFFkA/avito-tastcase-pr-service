package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"AVITOSAMPISHU/internal/handlers"
	"AVITOSAMPISHU/internal/infrastructure/database"
	"AVITOSAMPISHU/internal/middleware"
	pullrequest_repository "AVITOSAMPISHU/internal/repository/pullrequest_repository"
	reviewer_repository "AVITOSAMPISHU/internal/repository/reviewer_repository"
	team_repository "AVITOSAMPISHU/internal/repository/team_repository"
	user_repository "AVITOSAMPISHU/internal/repository/user_repository"
	"AVITOSAMPISHU/internal/server"
	pullrequest_service "AVITOSAMPISHU/internal/service/pullrequest_service"
	team_service "AVITOSAMPISHU/internal/service/team_service"
	user_service "AVITOSAMPISHU/internal/service/user_service"
	"AVITOSAMPISHU/pkg/logger"
	"AVITOSAMPISHU/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	shutdownTimeoutSeconds = 30
)

// Run инициализирует и запускает приложение
func Run() {
	// Инициализация логгера
	logger.InitLogger()
	defer logger.Sync()

	logger.Logger.Infow("initializing application")

	// Подключение к БД
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	db, err := database.NewDB(dbCtx)
	if err != nil {
		logger.Logger.Fatalw("error connecting to database", "error", err)
	}
	defer db.Close()

	logger.Logger.Infow("database connection established")

	// Инициализация репозиториев
	teamRepo := team_repository.NewTeamStorage(db)
	userRepo := user_repository.NewUserRepository(db)
	prRepo := pullrequest_repository.NewPullRequestStorage(db)
	prReviewersRepo := reviewer_repository.NewPrReviewersStorage(db)

	// Инициализация сервисов
	teamSvc := team_service.NewTeamService(teamRepo, userRepo)
	userSvc := user_service.NewUserService(userRepo, prReviewersRepo, teamRepo)
	prSvc := pullrequest_service.NewPullRequestService(prRepo, prReviewersRepo, userRepo, teamRepo)

	// Создание роутера
	mux := http.NewServeMux()

	// Регистрация метрик
	prometheus.MustRegister(metrics.ReviewerLoadDistribution)
	mux.Handle("/metrics", promhttp.Handler())

	logger.Logger.Infow("metrics registered")

	// Регистрация роутов
	handlers.RegisterRoutes(mux, teamSvc, userSvc, prSvc)

	logger.Logger.Infow("routes registered")

	// Применение middleware (сначала логирование, потом авторизация)
	handler := middleware.AuthMiddleware(middleware.LoggingMiddleware(mux))

	// Создание сервера
	srv := server.NewAPIServer(handler)

	logger.Logger.Infow("server created", "port", os.Getenv("API_PORT"))

	// Запуск сервера в горутине
	go func() {
		logger.Logger.Infow("starting HTTP server")
		if err := srv.Start(); err != nil {
			logger.Logger.Fatalw("error starting server", "error", err)
		}
	}()

	// Graceful shutdown
	// Реализацию взял из вашего репозитория с ошибками прошлых лет :)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Infow("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeoutSeconds*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Logger.Errorw("shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Logger.Infow("server stopped gracefully")
}
