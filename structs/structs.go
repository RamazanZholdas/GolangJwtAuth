package structs

import (
	"github.com/dgrijalva/jwt-go"
)

type Token struct {
	GUID string
	Jti  string
	jwt.StandardClaims
}

type User struct {
	GUID         string `json:"guid"`
	Jti          string `json:"jti"`
	RefreshToken string `json:"refreshToken"`
}
