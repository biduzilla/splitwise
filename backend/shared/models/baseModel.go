package models

import (
	"time"
	"uuid"
)

type BaseModel struct {
	Version   int
	Deleted   bool
	CreatedAt time.Time
	CreatedBy *uuid.UUID
	UpdatedAt *time.Time
	UpdatedBy *uuid.UUID
}
