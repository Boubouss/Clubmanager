package services

type AuthService any

type authService struct {}

func NewAuthService() *authService {
  return &authService{}
}
