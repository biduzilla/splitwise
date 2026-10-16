package auth

import (
	"shared/models"
	"time"

	"uuid"
)

type RefreshToken struct {
	models.BaseModel
	ID        uuid.UUID
	TokenHash string
	UserID    uuid.UUID
	ExpiresAt time.Time
	Family    uuid.UUID
	Revoked   bool
}

type SessionResponse struct {
	Authenticated bool      `json:"authenticated"`
	UserID        uuid.UUID `json:"user_id,omitempty"`
	Email         *string   `json:"email,omitempty"`
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}
