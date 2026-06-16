package middleware

import (
	"net/http"

	"github.com/notafacil/platform/backend/pkg/httpx"
)

// RequireRole returns middleware that allows the request only if the
// authenticated identity holds one of the given roles, else 403 (FR-008).
// It must be installed after Authenticate.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := IdentityFrom(r.Context())
			if !ok {
				httpx.Error(w, http.StatusUnauthorized, "unauthenticated", "Sessão ausente ou inválida.")
				return
			}
			if _, permitted := allowed[id.Role]; !permitted {
				httpx.Error(w, http.StatusForbidden, "forbidden", "Você não tem permissão para esta ação.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
