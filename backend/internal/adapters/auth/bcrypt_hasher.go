package auth

import (
	"golang.org/x/crypto/bcrypt"
)

type bcryptHasher struct {}

func NewBcryptHasher() *bcryptHasher {
  return &bcryptHasher{}
}

func (h bcryptHasher) Hash(pswd string) (string, error) {
  hash, err := bcrypt.GenerateFromPassword([]byte(pswd), bcrypt.DefaultCost)
  return string(hash), err
}
