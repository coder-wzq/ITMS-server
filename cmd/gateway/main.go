package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/gateway"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"itms-server/pkg/etcd"
	"itms-server/pkg/jwt"
	"itms-server/pkg/middleware"
	"itms-server/pkg/response"
)

type GatewayConfig struct {
	gateway.GatewayConf
	JWT       JWTConfig       `json:"jwt"`
	CORS      CORSConfig      `json:"cors"`
	RateLimit RateLimitConfig `json:"rateLimit"`
	Etcd      etcd.Config     `json:"etcd"` // global etcd
}

type JWTConfig struct {
	Secret              string `json:"secret"`
	AccessTokenTTLMin   int    `json:"accessTokenTTL"`
	RefreshTokenTTLHour int    `json:"refreshTokenTTL"`
}

type CORSConfig struct {
	AllowedOrigins []string `json:"allowedOrigins"`
	AllowedMethods []string `json:"allowedMethods"`
	AllowedHeaders []string `json:"allowedHeaders"`
	MaxAge         int      `json:"maxAge"`
}

type RateLimitConfig struct {
	Rate  int `json:"rate"`
	Burst int `json:"burst"`
}

// simpleClient wraps a gRPC conn to satisfy zrpc.Client.
type simpleClient struct{ conn *grpc.ClientConn }

func (c *simpleClient) Conn() *grpc.ClientConn { return c.conn }

func main() {
	configPath := flag.String("config", "deploy/configs/config.dev.yaml", "path to config file")
	flag.Parse()

	var c GatewayConfig
	conf.MustLoad(*configPath, &c)

	log.Printf("[Gateway] starting ITMS Gateway on %s:%d", c.Host, c.Port)

	jcfg := &jwt.Config{
		Secret:              c.JWT.Secret,
		AccessTokenTTLMin:   c.JWT.AccessTokenTTLMin,
		RefreshTokenTTLHour: c.JWT.RefreshTokenTTLHour,
	}
	if err := jcfg.Validate(); err != nil {
		log.Printf("[Gateway] WARNING: JWT %v", err)
	}
	middleware.InitJWTMiddleware(jcfg)
	middleware.InitCORS(&middleware.CORSConfig{
		AllowedOrigins: c.CORS.AllowedOrigins,
		AllowedMethods: c.CORS.AllowedMethods,
		AllowedHeaders: c.CORS.AllowedHeaders,
		MaxAge:         c.CORS.MaxAge,
	})

	// etcd-aware dialer: uses global etcd config + per-upstream Etcd.Key
	etcdCli, etcdErr := etcd.NewClient(&c.Etcd)

	customDialer := func(conf zrpc.RpcClientConf) zrpc.Client {
		// If Etcd discovery is configured and etcd is available
		if conf.Etcd.Key != "" && etcdErr == nil {
			serviceName := conf.Etcd.Key
			// Strip /itms/ prefix if present
			if len(serviceName) > 6 && serviceName[:6] == "/itms/" {
				serviceName = serviceName[6:]
			}
			conn, err := etcd.DialService(etcdCli, serviceName)
			if err != nil {
				log.Fatalf("[Gateway] etcd dial %s: %v", serviceName, err)
			}
			return &simpleClient{conn}
		}

		// Fallback to go-zero default dialer
		return zrpc.MustNewClient(conf)
	}

	gw := gateway.MustNewServer(c.GatewayConf,
		gateway.WithDialer(customDialer),
		gateway.WithMiddleware(middleware.CORS),
		gateway.WithMiddleware(middleware.JWTAuth),
		gateway.WithMiddleware(middleware.OperationLogger),
		gateway.WithMiddleware(middleware.WrapResponse),
	)

	gw.Server.Use(middleware.Recovery)
	gw.Server.Use(middleware.RequestLogger)
	gw.Server.Use(middleware.RateLimit(c.RateLimit.Rate, c.RateLimit.Burst))

	gw.Server.AddRoute(rest.Route{Method: http.MethodGet, Path: "/health", Handler: healthHandler})

	// svc-user health proxy via gRPC health check
	if etcdErr == nil {
		conn, err := etcd.DialService(etcdCli, "svc-user")
		if err == nil {
			hc := grpc_health_v1.NewHealthClient(conn)
			gw.Server.AddRoute(rest.Route{Method: http.MethodGet, Path: "/api/user/health",
				Handler: func(w http.ResponseWriter, r *http.Request) {
					resp, err := hc.Check(r.Context(), &grpc_health_v1.HealthCheckRequest{})
					if err != nil {
						response.WriteJSON(w, http.StatusServiceUnavailable,
							response.Error(response.CodeAuthServerErr, err.Error()))
						return
					}
					response.WriteJSON(w, http.StatusOK, response.Success(map[string]string{
						"status": resp.Status.String(),
					}))
				},
			})
		}
	}

	log.Println("[Gateway] listening...")
	gw.Start()
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	response.WriteJSON(w, http.StatusOK, response.Success(map[string]string{"status": "healthy"}))
}
