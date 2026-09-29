package group

import (
	"ms_group/internal/features/membership"
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

type GroupDetailDTO struct {
	GroupDTO
	Members []membership.MemberDTO `json:"members"`
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

func (d GroupDetail) ToDTO() GroupDetailDTO {
	members := make([]membership.MemberDTO, len(d.Members))
	for i, m := range d.Members {
		members[i] = m.ToDTO()
	}
	return GroupDetailDTO{
		GroupDTO: d.Group.ToDTO(),
		Members:  members,
	}
}
