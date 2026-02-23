package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the custom JWT claims for Portwhine access tokens.
// For API key authentication, Scopes is populated from the key's scopes.
type Claims struct {
	jwt.RegisteredClaims
	UserID   string   `json:"uid"`
	Username string   `json:"usr"`
	Role     string   `json:"role"`
	TeamIDs  []string `json:"teams,omitempty"`
	Scopes   []string `json:"-"` // populated only for API key auth, not in JWT
}

// JWTService handles JWT token creation and validation.
type JWTService struct {
	signingKey []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTService creates a new JWTService with the given signing key and token
// time-to-live durations.
func NewJWTService(signingKey []byte, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		signingKey: signingKey,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// GenerateTokenPair creates a new access/refresh token pair for the given user.
// It returns the signed access token string, the signed refresh token string,
// the access token expiration time, and any error encountered.
func (s *JWTService) GenerateTokenPair(userID, username, role string, teamIDs []string) (accessToken, refreshToken string, expiresAt time.Time, err error) {
	now := time.Now()
	accessExp := now.Add(s.accessTTL)

	accessClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "portwhine",
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   userID,
		Username: username,
		Role:     role,
		TeamIDs:  teamIDs,
	}

	accessJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessJWT.SignedString(s.signingKey)
	if err != nil {
		return "", "", time.Time{}, err
	}

	refreshExp := now.Add(s.refreshTTL)
	refreshClaims := &jwt.RegisteredClaims{
		Issuer:    "portwhine",
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(refreshExp),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	refreshJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshJWT.SignedString(s.signingKey)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, refreshToken, accessExp, nil
}

// ValidateAccessToken parses and validates an access token string, returning
// the embedded Claims on success.
func (s *JWTService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.signingKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid access token")
	}

	return claims, nil
}

// ValidateRefreshToken parses and validates a refresh token string, returning
// the subject (user ID) on success.
func (s *JWTService) ValidateRefreshToken(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.signingKey, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	return claims.Subject, nil
}
