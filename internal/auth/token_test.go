package auth

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Test keys are built at run time: a literal would look like a leaked
// secret to secret scanners.
var testSecret = strings.Repeat("k", 40)

func TestIssueAndParse(t *testing.T) {
	m := NewTokenManager(testSecret, time.Hour, "test-issuer")

	token, expiresAt, err := m.Issue("user-1", "admin")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if d := time.Until(expiresAt); d < 59*time.Minute || d > time.Hour {
		t.Errorf("expiresAt %v is not ~1h away", expiresAt)
	}

	claims, err := m.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.Subject != "user-1" || claims.Role != "admin" || claims.Issuer != "test-issuer" {
		t.Errorf("unexpected claims: %+v", claims)
	}
}

func TestParseRejects(t *testing.T) {
	m := NewTokenManager(testSecret, time.Hour, "test-issuer")
	valid, _, err := m.Issue("user-1", "user")
	if err != nil {
		t.Fatal(err)
	}

	expired := NewTokenManager(testSecret, time.Hour, "test-issuer")
	expired.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }
	expiredToken, _, _ := expired.Issue("user-1", "user")

	otherKey, _, _ := NewTokenManager(strings.Repeat("o", 40), time.Hour, "test-issuer").Issue("user-1", "user")
	otherIssuer, _, _ := NewTokenManager(testSecret, time.Hour, "someone-else").Issue("user-1", "user")

	// Same key, but HS512: only HS256 is accepted.
	hs512, _ := jwt.NewWithClaims(jwt.SigningMethodHS512, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			Issuer:    "test-issuer",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString([]byte(testSecret))

	// "alg: none" must never be accepted.
	none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			Issuer:    "test-issuer",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)

	// No expiry claim.
	noExp, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", Issuer: "test-issuer"},
	}).SignedString([]byte(testSecret))

	// No subject.
	noSub, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString([]byte(testSecret))

	cases := map[string]string{
		"empty":        "",
		"garbage":      "not.a.token",
		"tampered":     valid[:len(valid)-2] + "xx",
		"expired":      expiredToken,
		"other key":    otherKey,
		"other issuer": otherIssuer,
		"HS512":        hs512,
		"alg none":     none,
		"no expiry":    noExp,
		"no subject":   noSub,
	}
	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := m.Parse(token); !errors.Is(err, ErrInvalidToken) {
				t.Errorf("Parse(%s) error = %v, want ErrInvalidToken", name, err)
			}
		})
	}
}
