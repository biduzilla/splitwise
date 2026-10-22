package user

import (
	"context"
	"net/url"
	"shared/filters"
	"shared/httpx/httpclient"
	"strconv"
	"strings"
	"time"
	"uuid"
)

type HTTPClient struct {
	base *httpclient.BaseClient
}

func NewClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		base: httpclient.New(baseURL, timeout),
	}
}

type UserDTO struct {
	ID      uuid.UUID `json:"id"`
	Email   string    `json:"email"`
	Name    string    `json:"name"`
	Version int       `json:"version"`
}

type UserListResponse struct {
	Content  []UserDTO        `json:"content"`
	Metadata filters.Metadata `json:"metadata"`
}

func (c *HTTPClient) FindAllByIds(
	ctx context.Context,
	ids []uuid.UUID,
	page int,
	pageSize int,
) ([]UserDTO, filters.Metadata, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}

	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = id.String()
	}

	q := url.Values{}
	q.Set("ids", strings.Join(parts, ","))
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))

	path := "/batch?" + q.Encode()

	var response UserListResponse
	err := c.base.Get(ctx, path, &response, nil)
	if err != nil {
		return nil, filters.Metadata{}, err
	}

	return response.Content, response.Metadata, nil
}

func (c *HTTPClient) FindById(
	ctx context.Context,
	id uuid.UUID,
) (UserDTO, error) {
	path := "/" + id.String()

	var response UserDTO
	err := c.base.Get(ctx, path, &response, nil)
	if err != nil {
		return UserDTO{}, err
	}

	return response, nil
}
