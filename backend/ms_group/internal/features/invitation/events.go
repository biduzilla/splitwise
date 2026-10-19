package invitation

import (
	"time"
	"uuid"
)

type GroupMemberAdded struct {
	EventID      uuid.UUID `json:"event_id"`
	EventType    string    `json:"event_type"`
	GroupID      uuid.UUID `json:"group_id"`
	MembershipID uuid.UUID `json:"membership_id"`
	UserID       uuid.UUID `json:"user_id"`
	Role         string    `json:"role"`
	InviterID    uuid.UUID `json:"inviter_id"`
	OccurredAt   time.Time `json:"occurred_at"`
}

func (e GroupMemberAdded) Key() string { return e.GroupID.String() }
