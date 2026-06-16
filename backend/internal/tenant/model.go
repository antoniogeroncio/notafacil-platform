// Package tenant owns the company (tenant) entity, the root of data isolation.
package tenant

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Tenant is the customer company. It is the root of isolation and therefore
// carries no tenantId of its own (Princípio III).
type Tenant struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	RazaoSocial string             `bson:"razaoSocial"`
	CNPJ        string             `bson:"cnpj"`
	CriadoEm    time.Time          `bson:"criadoEm"`
}
