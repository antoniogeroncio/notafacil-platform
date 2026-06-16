// Package middleware provides HTTP middleware and the multi-tenant isolation
// core (Princípio III): the authenticated identity is extracted from the
// session token and injected into the request context; data access is scoped
// to that tenant centrally and by default.
package middleware

import "context"

type ctxKey int

const identityKey ctxKey = iota

// Identity is the authenticated caller derived exclusively from the session
// token — never from client input (FR-010).
type Identity struct {
	UserID   string
	TenantID string
	Role     string
}

// WithIdentity returns a context carrying the authenticated identity.
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityKey, id)
}

// IdentityFrom extracts the authenticated identity from the context.
func IdentityFrom(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey).(Identity)
	return id, ok
}
