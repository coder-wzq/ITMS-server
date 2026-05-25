package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"itms-server/pkg/config"
	"itms-server/pkg/db"
	"itms-server/pkg/etcd"
	"itms-server/pkg/redis"

	userpb "itms-server/api/proto/user"
	svcuser "itms-server/services/svc-user"
)

func main() {
	configPath := flag.String("config", "deploy/configs/config.dev.yaml", "path to config file")
	flag.Parse()

	cfg := config.MustLoad(*configPath)

	// Init infrastructure
	gdb, err := db.NewPostgres(&cfg.Postgres)
	if err != nil {
		log.Fatalf("[svc-user] Postgres: %v", err)
	}
	log.Printf("[svc-user] Postgres connected")
	_ = db.GetSnowflake()

	rdb, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("[svc-user] Redis: %v", err)
	}
	log.Printf("[svc-user] Redis connected")

	// Create gRPC server implementation
	grpcImpl := svcuser.NewServer(gdb, rdb)
	srv := grpc.NewServer()
	userpb.RegisterUserServiceServer(srv, grpcImpl)

	// gRPC standard health check
	hs := health.NewServer()
	hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(srv, hs)

	// Listen
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("[svc-user] listen %s: %v", addr, err)
	}

	// Resolve external address for etcd registration
	_, port, _ := net.SplitHostPort(lis.Addr().String())
	regAddr := fmt.Sprintf("%s:%s", getLocalIP(), port)
	log.Printf("[svc-user] gRPC listening on %s", lis.Addr().String())

	// etcd registration
	var reg *etcd.Registrar
	etcdCli, err := etcd.NewClient(&etcd.Config{
		Endpoints: cfg.Etcd.Endpoints, Username: cfg.Etcd.Username,
		Password: cfg.Etcd.Password, TTL: cfg.Etcd.TTL,
	})
	if err != nil {
		log.Printf("[svc-user] etcd: %v (running without service discovery)", err)
	} else {
		reg = etcd.NewRegistrar(etcdCli, "svc-user", int64(cfg.Etcd.TTL))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := reg.Register(ctx, regAddr); err != nil {
			log.Printf("[svc-user] etcd register: %v", err)
		}
		cancel()
	}

	// Start serving
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("[svc-user] gRPC serve: %v", err)
		}
	}()

	log.Printf("[svc-user] started on %s", regAddr)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[svc-user] shutting down...")
	if reg != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		reg.Deregister(ctx)
		cancel()
	}
	srv.GracefulStop()
	log.Println("[svc-user] stopped")
}

func getLocalIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "127.0.0.1"
}
