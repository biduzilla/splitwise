package group

import (
	"shared/models"
	"time"
	"uuid"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleMember Role = "member"
)

type Group struct {
	models.BaseModel
	ID       uuid.UUID
	Name     string
	Currency string
	OwnerID  uuid.UUID
}

type Membership struct {
	models.BaseModel
	ID      uuid.UUID
	GroupID uuid.UUID
	UserID  uuid.UUID
	Role    Role
}

type Invitation struct {
	models.BaseModel
	ID        uuid.UUID
	GroupID   uuid.UUID
	InviterID uuid.UUID
	JTI       uuid.UUID
	ExpiresAt time.Time
	UsedAt    *time.Time
	UsedBy    *uuid.UUID
}

type GroupDetail struct {
	Group
	Members []Membership
}
