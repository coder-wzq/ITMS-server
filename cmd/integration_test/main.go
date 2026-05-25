package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"itms-server/pkg/db"
	"itms-server/pkg/jwt"
	"itms-server/pkg/redis"
)

func main() {
	log.Println("=== ITMS Integration Test ===")

	// 1. Snowflake ID generation
	log.Println("\n--- Test 1: Snowflake ID ---")
	sf := db.GetSnowflake()
	id1 := sf.NextID()
	id2 := sf.NextID()
	fmt.Printf("ID1: %d, ID2: %d, Unique: %v\n", id1, id2, id1 != id2)

	// 2. PostgreSQL connection
	log.Println("\n--- Test 2: PostgreSQL ---")
	pgCfg := &db.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "itms",
		Password: "itms@2026",
		DBName:   "itms",
		SSLMode:  "disable",
	}
	pgDB, err := db.NewPostgres(pgCfg)
	if err != nil {
		log.Printf("PostgreSQL connection FAILED: %v", err)
	} else {
		log.Println("PostgreSQL connection OK")
		sqlDB, _ := pgDB.DB()
		stats := sqlDB.Stats()
		log.Printf("  Open: %d, InUse: %d, Idle: %d", stats.OpenConnections, stats.InUse, stats.Idle)

		// Test query
		var version string
		pgDB.Raw("SELECT version()").Scan(&version)
		log.Printf("  Version: %s", version[:60])
	}

	// 3. Redis connection
	log.Println("\n--- Test 3: Redis ---")
	redisCfg := &redis.Config{
		Addr:     "localhost:6379",
		Password: "itms@2026",
		DB:       0,
		PoolSize: 10,
	}
	rdb, err := redis.NewClient(redisCfg)
	if err != nil {
		log.Printf("Redis connection FAILED: %v", err)
	} else {
		log.Println("Redis connection OK")
		ctx := context.Background()

		// Test Set/Get
		testKey := "itms:test:key"
		err = rdb.Set(ctx, testKey, "hello-itms", 10*time.Second)
		if err != nil {
			log.Printf("Redis Set FAILED: %v", err)
		} else {
			val, err := rdb.Get(ctx, testKey)
			if err != nil {
				log.Printf("Redis Get FAILED: %v", err)
			} else {
				log.Printf("Set/Get OK: %s = %s", testKey, val)
			}
			rdb.Del(ctx, testKey)
		}
	}

	// 4. JWT token generation
	log.Println("\n--- Test 4: JWT ---")
	jwtCfg := &jwt.Config{
		Secret:              "itms-test-jwt-secret-min-32-chars-long!!",
		AccessTokenTTLMin:   120,
		RefreshTokenTTLHour: 168,
	}
	pair, err := jwt.GenerateTokenPair(jwtCfg, 12345, "testuser", "测试用户", []string{"admin", "operator"}, "uuid-token-001")
	if err != nil {
		log.Printf("JWT generation FAILED: %v", err)
	} else {
		log.Println("JWT generation OK")
		log.Printf("  AccessToken: %s...", pair.AccessToken[:50])
		log.Printf("  RefreshToken: %s...", pair.RefreshToken[:50])

		// Test parsing
		claims, err := jwt.ParseToken(jwtCfg.Secret, pair.AccessToken)
		if err != nil {
			log.Printf("JWT parse FAILED: %v", err)
		} else {
			log.Printf("JWT parse OK: userId=%d, username=%s, roles=%v",
				claims.UserID, claims.Username, claims.Roles)
		}
	}

	// 5. Etcd connection (optional, may not have permissions)
	log.Println("\n--- Test 5: Etcd ---")
	log.Println("Etcd test skipped (requires running etcd instance)")

	log.Println("\n=== All tests completed ===")
}
