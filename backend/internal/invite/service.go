package invite

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/notafacil/platform/backend/internal/user"
	"github.com/notafacil/platform/backend/pkg/email"
	"github.com/notafacil/platform/backend/pkg/middleware"
	"github.com/notafacil/platform/backend/pkg/token"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ErrValidation indicates invalid input (bad e-mail or role).
var ErrValidation = errors.New("invite: validation error")

// ErrEmailConflict indicates the e-mail already belongs to an account on the
// platform (global uniqueness, FR-004).
var ErrEmailConflict = errors.New("invite: email already exists")

// CreateInput is the data required to invite a team member.
type CreateInput struct {
	Email string
	Role  user.Role
}

// Service implements the team-invitation business rules (US1).
type Service struct {
	users      user.Repository
	invites    Repository
	sender     email.Sender
	inviteTTL  time.Duration
	appBaseURL string
	now        func() time.Time
}

// NewService builds an invite Service with its injected dependencies.
func NewService(users user.Repository, invites Repository, sender email.Sender, inviteTTL time.Duration, appBaseURL string) *Service {
	return &Service{
		users:      users,
		invites:    invites,
		sender:     sender,
		inviteTTL:  inviteTTL,
		appBaseURL: appBaseURL,
		now:        time.Now,
	}
}

// CreateInvite invites an e-mail to the authenticated Admin's tenant. It creates
// (or reuses) a pending user, issues a 48h token and sends the activation
// e-mail. The raw token leaves only via the e-mail; it is never persisted nor
// returned (Princípio VI).
func (s *Service) CreateInvite(ctx context.Context, in CreateInput) (*Invite, error) {
	if !validEmail(in.Email) {
		return nil, fmt.Errorf("%w: e-mail inválido", ErrValidation)
	}
	if !user.ValidRole(in.Role) {
		return nil, fmt.Errorf("%w: papel inválido", ErrValidation)
	}

	tenantID, err := middleware.TenantObjectID(ctx)
	if err != nil {
		return nil, err
	}
	identity, _ := middleware.IdentityFrom(ctx)
	invitedBy, _ := primitive.ObjectIDFromHex(identity.UserID)

	existing, err := s.users.FindByEmail(ctx, in.Email)
	switch {
	case err == nil:
		// E-mail already exists: only a pending user in the same tenant may be
		// renewed; anything else is a conflict.
		if existing.Status == user.StatusAtivo || existing.TenantID != tenantID {
			return nil, ErrEmailConflict
		}
		if err := s.invites.DeleteByEmail(ctx, in.Email); err != nil {
			return nil, err
		}
	case errors.Is(err, user.ErrNotFound):
		newUser := &user.User{
			TenantID: tenantID,
			Email:    in.Email,
			Role:     in.Role,
			Status:   user.StatusPendente,
			CriadoEm: s.now(),
		}
		if err := s.users.Create(ctx, newUser); err != nil {
			if errors.Is(err, user.ErrEmailTaken) {
				return nil, ErrEmailConflict
			}
			return nil, err
		}
	default:
		return nil, err
	}

	raw, hash, err := token.GenerateInviteToken()
	if err != nil {
		return nil, err
	}
	inv := &Invite{
		TenantID:        tenantID,
		Email:           in.Email,
		Role:            in.Role,
		TokenHash:       hash,
		ExpiresAt:       s.now().Add(s.inviteTTL),
		InvitedByUserID: invitedBy,
	}
	if err := s.invites.Create(ctx, inv); err != nil {
		return nil, err
	}

	if err := s.sender.Send(ctx, buildInviteEmail(in.Email, s.appBaseURL, raw)); err != nil {
		return nil, err
	}
	return inv, nil
}

func validEmail(addr string) bool {
	parsed, err := mail.ParseAddress(addr)
	return err == nil && parsed.Address == addr
}

func buildInviteEmail(to, baseURL, rawToken string) email.Message {
	link := fmt.Sprintf("%s/accept-invite/%s", baseURL, rawToken)
	html := fmt.Sprintf(
		`<p>Você foi convidado para a NotaFácil.</p>`+
			`<p>Para ativar sua conta e definir sua senha, acesse:</p>`+
			`<p><a href="%s">Ativar minha conta</a></p>`+
			`<p>Este link expira em 48 horas.</p>`, link)
	text := fmt.Sprintf("Ative sua conta na NotaFácil acessando: %s (o link expira em 48 horas).", link)
	return email.Message{
		To:       to,
		Subject:  "Ative sua conta na NotaFácil",
		HTMLBody: html,
		TextBody: text,
	}
}
