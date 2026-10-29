package groups

import (
	"context"
	"fmt"
	"net/http"
	"shared/apierror"
	"shared/httpx/httpclient"
	"time"
	"uuid"
)

type GroupDetailDTO struct {
	GroupDTO
	Members []MemberDTO `json:"members"`
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
	ID       uuid.UUID     `json:"id"`
	GroupID  uuid.UUID     `json:"group_id"`
	User     MemberUserDTO `json:"user"`
	UserID   uuid.UUID     `json:"user_id"`
	Role     string        `json:"role"`
	JoinedAt time.Time     `json:"joined_at"`
}

type MemberUserDTO struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type HTTPClient struct {
	base *httpclient.BaseClient
}

func NewClient(baseURL string, timout time.Duration) *HTTPClient {
	return &HTTPClient{
		base: httpclient.New(baseURL, timout),
	}
}

func (c *HTTPClient) GetGroup(
	ctx context.Context,
	id uuid.UUID,
) (GroupDTO, error) {
	path := "/" + id.String()
	var response GroupDTO
	err := c.base.Get(ctx, path, &response, func(status int, body []byte) error {
		switch status {
		case http.StatusNotFound:
			return apierror.NewBadRequestError(fmt.Errorf("group not found"))
		case http.StatusForbidden:
			return apierror.NewBadRequestError(fmt.Errorf("not a group member"))
		default:
			return fmt.Errorf("ms_group returned %d", status)
		}
	})
	if err != nil {
		return GroupDTO{}, nil
	}
	return response, nil
}
