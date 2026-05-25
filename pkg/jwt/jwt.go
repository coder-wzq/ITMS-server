package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   int64    `json:"userId"`
	Username string   `json:"username"`
	RealName string   `json:"realName"`
	Roles    []string `json:"roles"`
	TokenID  string   `json:"tokenId,omitempty"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type Config struct {
	Secret              string `json:"secret" env:"JWT_SECRET"`
	AccessTokenTTLMin   int    `json:"accessTokenTTL"`   // minutes
	RefreshTokenTTLHour int    `json:"refreshTokenTTL"`  // hours
}

func (c *Config) AccessTokenTTL() time.Duration {
	if c.AccessTokenTTLMin <= 0 {
		c.AccessTokenTTLMin = 120 // 2h default
	}
	return time.Duration(c.AccessTokenTTLMin) * time.Minute
}

func (c *Config) RefreshTokenTTL() time.Duration {
	if c.RefreshTokenTTLHour <= 0 {
		c.RefreshTokenTTLHour = 168 // 7d default
	}
	return time.Duration(c.RefreshTokenTTLHour) * time.Hour
}

func (c *Config) Validate() error {
	if c.Secret == "" {
		return fmt.Errorf("jwt secret must not be empty")
	}
	if len(c.Secret) < 32 {
		return fmt.Errorf("jwt secret must be at least 32 characters")
	}
	return nil
}

func GenerateTokenPair(cfg *Config, userID int64, username, realName string, roles []string, tokenID string) (*TokenPair, error) {
	now := time.Now()

	accessClaims := &Claims{
		UserID:   userID,
		Username: username,
		RealName: realName,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.AccessTokenTTL())),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "itms",
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshClaims := &Claims{
		UserID:  userID,
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.RefreshTokenTTL())),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "itms",
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(cfg.AccessTokenTTL().Seconds()),
	}, nil
}

func ParseToken(secret, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
