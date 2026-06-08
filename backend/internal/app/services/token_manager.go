package services

type TokenManager interface {
  GenerateToken(string) (string, error)
  ParseToken(string) (string, error)
}
