package database

import (
	"context"
	"time"

	"github.com/RamazanZholdas/MyGoPlayground/jwtGinGolangTest/structs"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(uri string) (*mongo.Client, context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}

func CreateDbAndDocument(client *mongo.Client, ctx context.Context, dbName string, collectionName string) error {
	demoDB := client.Database(dbName)
	if err := demoDB.CreateCollection(ctx, collectionName); err != nil {
		return err
	}
	return nil
}

func InsertOne(client *mongo.Client, ctx context.Context, dataBase, col string, doc interface{}) (*mongo.InsertOneResult, error) {
	collection := client.Database(dataBase).Collection(col)
	result, err := collection.InsertOne(ctx, doc)
	return result, err
}

func DropCollection(client *mongo.Client, ctx context.Context, dbName string, collectionName string) error {
	collection := client.Database(dbName).Collection(collectionName)
	if err := collection.Drop(ctx); err != nil {
		return err
	}
	return nil
}

func Close(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func DropDatabase(client *mongo.Client, ctx context.Context, dbName string) error {
	if err := client.Database(dbName).Drop(ctx); err != nil {
		return err
	}
	return nil
}

func FindOne(client *mongo.Client, ctx context.Context, dataBase, col string, filter interface{}) (user structs.User, err error) {
	collection := client.Database(dataBase).Collection(col)
	err = collection.FindOne(ctx, filter).Decode(&user)
	return user, err
}

func UpdateOne(client *mongo.Client, ctx context.Context, dataBase,
	col string, filter, update interface{}) (result *mongo.UpdateResult, err error) {
	collection := client.Database(dataBase).Collection(col)
	result, err = collection.UpdateOne(ctx, filter, update)
	return
}
