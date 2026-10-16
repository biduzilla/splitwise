package membership

import (
	"time"
	"uuid"
)

type MemberDTO struct {
	ID       uuid.UUID      `json:"id"`
	GroupID  uuid.UUID      `json:"group_id"`
	User     UserSummaryDTO `json:"user"`
	UserID   uuid.UUID      `json:"user_id"`
	Role     string         `json:"role"`
	JoinedAt time.Time      `json:"joined_at"`
}

type UserSummaryDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (m *Membership) ToDTO() MemberDTO {
	return MemberDTO{
		ID:       m.ID,
		GroupID:  m.GroupID,
		UserID:   m.UserID,
		User:     m.User.ToDTO(),
		Role:     string(m.Role),
		JoinedAt: m.CreatedAt,
	}
}

func (m *UserSummary) ToDTO() UserSummaryDTO {
	return UserSummaryDTO{
		Email: m.Email,
		Name:  m.Name,
	}
}
