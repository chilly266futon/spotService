package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	spotv1 "github.com/chilly266futon/spotService/gen/pb"
	"github.com/chilly266futon/spotService/internal/domain"
	"github.com/chilly266futon/spotService/internal/storage"
)

func TestViewMarkets_ReturnsOnlyActiveMarkets(t *testing.T) {
	now := time.Now()

	markets := []*domain.Market{
		{
			ID:      "active",
			Name:    "Active market",
			Enabled: true,
		},
		{
			ID:        "deleted",
			Enabled:   true,
			DeletedAt: &now,
		},
		{
			ID:      "disabled",
			Enabled: false,
		},
	}

	svc := NewService(storage.NewMarketStorage(markets))

	resp, err := svc.ViewMarkets(context.Background(), &spotv1.ViewMarketsRequest{})

	require.NoError(t, err)
	assert.Len(t, resp.Markets, 1)
	assert.Equal(t, "active", resp.Markets[0].Id)
}

func TestViewMarkets_FilterByRoles(t *testing.T) {
	markets := []*domain.Market{
		{
			ID:           "common-market",
			Name:         "Common Market",
			Enabled:      true,
			AllowedRoles: []spotv1.UserRole{spotv1.UserRole_USER_ROLE_COMMON},
		},
		{
			ID:           "premium-market",
			Name:         "Premium Market",
			Enabled:      true,
			AllowedRoles: []spotv1.UserRole{spotv1.UserRole_USER_ROLE_PREMIUM},
		},
		{
			ID:           "public-market",
			Name:         "Public Market",
			Enabled:      true,
			AllowedRoles: []spotv1.UserRole{}, // Доступно всем
		},
	}

	svc := NewService(storage.NewMarketStorage(markets))

	t.Run("common user sees common and public markets", func(t *testing.T) {
		resp, err := svc.ViewMarkets(context.Background(), &spotv1.ViewMarketsRequest{
			UserRoles: []spotv1.UserRole{spotv1.UserRole_USER_ROLE_COMMON},
		})

		require.NoError(t, err)
		assert.Len(t, resp.Markets, 2)

		ids := make(map[string]bool)
		for _, m := range resp.Markets {
			ids[m.Id] = true
		}

		assert.True(t, ids["common-market"])
		assert.True(t, ids["public-market"])
		assert.False(t, ids["premium-market"])
	})

	t.Run("premium user sees all markets", func(t *testing.T) {
		resp, err := svc.ViewMarkets(context.Background(), &spotv1.ViewMarketsRequest{
			UserRoles: []spotv1.UserRole{
				spotv1.UserRole_USER_ROLE_COMMON,
				spotv1.UserRole_USER_ROLE_PREMIUM,
			},
		})

		require.NoError(t, err)
		assert.Len(t, resp.Markets, 3)
	})

	t.Run("user without roles sees only public market", func(t *testing.T) {
		resp, err := svc.ViewMarkets(context.Background(), &spotv1.ViewMarketsRequest{
			UserRoles: []spotv1.UserRole{},
		})

		require.NoError(t, err)
		assert.Len(t, resp.Markets, 1)
		assert.Equal(t, "public-market", resp.Markets[0].Id)
	})
}

func TestViewMarkets_InvalidRequest(t *testing.T) {
	svc := NewService(storage.NewMarketStorage([]*domain.Market{}))

	resp, err := svc.ViewMarkets(context.Background(), nil)

	assert.Nil(t, resp)
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}
