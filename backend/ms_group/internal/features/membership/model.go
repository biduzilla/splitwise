package membership

import (
	"shared/models"
	"uuid"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleMember Role = "member"
)

type Membership struct {
	models.BaseModel
	ID      uuid.UUID
	GroupID uuid.UUID
	UserID  uuid.UUID
	Role    Role
}
