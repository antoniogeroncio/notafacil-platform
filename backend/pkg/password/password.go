// Package password hashes and verifies user passwords using bcrypt.
// Plaintext passwords are never stored (Princípio VI).
package password

import "golang.org/x/crypto/bcrypt"

// cost is the bcrypt work factor. 12 balances security and latency for v1.
const cost = 12

// Hash returns the bcrypt hash of the given plaintext password.
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify reports whether plain matches the previously generated hash.
func Verify(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
