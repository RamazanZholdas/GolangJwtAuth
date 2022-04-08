package tokens

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/structs"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	uuid "github.com/satori/go.uuid"
)

func getSecretKey() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	secret, ok := os.LookupEnv("SECRET")
	if !ok {
		fmt.Println("SECRET not found, using default: secret")
		secret = "secret"
	}
	return secret
}

func GenerateTokens(uid string) (signedToken string, signedRefreshToken string, jti string, err error) {
	jti = uuid.NewV4().String()
	claims := &structs.Token{
		GUID: uid,
		Jti:  jti,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &structs.Token{
		Jti: jti,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(getSecretKey()))
	if err != nil {
		log.Fatal(err)
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(getSecretKey()))
	if err != nil {
		log.Fatal(err)
		return
	}

	return token, refreshToken, jti, err
}

func ParseRefreshToken(refreshToken string) (jti string, err error) {
	claims := &structs.Token{}
	_, err = jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(getSecretKey()), nil
	})
	if err != nil {
		return
	}
	jti = claims.Jti
	return
}
