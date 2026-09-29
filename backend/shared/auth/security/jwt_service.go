package security

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"shared/apierror"
	"shared/auth/domain"
	"shared/config"
	"strings"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

type JwtService struct {
	config    config.Config
	publicKey *rsa.PublicKey
}

const (
	AccessTokenExpiration  = 3 * time.Hour
	RefreshTokenExpiration = 7 * 24 * time.Hour
	TokenIssuer            = "mini_ecomerce-go-api"
	TokenAudience          = "mini_ecomerce-go-clients"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type TokenClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	IsAtivo  bool      `json:"is_ativo"`
	Type     TokenType `json:"type"`
	Roles    []string  `json:"roles"`
	jwt.RegisteredClaims
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

	return service, nil
}

func (s *JwtService) loadKeys() error {
	publicKey, err := LoadPublicKey(s.config.Security.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load public key: %w", err)
	}
	s.publicKey = publicKey

	return nil
}

func (s *JwtService) ExtractAuthenticatedUser(
	tokenString string,
) (domain.UserDetails, error) {
	claims, err := s.ValidateToken(tokenString, TokenTypeAccess)
	if err != nil {
		return nil, err
	}

	return domain.NewAuthenticatedUser(
		claims.UserID,
		claims.Username,
		claims.IsAtivo,
		claims.Roles,
	), nil
}

func (s *JwtService) ValidateToken(
	tokenString string,
	expectedType TokenType,
) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&TokenClaims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return s.publicKey, nil
		},
		jwt.WithIssuer(TokenIssuer),
		jwt.WithAudience(TokenAudience),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apierror.ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenInvalidIssuer) || errors.Is(err, jwt.ErrTokenInvalidAudience) {
			return nil, apierror.ErrInvalidTokenClaims
		}

		if strings.Contains(err.Error(), "token has invalid claims") {
			return nil, apierror.ErrInvalidTokenClaims
		}
		return nil, apierror.NewHTTPError(
			fmt.Sprintf("failed to parse token: %s", err.Error()),
			http.StatusBadRequest,
			nil,
		)
	}

	if !token.Valid {
		return nil, apierror.ErrInvalidCredentials
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, apierror.ErrInvalidTokenClaims
	}

	if claims.Type != expectedType {
		return nil, apierror.ErrInvalidTokenType
	}

	if !claims.IsAtivo {
		return nil, apierror.ErrInactiveAccount
	}

	return claims, nil
}

func (s *JwtService) GetPublicKey() *rsa.PublicKey {
	return s.publicKey
}

func (s *JwtService) GetAccessTokenExpiration() time.Duration {
	return AccessTokenExpiration
}

func (s *JwtService) GetRefreshTokenExpiration() time.Duration {
	return RefreshTokenExpiration
}
