package invitation

import (
	"crypto/rsa"
	"errors"
	"shared/auth/security"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

const (
	issuer   = "rateio-ms-group"
	audience = "rateio-invitations"
	ttl      = 72 * time.Hour
)

type Claims struct {
	GroupID   uuid.UUID `json:"group_id"`
	InviterID uuid.UUID `json:"inviter_id"`
	jwt.RegisteredClaims
}

type TokenService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewTokenService(
	privPath, pubPath string,
) (*TokenService, error) {
	priv, err := security.LoadPrivateKey(privPath)
	if err != nil {
		return nil, err
	}
	pub, err := security.LoadPublicKey(pubPath)
	if err != nil {
		return nil, err
	}

	return &TokenService{privateKey: priv, publicKey: pub}, nil
}

func (s *TokenService) Generate(
	groupID, inviterID, jti uuid.UUID,
) (string, time.Time, error) {
	exp := time.Now().Add(ttl)
	claims := Claims{
		GroupID:   groupID,
		InviterID: inviterID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        jti.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.privateKey)
	return signed, exp, err
}

func (s *TokenService) Validate(raw string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.publicKey, nil
	},
		jwt.WithIssuer(issuer),
		jwt.WithAudience(audience),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid invitation token")
	}
	return claims, nil
}
