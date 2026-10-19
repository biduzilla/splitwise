package membership

import (
	"time"
	"uuid"
)

type GroupMemberRemoved struct {
	EventID      uuid.UUID `json:"event_id"`
	EventType    string    `json:"event_type"`
	GroupID      uuid.UUID `json:"group_id"`
	MembershipID uuid.UUID `json:"membership_id"`
	UserID       uuid.UUID `json:"user_id"`
	RemovedBy    uuid.UUID `json:"removed_by"`
	OccurredAt   time.Time `json:"occurred_at"`
}

func (e GroupMemberRemoved) Key() string { return e.GroupID.String() }
