package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type PostgresConfig struct {
	Host     string `json:"host" env:"PG_HOST"`
	Port     int    `json:"port" env:"PG_PORT"`
	User     string `json:"user" env:"PG_USER"`
	Password string `json:"password" env:"PG_PASSWORD"`
	DBName   string `json:"dbName" env:"PG_DBNAME"`
	SSLMode  string `json:"sslMode" env:"PG_SSLMODE"`

	MaxOpenConn    int           `json:"maxOpenConn"`
	MaxIdleConn    int           `json:"maxIdleConn"`
	MaxLifetimeMin int           `json:"maxLifetimeMin"`
	MaxLifetime    time.Duration `json:"-"`
	LogLevel       int           `json:"logLevel"`
}

func (c *PostgresConfig) DSN() string {
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func NewPostgres(cfg *PostgresConfig) (*gorm.DB, error) {
	if cfg.MaxOpenConn == 0 {
		cfg.MaxOpenConn = 100
	}
	if cfg.MaxIdleConn == 0 {
		cfg.MaxIdleConn = 25
	}
	if cfg.MaxLifetimeMin == 0 {
		cfg.MaxLifetimeMin = 5
	}
	cfg.MaxLifetime = time.Duration(cfg.MaxLifetimeMin) * time.Minute

	logLevel := logger.Warn
	switch cfg.LogLevel {
	case 1:
		logLevel = logger.Silent
	case 2:
		logLevel = logger.Error
	case 3:
		logLevel = logger.Warn
	case 4:
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger:      logger.Default.LogMode(logLevel),
		PrepareStmt: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	log.Printf("[PostgreSQL] connected successfully, maxOpen=%d maxIdle=%d lifetime=%v",
		cfg.MaxOpenConn, cfg.MaxIdleConn, cfg.MaxLifetime)
	return db, nil
}
