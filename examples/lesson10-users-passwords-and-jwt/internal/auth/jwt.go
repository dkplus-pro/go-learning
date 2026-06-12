package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

type UserClaims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

func (m *TokenManager) Generate(user User) (string, error) {
	now := time.Now()
	claims := UserClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *TokenManager) Verify(raw string) (User, error) {
	token, err := jwt.ParseWithClaims(raw, &UserClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %s", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return User{}, ErrUnauthorized
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return User{}, ErrUnauthorized
	}

	return User{ID: claims.UserID, Email: claims.Email}, nil
}
