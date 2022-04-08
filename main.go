package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/controllers"
	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
	c.GET("/:guid", controllers.GetTokens(c, bcryptCost, dbName, collectionName, client, ctx))
	c.POST("/refresh", controllers.Refresh(c, bcryptCost, dbName, collectionName, client, ctx))

	c.Run(":" + port)
}
