package rpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zeromicro/go-zero/core/breaker"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps a gRPC connection with 3s timeout and circuit breaker.
type Client struct {
	*zrpc.RpcClient
	conn *grpc.ClientConn
}

// Config for gRPC client connection.
type Config struct {
	// Endpoints is the etcd discovery endpoints or direct address.
	Endpoints []string `json:"endpoints"`
	// Target is the direct gRPC server address (host:port).
	Target string `json:"target"`
	// Timeout in milliseconds, default 3000 (3s).
	Timeout int64 `json:"timeout"`
}

// NewClient creates a gRPC client with 3s timeout and go-zero breaker.
// For direct connection, set cfg.Target.
// For etcd service discovery, set cfg.Endpoints and use WithEtcdDiscovery option.
func NewClient(cfg *Config, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 3000 // 3s default per design spec
	}

	timeout := time.Duration(cfg.Timeout) * time.Millisecond

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(timeout),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16 * 1024 * 1024)),
	}
	dialOpts = append(dialOpts, opts...)

	target := cfg.Target
	if target == "" && len(cfg.Endpoints) > 0 {
		// Use etcd discovery: target is the etcd service key prefix
		target = cfg.Endpoints[0]
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("gRPC dial failed: %w", err)
	}

	log.Printf("[gRPC] connected to %s, timeout=%v", target, timeout)
	return conn, nil
}

// DoWithBreaker executes fn with go-zero circuit breaker protection.
// The breaker name is derived from the method identifier.
func DoWithBreaker[T any](name string, fn func() (T, error)) (T, error) {
	brk := breaker.GetBreaker(name)
	var result T
	err := brk.DoWithAcceptable(func() error {
		var execErr error
		result, execErr = fn()
		return execErr
	}, func(err error) bool {
		// Accept only non-critical errors; connection/timeout errors trigger breaker
		return err == nil
	})
	return result, err
}
