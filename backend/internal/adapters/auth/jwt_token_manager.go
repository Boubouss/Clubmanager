package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtTokenManager struct {}

func NewJwtTokenManager() *jwtTokenManager {
  return &jwtTokenManager{}
}

func (t jwtTokenManager) GenerateToken(id string) (string, error) {
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "userId": id,
    "ttl": time.Now().Add(time.Hour * 24 * 30).Unix(),
  })

  tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

  return tokenStr, err
}
