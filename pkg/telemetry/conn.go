package telemetry

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Collector struct {
	Conn *grpc.ClientConn
}

func (c *Collector) ConnectGRPC(ctx context.Context, url string) (*Collector, error) {
	conn, err := grpc.DialContext(ctx, url,
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return c, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	c.Conn = conn

	return c, nil
}
