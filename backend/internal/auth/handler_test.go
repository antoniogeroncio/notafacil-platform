package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/notafacil/platform/backend/internal/auth"
	"github.com/notafacil/platform/backend/internal/invite"
	"github.com/notafacil/platform/backend/internal/testutil"
	"github.com/notafacil/platform/backend/internal/user"
	"github.com/notafacil/platform/backend/pkg/middleware"
	"github.com/notafacil/platform/backend/pkg/password"
	"github.com/notafacil/platform/backend/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type authHarness struct {
	router      http.Handler
	userRepo    *user.MongoRepository
	inviteRepo  *invite.MongoRepository
	usersColl   *mongo.Collection
	invitesColl *mongo.Collection
	tokens      *token.Manager
}

func newAuthHarness(t *testing.T) authHarness {
	t.Helper()
	db := testutil.NewMongoDB(t)
	tokens := token.NewManager("test-secret", 30*time.Minute)

	userRepo := user.NewMongoRepository(db)
	inviteRepo := invite.NewMongoRepository(db)
	svc := auth.NewService(userRepo, inviteRepo, tokens)
	h := auth.NewHandler(svc, 30*time.Minute, false)

	r := chi.NewRouter()
	r.Post("/api/v1/invites/{token}/accept", h.AcceptInvite)
	r.Post("/api/v1/auth/login", h.Login)
	r.Post("/api/v1/auth/logout", h.Logout)
	r.Group(func(pr chi.Router) {
		pr.Use(middleware.Authenticate(tokens))
		pr.Get("/api/v1/me", h.Me)
	})

	return authHarness{
		router:      r,
		userRepo:    userRepo,
		inviteRepo:  inviteRepo,
		usersColl:   db.Collection("users"),
		invitesColl: db.Collection("invites"),
		tokens:      tokens,
	}
}

func (h authHarness) postJSON(path string, payload any, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)
	return rec
}

func sessionCookie(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == middleware.SessionCookieName && c.Value != "" {
			return c
		}
	}
	return nil
}

func seedActiveUser(t *testing.T, h authHarness, email, plain string, role user.Role) primitive.ObjectID {
	t.Helper()
	hash, err := password.Hash(plain)
	require.NoError(t, err)
	tenantID := primitive.NewObjectID()
	_, err = h.usersColl.InsertOne(context.Background(), user.User{
		ID: primitive.NewObjectID(), TenantID: tenantID, Email: email, Nome: "Ativo",
		Role: role, Status: user.StatusAtivo, SenhaHash: hash, CriadoEm: time.Now(),
	})
	require.NoError(t, err)
	return tenantID
}

func TestLogin_ActiveUserSetsSessionCookie(t *testing.T) {
	h := newAuthHarness(t)
	seedActiveUser(t, h, "ativo@empresa.com", "senhaForte1", user.RoleAdmin)

	rec := h.postJSON("/api/v1/auth/login", map[string]string{"email": "ativo@empresa.com", "senha": "senhaForte1"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotNil(t, sessionCookie(rec), "a session cookie must be set")

	cookie := sessionCookie(rec)
	assert.True(t, cookie.HttpOnly, "session cookie must be httpOnly")
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.NotContains(t, rec.Body.String(), "senhaHash")
}

func TestLogin_WrongPasswordUnauthorized(t *testing.T) {
	h := newAuthHarness(t)
	seedActiveUser(t, h, "ativo@empresa.com", "senhaForte1", user.RoleAdmin)

	rec := h.postJSON("/api/v1/auth/login", map[string]string{"email": "ativo@empresa.com", "senha": "errada99"})
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid_credentials")
}

func TestMe_RequiresSessionAndReturnsUser(t *testing.T) {
	h := newAuthHarness(t)
	seedActiveUser(t, h, "ativo@empresa.com", "senhaForte1", user.RoleViewer)

	login := h.postJSON("/api/v1/auth/login", map[string]string{"email": "ativo@empresa.com", "senha": "senhaForte1"})
	require.Equal(t, http.StatusOK, login.Code)

	// With session.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(sessionCookie(login))
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), "ativo@empresa.com")

	// Without session.
	bare := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	bareRec := httptest.NewRecorder()
	h.router.ServeHTTP(bareRec, bare)
	assert.Equal(t, http.StatusUnauthorized, bareRec.Code)
}

func TestLogout_ClearsSessionCookie(t *testing.T) {
	h := newAuthHarness(t)
	rec := h.postJSON("/api/v1/auth/logout", nil)
	require.Equal(t, http.StatusNoContent, rec.Code)

	var cleared bool
	for _, c := range rec.Result().Cookies() {
		if c.Name == middleware.SessionCookieName && c.MaxAge < 0 {
			cleared = true
		}
	}
	assert.True(t, cleared, "logout must expire the session cookie")
}

// helper used by accept_test.go to assert persistence
func (h authHarness) countUsers(t *testing.T, filter bson.M) int64 {
	t.Helper()
	n, err := h.usersColl.CountDocuments(context.Background(), filter)
	require.NoError(t, err)
	return n
}
