package auth_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/notafacil/platform/backend/internal/invite"
	"github.com/notafacil/platform/backend/internal/user"
	"github.com/notafacil/platform/backend/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// seedPendingInvite inserts a pending user and a matching invite, returning the
// raw token to be redeemed.
func seedPendingInvite(t *testing.T, h authHarness, expiresAt time.Time, used bool) (raw, email string) {
	t.Helper()
	ctx := context.Background()
	email = "func@empresa.com"
	tenantID := primitive.NewObjectID()

	u := &user.User{TenantID: tenantID, Email: email, Role: user.RoleEditor, Status: user.StatusPendente, CriadoEm: time.Now()}
	require.NoError(t, h.userRepo.Create(ctx, u))

	raw, hash, err := token.GenerateInviteToken()
	require.NoError(t, err)
	inv := &invite.Invite{TenantID: tenantID, Email: email, Role: user.RoleEditor, TokenHash: hash, ExpiresAt: expiresAt}
	if used {
		ts := time.Now().Add(-time.Hour)
		inv.UsedAt = &ts
	}
	require.NoError(t, h.inviteRepo.Create(ctx, inv))
	return raw, email
}

func TestAccept_ValidActivatesAndAuthenticates(t *testing.T) {
	h := newAuthHarness(t)
	raw, email := seedPendingInvite(t, h, time.Now().Add(24*time.Hour), false)

	rec := h.postJSON("/api/v1/invites/"+raw+"/accept", map[string]string{"nome": "Maria", "senha": "senhaForte1"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotNil(t, sessionCookie(rec), "activation must establish a session")

	// User is now active in the database.
	var stored user.User
	require.NoError(t, h.usersColl.FindOne(context.Background(), bson.M{"email": email}).Decode(&stored))
	assert.Equal(t, user.StatusAtivo, stored.Status)
	assert.Equal(t, "Maria", stored.Nome)
	assert.NotEmpty(t, stored.SenhaHash, "password must be stored hashed")

	// Invite is marked used.
	var inv invite.Invite
	require.NoError(t, h.invitesColl.FindOne(context.Background(), bson.M{"email": email}).Decode(&inv))
	assert.NotNil(t, inv.UsedAt, "invite must be marked used")

	// Response never leaks secrets.
	assert.NotContains(t, rec.Body.String(), "senhaHash")
	assert.NotContains(t, rec.Body.String(), "tokenHash")
}

func TestAccept_ExpiredReturns410(t *testing.T) {
	h := newAuthHarness(t)
	raw, _ := seedPendingInvite(t, h, time.Now().Add(-time.Minute), false)

	rec := h.postJSON("/api/v1/invites/"+raw+"/accept", map[string]string{"nome": "Maria", "senha": "senhaForte1"})
	assert.Equal(t, http.StatusGone, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), "invite_expired")
}

func TestAccept_AlreadyUsedReturns410(t *testing.T) {
	h := newAuthHarness(t)
	raw, _ := seedPendingInvite(t, h, time.Now().Add(24*time.Hour), true)

	rec := h.postJSON("/api/v1/invites/"+raw+"/accept", map[string]string{"nome": "Maria", "senha": "senhaForte1"})
	assert.Equal(t, http.StatusGone, rec.Code, rec.Body.String())
}

func TestAccept_UnknownTokenReturns404(t *testing.T) {
	h := newAuthHarness(t)
	rec := h.postJSON("/api/v1/invites/0000deadbeef/accept", map[string]string{"nome": "Maria", "senha": "senhaForte1"})
	assert.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), "invite_not_found")
}

func TestAccept_WeakPasswordReturns422(t *testing.T) {
	h := newAuthHarness(t)
	raw, _ := seedPendingInvite(t, h, time.Now().Add(24*time.Hour), false)

	rec := h.postJSON("/api/v1/invites/"+raw+"/accept", map[string]string{"nome": "Maria", "senha": "fraca"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), "weak_password")

	// The account must remain pending.
	assert.Equal(t, int64(1), h.countUsers(t, bson.M{"status": user.StatusPendente}))
}
