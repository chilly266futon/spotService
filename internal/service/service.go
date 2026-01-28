package service

import (
	"context"

	"github.com/chilly266futon/spotService/pkg/shared/interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	spotv1 "github.com/chilly266futon/spotService/gen/pb"
	"github.com/chilly266futon/spotService/internal/domain"
	"github.com/chilly266futon/spotService/internal/storage"
)

// Service реализует SpotInstrumentService
type Service struct {
	spotv1.UnimplementedSpotInstrumentServiceServer
	storage *storage.MarketStorage
	logger  *zap.Logger
}

// NewService создает новый сервис
func NewService(storage *storage.MarketStorage, logger *zap.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  logger,
	}
}

// ViewMarkets возвращает список доступных рынков для указанных ролей
func (s *Service) ViewMarkets(ctx context.Context, req *spotv1.ViewMarketsRequest) (*spotv1.ViewMarketsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	traceID := interceptors.GetTraceID(ctx)

	// Получаем все доступные рынки
	markets := s.storage.GetAvailable()

	// Фильтруем по ролям пользователя
	var result []*spotv1.Market
	for _, market := range markets {
		if !market.Enabled || market.DeletedAt != nil {
			s.logger.Info("market unavailable",
				zap.String("trace_id", traceID),
				zap.String("market_id", market.ID),
				zap.Bool("enabled", market.Enabled),
				zap.Bool("deleted", market.DeletedAt != nil),
			)
		}
		if market.IsAccessibleForRoles(req.UserRoles) {
			result = append(result, mapDomainMarketToProto(market))
		}
	}

	return &spotv1.ViewMarketsResponse{
		Markets: result,
	}, nil
}

// mapDomainMarketToProto преобразует доменную модель в proto
func mapDomainMarketToProto(m *domain.Market) *spotv1.Market {
	return &spotv1.Market{
		Id:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Enabled:     m.Enabled,
	}
}
