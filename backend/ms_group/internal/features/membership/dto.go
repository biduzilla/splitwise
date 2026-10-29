package membership

import (
	"time"
	"uuid"
)

type MemberDTO struct {
	ID       uuid.UUID      `json:"id"`
	GroupID  uuid.UUID      `json:"group_id"`
	User     *MemberUserDTO `json:"user,omitempty"`
	UserID   uuid.UUID      `json:"user_id"`
	Role     string         `json:"role"`
	JoinedAt time.Time      `json:"joined_at"`
}

type MemberUserDTO struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

func (m *Membership) ToDTO() MemberDTO {
	u := m.User.ToDTO()
	return MemberDTO{
		ID:       m.ID,
		GroupID:  m.GroupID,
		UserID:   m.UserID,
		Role:     string(m.Role),
		JoinedAt: m.CreatedAt,
		User:     &u,
	}
}

func (m *MemberUser) ToDTO() MemberUserDTO {
	return MemberUserDTO{
		ID:    m.ID,
		Name:  m.Name,
		Email: m.Email,
	}
}
