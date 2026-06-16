package invite_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/notafacil/platform/backend/internal/invite"
	"github.com/notafacil/platform/backend/internal/testutil"
	"github.com/notafacil/platform/backend/internal/user"
	"github.com/notafacil/platform/backend/pkg/email"
	"github.com/notafacil/platform/backend/pkg/middleware"
	"github.com/notafacil/platform/backend/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type inviteHarness struct {
	router    http.Handler
	tokens    *token.Manager
	sender    *email.FakeSender
	users     *user.MongoRepository
	usersColl *mongo.Collection
}

func setupInviteHarness(t *testing.T) inviteHarness {
	t.Helper()
	db := testutil.NewMongoDB(t)
	tokens := token.NewManager("test-secret", 30*time.Minute)
	sender := &email.FakeSender{}

	userRepo := user.NewMongoRepository(db)
	inviteRepo := invite.NewMongoRepository(db)
	svc := invite.NewService(userRepo, inviteRepo, sender, 48*time.Hour, "http://localhost:3000")
	h := invite.NewHandler(svc)

	r := chi.NewRouter()
	r.Group(func(pr chi.Router) {
		pr.Use(middleware.Authenticate(tokens))
		pr.With(middleware.RequireRole(string(user.RoleAdmin))).Post("/api/v1/invites", h.Create)
	})

	return inviteHarness{router: r, tokens: tokens, sender: sender, users: userRepo, usersColl: db.Collection("users")}
}

func (h inviteHarness) request(t *testing.T, role string, tenantID primitive.ObjectID, email, inviteRole string) *httptest.ResponseRecorder {
	t.Helper()
	sess, err := h.tokens.Issue(token.SessionClaims{UserID: primitive.NewObjectID().Hex(), TenantID: tenantID.Hex(), Role: role})
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]string{"email": email, "role": inviteRole})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invites", bytes.NewReader(body))
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sess})
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)
	return rec
}

func TestInvitesEndpoint_AdminCreatesPendingAndSendsEmail(t *testing.T) {
	h := setupInviteHarness(t)
	tenantID := primitive.NewObjectID()

	rec := h.request(t, string(user.RoleAdmin), tenantID, "novo@empresa.com", string(user.RoleEditor))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	// Pending user persisted in the inviting tenant.
	var stored user.User
	err := h.usersColl.FindOne(context.Background(), bson.M{"email": "novo@empresa.com"}).Decode(&stored)
	require.NoError(t, err)
	assert.Equal(t, user.StatusPendente, stored.Status)
	assert.Equal(t, tenantID, stored.TenantID)

	// Activation e-mail captured.
	sent, ok := h.sender.Last()
	require.True(t, ok)
	assert.Equal(t, "novo@empresa.com", sent.To)

	// Response never leaks the token (Princípio VI).
	assert.NotContains(t, rec.Body.String(), "tokenHash")
	assert.NotContains(t, rec.Body.String(), "senhaHash")
}

func TestInvitesEndpoint_DuplicateActiveEmailConflicts(t *testing.T) {
	h := setupInviteHarness(t)
	tenantID := primitive.NewObjectID()

	// Seed an already active user with that e-mail.
	_, err := h.usersColl.InsertOne(context.Background(), user.User{
		ID: primitive.NewObjectID(), TenantID: tenantID, Email: "ativo@empresa.com",
		Role: user.RoleEditor, Status: user.StatusAtivo, CriadoEm: time.Now(),
	})
	require.NoError(t, err)

	rec := h.request(t, string(user.RoleAdmin), tenantID, "ativo@empresa.com", string(user.RoleEditor))
	assert.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), "email_conflict")
}

func TestInvitesEndpoint_ResendToPendingRenews(t *testing.T) {
	h := setupInviteHarness(t)
	tenantID := primitive.NewObjectID()

	first := h.request(t, string(user.RoleAdmin), tenantID, "pend@empresa.com", string(user.RoleEditor))
	require.Equal(t, http.StatusCreated, first.Code)

	second := h.request(t, string(user.RoleAdmin), tenantID, "pend@empresa.com", string(user.RoleEditor))
	require.Equal(t, http.StatusCreated, second.Code, "resend to pending must renew, not conflict")

	// Exactly one user exists for that e-mail (no duplicate).
	count, err := h.usersColl.CountDocuments(context.Background(), bson.M{"email": "pend@empresa.com"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestInvitesEndpoint_EditorForbidden(t *testing.T) {
	h := setupInviteHarness(t)
	rec := h.request(t, string(user.RoleEditor), primitive.NewObjectID(), "x@empresa.com", string(user.RoleViewer))
	assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestInvitesEndpoint_Unauthenticated(t *testing.T) {
	h := setupInviteHarness(t)
	body, _ := json.Marshal(map[string]string{"email": "x@empresa.com", "role": "Editor"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invites", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
