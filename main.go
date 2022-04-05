package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/database"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kataras/jwt"
)

var (
	secret   string
	port     string
	mongoUri string
	/*client   *mongo.Client
	ctx      context.Context
	cancel   context.CancelFunc*/
)

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type User struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

func init() {
	var ok bool
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	secret, ok = os.LookupEnv("SECRET")
	if !ok {
		fmt.Println("SECRET not found, using default: secret")
		secret = "secret"
	}

	port, ok = os.LookupEnv("PORT")
	if !ok {
		fmt.Println("PORT not found, using default: 8080")
		port = "8080"
	}

	mongoUri, ok = os.LookupEnv("MONGO_URI")
	if !ok {
		fmt.Println("MONGO_URI not found, using default: mongodb://localhost:27017")
		mongoUri = "mongodb://localhost:27017"
	}
}

func main() {
	client, ctx, cancel, err := database.Connect(mongoUri)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(client, ctx, cancel)

	c := gin.Default()

	c.POST("/", func(c *gin.Context) {
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

		document := bson.Marshal(&user)

		insertOneResult, err := database.InsertOne(client, ctx, "test", "users", document)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		fmt.Println("Result of InsertOne:")
		fmt.Println(insertOneResult.InsertedID)

		token, err := jwt.Sign(jwt.HS256, secret, user, jwt.MaxAge(15*time.Minute))
		if err != nil {
			panic(err)
		}

		refreshToken, err := jwt.Sign(jwt.HS256, secret, user, jwt.MaxAge(time.Hour))
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, Token{
			AccessToken:  string(token),
			RefreshToken: string(refreshToken),
		})
	})

	c.GET("/refresh", func(c *gin.Context) {
		//refresh access and refresh tokens
		refreshToken := c.Request.Header.Get("RefreshToken")
		if refreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token is empty"})
			return
		}

	})

	c.Run(":" + port)
}
