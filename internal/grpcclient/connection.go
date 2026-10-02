package grpcclient

import (
	"context"
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// DefaultMaxMsgSize sets a generous 50MB message limit to handle large responses comfortably.
	DefaultMaxMsgSize = 50 * 1024 * 1024
)

// DialTarget creates a gRPC client connection based on TargetConfig.
func DialTarget(ctx context.Context, cfg TargetConfig) (*grpc.ClientConn, error) {
	if cfg.Target == "" {
		return nil, fmt.Errorf("gRPC target cannot be empty")
	}

	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(DefaultMaxMsgSize),
			grpc.MaxCallSendMsgSize(DefaultMaxMsgSize),
		),
	}

	if cfg.Plaintext {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		}
		if cfg.Authority != "" {
			tlsConfig.ServerName = cfg.Authority
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	if cfg.Authority != "" {
		opts = append(opts, grpc.WithAuthority(cfg.Authority))
	}

	conn, err := grpc.DialContext(ctx, cfg.Target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC target %s: %w", cfg.Target, err)
	}

	return conn, nil
}
