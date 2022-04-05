package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(uri string) (*mongo.Client, context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

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

//check if the value in collection exist
func CheckIfExist(client *mongo.Client, ctx context.Context, dbName, collectionName, field, value string) (bool, error) {
	collection := client.Database(dbName).Collection(collectionName)
	filter := bson.M{
		field: value,
	}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}
