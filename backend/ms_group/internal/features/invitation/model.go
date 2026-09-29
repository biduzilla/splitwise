package invitation

import (
	"shared/models"
	"time"
	"uuid"
)

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
