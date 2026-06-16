// Package invite owns team invitations: a temporary, hashed token that lets a
// pending user activate their account (FR-003/FR-005/FR-006).
package invite

import (
	"time"

	"github.com/notafacil/platform/backend/internal/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Invite is a tenant-scoped activation grant. Only the SHA-256 hash of the
// token is persisted — never the raw token (Princípio VI).
type Invite struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	TenantID        primitive.ObjectID `bson:"tenantId"`
	Email           string             `bson:"email"`
	Role            user.Role          `bson:"role"`
	TokenHash       string             `bson:"tokenHash"`
	ExpiresAt       time.Time          `bson:"expiresAt"`
	UsedAt          *time.Time         `bson:"usedAt,omitempty"`
	InvitedByUserID primitive.ObjectID `bson:"invitedByUserId"`
}

// IsUsable reports whether the invite can still be redeemed at time now:
// not yet used and not expired.
func (i Invite) IsUsable(now time.Time) bool {
	return i.UsedAt == nil && now.Before(i.ExpiresAt)
}

// View is the API-safe projection of an invite (never exposes the token).
type View struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      user.Role `json:"role"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// ToView maps an Invite to its API-safe projection.
func (i Invite) ToView() View {
	return View{
		ID:        i.ID.Hex(),
		Email:     i.Email,
		Role:      i.Role,
		ExpiresAt: i.ExpiresAt,
	}
}
