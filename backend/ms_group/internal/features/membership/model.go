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

type UserSummary struct {
	Name  string
	Email string
}

type Membership struct {
	models.BaseModel
	ID      uuid.UUID
	GroupID uuid.UUID
	User    UserSummary
	UserID  uuid.UUID
	Role    Role
}
