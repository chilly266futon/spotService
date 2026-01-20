package service

import (
	"context"

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
}

// NewService создает новый сервис
func NewService(storage *storage.MarketStorage) *Service {
	return &Service{
		storage: storage,
	}
}

// ViewMarkets возвращает список доступных рынков для указанных ролей
func (s *Service) ViewMarkets(ctx context.Context, req *spotv1.ViewMarketsRequest) (*spotv1.ViewMarketsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Получаем все доступные рынки
	markets := s.storage.GetAvailable()

	// Фильтруем по ролям пользователя
	var result []*spotv1.Market
	for _, market := range markets {
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
