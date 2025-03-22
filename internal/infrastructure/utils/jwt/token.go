package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT interface {
	GenerateToken(claims Claims, expiry time.Duration) (string, time.Time, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type tokenService struct {
	secret string
}

func newTokenService(secret string) JWT {
	return &tokenService{secret: secret}
}

func (t *tokenService) GenerateToken(c Claims, expiry time.Duration) (string, time.Time, error) {
	expirationTime := time.Now().Add(expiry)
	c.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenString, err := token.SignedString([]byte(t.secret))

	return tokenString, expirationTime, err
}

func (t *tokenService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKeyType
}
