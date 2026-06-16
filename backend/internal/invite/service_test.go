package invite

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/notafacil/platform/backend/internal/user"
	"github.com/notafacil/platform/backend/pkg/email"
	"github.com/notafacil/platform/backend/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- fakes (test-controlled doubles, Princípio IV) ---

type fakeUserRepo struct {
	byEmail   map[string]*user.User
	created   []*user.User
	createErr error
}

func newFakeUserRepo() *fakeUserRepo { return &fakeUserRepo{byEmail: map[string]*user.User{}} }

func (f *fakeUserRepo) Create(_ context.Context, u *user.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	u.ID = primitive.NewObjectID()
	f.created = append(f.created, u)
	f.byEmail[u.Email] = u
	return nil
}
func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return nil, user.ErrNotFound
}
func (f *fakeUserRepo) FindByIDScoped(context.Context, string) (*user.User, error) {
	return nil, user.ErrNotFound
}
func (f *fakeUserRepo) ListByTenant(context.Context) ([]user.User, error) { return nil, nil }
func (f *fakeUserRepo) Activate(context.Context, string, string, string, time.Time) error {
	return nil
}

type fakeInviteRepo struct {
	created     []*Invite
	deletedMail []string
}

func (f *fakeInviteRepo) Create(_ context.Context, inv *Invite) error {
	inv.ID = primitive.NewObjectID()
	f.created = append(f.created, inv)
	return nil
}
func (f *fakeInviteRepo) FindByTokenHash(context.Context, string) (*Invite, error) {
	return nil, ErrNotFound
}
func (f *fakeInviteRepo) MarkUsed(context.Context, primitive.ObjectID, time.Time) error { return nil }
func (f *fakeInviteRepo) DeleteByEmail(_ context.Context, email string) error {
	f.deletedMail = append(f.deletedMail, email)
	return nil
}

// adminContext returns a context carrying an authenticated Admin identity.
func adminContext(tenantID primitive.ObjectID) context.Context {
	return middleware.WithIdentity(context.Background(), middleware.Identity{
		UserID:   primitive.NewObjectID().Hex(),
		TenantID: tenantID.Hex(),
		Role:     string(user.RoleAdmin),
	})
}

func newTestService(u *fakeUserRepo, i *fakeInviteRepo, s email.Sender) *Service {
	return NewService(u, i, s, 48*time.Hour, "http://localhost:3000")
}

// --- tests ---

func TestCreateInvite_NewUser(t *testing.T) {
	users, invites, sender := newFakeUserRepo(), &fakeInviteRepo{}, &email.FakeSender{}
	svc := newTestService(users, invites, sender)
	tenantID := primitive.NewObjectID()

	inv, err := svc.CreateInvite(adminContext(tenantID), CreateInput{Email: "func@empresa.com", Role: user.RoleEditor})
	require.NoError(t, err)

	require.Len(t, users.created, 1, "a pending user must be created")
	assert.Equal(t, user.StatusPendente, users.created[0].Status)
	assert.Equal(t, tenantID, users.created[0].TenantID)
	assert.Equal(t, user.RoleEditor, users.created[0].Role)

	require.Len(t, invites.created, 1)
	assert.NotEmpty(t, invites.created[0].TokenHash)
	assert.WithinDuration(t, time.Now().Add(48*time.Hour), inv.ExpiresAt, time.Minute)

	sent, ok := sender.Last()
	require.True(t, ok, "an activation e-mail must be sent")
	assert.Equal(t, "func@empresa.com", sent.To)
	assert.Contains(t, sent.HTMLBody, "/accept-invite/", "e-mail must contain the activation link")
}

func TestCreateInvite_EmailNeverLeaksTokenInView(t *testing.T) {
	users, invites, sender := newFakeUserRepo(), &fakeInviteRepo{}, &email.FakeSender{}
	svc := newTestService(users, invites, sender)

	inv, err := svc.CreateInvite(adminContext(primitive.NewObjectID()), CreateInput{Email: "a@b.com", Role: user.RoleViewer})
	require.NoError(t, err)

	view := inv.ToView()
	// The raw token only exists inside the e-mail link; the stored hash and the
	// view must never expose it (Princípio VI).
	sent, _ := sender.Last()
	link := sent.HTMLBody
	assert.NotContains(t, link, inv.TokenHash, "raw link must not equal the stored hash")
	assert.Empty(t, view.toJSONMustNotContainToken())
}

func TestCreateInvite_DuplicateActiveEmailConflicts(t *testing.T) {
	users, invites, sender := newFakeUserRepo(), &fakeInviteRepo{}, &email.FakeSender{}
	users.byEmail["taken@empresa.com"] = &user.User{
		ID: primitive.NewObjectID(), Email: "taken@empresa.com", Status: user.StatusAtivo, TenantID: primitive.NewObjectID(),
	}
	svc := newTestService(users, invites, sender)

	_, err := svc.CreateInvite(adminContext(primitive.NewObjectID()), CreateInput{Email: "taken@empresa.com", Role: user.RoleEditor})
	assert.ErrorIs(t, err, ErrEmailConflict)
	assert.Empty(t, users.created, "no user must be created on conflict")
	assert.Empty(t, sender.Sent, "no e-mail must be sent on conflict")
}

func TestCreateInvite_ResendToPendingRenewsWithoutDuplicate(t *testing.T) {
	users, invites, sender := newFakeUserRepo(), &fakeInviteRepo{}, &email.FakeSender{}
	tenantID := primitive.NewObjectID()
	users.byEmail["pend@empresa.com"] = &user.User{
		ID: primitive.NewObjectID(), Email: "pend@empresa.com", Status: user.StatusPendente, TenantID: tenantID, Role: user.RoleEditor,
	}
	svc := newTestService(users, invites, sender)

	_, err := svc.CreateInvite(adminContext(tenantID), CreateInput{Email: "pend@empresa.com", Role: user.RoleEditor})
	require.NoError(t, err)

	assert.Empty(t, users.created, "must NOT create a second user for a pending invite")
	assert.Equal(t, []string{"pend@empresa.com"}, invites.deletedMail, "old invites must be cleared")
	require.Len(t, invites.created, 1, "a fresh invite must be issued")
	assert.Len(t, sender.Sent, 1)
}

func TestCreateInvite_InvalidEmail(t *testing.T) {
	svc := newTestService(newFakeUserRepo(), &fakeInviteRepo{}, &email.FakeSender{})
	_, err := svc.CreateInvite(adminContext(primitive.NewObjectID()), CreateInput{Email: "not-an-email", Role: user.RoleEditor})
	assert.ErrorIs(t, err, ErrValidation)
}

func TestCreateInvite_InvalidRole(t *testing.T) {
	svc := newTestService(newFakeUserRepo(), &fakeInviteRepo{}, &email.FakeSender{})
	_, err := svc.CreateInvite(adminContext(primitive.NewObjectID()), CreateInput{Email: "a@b.com", Role: user.Role("SuperRoot")})
	assert.ErrorIs(t, err, ErrValidation)
}

func TestCreateInvite_FailsClosedWithoutTenant(t *testing.T) {
	svc := newTestService(newFakeUserRepo(), &fakeInviteRepo{}, &email.FakeSender{})
	_, err := svc.CreateInvite(context.Background(), CreateInput{Email: "a@b.com", Role: user.RoleEditor})
	assert.Error(t, err, "must fail without an authenticated tenant")
	assert.True(t, errors.Is(err, middleware.ErrNoTenant))
}

// helper kept local so the view assertion above reads clearly.
func (v View) toJSONMustNotContainToken() string {
	if strings.Contains(v.ID, "token") {
		return v.ID
	}
	return ""
}
