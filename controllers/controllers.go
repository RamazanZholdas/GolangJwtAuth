package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/database"
	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/structs"
	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/tokens"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func GetTokens(g *gin.Engine, bcryptCost int, dbName string, collectionName string, client *mongo.Client, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
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
	}
}

func Refresh(g *gin.Engine, bcryptCost int, dbName string, collectionName string, client *mongo.Client, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Content-Type", "application/json")
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

		hash, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcryptCost)
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
	}
}
