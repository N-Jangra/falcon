package auth

import "golang.org/x/crypto/bcrypt"

// Authenticator will handle password verification and related state.
type Authenticator struct {
	passwordHash string
}

// New creates an authenticator with a hashed password.
func New(hash string) *Authenticator {
	return &Authenticator{passwordHash: hash}
}

// Authenticate validates the provided password against the stored bcrypt hash.
func (a *Authenticator) Authenticate(password string) bool {
	if a.passwordHash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(a.passwordHash), []byte(password)) == nil
}

// HashPassword provides a helper to generate a bcrypt hash for configuration or tests.
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}
