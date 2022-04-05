package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/database"
	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/structs"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kataras/jwt"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	secret         string
	port           string
	mongoUri       string
	dbName         string
	collectionName string
)

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

	dbName, ok = os.LookupEnv("DB_NAME")
	if !ok {
		fmt.Println("DB_NAME not found, using default: test")
		dbName = "test"
	}

	collectionName, ok = os.LookupEnv("COLLECTION_NAME")
	if !ok {
		fmt.Println("COLLECTION_NAME not found, using default: users")
		collectionName = "users"
	}

}

func main() {
	client, ctx, cancel, err := database.Connect(mongoUri)
	if err != nil {
		log.Fatal(err)
	}
	err = database.CreateDbAndDocument(client, ctx, dbName, collectionName)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(client, ctx, cancel)
	defer database.DropDatabase(client, ctx, dbName)
	defer database.DropCollection(client, ctx, dbName, collectionName)

	c := gin.New()
	c.Use(gin.Logger())

	c.POST("/", func(c *gin.Context) {
		c.Writer.Header().Add("Content-Type", "application/json")
		var user structs.User
		validate := validator.New()

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			fmt.Println("1", err)
			return
		}

		valErr := validate.Struct(user)
		if valErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": valErr.Error()})
			fmt.Println("valErr", valErr)
			return
		}

		exist, err := database.CheckIfExist(client, ctx, dbName, collectionName, "email", user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			fmt.Println("check if exist", err)
			return
		}
		if exist {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User with this email already exists"})
			fmt.Println("user exist", err)
			return
		}

		fmt.Println(user)

		document, err := bson.Marshal(user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			fmt.Println("3", err)
			return
		}

		insertOneResult, err := database.InsertOne(client, ctx, dbName, collectionName, document)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			fmt.Println("4", err)
			return
		}
		fmt.Println("Result of InsertOne:")
		fmt.Println(insertOneResult.InsertedID)

		token, err := jwt.Sign(jwt.HS512, secret, user, jwt.MaxAge(15*time.Minute))
		if err != nil {
			panic(err)
		}

		refreshToken, err := jwt.Sign(jwt.HS512, secret, user, jwt.MaxAge(time.Hour*24*30))
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, structs.Token{
			AccessToken:  string(token),
			RefreshToken: string(refreshToken),
		})
	})

	c.POST("/refresh", func(c *gin.Context) {
		c.Writer.Header().Add("Content-Type", "application/json")
		/*refreshToken := c.Request.Header.Get("RefreshToken")
		if refreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token is empty"})
			return
		}*/
		accessToken := c.Request.Header.Get("AccessToken")
		if accessToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "access token is empty"})
			return
		}

		_, err := jwt.Verify(jwt.HS256, secret, []byte(accessToken))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		/*_, err = jwt.Verify(jwt.HS256, secret, []byte(refreshToken))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}*/

		Atoken, err := jwt.Sign(jwt.HS512, secret, jwt.Claims{}, jwt.MaxAge(15*time.Minute))
		if err != nil {
			panic(err)
		}

		Rtoken, err := jwt.Sign(jwt.HS512, secret, jwt.Claims{}, jwt.MaxAge(time.Hour*24*30))
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, structs.Token{
			AccessToken:  string(Atoken),
			RefreshToken: string(Rtoken),
		})
	})

	c.Run(":" + port)
}
