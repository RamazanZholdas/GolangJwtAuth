package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/database"
	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/structs"
	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/tokens"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var (
	port           string
	mongoUri       string
	dbName         string
	collectionName string
	bcryptCost     int
)

func init() {
	var ok bool
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	bcryptCostStr, ok := os.LookupEnv("BCRYPT_COST")
	if !ok {
		log.Fatal("BCRYPT_COST not found, using minimum cost: 4")
		bcryptCostStr = "4"
	}
	bcryptCost, _ = strconv.Atoi(bcryptCostStr)

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

	c := gin.Default()
	// http://localhost:8080/{paste guid here}
	c.GET("/:guid", func(c *gin.Context) {
		c.Writer.Header().Add("Content-Type", "application/json")
		var user structs.User

		token, refreshToken, jti, err := tokens.GenerateTokens(c.Param("guid"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcryptCost)
		if err != nil {
			log.Fatal(err)
		}
		user.GUID = c.Param("guid")
		user.RefreshToken = string(hash)
		user.Jti = jti

		insertOneResult, err := database.InsertOne(client, ctx, dbName, collectionName, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			fmt.Println("4", err)
			return
		}
		fmt.Println("Result of InsertOne:")
		fmt.Println(insertOneResult.InsertedID)

		c.JSON(http.StatusOK, gin.H{
			"AccessToken":  token,
			"RefreshToken": base64.RawURLEncoding.EncodeToString([]byte(refreshToken)),
		})
	})

	c.POST("/refresh", func(c *gin.Context) {
		refreshToken := c.Request.Header.Get("RefreshToken")
		if refreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "RefreshToken is empty"})
			return
		}
		userJti, err := tokens.ParseRefreshToken(refreshToken)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println("userJti:", userJti)
		var user structs.User
		user, err = database.FindOne(client, ctx, dbName, collectionName, bson.M{"jti": userJti})
		if err != nil {
			fmt.Println("user not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if user.GUID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.RefreshToken), []byte(refreshToken)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		token, refreshToken, jti, err := tokens.GenerateTokens(user.GUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.MinCost)
		if err != nil {
			log.Fatal(err)
		}
		user.RefreshToken = string(hash)
		user.Jti = jti
		update := bson.M{"$set": bson.M{"refreshToken": user.RefreshToken, "jti": user.Jti}}
		updateResult, err := database.UpdateOne(client, ctx, dbName, collectionName, bson.M{"guid": user.GUID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		fmt.Println("Result of UpdateOne:")
		fmt.Println(updateResult.UpsertedID)

		c.JSON(http.StatusOK, gin.H{
			"AccessToken":  token,
			"RefreshToken": base64.RawURLEncoding.EncodeToString([]byte(refreshToken)),
		})
	})

	c.Run(":" + port)
}
