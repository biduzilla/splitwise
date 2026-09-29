package membership

import (
	"time"
	"uuid"
)

type MemberDTO struct {
	ID       uuid.UUID `json:"id"`
	GroupID  uuid.UUID `json:"group_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

func (m *Membership) ToDTO() MemberDTO {
	return MemberDTO{
		ID:       m.ID,
		GroupID:  m.GroupID,
		UserID:   m.UserID,
		Role:     string(m.Role),
		JoinedAt: m.CreatedAt,
	}
}
