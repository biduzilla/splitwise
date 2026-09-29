package security

import (
	"crypto/rsa"
	"fmt"
	"ms_auth/internal/core/config"
	"shared/auth/domain"
	sharedsecurity "shared/auth/security"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

type JwtService struct {
	*sharedsecurity.JwtService
	config     config.Config
	privateKey *rsa.PrivateKey
}

func NewService(
	config config.Config,
) (*JwtService, error) {
	service := &JwtService{
		config: config,
	}

	if err := service.loadKeys(); err != nil {
		return nil, fmt.Errorf("failed to load RSA keys: %w", err)
	}

	sharedSvc, err := sharedsecurity.NewService(config.Base)
	if err != nil {
		return nil, err
	}

	service.JwtService = sharedSvc

	return service, nil
}

func (s *JwtService) loadKeys() error {
	privateKey, err := sharedsecurity.LoadPrivateKey(s.config.Base.Security.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}
	s.privateKey = privateKey

	return nil
}

func (s *JwtService) CreateToken(
	user domain.UserDetails,
	tokenType sharedsecurity.TokenType,
) (string, error) {
	var expiration time.Duration

	switch tokenType {
	case sharedsecurity.TokenTypeAccess:
		expiration = sharedsecurity.AccessTokenExpiration
	case sharedsecurity.TokenTypeRefresh:
		expiration = sharedsecurity.RefreshTokenExpiration
	default:
		expiration = 0
	}

	now := time.Now()
	claims := sharedsecurity.TokenClaims{
		UserID:   user.GetID(),
		Username: user.GetUsername(),
		IsAtivo:  user.GetIsAtivo(),
		Type:     tokenType,
		Roles:    user.GetRoles(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    sharedsecurity.TokenIssuer,
			Audience:  jwt.ClaimStrings{sharedsecurity.TokenAudience},
			Subject:   user.GetID().String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}
