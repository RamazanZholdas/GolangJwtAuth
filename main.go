package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kataras/jwt"
)

const (
	port      = "8080"
	secretKey = "secret"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type User struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

func main() {
	c := gin.Default()

	c.GET("/", func(c *gin.Context) {
		user := User{}
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := json.NewDecoder(c.Request.Body).Decode(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := jwt.Sign(jwt.HS256, secretKey, user, jwt.MaxAge(15*time.Minute))
		if err != nil {
			panic(err)
		}

		refreshToken, err := jwt.Sign(jwt.HS256, secretKey, user, jwt.MaxAge(time.Hour))
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, Token{
			AccessToken:  string(token),
			RefreshToken: string(refreshToken),
		})
	})

	c.Run(":" + port)
}
