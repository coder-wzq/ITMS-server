package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Endpoints []string `yaml:"endpoints"`
	Username  string   `yaml:"username,optional"`
	Password  string   `yaml:"password,optional"`
	TTL       int      `yaml:"ttl"`
}

type Registrar struct {
	client     *clientv3.Client
	leaseID    clientv3.LeaseID
	key        string
	instanceID string
	stopCh     chan struct{}
}

func NewClient(cfg *Config) (*clientv3.Client, error) {
	if cfg.TTL == 0 {
		cfg.TTL = 10
	}
	if len(cfg.Endpoints) == 0 {
		cfg.Endpoints = []string{"localhost:2379"}
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		Username:    cfg.Username,
		Password:    cfg.Password,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %w", err)
	}

	log.Printf("[Etcd] connected, endpoints=%v", cfg.Endpoints)
	return cli, nil
}

func NewRegistrar(client *clientv3.Client, serviceName string, ttl int64) *Registrar {
	hostname, _ := os.Hostname()
	instanceID := fmt.Sprintf("%s-%s-%d", hostname, serviceName, os.Getpid())

	return &Registrar{
		client:     client,
		instanceID: instanceID,
		key:        fmt.Sprintf("/itms/%s/%s", serviceName, instanceID),
		stopCh:     make(chan struct{}),
	}
}

// endpointUpdate is the JSON format the etcd naming resolver expects.
type endpointUpdate struct {
	Op       int         `json:"Op"`
	Addr     string      `json:"Addr"`
	Metadata interface{} `json:"Metadata"`
}

func (r *Registrar) Register(ctx context.Context, addr string) error {
	lease, err := r.client.Grant(ctx, 10)
	if err != nil {
		return fmt.Errorf("failed to create lease: %w", err)
	}

	r.leaseID = lease.ID

	val, _ := json.Marshal(endpointUpdate{Op: 0, Addr: addr})
	if _, err := r.client.Put(ctx, r.key, string(val), clientv3.WithLease(lease.ID)); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	log.Printf("[Etcd] service registered: key=%s addr=%s lease=%d", r.key, addr, lease.ID)

	go r.keepAlive()
	return nil
}

func (r *Registrar) keepAlive() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if _, err := r.client.KeepAliveOnce(ctx, r.leaseID); err != nil {
				log.Printf("[Etcd] keepalive failed: %v", err)
			}
			cancel()
		}
	}
}

func (r *Registrar) Deregister(ctx context.Context) error {
	close(r.stopCh)
	_, err := r.client.Delete(ctx, r.key)
	if err != nil {
		return fmt.Errorf("failed to deregister: %w", err)
	}
	log.Printf("[Etcd] service deregistered: key=%s", r.key)
	return nil
}

// DialService creates a gRPC client connection to a service discovered via etcd.
// It queries etcd for the current set of registered endpoints and connects to one.
func DialService(client *clientv3.Client, serviceName string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	prefix := fmt.Sprintf("/itms/%s/", serviceName)
	resp, err := client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("etcd get %s: %w", prefix, err)
	}

	// Collect addresses from etcd entries (both JSON and plain string formats)
	var addrs []string
	for _, kv := range resp.Kvs {
		val := string(kv.Value)
		// Try JSON format first
		var ep endpointUpdate
		if err := json.Unmarshal(kv.Value, &ep); err == nil && ep.Addr != "" {
			addrs = append(addrs, ep.Addr)
		} else if val != "" {
			// Plain string format
			addrs = append(addrs, val)
		}
	}

	if len(addrs) == 0 {
		return nil, fmt.Errorf("no endpoints found for %s in etcd", serviceName)
	}

	log.Printf("[Etcd] discovered %s → %v", serviceName, addrs)

	conn, err := grpc.NewClient(addrs[0],
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addrs[0], err)
	}

	return conn, nil
}
