package middleware

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTenantScopedFailsClosedWithoutIdentity(t *testing.T) {
	if _, err := TenantScoped(context.Background(), bson.M{"email": "a@b.com"}); !errors.Is(err, ErrNoTenant) {
		t.Fatalf("expected ErrNoTenant when no identity in context, got %v", err)
	}
}

func TestTenantScopedFailsClosedWithEmptyTenant(t *testing.T) {
	ctx := WithIdentity(context.Background(), Identity{UserID: "u1", TenantID: "", Role: "Admin"})
	if _, err := TenantScoped(ctx, bson.M{}); !errors.Is(err, ErrNoTenant) {
		t.Fatalf("expected ErrNoTenant with empty tenant, got %v", err)
	}
}

func TestTenantScopedRejectsMalformedTenant(t *testing.T) {
	ctx := WithIdentity(context.Background(), Identity{TenantID: "not-an-objectid"})
	if _, err := TenantScoped(ctx, bson.M{}); !errors.Is(err, ErrNoTenant) {
		t.Fatalf("expected ErrNoTenant with malformed tenant id, got %v", err)
	}
}

func TestTenantScopedInjectsTenantAndPreservesFilter(t *testing.T) {
	tid := primitive.NewObjectID()
	ctx := WithIdentity(context.Background(), Identity{TenantID: tid.Hex(), Role: "Admin"})

	scoped, err := TenantScoped(ctx, bson.M{"email": "a@b.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scoped["email"] != "a@b.com" {
		t.Errorf("base filter key was dropped: %+v", scoped)
	}
	got, ok := scoped["tenantId"].(primitive.ObjectID)
	if !ok || got != tid {
		t.Errorf("tenantId not injected correctly: %+v", scoped["tenantId"])
	}
}

func TestTenantScopedDoesNotMutateInput(t *testing.T) {
	tid := primitive.NewObjectID()
	ctx := WithIdentity(context.Background(), Identity{TenantID: tid.Hex()})
	base := bson.M{"email": "a@b.com"}
	if _, err := TenantScoped(ctx, base); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, exists := base["tenantId"]; exists {
		t.Error("input filter must not be mutated")
	}
}
