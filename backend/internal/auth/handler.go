package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/notafacil/platform/backend/pkg/httpx"
	"github.com/notafacil/platform/backend/pkg/middleware"
)

// Handler is the HTTP entry point for authentication (Controller layer).
type Handler struct {
	svc        *Service
	sessionTTL time.Duration
	secure     bool
}

// NewHandler builds an auth Handler. secure controls the Secure cookie flag
// (true in production over HTTPS).
func NewHandler(svc *Service, sessionTTL time.Duration, secure bool) *Handler {
	return &Handler{svc: svc, sessionTTL: sessionTTL, secure: secure}
}

type acceptRequest struct {
	Nome  string `json:"nome"`
	Senha string `json:"senha"`
}

type loginRequest struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

// AcceptInvite handles POST /api/v1/invites/{token}/accept (public).
func (h *Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	rawToken := chi.URLParam(r, "token")
	var body acceptRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation_error", "Corpo da requisição inválido.")
		return
	}

	u, session, err := h.svc.AcceptInvite(r.Context(), rawToken, body.Nome, body.Senha)
	switch {
	case errors.Is(err, ErrWeakPassword):
		httpx.Error(w, http.StatusUnprocessableEntity, "weak_password", "A senha não atende à política mínima (mínimo de 8 caracteres, com letras e números).")
		return
	case errors.Is(err, ErrInviteExpired):
		httpx.Error(w, http.StatusGone, "invite_expired", "Este convite expirou ou já foi utilizado. Solicite um novo convite.")
		return
	case errors.Is(err, ErrInviteNotFound):
		httpx.Error(w, http.StatusNotFound, "invite_not_found", "Convite inválido.")
		return
	case err != nil:
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "Não foi possível ativar a conta.")
		return
	}

	h.setSessionCookie(w, session)
	httpx.JSON(w, http.StatusOK, map[string]any{"user": u.ToView()})
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation_error", "Corpo da requisição inválido.")
		return
	}

	u, session, err := h.svc.Login(r.Context(), body.Email, body.Senha)
	if errors.Is(err, ErrInvalidCredentials) {
		httpx.Error(w, http.StatusUnauthorized, "invalid_credentials", "E-mail ou senha inválidos.")
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "Não foi possível autenticar.")
		return
	}

	h.setSessionCookie(w, session)
	httpx.JSON(w, http.StatusOK, map[string]any{"user": u.ToView()})
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, _ *http.Request) {
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// Me handles GET /api/v1/me (authenticated).
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.IdentityFrom(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthenticated", "Sessão ausente ou inválida.")
		return
	}
	u, err := h.svc.Me(r.Context(), id.UserID)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "unauthenticated", "Sessão ausente ou inválida.")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"user": u.ToView()})
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.sessionTTL.Seconds()),
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
