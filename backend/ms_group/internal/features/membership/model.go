package membership

import (
	"ms_group/internal/core/httpclient/user"
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
	User    *user.UserDTO
}
