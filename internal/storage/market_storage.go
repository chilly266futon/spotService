package storage

import (
	"spotService/internal/domain"
	"sync"
	"time"
)

type MarketStorage struct {
	markets map[string]*domain.Market
	mu      sync.RWMutex
}

func NewMarketStorage(markets []*domain.Market) *MarketStorage {
	storage := &MarketStorage{
		markets: make(map[string]*domain.Market),
	}

	for _, market := range markets {
		storage.markets[market.ID] = market
	}

	return storage
}

func (s *MarketStorage) GetByID(id string) (*domain.Market, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	market, exists := s.markets[id]
	if !exists {
		return nil, false
	}
	return market, true
}

func (s *MarketStorage) GetAll() []*domain.Market {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*domain.Market, 0, len(s.markets))
	for _, market := range s.markets {
		result = append(result, market)
	}
	return result
}

func (s *MarketStorage) GetAvailbale() []*domain.Market {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*domain.Market, 0)
	for _, market := range s.markets {
		if market.IsAvailable() {
			result = append(result, market)
		}
	}
	return result
}

func (s *MarketStorage) Add(market *domain.Market) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.markets[market.ID] = market
}

func (s *MarketStorage) Update(market *domain.Market) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.markets[market.ID]; !exists {
		return false
	}

	s.markets[market.ID] = market
	return true
}

func (s *MarketStorage) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	market, exists := s.markets[id]
	if !exists {
		return false
	}

	now := time.Now()
	market.DeletedAt = &now
	return true
}

func (s *MarketStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.markets)
}
