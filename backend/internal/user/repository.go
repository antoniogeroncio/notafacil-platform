package user

import (
	"context"
	"errors"
	"time"

	"github.com/notafacil/platform/backend/pkg/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrNotFound is returned when no user matches a query (including cross-tenant
// access, which must look like a missing resource — Princípio III/SC-002).
var ErrNotFound = errors.New("user: not found")

// ErrEmailTaken is returned when the global e-mail uniqueness is violated.
var ErrEmailTaken = errors.New("user: email already taken")

// Repository abstracts user persistence. Services depend on this interface.
type Repository interface {
	// Create inserts a new user. The tenant is taken from the struct (set by
	// the service from the authenticated context), not from client input.
	Create(ctx context.Context, u *User) error
	// FindByEmail looks a user up by its global login identity. This lookup is
	// intentionally NOT tenant-scoped: e-mail is unique platform-wide (FR-004)
	// and is also used by the public login/activation flows, which have no
	// authenticated tenant yet.
	FindByEmail(ctx context.Context, email string) (*User, error)
	// FindByIDScoped fetches a user by id within the authenticated tenant.
	// Accessing another tenant's id returns ErrNotFound (no existence leak).
	FindByIDScoped(ctx context.Context, id string) (*User, error)
	// ListByTenant returns all users of the authenticated tenant.
	ListByTenant(ctx context.Context) ([]User, error)
	// Activate transitions a pending user to active. Keyed by the globally
	// unique _id; used by the public activation flow (no tenant context).
	Activate(ctx context.Context, id string, nome, senhaHash string, when time.Time) error
}

// MongoRepository is the MongoDB-backed Repository.
type MongoRepository struct {
	col *mongo.Collection
}

// NewMongoRepository builds a MongoRepository over the "users" collection.
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{col: db.Collection("users")}
}

func (r *MongoRepository) Create(ctx context.Context, u *User) error {
	res, err := r.col.InsertOne(ctx, u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrEmailTaken
		}
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		u.ID = oid
	}
	return nil
}

func (r *MongoRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *MongoRepository) FindByIDScoped(ctx context.Context, id string) (*User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrNotFound
	}
	filter, err := middleware.TenantScoped(ctx, bson.M{"_id": oid})
	if err != nil {
		return nil, err
	}
	var u User
	err = r.col.FindOne(ctx, filter).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *MongoRepository) ListByTenant(ctx context.Context) ([]User, error) {
	filter, err := middleware.TenantScoped(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	cur, err := r.col.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "criadoEm", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var users []User
	if err := cur.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *MongoRepository) Activate(ctx context.Context, id string, nome, senhaHash string, when time.Time) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrNotFound
	}
	res, err := r.col.UpdateOne(ctx,
		bson.M{"_id": oid, "status": StatusPendente},
		bson.M{"$set": bson.M{
			"nome":      nome,
			"senhaHash": senhaHash,
			"status":    StatusAtivo,
			"ativadoEm": when,
		}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}
