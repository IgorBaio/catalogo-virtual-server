package utils

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type Auth struct {
	JwtSecret []byte
}

func NewAuth() *Auth {
	return &Auth{
		JwtSecret: []byte(GetEnvVar("JWT_SECRET")),
	}
}

func (a *Auth) GenerateJWT(username string, password string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"password": password,
		"exp":      time.Now().Add(time.Hour * 1).Unix(), // Expira em 1 hora
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.JwtSecret)
}

func (a *Auth) ReverseJWT(tokenStr string) (jwt.MapClaims, error) {
	tokenStr = strings.Replace(tokenStr, "Bearer ", "", -1)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return a.JwtSecret, nil
	})

	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

// ValidateJWT validates the JWT token.
func (a *Auth) ValidateJWT(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return a.JwtSecret, nil
	})
}

