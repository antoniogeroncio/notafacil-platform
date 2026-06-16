package auth

import (
	"context"
	"testing"
	"time"

	"github.com/notafacil/platform/backend/internal/invite"
	"github.com/notafacil/platform/backend/internal/user"
	"github.com/notafacil/platform/backend/pkg/middleware"
	"github.com/notafacil/platform/backend/pkg/password"
	"github.com/notafacil/platform/backend/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- fakes ---

type fakeUserRepo struct {
	byEmail   map[string]*user.User
	activated map[string]bool
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byEmail: map[string]*user.User{}, activated: map[string]bool{}}
}

func (f *fakeUserRepo) Create(context.Context, *user.User) error { return nil }
func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return nil, user.ErrNotFound
}
func (f *fakeUserRepo) FindByIDScoped(_ context.Context, id string) (*user.User, error) {
	for _, u := range f.byEmail {
		if u.ID.Hex() == id {
			return u, nil
		}
	}
	return nil, user.ErrNotFound
}
func (f *fakeUserRepo) ListByTenant(context.Context) ([]user.User, error) { return nil, nil }
func (f *fakeUserRepo) Activate(_ context.Context, id, nome, senhaHash string, _ time.Time) error {
	for _, u := range f.byEmail {
		if u.ID.Hex() == id {
			f.activated[id] = true
			return nil
		}
	}
	return user.ErrNotFound
}

type fakeInviteRepo struct {
	byHash map[string]*invite.Invite
	used   map[string]bool
}

func newFakeInviteRepo() *fakeInviteRepo {
	return &fakeInviteRepo{byHash: map[string]*invite.Invite{}, used: map[string]bool{}}
}

func (f *fakeInviteRepo) Create(context.Context, *invite.Invite) error { return nil }
func (f *fakeInviteRepo) FindByTokenHash(_ context.Context, h string) (*invite.Invite, error) {
	if inv, ok := f.byHash[h]; ok {
		return inv, nil
	}
	return nil, invite.ErrNotFound
}
func (f *fakeInviteRepo) MarkUsed(_ context.Context, id primitive.ObjectID, _ time.Time) error {
	f.used[id.Hex()] = true
	return nil
}
func (f *fakeInviteRepo) DeleteByEmail(context.Context, string) error { return nil }

func newTestService() (*Service, *fakeUserRepo, *fakeInviteRepo, *token.Manager) {
	users, invites := newFakeUserRepo(), newFakeInviteRepo()
	tokens := token.NewManager("test-secret", 30*time.Minute)
	return NewService(users, invites, tokens), users, invites, tokens
}

// seedInvite registers a pending user and its invite, returning the raw token.
func seedInvite(users *fakeUserRepo, invites *fakeInviteRepo, expiresAt time.Time, used bool) (raw string, tenantID primitive.ObjectID) {
	tenantID = primitive.NewObjectID()
	u := &user.User{ID: primitive.NewObjectID(), TenantID: tenantID, Email: "func@empresa.com", Role: user.RoleEditor, Status: user.StatusPendente}
	users.byEmail[u.Email] = u

	raw, hash, _ := token.GenerateInviteToken()
	inv := &invite.Invite{ID: primitive.NewObjectID(), TenantID: tenantID, Email: u.Email, Role: user.RoleEditor, TokenHash: hash, ExpiresAt: expiresAt}
	if used {
		t := time.Now().Add(-time.Hour)
		inv.UsedAt = &t
	}
	invites.byHash[hash] = inv
	return raw, tenantID
}

// --- AcceptInvite ---

func TestAcceptInvite_ValidActivatesAndIssuesSession(t *testing.T) {
	svc, users, invites, tokens := newTestService()
	raw, tenantID := seedInvite(users, invites, time.Now().Add(24*time.Hour), false)

	u, session, err := svc.AcceptInvite(context.Background(), raw, "Maria", "senhaForte1")
	require.NoError(t, err)

	assert.Equal(t, user.StatusAtivo, u.Status)
	assert.Empty(t, u.SenhaHash, "the returned user must never expose the hash")

	claims, err := tokens.Parse(session)
	require.NoError(t, err)
	assert.Equal(t, tenantID.Hex(), claims.TenantID)
	assert.Equal(t, string(user.RoleEditor), claims.Role)
	assert.True(t, invites.used[u.ID.Hex()] || len(invites.used) == 1, "invite must be marked used")
}

func TestAcceptInvite_ExpiredRejected(t *testing.T) {
	svc, users, invites, _ := newTestService()
	raw, _ := seedInvite(users, invites, time.Now().Add(-time.Minute), false)

	_, _, err := svc.AcceptInvite(context.Background(), raw, "Maria", "senhaForte1")
	assert.ErrorIs(t, err, ErrInviteExpired)
}

func TestAcceptInvite_AlreadyUsedRejected(t *testing.T) {
	svc, users, invites, _ := newTestService()
	raw, _ := seedInvite(users, invites, time.Now().Add(24*time.Hour), true)

	_, _, err := svc.AcceptInvite(context.Background(), raw, "Maria", "senhaForte1")
	assert.ErrorIs(t, err, ErrInviteExpired)
}

func TestAcceptInvite_UnknownTokenRejected(t *testing.T) {
	svc, _, _, _ := newTestService()
	_, _, err := svc.AcceptInvite(context.Background(), "deadbeef-token", "Maria", "senhaForte1")
	assert.ErrorIs(t, err, ErrInviteNotFound)
}

func TestAcceptInvite_WeakPasswordRejected(t *testing.T) {
	svc, users, invites, _ := newTestService()
	raw, _ := seedInvite(users, invites, time.Now().Add(24*time.Hour), false)

	for _, weak := range []string{"curta1", "semnumeros", "12345678"} {
		_, _, err := svc.AcceptInvite(context.Background(), raw, "Maria", weak)
		assert.ErrorIs(t, err, ErrWeakPassword, "password %q must be rejected", weak)
	}
}

// --- Login ---

func TestLogin_ActiveUserSucceeds(t *testing.T) {
	svc, users, _, tokens := newTestService()
	hash, _ := password.Hash("senhaForte1")
	tenantID := primitive.NewObjectID()
	users.byEmail["ativo@empresa.com"] = &user.User{
		ID: primitive.NewObjectID(), TenantID: tenantID, Email: "ativo@empresa.com",
		Role: user.RoleAdmin, Status: user.StatusAtivo, SenhaHash: hash,
	}

	u, session, err := svc.Login(context.Background(), "ativo@empresa.com", "senhaForte1")
	require.NoError(t, err)
	assert.Empty(t, u.SenhaHash)

	claims, err := tokens.Parse(session)
	require.NoError(t, err)
	assert.Equal(t, tenantID.Hex(), claims.TenantID)
	assert.Equal(t, string(user.RoleAdmin), claims.Role)
}

func TestLogin_WrongPasswordRejected(t *testing.T) {
	svc, users, _, _ := newTestService()
	hash, _ := password.Hash("senhaForte1")
	users.byEmail["ativo@empresa.com"] = &user.User{
		ID: primitive.NewObjectID(), TenantID: primitive.NewObjectID(), Email: "ativo@empresa.com",
		Status: user.StatusAtivo, SenhaHash: hash,
	}
	_, _, err := svc.Login(context.Background(), "ativo@empresa.com", "errada9")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_PendingUserRejected(t *testing.T) {
	svc, users, _, _ := newTestService()
	hash, _ := password.Hash("senhaForte1")
	users.byEmail["pend@empresa.com"] = &user.User{
		ID: primitive.NewObjectID(), TenantID: primitive.NewObjectID(), Email: "pend@empresa.com",
		Status: user.StatusPendente, SenhaHash: hash,
	}
	_, _, err := svc.Login(context.Background(), "pend@empresa.com", "senhaForte1")
	assert.ErrorIs(t, err, ErrInvalidCredentials, "pending users cannot authenticate")
}

func TestLogin_UnknownEmailRejected(t *testing.T) {
	svc, _, _, _ := newTestService()
	_, _, err := svc.Login(context.Background(), "ninguem@empresa.com", "senhaForte1")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

// --- Me ---

func TestMe_ReturnsScopedUserWithoutSecret(t *testing.T) {
	svc, users, _, _ := newTestService()
	tenantID := primitive.NewObjectID()
	u := &user.User{ID: primitive.NewObjectID(), TenantID: tenantID, Email: "me@empresa.com", Role: user.RoleViewer, Status: user.StatusAtivo, SenhaHash: "secret"}
	users.byEmail[u.Email] = u

	ctx := middleware.WithIdentity(context.Background(), middleware.Identity{UserID: u.ID.Hex(), TenantID: tenantID.Hex(), Role: string(u.Role)})
	got, err := svc.Me(ctx, u.ID.Hex())
	require.NoError(t, err)
	assert.Equal(t, "me@empresa.com", got.Email)
	assert.Empty(t, got.SenhaHash)
}
