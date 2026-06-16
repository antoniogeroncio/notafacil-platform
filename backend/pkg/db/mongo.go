// Package db manages the MongoDB connection and index setup.
//
// The platform is multi-tenant single-database: every tenant-scoped collection
// carries tenantId and an index that begins with it (Princípio III).
package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect opens a MongoDB connection and returns the target database.
func Connect(ctx context.Context, uri, dbName string) (*mongo.Database, *mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, err
	}
	return client.Database(dbName), client, nil
}

// EnsureIndexes creates all indexes required by feature 001. Safe to call on
// every startup (index creation is idempotent).
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	if _, err := db.Collection("tenants").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "cnpj", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_cnpj"),
	}); err != nil {
		return err
	}

	if _, err := db.Collection("users").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_email_global"),
		},
		{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "_id", Value: 1}}, Options: options.Index().SetName("tenant_id")},
		{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "role", Value: 1}}, Options: options.Index().SetName("tenant_role")},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("invites").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "tokenHash", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_token_hash"),
		},
		{Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "email", Value: 1}}, Options: options.Index().SetName("tenant_email")},
		// Plain index (not TTL): expired invites are kept so activation can
		// return 410 (expired) rather than 404 (not found), per the contract.
		{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetName("expires_at")},
	}); err != nil {
		return err
	}
	return nil
}
