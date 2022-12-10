package main

import (
	"context"
	"fmt"
	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	MONGO_HOST_ENV     = "MONGO_HOST"
	MONGO_PORT_ENV     = "MONGO_PORT"
	MONGO_USERNAME_ENV = "MONGO_USERNAME"
	MONGO_PASSWORD_ENV = "MONGO_PASSWORD"
	MONGO_DB_ENV       = "MONGO_DB"
)

func newMongoClient() (*mongo.Collection, error) {
	host := common.GetEnv(MONGO_HOST_ENV, "localhost")
	port := common.GetEnv(MONGO_PORT_ENV, "27017")
	username := common.GetEnv(MONGO_USERNAME_ENV, "admin")
	password := common.GetEnv(MONGO_PASSWORD_ENV, "admin")
	dbName := common.GetEnv(MONGO_DB_ENV, "pressandplay")

	credential := options.Credential{
		Username: username,
		Password: password,
	}
	client, err := mongo.Connect(context.TODO(),
		options.Client().
			ApplyURI(fmt.Sprintf("mongodb://%s:%s", host, port)),
		options.Client().
			SetAuth(credential).
			SetAppName(SERVICE_NAME))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	return client.Database(dbName).Collection(CourtTableName), nil
}

func verifyDatabaseConnection(databaseConnection *mongo.Collection) error {
	if databaseConnection == nil {
		return fmt.Errorf("database connection not initialized")
	}
	return nil
}
