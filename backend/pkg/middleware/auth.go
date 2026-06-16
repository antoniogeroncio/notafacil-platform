package middleware

import (
	"net/http"

	"github.com/notafacil/platform/backend/pkg/httpx"
	"github.com/notafacil/platform/backend/pkg/token"
)

// SessionCookieName is the name of the httpOnly session cookie.
const SessionCookieName = "nf_session"

// Authenticator parses session tokens.
type Authenticator interface {
	Parse(raw string) (token.SessionClaims, error)
}

// Authenticate returns middleware that requires a valid session cookie and
// injects the derived identity into the request context. It fails closed:
// any missing or invalid token yields 401 (Princípio III/FR-010).
func Authenticate(tm Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil || cookie.Value == "" {
				httpx.Error(w, http.StatusUnauthorized, "unauthenticated", "Sessão ausente ou inválida.")
				return
			}
			claims, err := tm.Parse(cookie.Value)
			if err != nil || claims.TenantID == "" || claims.UserID == "" {
				httpx.Error(w, http.StatusUnauthorized, "unauthenticated", "Sessão ausente ou inválida.")
				return
			}
			ctx := WithIdentity(r.Context(), Identity{
				UserID:   claims.UserID,
				TenantID: claims.TenantID,
				Role:     claims.Role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
