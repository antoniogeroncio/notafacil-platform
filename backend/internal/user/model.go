// Package user owns the platform user identity, its persistence and lifecycle.
package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role is the permission level a user holds within its tenant (FR-008).
type Role string

const (
	RoleAdmin  Role = "Admin"
	RoleEditor Role = "Editor"
	RoleViewer Role = "Viewer"
)

// ValidRole reports whether r is one of the supported roles.
func ValidRole(r Role) bool {
	switch r {
	case RoleAdmin, RoleEditor, RoleViewer:
		return true
	default:
		return false
	}
}

// Status is the account lifecycle state.
type Status string

const (
	StatusPendente Status = "Pendente"
	StatusAtivo    Status = "Ativo"
)

// User belongs to exactly one tenant. The e-mail is the global login identity
// (unique platform-wide, FR-004). SenhaHash is never serialized to the API
// (Princípio VI).
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	TenantID  primitive.ObjectID `bson:"tenantId"`
	Nome      string             `bson:"nome"`
	Email     string             `bson:"email"`
	SenhaHash string             `bson:"senhaHash,omitempty"`
	Role      Role               `bson:"role"`
	Status    Status             `bson:"status"`
	CriadoEm  time.Time          `bson:"criadoEm"`
	AtivadoEm *time.Time         `bson:"ativadoEm,omitempty"`
}

// View is the API-safe projection of a user (never exposes secrets).
type View struct {
	ID     string `json:"id"`
	Nome   string `json:"nome"`
	Email  string `json:"email"`
	Role   Role   `json:"role"`
	Status Status `json:"status"`
}

// ToView maps a User to its API-safe projection.
func (u User) ToView() View {
	return View{
		ID:     u.ID.Hex(),
		Nome:   u.Nome,
		Email:  u.Email,
		Role:   u.Role,
		Status: u.Status,
	}
}
