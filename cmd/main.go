package main

import (
	"flag"
	"log"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/chilly266futon/spotService/internal/config"
	"github.com/chilly266futon/spotService/internal/domain"
	"github.com/chilly266futon/spotService/internal/service"
	"github.com/chilly266futon/spotService/internal/storage"

	spotpb "github.com/chilly266futon/exchange-service-contracts/gen/pb/spot"

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

	// Инициализация logger из конфига
	l, err := logger.New(cfg.Logger)
	if err != nil {
		log.Fatalf("failed to create l: %v", err)
	}
	defer l.Sync()

	l.Info("starting spot-service",
		zap.String("version", "1.0.0"),
		zap.String("config", *configPath),
	)

	// Инициализация хранилища с данными из конфига
	markets := loadMarketsFromConfig(cfg)
	marketStorage := storage.NewMarketStorage(markets)

	l.Info("loaded markets",
		zap.Int("count", marketStorage.Count()),
	)

	// Создание сервиса
	spotService := service.NewService(marketStorage, l)

	// Настройка interceptors
	var interceptorChain []grpc.ServerOption

	// Trace ID
	interceptorChain = append(interceptorChain,
		grpc.ChainUnaryInterceptor(interceptors.TraceIDInterceptor()),
	)

	// Panic recovery
	interceptorChain = append(interceptorChain,
		grpc.ChainUnaryInterceptor(interceptors.UnaryPanicRecoveryInterceptor(l)),
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

		l.Info("rate limiting enabled")
	}

	// Logger
	interceptorChain = append(interceptorChain,
		grpc.ChainUnaryInterceptor(interceptors.LoggerInterceptor(l)),
	)

	// Создание gRPC сервера
	grpcServer, err := grpcutil.NewServer(
		grpcutil.ServerConfig{
			Host:            cfg.Server.Host,
			Port:            cfg.Server.Port,
			ShutdownTimeout: cfg.Server.ShutdownTimeout,
		},
		l,
		interceptorChain...,
	)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	// Регистрация сервисов
	spotpb.RegisterSpotInstrumentServiceServer(grpcServer.GRPCServer(), spotService)

	// Health check
	if cfg.Health.Enabled {
		healthServer := health.NewServer()
		healthServer.SetHealthy("spot_v1.SpotInstrumentService")
		grpc_health_v1.RegisterHealthServer(grpcServer.GRPCServer(), healthServer)
		l.Info("health check enabled")
	}

	// Reflection для grpcui
	reflection.Register(grpcServer.GRPCServer())

	// Запуск сервера (с graceful shutdown)
	l.Info("server ready to accept connections")
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
func parseRoles(roles []string) []spotpb.UserRole {
	result := make([]spotpb.UserRole, 0, len(roles))

	roleMap := map[string]spotpb.UserRole{
		"USER_ROLE_COMMON":   spotpb.UserRole_USER_ROLE_COMMON,
		"USER_ROLE_VERIFIED": spotpb.UserRole_USER_ROLE_VERIFIED,
		"USER_ROLE_PREMIUM":  spotpb.UserRole_USER_ROLE_PREMIUM,
		"USER_ROLE_ADMIN":    spotpb.UserRole_USER_ROLE_ADMIN,
	}

	for _, roleStr := range roles {
		if role, ok := roleMap[roleStr]; ok {
			result = append(result, role)
		}
	}

	return result
}
