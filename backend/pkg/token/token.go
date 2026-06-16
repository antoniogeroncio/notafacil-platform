// Package token issues and validates session JWTs and invite tokens.
//
// Session tokens are short-lived HS256 JWTs carrying the authenticated
// identity (userId, tenantId, role). Invite tokens are high-entropy random
// strings; only their SHA-256 hash is ever persisted (Princípio VI).
package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SessionClaims is the authenticated identity carried by a session token.
type SessionClaims struct {
	UserID   string
	TenantID string
	Role     string
}

// Manager issues and parses session tokens with a fixed secret and TTL.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager builds a session token Manager.
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

// Issue returns a signed session token for the given claims.
func (m *Manager) Issue(c SessionClaims) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  c.UserID,
		"tid":  c.TenantID,
		"role": c.Role,
		"iat":  now.Unix(),
		"exp":  now.Add(m.ttl).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse validates a session token and returns its claims.
func (m *Manager) Parse(raw string) (SessionClaims, error) {
	parsed, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return SessionClaims{}, err
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return SessionClaims{}, errors.New("token: invalid claims")
	}
	return SessionClaims{
		UserID:   stringClaim(claims, "sub"),
		TenantID: stringClaim(claims, "tid"),
		Role:     stringClaim(claims, "role"),
	}, nil
}

func stringClaim(c jwt.MapClaims, key string) string {
	if v, ok := c[key].(string); ok {
		return v
	}
	return ""
}

// inviteTokenBytes is the entropy of an invite token (256 bits).
const inviteTokenBytes = 32

// GenerateInviteToken returns a random invite token and its SHA-256 hash.
// The raw value is sent to the user; only the hash is persisted.
func GenerateInviteToken() (raw string, hash string, err error) {
	buf := make([]byte, inviteTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(buf)
	return raw, HashInviteToken(raw), nil
}

// HashInviteToken returns the SHA-256 hex digest of a raw invite token.
func HashInviteToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
