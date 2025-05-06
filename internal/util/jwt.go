package util

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

type Jwt interface {
	GenerateToken(userID string, role string) (string, error)
	ValidateToken(token string) (jwt.MapClaims, error)
}

const TOKEN_EXPIRED_TIME = 30 * time.Minute

type JwtImpl struct {
}

func NewJwtImpl() *JwtImpl {
	return &JwtImpl{}
}

func (j JwtImpl) GenerateToken(userID string, role string) (string, error) {
	expireTime := time.Now().Add(TOKEN_EXPIRED_TIME)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"role":   role,
		"exp":    expireTime.Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j JwtImpl) ValidateToken(token string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
