package invite

import (
	"context"
	"errors"
	"time"

	"github.com/notafacil/platform/backend/pkg/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrNotFound is returned when no invite matches a query.
var ErrNotFound = errors.New("invite: not found")

// Repository abstracts invite persistence. Services depend on this interface.
type Repository interface {
	// Create inserts a new invite within the authenticated tenant (taken from
	// the struct, set by the service from context).
	Create(ctx context.Context, inv *Invite) error
	// FindByTokenHash looks an invite up by the SHA-256 hash of its token.
	// Not tenant-scoped: the token is globally unique and the activation flow
	// is public (no authenticated tenant); the tenant is derived from the
	// returned invite, never from client input.
	FindByTokenHash(ctx context.Context, tokenHash string) (*Invite, error)
	// MarkUsed records that the invite has been redeemed. Keyed by _id.
	MarkUsed(ctx context.Context, id primitive.ObjectID, when time.Time) error
	// DeleteByEmail removes any existing invites for an e-mail within the
	// authenticated tenant (used to renew an invite without duplicates).
	DeleteByEmail(ctx context.Context, email string) error
}

// MongoRepository is the MongoDB-backed Repository.
type MongoRepository struct {
	col *mongo.Collection
}

// NewMongoRepository builds a MongoRepository over the "invites" collection.
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{col: db.Collection("invites")}
}

func (r *MongoRepository) Create(ctx context.Context, inv *Invite) error {
	res, err := r.col.InsertOne(ctx, inv)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		inv.ID = oid
	}
	return nil
}

func (r *MongoRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*Invite, error) {
	var inv Invite
	err := r.col.FindOne(ctx, bson.M{"tokenHash": tokenHash}).Decode(&inv)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *MongoRepository) MarkUsed(ctx context.Context, id primitive.ObjectID, when time.Time) error {
	res, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id, "usedAt": nil},
		bson.M{"$set": bson.M{"usedAt": when}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MongoRepository) DeleteByEmail(ctx context.Context, email string) error {
	filter, err := middleware.TenantScoped(ctx, bson.M{"email": email})
	if err != nil {
		return err
	}
	_, err = r.col.DeleteMany(ctx, filter)
	return err
}
