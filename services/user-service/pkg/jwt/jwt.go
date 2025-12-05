package jwtutil

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Signer struct {
	secret    []byte
	accessTTL time.Duration
}

func NewSigner(secret string, accessTTL time.Duration) *Signer {
	return &Signer{secret: []byte(secret), accessTTL: accessTTL}
}

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Status string `json:"status"`
	jwt.RegisteredClaims
}

func (s *Signer) NewAccess(userID uint, email, status string) (token string, jti string, err error) {
	now := time.Now()
	jti = uuid.NewString()

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Status: status,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "user-svc",
			Subject:   strconv.FormatUint(uint64(userID), 10),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = t.SignedString(s.secret)
	return
}
