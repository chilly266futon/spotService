package spotclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	spotpb "github.com/chilly266futon/exchange-service-contracts/gen/pb/spot"
)

type Client struct {
	conn   *grpc.ClientConn
	client spotpb.SpotInstrumentServiceClient
}

type Config struct {
	Address string
}

func New(cfg Config) (*Client, error) {
	conn, err := grpc.NewClient(
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to spot service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: spotpb.NewSpotInstrumentServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) ViewMarkets(ctx context.Context, userRoles []spotpb.UserRole) ([]*spotpb.Market, error) {
	resp, err := c.client.ViewMarkets(ctx, &spotpb.ViewMarketsRequest{
		UserRoles: userRoles,
	})
	if err != nil {
		return nil, err
	}
	return resp.Markets, nil
}

func (c *Client) GetClient() spotpb.SpotInstrumentServiceClient {
	return c.client
}
