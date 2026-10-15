package group

import (
	"ms_group/internal/features/membership"
	"shared/models"
	"uuid"
)

type Group struct {
	models.BaseModel
	ID       uuid.UUID
	Name     string
	Currency string
	OwnerID  uuid.UUID
}

type GroupDetail struct {
	Group
	Members []membership.Membership
}
