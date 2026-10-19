package group

import (
	"messaging"
	"time"
	"uuid"
)

type GroupCreated struct {
	EventID    uuid.UUID           `json:"event_id"`
	EventType  messaging.EventType `json:"event_type"`
	GroupID    uuid.UUID           `json:"group_id"`
	Name       string              `json:"name"`
	OwnerID    uuid.UUID           `json:"owner_id"`
	Currency   string              `json:"currency"`
	OccurredAt time.Time           `json:"occurred_at"`
}

func (e GroupCreated) Key() string { return e.GroupID.String() }
