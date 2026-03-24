package users

type PasswordHasher interface {
  Hash(string) (string, error)
}
