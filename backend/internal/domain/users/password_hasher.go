package users

type PasswordHasher interface {
  Hash(string) (string, error)
  Compare(plain, hash string) error
}
