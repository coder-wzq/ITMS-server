package config

import (
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/zeromicro/go-zero/core/conf"

	"itms-server/pkg/db"
	jwtpkg "itms-server/pkg/jwt"
	"itms-server/pkg/redis"
)

type AppConfig struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

type LogConfig struct {
	Level string `json:"level"`
	File  string `json:"file"`
	Mode  string `json:"mode"`
}

type EtcdConfig struct {
	Endpoints []string `json:"endpoints"`
	Username  string   `json:"username,optional"`
	Password  string   `json:"password,optional"`
	TTL       int      `json:"ttl"`
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

type GatewayConfig struct {
	HTTPPort int `json:"httpPort"`
	GRPCPort int `json:"grpcPort"`
	Timeout  int `json:"timeout"`
}

type Config struct {
	App       AppConfig         `json:"app"`
	Postgres  db.PostgresConfig `json:"postgres"`
	Redis     redis.Config      `json:"redis"`
	JWT       jwtpkg.Config     `json:"jwt"`
	Etcd      EtcdConfig        `json:"etcd"`
	Log       LogConfig         `json:"log"`
	CORS      CORSConfig        `json:"cors"`
	RateLimit RateLimitConfig   `json:"rateLimit"`
	Gateway   GatewayConfig     `json:"gateway"`
}

var defaultConfig = &Config{
	App: AppConfig{
		Name: "itms-server",
		Host: "0.0.0.0",
		Port: 8080,
	},
	Log: LogConfig{
		Level: "info",
		File:  "logs/app.log",
		Mode:  "console",
	},
	CORS: CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type", "X-Request-ID"},
		MaxAge:         3600,
	},
	RateLimit: RateLimitConfig{
		Rate:  100,
		Burst: 200,
	},
	Gateway: GatewayConfig{
		HTTPPort: 8080,
		GRPCPort: 9090,
		Timeout:  3000,
	},
}

func MustLoad(path string) *Config {
	cfg := &Config{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("[Config] file %s not found, using defaults", path)
		*cfg = *defaultConfig
	} else {
		if err := conf.Load(path, cfg, conf.UseEnv()); err != nil {
			log.Printf("[Config] failed to load %s: %v, using defaults", path, err)
			*cfg = *defaultConfig
		}
	}

	cfg.applyEnvOverrides()
	cfg.setDefaults()

	// Hot reload via fsnotify
	go watchConfig(path)

	return cfg
}

func watchConfig(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()

	if err := watcher.Add(path); err != nil {
		return
	}

	for {
		select {
		case ev, ok := <-watcher.Events:
			if !ok {
				return
			}
			if ev.Has(fsnotify.Write) || ev.Has(fsnotify.Create) {
				log.Println("[Config] file changed, reloading...")
				newCfg := &Config{}
				if err := conf.Load(path, newCfg, conf.UseEnv()); err != nil {
					log.Printf("[Config] reload error: %v", err)
					continue
				}
				newCfg.applyEnvOverrides()
				newCfg.setDefaults()
				log.Println("[Config] reloaded successfully")
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("[Config] watcher error: %v", err)
		}
	}
}

func (c *Config) applyEnvOverrides() {
	c.Postgres.Host = coalesceEnv("PG_HOST", c.Postgres.Host)
	c.Postgres.User = coalesceEnv("PG_USER", c.Postgres.User)
	c.Postgres.Password = coalesceEnv("PG_PASSWORD", c.Postgres.Password)
	c.Postgres.DBName = coalesceEnv("PG_DBNAME", c.Postgres.DBName)
	c.Redis.Addr = coalesceEnv("REDIS_ADDR", c.Redis.Addr)
	c.Redis.Password = coalesceEnv("REDIS_PASSWORD", c.Redis.Password)
	c.JWT.Secret = coalesceEnv("JWT_SECRET", c.JWT.Secret)
	if v := os.Getenv("ETCD_ENDPOINTS"); v != "" {
		c.Etcd.Endpoints = strings.Split(v, ",")
	}
}

func (c *Config) setDefaults() {
	if c.App.Port == 0 {
		c.App.Port = 8080
	}
	if c.Postgres.MaxOpenConn == 0 {
		c.Postgres.MaxOpenConn = 100
	}
	if c.Postgres.MaxIdleConn == 0 {
		c.Postgres.MaxIdleConn = 25
	}
	if c.Postgres.MaxLifetimeMin == 0 {
		c.Postgres.MaxLifetimeMin = 5
	}
	if c.Redis.Addr == "" {
		c.Redis.Addr = "localhost:6379"
	}
	if c.Gateway.Timeout == 0 {
		c.Gateway.Timeout = 3000
	}
	if c.Etcd.TTL == 0 {
		c.Etcd.TTL = 10
	}
	if c.JWT.AccessTokenTTLMin == 0 {
		c.JWT.AccessTokenTTLMin = 120
	}
	if c.JWT.RefreshTokenTTLHour == 0 {
		c.JWT.RefreshTokenTTLHour = 168
	}
}

func coalesceEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
