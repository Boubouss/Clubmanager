package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtTokenManager struct {
  secret  []byte
  ttlDays int
}

func NewJwtTokenManager(secret string, ttlDays int) (*jwtTokenManager, error) {
  if len(secret) < 32 {
    return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters (got %d)", len(secret))
  }
  if ttlDays <= 0 {
    ttlDays = 30
  }
  return &jwtTokenManager{secret: []byte(secret), ttlDays: ttlDays}, nil
}

type clubManagerClaims struct {
  UserId string `json:"userId"`
  jwt.RegisteredClaims
}

func (t jwtTokenManager) GenerateToken(id string) (string, error) {
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, clubManagerClaims{
    UserId: id,
    RegisteredClaims: jwt.RegisteredClaims{
      ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(t.ttlDays) * 24 * time.Hour)),
      IssuedAt:  jwt.NewNumericDate(time.Now()),
    },
  })

  return token.SignedString(t.secret)
}

func (t jwtTokenManager) ParseToken(tokenStr string) (string, error) {
  token, err := jwt.ParseWithClaims(tokenStr, &clubManagerClaims{}, func(token *jwt.Token) (any, error) {
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
      return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
    }
    return t.secret, nil
  })

  if err != nil {
    return "", err
  }

  claims, ok := token.Claims.(*clubManagerClaims)
  if !ok || !token.Valid {
    return "", fmt.Errorf("invalid token")
  }

  return claims.UserId, nil
}
