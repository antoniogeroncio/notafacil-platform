package middleware

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ErrNoTenant is returned when no tenant can be derived from the context.
// Repositories MUST treat this as fail-closed: deny the operation rather than
// querying without a tenant filter (Princípio III).
var ErrNoTenant = errors.New("middleware: no tenant in authenticated context")

// TenantObjectID returns the tenant ObjectID derived from the context.
// It fails closed when the identity or tenant is absent or malformed.
func TenantObjectID(ctx context.Context) (primitive.ObjectID, error) {
	id, ok := IdentityFrom(ctx)
	if !ok || id.TenantID == "" {
		return primitive.NilObjectID, ErrNoTenant
	}
	oid, err := primitive.ObjectIDFromHex(id.TenantID)
	if err != nil {
		return primitive.NilObjectID, ErrNoTenant
	}
	return oid, nil
}

// TenantScoped returns a copy of filter with the context tenant injected as
// {tenantId}. This is the central, default isolation applied to every query
// and must never be bypassed. The provided filter is not mutated.
func TenantScoped(ctx context.Context, filter bson.M) (bson.M, error) {
	oid, err := TenantObjectID(ctx)
	if err != nil {
		return nil, err
	}
	scoped := bson.M{}
	for k, v := range filter {
		scoped[k] = v
	}
	scoped["tenantId"] = oid
	return scoped, nil
}
