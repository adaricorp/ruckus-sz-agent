package main

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana/loki/v3/pkg/logproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type lokiWriteClient struct {
	client  logproto.PusherClient
	timeout time.Duration
}

func newLokiWriteClient(server string, timeout time.Duration) (lokiWriteClient, error) {
	grpcClient, err := grpc.NewClient(*lokiServer,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return lokiWriteClient{}, fmt.Errorf(
			"failed to create gRPC client for server %s: %w",
			server,
			err,
		)
	}

	return lokiWriteClient{
		client:  logproto.NewPusherClient(grpcClient),
		timeout: timeout,
	}, nil
}

func (l lokiWriteClient) write(entry logEntry) error {
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	_, err := loki.client.Push(ctx, &logproto.PushRequest{
		Streams: []logproto.Stream{entry.marshal()},
	})
	if err != nil {
		return err
	}

	return nil
}
