package token

import (
	"care-cordination/lib/nanoid"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
	audience      string
}

type AccessTokenClaims struct {
	Scope string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	Tokenfamily string `json:"token_family"`
	TokenHash   string `json:"token_hash"`
	jwt.RegisteredClaims
}

func NewTokenManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		issuer:        "care-coordination",
		audience:      "care-coordination",
	}
}

func (tm *TokenManager) GenerateAccessToken(userID string, now time.Time) (string, error) {

	accessExpire := now.Add(tm.accessTTL)

	accessClaims := &AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tm.issuer,
			Audience:  jwt.ClaimStrings{tm.audience},
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpire),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	return accessToken.SignedString(tm.accessSecret)
}

func (tm *TokenManager) GenerateRefreshToken(userID string, now time.Time) (string, *RefreshTokenClaims, error) {
	refreshExpire := now.Add(tm.refreshTTL)
	tokenHash := nanoid.Generate()
	tokenFamily := nanoid.Generate()

	refreshClaims := &RefreshTokenClaims{
		Tokenfamily: tokenFamily,
		TokenHash:   tokenHash,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tm.issuer,
			Audience:  jwt.ClaimStrings{tm.audience},
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpire),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	token, err := refreshToken.SignedString(tm.refreshSecret)
	if err != nil {
		return "", nil, err
	}
	return token, refreshClaims, nil
}

func (tm *TokenManager) ValidateAccessToken(tokenStr string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.accessSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid access token")
}

func (tm *TokenManager) ValidateRefreshToken(tokenStr string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.refreshSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid refresh token")
}
