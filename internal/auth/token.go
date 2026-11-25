// Package auth issues and verifies the JWT access tokens used by the API.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned for any token that is malformed, expired,
// signed with another key or algorithm, or issued by someone else.
var ErrInvalidToken = errors.New("invalid or expired token")

// Claims are the claims carried by an access token. The user ID is the
// standard "sub" claim.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// TokenManager signs and parses HS256 access tokens.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
	issuer string
	now    func() time.Time
}

// NewTokenManager creates a TokenManager. The secret must be kept private and
// should be at least 32 random bytes.
func NewTokenManager(secret string, ttl time.Duration, issuer string) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl, issuer: issuer, now: time.Now}
}

// Issue creates a signed access token for the given user.
func (m *TokenManager) Issue(userID, role string) (string, time.Time, error) {
	now := m.now()
	expiresAt := now.Add(m.ttl)
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

// Parse verifies a token and returns its claims. Only HS256 tokens from this
// issuer with an expiry and a subject are accepted.
func (m *TokenManager) Parse(token string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil || !parsed.Valid || claims.Subject == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
