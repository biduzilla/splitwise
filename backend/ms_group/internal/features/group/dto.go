package group

import (
	"time"
	"uuid"

	"shared/validator"
)

type CreateGroupDTO struct {
	Name string `json:"name"`
}

func (d CreateGroupDTO) Validate(v *validator.Validator) {
	v.CheckStringNotEmpty(d.Name, "name")
	v.CheckStringMinLen(d.Name, 3, "name")
	v.CheckStringMaxLen(d.Name, 50, "name")
}

type GroupDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	OwnerID   uuid.UUID `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	Version   int       `json:"version"`
}

type MemberDTO struct {
	ID       uuid.UUID `json:"id"`
	GroupID  uuid.UUID `json:"group_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type GroupDetailDTO struct {
	GroupDTO
	Members []MemberDTO `json:"members"`
}

type InvitationDTO struct {
	Token     string    `json:"token"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (g *Group) ToDTO() GroupDTO {
	return GroupDTO{
		ID:        g.ID,
		Name:      g.Name,
		Currency:  g.Currency,
		OwnerID:   g.OwnerID,
		CreatedAt: g.CreatedAt,
		Version:   g.Version,
	}
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

func (d GroupDetail) ToDTO() GroupDetailDTO {
	members := make([]MemberDTO, len(d.Members))
	for i, m := range d.Members {
		members[i] = m.ToDTO()
	}
	return GroupDetailDTO{
		GroupDTO: d.Group.ToDTO(),
		Members:  members,
	}
}
