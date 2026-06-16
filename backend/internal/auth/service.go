// Package auth implements authentication and account activation (US2):
// login, logout, session identity (/me) and invite acceptance.
package auth

import (
	"context"
	"errors"
	"time"
	"unicode"

	"github.com/notafacil/platform/backend/internal/invite"
	"github.com/notafacil/platform/backend/internal/user"
	"github.com/notafacil/platform/backend/pkg/password"
	"github.com/notafacil/platform/backend/pkg/token"
)

// Domain errors mapped to HTTP statuses by the handler.
var (
	ErrWeakPassword       = errors.New("auth: weak password")
	ErrInviteNotFound     = errors.New("auth: invite not found")
	ErrInviteExpired      = errors.New("auth: invite expired or already used")
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
)

// TokenIssuer issues session tokens.
type TokenIssuer interface {
	Issue(c token.SessionClaims) (string, error)
}

// Service implements authentication business rules.
type Service struct {
	users   user.Repository
	invites invite.Repository
	tokens  TokenIssuer
	now     func() time.Time
}

// NewService builds an auth Service with injected dependencies.
func NewService(users user.Repository, invites invite.Repository, tokens TokenIssuer) *Service {
	return &Service{users: users, invites: invites, tokens: tokens, now: time.Now}
}

// AcceptInvite activates a pending account from a valid invite token. It applies
// the password policy, hashes the password, marks the invite used and returns
// the activated user plus a fresh session token.
func (s *Service) AcceptInvite(ctx context.Context, rawToken, nome, senha string) (*user.User, string, error) {
	if !validPassword(senha) {
		return nil, "", ErrWeakPassword
	}

	inv, err := s.invites.FindByTokenHash(ctx, token.HashInviteToken(rawToken))
	if errors.Is(err, invite.ErrNotFound) {
		return nil, "", ErrInviteNotFound
	}
	if err != nil {
		return nil, "", err
	}
	if !inv.IsUsable(s.now()) {
		return nil, "", ErrInviteExpired
	}

	u, err := s.users.FindByEmail(ctx, inv.Email)
	if errors.Is(err, user.ErrNotFound) {
		return nil, "", ErrInviteNotFound
	}
	if err != nil {
		return nil, "", err
	}

	hash, err := password.Hash(senha)
	if err != nil {
		return nil, "", err
	}
	now := s.now()
	if err := s.users.Activate(ctx, u.ID.Hex(), nome, hash, now); err != nil {
		return nil, "", err
	}
	if err := s.invites.MarkUsed(ctx, inv.ID, now); err != nil {
		return nil, "", err
	}

	u.Nome = nome
	u.Status = user.StatusAtivo
	u.AtivadoEm = &now
	u.SenhaHash = ""

	session, err := s.issueSession(u)
	if err != nil {
		return nil, "", err
	}
	return u, session, nil
}

// Login authenticates an active user by e-mail and password.
func (s *Service) Login(ctx context.Context, email, senha string) (*user.User, string, error) {
	u, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, user.ErrNotFound) {
		return nil, "", ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", err
	}
	if u.Status != user.StatusAtivo || !password.Verify(u.SenhaHash, senha) {
		return nil, "", ErrInvalidCredentials
	}

	u.SenhaHash = ""
	session, err := s.issueSession(u)
	if err != nil {
		return nil, "", err
	}
	return u, session, nil
}

// Me returns the user of the authenticated context, scoped to its tenant.
func (s *Service) Me(ctx context.Context, userID string) (*user.User, error) {
	u, err := s.users.FindByIDScoped(ctx, userID)
	if err != nil {
		return nil, err
	}
	u.SenhaHash = ""
	return u, nil
}

func (s *Service) issueSession(u *user.User) (string, error) {
	return s.tokens.Issue(token.SessionClaims{
		UserID:   u.ID.Hex(),
		TenantID: u.TenantID.Hex(),
		Role:     string(u.Role),
	})
}

// validPassword enforces the v1 policy: at least 8 chars, with letters and
// digits (research.md §5).
func validPassword(p string) bool {
	if len(p) < 8 {
		return false
	}
	var hasLetter, hasDigit bool
	for _, r := range p {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}
