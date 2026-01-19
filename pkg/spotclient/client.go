package spotclient

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	spotv1 "spotService/gen/pb"
)

type Client struct {
	conn   *grpc.ClientConn
	client spotv1.SpotInstrumentServiceClient
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
		client: spotv1.NewSpotInstrumentServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) ViewMarkets(ctx context.Context, userRoles []spotv1.UserRole) ([]*spotv1.Market, error) {
	resp, err := c.client.ViewMarkets(ctx, &spotv1.ViewMarketsRequest{
		UserRoles: userRoles,
	})
	if err != nil {
		return nil, err
	}
	return resp.Markets, nil
}

func (c *Client) GetClient() spotv1.SpotInstrumentServiceClient {
	return c.client
}
