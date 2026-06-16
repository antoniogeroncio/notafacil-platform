// Package testutil provides ephemeral infrastructure for integration tests.
package testutil

import (
	"context"
	"testing"
	"time"

	appdb "github.com/notafacil/platform/backend/pkg/db"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

// NewMongoDB starts an ephemeral MongoDB container, applies the feature
// indexes, and returns a ready database. The container is terminated on test
// cleanup. The test is skipped if Docker is unavailable.
func NewMongoDB(t *testing.T) *mongo.Database {
	t.Helper()
	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Skipf("skipping integration test: cannot start MongoDB container: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	connectCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	database, client, err := appdb.Connect(connectCtx, uri, "test")
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Disconnect(context.Background())
	})

	if err := appdb.EnsureIndexes(ctx, database); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	return database
}
