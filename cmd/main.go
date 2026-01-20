package main

import (
	"flag"
	"log"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	spotv1 "github.com/chilly266futon/spotService/gen/pb"
	"github.com/chilly266futon/spotService/internal/config"
	"github.com/chilly266futon/spotService/internal/domain"
	"github.com/chilly266futon/spotService/internal/service"
	"github.com/chilly266futon/spotService/internal/storage"

	// Shared пакеты из того же репозитория
	"github.com/chilly266futon/spotService/pkg/shared/grpcutil"
	"github.com/chilly266futon/spotService/pkg/shared/health"
	"github.com/chilly266futon/spotService/pkg/shared/interceptors"
	"github.com/chilly266futon/spotService/pkg/shared/logger"
)

func main() {
	// Парсинг флагов
	configPath := flag.String("config", "configs/config.yaml", "Path to config file")
	flag.Parse()

	// Загрузка конфигурации
	cfg := config.MustLoad(*configPath)

	// Инициализация logger
	logger, err := logger.New(cfg.Logger)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("starting spot-service",
		zap.String("version", "1.0.0"),
		zap.String("config", *configPath),
	)

	// Инициализация хранилища с данными из конфига
	markets := loadMarketsFromConfig(cfg)
	marketStorage := storage.NewMarketStorage(markets)

	logger.Info("loaded markets",
		zap.Int("count", marketStorage.Count()),
	)

	// Создание сервиса
	spotService := service.NewService(marketStorage)

	// Настройка interceptors
	var interceptorChain []grpc.ServerOption

	// Trace ID
	interceptorChain = append(interceptorChain,
		grpc.ChainUnaryInterceptor(interceptors.TraceIDInterceptor()),
	)

	// Panic recovery
	interceptorChain = append(interceptorChain,
		grpc.ChainUnaryInterceptor(interceptors.UnaryPanicRecoveryInterceptor(logger)),
	)

	// Rate limiting
	if cfg.RateLimit.Enabled {
		rateLimiter := interceptors.NewMethodRateLimiterInterceptor(
			rate.Limit(cfg.RateLimit.RequestsPerSecond),
			cfg.RateLimit.Burst,
		)

		// Лимиты для конкретных методов
		for method, limit := range cfg.RateLimit.Methods {
			rateLimiter.SetMethodLimit(method, rate.Limit(limit.RequestsPerSecond), limit.Burst)
		}

		interceptorChain = append(interceptorChain,
			grpc.ChainUnaryInterceptor(rateLimiter.Interceptor()),
		)

		logger.Info("rate limiting enabled")
	}

	// Logger interceptor (последний в цепочке)
	interceptorChain = append(interceptorChain,
		grpc.ChainUnaryInterceptor(interceptors.LoggerInterceptor(logger)),
	)

	// Создание gRPC сервера
	grpcServer, err := grpcutil.NewServer(
		grpcutil.ServerConfig{
			Host:            cfg.Server.Host,
			Port:            cfg.Server.Port,
			ShutdownTimeout: cfg.Server.ShutdownTimeout,
		},
		logger,
		interceptorChain...,
	)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	// Регистрация сервисов
	spotv1.RegisterSpotInstrumentServiceServer(grpcServer.GRPCServer(), spotService)

	// Health check
	if cfg.Health.Enabled {
		healthServer := health.NewServer()
		healthServer.SetHealthy("spot.v1.SpotInstrumentService")
		grpc_health_v1.RegisterHealthServer(grpcServer.GRPCServer(), healthServer)
		logger.Info("health check enabled")
	}

	// Reflection для grpcui
	reflection.Register(grpcServer.GRPCServer())

	// Запуск сервера (с graceful shutdown)
	if err := grpcServer.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// loadMarketsFromConfig загружает рынки из конфига
func loadMarketsFromConfig(cfg *config.Config) []*domain.Market {
	markets := make([]*domain.Market, 0, len(cfg.Markets))

	for _, mc := range cfg.Markets {
		market := &domain.Market{
			ID:           mc.ID,
			Name:         mc.Name,
			Description:  mc.Description,
			Enabled:      mc.Enabled,
			AllowedRoles: parseRoles(mc.AllowedRoles),
		}
		markets = append(markets, market)
	}

	return markets
}

// parseRoles преобразует строки в proto enums
func parseRoles(roles []string) []spotv1.UserRole {
	result := make([]spotv1.UserRole, 0, len(roles))

	roleMap := map[string]spotv1.UserRole{
		"USER_ROLE_COMMON":   spotv1.UserRole_USER_ROLE_COMMON,
		"USER_ROLE_VERIFIED": spotv1.UserRole_USER_ROLE_VERIFIED,
		"USER_ROLE_PREMIUM":  spotv1.UserRole_USER_ROLE_PREMIUM,
		"USER_ROLE_ADMIN":    spotv1.UserRole_USER_ROLE_ADMIN,
	}

	for _, roleStr := range roles {
		if role, ok := roleMap[roleStr]; ok {
			result = append(result, role)
		}
	}

	return result
}
